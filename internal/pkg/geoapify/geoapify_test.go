package geoapify

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGeoapifyGeocoding(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		body, err := os.ReadFile("./response.json")
		if err != nil {
			t.Fatalf("unable to read file: %v", err)
		}
		_, _ = w.Write(body)
	}))
	defer ts.Close()

	client, err := NewClient(ts.URL, "api_key")
	require.NoError(t, err, "unexpected error")

	features, err := client.GeocodingSearch(t.Context(), "Highgate Hill, London")
	require.NoError(t, err, "unexpected error")

	require.Len(t, features, 5, "unexpected number of features")

	require.InEpsilon(t, 51.5691675, features[0].Properties.Lat, 0.0000001)
	require.InEpsilon(t, -0.1421935, features[0].Properties.Lon, 0.0000001)
	require.Equal(t, "Highgate Hill", features[0].Properties.Name)
	require.Equal(t, "Highgate Hill, London, N6 5XG, United Kingdom", features[0].Properties.Formatted)
	require.Equal(t, "street", features[0].Properties.ResultType)
}
