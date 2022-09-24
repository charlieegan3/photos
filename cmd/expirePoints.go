package cmd

import (
	"github.com/charlieegan3/photos/internal/pkg/database"
	"github.com/spf13/viper"
	"log"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	var olderThan string
	var expirePointsCmd = &cobra.Command{
		Use:   "expire-points",
		Short: "Delete old points data",
		Run: func(cmd *cobra.Command, args []string) {
			if olderThan == "" {
				log.Fatalf("--older-than must be set to a value")
			}

			olderThanDate, err := time.Parse("2006-01-02", olderThan)
			if err != nil {
				log.Fatalf("failed to parse older-than date value: %s", err)
			}

			log.Println("Removing points older than", olderThanDate)

			params := viper.GetStringMapString("database.params")
			connectionString := viper.GetString("database.connectionString")
			db, err := database.Init(connectionString, params, params["dbname"], viper.GetBool("database.createDatabase"))
			if err != nil {
				log.Fatalf("failed to init DB: %s", err)
			}

			oldCount, err := database.CountPoints(db)
			if err != nil {
				log.Fatalf("failed to get point count: %s", err)
			}

			err = database.DeletePointsBefore(db, olderThanDate)
			if err != nil {
				log.Fatalf("failed to delete points: %s", err)
			}

			newCount, err := database.CountPoints(db)
			if err != nil {
				log.Fatalf("failed to get point count: %s", err)
			}

			log.Printf("Deleted %d points", oldCount-newCount)
		},
	}

	expirePointsCmd.Flags().StringVarP(&olderThan, "older-than", "", "", "older-than date")

	jobsCmd.AddCommand(expirePointsCmd)
}
