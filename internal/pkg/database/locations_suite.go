package database

import (
	"database/sql"

	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/charlieegan3/photos/cms/internal/pkg/models"
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

	returnedLocations, err := FindLocationsByID(s.DB, persistedLocations[0].ID)
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

	returnedLocations, err = FindLocationsByID(s.DB, returnedLocations[0].ID)
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
