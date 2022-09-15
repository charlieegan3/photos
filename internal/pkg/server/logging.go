package server

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (lw *loggingResponseWriter) WriteHeader(code int) {
	lw.statusCode = code
	lw.ResponseWriter.WriteHeader(code)
}

func (lw *loggingResponseWriter) Write(b []byte) (int, error) {
	if lw.statusCode >= 400 {
		lw.body = append(lw.body, b...)
	}
	return lw.ResponseWriter.Write(b)
}

func InitMiddlewareLogging() func(http.Handler) http.Handler {
	logger := logrus.New()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			entry := logrus.NewEntry(logger)

			if r == nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("request to logging middleware was nil"))
				return
			}

			lw := &loggingResponseWriter{w, http.StatusOK, []byte{}}

			method := r.Method
			if val, ok := r.Header["Content-Type"]; ok && val[0] == "application/x-www-form-urlencoded" {
				err := r.ParseForm()
				if err == nil {
					if formMethod := r.PostForm.Get("_method"); formMethod != "" {
						method = fmt.Sprintf("%s (%s)", formMethod, r.Method)
					}
				}
			}

			next.ServeHTTP(lw, r)

			path := r.URL.Path
			if len(r.URL.RawQuery) > 0 {
				path += "?" + r.URL.RawQuery
			}

			entry = entry.WithFields(logrus.Fields{
				"status": lw.statusCode,
				"path":   path,
				"method": method,
			})

			switch {
			case lw.statusCode > 0 && lw.statusCode < 400:
				entry.Info()
			case lw.statusCode >= 400 && lw.statusCode < 500:
				entry.Warn(string(lw.body))
			case lw.statusCode >= 500:
				entry.Error(string(lw.body))
			default:
				entry.Warnf("unknown code: %d", lw.statusCode)
			}
		})
	}
}
