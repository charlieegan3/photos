package public

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gorilla/mux"
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
	err = database.Truncate(s.T().Context(), s.DB, "photos.post_collections")
	s.Require().NoError(err)
	err = database.Truncate(s.T().Context(), s.DB, "photos.posts")
	s.Require().NoError(err)
	err = database.Truncate(s.T().Context(), s.DB, "photos.medias")
	s.Require().NoError(err)
	err = database.Truncate(s.T().Context(), s.DB, "photos.devices")
	s.Require().NoError(err)
	err = database.Truncate(s.T().Context(), s.DB, "photos.locations")
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
	router.HandleFunc("/collections",
		BuildIndexHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodGet)

	req, err := http.NewRequestWithContext(s.T().Context(), http.MethodGet, "/collections", nil)
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

func (s *EndpointsCollectionsSuite) TestGetCollectionWithPosts() {
	// Create a location first
	locationRepo := database.NewLocationRepository(s.DB)
	locations, err := locationRepo.Create(s.T().Context(), []models.Location{
		{
			Name:      "Test Location",
			Latitude:  51.5074,
			Longitude: -0.1278,
		},
	})
	s.Require().NoError(err)

	// Create a device first
	deviceRepo := database.NewDeviceRepository(s.DB)
	devices, err := deviceRepo.Create(s.T().Context(), []models.Device{
		{Name: "Test Device"},
	})
	s.Require().NoError(err)

	// Create a media
	mediaRepo := database.NewMediaRepository(s.DB)
	medias, err := mediaRepo.Create(s.T().Context(), []models.Media{
		{
			DeviceID:    devices[0].ID,
			Kind:        "jpg",
			TakenAt:     time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			Orientation: 1,
		},
	})
	s.Require().NoError(err)

	// Create a post
	postRepo := database.NewPostRepository(s.DB)
	posts, err := postRepo.Create(s.T().Context(), []models.Post{
		{
			Description: "Test post for collection",
			PublishDate: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			MediaID:     medias[0].ID,
			LocationID:  locations[0].ID,
		},
	})
	s.Require().NoError(err)

	// Create a collection
	collectionRepo := database.NewCollectionRepository(s.DB)
	collections, err := collectionRepo.Create(s.T().Context(), []models.Collection{
		{
			Title:       "Wildlife",
			Description: "Wild animals in their natural habitat",
		},
	})
	s.Require().NoError(err)

	// Create post-collection relationship
	postCollectionRepo := database.NewPostCollectionRepository(s.DB)
	_, err = postCollectionRepo.Create(s.T().Context(), []models.PostCollection{
		{
			PostID:       posts[0].ID,
			CollectionID: collections[0].ID,
		},
	})
	s.Require().NoError(err)

	router := mux.NewRouter()
	router.HandleFunc("/collections/{collectionID}",
		BuildGetHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodGet)

	req, err := http.NewRequestWithContext(
		s.T().Context(), http.MethodGet, fmt.Sprintf("/collections/%d", collections[0].ID), nil,
	)
	s.Require().NoError(err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	s.Require().Equal(http.StatusOK, rr.Code)

	body, err := io.ReadAll(rr.Body)
	s.Require().NoError(err)

	s.Contains(string(body), "Wildlife")
	s.Contains(string(body), "Wild animals in their natural habitat")
	s.Contains(string(body), "Test post for collection")
}

func (s *EndpointsCollectionsSuite) TestGetCollectionNotFound() {
	router := mux.NewRouter()
	router.HandleFunc("/collections/{collectionID}",
		BuildGetHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodGet)

	req, err := http.NewRequestWithContext(
		s.T().Context(), http.MethodGet, "/collections/999", nil,
	)
	s.Require().NoError(err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	s.Require().Equal(http.StatusNotFound, rr.Code)
}

func (s *EndpointsCollectionsSuite) TestGetCollectionWithNoPosts() {
	// Create a collection but no posts
	collectionRepo := database.NewCollectionRepository(s.DB)
	collections, err := collectionRepo.Create(s.T().Context(), []models.Collection{
		{
			Title:       "Empty Collection",
			Description: "This collection has no posts",
		},
	})
	s.Require().NoError(err)

	router := mux.NewRouter()
	router.HandleFunc("/collections/{collectionID}",
		BuildGetHandler(s.DB, templating.BuildPageRenderFunc(true, ""))).
		Methods(http.MethodGet)

	req, err := http.NewRequestWithContext(
		s.T().Context(), http.MethodGet, fmt.Sprintf("/collections/%d", collections[0].ID), nil,
	)
	s.Require().NoError(err)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Should return 404 when collection has no posts
	s.Require().Equal(http.StatusNotFound, rr.Code)
}
