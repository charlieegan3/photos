package database

import (
	"database/sql"
	"time"

	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/suite"

	"github.com/charlieegan3/photos/internal/pkg/models"
)

// LensesSuite is a number of tests to define the database integration for
// storing lenses
type LensesSuite struct {
	suite.Suite
	DB *sql.DB
}

func (s *LensesSuite) SetupTest() {
	var err error

	err = Truncate(s.DB, "photos.devices")
	if err != nil {
		s.T().Fatalf("failed to truncate table: %s", err)
	}

	err = Truncate(s.DB, "photos.medias")
	if err != nil {
		s.T().Fatalf("failed to truncate table: %s", err)
	}

	err = Truncate(s.DB, "photos.lenses")
	if err != nil {
		s.T().Fatalf("failed to truncate table: %s", err)
	}
}

func (s *LensesSuite) TestMostRecentlyUsedLens() {
	devices := []models.Device{
		{
			Name: "iPhone 11 Pro Max",
		},
		{
			Name: "X100F",
		},
	}
	returnedDevices, err := CreateDevices(s.DB, devices)
	if err != nil {
		s.T().Fatalf("failed to create devices: %s", err)
	}

	lenses := []models.Lens{
		{
			Name: "iPhone",
		},
		{
			Name: "X100F",
		},
	}

	returnedLenses, err := CreateLenses(s.DB, lenses)
	if err != nil {
		s.T().Fatalf("failed to create lenses: %s", err)
	}

	medias := []models.Media{
		{
			LensID:   returnedLenses[0].ID,
			DeviceID: returnedDevices[0].ID,

			Make:  "FujiFilm",
			Model: "X100F",

			TakenAt: time.Date(2021, time.November, 23, 19, 56, 0, 0, time.UTC),

			FNumber:                 2.0,
			ExposureTimeNumerator:   1,
			ExposureTimeDenominator: 100,
			ISOSpeed:                100,

			Latitude:  51.1,
			Longitude: 52.2,
			Altitude:  100.0,
		},
		{
			LensID:   returnedLenses[1].ID,
			DeviceID: returnedDevices[1].ID,

			Make:  "Apple",
			Model: "iPhone",

			TakenAt: time.Date(2020, time.June, 22, 18, 56, 0, 0, time.UTC),

			FNumber:  4.0,
			ISOSpeed: 400,

			Latitude:  53.1,
			Longitude: 54.2,
			Altitude:  200.0,
		},
	}
	_, err = CreateMedias(s.DB, medias)
	if err != nil {
		s.T().Fatalf("failed to create medias: %s", err)
	}

	lens, err := MostRecentlyUsedLens(s.DB)
	if err != nil {
		s.T().Fatalf("failed to create medias: %s", err)
	}

	td.Cmp(s.T(), lens, returnedLenses[0])
}

func (s *LensesSuite) TestCreateLenses() {
	lenses := []models.Lens{
		{
			Name: "iPhone",
		},
		{
			Name: "X100F",
		},
	}

	returnedLenses, err := CreateLenses(s.DB, lenses)
	if err != nil {
		s.T().Fatalf("failed to create lenses: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Lens{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Lens{
					Name: "iPhone",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
			1: td.SStruct(
				models.Lens{
					Name: "X100F",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedLenses, expectedResult)
}

func (s *LensesSuite) TestFindLensesByID() {
	lenses := []models.Lens{
		{
			Name: "iPhone",
		},
		{
			Name: "X100F",
		},
	}

	returnedLenses, err := CreateLenses(s.DB, lenses)
	if err != nil {
		s.T().Fatalf("failed to create lenses needed for test: %s", err)
	}

	returnedLenses, err = FindLensesByID(s.DB, []int64{returnedLenses[0].ID})
	if err != nil {
		s.T().Fatalf("failed get lenses: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Lens{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Lens{
					Name: "iPhone",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedLenses, expectedResult)
}

func (s *LensesSuite) TestFindLensesByName() {
	lenses := []models.Lens{
		{
			Name: "iPhone",
		},
		{
			Name: "X100F",
		},
	}

	_, err := CreateLenses(s.DB, lenses)
	if err != nil {
		s.T().Fatalf("failed to create lenses needed for test: %s", err)
	}

	returnedLenses, err := FindLensesByName(s.DB, "iPhone")
	if err != nil {
		s.T().Fatalf("failed get lenses: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Lens{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Lens{
					Name: "iPhone",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedLenses, expectedResult)
}

func (s *LensesSuite) TestAllLenses() {
	lenses := []models.Lens{
		{
			Name: "IPhone",
		},
		{
			Name: "X100F",
		},
	}

	_, err := CreateLenses(s.DB, lenses)
	if err != nil {
		s.T().Fatalf("failed to create lenses needed for test: %s", err)
	}

	returnedLenses, err := AllLenses(s.DB)
	if err != nil {
		s.T().Fatalf("failed get lenses: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Lens{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Lens{
					Name: "IPhone",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
			1: td.SStruct(
				models.Lens{
					Name: "X100F",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedLenses, expectedResult)
}

func (s *LensesSuite) TestUpdateLenses() {
	initialLenses := []models.Lens{
		{
			Name: "iPhone",
		},
		{
			Name: "X100F",
		},
	}

	returnedLenses, err := CreateLenses(s.DB, initialLenses)
	if err != nil {
		s.T().Fatalf("failed to create lenses needed for test: %s", err)
	}

	expectedLenses := td.Slice(
		[]models.Lens{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Lens{
					Name: "iPhone",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
			1: td.SStruct(
				models.Lens{
					Name: "X100F",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedLenses, expectedLenses)

	updatedLenses := returnedLenses
	updatedLenses[0].Name = "iPod"

	returnedLenses, err = UpdateLenses(s.DB, updatedLenses)
	if err != nil {
		s.T().Fatalf("failed to update lenses: %s", err)
	}

	expectedLenses = td.Slice(
		[]models.Lens{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Lens{
					Name: "iPod",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
			1: td.SStruct(
				models.Lens{
					Name: "X100F",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedLenses, expectedLenses)

	returnedLenses, err = FindLensesByID(s.DB, []int64{returnedLenses[0].ID})
	if err != nil {
		s.T().Fatalf("failed get lenses: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Lens{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Lens{
					Name: "iPod",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedLenses, expectedResult)
}

func (s *LensesSuite) TestDeleteLenses() {
	lenses := []models.Lens{
		{
			Name: "iPhone",
		},
		{
			Name: "X100F",
		},
	}

	returnedLenses, err := CreateLenses(s.DB, lenses)
	if err != nil {
		s.T().Fatalf("failed to create lenses: %s", err)
	}

	lensToDelete := returnedLenses[0]

	err = DeleteLenses(s.DB, []models.Lens{lensToDelete})
	s.Require().NoError(err, "unexpected error deleting lenses")

	allLenses, err := AllLenses(s.DB)
	if err != nil {
		s.T().Fatalf("failed get lenses: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Lens{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Lens{
					Name: "X100F",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), allLenses, expectedResult)
}

func (s *LensesSuite) TestFindLensByLensMatches() {
	lenses := []models.Lens{
		{
			Name:        "iPhone",
			LensMatches: "foobar",
		},
		{
			Name:        "X100F",
			LensMatches: "iPhone 11 Pro Max back triple camera 6mm f/2",
		},
	}

	returnedLenses, err := CreateLenses(s.DB, lenses)
	if err != nil {
		s.T().Fatalf("failed to create lenses needed for test: %s", err)
	}

	lens, err := FindLensByLensMatches(s.DB, "back triple camera")
	if err != nil {
		s.T().Fatalf("failed to find devices by model matches: %s", err)
	}

	if lens == nil {
		s.T().Fatalf("failed to find lens by lens matches")
	}

	td.Cmp(s.T(), *lens, returnedLenses[1])
}
