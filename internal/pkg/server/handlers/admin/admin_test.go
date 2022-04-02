package admin

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/charlieegan3/photos/cms/internal/pkg/server/templating"
)

func TestIndexPage(t *testing.T) {
	router := mux.NewRouter()

	renderer := templating.BuildPageRenderFunc(true, "", "admin")

	router.HandleFunc("/admin", BuildAdminIndexHandler(renderer)).Methods("GET")

	req, err := http.NewRequest("GET", "/admin", nil)
	require.NoError(t, err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	body, err := ioutil.ReadAll(rr.Body)
	require.NoError(t, err)

	// items from the shared header
	assert.Contains(t, string(body), "Posts")
	// items from the page
	assert.Contains(t, string(body), "Manage...")
}
