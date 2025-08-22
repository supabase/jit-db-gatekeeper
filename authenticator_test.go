package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthenticate_Discover(t *testing.T) {
	// config to  use  for testing
	c := &config{
		AuthAPIURL: "https://mocked.api",
	}
	ctx := context.Background()

	t.Run("discovers PAT token for auth", func(t *testing.T) {
		token := "sbp_1112223336d4dddd54e60cfa33441499b182bbbb"
		auth, err := discoverAuthenticator(ctx, c, token)
		assert.NoError(t, err)
		assert.Equal(t, &authenticator{AuthMethod: AuthPat, ApiUrl: c.AuthAPIURL}, auth)
	})

	t.Run("discovers JWT token for auth", func(t *testing.T) {
		token := "eyJhbGciOiJSUzI1NiIsImtpZCI6IjcyYjY2NjA1IiwidHlwIjoiSldUIn0.eyJhYWwiOiJhYWwyIiwiYW1yIjpbeyJtZXRob2QiOiJ0b3RwIiwidGltZXN0YW1wIjoxNzUxOTc4NzU1fSx7Im1ldGhvZCI6InBhc3N3b3JkIiwidGltZXN0YW1wIjoxNzUxOTc4NzM2fV0sImFwcF9tZXRhZGF0YSI6eyJwcm92aWRlciI6ImVtYWlsIiwicHJvdmlkZXJzIjpbImVtYWlsIiwiZ2l0aHViIl19LCJhdWQiOiJhdXRoZW50aWNhdGVkIiwiZW1haWwiOiJldGllbm5lQHN1cGFiYXNlLmlvIiwiZXhwIjoxNzUzOTUzMTQ0LCJpYXQiOjE3NTM5NTI1NDQsImlzX2Fub255bW91cyI6ZmFsc2UsImlzcyI6Imh0dHBzOi8vYWx0LnN1cGFiYXNlLmdyZWVuL2F1dGgvdjEiLCJwaG9uZSI6IiIsInJvbGUiOiJhdXRoZW50aWNhdGVkIiwic2Vzc2lvbl9pZCI6ImI2MjZhNjMzLThhMzktNGFkYy1hY2FmLTVhYTNhMmQwYTg1ZiIsInN1YiI6ImZmOTIxZDE5LTk0NWYtNDRiMy1iNzg2LTE5MTViNmViMWQwZSIsInVzZXJfbWV0YWRhdGEiOnsiYXZhdGFyX3VybCI6Imh0dHBzOi8vYXZhdGFycy5naXRodWJ1c2VyY29udGVudC5jb20vdS80MjAwODMyP3Y9NCIsImVtYWlsIjoiZXN0YWxtYW5zQGdtYWlsLmNvbSIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJmdWxsX25hbWUiOiJFdGllbm5lIFN0YWxtYW5zIiwiaXNzIjoiaHR0cHM6Ly9hcGkuZ2l0aHViLmNvbSIsIm5hbWUiOiJFdGllbm5lIFN0YWxtYW5zIiwicGhvbmVfdmVyaWZpZWQiOmZhbHNlLCJwcmVmZXJyZWRfdXNlcm5hbWUiOiJzdGFhbGRyYWFkIiwicHJvdmlkZXJfaWQiOiI0MjAwODMyIiwic3ViIjoiNDIwMDgzMiIsInVzZXJfbmFtZSI6InN0YWFsZHJhYWQifX0.E4sLsEcjxg3WWsHVV7-37RqhvZqPRShVpCcaavEf2or88J4o1HS0bM_fGzUPULDfE0jzzu-2N9vvlvX41XVaCMsuVh5kspfcTBhQ9eCaqIYok_5nh5AiNafdI8mvSJTpwo9Qem1Fqj9Ka9pqg5mUCkU39r1N04a30py9xmI1hERN8C1rJK2BxMMDQkjjcWA0Bgyk8fyj8kwJ-CoYuIGVsOqI1rc__U3yxG48RmZrIXsgyl6PxQDjM724lVI2gjSQG2zIugT7QDn41OuZdEFKnVf5jPslt9zMl39CwnDhNiMSFBLKfaT6X6N9LcDU7N0vDw2Xp8YOj6RhKFxNUAVNYQ"
		auth, err := discoverAuthenticator(ctx, c, token)
		assert.NoError(t, err)
		assert.Equal(t, &authenticator{AuthMethod: AuthJwt, ApiUrl: c.AuthAPIURL}, auth)
	})

	t.Run("discovers Password for auth", func(t *testing.T) {
		token := "aPasswordString"
		auth, err := discoverAuthenticator(ctx, c, token)
		assert.NoError(t, err)
		assert.Equal(t, &authenticator{AuthMethod: AuthPassword}, auth)
	})
}

func mockServer(resp *UserPermissionSet) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			return
		}
	}))
}
func TestAuthenticate_authApi(t *testing.T) {
	t.Run("successful auth against Api", func(t *testing.T) {
		validUser := &UserPermissionSet{UserId: "99cf6d1d-7c39-46b4-bc58-688f6dd897ad", Role: UserRole{Role: "postgres"}}

		mockServer := mockServer(validUser)
		defer mockServer.Close()

		c := &config{
			AuthAPIURL: mockServer.URL,
		}
		ctx := context.Background()
		token := "sbp_1112223336d4dddd54e60cfa33441499b182bbbb"
		auth, err := discoverAuthenticator(ctx, c, token)
		assert.NoError(t, err)
		err = auth.Authenticate(ctx, "postgres", token)
		assert.NoError(t, err)
		defer mockServer.Close()
	})

	t.Run("empty user fails", func(t *testing.T) {
		validUser := &UserPermissionSet{UserId: "99cf6d1d-7c39-46b4-bc58-688f6dd897ad", Role: UserRole{Role: "postgres"}}

		mockServer := mockServer(validUser)
		defer mockServer.Close()

		c := &config{
			AuthAPIURL: mockServer.URL,
		}
		ctx := context.Background()
		token := "sbp_1112223336d4dddd54e60cfa33441499b182bbbb"
		auth, err := discoverAuthenticator(ctx, c, token)
		assert.NoError(t, err)
		err = auth.Authenticate(ctx, "", token)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "empty username")
		defer mockServer.Close()
	})

	t.Run("fails when user is not permitted for requested role", func(t *testing.T) {
		validUser := &UserPermissionSet{UserId: "99cf6d1d-7c39-46b4-bc58-688f6dd897ad", Role: UserRole{Role: "supabase_read_only_user"}}

		mockServer := mockServer(validUser)
		defer mockServer.Close()

		c := &config{
			AuthAPIURL: mockServer.URL,
		}
		ctx := context.Background()
		token := "sbp_1112223336d4dddd54e60cfa33441499b182bbbb"
		auth, err := discoverAuthenticator(ctx, c, token)
		assert.NoError(t, err)
		err = auth.Authenticate(ctx, "postgres", token)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "not permitted to assume postgres")
		defer mockServer.Close()
	})
}
