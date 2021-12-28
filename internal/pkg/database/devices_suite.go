package database

import (
	"database/sql"

	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/require"
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
			Name: "iPhone",
		},
		{
			Name: "X100F",
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

	returnedDevices, err := CreateDevices(s.DB, devices)
	if err != nil {
		s.T().Fatalf("failed to create devices needed for test: %s", err)
	}

	returnedDevices, err = FindDevicesByID(s.DB, returnedDevices[0].ID)
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

	_, err := CreateDevices(s.DB, devices)
	if err != nil {
		s.T().Fatalf("failed to create devices needed for test: %s", err)
	}

	returnedDevices, err := FindDevicesByName(s.DB, "iPhone")
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
					Name: "IPhone",
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

func (s *DevicesSuite) TestUpdateDevices() {
	initialDevices := []models.Device{
		{
			Name: "iPhone",
		},
		{
			Name: "X100F",
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

	returnedDevices, err = UpdateDevices(s.DB, updatedDevices)
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

	returnedDevices, err = FindDevicesByID(s.DB, returnedDevices[0].ID)
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

	returnedDevices, err := CreateDevices(s.DB, devices)
	if err != nil {
		s.T().Fatalf("failed to create devices: %s", err)
	}

	deviceToDelete := returnedDevices[0]

	err = DeleteDevices(s.DB, []models.Device{deviceToDelete})
	require.NoError(s.T(), err, "unexpected error deleting devices")

	allDevices, err := AllDevices(s.DB)
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
