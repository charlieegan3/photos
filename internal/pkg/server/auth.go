package server

import (
	"crypto/subtle"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/charlieegan3/photos/internal/pkg/constants"
)

func InitMiddlewareAuth(username, password string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, p, ok := r.BasicAuth()
			if !ok ||
				subtle.ConstantTimeCompare([]byte(username), []byte(u)) != 1 ||
				subtle.ConstantTimeCompare([]byte(password), []byte(p)) != 1 {
				w.Header().Set(
					"WWW-Authenticate",
					fmt.Sprintf(`Basic realm="%s"`, constants.Realm),
				)
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte("Unauthorised.\n"))

				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func InitMiddlewareEmailAuth(permittedEmailSuffix string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			email := r.Header.Get("X-Email")
			if email == "" {
				log.Println("X-Email header is missing")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte("not authenticated\n"))
				return
			}

			if permittedEmailSuffix == "" {
				log.Println("email suffix was blank and so no emails are allowed")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte("Unauthorized: no permitted email suffix configured\n"))
				return
			}

			if !strings.HasSuffix(email, permittedEmailSuffix) {
				log.Printf("email %s does not have suffix %s", email, permittedEmailSuffix)
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte("Forbidden: email not permitted\n"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
