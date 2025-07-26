package databasetest

import (
	"context"
	"database/sql"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"gocloud.dev/blob"

	"github.com/charlieegan3/photos/internal/pkg/database"
	"github.com/charlieegan3/photos/internal/pkg/server/handlers/admin/devices"
	"github.com/charlieegan3/photos/internal/pkg/server/handlers/admin/lenses"
	"github.com/charlieegan3/photos/internal/pkg/server/handlers/admin/locations"
	"github.com/charlieegan3/photos/internal/pkg/server/handlers/admin/medias"
	"github.com/charlieegan3/photos/internal/pkg/server/handlers/admin/posts"
	"github.com/charlieegan3/photos/internal/pkg/server/handlers/admin/tags"
	"github.com/charlieegan3/photos/internal/pkg/server/handlers/admin/trips"
	publicdevices "github.com/charlieegan3/photos/internal/pkg/server/handlers/public/devices"
	publiclenses "github.com/charlieegan3/photos/internal/pkg/server/handlers/public/lenses"
	publiclocations "github.com/charlieegan3/photos/internal/pkg/server/handlers/public/locations"
	publicmedias "github.com/charlieegan3/photos/internal/pkg/server/handlers/public/medias"
	publicposts "github.com/charlieegan3/photos/internal/pkg/server/handlers/public/posts"
	publictags "github.com/charlieegan3/photos/internal/pkg/server/handlers/public/tags"
	publictrips "github.com/charlieegan3/photos/internal/pkg/server/handlers/public/trips"
)

func TestDatabaseSuite(t *testing.T) {
	t.Parallel()

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
	viper.SetConfigFile("../../../../config.test.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		s.T().Fatalf("failed to load test config: %s", err)
	}

	// initialize a database connection to init the db
	params := viper.GetStringMapString("database.params")
	connectionString := viper.GetString("database.connectionString")
	migrationsTable := viper.GetString("database.migrationsTable")
	db, err := database.Init(s.T().Context(), connectionString, params, "postgres", false)
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
	exists, err := database.Exists(s.T().Context(), db, dbname)
	if err != nil {
		s.T().Fatalf("failed to check if test DB exists: %s", err)
	}
	if exists {
		// drop existing test db
		err = database.Drop(s.T().Context(), db, dbname)
		if err != nil {
			s.T().Fatalf("failed to drop test database: %s", err)
		}
	}

	// create the test db for this test run
	err = database.Create(s.T().Context(), db, dbname)
	if err != nil {
		s.T().Fatalf("failed to create test database: %s", err)
	}

	// init the db for the test suite with the name of the new db
	s.DB, err = database.Init(s.T().Context(), connectionString, params, dbname, true)
	if err != nil {
		s.T().Fatalf("failed to init DB: %s", err)
	}

	conn, err := s.DB.Conn(context.Background())
	if err != nil {
		s.T().Fatalf("failed to get connection: %s", err)
	}
	// close to avoid leaking connection for migrations
	defer conn.Close()

	migrations := database.Migrations

	driver, err := postgres.WithConnection(context.Background(), conn, &postgres.Config{
		MigrationsTable: migrationsTable,
	})
	if err != nil {
		s.T().Fatalf("failed to init migrations driver: %s", err)
	}

	source, err := iofs.New(migrations, "migrations")
	if err != nil {
		s.T().Fatalf("failed to create migration source: %s", err)
	}
	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		s.T().Fatalf("failed to init migrations: %s", err)
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
	err := database.Ping(s.T().Context(), s.DB)
	if err != nil {
		s.T().Fatalf("failed to ping database: %s", err)
	}
}

//  Tests for dependent suites which use the database from the DatabaseSuite
//  follow

func (s *DatabaseSuite) TestDevicesSuite() {
	suite.Run(s.T(), &database.DevicesSuite{DB: s.DB})
}

func (s *DatabaseSuite) TestLensesSuite() {
	suite.Run(s.T(), &database.LensesSuite{DB: s.DB})
}

func (s *DatabaseSuite) TestTagsSuite() {
	suite.Run(s.T(), &database.TagsSuite{DB: s.DB})
}

func (s *DatabaseSuite) TestLocationsSuite() {
	suite.Run(s.T(), &database.LocationsSuite{DB: s.DB})
}

func (s *DatabaseSuite) TestMediasSuite() {
	suite.Run(s.T(), &database.MediasSuite{DB: s.DB})
}

func (s *DatabaseSuite) TestPostsSuite() {
	suite.Run(s.T(), &database.PostsSuite{DB: s.DB})
}

func (s *DatabaseSuite) TestTaggingsSuite() {
	suite.Run(s.T(), &database.TaggingsSuite{DB: s.DB})
}

func (s *DatabaseSuite) TestTripsSuite() {
	suite.Run(s.T(), &database.TripsSuite{DB: s.DB})
}

func (s *DatabaseSuite) TestEndpointsDevicesSuite() {
	// TODO move to suite to be shared
	bucketBaseURL := "mem://test_bucket/"
	bucket, err := blob.OpenBucket(context.Background(), bucketBaseURL)
	s.Require().NoError(err)
	defer bucket.Close()

	suite.Run(s.T(), &devices.EndpointsDevicesSuite{
		DB:            s.DB,
		Bucket:        bucket,
		BucketBaseURL: bucketBaseURL,
	})
}

func (s *DatabaseSuite) TestEndpointsTripsSuite() {
	suite.Run(s.T(), &trips.EndpointsTripsSuite{
		DB: s.DB,
	})
}

func (s *DatabaseSuite) TestEndpointsLensesSuite() {
	// TODO move to suite to be shared
	bucketBaseURL := "mem://test_bucket/"
	bucket, err := blob.OpenBucket(context.Background(), bucketBaseURL)
	s.Require().NoError(err)
	defer bucket.Close()

	suite.Run(s.T(), &lenses.EndpointsLensesSuite{
		DB:            s.DB,
		Bucket:        bucket,
		BucketBaseURL: bucketBaseURL,
	})
}

func (s *DatabaseSuite) TestEndpointsPostsSuite() {
	suite.Run(s.T(), &posts.EndpointsPostsSuite{
		DB: s.DB,
	})
}

func (s *DatabaseSuite) TestEndpointsTagsSuite() {
	suite.Run(s.T(), &tags.EndpointsTagsSuite{
		DB: s.DB,
	})
}

func (s *DatabaseSuite) TestEndpointsLocationsSuite() {
	bucketBaseURL := "mem://test_bucket/"
	bucket, err := blob.OpenBucket(context.Background(), bucketBaseURL)
	s.Require().NoError(err)
	defer bucket.Close()
	suite.Run(s.T(), &locations.EndpointsLocationsSuite{
		DB:     s.DB,
		Bucket: bucket,
	})
}

func (s *DatabaseSuite) TestPublicPostsSuite() {
	suite.Run(s.T(), &publicposts.PostsSuite{
		DB: s.DB,
	})
}

func (s *DatabaseSuite) TestPublicTagsSuite() {
	suite.Run(s.T(), &publictags.TagsSuite{
		DB: s.DB,
	})
}

func (s *DatabaseSuite) TestPublicLocationsSuite() {
	bucketBaseURL := "mem://test_bucket/"
	bucket, err := blob.OpenBucket(context.Background(), bucketBaseURL)
	s.Require().NoError(err)
	defer bucket.Close()

	suite.Run(s.T(), &publiclocations.LocationsSuite{
		DB:     s.DB,
		Bucket: bucket,
	})
}

func (s *DatabaseSuite) TestPublicTripsSuite() {
	suite.Run(s.T(), &publictrips.TripsSuite{
		DB: s.DB,
	})
}

func (s *DatabaseSuite) TestPublicDevicesSuite() {
	bucketBaseURL := "mem://test_bucket/"
	bucket, err := blob.OpenBucket(context.Background(), bucketBaseURL)
	s.Require().NoError(err)
	defer bucket.Close()

	suite.Run(s.T(), &publicdevices.DevicesSuite{
		DB:     s.DB,
		Bucket: bucket,
	})
}

func (s *DatabaseSuite) TestPublicLensesSuite() {
	bucketBaseURL := "mem://test_bucket/"
	bucket, err := blob.OpenBucket(context.Background(), bucketBaseURL)
	s.Require().NoError(err)
	defer bucket.Close()

	suite.Run(s.T(), &publiclenses.LensesSuite{
		DB:     s.DB,
		Bucket: bucket,
	})
}

func (s *DatabaseSuite) TestPublicMediasSuite() {
	bucketBaseURL := "mem://test_bucket/"
	bucket, err := blob.OpenBucket(context.Background(), bucketBaseURL)
	s.Require().NoError(err)
	defer bucket.Close()

	suite.Run(s.T(), &publicmedias.MediasSuite{
		DB:     s.DB,
		Bucket: bucket,
	})
}

func (s *DatabaseSuite) TestEndpointsMediasSuite() {
	// TODO move to suite to be shared
	bucketBaseURL := "mem://test_bucket/"
	bucket, err := blob.OpenBucket(context.Background(), bucketBaseURL)
	s.Require().NoError(err)
	defer bucket.Close()

	suite.Run(s.T(), &medias.EndpointsMediasSuite{
		DB:            s.DB,
		Bucket:        bucket,
		BucketBaseURL: bucketBaseURL,
	})
}
