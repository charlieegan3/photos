package medias

import (
	"bytes"
	"context"
	"crypto/md5"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/maxatome/go-testdeep/td"
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

type EndpointsMediasSuite struct {
	suite.Suite
	DB            *sql.DB
	Bucket        *blob.Bucket
	BucketBaseURL string
}

func (s *EndpointsMediasSuite) SetupTest() {
	err := database.Truncate(s.DB, "medias")
	require.NoError(s.T(), err)
}

func (s *EndpointsMediasSuite) TestListMedias() {
	testData := []models.Media{
		{
			Make:  "FujiFilm",
			Model: "X100F",

			TakenAt: time.Date(2021, time.November, 23, 19, 56, 0, 0, time.UTC),

			FNumber:      2.0,
			ShutterSpeed: 0.004,
			ISOSpeed:     100,

			Latitude:  51.1,
			Longitude: 52.2,
			Altitude:  100.0,
		},
		{
			Make:  "Apple",
			Model: "iPhone",

			TakenAt: time.Date(2021, time.September, 22, 18, 56, 0, 0, time.UTC),

			FNumber:      4.0,
			ShutterSpeed: 0.05,
			ISOSpeed:     400,

			Latitude:  53.1,
			Longitude: 54.2,
			Altitude:  200.0,
		},
	}

	returnedMedias, err := database.CreateMedias(s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create medias: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/admin/medias", BuildIndexHandler(s.DB, templating.BuildPageRenderFunc("http://"))).Methods("GET")

	req, err := http.NewRequest("GET", "/admin/medias", nil)
	require.NoError(s.T(), err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(s.T(), http.StatusOK, rr.Code)

	body, err := ioutil.ReadAll(rr.Body)
	require.NoError(s.T(), err)

	assert.Contains(s.T(), string(body), fmt.Sprintf("id: %d", returnedMedias[0].ID))
	assert.Contains(s.T(), string(body), fmt.Sprintf("id: %d", returnedMedias[1].ID))
}

func (s *EndpointsMediasSuite) TestGetMedia() {
	testData := []models.Media{
		{
			Make:  "FujiFilm",
			Model: "X100F",

			TakenAt: time.Date(2021, time.November, 23, 19, 56, 0, 0, time.UTC),

			FNumber:      2.0,
			ShutterSpeed: 0.004,
			ISOSpeed:     100,

			Latitude:  51.1,
			Longitude: 52.2,
			Altitude:  100.0,
		},
	}

	persistedMedias, err := database.CreateMedias(s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create medias: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/admin/medias/{mediaID}", BuildGetHandler(s.DB, templating.BuildPageRenderFunc("http://"))).Methods("GET")

	req, err := http.NewRequest("GET", fmt.Sprintf("/admin/medias/%d", persistedMedias[0].ID), nil)
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

	assert.Contains(s.T(), string(body), fmt.Sprintf("ID: %d", persistedMedias[0].ID))
}

func (s *EndpointsMediasSuite) TestUpdateMedia() {
	testData := []models.Media{
		{
			Make:  "FujiFilm",
			Model: "X100F",

			TakenAt: time.Date(2021, time.November, 23, 19, 56, 0, 0, time.UTC),

			FNumber:      2.0,
			ShutterSpeed: 0.004,
			ISOSpeed:     100,

			Latitude:  51.1,
			Longitude: 52.2,
			Altitude:  100.0,
		},
	}

	persistedMedias, err := database.CreateMedias(s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create medias: %s", err)
	}

	// store the an icon in the bucket
	imageFilePath := "../../../pkg/mediametadata/samples/iphone-11-pro-max.jpg"
	imageFile, err := os.Open(imageFilePath)
	require.NoError(s.T(), err)
	bw, err := s.Bucket.NewWriter(context.Background(), fmt.Sprintf("media/%d.jpg", persistedMedias[0].ID), nil)
	require.NoError(s.T(), err)
	_, err = io.Copy(bw, imageFile)
	err = bw.Close()
	require.NoError(s.T(), err)

	router := mux.NewRouter()
	router.HandleFunc("/admin/medias/{mediaID}", BuildFormHandler(s.DB, s.Bucket, templating.BuildPageRenderFunc("http://"))).Methods("POST")

	// open the image to be uploaded in the form

	// build the form to be posted
	values := map[string]io.Reader{
		"Make":    strings.NewReader("Fuji"),
		"File":    imageFile,
		"_method": strings.NewReader("PUT"),
	}
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		if x, ok := r.(*os.File); ok {
			fw, err = w.CreateFormFile(key, x.Name())
			require.NoError(s.T(), err)
		} else {
			fw, err = w.CreateFormField(key)
			require.NoError(s.T(), err)
		}
		_, err = io.Copy(fw, r)
		require.NoError(s.T(), err)
	}
	w.Close()

	// make the request to the handler
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("/admin/medias/%d", persistedMedias[0].ID),
		&b,
	)
	require.NoError(s.T(), err)

	req.Header.Set("Content-Type", w.FormDataContentType())

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(s.T(), http.StatusSeeOther, rr.Code)

	// check that the database content is also correct
	returnedMedias, err := database.AllMedias(s.DB)
	if err != nil {
		s.T().Fatalf("failed to list medias: %s", err)
	}
	expectedMedias := td.Slice(
		[]models.Media{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Media{
					ID:   persistedMedias[0].ID,
					Make: "Fuji",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedMedias, expectedMedias)
}

func (s *EndpointsMediasSuite) TestDeleteMedia() {
	testData := []models.Media{
		{
			Kind: "jpg",

			Make:  "FujiFilm",
			Model: "X100F",

			TakenAt: time.Date(2021, time.November, 23, 19, 56, 0, 0, time.UTC),

			FNumber:      2.0,
			ShutterSpeed: 0.004,
			ISOSpeed:     100,

			Latitude:  51.1,
			Longitude: 52.2,
			Altitude:  100.0,
		},
	}

	persistedMedias, err := database.CreateMedias(s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create medias: %s", err)
	}

	// store the an icon in the bucket, check it's deleted
	imageFilePath := "../../../pkg/mediametadata/samples/iphone-11-pro-max.jpg"
	imageFile, err := os.Open(imageFilePath)
	require.NoError(s.T(), err)
	bw, err := s.Bucket.NewWriter(context.Background(), fmt.Sprintf("media/%d.jpg", persistedMedias[0].ID), nil)
	require.NoError(s.T(), err)
	_, err = io.Copy(bw, imageFile)
	err = bw.Close()
	require.NoError(s.T(), err)

	router := mux.NewRouter()
	router.HandleFunc(
		"/admin/medias/{mediaID}",
		BuildFormHandler(s.DB, s.Bucket, templating.BuildPageRenderFunc("http://")),
	).Methods("POST")

	form := url.Values{}
	form.Add("_method", "DELETE")

	// make the request to the handler
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("/admin/medias/%d", persistedMedias[0].ID),
		strings.NewReader(form.Encode()),
	)
	require.NoError(s.T(), err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(s.T(), http.StatusSeeOther, rr.Code)

	// check that the database content is also correct
	returnedMedias, err := database.AllMedias(s.DB)
	if err != nil {
		s.T().Fatalf("failed to list medias: %s", err)
	}

	expectedMedias := []models.Media{}
	td.Cmp(s.T(), returnedMedias, expectedMedias)

	// should have a not found error as the icon should have been deleted
	_, err = s.Bucket.Attributes(context.Background(), "media/iphone.jpg")
	require.Error(s.T(), err)
}

func (s *EndpointsMediasSuite) TestNewMedia() {
	router := mux.NewRouter()
	router.HandleFunc("/admin/medias/new", BuildNewHandler(templating.BuildPageRenderFunc("http://"))).Methods("GET")

	req, err := http.NewRequest("GET", "/admin/medias/new", nil)
	require.NoError(s.T(), err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(s.T(), http.StatusOK, rr.Code)

	body, err := ioutil.ReadAll(rr.Body)
	require.NoError(s.T(), err)

	assert.Contains(s.T(), string(body), "File")
	// only show the file field on create
	assert.NotContains(s.T(), string(body), "Make")
}

func (s *EndpointsMediasSuite) TestCreateMedia() {
	router := mux.NewRouter()
	router.HandleFunc("/admin/medias", BuildCreateHandler(s.DB, s.Bucket, templating.BuildPageRenderFunc("http://"))).Methods("POST")

	// open the image to be uploaded in the form
	imageFilePath := "../../../pkg/mediametadata/samples/iphone-11-pro-max.jpg"
	imageFile, err := os.Open(imageFilePath)
	require.NoError(s.T(), err)

	// build the form to be posted
	values := map[string]io.Reader{
		"File": imageFile,
	}
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		if x, ok := r.(*os.File); ok {
			fw, err = w.CreateFormFile(key, x.Name())
			require.NoError(s.T(), err)
		} else {
			fw, err = w.CreateFormField(key)
			require.NoError(s.T(), err)
		}
		_, err = io.Copy(fw, r)
		require.NoError(s.T(), err)
	}
	w.Close()

	// make the request to the handler
	req, err := http.NewRequest(
		"POST",
		"/admin/medias",
		&b,
	)
	require.NoError(s.T(), err)

	req.Header.Set("Content-Type", w.FormDataContentType())

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// check that we get a see other response to the right location
	require.Equal(s.T(), http.StatusSeeOther, rr.Code)
	if !strings.HasPrefix(rr.HeaderMap["Location"][0], "/admin/medias/") {
		s.T().Fatalf("%v doesn't appear to be the correct path", rr.HeaderMap["Location"])
	}

	// check that the database content is also correct
	returnedMedias, err := database.AllMedias(s.DB)
	require.NoError(s.T(), err)

	expectedMedias := td.Slice(
		[]models.Media{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Media{
					// set from image exif data
					Make:     "Apple",
					Altitude: 97.99822998046875,
					FNumber:  2.0,
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedMedias, expectedMedias)

	// check that the image has been uploaded ok
	// get a digest for the image in the bucket
	r, err := s.Bucket.NewReader(context.Background(), fmt.Sprintf("media/%d.jpg", returnedMedias[0].ID), nil)
	defer r.Close()
	require.NoError(s.T(), err)
	bucketHash := md5.New()
	_, err = io.Copy(bucketHash, r)
	require.NoError(s.T(), err)
	bucketMD5 := fmt.Sprintf("%x", bucketHash.Sum(nil))

	// get a digest for the image originally uploaded
	f, err := os.Open(imageFilePath)
	require.NoError(s.T(), err)
	defer f.Close()
	sourceHash := md5.New()
	_, err = io.Copy(sourceHash, f)
	require.NoError(s.T(), err)
	sourceMD5 := fmt.Sprintf("%x", bucketHash.Sum(nil))

	require.Equal(s.T(), bucketMD5, sourceMD5)
}
