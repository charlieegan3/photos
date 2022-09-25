package cmd

import (
	bq "cloud.google.com/go/bigquery"
	"context"
	"google.golang.org/api/option"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/gcsblob"

	"github.com/charlieegan3/photos/internal/pkg/database"
	"github.com/charlieegan3/photos/internal/pkg/server"
)

// serverCmd wraps server.Serve and starts the cms webserver
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "start photos server",
	Run: func(cmd *cobra.Command, args []string) {
		environment := viper.GetString("environment")
		if environment != "production" && environment != "development" && environment != "test" {
			log.Fatalf("unknown environment %q", environment)
		}

		params := viper.GetStringMapString("database.params")
		connectionString := viper.GetString("database.connectionString")
		db, err := database.Init(connectionString, params, params["dbname"], viper.GetBool("database.createDatabase"))
		if err != nil {
			log.Fatalf("failed to init DB: %s", err)
		}

		driver, err := postgres.WithInstance(db, &postgres.Config{})
		m, err := migrate.NewWithDatabaseInstance(
			viper.GetString("database.migrationsPath"),
			"postgres",
			driver,
		)
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

		bucket, err := blob.OpenBucket(context.Background(), viper.GetString("bucket.url"))
		if err != nil {
			log.Fatalf("failed to open bucket: %s", err)
		}

		bigqueryClient, err := bq.NewClient(
			cmd.Context(),
			viper.GetString("google.bigquery.project_id"),
			option.WithCredentialsJSON([]byte(viper.GetString("google.service_account_key"))),
		)
		if err != nil {
			log.Fatalf("failed to create bigquery client: %s", err)
		}

		log.Println("Listening on", port)

		server.Serve(
			environment,
			viper.GetString("hostname"),
			viper.GetString("server.address"),
			port,
			viper.GetString("server.adminUsername"),
			viper.GetString("server.adminPassword"),
			db,
			bucket,
			bigqueryClient,
			viper.GetString("google.bigquery.dataset_id"),
			viper.GetString("google.bigquery.table_id"),
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
