package locations

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
	"github.com/charlieegan3/photos/cms/internal/pkg/server/templating"
	"github.com/gorilla/mux"
	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/memblob"
)

type EndpointsLocationsSuite struct {
	suite.Suite
	DB            *sql.DB
	Bucket        *blob.Bucket
	BucketBaseURL string
}

func (s *EndpointsLocationsSuite) SetupTest() {
	err := database.Truncate(s.DB, "locations")
	require.NoError(s.T(), err)
}

func (s *EndpointsLocationsSuite) TestListLocations() {
	testData := []models.Location{
		{
			Name:      "London",
			Latitude:  1.1,
			Longitude: 1.2,
		},
		{
			Name:      "New York",
			Latitude:  1.3,
			Longitude: 1.4,
		},
	}

	_, err := database.CreateLocations(s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create locations: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/admin/locations", BuildIndexHandler(s.DB, templating.BuildPageRenderFunc("http://", ""))).Methods("GET")

	req, err := http.NewRequest("GET", "/admin/locations", nil)
	require.NoError(s.T(), err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(s.T(), http.StatusOK, rr.Code)

	body, err := ioutil.ReadAll(rr.Body)
	require.NoError(s.T(), err)

	assert.Contains(s.T(), string(body), "London")
	assert.Contains(s.T(), string(body), "1.1")
	assert.Contains(s.T(), string(body), "New York")
	assert.Contains(s.T(), string(body), "1.3")
}

func (s *EndpointsLocationsSuite) TestGetLocation() {
	testData := []models.Location{
		{
			Name:      "Inverness",
			Longitude: 1.1,
			Latitude:  1.2,
		},
	}

	persistedLocations, err := database.CreateLocations(s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create locations: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/admin/locations/{locationSlug}", BuildGetHandler(s.DB, templating.BuildPageRenderFunc("http://", ""))).Methods("GET")

	req, err := http.NewRequest("GET", fmt.Sprintf("/admin/locations/%s", persistedLocations[0].Slug), nil)
	require.NoError(s.T(), err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(s.T(), http.StatusOK, rr.Code)

	body, err := ioutil.ReadAll(rr.Body)
	require.NoError(s.T(), err)

	assert.Contains(s.T(), string(body), "Inverness")
	assert.Contains(s.T(), string(body), "1.1")
	assert.Contains(s.T(), string(body), "1.2")
}

func (s *EndpointsLocationsSuite) TestNewLocation() {
	router := mux.NewRouter()
	router.HandleFunc("/admin/locations/new", BuildNewHandler(templating.BuildPageRenderFunc("http://", ""))).Methods("GET")

	req, err := http.NewRequest("GET", "/admin/locations/new", nil)
	require.NoError(s.T(), err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	require.Equal(s.T(), http.StatusOK, rr.Code)

	body, err := ioutil.ReadAll(rr.Body)
	require.NoError(s.T(), err)

	assert.Contains(s.T(), string(body), "Name")
}

func (s *EndpointsLocationsSuite) TestCreateLocation() {
	router := mux.NewRouter()
	router.HandleFunc("/admin/locations", BuildCreateHandler(s.DB, templating.BuildPageRenderFunc("http://", ""))).Methods("POST")

	form := url.Values{}
	form.Add("Name", "London")

	// make the request to the handler
	req, err := http.NewRequest(
		"POST",
		"/admin/locations",
		strings.NewReader(form.Encode()),
	)
	require.NoError(s.T(), err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// check that we get a see other response to the right location
	require.Equal(s.T(), http.StatusSeeOther, rr.Code)
	td.Cmp(s.T(), rr.HeaderMap["Location"], []string{"/admin/locations/london"})

	// check that the database content is also correct
	returnedLocations, err := database.AllLocations(s.DB)
	require.NoError(s.T(), err)

	expectedLocations := td.Slice(
		[]models.Location{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Location{
					Name: "London",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedLocations, expectedLocations)
}

func (s *EndpointsLocationsSuite) TestUpdateLocation() {
	testData := []models.Location{
		{
			Name: "Paris",
		},
	}

	persistedLocations, err := database.CreateLocations(s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create locations: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/admin/locations/{locationSlug}", BuildFormHandler(s.DB, templating.BuildPageRenderFunc("http://", ""))).Methods("POST")

	form := url.Values{}
	form.Add("_method", "PUT")
	form.Add("Name", "Berlin")
	form.Add("Longitude", "1.1")

	// make the request to the handler
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("/admin/locations/%s", persistedLocations[0].Slug),
		strings.NewReader(form.Encode()),
	)
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(s.T(), http.StatusSeeOther, rr.Code)

	// check that the database content is also correct
	returnedLocations, err := database.AllLocations(s.DB)
	if err != nil {
		s.T().Fatalf("failed to list locations: %s", err)
	}
	expectedLocations := td.Slice(
		[]models.Location{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Location{
					ID:        persistedLocations[0].ID,
					Name:      "Berlin",
					Longitude: 1.1,
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedLocations, expectedLocations)
}

func (s *EndpointsLocationsSuite) TestDeleteLocation() {
	testData := []models.Location{
		{
			Name: "nofilter",
		},
	}

	persistedLocations, err := database.CreateLocations(s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create locations: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc(
		"/admin/locations/{locationSlug}",
		BuildFormHandler(s.DB, templating.BuildPageRenderFunc("http://", "")),
	).Methods("POST")

	form := url.Values{}
	form.Add("_method", "DELETE")

	// make the request to the handler
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("/admin/locations/%s", persistedLocations[0].Slug),
		strings.NewReader(form.Encode()),
	)
	require.NoError(s.T(), err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(s.T(), http.StatusSeeOther, rr.Code)

	// check that the database content is also correct
	returnedLocations, err := database.AllLocations(s.DB)
	if err != nil {
		s.T().Fatalf("failed to list locations: %s", err)
	}

	expectedLocations := []models.Location{}
	td.Cmp(s.T(), returnedLocations, expectedLocations)
}
