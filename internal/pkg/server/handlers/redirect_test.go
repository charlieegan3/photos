package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/require"
)

func TestBuildRedirectHandler(t *testing.T) {
	t.Parallel()

	router := mux.NewRouter()
	router.HandleFunc("/admin/", BuildRedirectHandler("/admin")).Methods(http.MethodGet)

	req, err := http.NewRequest(http.MethodGet, "/admin/", nil)
	require.NoError(t, err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusSeeOther, rr.Code)

	td.Cmp(t, rr.Result().Header["Location"], []string{"/admin"})
}
