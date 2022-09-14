package public

import (
	"context"
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

type DevicesSuite struct {
	suite.Suite
	DB            *sql.DB
	Bucket        *blob.Bucket
	BucketBaseURL string
}

func (s *DevicesSuite) SetupTest() {
	var err error
	err = database.Truncate(s.DB, "posts")
	require.NoError(s.T(), err)
	err = database.Truncate(s.DB, "locations")
	require.NoError(s.T(), err)
	err = database.Truncate(s.DB, "medias")
	require.NoError(s.T(), err)
	err = database.Truncate(s.DB, "devices")
	require.NoError(s.T(), err)
}

func (s *DevicesSuite) TestIndex() {
	devices := []models.Device{
		{
			Name:     "iPhone",
			IconKind: "jpg",
		},
		{
			Name:     "X100F",
			IconKind: "png",
		},
	}
	returnedDevices, err := database.CreateDevices(s.DB, devices)
	require.NoError(s.T(), err)

	router := mux.NewRouter()
	router.HandleFunc("/devices", BuildIndexHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).Methods("GET")

	req, err := http.NewRequest("GET", "/devices", nil)
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

	assert.Contains(s.T(), string(body), fmt.Sprintf("/devices/%d", returnedDevices[0].ID))
	assert.Contains(s.T(), string(body), fmt.Sprintf("/devices/%d", returnedDevices[1].ID))
	assert.Contains(s.T(), string(body), returnedDevices[0].Name)
	assert.Contains(s.T(), string(body), returnedDevices[0].Name)
}

func (s *DevicesSuite) TestShow() {
	devices := []models.Device{
		{
			Name:     "iPhone",
			IconKind: "jpg",
		},
	}
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
	}
	returnedLocations, err := database.CreateLocations(s.DB, locations)
	require.NoError(s.T(), err)

	posts := []models.Post{
		{
			Description: "post from 2022",
			PublishDate: time.Date(2021, time.January, 1, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
		},
		{
			Description: "post from 2021",
			PublishDate: time.Date(2021, time.January, 1, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[1].ID,
			LocationID:  returnedLocations[0].ID,
		},
	}
	returnedPosts, err := database.CreatePosts(s.DB, posts)
	require.NoError(s.T(), err)

	router := mux.NewRouter()
	router.HandleFunc("/devices/{deviceID}", BuildShowHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).Methods("GET")

	req, err := http.NewRequest("GET", fmt.Sprintf("/devices/%d", returnedDevices[0].ID), nil)
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

	assert.Contains(s.T(), string(body), fmt.Sprintf("/posts/%d", returnedPosts[0].ID))
	assert.Contains(s.T(), string(body), fmt.Sprintf("/posts/%d", returnedPosts[1].ID))
}

func (s *DevicesSuite) TestGetIcon() {
	devices := []models.Device{
		{
			Name:     "Example Device",
			IconKind: "jpg",
		},
	}
	returnedDevices, err := database.CreateDevices(s.DB, devices)
	require.NoError(s.T(), err)

	// store an image for the device in the bucket to be served in the request.
	imageFilePath := "../../../pkg/mediametadata/samples/iphone-11-pro-max.jpg"
	imageBytes, err := os.ReadFile(imageFilePath)
	require.NoError(s.T(), err)
	imageFile, err := os.Open(imageFilePath)
	require.NoError(s.T(), err)
	bw, err := s.Bucket.NewWriter(context.Background(), fmt.Sprintf("device_icons/%s.jpg", returnedDevices[0].Slug), nil)
	require.NoError(s.T(), err)
	_, err = io.Copy(bw, imageFile)
	err = bw.Close()
	require.NoError(s.T(), err)

	router := mux.NewRouter()
	router.HandleFunc("/devices/{deviceID}/icon.{kind}", BuildIconHandler(s.DB, s.Bucket)).Methods("GET")

	req, err := http.NewRequest("GET", fmt.Sprintf("/devices/%d/icon.jpg", returnedDevices[0].ID), nil)
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

	// validate that the images are the same
	h := sha1.New()
	h.Write(body)
	bodySha := hex.EncodeToString(h.Sum(nil))
	h = sha1.New()
	h.Write(imageBytes)
	imageSha := hex.EncodeToString(h.Sum(nil))
	assert.Equal(s.T(), bodySha, imageSha)
}
