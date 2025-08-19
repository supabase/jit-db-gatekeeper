package main

/*
#cgo CFLAGS: -I.
#cgo LDFLAGS: -lpam -fPIC

#include <stdlib.h>
#include <security/pam_appl.h>
#include <security/pam_modules.h>

#ifdef __linux__
#include <security/pam_ext.h>
#endif

char* argv_i(const char **argv, int i);
void pam_syslog_str(pam_handle_t *pamh, int priority, const char *str);
*/
import "C"

import (
	"context"
	"fmt"
	"log/syslog"
	"unsafe"
)

type key int

const (
	rhostKey key = iota
)

func main() {
}

//export pam_sm_authenticate_go
func pam_sm_authenticate_go(pamh *C.pam_handle_t, flags C.int, argc C.int, argv **C.char) C.int {
	ctx := context.Background()

	// Copy args to Go strings
	args := make([]string, int(argc))
	for i := 0; i < int(argc); i++ {
		args[i] = C.GoString(C.argv_i(argv, C.int(i)))
	}
	// Parse config
	cfg, err := configFromArgs(args)
	if err != nil {
		pamSyslog(pamh, syslog.LOG_ERR, "failed to parse config: %v", err)
		return C.PAM_SERVICE_ERR
	}

	// Validate config
	if cfg.AuthAPIURL == "" {
		pamSyslog(pamh, syslog.LOG_WARNING, "no apiUrl set, only password auth will work")
	}

	// get the remote host from PAM_RHOST
	var cRhost *C.char
	if errnum := C.pam_get_item(pamh, C.PAM_RHOST, (*unsafe.Pointer)(unsafe.Pointer(&cRhost))); errnum != C.PAM_SUCCESS {
		pamSyslog(pamh, syslog.LOG_ERR, "failed to get rhost: %v", pamStrError(pamh, errnum))
		return errnum
	}

	// store the rhost in the context so it can be used for
	// authz decisions later
	rhost := C.GoString(cRhost)
	ctx = context.WithValue(ctx, rhostKey, rhost)

	pamSyslog(pamh, syslog.LOG_INFO, "connection from %s", rhost)
	// if localhost connection and configuration is to trust localhost
	// emmulate the pg_hba.conf setting
	// host all all 127.0.0.1/32 trust
	// host all all ::1/0 trust
	// if cfg.TrustLocal && (rhost == "127.0.0.1" || rhost == "::1") {
	// 	pamSyslog(pamh, syslog.LOG_INFO, "local connection trusted")
	// 	return c.PAM_SUCCESS
	// }

	// Get (or prompt for) user
	var cUser *C.char
	if errnum := C.pam_get_user(pamh, &cUser, nil); errnum != C.PAM_SUCCESS {
		pamSyslog(pamh, syslog.LOG_ERR, "failed to get user: %v", pamStrError(pamh, errnum))
		return errnum
	}

	user := C.GoString(cUser)
	if len(user) == 0 {
		pamSyslog(pamh, syslog.LOG_WARNING, "empty user")
		return C.PAM_USER_UNKNOWN
	}

	// Get (or prompt for) password (token)
	var cToken *C.char
	if errnum := C.pam_get_authtok(pamh, C.PAM_AUTHTOK, &cToken, nil); errnum != C.PAM_SUCCESS {
		pamSyslog(pamh, syslog.LOG_ERR, "failed to get token: %v", pamStrError(pamh, errnum))
		return errnum
	}
	token := C.GoString(cToken)

	// determine which authenticator to use
	auth, err := discoverAuthenticator(ctx, cfg, token)
	if err != nil {
		pamSyslog(pamh, syslog.LOG_ERR, "failed to discover authenticator: %v", err)
		return C.PAM_AUTH_ERR
	}

	// do the actual authentication and authorization
	// this will use the correct authenticator automatically, either password or PAT/JWT against an api
	if err := auth.Authenticate(ctx, user, token); err != nil {
		pamSyslog(pamh, syslog.LOG_WARNING, "failed to authenticate: %v", err)
		return C.PAM_AUTH_ERR
	}
	pamSyslog(pamh, syslog.LOG_INFO, "authenticated: %v", user)
	return C.PAM_SUCCESS
}

//export pam_sm_setcred_go
func pam_sm_setcred_go(pamh *C.pam_handle_t, flags C.int, argc C.int, argv **C.char) C.int {
	pamSyslog(pamh, syslog.LOG_WARNING, "SET CRED:")
	return C.PAM_IGNORE
}

//export pam_sm_open_session_go
func pam_sm_open_session_go(pamh *C.pam_handle_t, flags C.int, argc C.int, argv **C.char) C.int {
	pamSyslog(pamh, syslog.LOG_WARNING, "OPEN SESSION")
	return C.PAM_IGNORE
}

//export pam_sm_close_session_go
func pam_sm_close_session_go(pamh *C.pam_handle_t, flags C.int, argc C.int, argv **C.char) C.int {
	pamSyslog(pamh, syslog.LOG_WARNING, "CLOSE SESSION")
	return C.PAM_IGNORE
}

//export pam_sm_acct_mgmt_go
func pam_sm_acct_mgmt_go(pamh *C.pam_handle_t, flags C.int, argc C.int, argv **C.char) C.int {
	// From pam_sm_acct_mgmt
	// This function performs the task of establishing whether the user
	// is permitted to gain access at this time. It should be understood
	// that the user has previously been validated by an authentication module.
	// We assume that pam_sm_authenticate_go has been called and done
	// the authentication.
	// Because token expiration checks happen in pam_sm_acct_mgmt_go, we do not
	// need to do another expiration check here (there shouldn't be a multi-minute delay between calls)
	return C.PAM_SUCCESS
}

func pamStrError(pamh *C.pam_handle_t, errnum C.int) string {
	return C.GoString(C.pam_strerror(pamh, errnum))
}

func pamSyslog(pamh *C.pam_handle_t, priority syslog.Priority, format string, a ...interface{}) {
	cstr := C.CString(fmt.Sprintf(format, a...))
	defer C.free(unsafe.Pointer(cstr))

	C.pam_syslog_str(pamh, C.int(priority), cstr)
}
