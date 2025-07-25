package admin

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/charlieegan3/photos/internal/pkg/server/templating"
)

func TestIndexPage(t *testing.T) {
	t.Parallel()

	router := mux.NewRouter()

	renderer := templating.BuildPageRenderFunc(true, "", "admin")

	router.HandleFunc("/admin", BuildAdminIndexHandler(renderer)).Methods(http.MethodGet)

	req, err := http.NewRequest(http.MethodGet, "/admin", nil)
	require.NoError(t, err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	body, err := io.ReadAll(rr.Body)
	require.NoError(t, err)

	// items from the shared header
	assert.Contains(t, string(body), "Posts")
}
