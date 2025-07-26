package database

import (
	"database/sql"
	"time"

	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/suite"

	"github.com/charlieegan3/photos/internal/pkg/models"
)

// TripsSuite is a number of tests to define the database integration for
// storing trips
type TripsSuite struct {
	suite.Suite
	DB *sql.DB
}

func (s *TripsSuite) SetupTest() {
	err := Truncate(s.T().Context(), s.DB, "photos.trips")
	if err != nil {
		s.T().Fatalf("failed to truncate table: %s", err)
	}
}

func (s *TripsSuite) TestCreateTrips() {
	trips := []models.Trip{
		{
			Title:       "London",
			Description: "A trip to London",
			StartDate:   time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
		},
		{
			Title:       "New York",
			Description: "A trip to New York",
			StartDate:   time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
			EndDate:     time.Date(2020, 2, 2, 0, 0, 0, 0, time.UTC),
		},
	}

	returnedTrips, err := CreateTrips(s.T().Context(), s.DB, trips)
	if err != nil {
		s.T().Fatalf("failed to create trips: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Trip{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Trip{
					Title:       "London",
					Description: "A trip to London",
					StartDate:   time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
					EndDate:     time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC),
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
			1: td.SStruct(
				models.Trip{
					Title:       "New York",
					Description: "A trip to New York",
					StartDate:   time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
					EndDate:     time.Date(2020, 2, 2, 0, 0, 0, 0, time.UTC),
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedTrips, expectedResult)
}

func (s *TripsSuite) TestFindTripsByID() {
	trips := []models.Trip{
		{
			Title: "Edinburgh",
		},
		{
			Title: "Norwich",
		},
	}

	persistedTrips, err := CreateTrips(s.T().Context(), s.DB, trips)
	if err != nil {
		s.T().Fatalf("failed to create trips needed for test: %s", err)
	}

	returnedTrips, err := FindTripsByID(s.T().Context(), s.DB, []int{persistedTrips[0].ID})
	if err != nil {
		s.T().Fatalf("failed get trips: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Trip{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Trip{
					Title: "Edinburgh",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedTrips, expectedResult)
}

func (s *TripsSuite) TestAllTrips() {
	trips := []models.Trip{
		{
			Title: "London",
		},
		{
			Title: "New York",
		},
	}

	_, err := CreateTrips(s.T().Context(), s.DB, trips)
	if err != nil {
		s.T().Fatalf("failed to create trips needed for test: %s", err)
	}

	returnedTrips, err := AllTrips(s.T().Context(), s.DB)
	if err != nil {
		s.T().Fatalf("failed get trips: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Trip{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Trip{
					Title: "London",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
			1: td.SStruct(
				models.Trip{
					Title: "New York",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedTrips, expectedResult)
}

func (s *TripsSuite) TestUpdateTrips() {
	initialTrips := []models.Trip{
		{
			Title: "Hereford",
		},
		{
			Title: "Reading",
		},
	}

	returnedTrips, err := CreateTrips(s.T().Context(), s.DB, initialTrips)
	if err != nil {
		s.T().Fatalf("failed to create trips needed for test: %s", err)
	}

	expectedTrips := td.Slice(
		[]models.Trip{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Trip{
					Title: "Hereford",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
			1: td.SStruct(
				models.Trip{
					Title: "Reading",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedTrips, expectedTrips)

	updatedTrips := returnedTrips
	updatedTrips[0].Title = "'ereford"
	updatedTrips[1].StartDate = time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	returnedTrips, err = UpdateTrips(s.T().Context(), s.DB, updatedTrips)
	if err != nil {
		s.T().Fatalf("failed to update trips: %s", err)
	}

	expectedTrips = td.Slice(
		[]models.Trip{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Trip{
					Title: "'ereford",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
			1: td.SStruct(
				models.Trip{
					Title:     "Reading",
					StartDate: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedTrips, expectedTrips)

	returnedTrips, err = FindTripsByID(s.T().Context(), s.DB, []int{returnedTrips[0].ID})
	if err != nil {
		s.T().Fatalf("failed get trips: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Trip{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Trip{
					Title: "'ereford",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedTrips, expectedResult)
}

func (s *TripsSuite) TestDeleteTrips() {
	trips := []models.Trip{
		{
			Title: "Inverness",
		},
		{
			Title: "Dingwall",
		},
	}

	returnedTrips, err := CreateTrips(s.T().Context(), s.DB, trips)
	if err != nil {
		s.T().Fatalf("failed to create trips: %s", err)
	}

	tripToDelete := returnedTrips[1]

	err = DeleteTrips(s.T().Context(), s.DB, []models.Trip{tripToDelete})
	s.Require().NoError(err, "unexpected error deleting trips")

	allTrips, err := AllTrips(s.T().Context(), s.DB)
	if err != nil {
		s.T().Fatalf("failed get trips: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Trip{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Trip{
					Title: "Inverness",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), allTrips, expectedResult)
}
