package database

import (
	"database/sql"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

// DatabaseSuite is the top of the test suite hierarchy for all tests that use
// the database.
type DatabaseSuite struct {
	suite.Suite
	DB *sql.DB
}

// SetupSuite configures the test database, dropping and recreating if need be
// to get a clean state
func (suite *DatabaseSuite) SetupSuite() {
	// use viper as we do in commands to load in the config, this time, the
	// config is hardcoded to the test config file
	viper.SetConfigFile("../../../config.test.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		suite.T().Fatalf("failed to load test config: %s", err)
	}

	// initialize the database connection
	params := viper.GetStringMapString("database.params")
	connectionString := viper.GetString("database.connection_string")
	suite.DB, err = Init(connectionString, params, "postgres", true)
	if err != nil {
		suite.T().Fatalf("failed to init DB: %s", err)
	}

	// dbname must be set to a test db name
	dbname, ok := params["dbname"]
	if !ok {
		suite.T().Fatalf("test dbname param was not set, failing as unsure what DB to use: %s", err)
	}

	// if the database exists, then we drop it to give a clean test state
	// this happens at the start of the test suite so that the state is there
	// after a test run to inspect if need be
	exists, err := Exists(suite.DB, dbname)
	if err != nil {
		suite.T().Fatalf("failed to check if test DB exists: %s", err)
	}
	if exists {
		// drop existing test db
		err = Drop(suite.DB, dbname)
		if err != nil {
			suite.T().Fatalf("failed to drop test database: %s", err)
		}
	}

	// create the test db for this test run
	err = Create(suite.DB, dbname)
	if err != nil {
		suite.T().Fatalf("failed to create test database: %s", err)
	}
}

func (suite *DatabaseSuite) TestExample() {
	// example test, check that the connection is ok
	err := Ping(suite.DB)
	if err != nil {
		suite.T().Fatalf("failed to ping database: %s", err)
	}

	// TODO add dependant suites here...
}

func TestDatabaseSuite(t *testing.T) {
	suite.Run(t, new(DatabaseSuite))
}
