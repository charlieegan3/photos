package locations

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
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

	"github.com/charlieegan3/photos/internal/pkg/database"
	"github.com/charlieegan3/photos/internal/pkg/models"
	"github.com/charlieegan3/photos/internal/pkg/server/templating"
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
	router.HandleFunc("/admin/locations", BuildIndexHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).Methods("GET")

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
	router.HandleFunc("/admin/locations/{locationID}", BuildGetHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).Methods("GET")

	req, err := http.NewRequest("GET", fmt.Sprintf("/admin/locations/%d", persistedLocations[0].ID), nil)
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
	router.HandleFunc("/admin/locations/new", BuildNewHandler(templating.BuildPageRenderFunc(true, ""))).Methods("GET")

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
	router.HandleFunc("/admin/locations", BuildCreateHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).Methods("POST")

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
	if !strings.HasPrefix(rr.HeaderMap["Location"][0], "/admin/locations/") {
		s.T().Fatalf("%v doesn't appear to be the correct path", rr.HeaderMap["Location"])
	}

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
	router.HandleFunc("/admin/locations/{locationID}", BuildFormHandler(s.DB, s.Bucket, templating.BuildPageRenderFunc(true, ""))).Methods("POST")

	form := url.Values{}
	form.Add("_method", "PUT")
	form.Add("Name", "Berlin")
	form.Add("Longitude", "1.1")

	// make the request to the handler
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("/admin/locations/%d", persistedLocations[0].ID),
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

func (s *EndpointsLocationsSuite) TestUpdateLocationMergeName() {
	devices := []models.Device{
		{
			Name: "Example Device",
		},
	}
	returnedDevices, err := database.CreateDevices(s.DB, devices)
	require.NoError(s.T(), err)

	medias := []models.Media{
		{DeviceID: returnedDevices[0].ID},
	}
	returnedMedias, err := database.CreateMedias(s.DB, medias)
	require.NoError(s.T(), err)

	locations := []models.Location{
		{Name: "Paris"},
		{Name: "Berlin"},
	}
	returnedLocations, err := database.CreateLocations(s.DB, locations)
	require.NoError(s.T(), err)

	posts := []models.Post{
		{
			Description: "photo to move to berlin",
			PublishDate: time.Date(2021, time.November, 24, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
		},
		{
			Description: "photo already in berlin",
			PublishDate: time.Date(2021, time.November, 25, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[1].ID,
		},
	}
	returnedPosts, err := database.CreatePosts(s.DB, posts)
	require.NoError(s.T(), err)

	router := mux.NewRouter()
	router.HandleFunc("/admin/locations/{locationID}", BuildFormHandler(s.DB, s.Bucket, templating.BuildPageRenderFunc(true, ""))).Methods("POST")

	form := url.Values{}
	form.Add("_method", "PUT")
	form.Add("Name", "Berlin")

	// make the request to the handler
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("/admin/locations/%d", returnedLocations[0].ID),
		strings.NewReader(form.Encode()),
	)
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	require.Equal(s.T(), http.StatusSeeOther, rr.Code)

	// check that the database content is also correct
	endingLocations, err := database.AllLocations(s.DB)
	require.NoError(s.T(), err)

	expectedLocations := td.Slice(
		[]models.Location{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Location{
					ID:   returnedLocations[1].ID,
					Name: "Berlin",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)
	td.Cmp(s.T(), endingLocations, expectedLocations)

	endingPosts, err := database.AllPosts(s.DB, false, database.SelectOptions{SortField: "id"})
	require.NoError(s.T(), err)

	expectedPosts := td.Slice(
		[]models.Post{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Post{
					ID:         returnedPosts[0].ID,
					LocationID: returnedLocations[1].ID,
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
			1: td.SStruct(
				models.Post{
					ID:         returnedPosts[1].ID,
					LocationID: returnedLocations[1].ID,
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)
	td.Cmp(s.T(), endingPosts, expectedPosts)
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
		"/admin/locations/{locationID}",
		BuildFormHandler(s.DB, s.Bucket, templating.BuildPageRenderFunc(true, "")),
	).Methods("POST")

	form := url.Values{}
	form.Add("_method", "DELETE")

	// make the request to the handler
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("/admin/locations/%d", persistedLocations[0].ID),
		strings.NewReader(form.Encode()),
	)
	require.NoError(s.T(), err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if !assert.Equal(s.T(), http.StatusSeeOther, rr.Code) {
		bodyString, err := ioutil.ReadAll(rr.Body)
		require.NoError(s.T(), err)
		s.T().Fatalf("request failed with: %s", bodyString)
	}

	// check that the database content is also correct
	returnedLocations, err := database.AllLocations(s.DB)
	if err != nil {
		s.T().Fatalf("failed to list locations: %s", err)
	}

	expectedLocations := []models.Location{}
	td.Cmp(s.T(), returnedLocations, expectedLocations)
}

func (s *EndpointsLocationsSuite) TestLocationSelector() {
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

	persistedLocations, err := database.CreateLocations(s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create locations: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/admin/locations/select", BuildSelectHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).Methods("GET")

	req, err := http.NewRequest("GET", "/admin/locations/select?redirectTo=%2Fadmin%2Fposts%2Fnew&param1=1&param2=2&timestamp=1234", nil)
	require.NoError(s.T(), err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	body, err := ioutil.ReadAll(rr.Body)
	require.NoError(s.T(), err)

	if !assert.Equal(s.T(), http.StatusOK, rr.Code) {
		s.T().Fatalf("request failed with: %s", string(body))
	}

	assert.Contains(s.T(), string(body), "London")
	assert.Contains(s.T(), string(body), "New York")

	assert.Contains(s.T(), string(body), fmt.Sprintf(`<a href="/admin/posts/new?param1=1&param2=2&timestamp=1234&locationID=%d">`, persistedLocations[0].ID))
}
