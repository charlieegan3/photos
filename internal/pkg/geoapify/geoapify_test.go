package geoapify

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGeoapifyGeocoding(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadFile("./response.json")
		if err != nil {
			t.Fatalf("unable to read file: %v", err)
		}
		w.Write(body)
	}))
	defer ts.Close()

	client, err := NewClient(ts.URL, "api_key")
	require.NoError(t, err, "unexpected error")

	features, err := client.GeocodingSearch("Highgate Hill, London")
	require.NoError(t, err, "unexpected error")

	require.Equal(t, 5, len(features), "unexpected number of features")

	require.Equal(t, 51.5691675, features[0].Properties.Lat)
	require.Equal(t, -0.1421935, features[0].Properties.Lon)
	require.Equal(t, "Highgate Hill", features[0].Properties.Name)
	require.Equal(t, "Highgate Hill, London, N6 5XG, United Kingdom", features[0].Properties.Formatted)
	require.Equal(t, "street", features[0].Properties.ResultType)
}
