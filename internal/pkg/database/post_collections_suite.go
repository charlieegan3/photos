package database

import (
	"database/sql"
	"time"

	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/suite"

	"github.com/charlieegan3/photos/internal/pkg/models"
)

// PostCollectionsSuite is a number of tests to define the database integration for
// storing post-collection relationships.
type PostCollectionsSuite struct {
	suite.Suite

	DB *sql.DB
}

func (s *PostCollectionsSuite) SetupTest() {
	err := Truncate(s.T().Context(), s.DB, "photos.post_collections")
	if err != nil {
		s.T().Fatalf("failed to truncate post_collections table: %s", err)
	}
	err = Truncate(s.T().Context(), s.DB, "photos.collections")
	if err != nil {
		s.T().Fatalf("failed to truncate collections table: %s", err)
	}
	err = Truncate(s.T().Context(), s.DB, "photos.posts")
	if err != nil {
		s.T().Fatalf("failed to truncate posts table: %s", err)
	}
	err = Truncate(s.T().Context(), s.DB, "photos.medias")
	if err != nil {
		s.T().Fatalf("failed to truncate medias table: %s", err)
	}
	err = Truncate(s.T().Context(), s.DB, "photos.devices")
	if err != nil {
		s.T().Fatalf("failed to truncate devices table: %s", err)
	}
	err = Truncate(s.T().Context(), s.DB, "photos.locations")
	if err != nil {
		s.T().Fatalf("failed to truncate locations table: %s", err)
	}
}

func (s *PostCollectionsSuite) TestCreatePostCollections() {
	// Create test data first
	locationRepo := NewLocationRepository(s.DB)
	locations, err := locationRepo.Create(s.T().Context(), []models.Location{
		{Name: "Test Location", Latitude: 51.5074, Longitude: -0.1278},
	})
	s.Require().NoError(err)

	deviceRepo := NewDeviceRepository(s.DB)
	devices, err := deviceRepo.Create(s.T().Context(), []models.Device{
		{Name: "Test Device"},
	})
	s.Require().NoError(err)

	mediaRepo := NewMediaRepository(s.DB)
	medias, err := mediaRepo.Create(s.T().Context(), []models.Media{
		{Kind: "jpg", TakenAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC), Orientation: 1, DeviceID: devices[0].ID},
	})
	s.Require().NoError(err)

	postRepo := NewPostRepository(s.DB)
	posts, err := postRepo.Create(s.T().Context(), []models.Post{
		{
			Description: "Test post",
			PublishDate: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			MediaID:     medias[0].ID,
			LocationID:  locations[0].ID,
		},
	})
	s.Require().NoError(err)

	collectionRepo := NewCollectionRepository(s.DB)
	collections, err := collectionRepo.Create(s.T().Context(), []models.Collection{
		{Title: "Test Collection", Description: "Test description"},
	})
	s.Require().NoError(err)

	// Create post-collection relationships
	postCollections := []models.PostCollection{
		{
			PostID:       posts[0].ID,
			CollectionID: collections[0].ID,
		},
	}

	repo := NewPostCollectionRepository(s.DB)
	returnedPostCollections, err := repo.Create(s.T().Context(), postCollections)
	if err != nil {
		s.T().Fatalf("failed to create post collections: %s", err)
	}

	expectedResult := td.Slice(
		[]models.PostCollection{},
		td.ArrayEntries{
			0: td.SStruct(
				models.PostCollection{
					PostID:       posts[0].ID,
					CollectionID: collections[0].ID,
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedPostCollections, expectedResult)
}

func (s *PostCollectionsSuite) TestFindByPostID() {
	// Create test data
	locationRepo := NewLocationRepository(s.DB)
	locations, err := locationRepo.Create(s.T().Context(), []models.Location{
		{Name: "Test Location", Latitude: 51.5074, Longitude: -0.1278},
	})
	s.Require().NoError(err)

	deviceRepo := NewDeviceRepository(s.DB)
	devices, err := deviceRepo.Create(s.T().Context(), []models.Device{
		{Name: "Test Device"},
	})
	s.Require().NoError(err)

	mediaRepo := NewMediaRepository(s.DB)
	medias, err := mediaRepo.Create(s.T().Context(), []models.Media{
		{Kind: "jpg", TakenAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC), Orientation: 1, DeviceID: devices[0].ID},
	})
	s.Require().NoError(err)

	postRepo := NewPostRepository(s.DB)
	posts, err := postRepo.Create(s.T().Context(), []models.Post{
		{
			Description: "Test post",
			PublishDate: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			MediaID:     medias[0].ID,
			LocationID:  locations[0].ID,
		},
	})
	s.Require().NoError(err)

	collectionRepo := NewCollectionRepository(s.DB)
	collections, err := collectionRepo.Create(s.T().Context(), []models.Collection{
		{Title: "Test Collection 1", Description: "Test description 1"},
		{Title: "Test Collection 2", Description: "Test description 2"},
	})
	s.Require().NoError(err)

	// Create post-collection relationships
	repo := NewPostCollectionRepository(s.DB)
	_, err = repo.Create(s.T().Context(), []models.PostCollection{
		{PostID: posts[0].ID, CollectionID: collections[0].ID},
		{PostID: posts[0].ID, CollectionID: collections[1].ID},
	})
	s.Require().NoError(err)

	// Find by post ID
	foundPostCollections, err := repo.FindByPostID(s.T().Context(), posts[0].ID)
	if err != nil {
		s.T().Fatalf("failed to find post collections: %s", err)
	}

	s.Len(foundPostCollections, 2)
}

func (s *PostCollectionsSuite) TestFindByCollectionID() {
	// Create test data
	locationRepo := NewLocationRepository(s.DB)
	locations, err := locationRepo.Create(s.T().Context(), []models.Location{
		{Name: "Test Location", Latitude: 51.5074, Longitude: -0.1278},
	})
	s.Require().NoError(err)

	deviceRepo := NewDeviceRepository(s.DB)
	devices, err := deviceRepo.Create(s.T().Context(), []models.Device{
		{Name: "Test Device"},
	})
	s.Require().NoError(err)

	mediaRepo := NewMediaRepository(s.DB)
	medias, err := mediaRepo.Create(s.T().Context(), []models.Media{
		{Kind: "jpg", TakenAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC), Orientation: 1, DeviceID: devices[0].ID},
	})
	s.Require().NoError(err)

	postRepo := NewPostRepository(s.DB)
	posts, err := postRepo.Create(s.T().Context(), []models.Post{
		{
			Description: "Test post 1",
			PublishDate: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			MediaID:     medias[0].ID,
			LocationID:  locations[0].ID,
		},
		{
			Description: "Test post 2",
			PublishDate: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
			MediaID:     medias[0].ID,
			LocationID:  locations[0].ID,
		},
	})
	s.Require().NoError(err)

	collectionRepo := NewCollectionRepository(s.DB)
	collections, err := collectionRepo.Create(s.T().Context(), []models.Collection{
		{Title: "Test Collection", Description: "Test description"},
	})
	s.Require().NoError(err)

	// Create post-collection relationships
	repo := NewPostCollectionRepository(s.DB)
	_, err = repo.Create(s.T().Context(), []models.PostCollection{
		{PostID: posts[0].ID, CollectionID: collections[0].ID},
		{PostID: posts[1].ID, CollectionID: collections[0].ID},
	})
	s.Require().NoError(err)

	// Find by collection ID
	foundPostCollections, err := repo.FindByCollectionID(s.T().Context(), collections[0].ID)
	if err != nil {
		s.T().Fatalf("failed to find post collections: %s", err)
	}

	s.Len(foundPostCollections, 2)
}

func (s *PostCollectionsSuite) TestDeleteByPostID() {
	// Create test data
	locationRepo := NewLocationRepository(s.DB)
	locations, err := locationRepo.Create(s.T().Context(), []models.Location{
		{Name: "Test Location", Latitude: 51.5074, Longitude: -0.1278},
	})
	s.Require().NoError(err)

	deviceRepo := NewDeviceRepository(s.DB)
	devices, err := deviceRepo.Create(s.T().Context(), []models.Device{
		{Name: "Test Device"},
	})
	s.Require().NoError(err)

	mediaRepo := NewMediaRepository(s.DB)
	medias, err := mediaRepo.Create(s.T().Context(), []models.Media{
		{Kind: "jpg", TakenAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC), Orientation: 1, DeviceID: devices[0].ID},
	})
	s.Require().NoError(err)

	postRepo := NewPostRepository(s.DB)
	posts, err := postRepo.Create(s.T().Context(), []models.Post{
		{
			Description: "Test post",
			PublishDate: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			MediaID:     medias[0].ID,
			LocationID:  locations[0].ID,
		},
	})
	s.Require().NoError(err)

	collectionRepo := NewCollectionRepository(s.DB)
	collections, err := collectionRepo.Create(s.T().Context(), []models.Collection{
		{Title: "Test Collection 1", Description: "Test description 1"},
		{Title: "Test Collection 2", Description: "Test description 2"},
	})
	s.Require().NoError(err)

	// Create post-collection relationships
	repo := NewPostCollectionRepository(s.DB)
	_, err = repo.Create(s.T().Context(), []models.PostCollection{
		{PostID: posts[0].ID, CollectionID: collections[0].ID},
		{PostID: posts[0].ID, CollectionID: collections[1].ID},
	})
	s.Require().NoError(err)

	// Delete all relationships for the post
	err = repo.DeleteByPostID(s.T().Context(), posts[0].ID)
	if err != nil {
		s.T().Fatalf("failed to delete post collections: %s", err)
	}

	// Verify relationships are deleted
	foundPostCollections, err := repo.FindByPostID(s.T().Context(), posts[0].ID)
	if err != nil {
		s.T().Fatalf("failed to find post collections: %s", err)
	}

	s.Empty(foundPostCollections)
}

func (s *PostCollectionsSuite) TestCreateWithConflictHandling() {
	// Create test data
	locationRepo := NewLocationRepository(s.DB)
	locations, err := locationRepo.Create(s.T().Context(), []models.Location{
		{Name: "Test Location", Latitude: 51.5074, Longitude: -0.1278},
	})
	s.Require().NoError(err)

	deviceRepo := NewDeviceRepository(s.DB)
	devices, err := deviceRepo.Create(s.T().Context(), []models.Device{
		{Name: "Test Device"},
	})
	s.Require().NoError(err)

	mediaRepo := NewMediaRepository(s.DB)
	medias, err := mediaRepo.Create(s.T().Context(), []models.Media{
		{Kind: "jpg", TakenAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC), Orientation: 1, DeviceID: devices[0].ID},
	})
	s.Require().NoError(err)

	postRepo := NewPostRepository(s.DB)
	posts, err := postRepo.Create(s.T().Context(), []models.Post{
		{
			Description: "Test post",
			PublishDate: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			MediaID:     medias[0].ID,
			LocationID:  locations[0].ID,
		},
	})
	s.Require().NoError(err)

	collectionRepo := NewCollectionRepository(s.DB)
	collections, err := collectionRepo.Create(s.T().Context(), []models.Collection{
		{Title: "Test Collection", Description: "Test description"},
	})
	s.Require().NoError(err)

	repo := NewPostCollectionRepository(s.DB)

	// Create initial relationship
	postCollections := []models.PostCollection{
		{PostID: posts[0].ID, CollectionID: collections[0].ID},
	}

	_, err = repo.CreateWithConflictHandling(s.T().Context(), postCollections)
	s.Require().NoError(err)

	// Try to create duplicate - should not error
	_, err = repo.CreateWithConflictHandling(s.T().Context(), postCollections)
	s.Require().NoError(err)

	// Verify only one relationship exists
	foundPostCollections, err := repo.FindByPostID(s.T().Context(), posts[0].ID)
	if err != nil {
		s.T().Fatalf("failed to find post collections: %s", err)
	}

	s.Len(foundPostCollections, 1)
}
