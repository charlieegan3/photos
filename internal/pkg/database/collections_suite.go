package database

import (
	"database/sql"

	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/suite"

	"github.com/charlieegan3/photos/internal/pkg/models"
)

// CollectionsSuite is a number of tests to define the database integration for
// storing collections.
type CollectionsSuite struct {
	suite.Suite

	DB *sql.DB
}

func (s *CollectionsSuite) SetupTest() {
	err := Truncate(s.T().Context(), s.DB, "photos.collections")
	if err != nil {
		s.T().Fatalf("failed to truncate collections table: %s", err)
	}
	err = Truncate(s.T().Context(), s.DB, "photos.post_collections")
	if err != nil {
		s.T().Fatalf("failed to truncate post_collections table: %s", err)
	}
}

func (s *CollectionsSuite) TestCreateCollections() {
	collections := []models.Collection{
		{
			Title:       "Nature Photography",
			Description: "Beautiful nature shots",
		},
		{
			Title:       "Street Photography",
			Description: "Urban street photography",
		},
	}

	repo := NewCollectionRepository(s.DB)
	returnedCollections, err := repo.Create(s.T().Context(), collections)
	if err != nil {
		s.T().Fatalf("failed to create collections: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Collection{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Collection{
					Title:       "Nature Photography",
					Description: "Beautiful nature shots",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
			1: td.SStruct(
				models.Collection{
					Title:       "Street Photography",
					Description: "Urban street photography",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedCollections, expectedResult)
}

func (s *CollectionsSuite) TestFindCollectionsByID() {
	collections := []models.Collection{
		{
			Title:       "Wildlife",
			Description: "Wild animals",
		},
	}

	repo := NewCollectionRepository(s.DB)
	createdCollections, err := repo.Create(s.T().Context(), collections)
	if err != nil {
		s.T().Fatalf("failed to create collections: %s", err)
	}

	foundCollections, err := repo.FindByIDs(s.T().Context(), []int64{int64(createdCollections[0].ID)})
	if err != nil {
		s.T().Fatalf("failed to find collections: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Collection{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Collection{
					Title:       "Wildlife",
					Description: "Wild animals",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), foundCollections, expectedResult)
}

func (s *CollectionsSuite) TestUpdateCollections() {
	collections := []models.Collection{
		{
			Title:       "Original Title",
			Description: "Original description",
		},
	}

	repo := NewCollectionRepository(s.DB)
	createdCollections, err := repo.Create(s.T().Context(), collections)
	if err != nil {
		s.T().Fatalf("failed to create collections: %s", err)
	}

	// Update the collection
	createdCollections[0].Title = "Updated Title"
	createdCollections[0].Description = "Updated description"

	updatedCollections, err := repo.Update(s.T().Context(), createdCollections)
	if err != nil {
		s.T().Fatalf("failed to update collections: %s", err)
	}

	expectedResult := td.Slice(
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

	td.Cmp(s.T(), updatedCollections, expectedResult)
}

func (s *CollectionsSuite) TestDeleteCollections() {
	collections := []models.Collection{
		{
			Title:       "To Be Deleted",
			Description: "This will be deleted",
		},
	}

	repo := NewCollectionRepository(s.DB)
	createdCollections, err := repo.Create(s.T().Context(), collections)
	if err != nil {
		s.T().Fatalf("failed to create collections: %s", err)
	}

	err = repo.Delete(s.T().Context(), createdCollections)
	if err != nil {
		s.T().Fatalf("failed to delete collections: %s", err)
	}

	// Try to find the deleted collection
	foundCollections, err := repo.FindByIDs(s.T().Context(), []int64{int64(createdCollections[0].ID)})
	if err != nil {
		s.T().Fatalf("failed to search for collections: %s", err)
	}

	s.Empty(foundCollections)
}

func (s *CollectionsSuite) TestAllOrderedByPostCount() {
	collections := []models.Collection{
		{
			Title:       "Empty Collection",
			Description: "Has no posts",
		},
		{
			Title:       "Popular Collection",
			Description: "Has posts",
		},
	}

	repo := NewCollectionRepository(s.DB)
	_, err := repo.Create(s.T().Context(), collections)
	if err != nil {
		s.T().Fatalf("failed to create collections: %s", err)
	}

	allCollections, err := repo.AllOrderedByPostCount(s.T().Context())
	if err != nil {
		s.T().Fatalf("failed to get all collections: %s", err)
	}

	// Should return both collections, ordered by post count (both have 0 posts, so order by created_at desc)
	s.Len(allCollections, 2)
	s.Equal("Popular Collection", allCollections[0].Title) // Most recent first when post count is equal
}
