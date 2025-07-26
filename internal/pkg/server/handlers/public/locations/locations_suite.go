package public

import (
	"bytes"
	//nolint:gosec
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/charlieegan3/photos/internal/pkg/server/templating"

	"github.com/gorilla/mux"
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
	err = database.Truncate(s.T().Context(), s.DB, "photos.posts")
	s.Require().NoError(err)
	err = database.Truncate(s.T().Context(), s.DB, "photos.medias")
	s.Require().NoError(err)
	err = database.Truncate(s.T().Context(), s.DB, "photos.locations")
	s.Require().NoError(err)
	err = database.Truncate(s.T().Context(), s.DB, "photos.devices")
	s.Require().NoError(err)
}

func (s *LocationsSuite) TestLocationsMapIndex() {
	locations := []models.Location{
		{
			Name:      "London",
			Latitude:  1.1,
			Longitude: 1.2,
		},
	}

	_, err := database.CreateLocations(s.T().Context(), s.DB, locations)
	s.Require().NoError(err)

	router := mux.NewRouter()
	router.HandleFunc(
		"/locations",
		BuildIndexHandler(
			s.DB,
			"",
			templating.BuildPageRenderFunc(true, HeadContent),
		),
	).Methods(http.MethodGet)

	req, err := http.NewRequestWithContext(s.T().Context(), http.MethodGet, "/locations", nil)
	s.Require().NoError(err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if !s.Equal(http.StatusOK, rr.Code) {
		bodyString, err := io.ReadAll(rr.Body)
		s.Require().NoError(err)
		s.T().Fatalf("request failed with: %s", bodyString)
	}

	body, err := io.ReadAll(rr.Body)
	s.Require().NoError(err)

	s.Contains(string(body), `"London"`)
}

func (s *LocationsSuite) TestGetLocation() {
	devices := []models.Device{{Name: "Example Device"}}
	returnedDevices, err := database.CreateDevices(s.T().Context(), s.DB, devices)
	s.Require().NoError(err)

	medias := []models.Media{
		{DeviceID: returnedDevices[0].ID},
		{DeviceID: returnedDevices[0].ID},
	}
	returnedMedias, err := database.CreateMedias(s.T().Context(), s.DB, medias)
	s.Require().NoError(err)

	locations := []models.Location{
		{Name: "London"},
		{Name: "New York"},
	}
	returnedLocations, err := database.CreateLocations(s.T().Context(), s.DB, locations)
	s.Require().NoError(err)

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
	_, err = database.CreatePosts(s.T().Context(), s.DB, posts)
	s.Require().NoError(err)

	router := mux.NewRouter()
	router.HandleFunc(
		"/locations/{locationID}",
		BuildGetHandler(s.DB, templating.BuildPageRenderFunc(true, "")),
	).Methods(http.MethodGet)

	req, err := http.NewRequestWithContext(
		s.T().Context(), http.MethodGet, fmt.Sprintf("/locations/%d", returnedLocations[0].ID), nil,
	)
	s.Require().NoError(err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if !s.Equal(http.StatusOK, rr.Code) {
		bodyString, err := io.ReadAll(rr.Body)
		s.Require().NoError(err)
		s.T().Fatalf("request failed with: %s", bodyString)
	}

	body, err := io.ReadAll(rr.Body)
	s.Require().NoError(err)

	s.Contains(string(body), `London`)
	s.Contains(string(body), `post from london`)
	s.NotContains(string(body), `New York`)
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

	mapServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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

	returnedLocations, err := database.CreateLocations(s.T().Context(), s.DB, locations)
	s.Require().NoError(err)

	router := mux.NewRouter()
	router.HandleFunc("/locations/{locationID}/map.jpg",
		BuildMapHandler(s.DB, s.Bucket, mapServer.URL, "")).
		Methods(http.MethodGet)

	req, err := http.NewRequestWithContext(
		s.T().Context(), http.MethodGet, fmt.Sprintf("/locations/%d/map.jpg", returnedLocations[0].ID), nil,
	)
	s.Require().NoError(err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// we expect that the raw map is only requested once, since the second time it's in object storage
	s.Equal(1, requested)

	if !s.Equal(http.StatusOK, rr.Code) {
		bodyString, err := io.ReadAll(rr.Body)
		s.Require().NoError(err)
		s.T().Fatalf("request failed with: %s", bodyString)
	}

	body, err := io.ReadAll(rr.Body)
	s.Require().NoError(err)

	h = sha1.New()
	h.Write(body)
	bodySha := hex.EncodeToString(h.Sum(nil))
	s.Equal(bodySha, mapSha)
}
