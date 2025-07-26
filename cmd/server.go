package cmd

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gocloud.dev/blob"
	"gocloud.dev/blob/gcsblob"
	"gocloud.dev/gcp"
	"golang.org/x/oauth2/google"

	"github.com/charlieegan3/photos/internal/pkg/database"
	"github.com/charlieegan3/photos/internal/pkg/server"
)

// serverCmd wraps server.Serve and starts the cms webserver
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "start photos server",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		environment := viper.GetString("environment")
		if environment != "production" && environment != "development" && environment != "test" {
			log.Fatalf("unknown environment %q", environment)
		}

		params := viper.GetStringMapString("database.params")
		connectionString := viper.GetString("database.connectionString")
		db, err := database.Init(
			ctx,
			connectionString,
			params,
			params["dbname"],
			viper.GetBool("database.createDatabase"),
		)
		if err != nil {
			log.Fatalf("failed to init DB: %s", err)
		}

		conn, err := db.Conn(ctx)
		if err != nil {
			log.Fatalf("failed to get DB connection: %s", err)
		}

		// if the schema migrations table has the old name, move it
		_, err = conn.ExecContext(
			ctx,
			"ALTER TABLE IF EXISTS schema_migrations RENAME TO "+viper.GetString("database.migrationsTable"),
		)
		if err != nil {
			log.Fatalf("failed to rename migrations table: %s", err)
		}

		migrations := database.Migrations

		driver, err := postgres.WithConnection(ctx, conn, &postgres.Config{
			MigrationsTable: viper.GetString("database.migrationsTable"),
		})
		if err != nil {
			log.Fatalf("failed to init DB driver: %s", err)
		}

		source, err := iofs.New(migrations, "migrations")
		if err != nil {
			log.Fatalf("failed to create migration source: %s", err)
		}
		m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
		if err != nil {
			log.Fatalf("failed to load migrations: %s", err)
		}

		err = m.Up()
		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("failed to migrate up: %s", err)
		} else {
			log.Println("migrated up")
		}

		port := viper.GetString("server.port")
		// PORT env var will override config value
		if p := os.Getenv("PORT"); p != "" {
			port = p
		}

		var bucket *blob.Bucket
		if strings.HasPrefix(viper.GetString("bucket.url"), "gs://") {
			keyString := viper.GetString("google.service_account_key")
			if keyString == "" {
				log.Fatalf("failed to get default GCP credentials: %s", err)
			}

			creds, err := google.CredentialsFromJSON(
				context.Background(),
				[]byte(keyString),
				"https://www.googleapis.com/auth/cloud-platform",
			)
			if err != nil {
				log.Fatalf("failed to get default GCP credentials: %s", err)
			}

			client, err := gcp.NewHTTPClient(
				gcp.DefaultTransport(),
				gcp.CredentialsTokenSource(creds))
			if err != nil {
				log.Fatalf("failed to create bucket HTTP client: %s", err)
			}

			bucketName := strings.TrimPrefix(viper.GetString("bucket.url"), "gs://")

			bucket, err = gcsblob.OpenBucket(ctx, client, bucketName, nil)
			if err != nil {
				log.Fatalf("failed to open bucket: %s", err)
			}
		} else {
			bucket, err = blob.OpenBucket(ctx, viper.GetString("bucket.url"))
			if err != nil {
				log.Fatalf("failed to open bucket: %s", err)
			}
		}

		log.Printf("starting server on http://%s:%s", viper.GetString("hostname"), port)

		server.Serve(
			environment,
			viper.GetString("hostname"),
			viper.GetString("server.address"),
			port,
			viper.GetString("server.adminUsername"),
			viper.GetString("server.adminPassword"),
			db,
			bucket,
			viper.GetString("geoapify.url"),
			viper.GetString("geoapify.key"),
		)

		err = bucket.Close()
		if err != nil {
			log.Fatalf("failed to close bucket: %s", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
