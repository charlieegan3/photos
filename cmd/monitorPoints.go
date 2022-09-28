package cmd

import (
	"bytes"
	bq "cloud.google.com/go/bigquery"
	"encoding/json"
	"fmt"
	"github.com/charlieegan3/photos/internal/pkg/bigquery"
	"github.com/charlieegan3/photos/internal/pkg/database"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/api/option"
	"log"
	"net/http"
	"time"
)

func notify(title, message string) {
	log.Println("notifying", title, message)
	datab := []map[string]string{
		{
			"title": title,
			"body":  message,
			"url":   "",
		},
	}

	b, err := json.Marshal(datab)
	if err != nil {
		log.Fatal(fmt.Sprintf("failed to marshal data: %s", err))
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", viper.GetString("notification_webhook.endpoint"), bytes.NewBuffer(b))
	if err != nil {
		log.Fatal(fmt.Sprintf("failed to create request: %s", err))
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")

	_, err = client.Do(req)
	if err != nil {
		log.Fatal(fmt.Sprintf("failed to send request: %s", err))
	}
}

var monitorPointsCmd = &cobra.Command{
	Use:   "monitor-points",
	Short: "Monitor the delta between current time, points in the database and points in BiqQuery",
	Run: func(cmd *cobra.Command, args []string) {
		params := viper.GetStringMapString("database.params")
		connectionString := viper.GetString("database.connectionString")
		db, err := database.Init(connectionString, params, params["dbname"], viper.GetBool("database.createDatabase"))
		if err != nil {
			notify("photos: Points Monitor Error", fmt.Sprintf("failed to init DB: %s", err))
			return
		}

		currentTime := time.Now()
		fmt.Println("current time\t", currentTime)

		databasePoints, err := database.PointsInRange(db, currentTime.AddDate(0, 0, -30), currentTime)
		if err != nil {
			notify("photos: Points Monitor Error", fmt.Sprintf("failed to get points from database: %s", err))
			return
		}
		latestDatabasePoint := databasePoints[len(databasePoints)-1].CreatedAt
		fmt.Println("latest database\t", latestDatabasePoint)

		client, err := bq.NewClient(
			cmd.Context(),
			viper.GetString("google.bigquery.project_id"),
			option.WithCredentialsJSON([]byte(viper.GetString("google.service_account_key"))),
		)
		if err != nil {
			notify("photos: Points Monitor Error", fmt.Sprintf("failed to init BigQuery client: %s", err))
			return
		}

		bqPoints, err := bigquery.PointsInRange(cmd.Context(), client,
			viper.GetString("google.bigquery.dataset_id"),
			viper.GetString("google.bigquery.table_id"),
			currentTime.AddDate(0, 0, -30),
			currentTime,
		)
		if err != nil {
			notify("photos: Points Monitor Error", fmt.Sprintf("failed to get points from BigQuery: %s", err))
		}
		latestBqPoint := bqPoints[len(bqPoints)-1].CreatedAt
		fmt.Println("latest bq\t", latestBqPoint)

		if diff := latestDatabasePoint.Sub(latestBqPoint); diff > 4*24*time.Hour {
			notify("photos: Sync Offset Warning", fmt.Sprintf("bigquery is %s behind database", diff))
		}

		if diff := currentTime.Sub(latestDatabasePoint); diff > 4*24*time.Hour {
			notify("photos: Sync Offset Warning", fmt.Sprintf("database is %s behind current time", diff))
		}
	},
}

func init() {
	jobsCmd.AddCommand(monitorPointsCmd)
}
