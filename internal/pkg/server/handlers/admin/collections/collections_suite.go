package collections

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/suite"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/memblob"

	"github.com/charlieegan3/photos/internal/pkg/database"
	"github.com/charlieegan3/photos/internal/pkg/models"
	"github.com/charlieegan3/photos/internal/pkg/server/templating"
)

type EndpointsCollectionsSuite struct {
	suite.Suite

	DB *sql.DB
}

func (s *EndpointsCollectionsSuite) SetupTest() {
	err := database.Truncate(s.T().Context(), s.DB, "photos.collections")
	s.Require().NoError(err)
}

func (s *EndpointsCollectionsSuite) TestListCollections() {
	testData := []models.Collection{
		{
			Title:       "Nature Photography",
			Description: "Beautiful nature shots from various trips",
		},
		{
			Title:       "Street Photography",
			Description: "Urban street photography collection",
		},
	}

	repo := database.NewCollectionRepository(s.DB)
	_, err := repo.Create(s.T().Context(), testData)
	if err != nil {
		s.T().Fatalf("failed to create collections: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/admin/collections",
		BuildIndexHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodGet)

	req, err := http.NewRequestWithContext(s.T().Context(), http.MethodGet, "/admin/collections", nil)
	s.Require().NoError(err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if !s.Equal(http.StatusOK, rr.Code) {
		s.T().Log(rr.Body.String())
		s.T().Fatalf("failed to get OK response: %d", rr.Code)
	}

	body, err := io.ReadAll(rr.Body)
	s.Require().NoError(err)

	s.Contains(string(body), "Nature Photography")
	s.Contains(string(body), "Street Photography")
}

func (s *EndpointsCollectionsSuite) TestGetCollection() {
	testData := []models.Collection{
		{
			Title:       "Wildlife",
			Description: "Wild animals in their natural habitat",
		},
	}

	repo := database.NewCollectionRepository(s.DB)
	persistedCollections, err := repo.Create(s.T().Context(), testData)
	if err != nil {
		s.T().Fatalf("failed to create collections: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/admin/collections/{collectionID}",
		BuildGetHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodGet)

	req, err := http.NewRequestWithContext(
		s.T().Context(), http.MethodGet, fmt.Sprintf("/admin/collections/%d", persistedCollections[0].ID), nil,
	)
	s.Require().NoError(err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	s.Require().Equal(http.StatusOK, rr.Code)

	body, err := io.ReadAll(rr.Body)
	s.Require().NoError(err)

	s.Contains(string(body), "Wildlife")
	s.Contains(string(body), "Wild animals in their natural habitat")
}

func (s *EndpointsCollectionsSuite) TestNewCollection() {
	router := mux.NewRouter()
	router.HandleFunc("/admin/collections/new",
		BuildNewHandler(templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodGet)

	req, err := http.NewRequestWithContext(s.T().Context(), http.MethodGet, "/admin/collections/new", nil)
	s.Require().NoError(err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	s.Require().Equal(http.StatusOK, rr.Code)

	body, err := io.ReadAll(rr.Body)
	s.Require().NoError(err)

	s.Contains(string(body), "Title")
	s.Contains(string(body), "Description")
}

func (s *EndpointsCollectionsSuite) TestCreateCollection() {
	router := mux.NewRouter()
	router.HandleFunc(
		"/admin/collections",
		BuildCreateHandler(s.DB, templating.BuildPageRenderFunc(true, "")),
	).Methods(http.MethodPost)

	form := url.Values{}
	form.Add("Title", "Landscape Photography")
	form.Add("Description", "Beautiful landscapes from around the world")

	// make the request to the handler
	req, err := http.NewRequestWithContext(
		s.T().Context(),
		http.MethodPost,
		"/admin/collections",
		strings.NewReader(form.Encode()),
	)
	s.Require().NoError(err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// check that we get a see other response to the right collection
	if !s.Equal(http.StatusSeeOther, rr.Code) {
		s.T().Log(rr.Body.String())
	}
	s.Require().Equal(http.StatusSeeOther, rr.Code)
	if !strings.HasPrefix(rr.Result().Header["Location"][0], "/admin/collections/") {
		s.T().Fatalf("%v doesn't appear to be the correct path", rr.Result().Header["Location"])
	}

	// check that the database content is also correct
	repo := database.NewCollectionRepository(s.DB)
	returnedCollections, err := repo.All(s.T().Context())
	s.Require().NoError(err)

	expectedCollections := td.Slice(
		[]models.Collection{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Collection{
					Title:       "Landscape Photography",
					Description: "Beautiful landscapes from around the world",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedCollections, expectedCollections)
}

func (s *EndpointsCollectionsSuite) TestUpdateCollection() {
	testData := []models.Collection{
		{
			Title:       "Old Title",
			Description: "Old description",
		},
	}

	repo := database.NewCollectionRepository(s.DB)
	persistedCollections, err := repo.Create(s.T().Context(), testData)
	if err != nil {
		s.T().Fatalf("failed to create collections: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc(
		"/admin/collections/{collectionID}",
		BuildFormHandler(s.DB, templating.BuildPageRenderFunc(true, "")),
	).Methods(http.MethodPost)

	form := url.Values{}
	form.Add("_method", "PUT")
	form.Add("Title", "Updated Title")
	form.Add("Description", "Updated description")

	req, err := http.NewRequestWithContext(
		s.T().Context(),
		http.MethodPost,
		fmt.Sprintf("/admin/collections/%d", persistedCollections[0].ID),
		strings.NewReader(form.Encode()),
	)
	s.Require().NoError(err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	s.Require().Equal(http.StatusSeeOther, rr.Code)

	// check that the database content is updated
	returnedCollections, err := repo.All(s.T().Context())
	s.Require().NoError(err)

	expectedCollections := td.Slice(
		[]models.Collection{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Collection{
					Title:       "Updated Title",
					Description: "Updated description",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedCollections, expectedCollections)
}

func (s *EndpointsCollectionsSuite) TestDeleteCollection() {
	testData := []models.Collection{
		{
			Title:       "To Be Deleted",
			Description: "This collection will be deleted",
		},
	}

	repo := database.NewCollectionRepository(s.DB)
	persistedCollections, err := repo.Create(s.T().Context(), testData)
	if err != nil {
		s.T().Fatalf("failed to create collections: %s", err)
	}

	router := mux.NewRouter()
	router.HandleFunc(
		"/admin/collections/{collectionID}",
		BuildFormHandler(s.DB, templating.BuildPageRenderFunc(true, "")),
	).Methods(http.MethodPost)

	form := url.Values{}
	form.Add("_method", "DELETE")

	req, err := http.NewRequestWithContext(
		s.T().Context(),
		http.MethodPost,
		fmt.Sprintf("/admin/collections/%d", persistedCollections[0].ID),
		strings.NewReader(form.Encode()),
	)
	s.Require().NoError(err)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	s.Require().Equal(http.StatusSeeOther, rr.Code)
	s.Equal("/admin/collections", rr.Result().Header["Location"][0])

	// check that the collection is deleted from database
	returnedCollections, err := repo.All(s.T().Context())
	s.Require().NoError(err)
	s.Empty(returnedCollections)
}
