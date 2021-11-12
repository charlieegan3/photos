/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"context"
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

	"github.com/charlieegan3/photos/cms/internal/pkg/database"
	"github.com/charlieegan3/photos/cms/internal/pkg/server"
)

// serverCmd wraps server.Serve and starts the cms webserver
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "start cms server",
	Run: func(cmd *cobra.Command, args []string) {
		params := viper.GetStringMapString("database.params")
		connectionString := viper.GetString("database.connectionString")
		db, err := database.Init(connectionString, params, params["dbname"], viper.GetBool("database.createDatabase"))
		if err != nil {
			log.Fatalf("failed to init DB: %s", err)
		}

		driver, err := postgres.WithInstance(db, &postgres.Config{})
		m, err := migrate.NewWithDatabaseInstance(
			"file://migrations",
			"postgres",
			driver,
		)
		if err != nil {
			log.Fatalf("failed to load migrations: %s", err)
		}

		err = m.Up()
		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("failed to migrate up: %s", err)
		}

		port := viper.GetString("server.port")
		// PORT env var is used on heroku and should be used if set
		if p := os.Getenv("PORT"); p != "" {
			port = p
		}

		bucket, err := blob.OpenBucket(context.Background(), viper.GetString("bucket.url"))
		if err != nil {
			log.Fatalf("failed to open bucket: %s", err)
		}
		defer bucket.Close()

		log.Println("Listening on", port)

		server.Serve(
			viper.GetString("server.address"),
			port,
			viper.GetString("server.adminUsername"),
			viper.GetString("server.adminPassword"),
			db,
			bucket,
			viper.GetString("bucket.webUrl"),
		)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
