package cmd

import (
	"context"
	"fmt"
	"io"
	"log"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/gcsblob"

	"github.com/charlieegan3/photos/cms/internal/pkg/database"
	"github.com/charlieegan3/photos/cms/internal/pkg/mediametadata"
	"github.com/charlieegan3/photos/cms/internal/pkg/models"
)

var expTimeCmd = &cobra.Command{
	Use:   "exp",
	Short: "exp",
	Run: func(cmd *cobra.Command, args []string) {
		params := viper.GetStringMapString("database.params")
		connectionString := viper.GetString("database.connectionString")
		db, err := database.Init(connectionString, params, params["dbname"], viper.GetBool("database.createDatabase"))
		if err != nil {
			log.Fatalf("failed to init DB: %s", err)
		}

		bucket, err := blob.OpenBucket(context.Background(), viper.GetString("bucket.url"))
		if err != nil {
			log.Fatalf("failed to open bucket: %s", err)
		}
		defer bucket.Close()

		medias, err := database.AllMedias(db, false)
		if err != nil {
			log.Fatal(err.Error())
		}

		for _, m := range medias {
			if m.ShutterSpeed != 0 && m.ExposureTimeNumerator == 0 {

				br, err := bucket.NewReader(context.TODO(), fmt.Sprintf("media/%d.jpg", m.ID), nil)
				fileBytes, err := io.ReadAll(br)
				if err != nil {
					log.Fatal(err.Error())
				}

				exifData, err := mediametadata.ExtractMetadata(fileBytes)
				if err != nil {
					log.Fatal(err.Error())
				}

				m.Make = exifData.Make
				m.Model = exifData.Model
				m.TakenAt = exifData.DateTime
				m.FNumber, err = exifData.FNumber.ToDecimal()
				m.ShutterSpeed, err = exifData.ShutterSpeed.ToDecimal()
				m.ExposureTimeNumerator = exifData.ExposureTime.Numerator
				m.ExposureTimeDenominator = exifData.ExposureTime.Denominator
				m.ISOSpeed = int(exifData.ISOSpeed)
				m.Latitude, err = exifData.Latitude.ToDecimal()
				m.Longitude, err = exifData.Longitude.ToDecimal()
				m.Altitude, err = exifData.Altitude.ToDecimal()

				_, err = database.UpdateMedias(db, []models.Media{m})
				if err != nil {
					log.Fatal(err.Error())
				}
				fmt.Println(m.ID, "updated")
				br.Close()
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(expTimeCmd)
}
