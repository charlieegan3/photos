package devices

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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/memblob"

	"github.com/charlieegan3/photos/internal/pkg/database"
	"github.com/charlieegan3/photos/internal/pkg/models"
	"github.com/charlieegan3/photos/internal/pkg/server/templating"
)

type EndpointsDevicesSuite struct {
	suite.Suite
	DB            *sql.DB
	Bucket        *blob.Bucket
	BucketBaseURL string
}

func (s *EndpointsDevicesSuite) SetupTest() {
	err := database.Truncate(s.DB, "photos.devices")
	require.NoError(s.T(), err)
}

func (s *EndpointsDevicesSuite) TestListDevices() {
	testData := []models.Device{
		{
			Name: "iPhone",
		},
		{
			Name: "X100F",
		},
	}

	_, err := database.CreateDevices(s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create devices: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/admin/devices",
		BuildIndexHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodGet)

	req, err := http.NewRequest(http.MethodGet, "/admin/devices", nil)
	require.NoError(s.T(), err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(s.T(), http.StatusOK, rr.Code)

	body, err := io.ReadAll(rr.Body)
	require.NoError(s.T(), err)

	assert.Contains(s.T(), string(body), "iPhone")
	assert.Contains(s.T(), string(body), "X100F")
}

func (s *EndpointsDevicesSuite) TestGetDevice() {
	testData := []models.Device{
		{
			Name: "iPhone",
		},
	}

	persistedDevices, err := database.CreateDevices(s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create devices: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/admin/devices/{deviceID}",
		BuildGetHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).Methods(http.MethodGet)

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/admin/devices/%d", persistedDevices[0].ID), nil)
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

	assert.Contains(s.T(), string(body), "iPhone")
}

func (s *EndpointsDevicesSuite) TestUpdateDevice() {
	testData := []models.Device{
		{
			Name:     "iPhone",
			IconKind: "jpg",
		},
	}

	persistedDevices, err := database.CreateDevices(s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create devices: %s", err)
	}

	// store the an icon in the bucket, check it's deleted
	imageIconPath := "../../../pkg/server/handlers/admin/devices/testdata/x100f.jpg"
	imageFile, err := os.Open(imageIconPath)
	require.NoError(s.T(), err)
	bw, err := s.Bucket.NewWriter(context.Background(), "device_icons/iphone.jpg", nil)
	require.NoError(s.T(), err)
	_, err = io.Copy(bw, imageFile)
	require.NoError(s.T(), err)
	err = bw.Close()
	require.NoError(s.T(), err)

	router := mux.NewRouter()
	router.HandleFunc("/admin/devices/{deviceID}",
		BuildFormHandler(s.DB, s.Bucket, templating.BuildPageRenderFunc(true, ""))).Methods(http.MethodPost)

	// open the image to be uploaded in the form

	// build the form to be posted
	values := map[string]io.Reader{
		"Name":    strings.NewReader("iPad"),
		"_method": strings.NewReader(http.MethodPut),
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

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("/admin/devices/%d", persistedDevices[0].ID),
		&b,
	)
	require.NoError(s.T(), err)

	req.Header.Set("Content-Type", w.FormDataContentType())

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if !assert.Equal(s.T(), http.StatusSeeOther, rr.Code) {
		bodyString, err := io.ReadAll(rr.Body)
		require.NoError(s.T(), err)
		s.T().Fatalf("request failed with: %s", bodyString)
	}

	// check that the database content is also correct
	returnedDevices, err := database.AllDevices(s.DB)
	if err != nil {
		s.T().Fatalf("failed to list devices: %s", err)
	}
	expectedDevices := td.Slice(
		[]models.Device{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Device{
					ID:   persistedDevices[0].ID,
					Name: "iPad",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedDevices, expectedDevices)
}

func (s *EndpointsDevicesSuite) TestDeleteDevice() {
	testData := []models.Device{
		{
			Name:     "iPhone",
			IconKind: "jpg",
		},
	}

	persistedDevices, err := database.CreateDevices(s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create devices: %s", err)
	}

	// store the an icon in the bucket, check it's deleted
	imageIconPath := "../../../pkg/server/handlers/admin/devices/testdata/x100f.jpg"
	imageFile, err := os.Open(imageIconPath)
	require.NoError(s.T(), err)
	bw, err := s.Bucket.NewWriter(context.Background(), "device_icons/iphone.jpg", nil)
	require.NoError(s.T(), err)
	_, err = io.Copy(bw, imageFile)
	require.NoError(s.T(), err)
	err = bw.Close()
	require.NoError(s.T(), err)

	router := mux.NewRouter()
	router.HandleFunc(
		"/admin/devices/{deviceID}",
		BuildFormHandler(s.DB, s.Bucket, templating.BuildPageRenderFunc(true, "")),
	).Methods(http.MethodPost)

	form := url.Values{}
	form.Add("_method", http.MethodDelete)

	// make the request to the handler
	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("/admin/devices/%d", persistedDevices[0].ID),
		strings.NewReader(form.Encode()),
	)
	require.NoError(s.T(), err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if !assert.Equal(s.T(), http.StatusSeeOther, rr.Code) {
		bodyString, err := io.ReadAll(rr.Body)
		require.NoError(s.T(), err)
		s.T().Fatalf("request failed with: %s", bodyString)
	}

	// check that the database content is also correct
	returnedDevices, err := database.AllDevices(s.DB)
	if err != nil {
		s.T().Fatalf("failed to list devices: %s", err)
	}

	expectedDevices := []models.Device{}
	td.Cmp(s.T(), returnedDevices, expectedDevices)

	// should have a not found error as the icon should have been deleted
	_, err = s.Bucket.Attributes(context.Background(), "device_icons/iphone.jpg")
	require.Error(s.T(), err)
}

func (s *EndpointsDevicesSuite) TestNewDevice() {
	router := mux.NewRouter()
	router.HandleFunc("/admin/devices/new",
		BuildNewHandler(templating.BuildPageRenderFunc(true, ""))).Methods(http.MethodGet)

	req, err := http.NewRequest(http.MethodGet, "/admin/devices/new", nil)
	require.NoError(s.T(), err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(s.T(), http.StatusOK, rr.Code)

	body, err := io.ReadAll(rr.Body)
	require.NoError(s.T(), err)

	assert.Contains(s.T(), string(body), "Name")
	assert.Contains(s.T(), string(body), "Icon")
}

func (s *EndpointsDevicesSuite) TestCreateDevice() {
	router := mux.NewRouter()
	router.HandleFunc("/admin/devices",
		BuildCreateHandler(s.DB, s.Bucket, templating.BuildPageRenderFunc(true, ""))).Methods(http.MethodPost)

	// open the image to be uploaded in the form
	imageIconPath := "../../../pkg/server/handlers/admin/devices/testdata/x100f.jpg"
	imageFile, err := os.Open(imageIconPath)
	require.NoError(s.T(), err)

	// build the form to be posted
	values := map[string]io.Reader{
		"Icon": imageFile,
		"Name": strings.NewReader("X100F"),
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
		http.MethodPost,
		"/admin/devices",
		&b,
	)
	require.NoError(s.T(), err)

	req.Header.Set("Content-Type", w.FormDataContentType())

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// check that we get a see other response to the right location
	require.Equal(s.T(), http.StatusSeeOther, rr.Code)
	if !strings.HasPrefix(rr.Result().Header["Location"][0], "/admin/devices/") {
		s.T().Fatalf("%v doesn't appear to be the correct path", rr.Result().Header["Location"])
	}

	// check that the database content is also correct
	returnedDevices, err := database.AllDevices(s.DB)
	require.NoError(s.T(), err)

	expectedDevices := td.Slice(
		[]models.Device{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Device{
					Name: "X100F",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedDevices, expectedDevices)

	// check that the image has been uploaded ok
	// get a digest for the image in the bucket
	r, err := s.Bucket.NewReader(context.Background(), "device_icons/x100f.jpg", nil)
	require.NoError(s.T(), err)
	defer r.Close()
	bucketHash := md5.New()
	_, err = io.Copy(bucketHash, r)
	require.NoError(s.T(), err)
	bucketMD5 := fmt.Sprintf("%x", bucketHash.Sum(nil))

	// get a digest for the image originally uploaded
	f, err := os.Open(imageIconPath)
	require.NoError(s.T(), err)
	defer f.Close()
	sourceHash := md5.New()
	_, err = io.Copy(sourceHash, f)
	require.NoError(s.T(), err)
	sourceMD5 := fmt.Sprintf("%x", bucketHash.Sum(nil))

	require.Equal(s.T(), bucketMD5, sourceMD5)
}
