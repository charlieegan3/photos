package database

import (
	"database/sql"
	"github.com/charlieegan3/photos/internal/pkg/models"
	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"time"
)

// PointsSuite is a number of tests to define the database integration for
// storing points
type PointsSuite struct {
	suite.Suite
	DB *sql.DB
}

func (s *PointsSuite) SetupTest() {
	var err error
	err = Truncate(s.DB, "locations.points")
	require.NoError(s.T(), err)
	err = Truncate(s.DB, "locations.reasons")
	require.NoError(s.T(), err)
	err = Truncate(s.DB, "locations.activities")
	require.NoError(s.T(), err)
	err = Truncate(s.DB, "locations.callers")
	require.NoError(s.T(), err)
	err = Truncate(s.DB, "locations.importers")
	require.NoError(s.T(), err)
}

func (s *PointsSuite) TestCreatePoints() {
	points := []models.Point{
		{
			Latitude:  1.0,
			Longitude: 2.0,
			Altitude:  3.0,

			Velocity: 3.0,

			Accuracy:         1.0,
			VerticalAccuracy: 2.0,

			WasOffline: false,
		},
	}

	returnedPoints, err := CreatePoints(
		s.DB,
		"example_importer",
		"example_caller",
		"example_reason",
		nil, // no activity set
		points,
	)
	require.NoError(s.T(), err)

	expectedResult := td.Slice(
		[]models.Point{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Point{
					Latitude:  1.0,
					Longitude: 2.0,
					Altitude:  3.0,

					Velocity: 3.0,

					Accuracy:         1.0,
					VerticalAccuracy: 2.0,

					WasOffline: false,

					ActivityID: nil,
				},
				td.StructFields{
					"ID":         td.Not(0),
					"CallerID":   td.Not(0),
					"ImporterID": td.Not(0),
					"ReasonID":   td.Not(0),
					"=*":         td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedPoints, expectedResult)
}

func (s *PointsSuite) TestPointsNearTime() {
	points := []models.Point{
		{
			Latitude:  3.0,
			Longitude: 4.0,
			CreatedAt: time.Date(1994, 4, 23, 13, 22, 0, 0, time.UTC),
		},
		{
			Latitude:  3.0,
			Longitude: 4.0,
			CreatedAt: time.Date(2022, 4, 23, 13, 22, 0, 0, time.UTC),
		},
		{
			Latitude:  1.0,
			Longitude: 2.0,
			CreatedAt: time.Date(2022, 5, 23, 13, 19, 0, 0, time.UTC),
		},
		{
			Latitude:  3.0,
			Longitude: 4.0,
			CreatedAt: time.Date(2022, 5, 23, 13, 22, 0, 0, time.UTC),
		},
		{
			Latitude:  3.0,
			Longitude: 4.0,
			CreatedAt: time.Date(2022, 4, 23, 13, 22, 0, 0, time.UTC),
		},
		{
			Latitude:  3.0,
			Longitude: 4.0,
			CreatedAt: time.Date(2022, 4, 23, 13, 22, 0, 0, time.UTC),
		},
	}

	_, err := CreatePoints(
		s.DB,
		"example_importer",
		"example_caller",
		"example_reason",
		nil, // no activity set
		points,
	)
	require.NoError(s.T(), err)

	returnedPoints, err := PointsNearTime(s.DB, time.Date(2022, 5, 23, 13, 20, 0, 0, time.UTC))
	require.NoError(s.T(), err)

	expectedResult := td.Slice(
		[]models.Point{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Point{
					Latitude:  1.0,
					Longitude: 2.0,
					CreatedAt: time.Date(2022, 5, 23, 13, 19, 0, 0, time.UTC),
				},
				td.StructFields{"=*": td.Ignore()}),
			1: td.SStruct(
				models.Point{
					Latitude:  3.0,
					Longitude: 4.0,
					CreatedAt: time.Date(2022, 5, 23, 13, 22, 0, 0, time.UTC),
				},
				td.StructFields{"=*": td.Ignore()}),
			2: td.SStruct(
				models.Point{
					Latitude:  3.0,
					Longitude: 4.0,
					CreatedAt: time.Date(2022, 4, 23, 13, 22, 0, 0, time.UTC),
				},
				td.StructFields{"=*": td.Ignore()}),
			3: td.SStruct(
				models.Point{
					Latitude:  3.0,
					Longitude: 4.0,
					CreatedAt: time.Date(2022, 4, 23, 13, 22, 0, 0, time.UTC),
				},
				td.StructFields{"=*": td.Ignore()}),
			4: td.SStruct(
				models.Point{
					Latitude:  3.0,
					Longitude: 4.0,
					CreatedAt: time.Date(2022, 4, 23, 13, 22, 0, 0, time.UTC),
				},
				td.StructFields{"=*": td.Ignore()}),
		},
	)

	td.Cmp(s.T(), returnedPoints, expectedResult)
}

func (s *PointsSuite) TestPointsInRange() {
	points := []models.Point{
		{
			Latitude:  3.0,
			Longitude: 4.0,
			CreatedAt: time.Date(1994, 4, 23, 13, 22, 0, 0, time.UTC),
		},
		{
			Latitude:  3.0,
			Longitude: 4.0,
			CreatedAt: time.Date(2021, 4, 23, 13, 22, 0, 0, time.UTC),
		},
		{
			Latitude:  1.0,
			Longitude: 2.0,
			CreatedAt: time.Date(2021, 5, 23, 13, 19, 0, 0, time.UTC),
		},
		{
			Latitude:  3.0,
			Longitude: 4.0,
			CreatedAt: time.Date(2022, 4, 23, 13, 22, 0, 0, time.UTC),
		},
	}

	_, err := CreatePoints(
		s.DB,
		"example_importer",
		"example_caller",
		"example_reason",
		nil, // no activity set
		points,
	)
	require.NoError(s.T(), err)

	returnedPoints, err := PointsInRange(
		s.DB,
		time.Date(2021, 0, 0, 0, 0, 0, 0, time.UTC),
		time.Date(2022, 0, 0, 0, 0, 0, 0, time.UTC),
	)
	require.NoError(s.T(), err)

	expectedResult := td.Slice(
		[]models.Point{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Point{
					Latitude:  3.0,
					Longitude: 4.0,
					CreatedAt: time.Date(2021, 4, 23, 13, 22, 0, 0, time.UTC),
				},
				td.StructFields{"=*": td.Ignore()}),
			1: td.SStruct(
				models.Point{
					Latitude:  1.0,
					Longitude: 2.0,
					CreatedAt: time.Date(2021, 5, 23, 13, 19, 0, 0, time.UTC),
				},
				td.StructFields{"=*": td.Ignore()}),
		},
	)

	td.Cmp(s.T(), returnedPoints, expectedResult)
}
