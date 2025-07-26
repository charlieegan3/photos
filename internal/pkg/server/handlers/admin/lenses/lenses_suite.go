package lenses

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
	"strings"

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

type EndpointsLensesSuite struct {
	suite.Suite
	DB            *sql.DB
	Bucket        *blob.Bucket
	BucketBaseURL string
}

func (s *EndpointsLensesSuite) SetupTest() {
	err := database.Truncate(s.DB, "photos.lenses")
	s.Require().NoError(err)
}

func (s *EndpointsLensesSuite) TestListLenses() {
	testData := []models.Lens{
		{
			Name: "iPhone",
		},
		{
			Name: "X100F",
		},
	}

	_, err := database.CreateLenses(s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create lenses: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/admin/lenses",
		BuildIndexHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodGet)

	req, err := http.NewRequest(http.MethodGet, "/admin/lenses", nil)
	s.Require().NoError(err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	s.Require().Equal( http.StatusOK, rr.Code)

	body, err := io.ReadAll(rr.Body)
	s.Require().NoError(err)

	s.Contains( string(body), "iPhone")
	s.Contains( string(body), "X100F")
}

func (s *EndpointsLensesSuite) TestGetLens() {
	testData := []models.Lens{
		{
			Name: "iPhone",
		},
	}

	persistedLenses, err := database.CreateLenses(s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create lenses: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/admin/lenses/{lensID}",
		BuildGetHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodGet)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/admin/lenses/%d", persistedLenses[0].ID), nil)
	s.Require().NoError(err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if !s.Equal( http.StatusOK, rr.Code) {
		bodyString, err := io.ReadAll(rr.Body)
		s.Require().NoError(err)
		s.T().Fatalf("request failed with: %s", bodyString)
	}

	body, err := io.ReadAll(rr.Body)
	s.Require().NoError(err)

	s.Contains( string(body), "iPhone")
}

func (s *EndpointsLensesSuite) TestUpdateLens() {
	testData := []models.Lens{
		{
			Name:        "iPhone",
			LensMatches: "",
		},
	}

	persistedLenses, err := database.CreateLenses(s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create lenses: %s", err)
	}

	// store the an icon in the bucket, check it's deleted
	imageIconPath := "../../../pkg/server/handlers/admin/lenses/testdata/fisheye.png"
	imageFile, err := os.Open(imageIconPath)
	s.Require().NoError(err)
	bw, err := s.Bucket.NewWriter(context.Background(), "lens_icons/iphone.jpg", nil)
	s.Require().NoError(err)
	_, err = io.Copy(bw, imageFile)
	s.Require().NoError(err)
	err = bw.Close()
	s.Require().NoError(err)

	router := mux.NewRouter()
	router.HandleFunc("/admin/lenses/{lensID}",
		BuildFormHandler(s.DB, s.Bucket, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodPost)

	// open the image to be uploaded in the form

	// build the form to be posted
	values := map[string]io.Reader{
		"Name":        strings.NewReader("iPad"),
		"LensMatches": strings.NewReader("wow"),
		"_method":     strings.NewReader(http.MethodPut),
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

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("/admin/lenses/%d", persistedLenses[0].ID),
		&b,
	)
	s.Require().NoError(err)

	req.Header.Set("Content-Type", w.FormDataContentType())

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if !s.Equal( http.StatusSeeOther, rr.Code) {
		bodyString, err := io.ReadAll(rr.Body)
		s.Require().NoError(err)
		s.T().Fatalf("request failed with: %s", bodyString)
	}

	// check that the database content is also correct
	returnedLenses, err := database.AllLenses(s.DB)
	if err != nil {
		s.T().Fatalf("failed to list lenses: %s", err)
	}
	expectedLenses := td.Slice(
		[]models.Lens{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Lens{
					ID:          persistedLenses[0].ID,
					Name:        "iPad",
					LensMatches: "wow",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedLenses, expectedLenses)
}

func (s *EndpointsLensesSuite) TestDeleteLens() {
	testData := []models.Lens{
		{Name: "iPhone"},
	}

	persistedLenses, err := database.CreateLenses(s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create lenses: %s", err)
	}

	// store the an icon in the bucket, check it's deleted
	imageIconPath := "../../../pkg/server/handlers/admin/lenses/testdata/fisheye.png"
	imageFile, err := os.Open(imageIconPath)
	s.Require().NoError(err)
	bw, err := s.Bucket.NewWriter(context.Background(), fmt.Sprintf("lens_icons/%d.png", persistedLenses[0].ID), nil)
	s.Require().NoError(err)
	_, err = io.Copy(bw, imageFile)
	s.Require().NoError(err)
	err = bw.Close()
	s.Require().NoError(err)

	router := mux.NewRouter()
	router.HandleFunc(
		"/admin/lenses/{lensID}",
		BuildFormHandler(s.DB, s.Bucket, templating.BuildPageRenderFunc(true, "")),
	).Methods(http.MethodPost)

	form := url.Values{}
	form.Add("_method", http.MethodDelete)

	// make the request to the handler
	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("/admin/lenses/%d", persistedLenses[0].ID),
		strings.NewReader(form.Encode()),
	)
	s.Require().NoError(err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if !s.Equal( http.StatusSeeOther, rr.Code) {
		bodyString, err := io.ReadAll(rr.Body)
		s.Require().NoError(err)
		s.T().Fatalf("request failed with: %s", bodyString)
	}

	// check that the database content is also correct
	returnedLenses, err := database.AllLenses(s.DB)
	if err != nil {
		s.T().Fatalf("failed to list lenses: %s", err)
	}

	expectedLenses := []models.Lens{}
	td.Cmp(s.T(), returnedLenses, expectedLenses)

	// should have a not found error as the icon should have been deleted
	_, err = s.Bucket.Attributes(context.Background(), "lens_icons/iphone.jpg")
	s.Require().Error(err)
}

func (s *EndpointsLensesSuite) TestNewLens() {
	router := mux.NewRouter()
	router.HandleFunc("/admin/lenses/new",
		BuildNewHandler(templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodGet)

	req, err := http.NewRequest(http.MethodGet, "/admin/lenses/new", nil)
	s.Require().NoError(err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	s.Require().Equal( http.StatusOK, rr.Code)

	body, err := io.ReadAll(rr.Body)
	s.Require().NoError(err)

	s.Contains( string(body), "Name")
	s.Contains( string(body), "Icon")
}

func (s *EndpointsLensesSuite) TestCreateLens() {
	router := mux.NewRouter()
	router.HandleFunc("/admin/lenses",
		BuildCreateHandler(s.DB, s.Bucket, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodPost)

	// open the image to be uploaded in the form
	imageIconPath := "../../../pkg/server/handlers/admin/lenses/testdata/fisheye.png"
	imageFile, err := os.Open(imageIconPath)
	s.Require().NoError(err)

	// build the form to be posted
	values := map[string]io.Reader{
		"Icon":        imageFile,
		"Name":        strings.NewReader("X100F"),
		"LensMatches": strings.NewReader("wow"),
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
		"/admin/lenses",
		&b,
	)
	s.Require().NoError(err)

	req.Header.Set("Content-Type", w.FormDataContentType())

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// check that we get a see other response to the right location
	if !s.Equal( http.StatusSeeOther, rr.Code) {
		bodyString, err := io.ReadAll(rr.Body)
		s.Require().NoError(err)
		s.T().Fatalf("request failed with: %s", bodyString)
	}
	if !strings.HasPrefix(rr.Result().Header["Location"][0], "/admin/lenses") {
		s.T().Fatalf("%v doesn't appear to be the correct path", rr.Result().Header["Location"])
	}

	// check that the database content is also correct
	returnedLenses, err := database.AllLenses(s.DB)
	s.Require().NoError(err)

	expectedLenses := td.Slice(
		[]models.Lens{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Lens{
					Name:        "X100F",
					LensMatches: "wow",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedLenses, expectedLenses)

	// check that the image has been uploaded ok
	// get a digest for the image in the bucket
	r, err := s.Bucket.NewReader(context.Background(), fmt.Sprintf("lens_icons/%d.png", returnedLenses[0].ID), nil)
	s.Require().NoError(err)
	defer r.Close()
	bucketHash := md5.New()
	_, err = io.Copy(bucketHash, r)
	s.Require().NoError(err)
	bucketMD5 := fmt.Sprintf("%x", bucketHash.Sum(nil))

	// get a digest for the image originally uploaded
	f, err := os.Open(imageIconPath)
	s.Require().NoError(err)
	defer f.Close()
	sourceHash := md5.New()
	_, err = io.Copy(sourceHash, f)
	s.Require().NoError(err)
	sourceMD5 := fmt.Sprintf("%x", bucketHash.Sum(nil))

	s.Require().Equal( bucketMD5, sourceMD5)
}
