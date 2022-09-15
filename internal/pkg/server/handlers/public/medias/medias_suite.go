package public

import (
	"bytes"
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"

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

type MediasSuite struct {
	suite.Suite
	DB            *sql.DB
	Bucket        *blob.Bucket
	BucketBaseURL string
}

func (s *MediasSuite) SetupTest() {
	err := database.Truncate(s.DB, "devices")
	require.NoError(s.T(), err)
	err = database.Truncate(s.DB, "medias")
	require.NoError(s.T(), err)
}

func (s *MediasSuite) TestGetMedia() {
	devices := []models.Device{
		{
			Name: "Example Device",
		},
	}
	returnedDevices, err := database.CreateDevices(s.DB, devices)
	require.NoError(s.T(), err)

	// insert a sample media to allow the request to be validated as being for a valid media item
	medias := []models.Media{
		{
			DeviceID: returnedDevices[0].ID,
			Make:     "Apple",
			Kind:     "jpg",
		},
	}

	returnedMedias, err := database.CreateMedias(s.DB, medias)
	require.NoError(s.T(), err)

	// store an image for the media in the bucket to be served in the request.
	imageFilePath := "../../../pkg/server/handlers/public/medias/fixtures/image-original.jpg"
	imageFile, err := os.Open(imageFilePath)
	require.NoError(s.T(), err)
	bw, err := s.Bucket.NewWriter(context.Background(), fmt.Sprintf("media/%d.jpg", returnedMedias[0].ID), nil)
	require.NoError(s.T(), err)
	_, err = io.Copy(bw, imageFile)
	err = bw.Close()
	require.NoError(s.T(), err)

	router := mux.NewRouter()
	router.HandleFunc("/medias/{mediaID}/{file}.{kind}", BuildMediaHandler(s.DB, s.Bucket)).Methods("GET")

	req, err := http.NewRequest("GET", fmt.Sprintf("/medias/%d/file.jpg?o=100x", returnedMedias[0].ID), nil)
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

	// load the resized image to test the response content
	imageFilePath = "../../../pkg/server/handlers/public/medias/fixtures/image-100x.jpg"
	imageBytes, err := os.ReadFile(imageFilePath)
	require.NoError(s.T(), err)
	h = sha1.New()
	h.Write(imageBytes)
	imageSha := hex.EncodeToString(h.Sum(nil))
	assert.Equal(s.T(), bodySha, imageSha)

	// check that the image has been stashed in the bucket for future requests
	br, err := s.Bucket.NewReader(context.Background(), fmt.Sprintf("thumbs/media/%d-100x.jpg", returnedMedias[0].ID), nil)
	require.NoError(s.T(), err)
	defer br.Close()

	buf := bytes.NewBuffer([]byte{})
	_, err = io.Copy(buf, br)
	require.NoError(s.T(), err)

	h = sha1.New()
	h.Write(buf.Bytes())
	objectSha := hex.EncodeToString(h.Sum(nil))
	assert.Equal(s.T(), objectSha, imageSha)
}
