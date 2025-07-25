package tool

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"strings"

	"github.com/Jeffail/gabs/v2"
	"github.com/charlieegan3/toolbelt/pkg/apis"
	"github.com/coreos/go-oidc"
	"github.com/gorilla/mux"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	"gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/gcsblob"
	"gocloud.dev/gcp"
	"golang.org/x/oauth2"
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

	host, scheme string

	environment, hostname         string
	mapServerURL, mapServerAPIKey string

	adminUserName, adminPassword string
	adminPath                    string
	adminParam                   string
	permittedEmailSuffix         string

	oauth2Config    *oauth2.Config
	oidcProvider    *oidc.Provider
	idTokenVerifier *oidc.IDTokenVerifier
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

	p.adminPath = "/admin"

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

	path = "server.address"
	address, ok := p.config.Path(path).Data().(string)
	if !ok {
		return fmt.Errorf("config value %s not set", path)
	}

	path = "server.port"
	port, ok := p.config.Path(path).Data().(string)
	if !ok {
		return fmt.Errorf("config value %s not set", path)
	}

	p.host = address
	if port != "" {
		p.host = address + ":" + port
	}

	path = "server.https"
	isHttps, ok := p.config.Path(path).Data().(bool)
	if !ok {
		return fmt.Errorf("config value %s not set", path)
	}

	p.scheme = "https://"
	if !isHttps {
		p.scheme = "http://"
	}

	if p.adminPassword == "" {
		path = "admin.auth.param"
		p.adminParam, ok = p.config.Path(path).Data().(string)
		if !ok {
			return fmt.Errorf("config value %s not set", path)
		}

		path = "admin.auth.provider_url"
		providerURL, ok := p.config.Path(path).Data().(string)
		if !ok {
			return fmt.Errorf("config value %s not set", path)
		}

		p.oidcProvider, err = oidc.NewProvider(context.TODO(), providerURL)
		if err != nil {
			return fmt.Errorf("failed to create oidc provider: %w", err)
		}

		p.oauth2Config = &oauth2.Config{
			Endpoint: p.oidcProvider.Endpoint(),
		}

		path = "admin.auth.client_id"
		p.oauth2Config.ClientID, ok = p.config.Path(path).Data().(string)
		if !ok {
			return fmt.Errorf("config value %s not set", path)
		}

		path = "admin.auth.client_secret"
		p.oauth2Config.ClientSecret, ok = p.config.Path(path).Data().(string)
		if !ok {
			return fmt.Errorf("config value %s not set", path)
		}

		p.oauth2Config.RedirectURL = p.scheme + p.host + "/admin/auth/callback"

		// offline_access is required for refresh tokens
		p.oauth2Config.Scopes = []string{oidc.ScopeOpenID, "profile", "email", "offline_access"}

		p.idTokenVerifier = p.oidcProvider.Verifier(&oidc.Config{ClientID: p.oauth2Config.ClientID})

		path = "admin.auth.permitted_email_suffix"
		p.permittedEmailSuffix, ok = p.config.Path(path).Data().(string)
		if !ok {
			return fmt.Errorf("config value %s not set", path)
		}
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
		p.db,
		p.bucket,
		p.mapServerURL, p.mapServerAPIKey,
		p.adminUserName, p.adminPassword,
		p.oauth2Config,
		p.idTokenVerifier,
		p.adminPath,
		p.adminParam,
		p.permittedEmailSuffix,
	)

	return err
}

func (p *PhotosWebsite) HTTPHost() string {
	return p.host
}
func (p *PhotosWebsite) HTTPPath() string { return "" }

func (p *PhotosWebsite) ExternalJobsFuncSet(f func(job apis.ExternalJob) error) {}
