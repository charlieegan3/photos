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
	var days int

	var expirePointsCmd = &cobra.Command{
		Use:   "expire-points",
		Short: "Delete old points data",
		Run: func(cmd *cobra.Command, args []string) {
			if olderThan == "" && days == 0 {
				log.Fatalf("--older-than or --days must be set to a value")
			}

			var olderThanDate time.Time
			if days > 0 {
				olderThanDate = time.Now().AddDate(0, 0, -days)
			} else {
				var err error
				olderThanDate, err = time.Parse("2006-01-02", olderThan)
				if err != nil {
					log.Fatalf("failed to parse older-than date: %s", err)
				}

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
	expirePointsCmd.Flags().IntVarP(&days, "days", "", 0, "more than N days ago")

	jobsCmd.AddCommand(expirePointsCmd)
}
