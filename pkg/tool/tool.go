package tool

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"strings"

	"github.com/Jeffail/gabs/v2"
	"github.com/charlieegan3/toolbelt/pkg/apis"
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
		Database: true,
	}
}

func (p *PhotosWebsite) DatabaseMigrations() (*embed.FS, string, error) {
	return &database.Migrations, "migrations", nil
}

func (p *PhotosWebsite) DatabaseSet(db *sql.DB) {
	p.db = db
}

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
	if p.db == nil {
		return fmt.Errorf("database not set")
	}

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
