package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	t.Parallel()

	router := mux.NewRouter()
	adminRouter := router.PathPrefix("/admin").Subrouter()
	adminRouter.Use(InitMiddlewareAuth("username", "password"))

	router.HandleFunc("/public", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "public")
	})
	adminRouter.HandleFunc("/secret", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "secret")
	})

	// test the public request
	req, err := http.NewRequest(http.MethodGet, "/public", nil)
	assert.NoError(t, err, "unexpected error getting public page")

	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)

	body, err := io.ReadAll(rr.Body)
	assert.NoError(t, err, "unexpected error reading public page body")
	assert.Equal(t, "public", string(body))

	// test the admin request without credentials
	req, err = http.NewRequest(http.MethodGet, "/admin/secret", nil)
	assert.NoError(t, err, "unexpected error getting admin page")

	rr = httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Result().StatusCode)

	body, err = io.ReadAll(rr.Body)
	assert.NoError(t, err, "unexpected error reading private page body")
	assert.Equal(t, "Unauthorised.\n", string(body))

	// test the admin request with credentials set
	req, err = http.NewRequest(http.MethodGet, "/admin/secret", nil)
	req.SetBasicAuth("username", "password")
	assert.NoError(t, err, "unexpected error getting admin page")

	rr = httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)

	body, err = io.ReadAll(rr.Body)
	assert.NoError(t, err, "unexpected error reading private page body")
	assert.Equal(t, "secret", string(body))
}
