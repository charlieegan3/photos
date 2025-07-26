package public

import (
	"context"
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

type DevicesSuite struct {
	suite.Suite

	DB            *sql.DB
	Bucket        *blob.Bucket
	BucketBaseURL string
}

func (s *DevicesSuite) SetupTest() {
	var err error
	err = database.Truncate(s.T().Context(), s.DB, "photos.posts")
	s.Require().NoError(err)
	err = database.Truncate(s.T().Context(), s.DB, "photos.locations")
	s.Require().NoError(err)
	err = database.Truncate(s.T().Context(), s.DB, "photos.medias")
	s.Require().NoError(err)
	err = database.Truncate(s.T().Context(), s.DB, "photos.devices")
	s.Require().NoError(err)
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
	returnedDevices, err := database.CreateDevices(s.T().Context(), s.DB, devices)
	s.Require().NoError(err)

	router := mux.NewRouter()
	router.HandleFunc("/devices",
		BuildIndexHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodGet)

	req, err := http.NewRequestWithContext(s.T().Context(), http.MethodGet, "/devices", nil)
	s.Require().NoError(err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if !s.Equal(http.StatusOK, rr.Code) {
		bodyString, err := io.ReadAll(rr.Body)
		s.Require().NoError(err)
		s.T().Fatalf("request failed with: %s", bodyString)
	}

	body, err := io.ReadAll(rr.Body)
	s.Require().NoError(err)

	s.Contains(string(body), fmt.Sprintf("/devices/%d", returnedDevices[0].ID))
	s.Contains(string(body), fmt.Sprintf("/devices/%d", returnedDevices[1].ID))
	s.Contains(string(body), returnedDevices[0].Name)
	s.Contains(string(body), returnedDevices[0].Name)
}

func (s *DevicesSuite) TestShow() {
	devices := []models.Device{
		{
			Name:     "iPhone",
			IconKind: "jpg",
		},
	}
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
	}
	returnedLocations, err := database.CreateLocations(s.T().Context(), s.DB, locations)
	s.Require().NoError(err)

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
	returnedPosts, err := database.CreatePosts(s.T().Context(), s.DB, posts)
	s.Require().NoError(err)

	router := mux.NewRouter()
	router.HandleFunc("/devices/{deviceID}",
		BuildShowHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodGet)

	req, err := http.NewRequestWithContext(
		s.T().Context(), http.MethodGet, fmt.Sprintf("/devices/%d", returnedDevices[0].ID), nil,
	)
	s.Require().NoError(err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if !s.Equal(http.StatusOK, rr.Code) {
		bodyString, err := io.ReadAll(rr.Body)
		s.Require().NoError(err)
		s.T().Fatalf("request failed with: %s", bodyString)
	}

	body, err := io.ReadAll(rr.Body)
	s.Require().NoError(err)

	s.Contains(string(body), fmt.Sprintf("/posts/%d", returnedPosts[0].ID))
	s.Contains(string(body), fmt.Sprintf("/posts/%d", returnedPosts[1].ID))
}

func (s *DevicesSuite) TestGetIcon() {
	devices := []models.Device{
		{
			Name:     "Example Device",
			IconKind: "jpg",
		},
	}
	returnedDevices, err := database.CreateDevices(s.T().Context(), s.DB, devices)
	s.Require().NoError(err)

	// store an image for the device in the bucket to be served in the request.
	imageFilePath := "../../../pkg/mediametadata/samples/iphone-11-pro-max.jpg"
	imageBytes, err := os.ReadFile(imageFilePath)
	s.Require().NoError(err)
	imageFile, err := os.Open(imageFilePath)
	s.Require().NoError(err)
	bw, err := s.Bucket.NewWriter(context.Background(), fmt.Sprintf("device_icons/%s.jpg", returnedDevices[0].Slug), nil)
	s.Require().NoError(err)
	_, err = io.Copy(bw, imageFile)
	s.Require().NoError(err)
	err = bw.Close()
	s.Require().NoError(err)

	router := mux.NewRouter()
	router.HandleFunc("/devices/{deviceID}/icon.{kind}",
		BuildIconHandler(s.DB, s.Bucket)).
		Methods(http.MethodGet)

	req, err := http.NewRequestWithContext(
		s.T().Context(), http.MethodGet, fmt.Sprintf("/devices/%d/icon.jpg", returnedDevices[0].ID), nil,
	)
	s.Require().NoError(err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if !s.Equal(http.StatusOK, rr.Code) {
		bodyString, err := io.ReadAll(rr.Body)
		s.Require().NoError(err)
		s.T().Fatalf("request failed with: %s", bodyString)
	}

	body, err := io.ReadAll(rr.Body)
	s.Require().NoError(err)

	// validate that the images are the same
	h := sha1.New()
	h.Write(body)
	bodySha := hex.EncodeToString(h.Sum(nil))
	h = sha1.New()
	h.Write(imageBytes)
	imageSha := hex.EncodeToString(h.Sum(nil))
	s.Equal(bodySha, imageSha)
}
