package server

import (
	"crypto/subtle"
	"fmt"
	"net/http"

	"github.com/charlieegan3/photos/cms/internal/pkg/constants"
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
				w.WriteHeader(401)
				w.Write([]byte("Unauthorised.\n"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
