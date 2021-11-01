package server

import "net/http"

type notFoundIntercept struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

func (nf *notFoundIntercept) WriteHeader(code int) {
	nf.statusCode = code
	nf.ResponseWriter.WriteHeader(code)
}

func (nf *notFoundIntercept) Write(b []byte) (int, error) {
	return nf.ResponseWriter.Write(b)
}

func InitMiddleware404() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			nf := &notFoundIntercept{w, http.StatusOK, []byte{}}
			next.ServeHTTP(nf, r)

			if nf.statusCode == http.StatusNotFound {
				nf.ResponseWriter.Write([]byte("Not found"))
			}
		})
	}
}
