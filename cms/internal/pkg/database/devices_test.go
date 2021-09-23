package database

import (
	"database/sql"

	"github.com/charlieegan3/cms/internal/pkg/models"
	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/suite"
)

// DevicesSuite is a number of tests to define the database integration for
// storing devices
type DevicesSuite struct {
	suite.Suite
	DB *sql.DB
}

func (s *DevicesSuite) SetupSuite() {
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
