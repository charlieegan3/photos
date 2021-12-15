package public

import (
	"bytes"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/memblob"

	"github.com/charlieegan3/photos/cms/internal/pkg/database"
	"github.com/charlieegan3/photos/cms/internal/pkg/models"
)

type LocationsSuite struct {
	suite.Suite
	DB            *sql.DB
	Bucket        *blob.Bucket
	BucketBaseURL string
}

func (s *LocationsSuite) SetupTest() {
	err := database.Truncate(s.DB, "locations")
	require.NoError(s.T(), err)
}

func (s *LocationsSuite) TestGetLocationMap() {
	var requested int
	mapBytes, err := ioutil.ReadFile("../../server/handlers/public/locations/fixtures/map.jpg")
	if err != nil {
		s.T().Fatal(err)
	}
	h := sha1.New()
	h.Write(mapBytes)
	mapSha := hex.EncodeToString(h.Sum(nil))

	mapServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err = io.Copy(w, bytes.NewReader(mapBytes))
		if err != nil {
			s.T().Error(err)
		}

		requested += 1
	}))

	locations := []models.Location{
		{
			Name:      "London",
			Latitude:  1.1,
			Longitude: 1.2,
		},
	}

	returnedLocations, err := database.CreateLocations(s.DB, locations)
	require.NoError(s.T(), err)

	router := mux.NewRouter()
	router.HandleFunc("/locations/{locationID}/map.jpg", BuildMapHandler(s.DB, s.Bucket, mapServer.URL, "")).Methods("GET")

	req, err := http.NewRequest("GET", fmt.Sprintf("/locations/%d/map.jpg", returnedLocations[0].ID), nil)
	require.NoError(s.T(), err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// we expect that the raw map is only requested once, since the second time it's in object storage
	assert.Equal(s.T(), 1, requested)

	require.Equal(s.T(), http.StatusOK, rr.Code)

	body, err := ioutil.ReadAll(rr.Body)
	require.NoError(s.T(), err)

	h = sha1.New()
	h.Write(body)
	bodySha := hex.EncodeToString(h.Sum(nil))
	assert.Equal(s.T(), bodySha, mapSha)
}
