package points

import (
	"database/sql"
	"github.com/charlieegan3/photos/cms/internal/pkg/database"
	"github.com/doug-martin/goqu/v9"
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
	"strings"
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
	err = database.Truncate(s.DB, "locations.reasons")
	require.NoError(s.T(), err)
	err = database.Truncate(s.DB, "locations.activities")
	require.NoError(s.T(), err)
	err = database.Truncate(s.DB, "locations.callers")
	require.NoError(s.T(), err)
	err = database.Truncate(s.DB, "locations.importers")
	require.NoError(s.T(), err)
}

func (s *EndpointsPointsSuite) TestOwnTracksCreatePoint() {
	router := mux.NewRouter()
	router.HandleFunc("/private/points", BuildOwnTracksEndpointHandler(s.DB)).Methods("POST")

	// make the request to the handler
	req, err := http.NewRequest(
		"POST",
		"/private/points",
		strings.NewReader(`{
        "_type": "location",
        "lat": 51.567312,
        "lon": -0.138803,
        "batt": 41,
        "acc": 7,
        "bs": 1,
        "p": 101.311,
        "vel": 1,
        "BSSID": "xx:xx:xx:xx:xx:xx",
        "SSID": "Home",
        "vac": 1,
        "topic": "owntracks/user/iphone",
        "conn": "w",
        "m": 2,
        "tst": 1651412865,
        "alt": 75,
        "tid": "NE"
    }`))
	require.NoError(s.T(), err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic dXNlcjpoZXk=")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if !assert.Equal(s.T(), http.StatusOK, rr.Code) {
		bodyString, err := ioutil.ReadAll(rr.Body)
		require.NoError(s.T(), err)
		s.T().Fatalf("request failed with: %s", bodyString)
	}

	// TODO impl a means of selecting points in database
	goquDB := goqu.New("postgres", s.DB)
	selectPoints := goquDB.From("locations.points").
		Select("latitude").
		Executor()

	var lat float64
	_, err = selectPoints.ScanVal(&lat)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), 51.567312, lat, "point value")
}
