package server

import (
	"net/http"
	"net/url"
)

func InitMiddlewareHTTPS(hostname, environment string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if environment == "production" {
				if r.Header.Get("X-Forwarded-Proto") != "https" {
					newURL, err := url.Parse(r.URL.String())
					if err != nil {
						w.WriteHeader(http.StatusBadRequest)
						w.Write([]byte(err.Error()))
						return
					}
					newURL.Host = hostname + ":443"
					newURL.Scheme = "https"

					w.Header().Set("Strict-Transport-Security", "max-age=3600")

					http.Redirect(w, r, newURL.String(), http.StatusPermanentRedirect)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
