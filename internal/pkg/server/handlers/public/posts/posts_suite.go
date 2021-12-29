package public

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/memblob"

	"github.com/charlieegan3/photos/cms/internal/pkg/database"
	"github.com/charlieegan3/photos/cms/internal/pkg/models"
	"github.com/charlieegan3/photos/cms/internal/pkg/server/templating"
)

type PostsSuite struct {
	suite.Suite
	DB            *sql.DB
	Bucket        *blob.Bucket
	BucketBaseURL string
}

func (s *PostsSuite) SetupTest() {
	err := database.Truncate(s.DB, "posts")
	require.NoError(s.T(), err)
	err = database.Truncate(s.DB, "devices")
	require.NoError(s.T(), err)
	err = database.Truncate(s.DB, "locations")
	require.NoError(s.T(), err)
	err = database.Truncate(s.DB, "medias")
	require.NoError(s.T(), err)
	err = database.Truncate(s.DB, "tags")
	require.NoError(s.T(), err)
	err = database.Truncate(s.DB, "taggings")
	require.NoError(s.T(), err)
}

func (s *PostsSuite) TestListPosts() {
	devices := []models.Device{
		{
			Name: "Example Device",
		},
	}
	returnedDevices, err := database.CreateDevices(s.DB, devices)
	require.NoError(s.T(), err)

	medias := []models.Media{
		{
			DeviceID: returnedDevices[0].ID,

			Make:  "FujiFilm",
			Model: "X100F",

			TakenAt: time.Date(2021, time.November, 23, 19, 56, 0, 0, time.UTC),

			FNumber:  2.0,
			ISOSpeed: 100,

			Latitude:  51.1,
			Longitude: 52.2,
			Altitude:  100.0,
		},
	}
	returnedMedias, err := database.CreateMedias(s.DB, medias)
	require.NoError(s.T(), err)
	locations := []models.Location{
		{
			Name:      "London",
			Latitude:  1.1,
			Longitude: 1.2,
		},
	}

	returnedLocations, err := database.CreateLocations(s.DB, locations)
	require.NoError(s.T(), err)

	posts := []models.Post{
		{
			Description: "Here is a shot I took",
			PublishDate: time.Date(2021, time.November, 24, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
		},
		{
			Description: "Here is another photo I took, same but diff",
			PublishDate: time.Date(2021, time.November, 25, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
		},
	}

	_, err = database.CreatePosts(s.DB, posts)
	require.NoError(s.T(), err)

	router := mux.NewRouter()
	router.HandleFunc("/", BuildIndexHandler(s.DB, templating.BuildPageRenderFunc())).Methods("GET")

	req, err := http.NewRequest("GET", "/", nil)
	require.NoError(s.T(), err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(s.T(), http.StatusOK, rr.Code)

	body, err := ioutil.ReadAll(rr.Body)
	require.NoError(s.T(), err)

	assert.Contains(s.T(), string(body), "shot")
	assert.Contains(s.T(), string(body), "another photo")
}

func (s *PostsSuite) TestGetPost() {
	devices := []models.Device{
		{
			Name: "Example Device",
		},
	}
	returnedDevices, err := database.CreateDevices(s.DB, devices)
	require.NoError(s.T(), err)

	medias := []models.Media{
		{
			DeviceID: returnedDevices[0].ID,

			Make:  "FujiFilm",
			Model: "X100F",

			TakenAt: time.Date(2021, time.November, 23, 19, 56, 0, 0, time.UTC),

			FNumber:  2.0,
			ISOSpeed: 100,

			Latitude:  51.1,
			Longitude: 52.2,
			Altitude:  100.0,
		},
	}
	returnedMedias, err := database.CreateMedias(s.DB, medias)
	require.NoError(s.T(), err)
	locations := []models.Location{
		{
			Name:      "London",
			Latitude:  1.1,
			Longitude: 1.2,
		},
	}

	returnedLocations, err := database.CreateLocations(s.DB, locations)
	require.NoError(s.T(), err)

	posts := []models.Post{
		{
			Description: "Here is a shot I took",
			PublishDate: time.Date(2021, time.November, 24, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
		},
		{
			Description: "Here is another photo I took, same but diff",
			PublishDate: time.Date(2021, time.November, 25, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
		},
	}

	persistedPosts, err := database.CreatePosts(s.DB, posts)
	require.NoError(s.T(), err)

	router := mux.NewRouter()
	router.HandleFunc("/posts/{postID}", BuildGetHandler(s.DB, templating.BuildPageRenderFunc())).Methods("GET")

	req, err := http.NewRequest("GET", fmt.Sprintf("/posts/%d", persistedPosts[0].ID), nil)
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

	assert.Contains(s.T(), string(body), "Here is a shot I took")
	assert.NotContains(s.T(), string(body), "another photo")
}

func (s *PostsSuite) TestPeriodHandler() {
	devices := []models.Device{{Name: "Example Device"}}
	returnedDevices, err := database.CreateDevices(s.DB, devices)
	require.NoError(s.T(), err)

	medias := []models.Media{{DeviceID: returnedDevices[0].ID}}
	returnedMedias, err := database.CreateMedias(s.DB, medias)
	require.NoError(s.T(), err)

	locations := []models.Location{{Name: "London", Latitude: 1.1, Longitude: 1.2}}
	returnedLocations, err := database.CreateLocations(s.DB, locations)
	require.NoError(s.T(), err)

	posts := []models.Post{
		{
			Description: "older post",
			PublishDate: time.Date(2021, time.October, 24, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
		},
		{
			Description: "post in range",
			PublishDate: time.Date(2021, time.November, 24, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
		},
		{
			Description: "future post",
			PublishDate: time.Date(2021, time.December, 25, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
		},
	}

	_, err = database.CreatePosts(s.DB, posts)
	require.NoError(s.T(), err)

	router := mux.NewRouter()
	router.HandleFunc("/posts/period/{from}-to-{to}", BuildPeriodHandler(s.DB, templating.BuildPageRenderFunc())).Methods("GET")

	req, err := http.NewRequest("GET", "/posts/period/2021-11-01-to-2021-11-29", nil)
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

	assert.NotContains(s.T(), string(body), "older post")
	assert.Contains(s.T(), string(body), "post in range")
	assert.NotContains(s.T(), string(body), "future post")
}
