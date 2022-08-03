package points

import (
	"database/sql"
	"github.com/charlieegan3/photos/cms/internal/pkg/models"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/memblob"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/charlieegan3/photos/cms/internal/pkg/database"
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
			CreatedAt: time.Date(1994, 4, 23, 13, 22, 0, 0, time.UTC),
		},
		{
			Latitude:  3.0,
			Longitude: 4.0,
			CreatedAt: time.Date(2021, 4, 23, 13, 22, 0, 0, time.UTC),
		},
		{
			Latitude:  1.0,
			Longitude: 2.0,
			CreatedAt: time.Date(2021, 5, 23, 13, 19, 0, 0, time.UTC),
		},
		{
			Latitude:  3.0,
			Longitude: 4.0,
			CreatedAt: time.Date(2022, 4, 23, 13, 22, 0, 0, time.UTC),
		},
	}

	_, err := database.CreatePoints(
		s.DB,
		"example_importer",
		"example_caller",
		"example_reason",
		nil, // no activity set
		points,
	)
	require.NoError(s.T(), err)

	router := mux.NewRouter()
	router.HandleFunc("/admin/points/period/gpx", BuildPeriodGPXHandler(s.DB)).Methods("GET")

	req, err := http.NewRequest("GET", "/admin/points/period/gpx?from=2021-01-01&to=2022-12-31", nil)
	require.NoError(s.T(), err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if !assert.Equal(s.T(), http.StatusOK, rr.Code) {
		bodyString, err := ioutil.ReadAll(rr.Body)
		require.NoError(s.T(), err)
		s.T().Fatalf("request failed with: %s", bodyString)
	}

	body, err := ioutil.ReadAll(rr.Body)
	require.NoError(s.T(), err)

	expectedGPX := `<?xml version="1.0" encoding="UTF-8"?>
<gpx xmlns="http://www.topografix.com/GPX/1/0" version="1.0" creator="photos.charlieegan3.com">
	<trk>
		<name>2021-01-01 to 2022-12-31</name>
		<trkseg>
			<trkpt lat="3" lon="4">
				<time>2021-04-23T13:22:00Z</time>
				<Speed></Speed>
			</trkpt>
			<trkpt lat="1" lon="2">
				<time>2021-05-23T13:19:00Z</time>
				<Speed></Speed>
			</trkpt>
			<trkpt lat="3" lon="4">
				<time>2022-04-23T13:22:00Z</time>
				<Speed></Speed>
			</trkpt>
		</trkseg>
	</trk>
</gpx>`

	assert.Equal(s.T(), expectedGPX, string(body))
}
