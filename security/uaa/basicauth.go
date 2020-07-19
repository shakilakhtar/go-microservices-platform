package uaa

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"

	dtlogger "github.com/sirupsen/logrus"
)

var (
	// ErrInvalidCredFormat is returned when credential format is incorrect
	ErrInvalidCredFormat = errors.New("credentials not provided in the correct format")

	// ErrInvalidCreds is returned when the credential is incorrect
	ErrInvalidCreds = errors.New("incorrect credentials provided")
)

// BasicAuthenticator represents a type for handling basic authentication for http handlers.
type BasicAuthenticator struct {
	username string
	password string
}

// NewBasicAuthenticator creates a BasicAuthHandler.
func NewBasicAuthenticator(username, password string) BasicAuthenticator {
	return BasicAuthenticator{
		username: username,
		password: password,
	}
}

// Wrap will wrap around a handler passed in and enforces authentication on top.
func (ba BasicAuthenticator) Wrap(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()

		if !ok {
			var strRequest string
			dumpRequest, err := httputil.DumpRequest(r, true)
			if err == nil {
				strRequest = string(dumpRequest)
			}
			errStr := fmt.Sprint("credentials error", " error:", ErrInvalidCredFormat, " request:", strRequest)
			w.Header().Set("WWW-Authenticate", "Basic realm=Authorization Required")
			w.WriteHeader(http.StatusUnauthorized)
			dtlogger.Error(errStr)
			//http.Error(w, errStr, http.StatusUnauthorized)
			return
		}

		if username != ba.username || password != ba.password {
			var strRequest string
			dumpRequest, err := httputil.DumpRequest(r, true)
			if err == nil {
				strRequest = string(dumpRequest)
			}
			errStr := fmt.Sprint("credentials error", " error:", ErrInvalidCreds, " request:", strRequest)
			w.WriteHeader(http.StatusUnauthorized)
			dtlogger.Error(errStr)
			//http.Error(w, errStr, http.StatusUnauthorized)
			return
		}

		handler.ServeHTTP(w, r)
	})
}
