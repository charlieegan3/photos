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
	fmt.Println(code)
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

			lw := &loggingResponseWriter{w, http.StatusOK, []byte{}}

			next.ServeHTTP(lw, r)

			entry = entry.WithFields(logrus.Fields{
				"status": lw.statusCode,
				"path":   r.URL.Path,
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
