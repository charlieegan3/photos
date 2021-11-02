package handlers

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/charlieegan3/photos/cms/internal/pkg/database"
	"github.com/charlieegan3/photos/cms/internal/pkg/models"
	"github.com/gorilla/mux"
	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type EndpointsDevicesSuite struct {
	suite.Suite
	DB *sql.DB
}

func (s *EndpointsDevicesSuite) SetupTest() {
	err := database.Truncate(s.DB, "devices")
	require.NoError(s.T(), err)
}

func (s *EndpointsDevicesSuite) TestListDevices() {
	testData := []models.Device{
		{
			Name:    "iPhone",
			IconURL: "https://example.com/image.jpg",
		},
		{
			Name:    "X100F",
			IconURL: "https://example.com/image2.jpg",
		},
	}

	_, err := database.CreateDevices(s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create devices: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/admin/devices", BuildIndexHandler(s.DB)).Methods("GET")

	req, err := http.NewRequest("GET", "/admin/devices", nil)
	require.NoError(s.T(), err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(s.T(), http.StatusOK, rr.Code)

	body, err := ioutil.ReadAll(rr.Body)
	require.NoError(s.T(), err)

	assert.Contains(s.T(), string(body), "iPhone")
	assert.Contains(s.T(), string(body), "X100F")
}

func (s *EndpointsDevicesSuite) TestGetDevice() {
	testData := []models.Device{
		{
			Name:    "iPhone",
			IconURL: "https://example.com/image.jpg",
		},
	}

	persistedDevices, err := database.CreateDevices(s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create devices: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/admin/devices/{deviceName}", BuildGetHandler(s.DB)).Methods("GET")

	req, err := http.NewRequest("GET", fmt.Sprintf("/admin/devices/%s", persistedDevices[0].Name), nil)
	require.NoError(s.T(), err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(s.T(), http.StatusOK, rr.Code)

	body, err := ioutil.ReadAll(rr.Body)
	require.NoError(s.T(), err)

	assert.Contains(s.T(), string(body), "iPhone")
}

func (s *EndpointsDevicesSuite) TestUpdateDevice() {
	testData := []models.Device{
		{
			Name:    "iPhone",
			IconURL: "https://example.com/image.jpg",
		},
	}

	persistedDevices, err := database.CreateDevices(s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create devices: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/admin/devices/{deviceName}", BuildFormHandler(s.DB)).Methods("POST")

	form := url.Values{}
	form.Add("_method", "PUT")
	form.Add("Name", "iPad")
	form.Add("IconURL", "https://example.com/image.jpg")

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("/admin/devices/%s", persistedDevices[0].Name),
		strings.NewReader(form.Encode()),
	)
	require.NoError(s.T(), err)

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(s.T(), http.StatusSeeOther, rr.Code)

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
					ID:      persistedDevices[0].ID,
					Name:    "iPad",
					IconURL: "https://example.com/image.jpg",
				},
				td.StructFields{
					"CreatedAt": td.Ignore(),
					"UpdatedAt": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedDevices, expectedDevices)
}

func (s *EndpointsDevicesSuite) TestDeleteDevice() {
	fmt.Println("------start")
	testData := []models.Device{
		{
			Name:    "iPhone",
			IconURL: "https://example.com/image.jpg",
		},
	}

	persistedDevices, err := database.CreateDevices(s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create devices: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/admin/devices/{deviceName}", BuildFormHandler(s.DB)).Methods("POST")

	form := url.Values{}
	form.Add("_method", "DELETE")

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("/admin/devices/%s", persistedDevices[0].Name),
		strings.NewReader(form.Encode()),
	)
	require.NoError(s.T(), err)

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(s.T(), http.StatusSeeOther, rr.Code)

	// check that the database content is also correct
	returnedDevices, err := database.AllDevices(s.DB)
	if err != nil {
		s.T().Fatalf("failed to list devices: %s", err)
	}

	expectedDevices := []models.Device{}
	td.Cmp(s.T(), returnedDevices, expectedDevices)
	fmt.Println("------------end")
}

func (s *EndpointsDevicesSuite) TestNewDevice() {
	router := mux.NewRouter()
	router.HandleFunc("/admin/devices/new", BuildNewHandler()).Methods("GET")

	req, err := http.NewRequest("GET", "/admin/devices/new", nil)
	require.NoError(s.T(), err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(s.T(), http.StatusOK, rr.Code)

	body, err := ioutil.ReadAll(rr.Body)
	require.NoError(s.T(), err)

	assert.Contains(s.T(), string(body), "Name")
}

func (s *EndpointsDevicesSuite) TestCreateDevice() {
	router := mux.NewRouter()
	router.HandleFunc("/admin/devices", BuildCreateHandler(s.DB)).Methods("POST")

	form := url.Values{}
	form.Add("Name", "iPhone")
	form.Add("IconURL", "https://example.com/image.jpg")

	req, err := http.NewRequest(
		"POST",
		"/admin/devices",
		strings.NewReader(form.Encode()),
	)
	require.NoError(s.T(), err)

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// check that we get a see other to the right location
	require.Equal(s.T(), http.StatusSeeOther, rr.Code)
	td.Cmp(s.T(), rr.HeaderMap["Location"], []string{"/admin/devices"})

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
					Name:    "iPhone",
					IconURL: "https://example.com/image.jpg",
				},
				td.StructFields{
					"ID":        td.Ignore(),
					"CreatedAt": td.Ignore(),
					"UpdatedAt": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedDevices, expectedDevices)
}
