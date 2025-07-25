package public

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/charlieegan3/photos/internal/pkg/server/templating"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/memblob"

	"github.com/charlieegan3/photos/internal/pkg/database"
	"github.com/charlieegan3/photos/internal/pkg/models"
)

type TripsSuite struct {
	suite.Suite
	DB *sql.DB
}

func (s *TripsSuite) SetupTest() {
	var err error
	err = database.Truncate(s.DB, "photos.posts")
	require.NoError(s.T(), err)
	err = database.Truncate(s.DB, "photos.medias")
	require.NoError(s.T(), err)
	err = database.Truncate(s.DB, "photos.locations")
	require.NoError(s.T(), err)
	err = database.Truncate(s.DB, "photos.devices")
	require.NoError(s.T(), err)
	err = database.Truncate(s.DB, "photos.trips")
	require.NoError(s.T(), err)
}

func (s *TripsSuite) TestGetTrip() {
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

	trips := []models.Trip{
		{
			Title:     "London",
			StartDate: time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC),
			EndDate:   time.Date(2021, time.January, 2, 0, 0, 0, 0, time.UTC),
		},
		{Title: "New York"},
	}
	returnedTrips, err := database.CreateTrips(s.DB, trips)
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
			PublishDate: time.Date(2021, time.January, 3, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[1].ID,
			LocationID:  returnedLocations[1].ID,
		},
	}
	_, err = database.CreatePosts(s.DB, posts)
	require.NoError(s.T(), err)

	router := mux.NewRouter()
	router.HandleFunc(
		"/trips/{tripID}",
		BuildGetHandler(s.DB, templating.BuildPageRenderFunc(true, "")),
	).Methods(http.MethodGet)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/trips/%d", returnedTrips[0].ID), nil)
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
