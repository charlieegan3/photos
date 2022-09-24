package cmd

import (
	"log"
	"time"

	bq "cloud.google.com/go/bigquery"
	"github.com/charlieegan3/photos/internal/pkg/bigquery"
	"github.com/charlieegan3/photos/internal/pkg/database"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/api/option"
)

func init() {
	var from, to string
	var days int

	var archivePointsCmd = &cobra.Command{
		Use:   "archive-points",
		Short: "Archive points to BigQuery",
		Run: func(cmd *cobra.Command, args []string) {
			params := viper.GetStringMapString("database.params")
			connectionString := viper.GetString("database.connectionString")
			db, err := database.Init(connectionString, params, params["dbname"], viper.GetBool("database.createDatabase"))
			if err != nil {
				log.Fatalf("failed to init DB: %s", err)
			}

			var fromTime, toTime time.Time
			if days > 0 {
				fromTime = time.Now().AddDate(0, 0, -days)
				toTime = time.Now()
			} else {
				fromTime, err = time.Parse("2006-01-02", from)
				if err != nil {
					log.Fatalf("failed to parse from date: %s", err)
				}
				toTime, err = time.Parse("2006-01-02", to)
				if err != nil {
					log.Fatalf("failed to parse to date: %s", err)
				}
				// to the end of the day
				toTime.Add((24 * time.Hour) - 1)
			}

			points, err := database.PointsInRange(db, fromTime, toTime)
			if err != nil {
				log.Fatalf("failed to get points: %s", err)
			}

			client, err := bq.NewClient(
				cmd.Context(),
				viper.GetString("google.bigquery.project_id"),
				option.WithCredentialsJSON([]byte(viper.GetString("google.service_account_key"))),
			)
			if err != nil {
				log.Fatalf("failed to create bigquery client: %s", err)
			}

			unarchivedPoints, err := bigquery.UnarchivedPoints(
				cmd.Context(),
				client,
				points,
				viper.GetString("google.bigquery.dataset_id"),
				viper.GetString("google.bigquery.table_id"),
			)
			if err != nil {
				log.Fatalf("failed to get unarchived points: %s", err)
			}

			err = bigquery.InsertPoints(
				cmd.Context(),
				client,
				unarchivedPoints,
				viper.GetString("google.bigquery.dataset_id"),
				viper.GetString("google.bigquery.table_id"),
			)
			if err != nil {
				log.Fatalf("failed to archive points: %s", err)
			}

			log.Println("Points:", len(points))
			log.Println("New Points:", len(unarchivedPoints))
		},
	}

	archivePointsCmd.Flags().StringVarP(&from, "from", "", "", "from date")
	archivePointsCmd.Flags().StringVarP(&to, "to", "", "", "to date")
	archivePointsCmd.Flags().IntVarP(&days, "days", "", 0, "number of recent days to archive")
	jobsCmd.AddCommand(archivePointsCmd)
}
