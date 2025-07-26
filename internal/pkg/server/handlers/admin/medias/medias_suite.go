package medias

import (
	"bytes"
	"context"
	"crypto/md5"
	"database/sql"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/suite"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/memblob"

	"github.com/charlieegan3/photos/internal/pkg/database"
	"github.com/charlieegan3/photos/internal/pkg/models"
	"github.com/charlieegan3/photos/internal/pkg/server/templating"
)

type EndpointsMediasSuite struct {
	suite.Suite
	DB            *sql.DB
	Bucket        *blob.Bucket
	BucketBaseURL string
}

func (s *EndpointsMediasSuite) SetupTest() {
	err := database.Truncate(s.DB, "photos.medias")
	s.Require().NoError(err)

	err = database.Truncate(s.DB, "photos.devices")
	s.Require().NoError(err)
}

func (s *EndpointsMediasSuite) TestListMedias() {
	devices := []models.Device{
		{
			Name: "Example Device",
		},
	}

	returnedDevices, err := database.CreateDevices(s.DB, devices)
	if err != nil {
		s.T().Fatalf("failed to create devices: %s", err)
	}

	testData := []models.Media{
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
		{
			DeviceID: returnedDevices[0].ID,

			Make:  "Apple",
			Model: "iPhone",

			TakenAt: time.Date(2021, time.September, 22, 18, 56, 0, 0, time.UTC),

			FNumber:  4.0,
			ISOSpeed: 400,

			Latitude:  53.1,
			Longitude: 54.2,
			Altitude:  200.0,
		},
	}

	returnedMedias, err := database.CreateMedias(s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create medias: %s", err)
	}

	locations := []models.Location{
		{
			Name:      "London",
			Latitude:  1.1,
			Longitude: 1.2,
		},
	}

	returnedLocations, err := database.CreateLocations(s.DB, locations)
	s.Require().NoError(err)

	posts := []models.Post{
		{
			Description: "Here is a shot I took",
			PublishDate: time.Date(2021, time.November, 24, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
		},
	}

	_, err = database.CreatePosts(s.DB, posts)
	s.Require().NoError(err)

	router := mux.NewRouter()
	router.HandleFunc("/admin/medias",
		BuildIndexHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodGet)

	req, err := http.NewRequest(http.MethodGet, "/admin/medias", nil)
	s.Require().NoError(err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	s.Require().Equal(http.StatusOK, rr.Code)

	body, err := io.ReadAll(rr.Body)
	s.Require().NoError(err)

	s.Contains(string(body), fmt.Sprintf("id: %d", returnedMedias[0].ID))
	cleanedBody := regexp.MustCompile(`\s+`).ReplaceAllString(string(body), " ")
	s.Contains(cleanedBody, fmt.Sprintf("id: %d (not posted)", returnedMedias[1].ID))
}

func (s *EndpointsMediasSuite) TestGetMedia() {
	devices := []models.Device{
		{
			Name: "Example Device",
		},
	}

	returnedDevices, err := database.CreateDevices(s.DB, devices)
	if err != nil {
		s.T().Fatalf("failed to create devices: %s", err)
	}

	testData := []models.Media{
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

	persistedMedias, err := database.CreateMedias(s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create medias: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/admin/medias/{mediaID}",
		BuildGetHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodGet)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/admin/medias/%d", persistedMedias[0].ID), nil)
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

	s.Contains(string(body), fmt.Sprintf("%d", persistedMedias[0].ID))
}

func (s *EndpointsMediasSuite) TestUpdateMedia() {
	devices := []models.Device{
		{
			Name: "Example Device",
		},
	}

	returnedDevices, err := database.CreateDevices(s.DB, devices)
	s.Require().NoError(err)

	lenses := []models.Lens{{Name: "Example Lens"}}
	returnedLenses, err := database.CreateLenses(s.DB, lenses)
	s.Require().NoError(err)

	testData := []models.Media{
		{
			DeviceID: returnedDevices[0].ID,
			LensID:   returnedLenses[0].ID,

			Make:  "FujiFilm",
			Model: "X100F",

			TakenAt: time.Date(2021, time.November, 23, 19, 56, 0, 0, time.UTC),

			FNumber:  2.0,
			ISOSpeed: 100,

			Latitude:  51.1,
			Longitude: 52.2,
			Altitude:  100.0,

			DisplayOffset: 10,
		},
	}

	persistedMedias, err := database.CreateMedias(s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create medias: %s", err)
	}

	// store an icon in the bucket
	imageFilePath := "../../../pkg/mediametadata/samples/iphone-11-pro-max.jpg"
	imageFile, err := os.Open(imageFilePath)
	s.Require().NoError(err)
	bw, err := s.Bucket.NewWriter(context.Background(), fmt.Sprintf("media/%d.jpg", persistedMedias[0].ID), nil)
	s.Require().NoError(err)
	_, err = io.Copy(bw, imageFile)
	s.Require().NoError(err)
	err = bw.Close()
	s.Require().NoError(err)

	router := mux.NewRouter()
	router.HandleFunc("/admin/medias/{mediaID}",
		BuildFormHandler(s.DB, s.Bucket, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodPost)

	// open the image to be uploaded in the form

	// build the form to be posted
	values := map[string]io.Reader{
		"Make":          strings.NewReader("Fuji"),
		"File":          imageFile,
		"DeviceID":      strings.NewReader(fmt.Sprintf("%d", returnedDevices[0].ID)),
		"DisplayOffset": strings.NewReader("50"),
		"LensID":        strings.NewReader("0"), // remove the previously set lens
		"_method":       strings.NewReader(http.MethodPut),
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
			s.Require().NoError(err)
		} else {
			fw, err = w.CreateFormField(key)
			s.Require().NoError(err)
		}
		_, err = io.Copy(fw, r)
		s.Require().NoError(err)
	}
	w.Close()

	// make the request to the handler
	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("/admin/medias/%d", persistedMedias[0].ID),
		&b,
	)
	s.Require().NoError(err)

	req.Header.Set("Content-Type", w.FormDataContentType())

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if !s.Equal(http.StatusSeeOther, rr.Code) {
		bodyString, err := io.ReadAll(rr.Body)
		s.Require().NoError(err)
		s.T().Fatalf("request failed with: %s", bodyString)
	}

	// check that the database content is also correct
	returnedMedias, err := database.AllMedias(s.DB, false)
	if err != nil {
		s.T().Fatalf("failed to list medias: %s", err)
	}
	expectedMedias := td.Slice(
		[]models.Media{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Media{
					ID:            persistedMedias[0].ID,
					Make:          "Fuji",
					LensID:        0,
					DisplayOffset: 50,
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedMedias, expectedMedias)
}

func (s *EndpointsMediasSuite) TestDeleteMedia() {
	devices := []models.Device{
		{
			Name: "Example Device",
		},
	}

	returnedDevices, err := database.CreateDevices(s.DB, devices)
	if err != nil {
		s.T().Fatalf("failed to create devices: %s", err)
	}

	testData := []models.Media{
		{
			DeviceID: returnedDevices[0].ID,

			Kind: "jpg",

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

	persistedMedias, err := database.CreateMedias(s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create medias: %s", err)
	}

	// store the image in the bucket, check it's deleted
	imageFilePath := "../../../pkg/mediametadata/samples/iphone-11-pro-max.jpg"
	imageFile, err := os.Open(imageFilePath)
	s.Require().NoError(err)
	bw, err := s.Bucket.NewWriter(context.Background(), fmt.Sprintf("media/%d.jpg", persistedMedias[0].ID), nil)
	s.Require().NoError(err)
	_, err = io.Copy(bw, imageFile)
	s.Require().NoError(err)
	err = bw.Close()
	s.Require().NoError(err)
	imageFile.Close()

	// write a thumbnail to test these are also deleted, in this case there's only one thumb
	imageFile, err = os.Open(imageFilePath)
	s.Require().NoError(err)
	bw, err = s.Bucket.NewWriter(context.Background(), fmt.Sprintf("thumbs/%d-foobar.jpg", persistedMedias[0].ID), nil)
	s.Require().NoError(err)
	_, err = io.Copy(bw, imageFile)
	s.Require().NoError(err)
	err = bw.Close()
	s.Require().NoError(err)
	imageFile.Close()

	router := mux.NewRouter()
	router.HandleFunc(
		"/admin/medias/{mediaID}",
		BuildFormHandler(s.DB, s.Bucket, templating.BuildPageRenderFunc(true, "")),
	).Methods(http.MethodPost)

	form := url.Values{}
	form.Add("_method", http.MethodDelete)

	// make the request to the handler
	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("/admin/medias/%d", persistedMedias[0].ID),
		strings.NewReader(form.Encode()),
	)
	s.Require().NoError(err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	s.Require().Equal(http.StatusSeeOther, rr.Code)

	// check that the database content is also correct
	returnedMedias, err := database.AllMedias(s.DB, false)
	if err != nil {
		s.T().Fatalf("failed to list medias: %s", err)
	}

	expectedMedias := []models.Media{}
	td.Cmp(s.T(), returnedMedias, expectedMedias)

	// should have a not found error as the icon should have been deleted
	_, err = s.Bucket.Attributes(context.Background(), "media/iphone.jpg")
	s.Require().Error(err)

	var thumbs []string
	listOptions := &blob.ListOptions{
		Prefix: fmt.Sprintf("thumbs/media/%d-", persistedMedias[0].ID),
	}
	iter := s.Bucket.List(listOptions)
	for {
		obj, err := iter.Next(context.Background())
		if err == io.EOF {
			break
		}
		s.Require().NoError(err)

		thumbs = append(thumbs, obj.Key)
	}

	if len(thumbs) > 0 {
		s.T().Fatalf("thumbs not deleted correctly")
	}
}

func (s *EndpointsMediasSuite) TestNewMedia() {
	router := mux.NewRouter()
	router.HandleFunc("/admin/medias/new",
		BuildNewHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodGet)

	req, err := http.NewRequest(http.MethodGet, "/admin/medias/new", nil)
	s.Require().NoError(err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	s.Require().Equal(http.StatusOK, rr.Code)

	body, err := io.ReadAll(rr.Body)
	s.Require().NoError(err)

	s.Contains(string(body), "File")
	// only show the file field on create
	s.NotContains(string(body), "Make")
}

func (s *EndpointsMediasSuite) TestCreateMedia() {
	devices := []models.Device{
		{
			Name: "Example Device",
		},
	}

	returnedDevices, err := database.CreateDevices(s.DB, devices)
	if err != nil {
		s.T().Fatalf("failed to create devices: %s", err)
	}

	lenses := []models.Lens{
		{
			Name:        "Dummy Lens",
			LensMatches: "back camera",
		},
		{
			Name:        "Dummy Lens 2",
			LensMatches: "iPhone 11 Pro Max back triple camera 6mm f/2",
		},
	}

	returnedLenses, err := database.CreateLenses(s.DB, lenses)
	if err != nil {
		s.T().Fatalf("failed to create lenses: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/admin/medias",
		BuildCreateHandler(s.DB, s.Bucket, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodPost)

	// open the image to be uploaded in the form
	imageFilePath := "../../../pkg/mediametadata/samples/iphone-11-pro-max.jpg"
	imageFile, err := os.Open(imageFilePath)
	s.Require().NoError(err)

	// build the form to be posted
	values := map[string]io.Reader{
		"File":     imageFile,
		"DeviceID": strings.NewReader(fmt.Sprintf("%d", returnedDevices[0].ID)),
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
			s.Require().NoError(err)
		} else {
			fw, err = w.CreateFormField(key)
			s.Require().NoError(err)
		}
		_, err = io.Copy(fw, r)
		s.Require().NoError(err)
	}
	w.Close()

	// make the request to the handler
	req, err := http.NewRequest(
		http.MethodPost,
		"/admin/medias",
		&b,
	)
	s.Require().NoError(err)

	req.Header.Set("Content-Type", w.FormDataContentType())

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// check that we get a see other response to the right location
	if !s.Equal(http.StatusSeeOther, rr.Code) {
		bodyString, err := io.ReadAll(rr.Body)
		s.Require().NoError(err)
		s.T().Fatalf("request failed with: %s", bodyString)
	}
	if !strings.HasPrefix(rr.Result().Header["Location"][0], "/admin/medias/") {
		s.T().Fatalf("%v doesn't appear to be the correct path", rr.Result().Header["Location"])
	}

	// check that the database content is also correct
	returnedMedias, err := database.AllMedias(s.DB, false)
	s.Require().NoError(err)

	expectedMedias := td.Slice(
		[]models.Media{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Media{
					// set from image exif data
					Make:       "Apple",
					Lens:       "iPhone 11 Pro Max back triple camera 6mm f/2",
					Altitude:   97.99822998046875,
					FNumber:    2.0,
					UTCCorrect: true,

					DeviceID: returnedDevices[0].ID,
					LensID:   returnedLenses[1].ID,

					Width:  4032,
					Height: 3024,

					ExposureTimeDenominator: 122,
					ExposureTimeNumerator:   1,
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
	s.Require().NoError(err)
	defer r.Close()
	bucketHash := md5.New()
	_, err = io.Copy(bucketHash, r)
	s.Require().NoError(err)
	bucketMD5 := fmt.Sprintf("%x", bucketHash.Sum(nil))

	// get a digest for the image originally uploaded
	f, err := os.Open(imageFilePath)
	s.Require().NoError(err)
	defer f.Close()
	sourceHash := md5.New()
	_, err = io.Copy(sourceHash, f)
	s.Require().NoError(err)
	sourceMD5 := fmt.Sprintf("%x", bucketHash.Sum(nil))

	s.Require().Equal(bucketMD5, sourceMD5)

	// check that the thumbs have been created too
	var thumbs []string
	listOptions := &blob.ListOptions{
		Prefix: fmt.Sprintf("thumbs/media/%d-", returnedMedias[0].ID),
	}
	iter := s.Bucket.List(listOptions)
	for {
		obj, err := iter.Next(context.Background())
		if err == io.EOF {
			break
		}
		s.Require().NoError(err)

		thumbs = append(thumbs, obj.Key)
	}

	s.Require().ElementsMatchf(thumbs, []string{
		fmt.Sprintf("thumbs/media/%d-200-fit.jpg", returnedMedias[0].ID),
		fmt.Sprintf("thumbs/media/%d-500-fit.jpg", returnedMedias[0].ID),
		fmt.Sprintf("thumbs/media/%d-1000-fit.jpg", returnedMedias[0].ID),
		fmt.Sprintf("thumbs/media/%d-2000-fit.jpg", returnedMedias[0].ID),
	}, "thumbs not created correctly")
}
