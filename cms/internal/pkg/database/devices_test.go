package database

import (
	"database/sql"

	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/suite"

	"github.com/charlieegan3/photos/cms/internal/pkg/models"
)

// DevicesSuite is a number of tests to define the database integration for
// storing devices
type DevicesSuite struct {
	suite.Suite
	DB *sql.DB
}

func (s *DevicesSuite) SetupTest() {
	err := Truncate(s.DB, "devices")
	if err != nil {
		s.T().Fatalf("failed to truncate table: %s", err)
	}
}

func (s *DevicesSuite) TestCreateDevices() {
	devices := []models.Device{
		{
			Name:    "iPhone",
			IconURL: "https://example.com/image.jpg",
		},
		{
			Name:    "X100F",
			IconURL: "https://example.com/image2.jpg",
		},
	}

	returnedDevices, err := CreateDevices(s.DB, devices)
	if err != nil {
		s.T().Fatalf("failed to create devices: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Device{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Device{
					Name:    "iPhone",
					IconURL: "https://example.com/image.jpg",
				},
				td.StructFields{
					"ID":        td.Ignore(),
					"CreatedAt": td.Ignore(),
					"UpdatedAt": td.Ignore(),
				}),
			1: td.SStruct(
				models.Device{
					Name:    "X100F",
					IconURL: "https://example.com/image2.jpg",
				},
				td.StructFields{
					"ID":        td.Ignore(),
					"CreatedAt": td.Ignore(),
					"UpdatedAt": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedDevices, expectedResult)
}

func (s *DevicesSuite) TestFindDevicesByName() {
	devices := []models.Device{
		{
			Name:    "iPhone",
			IconURL: "https://example.com/image.jpg",
		},
		{
			Name:    "X100F",
			IconURL: "https://example.com/image2.jpg",
		},
	}

	_, err := CreateDevices(s.DB, devices)
	if err != nil {
		s.T().Fatalf("failed to create devices needed for test: %s", err)
	}

	returnedDevices, err := FindDevicesByName(s.DB, "X100F")
	if err != nil {
		s.T().Fatalf("failed get devices: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Device{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Device{
					Name:    "X100F",
					IconURL: "https://example.com/image2.jpg",
				},
				td.StructFields{
					"ID":        td.Ignore(),
					"CreatedAt": td.Ignore(),
					"UpdatedAt": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedDevices, expectedResult)
}

func (s *DevicesSuite) TestAllDevices() {
	devices := []models.Device{
		{
			Name:    "iPhone",
			IconURL: "https://example.com/image.jpg",
		},
		{
			Name:    "X100F",
			IconURL: "https://example.com/image2.jpg",
		},
	}

	_, err := CreateDevices(s.DB, devices)
	if err != nil {
		s.T().Fatalf("failed to create devices needed for test: %s", err)
	}

	returnedDevices, err := AllDevices(s.DB)
	if err != nil {
		s.T().Fatalf("failed get devices: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Device{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Device{
					Name:    "iPhone",
					IconURL: "https://example.com/image.jpg",
				},
				td.StructFields{
					"ID":        td.Ignore(),
					"CreatedAt": td.Ignore(),
					"UpdatedAt": td.Ignore(),
				}),
			1: td.SStruct(
				models.Device{
					Name:    "X100F",
					IconURL: "https://example.com/image2.jpg",
				},
				td.StructFields{
					"ID":        td.Ignore(),
					"CreatedAt": td.Ignore(),
					"UpdatedAt": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedDevices, expectedResult)
}

func (s *DevicesSuite) TestUpdateDevices() {
	initialDevices := []models.Device{
		{
			Name:    "iPhone",
			IconURL: "https://example.com/image.jpg",
		},
		{
			Name:    "X100F",
			IconURL: "https://example.com/image2.jpg",
		},
	}

	returnedDevices, err := CreateDevices(s.DB, initialDevices)
	if err != nil {
		s.T().Fatalf("failed to create devices needed for test: %s", err)
	}

	expectedDevices := td.Slice(
		[]models.Device{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Device{
					Name:    "iPhone",
					IconURL: "https://example.com/image.jpg",
				},
				td.StructFields{
					"ID":        td.Ignore(),
					"CreatedAt": td.Ignore(),
					"UpdatedAt": td.Ignore(),
				}),
			1: td.SStruct(
				models.Device{
					Name:    "X100F",
					IconURL: "https://example.com/image2.jpg",
				},
				td.StructFields{
					"ID":        td.Ignore(),
					"CreatedAt": td.Ignore(),
					"UpdatedAt": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedDevices, expectedDevices)

	updatedDevices := returnedDevices
	updatedDevices[0].Name = "iPod"

	returnedDevices, err = UpdateDevices(s.DB, updatedDevices)
	if err != nil {
		s.T().Fatalf("failed to update devices: %s", err)
	}

	expectedDevices = td.Slice(
		[]models.Device{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Device{
					Name:    "iPod",
					IconURL: "https://example.com/image.jpg",
				},
				td.StructFields{
					"ID":        td.Ignore(),
					"CreatedAt": td.Ignore(),
					"UpdatedAt": td.Ignore(),
				}),
			1: td.SStruct(
				models.Device{
					Name:    "X100F",
					IconURL: "https://example.com/image2.jpg",
				},
				td.StructFields{
					"ID":        td.Ignore(),
					"CreatedAt": td.Ignore(),
					"UpdatedAt": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedDevices, expectedDevices)

	returnedDevices, err = FindDevicesByName(s.DB, "iPod")
	if err != nil {
		s.T().Fatalf("failed get devices: %s", err)
	}

	expectedResult := td.Slice(
		[]models.Device{},
		td.ArrayEntries{
			0: td.SStruct(
				models.Device{
					Name:    "iPod",
					IconURL: "https://example.com/image.jpg",
				},
				td.StructFields{
					"ID":        td.Ignore(),
					"CreatedAt": td.Ignore(),
					"UpdatedAt": td.Ignore(),
				}),
		},
	)

	td.Cmp(s.T(), returnedDevices, expectedResult)
}
