package points

import (
	"cloud.google.com/go/bigquery"
	"context"
	"database/sql"
	"fmt"
	"github.com/charlieegan3/photos/internal/pkg/models"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/memblob"
	"google.golang.org/api/option"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/charlieegan3/photos/internal/pkg/database"
)

type EndpointsPointsSuite struct {
	suite.Suite
	DB            *sql.DB
	Bucket        *blob.Bucket
	BucketBaseURL string
}

func (s *EndpointsPointsSuite) SetupTest() {
	var err error
	err = database.Truncate(s.DB, "locations.points")
	require.NoError(s.T(), err)
}

func (s *EndpointsPointsSuite) TestPeriodGPXHandler() {
	points := []models.Point{
		{
			Latitude:  3.0,
			Longitude: 4.0,
			Altitude:  1.0,
			CreatedAt: time.Date(1994, 4, 23, 13, 22, 0, 0, time.UTC),
		},
		{
			Latitude:  3.0,
			Longitude: 4.0,
			Altitude:  2.0,
			CreatedAt: time.Date(2021, 4, 23, 13, 22, 0, 0, time.UTC),
		},
		{
			Latitude:  1.0,
			Longitude: 2.0,
			Altitude:  3.0,
			CreatedAt: time.Date(2021, 5, 23, 13, 19, 0, 0, time.UTC),
		},
		{
			Latitude:  3.0,
			Longitude: 4.0,
			Altitude:  4.0,
			CreatedAt: time.Date(2022, 4, 23, 13, 22, 0, 0, time.UTC),
		},
	}

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := fmt.Sprintf(`
{
  "kind": "result",
  "schema": {
	"fields": [
  { "name": "created_at", "type": "TIMESTAMP" },
  { "name": "latitude", "type": "FLOAT" },
  { "name": "longitude", "type": "FLOAT" },
  { "name": "altitude", "type": "FLOAT" },
  { "name": "accuracy", "type": "FLOAT" },
  { "name": "vertical_accuracy", "type": "FLOAT" },
  { "name": "velocity", "type": "FLOAT" },
  { "name": "was_offline", "type": "BOOLEAN" },
  { "name": "importer_id", "type": "INTEGER" },
  { "name": "caller_id", "type": "INTEGER" },
  { "name": "reason_id", "type": "INTEGER" },
  { "mode": "NULLABLE", "name": "activity_id", "type": "INTEGER" }
]
  },
  "jobReference": {
    "projectId": "example",
    "jobId": "example",
    "location":  "example"	
  },
  "totalRows": "3",
  "pageToken": "",
  "rows": [
    {
      "f": [
        { "v" : "%d" },
        { "v" : "%f" },
        { "v" : "%f" },
        { "v" : "%f" },
        { "v" : "0.0" },
        { "v" : "0.0" },
        { "v" : "0.0" },
        { "v" : "false" },
        { "v" : "0" },
        { "v" : "0" },
        { "v" : "0" },
        { "v" : null }
      ]
    },
    {
      "f": [
        { "v" : "%d" },
        { "v" : "%f" },
        { "v" : "%f" },
        { "v" : "%f" },
        { "v" : "0.0" },
        { "v" : "0.0" },
        { "v" : "0.0" },
        { "v" : "false" },
        { "v" : "0" },
        { "v" : "0" },
        { "v" : "0" },
        { "v" : null }
      ]
    },
    {
      "f": [
        { "v" : "%d" },
        { "v" : "%f" },
        { "v" : "%f" },
        { "v" : "%f" },
        { "v" : "0.0" },
        { "v" : "0.0" },
        { "v" : "0.0" },
        { "v" : "false" },
        { "v" : "0" },
        { "v" : "0" },
        { "v" : "0" },
        { "v" : null }
      ]
    }
  ],
  "totalBytesProcessed": "100",
  "jobComplete": true,
  "errors": [],
  "cacheHit": false
}`,
			points[1].CreatedAt.Unix(),
			points[1].Latitude,
			points[1].Longitude,
			points[1].Altitude,
			points[2].CreatedAt.Unix(),
			points[2].Latitude,
			points[2].Longitude,
			points[2].Altitude,
			points[3].CreatedAt.Unix(),
			points[3].Latitude,
			points[3].Longitude,
			points[3].Altitude,
		)

		w.Write([]byte(response))
	}))

	client, err := bigquery.NewClient(
		context.Background(),
		"test-project",
		option.WithEndpoint(testServer.URL),
		option.WithoutAuthentication(),
	)
	require.NoError(s.T(), err)

	router := mux.NewRouter()
	router.HandleFunc(
		"/admin/points/period/gpx",
		BuildPeriodGPXHandler(client, "dataset", "table"),
	).Methods("GET")

	req, err := http.NewRequest("GET", "/admin/points/period/gpx?from=2021-01-01&to=2022-12-31", nil)
	require.NoError(s.T(), err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if !assert.Equal(s.T(), http.StatusOK, rr.Code) {
		bodyString, err := io.ReadAll(rr.Body)
		require.NoError(s.T(), err)
		s.T().Fatalf("request failed with: %s", bodyString)
	}

	body, err := io.ReadAll(rr.Body)
	require.NoError(s.T(), err)

	expectedGPX := `<?xml version="1.0" encoding="UTF-8"?>
<gpx xmlns="http://www.topografix.com/GPX/1/0" version="1.0" creator="photos.charlieegan3.com">
	<trk>
		<name>2021-01-01 to 2022-12-31</name>
		<trkseg>
			<trkpt lat="3" lon="4">
				<ele>2</ele>
				<time>2021-04-23T13:22:00Z</time>
				<Speed></Speed>
			</trkpt>
			<trkpt lat="1" lon="2">
				<ele>3</ele>
				<time>2021-05-23T13:19:00Z</time>
				<Speed></Speed>
			</trkpt>
			<trkpt lat="3" lon="4">
				<ele>4</ele>
				<time>2022-04-23T13:22:00Z</time>
				<Speed></Speed>
			</trkpt>
		</trkseg>
	</trk>
</gpx>`

	assert.Equal(s.T(), expectedGPX, string(body))
}
