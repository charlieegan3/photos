package public

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
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

type LensesSuite struct {
	suite.Suite
	DB            *sql.DB
	Bucket        *blob.Bucket
	BucketBaseURL string
}

func (s *LensesSuite) SetupTest() {
	err := database.Truncate(s.DB, "lenses")
	require.NoError(s.T(), err)
}

func (s *LensesSuite) TestGetIcon() {
	lenses := []models.Lens{{Name: "Example Lens"}}
	returnedLenses, err := database.CreateLenses(s.DB, lenses)
	require.NoError(s.T(), err)

	// store an image for the lens in the bucket to be served in the request.
	imageFilePath := "../../../pkg/mediametadata/samples/iphone-11-pro-max.jpg"
	imageBytes, err := ioutil.ReadFile(imageFilePath)
	require.NoError(s.T(), err)
	imageFile, err := os.Open(imageFilePath)
	require.NoError(s.T(), err)
	bw, err := s.Bucket.NewWriter(context.Background(), fmt.Sprintf("lens_icons/%d.jpg", returnedLenses[0].ID), nil)
	require.NoError(s.T(), err)
	_, err = io.Copy(bw, imageFile)
	err = bw.Close()
	require.NoError(s.T(), err)

	router := mux.NewRouter()
	router.HandleFunc("/lenses/{lensID}/icon.{kind}", BuildIconHandler(s.DB, s.Bucket)).Methods("GET")

	req, err := http.NewRequest("GET", fmt.Sprintf("/lenses/%d/icon.jpg", returnedLenses[0].ID), nil)
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

	// validate that the images are the same
	h := sha1.New()
	h.Write(body)
	bodySha := hex.EncodeToString(h.Sum(nil))
	h = sha1.New()
	h.Write(imageBytes)
	imageSha := hex.EncodeToString(h.Sum(nil))
	assert.Equal(s.T(), bodySha, imageSha)
}
