package trips

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/suite"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/memblob"

	"github.com/charlieegan3/photos/internal/pkg/database"
	"github.com/charlieegan3/photos/internal/pkg/models"
	"github.com/charlieegan3/photos/internal/pkg/server/templating"
)

type EndpointsTripsSuite struct {
	suite.Suite
	DB *sql.DB
}

func (s *EndpointsTripsSuite) SetupTest() {
	err := database.Truncate(s.T().Context(), s.DB, "photos.trips")
	s.Require().NoError(err)
}

func (s *EndpointsTripsSuite) TestListTrips() {
	testData := []models.Trip{
		{
			Title:       "London",
			Description: "A trip to London",
			StartDate:   time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			Title:       "New York",
			Description: "A trip to New York",
			StartDate:   time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2020, 1, 4, 0, 0, 0, 0, time.UTC),
		},
	}

	_, err := database.CreateTrips(s.T().Context(), s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create trips: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/admin/trips",
		BuildIndexHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodGet)

	req, err := http.NewRequestWithContext(s.T().Context(), http.MethodGet, "/admin/trips", nil)
	s.Require().NoError(err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if !s.Equal(http.StatusOK, rr.Code) {
		s.T().Log(rr.Body.String())
		s.T().Fatalf("failed to read response body: %s", err)
	}

	body, err := io.ReadAll(rr.Body)
	s.Require().NoError(err)

	s.Contains(string(body), "London")
	s.Contains(string(body), "2020-01")
	s.Contains(string(body), "New York")
}

func (s *EndpointsTripsSuite) TestGetTrip() {
	testData := []models.Trip{
		{
			Title: "Inverness",
		},
	}

	persistedTrips, err := database.CreateTrips(s.T().Context(), s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create trips: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/admin/trips/{tripID}",
		BuildGetHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodGet)

	req, err := http.NewRequestWithContext(s.T().Context(), http.MethodGet, fmt.Sprintf("/admin/trips/%d", persistedTrips[0].ID), nil)
	s.Require().NoError(err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	s.Require().Equal(http.StatusOK, rr.Code)

	body, err := io.ReadAll(rr.Body)
	s.Require().NoError(err)

	s.Contains(string(body), "Inverness")
}

func (s *EndpointsTripsSuite) TestNewTrip() {
	router := mux.NewRouter()
	router.HandleFunc("/admin/trips/new",
		BuildNewHandler(templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodGet)

	req, err := http.NewRequestWithContext(s.T().Context(), http.MethodGet, "/admin/trips/new", nil)
	s.Require().NoError(err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	s.Require().Equal(http.StatusOK, rr.Code)

	body, err := io.ReadAll(rr.Body)
	s.Require().NoError(err)

	s.Contains(string(body), "Title")
}

func (s *EndpointsTripsSuite) TestCreateTrip() {
	router := mux.NewRouter()
	router.HandleFunc(
		"/admin/trips",
		BuildCreateHandler(s.DB, templating.BuildPageRenderFunc(true, "")),
	).Methods(http.MethodPost)

	form := url.Values{}
	form.Add("Title", "London")
	form.Add("Description", "Desc")
	form.Add("StartDate", "2023-01-01")
	form.Add("EndDate", "2023-01-02")

	// make the request to the handler
	req, err := http.NewRequestWithContext(
		s.T().Context(),
		http.MethodPost,
		"/admin/trips",
		strings.NewReader(form.Encode()),
	)
	s.Require().NoError(err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// check that we get a see other response to the right trip
	if !s.Equal(http.StatusSeeOther, rr.Code) {
		s.T().Log(rr.Body.String())
	}
	s.Require().Equal(http.StatusSeeOther, rr.Code)
	if !strings.HasPrefix(rr.Result().Header["Location"][0], "/admin/trips/") {
		s.T().Fatalf("%v doesn't appear to be the correct path", rr.Result().Header["Location"])
	}

	// check that the database content is also correct
	returnedTrips, err := database.AllTrips(s.T().Context(), s.DB)
	s.Require().NoError(err)

	expectedTrips := td.Slice(
		[]models.Trip{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Trip{
					Title:       "London",
					Description: "Desc",
					StartDate:   time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
					EndDate:     time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedTrips, expectedTrips)
}

func (s *EndpointsTripsSuite) TestUpdateTrip() {
	testData := []models.Trip{
		{
			Title: "Paris",
		},
	}

	persistedTrips, err := database.CreateTrips(s.T().Context(), s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create trips: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/admin/trips/{tripID}",
		BuildFormHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodPost)

	form := url.Values{}
	form.Add("_method", http.MethodPut)
	form.Add("Title", "Berlin")
	form.Add("StartDate", "2023-01-01")
	form.Add("EndDate", "2023-01-02")

	// make the request to the handler
	req, err := http.NewRequestWithContext(
		s.T().Context(),
		http.MethodPost,
		fmt.Sprintf("/admin/trips/%d", persistedTrips[0].ID),
		strings.NewReader(form.Encode()),
	)
	s.Require().NoError(err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if !s.Equal(http.StatusSeeOther, rr.Code) {
		s.T().Log(rr.Body.String())
	}
	s.Require().Equal(http.StatusSeeOther, rr.Code)

	// check that the database content is also correct
	returnedTrips, err := database.AllTrips(s.T().Context(), s.DB)
	if err != nil {
		s.T().Fatalf("failed to list trips: %s", err)
	}
	expectedTrips := td.Slice(
		[]models.Trip{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Trip{
					ID:    persistedTrips[0].ID,
					Title: "Berlin",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedTrips, expectedTrips)
}

func (s *EndpointsTripsSuite) TestDeleteTrip() {
	testData := []models.Trip{
		{
			Title: "nofilter",
		},
	}

	persistedTrips, err := database.CreateTrips(s.T().Context(), s.DB, testData)
	if err != nil {
		s.T().Fatalf("failed to create trips: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc(
		"/admin/trips/{tripID}",
		BuildFormHandler(s.DB, templating.BuildPageRenderFunc(true, "")),
	).Methods(http.MethodPost)

	form := url.Values{}
	form.Add("_method", http.MethodDelete)

	// make the request to the handler
	req, err := http.NewRequestWithContext(
		s.T().Context(),
		http.MethodPost,
		fmt.Sprintf("/admin/trips/%d", persistedTrips[0].ID),
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
	returnedTrips, err := database.AllTrips(s.T().Context(), s.DB)
	if err != nil {
		s.T().Fatalf("failed to list trips: %s", err)
	}

	expectedTrips := []models.Trip{}
	td.Cmp(s.T(), returnedTrips, expectedTrips)
}
