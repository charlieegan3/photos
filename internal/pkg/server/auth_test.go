package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware(t *testing.T) {
	t.Parallel()

	router := mux.NewRouter()
	adminRouter := router.PathPrefix("/admin").Subrouter()
	adminRouter.Use(InitMiddlewareAuth("username", "password"))

	router.HandleFunc("/public", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "public")
	})
	adminRouter.HandleFunc("/secret", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "secret")
	})

	// test the public request
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/public", nil)
	require.NoError(t, err, "unexpected error getting public page")

	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)

	body, err := io.ReadAll(rr.Body)
	require.NoError(t, err, "unexpected error reading public page body")
	assert.Equal(t, "public", string(body))

	// test the admin request without credentials
	req, err = http.NewRequestWithContext(t.Context(), http.MethodGet, "/admin/secret", nil)
	require.NoError(t, err, "unexpected error getting admin page")

	rr = httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Result().StatusCode)

	body, err = io.ReadAll(rr.Body)
	require.NoError(t, err, "unexpected error reading private page body")
	assert.Equal(t, "Unauthorised.\n", string(body))

	// test the admin request with credentials set
	req, err = http.NewRequestWithContext(t.Context(), http.MethodGet, "/admin/secret", nil)
	req.SetBasicAuth("username", "password")
	require.NoError(t, err, "unexpected error getting admin page")

	rr = httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)

	body, err = io.ReadAll(rr.Body)
	require.NoError(t, err, "unexpected error reading private page body")
	assert.Equal(t, "secret", string(body))
}

func TestEmailAuthMiddleware(t *testing.T) {
	t.Parallel()

	router := mux.NewRouter()
	adminRouter := router.PathPrefix("/admin").Subrouter()
	adminRouter.Use(InitMiddlewareEmailAuth("@charlieegan3.com"))

	router.HandleFunc("/public", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "public")
	})
	adminRouter.HandleFunc("/secret", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "secret")
	})

	// test the public request
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/public", nil)
	require.NoError(t, err, "unexpected error getting public page")

	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)

	body, err := io.ReadAll(rr.Body)
	require.NoError(t, err, "unexpected error reading public page body")
	assert.Equal(t, "public", string(body))

	// test the admin request without email header
	req, err = http.NewRequestWithContext(t.Context(), http.MethodGet, "/admin/secret", nil)
	require.NoError(t, err, "unexpected error getting admin page")

	rr = httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Result().StatusCode)

	body, err = io.ReadAll(rr.Body)
	require.NoError(t, err, "unexpected error reading private page body")
	assert.Equal(t, "not authenticated\n", string(body))

	// test the admin request with wrong email suffix
	req, err = http.NewRequestWithContext(t.Context(), http.MethodGet, "/admin/secret", nil)
	req.Header.Set("X-Email", "user@wrongdomain.com")
	require.NoError(t, err, "unexpected error getting admin page")

	rr = httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Result().StatusCode)

	body, err = io.ReadAll(rr.Body)
	require.NoError(t, err, "unexpected error reading private page body")
	assert.Equal(t, "Forbidden: email not permitted\n", string(body))

	// test the admin request with correct email
	req, err = http.NewRequestWithContext(t.Context(), http.MethodGet, "/admin/secret", nil)
	req.Header.Set("X-Email", "user@charlieegan3.com")
	require.NoError(t, err, "unexpected error getting admin page")

	rr = httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Result().StatusCode)

	body, err = io.ReadAll(rr.Body)
	require.NoError(t, err, "unexpected error reading private page body")
	assert.Equal(t, "secret", string(body))

	// test with empty permitted suffix
	router2 := mux.NewRouter()
	adminRouter2 := router2.PathPrefix("/admin").Subrouter()
	adminRouter2.Use(InitMiddlewareEmailAuth(""))

	adminRouter2.HandleFunc("/secret", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "secret")
	})

	req, err = http.NewRequestWithContext(t.Context(), http.MethodGet, "/admin/secret", nil)
	req.Header.Set("X-Email", "user@charlieegan3.com")
	require.NoError(t, err, "unexpected error getting admin page")

	rr = httptest.NewRecorder()

	router2.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Result().StatusCode)

	body, err = io.ReadAll(rr.Body)
	require.NoError(t, err, "unexpected error reading private page body")
	assert.Equal(t, "Unauthorized: no permitted email suffix configured\n", string(body))
}
