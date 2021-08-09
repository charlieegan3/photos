package database

import (
	"github.com/charlieegan3/cms/internal/pkg/models"
	"github.com/jmoiron/sqlx"
	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/suite"
)

// DevicesSuite is a number of tests to define the database integration for
// storing devices
type DevicesSuite struct {
	suite.Suite
	DB *sqlx.DB
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

	td.Cmp(s.T(), returnedDevices, devices)
}
