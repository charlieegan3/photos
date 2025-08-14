package locations

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
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

type EndpointsLocationsSuite struct {
	suite.Suite

	DB            *sql.DB
	Bucket        *blob.Bucket
	BucketBaseURL string
}

func (s *EndpointsLocationsSuite) SetupTest() {
	var err error
	err = database.Truncate(s.T().Context(), s.DB, "photos.locations")
	s.Require().NoError(err)
	err = database.Truncate(s.T().Context(), s.DB, "photos.medias")
	s.Require().NoError(err)
	err = database.Truncate(s.T().Context(), s.DB, "photos.devices")
	s.Require().NoError(err)
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

	_, err := database.CreateLocations(s.T().Context(), s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create locations: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/admin/locations",
		BuildIndexHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodGet)

	req, err := http.NewRequestWithContext(s.T().Context(), http.MethodGet, "/admin/locations", nil)
	s.Require().NoError(err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	s.Require().Equal(http.StatusOK, rr.Code)

	body, err := io.ReadAll(rr.Body)
	s.Require().NoError(err)

	s.Contains(string(body), "London")
	s.Contains(string(body), "1.1")
	s.Contains(string(body), "New York")
	s.Contains(string(body), "1.3")
}

func (s *EndpointsLocationsSuite) TestGetLocation() {
	testData := []models.Location{
		{
			Name:      "Inverness",
			Longitude: 1.1,
			Latitude:  1.2,
		},
	}

	persistedLocations, err := database.CreateLocations(s.T().Context(), s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create locations: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/admin/locations/{locationID}",
		BuildGetHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodGet)

	req, err := http.NewRequestWithContext(
		s.T().Context(), http.MethodGet, fmt.Sprintf("/admin/locations/%d", persistedLocations[0].ID), nil,
	)
	s.Require().NoError(err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	s.Require().Equal(http.StatusOK, rr.Code)

	body, err := io.ReadAll(rr.Body)
	s.Require().NoError(err)

	s.Contains(string(body), "Inverness")
	s.Contains(string(body), "1.1")
	s.Contains(string(body), "1.2")
}

func (s *EndpointsLocationsSuite) TestNewLocation() {
	router := mux.NewRouter()
	router.HandleFunc("/admin/locations/new",
		BuildNewHandler(templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodGet)

	req, err := http.NewRequestWithContext(s.T().Context(), http.MethodGet, "/admin/locations/new", nil)
	s.Require().NoError(err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	s.Require().Equal(http.StatusOK, rr.Code)

	body, err := io.ReadAll(rr.Body)
	s.Require().NoError(err)

	s.Contains(string(body), "Name")
}

func (s *EndpointsLocationsSuite) TestCreateLocation() {
	router := mux.NewRouter()
	router.HandleFunc("/admin/locations",
		BuildCreateHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodPost)

	form := url.Values{}
	form.Add("Name", "London")

	// make the request to the handler
	req, err := http.NewRequestWithContext(
		s.T().Context(),
		http.MethodPost,
		"/admin/locations",
		strings.NewReader(form.Encode()),
	)
	s.Require().NoError(err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// check that we get a see other response to the right location
	s.Require().Equal(http.StatusSeeOther, rr.Code)
	if !strings.HasPrefix(rr.Result().Header["Location"][0], "/admin/locations/") {
		s.T().Fatalf("%v doesn't appear to be the correct path", rr.Result().Header["Location"])
	}

	// check that the database content is also correct
	returnedLocations, err := database.AllLocations(s.T().Context(), s.DB)
	s.Require().NoError(err)

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

	persistedLocations, err := database.CreateLocations(s.T().Context(), s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create locations: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/admin/locations/{locationID}",
		BuildFormHandler(s.DB, s.Bucket, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodPost)

	form := url.Values{}
	form.Add("_method", http.MethodPut)
	form.Add("Name", "Berlin")
	form.Add("Longitude", "1.1")

	// make the request to the handler
	req, err := http.NewRequestWithContext(
		s.T().Context(),
		http.MethodPost,
		fmt.Sprintf("/admin/locations/%d", persistedLocations[0].ID),
		strings.NewReader(form.Encode()),
	)
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	s.Require().Equal(http.StatusSeeOther, rr.Code)

	// check that the database content is also correct
	returnedLocations, err := database.AllLocations(s.T().Context(), s.DB)
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
	returnedDevices, err := database.CreateDevices(s.T().Context(), s.DB, devices)
	s.Require().NoError(err)

	medias := []models.Media{
		{DeviceID: returnedDevices[0].ID, Orientation: 1},
	}
	returnedMedias, err := database.CreateMedias(s.T().Context(), s.DB, medias)
	s.Require().NoError(err)

	locations := []models.Location{
		{Name: "Paris"},
		{Name: "Berlin"},
	}
	returnedLocations, err := database.CreateLocations(s.T().Context(), s.DB, locations)
	s.Require().NoError(err)

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
	returnedPosts, err := database.CreatePosts(s.T().Context(), s.DB, posts)
	s.Require().NoError(err)

	router := mux.NewRouter()
	router.HandleFunc("/admin/locations/{locationID}",
		BuildFormHandler(s.DB, s.Bucket, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodPost)

	form := url.Values{}
	form.Add("_method", http.MethodPut)
	form.Add("Name", "Berlin")

	// make the request to the handler
	req, err := http.NewRequestWithContext(
		s.T().Context(),
		http.MethodPost,
		fmt.Sprintf("/admin/locations/%d", returnedLocations[0].ID),
		strings.NewReader(form.Encode()),
	)
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	s.Require().Equal(http.StatusSeeOther, rr.Code)

	// check that the database content is also correct
	endingLocations, err := database.AllLocations(s.T().Context(), s.DB)
	s.Require().NoError(err)

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

	endingPosts, err := database.AllPosts(s.T().Context(), s.DB, false, database.SelectOptions{SortField: "id"})
	s.Require().NoError(err)

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

	persistedLocations, err := database.CreateLocations(s.T().Context(), s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create locations: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc(
		"/admin/locations/{locationID}",
		BuildFormHandler(s.DB, s.Bucket, templating.BuildPageRenderFunc(true, "")),
	).Methods(http.MethodPost)

	form := url.Values{}
	form.Add("_method", http.MethodDelete)

	// make the request to the handler
	req, err := http.NewRequestWithContext(
		s.T().Context(),
		http.MethodPost,
		fmt.Sprintf("/admin/locations/%d", persistedLocations[0].ID),
		strings.NewReader(form.Encode()),
	)
	s.Require().NoError(err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if !s.Equal(http.StatusSeeOther, rr.Code) {
		bodyString, err := io.ReadAll(rr.Body)
		s.Require().NoError(err)
		s.T().Fatalf("request failed with: %s", bodyString)
	}

	// check that the database content is also correct
	returnedLocations, err := database.AllLocations(s.T().Context(), s.DB)
	if err != nil {
		s.T().Fatalf("failed to list locations: %s", err)
	}

	expectedLocations := []models.Location{}
	td.Cmp(s.T(), returnedLocations, expectedLocations)
}

func (s *EndpointsLocationsSuite) TestLocationSelector() {
	testDataLocations := []models.Location{
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

	persistedLocations, err := database.CreateLocations(s.T().Context(), s.DB, testDataLocations)
	if err != nil {
		s.T().Fatalf("failed to create locations: %s", err)
	}

	persistedDevices, err := database.CreateDevices(s.T().Context(), s.DB, []models.Device{
		{
			Name: "Example",
			Slug: "example",
		},
	})
	if err != nil {
		s.T().Fatalf("failed to create devices: %s", err)
	}

	testDataMedias := []models.Media{
		{
			DeviceID:  persistedDevices[0].ID,
			Latitude:  1.1,
			Longitude: 1.1,
			Orientation: 1,
		},
	}

	persistedMedias, err := database.CreateMedias(s.T().Context(), s.DB, testDataMedias)
	if err != nil {
		s.T().Fatalf("failed to create medias: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/admin/locations/select",
		BuildSelectHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodGet)

	req, err := http.NewRequestWithContext(
		s.T().Context(),
		http.MethodGet,
		"/admin/locations/select?redirectTo=%2Fadmin%2Fposts%2Fnew&param1=1&param2=2&timestamp=1234&mediaID="+
			strconv.Itoa(persistedMedias[0].ID),
		nil,
	)
	s.Require().NoError(err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	body, err := io.ReadAll(rr.Body)
	s.Require().NoError(err)

	if !s.Equal(http.StatusOK, rr.Code) {
		s.T().Fatalf("request failed with: %s", string(body))
	}

	s.Contains(string(body), "London")
	s.Contains(string(body), "New York")

	s.Contains(string(body), fmt.Sprintf(
		`/admin/posts/new?mediaID=%d&param1=1&param2=2&timestamp=1234&locationID=%d`,
		persistedMedias[0].ID,
		persistedLocations[0].ID,
	))
}
