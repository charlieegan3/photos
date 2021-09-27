package database

import (
	"database/sql"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

func TestDatabaseSuite(t *testing.T) {
	suite.Run(t, new(DatabaseSuite))
}

// DatabaseSuite is the top of the test suite hierarchy for all tests that use
// the database.
type DatabaseSuite struct {
	suite.Suite
	DB *sql.DB
}

// SetupSuite configures the test database, dropping and recreating if need be
// to get a clean state
func (s *DatabaseSuite) SetupSuite() {
	// use viper as we do in commands to load in the config, this time, the
	// config is hardcoded to the test config file
	viper.SetConfigFile("../../../config.test.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		s.T().Fatalf("failed to load test config: %s", err)
	}

	// initialize a database connection to init the db
	params := viper.GetStringMapString("database.params")
	connectionString := viper.GetString("database.connection_string")
	db, err := Init(connectionString, params, "postgres", true)
	if err != nil {
		s.T().Fatalf("failed to init DB: %s", err)
	}

	// dbname must be set to a test db name
	dbname, ok := params["dbname"]
	if !ok {
		s.T().Fatalf("test dbname param was not set, failing as unsure what DB to use: %s", err)
	}

	// if the database exists, then we drop it to give a clean test state
	// this happens at the start of the test suite so that the state is there
	// after a test run to inspect if need be
	exists, err := Exists(db, dbname)
	if err != nil {
		s.T().Fatalf("failed to check if test DB exists: %s", err)
	}
	if exists {
		// drop existing test db
		err = Drop(db, dbname)
		if err != nil {
			s.T().Fatalf("failed to drop test database: %s", err)
		}
	}

	// create the test db for this test run
	err = Create(db, dbname)
	if err != nil {
		s.T().Fatalf("failed to create test database: %s", err)
	}

	// init the db for the test suite with the name of the new db
	s.DB, err = Init(connectionString, params, dbname, true)
	if err != nil {
		s.T().Fatalf("failed to init DB: %s", err)
	}

	// prepare to run the migrations
	driver, err := postgres.WithInstance(s.DB, &postgres.Config{})
	m, err := migrate.NewWithDatabaseInstance(
		"file://../../../migrations",
		"postgres",
		driver,
	)
	if err != nil {
		s.T().Fatalf("failed to load migrations: %s", err)
	}

	// migrate up, down and up again to test that both directions work
	err = m.Up()
	if err != nil {
		s.T().Fatalf("failed to Up migrate: %s", err)
	}
	err = m.Down()
	if err != nil {
		s.T().Fatalf("failed to Down migrate: %s", err)
	}
	err = m.Up()
	if err != nil {
		s.T().Fatalf("failed to re Up migrate: %s", err)
	}
}

func (s *DatabaseSuite) TestPing() {
	// example test, check that the connection is ok
	err := Ping(s.DB)
	if err != nil {
		s.T().Fatalf("failed to ping database: %s", err)
	}
}

//  Tests for dependent suites which use the database from the DatabaseSuite
//  follow

func (s *DatabaseSuite) TestDevicesSuite() {
	suite.Run(s.T(), &DevicesSuite{DB: s.DB})
}
