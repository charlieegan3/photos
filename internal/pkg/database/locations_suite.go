package database

import (
	"database/sql"

	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/suite"

	"github.com/charlieegan3/photos/internal/pkg/models"
)

// LocationsSuite is a number of tests to define the database integration for
// storing locations.
type LocationsSuite struct {
	suite.Suite

	DB *sql.DB
}

func (s *LocationsSuite) SetupTest() {
	err := Truncate(s.T().Context(), s.DB, "photos.locations")
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

	returnedLocations, err := CreateLocations(s.T().Context(), s.DB, locations)
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

	persistedLocations, err := CreateLocations(s.T().Context(), s.DB, locations)
	if err != nil {
		s.T().Fatalf("failed to create locations needed for test: %s", err)
	}

	returnedLocations, err := FindLocationsByID(s.T().Context(), s.DB, []int{persistedLocations[0].ID})
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

	_, err := CreateLocations(s.T().Context(), s.DB, locations)
	if err != nil {
		s.T().Fatalf("failed to create locations needed for test: %s", err)
	}

	returnedLocations, err := FindLocationsByName(s.T().Context(), s.DB, "Edinburgh")
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

	_, err := CreateLocations(s.T().Context(), s.DB, locations)
	if err != nil {
		s.T().Fatalf("failed to create locations needed for test: %s", err)
	}

	returnedLocations, err := AllLocations(s.T().Context(), s.DB)
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

	returnedLocations, err := CreateLocations(s.T().Context(), s.DB, initialLocations)
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

	returnedLocations, err = UpdateLocations(s.T().Context(), s.DB, updatedLocations)
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

	returnedLocations, err = FindLocationsByID(s.T().Context(), s.DB, []int{returnedLocations[0].ID})
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

	returnedLocations, err := CreateLocations(s.T().Context(), s.DB, locations)
	if err != nil {
		s.T().Fatalf("failed to create locations: %s", err)
	}

	locationToDelete := returnedLocations[1]

	err = DeleteLocations(s.T().Context(), s.DB, []models.Location{locationToDelete})
	s.Require().NoError(err, "unexpected error deleting locations")

	allLocations, err := AllLocations(s.T().Context(), s.DB)
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
	returnedDevices, err := CreateDevices(s.T().Context(), s.DB, devices)
	s.Require().NoError(err)

	medias := []models.Media{
		{DeviceID: returnedDevices[0].ID},
	}
	returnedMedias, err := CreateMedias(s.T().Context(), s.DB, medias)
	s.Require().NoError(err)

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
	returnedLocations, err := CreateLocations(s.T().Context(), s.DB, locations)
	s.Require().NoError(err)

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
	_, err = CreatePosts(s.T().Context(), s.DB, posts)
	s.Require().NoError(err)

	s.Run("simple merge", func() {
		remainingLocationID, err := MergeLocations(s.T().Context(), s.DB, "London 1", "London 2")
		s.Require().NoError(err)

		s.Equal(returnedLocations[0].ID, remainingLocationID)
	})

	s.Run("merge when target name is missing from table", func() {
		remainingLocationID, err := MergeLocations(s.T().Context(), s.DB, "London X", "London 3")
		s.Require().NoError(err)

		s.Equal(0, remainingLocationID)

		locations, err := FindLocationsByName(s.T().Context(), s.DB, "London 3")
		s.Require().NoError(err)

		s.Len(locations, 1)
	})

	s.Run("merge when old name is missing from table", func() {
		remainingLocationID, err := MergeLocations(s.T().Context(), s.DB, "London 3", "London X")
		s.Require().NoError(err)

		s.Equal(0, remainingLocationID)

		locations, err := FindLocationsByName(s.T().Context(), s.DB, "London 3")
		s.Require().NoError(err)

		s.Len(locations, 1)
	})
}

func (s *LocationsSuite) TestNearbyLocations() {
	locations := []models.Location{
		{
			Name:      "Whittington Hospital",
			Latitude:  51.5657752,
			Longitude: -0.1388468,
		},
		{
			Name:      "Archway Station",
			Latitude:  51.565462952567,
			Longitude: -0.13486676038084,
		},
		{
			Name:      "Tokyo National Museum",
			Latitude:  35.718889,
			Longitude: 139.775833,
		},
		{
			Name:      "Highgate Hill",
			Latitude:  51.567761,
			Longitude: -0.138742,
		},
	}

	_, err := CreateLocations(s.T().Context(), s.DB, locations)
	s.Require().NoError(err)

	nearbyLocations, err := NearbyLocations(s.DB, 51.56748, -0.138666)
	s.Require().NoError(err)

	expectedResult := td.Slice(
		[]models.Location{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Location{
					Name:      "Highgate Hill",
					Latitude:  51.567761,
					Longitude: -0.138742,
				},
				td.StructFields{"=*": td.Ignore()}),
			1: td.SStruct(
				models.Location{
					Name:      "Whittington Hospital",
					Latitude:  51.5657752,
					Longitude: -0.1388468,
				},
				td.StructFields{"=*": td.Ignore()}),
			2: td.SStruct(
				models.Location{
					Name:      "Archway Station",
					Latitude:  51.565462952567,
					Longitude: -0.13486676038084,
				},
				td.StructFields{"=*": td.Ignore()}),
		},
	)

	td.Cmp(s.T(), nearbyLocations, expectedResult)
}
