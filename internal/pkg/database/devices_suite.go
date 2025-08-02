package database

import (
	"database/sql"
	"time"

	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/suite"

	"github.com/charlieegan3/photos/internal/pkg/models"
)

// DevicesSuite is a number of tests to define the database integration for
// storing devices.
type DevicesSuite struct {
	suite.Suite

	DB *sql.DB
}

func (s *DevicesSuite) SetupTest() {
	err := Truncate(s.T().Context(), s.DB, "photos.devices")
	if err != nil {
		s.T().Fatalf("failed to truncate table: %s", err)
	}
}

func (s *DevicesSuite) TestMostRecentlyUsedDevice() {
	devices := []models.Device{
		{
			Name: "iPhone",
		},
		{
			Name: "X100F",
		},
	}

	deviceRepo := NewDeviceRepository(s.DB)
	returnedDevices, err := deviceRepo.Create(s.T().Context(), devices)
	if err != nil {
		s.T().Fatalf("failed to create devices: %s", err)
	}

	medias := []models.Media{
		{
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
	_, err = CreateMedias(s.T().Context(), s.DB, medias)
	if err != nil {
		s.T().Fatalf("failed to create medias: %s", err)
	}

	device, err := deviceRepo.MostRecentlyUsed(s.T().Context())
	if err != nil {
		s.T().Fatalf("failed to create medias: %s", err)
	}

	td.Cmp(s.T(), device, returnedDevices[0])
}

func (s *DevicesSuite) TestCreateDevices() {
	devices := []models.Device{
		{
			Name: "iPhone",
		},
		{
			Name: "X100F",
		},
	}

	deviceRepo := NewDeviceRepository(s.DB)
	returnedDevices, err := deviceRepo.Create(s.T().Context(), devices)
	if err != nil {
		s.T().Fatalf("failed to create devices: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Device{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Device{
					Name: "iPhone",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
			1: td.SStruct(
				models.Device{
					Name: "X100F",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedDevices, expectedResult)
}

func (s *DevicesSuite) TestFindDevicesByID() {
	devices := []models.Device{
		{
			Name: "iPhone",
		},
		{
			Name: "X100F",
		},
	}

	deviceRepo := NewDeviceRepository(s.DB)
	returnedDevices, err := deviceRepo.Create(s.T().Context(), devices)
	if err != nil {
		s.T().Fatalf("failed to create devices needed for test: %s", err)
	}

	returnedDevices, err = deviceRepo.FindByIDs(s.T().Context(), []int64{returnedDevices[0].ID})
	if err != nil {
		s.T().Fatalf("failed get devices: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Device{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Device{
					Name: "iPhone",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedDevices, expectedResult)
}

func (s *DevicesSuite) TestFindDevicesByName() {
	devices := []models.Device{
		{
			Name: "iPhone",
		},
		{
			Name: "X100F",
		},
	}

	deviceRepo := NewDeviceRepository(s.DB)
	_, err := deviceRepo.Create(s.T().Context(), devices)
	if err != nil {
		s.T().Fatalf("failed to create devices needed for test: %s", err)
	}

	returnedDevices, err := deviceRepo.FindByField(s.T().Context(), "name", "iPhone")
	if err != nil {
		s.T().Fatalf("failed get devices: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Device{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Device{
					Name: "iPhone",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedDevices, expectedResult)
}

func (s *DevicesSuite) TestAllDevices() {
	devices := []models.Device{
		{
			Name: "IPhone",
		},
		{
			Name: "X100F",
		},
	}

	deviceRepo := NewDeviceRepository(s.DB)
	_, err := deviceRepo.Create(s.T().Context(), devices)
	if err != nil {
		s.T().Fatalf("failed to create devices needed for test: %s", err)
	}

	returnedDevices, err := deviceRepo.All(s.T().Context())
	if err != nil {
		s.T().Fatalf("failed get devices: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Device{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Device{
					Name: "X100F",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
			1: td.SStruct(
				models.Device{
					Name: "IPhone",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedDevices, expectedResult)
}

func (s *DevicesSuite) TestUpdateDevices() {
	initialDevices := []models.Device{
		{
			Name: "iPhone",
		},
		{
			Name: "X100F",
		},
	}

	deviceRepo := NewDeviceRepository(s.DB)
	returnedDevices, err := deviceRepo.Create(s.T().Context(), initialDevices)
	if err != nil {
		s.T().Fatalf("failed to create devices needed for test: %s", err)
	}

	expectedDevices := td.Slice(
		[]models.Device{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Device{
					Name: "iPhone",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
			1: td.SStruct(
				models.Device{
					Name: "X100F",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedDevices, expectedDevices)

	updatedDevices := returnedDevices
	updatedDevices[0].Name = "iPod"

	returnedDevices, err = deviceRepo.Update(s.T().Context(), updatedDevices)
	if err != nil {
		s.T().Fatalf("failed to update devices: %s", err)
	}

	expectedDevices = td.Slice(
		[]models.Device{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Device{
					Name: "iPod",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
			1: td.SStruct(
				models.Device{
					Name: "X100F",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedDevices, expectedDevices)

	returnedDevices, err = deviceRepo.FindByIDs(s.T().Context(), []int64{returnedDevices[0].ID})
	if err != nil {
		s.T().Fatalf("failed get devices: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Device{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Device{
					Name: "iPod",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedDevices, expectedResult)
}

func (s *DevicesSuite) TestDeleteDevices() {
	devices := []models.Device{
		{
			Name: "iPhone",
		},
		{
			Name: "X100F",
		},
	}

	deviceRepo := NewDeviceRepository(s.DB)
	returnedDevices, err := deviceRepo.Create(s.T().Context(), devices)
	if err != nil {
		s.T().Fatalf("failed to create devices: %s", err)
	}

	deviceToDelete := returnedDevices[0]

	err = deviceRepo.Delete(s.T().Context(), []models.Device{deviceToDelete})
	s.Require().NoError(err, "unexpected error deleting devices")

	allDevices, err := deviceRepo.All(s.T().Context())
	if err != nil {
		s.T().Fatalf("failed get devices: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Device{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Device{
					Name: "X100F",
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), allDevices, expectedResult)
}

func (s *DevicesSuite) TestDevicePosts() {
	devices := []models.Device{
		{
			Name: "iPhone",
		},
		{
			Name: "X100F",
		},
	}

	deviceRepo := NewDeviceRepository(s.DB)
	returnedDevices, err := deviceRepo.Create(s.T().Context(), devices)
	if err != nil {
		s.T().Fatalf("failed to create devices: %s", err)
	}

	medias := []models.Media{
		{DeviceID: returnedDevices[0].ID},
		{DeviceID: returnedDevices[1].ID},
	}
	returnedMedias, err := CreateMedias(s.T().Context(), s.DB, medias)
	if err != nil {
		s.T().Fatalf("failed to create medias: %s", err)
	}

	locations := []models.Location{
		{Name: "London"},
	}
	returnedLocations, err := CreateLocations(s.T().Context(), s.DB, locations)
	s.Require().NoError(err)

	posts := []models.Post{
		{
			Description: "post from 2022",
			PublishDate: time.Date(2021, time.January, 1, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[0].ID,
			LocationID:  returnedLocations[0].ID,
		},
		{
			Description: "post from 2021",
			PublishDate: time.Date(2021, time.January, 1, 19, 56, 0, 0, time.UTC),
			MediaID:     returnedMedias[1].ID,
			LocationID:  returnedLocations[0].ID,
		},
	}
	returnedPosts, err := CreatePosts(s.T().Context(), s.DB, posts)
	s.Require().NoError(err)

	result, err := deviceRepo.Posts(s.T().Context(), returnedDevices[0].ID)
	if err != nil {
		s.T().Fatalf("failed to get posts for devices: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Post{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Post{
					ID: returnedPosts[0].ID,
				},
				td.StructFields{
					"=*": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), result, expectedResult)
}

func (s *DevicesSuite) TestFindDeviceByModelMatches() {
	devices := []models.Device{
		{
			Name:         "iPhone",
			ModelMatches: "iPhone 11 Pro Max",
		},
		{
			Name:         "X100F",
			ModelMatches: "X100F",
		},
	}

	deviceRepo := NewDeviceRepository(s.DB)
	returnedDevices, err := deviceRepo.Create(s.T().Context(), devices)
	if err != nil {
		s.T().Fatalf("failed to create devices: %s", err)
	}

	device, err := deviceRepo.FindByModelMatches(s.T().Context(), "iPhone 11 Pro Max")
	if err != nil {
		s.T().Fatalf("failed to find devices by model matches: %s", err)
	}

	td.Cmp(s.T(), *device, returnedDevices[0])
}
