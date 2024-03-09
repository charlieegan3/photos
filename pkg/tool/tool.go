package tool

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log"
	"strings"

	"github.com/Jeffail/gabs/v2"
	"github.com/charlieegan3/toolbelt/pkg/apis"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/gorilla/mux"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	"gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/gcsblob"
	"gocloud.dev/gcp"
	"golang.org/x/oauth2/google"

	"github.com/charlieegan3/photos/internal/pkg/database"
	"github.com/charlieegan3/photos/internal/pkg/server"
)

// PhotosWebsite is a tool that runs my personal site
type PhotosWebsite struct {
	config *gabs.Container

	db              *sql.DB
	migrationsTable string
	bucket          *blob.Bucket

	environment, hostname         string
	adminUserName, adminPassword  string
	mapServerURL, mapServerAPIKey string
}

func (p *PhotosWebsite) Name() string {
	return "photos"
}

func (p *PhotosWebsite) FeatureSet() apis.FeatureSet {
	return apis.FeatureSet{
		HTTP:     true,
		HTTPHost: true,
		Config:   true,
	}
}

// Due to legacy reasons, photos manages it's own migrations
func (p *PhotosWebsite) DatabaseMigrations() (*embed.FS, string, error) {
	return nil, "migrations", nil
}

func (p *PhotosWebsite) DatabaseSet(db *sql.DB) {}

func (p *PhotosWebsite) SetConfig(config map[string]any) error {
	var ok bool
	var err error
	var path string
	p.config = gabs.Wrap(config)

	path = "environment"
	p.environment, ok = p.config.Path(path).Data().(string)
	if !ok {
		return fmt.Errorf("config value %s not set", path)
	}

	if p.environment != "production" && p.environment != "development" && p.environment != "test" {
		return fmt.Errorf("invalid environment %s", p.environment)
	}

	path = "database.params"

	rawDatabaseParams, ok := p.config.Path(path).Data().(map[string]any)
	if !ok {
		return fmt.Errorf("config value %s not set", path)
	}

	databaseParams := make(map[string]string)
	for k, v := range rawDatabaseParams {
		databaseParams[k], ok = v.(string)
		if !ok {
			return fmt.Errorf("config value %s not set", path)
		}
	}

	path = "database.connection_string"
	connectionString, ok := p.config.Path(path).Data().(string)
	if !ok {
		return fmt.Errorf("config value %s not set", path)
	}

	path = "database.create_database"
	createDatabase, ok := p.config.Path(path).Data().(bool)
	if !ok {
		createDatabase = false
	}

	path = "database.migrations_table"
	p.migrationsTable, ok = p.config.Path(path).Data().(string)
	if !ok {
		return fmt.Errorf("config value %s not set", path)
	}

	p.db, err = database.Init(
		connectionString,
		databaseParams,
		databaseParams["dbname"],
		createDatabase,
	)
	if err != nil {
		return fmt.Errorf("failed to init database: %s", err)
	}

	conn, err := p.db.Conn(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get DB connection: %s", err)
	}
	// close to avoid leaking connection for migrations
	defer conn.Close()

	// if the schema migrations table has the old name, move it
	_, err = conn.ExecContext(
		context.Background(),
		"ALTER TABLE IF EXISTS schema_migrations RENAME TO "+p.migrationsTable,
	)
	if err != nil {
		log.Fatalf("failed to rename migrations table: %s", err)
	}

	fmt.Println("migrations table", p.migrationsTable)

	migrations := database.Migrations

	migFiles, err := migrations.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations: %s", err)
	}
	for _, f := range migFiles {
		fmt.Println(f.Name())
	}

	driver, err := postgres.WithConnection(context.Background(), conn, &postgres.Config{
		MigrationsTable: p.migrationsTable,
	})
	if err != nil {
		return fmt.Errorf("failed to init migrations driver: %s", err)
	}

	source, err := iofs.New(migrations, "migrations")
	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to init migrations: %s", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to migrate up: %s", err)
	}

	path = "bucket.url"
	bucketURL, ok := p.config.Path(path).Data().(string)
	if !ok {
		return fmt.Errorf("config value %s not set", path)
	}

	// handle for when running in production
	if strings.HasPrefix(bucketURL, "gs://") {
		path = "google.service_account_key"
		keyString, ok := p.config.Path(path).Data().(string)
		if !ok {
			return fmt.Errorf("config value %s not set", path)
		}

		creds, err := google.CredentialsFromJSON(
			context.Background(),
			[]byte(keyString),
			"https://www.googleapis.com/auth/cloud-platform",
		)
		if err != nil {
			return fmt.Errorf("failed to create credentials: %s", err)
		}

		client, err := gcp.NewHTTPClient(
			gcp.DefaultTransport(),
			gcp.CredentialsTokenSource(creds))
		if err != nil {
			return fmt.Errorf("failed to create client: %s", err)
		}

		bucketName := strings.TrimPrefix(bucketURL, "gs://")

		p.bucket, err = gcsblob.OpenBucket(context.Background(), client, bucketName, nil)
		if err != nil {
			return fmt.Errorf("failed to open bucket: %s", err)
		}
	} else {
		p.bucket, err = blob.OpenBucket(context.Background(), bucketURL)
		if err != nil {
			return fmt.Errorf("failed to open bucket: %s", err)
		}
	}

	path = "geoapify.url"
	p.mapServerURL, ok = p.config.Path(path).Data().(string)
	if !ok {
		return fmt.Errorf("config value %s not set", path)
	}

	path = "geoapify.key"
	p.mapServerAPIKey, ok = p.config.Path(path).Data().(string)
	if !ok {
		return fmt.Errorf("config value %s not set", path)
	}

	path = "admin.username"
	p.adminUserName, ok = p.config.Path(path).Data().(string)
	if !ok {
		return fmt.Errorf("config value %s not set", path)
	}

	path = "admin.password"
	p.adminPassword, ok = p.config.Path(path).Data().(string)
	if !ok {
		return fmt.Errorf("config value %s not set", path)
	}

	return nil
}

func (p *PhotosWebsite) Jobs() ([]apis.Job, error) { return []apis.Job{}, nil }

func (p *PhotosWebsite) HTTPAttach(router *mux.Router) error {
	err := server.Attach(
		router,
		p.adminUserName, p.adminPassword,
		p.db,
		p.bucket,
		p.mapServerURL, p.mapServerAPIKey,
	)

	return err
}
func (p *PhotosWebsite) HTTPHost() string {
	path := "hostname"
	host, ok := p.config.Path(path).Data().(string)
	if !ok {
		return "example.com"
	}
	return host
}
func (p *PhotosWebsite) HTTPPath() string { return "" }

func (p *PhotosWebsite) ExternalJobsFuncSet(f func(job apis.ExternalJob) error) {}
