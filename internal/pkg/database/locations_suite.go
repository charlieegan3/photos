package database

import (
	"database/sql"
	"github.com/charlieegan3/photos/cms/internal/pkg/models"
	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

// LocationsSuite is a number of tests to define the database integration for
// storing locations
type LocationsSuite struct {
	suite.Suite
	DB *sql.DB
}

func (s *LocationsSuite) SetupTest() {
	err := Truncate(s.DB, "locations")
	if err != nil {
		s.T().Fatalf("failed to truncate table: %s", err)
	}
}

func (s *LocationsSuite) TestCreateLocations() {
	locations := []models.Location{
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

	returnedLocations, err := CreateLocations(s.DB, locations)
	if err != nil {
		s.T().Fatalf("failed to create locations: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Location{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Location{
					Name:      "London",
					Slug:      "london",
					Latitude:  1.1,
					Longitude: 1.2,
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
			1: td.SStruct(
				models.Location{
					Name:      "New York",
					Slug:      "new-york",
					Latitude:  1.3,
					Longitude: 1.4,
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedLocations, expectedResult)
}

func (s *LocationsSuite) TestFindLocationsByID() {
	locations := []models.Location{
		{
			Name: "Edinburgh",
		},
		{
			Name: "Norwich",
		},
	}

	persistedLocations, err := CreateLocations(s.DB, locations)
	if err != nil {
		s.T().Fatalf("failed to create locations needed for test: %s", err)
	}

	returnedLocations, err := FindLocationsByID(s.DB, []int{persistedLocations[0].ID})
	if err != nil {
		s.T().Fatalf("failed get locations: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Location{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Location{
					Name: "Edinburgh",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedLocations, expectedResult)
}

func (s *LocationsSuite) TestFindLocationsByName() {
	locations := []models.Location{
		{
			Name: "Edinburgh",
		},
		{
			Name: "Norwich",
		},
	}

	_, err := CreateLocations(s.DB, locations)
	if err != nil {
		s.T().Fatalf("failed to create locations needed for test: %s", err)
	}

	returnedLocations, err := FindLocationsByName(s.DB, "Edinburgh")
	if err != nil {
		s.T().Fatalf("failed get locations: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Location{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Location{
					Name: "Edinburgh",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedLocations, expectedResult)
}

func (s *LocationsSuite) TestAllLocations() {
	locations := []models.Location{
		{
			Name: "London",
		},
		{
			Name: "New York",
		},
	}

	_, err := CreateLocations(s.DB, locations)
	if err != nil {
		s.T().Fatalf("failed to create locations needed for test: %s", err)
	}

	returnedLocations, err := AllLocations(s.DB)
	if err != nil {
		s.T().Fatalf("failed get locations: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Location{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Location{
					Name: "London",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
			1: td.SStruct(
				models.Location{
					Name: "New York",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedLocations, expectedResult)
}

func (s *LocationsSuite) TestUpdateLocations() {
	initialLocations := []models.Location{
		{
			Name: "Hereford",
		},
		{
			Name: "Reading",
		},
	}

	returnedLocations, err := CreateLocations(s.DB, initialLocations)
	if err != nil {
		s.T().Fatalf("failed to create locations needed for test: %s", err)
	}

	expectedLocations := td.Slice(
		[]models.Location{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Location{
					Name: "Hereford",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
			1: td.SStruct(
				models.Location{
					Name: "Reading",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedLocations, expectedLocations)

	updatedLocations := returnedLocations
	updatedLocations[0].Name = "'ereford"
	updatedLocations[0].Longitude = 1.1

	returnedLocations, err = UpdateLocations(s.DB, updatedLocations)
	if err != nil {
		s.T().Fatalf("failed to update locations: %s", err)
	}

	expectedLocations = td.Slice(
		[]models.Location{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Location{
					Name: "'ereford",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
			1: td.SStruct(
				models.Location{
					Name: "Reading",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedLocations, expectedLocations)

	returnedLocations, err = FindLocationsByID(s.DB, []int{returnedLocations[0].ID})
	if err != nil {
		s.T().Fatalf("failed get locations: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Location{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Location{
					Name: "'ereford",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedLocations, expectedResult)
}

func (s *LocationsSuite) TestDeleteLocations() {
	locations := []models.Location{
		{
			Name: "Inverness",
		},
		{
			Name: "Dingwall",
		},
	}

	returnedLocations, err := CreateLocations(s.DB, locations)
	if err != nil {
		s.T().Fatalf("failed to create locations: %s", err)
	}

	locationToDelete := returnedLocations[1]

	err = DeleteLocations(s.DB, []models.Location{locationToDelete})
	require.NoError(s.T(), err, "unexpected error deleting locations")

	allLocations, err := AllLocations(s.DB)
	if err != nil {
		s.T().Fatalf("failed get locations: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Location{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Location{
					Name: "Inverness",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), allLocations, expectedResult)
}

func (s *PostsSuite) TestMergeLocations() {
	devices := []models.Device{
		{
			Name: "Example Device",
		},
	}
	returnedDevices, err := CreateDevices(s.DB, devices)
	require.NoError(s.T(), err)

	medias := []models.Media{
		{DeviceID: returnedDevices[0].ID},
	}
	returnedMedias, err := CreateMedias(s.DB, medias)
	require.NoError(s.T(), err)

	locations := []models.Location{
		{
			Name:      "London 1",
			Latitude:  1.1,
			Longitude: 1.2,
		},
		{
			Name:      "London 2",
			Latitude:  1.1,
			Longitude: 1.2,
		},
		{
			Name:      "London 3",
			Latitude:  1.1,
			Longitude: 1.2,
		},
	}
	returnedLocations, err := CreateLocations(s.DB, locations)
	require.NoError(s.T(), err)

	posts := []models.Post{
		{
			MediaID:    returnedMedias[0].ID,
			LocationID: returnedLocations[0].ID,
		},
		{
			MediaID:    returnedMedias[0].ID,
			LocationID: returnedLocations[1].ID,
		},
	}
	_, err = CreatePosts(s.DB, posts)
	require.NoError(s.T(), err)

	s.T().Run("simple merge", func(t *testing.T) {
		remainingLocationID, err := MergeLocations(s.DB, "London 1", "London 2")
		require.NoError(s.T(), err)

		assert.Equal(t, returnedLocations[0].ID, remainingLocationID)
	})

	s.T().Run("merge when target name is missing from table", func(t *testing.T) {
		remainingLocationID, err := MergeLocations(s.DB, "London X", "London 3")
		require.NoError(s.T(), err)

		assert.Equal(t, 0, remainingLocationID)

		locations, err := FindLocationsByName(s.DB, "London 3")
		require.NoError(s.T(), err)

		assert.Equal(t, 1, len(locations))
	})

	s.T().Run("merge when old name is missing from table", func(t *testing.T) {
		remainingLocationID, err := MergeLocations(s.DB, "London 3", "London X")
		require.NoError(s.T(), err)

		assert.Equal(t, 0, remainingLocationID)

		locations, err := FindLocationsByName(s.DB, "London 3")
		require.NoError(s.T(), err)

		assert.Equal(t, 1, len(locations))
	})
}
