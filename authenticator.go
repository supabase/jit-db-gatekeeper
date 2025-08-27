package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	_ "github.com/lib/pq"
)

type AuthMethod string

const (
	AuthPassword AuthMethod = "password"
	AuthPat      AuthMethod = "pat"
	AuthJwt      AuthMethod = "jwt"
)

type authenticator struct {
	// use password authentication, this falls through to the native scram-sha256 auth
	AuthMethod AuthMethod

	ApiUrl string
}

type AuthZRequest struct {
	Role  string `json:"role"`
	Rhost string `json:"rhost"`
}
type UserPermissionSet struct {
	UserId string   `json:"user_id"`
	Role   UserRole `json:"user_role"`
}

type UserRole struct {
	Role            string          `json:"role"`
	ExpiresAt       string          `json:"expires_at,omitempty"`
	AllowedNetworks AllowedNetworks `json:"allowed_networks"`
}

type AllowedNetworks struct {
	AllowedCidrs   []CidrRange `json:"allowed_cidrs,omitempty"`
	AllowedCidrsV6 []CidrRange `json:"allowed_cidrs_v6,omitempty"`
}

type CidrRange struct {
	Cidr string `json:"cidr,omitempty"`
}

/* discoverAuthenticator uses the auth token to determine which authentication mechanism to use */
func discoverAuthenticator(ctx context.Context, config *config, token string) (*authenticator, error) {
	if looksLikePAT(token) {
		return &authenticator{
			ApiUrl:     config.AuthAPIURL,
			AuthMethod: AuthPat,
		}, nil
	}
	if looksLikeJWT(token) {
		return &authenticator{
			ApiUrl:     config.AuthAPIURL,
			AuthMethod: AuthJwt,
		}, nil
	}
	return &authenticator{
		AuthMethod: AuthPassword,
	}, nil
}

// Authenticate authenticates a user with the provided token.
func (a *authenticator) Authenticate(ctx context.Context, user, token string) error {
	if a.AuthMethod == AuthPassword {
		return authPassword(ctx, user, token)
	}
	// AuthPat and AuthJwt use the same API
	return authApi(ctx, a.ApiUrl, user, token)
}

// looksLikePAT simply checks if a supplied token has the supabase PAT prefix
func looksLikePAT(token string) bool {
	return token[:4] == "sbp_"
}

// looksLikeJWT checks if a token has the format of a JWT. It does not validate the JWT
func looksLikeJWT(token string) bool {
	parts := strings.Split(token, ".")
	return len(parts) == 3 && parts[0][:3] == "eyJ" && parts[1][:3] == "eyJ"
}

/* authPassword will attempt to auth  to the local postgres database */
func authPassword(ctx context.Context, username, password string) error {
	connStr := fmt.Sprintf("user=%s password=%s dbname=authdbsupabase sslmode=disable host=127.0.0.1", username, password)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		// valid username + password and permitted to login
		if strings.Contains(err.Error(), "database \"authdbsupabase\" does not exist") {
			return nil
		}
		return err
	}
	return nil
}

func authApi(ctx context.Context, apiUrl, username, token string) error {
	// make an API request to check if the user is authorized to login
	// uses the incoming token to authenticate against the API
	// giving the guarantee that the user token is still valid and permitted
	// to  interact with the project
	if rhost, ok := ctx.Value(rhostKey).(string); !ok {
		return fmt.Errorf("context does not have rhost")
	} else {
		jsonData, err := json.Marshal(&AuthZRequest{username, rhost})
		if err != nil {
			panic(err)
		}

		client := &http.Client{}
		req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}
		// set auth for API server, only bearer support for now
		req.Header.Add("Authorization", "Bearer "+token)
		req.Header.Add("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			// user has no authorization setup if a 406 error is returned
			if resp.StatusCode == http.StatusNotAcceptable {
				return fmt.Errorf("user not authorized for JIT access to database")
			}

			if resp.StatusCode == http.StatusForbidden {
				return fmt.Errorf("user not authorized due to restriction")
			}
			// something else went wrong
			return fmt.Errorf("failed with status: %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var perms UserPermissionSet
		if err := json.Unmarshal(body, &perms); err != nil {
			return err
		}

		// validate the user's permission
		return isPermitted(ctx, username, perms)
	}
}

func isPermitted(ctx context.Context, username string, perms UserPermissionSet) error {
	if username == "" {
		return fmt.Errorf("empty username")
	}
	// not strictly required anymore,
	// but double check the response from the server is what we were expecting
	if perms.Role.Role == username {
		return nil
	}
	return fmt.Errorf("not permitted to assume %s", username)
}
