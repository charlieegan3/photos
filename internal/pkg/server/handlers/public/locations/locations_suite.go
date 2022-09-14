package public

import (
	"bytes"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/charlieegan3/photos/internal/pkg/server/templating"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/memblob"

	"github.com/charlieegan3/photos/internal/pkg/database"
	"github.com/charlieegan3/photos/internal/pkg/models"
)

type LocationsSuite struct {
	suite.Suite
	DB            *sql.DB
	Bucket        *blob.Bucket
	BucketBaseURL string
}

func (s *LocationsSuite) SetupTest() {
	var err error
	err = database.Truncate(s.DB, "posts")
	require.NoError(s.T(), err)
	err = database.Truncate(s.DB, "medias")
	require.NoError(s.T(), err)
	err = database.Truncate(s.DB, "locations")
	require.NoError(s.T(), err)
	err = database.Truncate(s.DB, "devices")
	require.NoError(s.T(), err)
}

func (s *LocationsSuite) TestLocationsMapIndex() {
	locations := []models.Location{
		{
			Name:      "London",
			Latitude:  1.1,
			Longitude: 1.2,
		},
	}

	_, err := database.CreateLocations(s.DB, locations)
	require.NoError(s.T(), err)

	router := mux.NewRouter()
	router.HandleFunc(
		"/locations",
		BuildIndexHandler(
			s.DB,
			"",
			templating.BuildPageRenderFunc(true, HeadContent),
		),
	).Methods("GET")

	req, err := http.NewRequest("GET", "/locations", nil)
	require.NoError(s.T(), err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if !assert.Equal(s.T(), http.StatusOK, rr.Code) {
		bodyString, err := io.ReadAll(rr.Body)
		require.NoError(s.T(), err)
		s.T().Fatalf("request failed with: %s", bodyString)
	}

	body, err := io.ReadAll(rr.Body)
	require.NoError(s.T(), err)

	assert.Contains(s.T(), string(body), `"London"`)
}

func (s *LocationsSuite) TestGetLocation() {
	devices := []models.Device{{Name: "Example Device"}}
	returnedDevices, err := database.CreateDevices(s.DB, devices)
	require.NoError(s.T(), err)

	medias := []models.Media{
		{DeviceID: returnedDevices[0].ID},
		{DeviceID: returnedDevices[0].ID},
	}
	returnedMedias, err := database.CreateMedias(s.DB, medias)
	require.NoError(s.T(), err)

	locations := []models.Location{
		{Name: "London"},
		{Name: "New York"},
	}
	returnedLocations, err := database.CreateLocations(s.DB, locations)
	require.NoError(s.T(), err)

	posts := []models.Post{
		{
			Description: "post from london",
			PublishDate: time.Date(2021, time.January, 1, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
		},
		{
			Description: "post from new york",
			PublishDate: time.Date(2021, time.January, 1, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[1].ID,
			LocationID:  returnedLocations[1].ID,
		},
	}
	_, err = database.CreatePosts(s.DB, posts)
	require.NoError(s.T(), err)

	router := mux.NewRouter()
	router.HandleFunc(
		"/locations/{locationID}",
		BuildGetHandler(s.DB, templating.BuildPageRenderFunc(true, "")),
	).Methods("GET")

	req, err := http.NewRequest("GET", fmt.Sprintf("/locations/%d", returnedLocations[0].ID), nil)
	require.NoError(s.T(), err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if !assert.Equal(s.T(), http.StatusOK, rr.Code) {
		bodyString, err := io.ReadAll(rr.Body)
		require.NoError(s.T(), err)
		s.T().Fatalf("request failed with: %s", bodyString)
	}

	body, err := io.ReadAll(rr.Body)
	require.NoError(s.T(), err)

	assert.Contains(s.T(), string(body), `London`)
	assert.Contains(s.T(), string(body), `post from london`)
	assert.NotContains(s.T(), string(body), `New York`)
}

func (s *LocationsSuite) TestGetLocationMap() {
	var requested int
	mapBytes, err := os.ReadFile("../../server/handlers/public/locations/fixtures/map.jpg")
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

	if !assert.Equal(s.T(), http.StatusOK, rr.Code) {
		bodyString, err := io.ReadAll(rr.Body)
		require.NoError(s.T(), err)
		s.T().Fatalf("request failed with: %s", bodyString)
	}

	body, err := io.ReadAll(rr.Body)
	require.NoError(s.T(), err)

	h = sha1.New()
	h.Write(body)
	bodySha := hex.EncodeToString(h.Sum(nil))
	assert.Equal(s.T(), bodySha, mapSha)
}
