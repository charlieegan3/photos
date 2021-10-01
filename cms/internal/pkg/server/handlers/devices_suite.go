package handlers

import (
	"database/sql"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/charlieegan3/photos/cms/internal/pkg/database"
	"github.com/charlieegan3/photos/cms/internal/pkg/models"
	"github.com/gorilla/mux"
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
