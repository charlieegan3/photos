package database

import (
	"database/sql"
	"time"

	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/charlieegan3/photos/cms/internal/pkg/models"
)

// MediasSuite is a number of tests to define the database integration for
// storing medias
type MediasSuite struct {
	suite.Suite
	DB *sql.DB
}

func (s *MediasSuite) SetupTest() {
	err := Truncate(s.DB, "medias")
	if err != nil {
		s.T().Fatalf("failed to truncate table: %s", err)
	}
}

func (s *MediasSuite) TestCreateMedias() {
	medias := []models.Media{
		{
			Make:  "FujiFilm",
			Model: "X100F",

			TakenAt: time.Date(2021, time.November, 23, 19, 56, 0, 0, time.UTC),

			FNumber:      2.0,
			ShutterSpeed: 0.004,
			ISOSpeed:     100,

			Latitude:  51.1,
			Longitude: 52.2,
			Altitude:  100.0,
		},
		{
			Make:  "Apple",
			Model: "iPhone",

			TakenAt: time.Date(2021, time.September, 22, 18, 56, 0, 0, time.UTC),

			FNumber:      4.0,
			ShutterSpeed: 0.05,
			ISOSpeed:     400,

			Latitude:  53.1,
			Longitude: 54.2,
			Altitude:  200.0,
		},
	}

	returnedMedias, err := CreateMedias(s.DB, medias)
	if err != nil {
		s.T().Fatalf("failed to create medias: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Media{},
		td.ArrayEntries{
			0: td.SStruct(
				medias[0],
				td.StructFields{
					"=*": td.Ignore(),
				}),
			1: td.SStruct(
				medias[1],
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedMedias, expectedResult)
}

func (s *MediasSuite) TestFindMediasByID() {
	medias := []models.Media{
		{
			Make:  "FujiFilm",
			Model: "X100F",

			TakenAt: time.Date(2021, time.November, 23, 19, 56, 0, 0, time.UTC),

			FNumber:      2.0,
			ShutterSpeed: 0.004,
			ISOSpeed:     100,

			Latitude:  51.1,
			Longitude: 52.2,
			Altitude:  100.0,
		},
		{
			Make:  "Apple",
			Model: "iPhone",

			TakenAt: time.Date(2021, time.September, 22, 18, 56, 0, 0, time.UTC),

			FNumber:      4.0,
			ShutterSpeed: 0.05,
			ISOSpeed:     400,

			Latitude:  53.1,
			Longitude: 54.2,
			Altitude:  200.0,
		},
	}

	returnedMedias, err := CreateMedias(s.DB, medias)
	if err != nil {
		s.T().Fatalf("failed to create medias needed for test: %s", err)
	}

	medias[0].ID = returnedMedias[0].ID

	returnedMedias, err = FindMediasByID(s.DB, medias[0].ID)
	if err != nil {
		s.T().Fatalf("failed get medias: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Media{},
		td.ArrayEntries{
			0: td.SStruct(
				medias[0],
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedMedias, expectedResult)
}

func (s *MediasSuite) TestAllMedias() {
	medias := []models.Media{
		{
			Make:  "FujiFilm",
			Model: "X100F",

			TakenAt: time.Date(2021, time.November, 23, 19, 56, 0, 0, time.UTC),

			FNumber:      2.0,
			ShutterSpeed: 0.004,
			ISOSpeed:     100,

			Latitude:  51.1,
			Longitude: 52.2,
			Altitude:  100.0,
		},
		{
			Make:  "Apple",
			Model: "iPhone",

			TakenAt: time.Date(2021, time.September, 22, 18, 56, 0, 0, time.UTC),

			FNumber:      4.0,
			ShutterSpeed: 0.05,
			ISOSpeed:     400,

			Latitude:  53.1,
			Longitude: 54.2,
			Altitude:  200.0,
		},
	}

	_, err := CreateMedias(s.DB, medias)
	if err != nil {
		s.T().Fatalf("failed to create medias needed for test: %s", err)
	}

	returnedMedias, err := AllMedias(s.DB)
	if err != nil {
		s.T().Fatalf("failed get medias: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Media{},
		td.ArrayEntries{
			0: td.SStruct(
				medias[0],
				td.StructFields{
					"=*": td.Ignore(),
				}),
			1: td.SStruct(
				medias[1],
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedMedias, expectedResult)
}

func (s *MediasSuite) TestDeleteMedias() {
	medias := []models.Media{
		{
			Make:  "FujiFilm",
			Model: "X100F",

			TakenAt: time.Date(2021, time.November, 23, 19, 56, 0, 0, time.UTC),

			FNumber:      2.0,
			ShutterSpeed: 0.004,
			ISOSpeed:     100,

			Latitude:  51.1,
			Longitude: 52.2,
			Altitude:  100.0,
		},
		{
			Make:  "Apple",
			Model: "iPhone",

			TakenAt: time.Date(2021, time.September, 22, 18, 56, 0, 0, time.UTC),

			FNumber:      4.0,
			ShutterSpeed: 0.05,
			ISOSpeed:     400,

			Latitude:  53.1,
			Longitude: 54.2,
			Altitude:  200.0,
		},
	}

	returnedMedias, err := CreateMedias(s.DB, medias)
	if err != nil {
		s.T().Fatalf("failed to create medias: %s", err)
	}

	mediaToDelete := returnedMedias[0]

	err = DeleteMedias(s.DB, []models.Media{mediaToDelete})
	require.NoError(s.T(), err, "unexpected error deleting medias")

	allMedias, err := AllMedias(s.DB)
	if err != nil {
		s.T().Fatalf("failed get medias: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Media{},
		td.ArrayEntries{
			0: td.SStruct(
				medias[1],
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), allMedias, expectedResult)
}

func (s *MediasSuite) TestUpdateMedias() {
	initialMedias := []models.Media{
		{
			Make:  "FujiFilm",
			Model: "X100F",

			TakenAt: time.Date(2021, time.November, 23, 19, 56, 0, 0, time.UTC),

			FNumber:      2.0,
			ShutterSpeed: 0.004,
			ISOSpeed:     100,

			Latitude:  51.1,
			Longitude: 52.2,
			Altitude:  100.0,
		},
		{
			Make:  "Apple",
			Model: "iPhone",

			TakenAt: time.Date(2021, time.September, 22, 18, 56, 0, 0, time.UTC),

			FNumber:      4.0,
			ShutterSpeed: 0.05,
			ISOSpeed:     400,

			Latitude:  53.1,
			Longitude: 54.2,
			Altitude:  200.0,
		},
	}

	returnedMedias, err := CreateMedias(s.DB, initialMedias)
	if err != nil {
		s.T().Fatalf("failed to create medias needed for test: %s", err)
	}

	expectedMedias := td.Slice(
		[]models.Media{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Media{
					Make:  "FujiFilm",
					Model: "X100F",

					TakenAt: time.Date(2021, time.November, 23, 19, 56, 0, 0, time.UTC),

					FNumber:      2.0,
					ShutterSpeed: 0.004,
					ISOSpeed:     100,

					Latitude:  51.1,
					Longitude: 52.2,
					Altitude:  100.0,
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
			1: td.SStruct(
				models.Media{
					Make:  "Apple",
					Model: "iPhone",

					TakenAt: time.Date(2021, time.September, 22, 18, 56, 0, 0, time.UTC),

					FNumber:      4.0,
					ShutterSpeed: 0.05,
					ISOSpeed:     400,

					Latitude:  53.1,
					Longitude: 54.2,
					Altitude:  200.0,
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedMedias, expectedMedias)

	updatedMedias := returnedMedias
	updatedMedias[0].Make = "Fuji"

	returnedMedias, err = UpdateMedias(s.DB, updatedMedias)
	if err != nil {
		s.T().Fatalf("failed to update medias: %s", err)
	}

	expectedMedias = td.Slice(
		[]models.Media{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Media{
					Make:  "Fuji",
					Model: "X100F",

					TakenAt: time.Date(2021, time.November, 23, 19, 56, 0, 0, time.UTC),

					FNumber:      2.0,
					ShutterSpeed: 0.004,
					ISOSpeed:     100,

					Latitude:  51.1,
					Longitude: 52.2,
					Altitude:  100.0,
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
			1: td.SStruct(
				models.Media{
					Make:  "Apple",
					Model: "iPhone",

					TakenAt: time.Date(2021, time.September, 22, 18, 56, 0, 0, time.UTC),

					FNumber:      4.0,
					ShutterSpeed: 0.05,
					ISOSpeed:     400,

					Latitude:  53.1,
					Longitude: 54.2,
					Altitude:  200.0,
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedMedias, expectedMedias)

	returnedMedias, err = FindMediasByID(s.DB, updatedMedias[0].ID)
	if err != nil {
		s.T().Fatalf("failed get medias: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Media{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Media{
					Make:  "Fuji",
					Model: "X100F",

					TakenAt: time.Date(2021, time.November, 23, 19, 56, 0, 0, time.UTC),

					FNumber:      2.0,
					ShutterSpeed: 0.004,
					ISOSpeed:     100,

					Latitude:  51.1,
					Longitude: 52.2,
					Altitude:  100.0,
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedMedias, expectedResult)
}
