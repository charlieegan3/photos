package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/gcsblob"

	"github.com/charlieegan3/photos/cms/internal/pkg/database"
)

var videoCmd = &cobra.Command{
	Use: "video",
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
			if m.Kind == "mp4" {
				// posts, err := database.FindPostsByMediaID(db, m.ID)

				mediaFileName := fmt.Sprintf("./videos/%d.mp4", m.ID)
				if _, err := os.Stat(mediaFileName); os.IsNotExist(err) {
					br, err := bucket.NewReader(context.Background(), fmt.Sprintf("/media/%d.%s", m.ID, m.Kind), nil)
					if err != nil {
						log.Fatal(err.Error())
					}

					f, err := os.Create(mediaFileName)
					if err != nil {
						log.Fatal(err.Error())
					}

					_, err = io.Copy(f, br)
					if err != nil {
						log.Fatal(err.Error())
					}
					fmt.Println("saved", m.ID)

					f.Close()
					br.Close()
				}

				posts, err := database.FindPostsByMediaID(db, m.ID)
				if err != nil {
					log.Fatal(err.Error())
				}

				for _, p := range posts {
					postJSONFile := fmt.Sprintf("./videos/%d.%d.json", m.ID, p.ID)
					if _, err := os.Stat(postJSONFile); os.IsNotExist(err) {
						b, err := json.MarshalIndent(p, "", "  ")
						if err != nil {
							log.Fatal(err.Error())
						}

						f, err := os.Create(postJSONFile)
						if err != nil {
							log.Fatal(err.Error())
						}

						_, err = io.Copy(f, bytes.NewBuffer(b))
						if err != nil {
							log.Fatal(err.Error())
						}
						f.Close()
						fmt.Println("created post json", p.ID)
					}

					//	err := database.DeletePosts(db, []models.Post{p})
					//	if err != nil {
					//		log.Fatal(err.Error())
					//	}
					fmt.Println("deleted post", p.ID)
				}

				//	err = database.DeleteMedias(db, []models.Media{m})
				//	if err != nil {
				//		log.Fatal(err.Error())
				//	}
				fmt.Println("deleted media", m.ID)

				//	err = bucket.Delete(context.Background(), fmt.Sprintf("/media/%d.%s", m.ID, m.Kind))
				//	if err != nil {
				//		log.Fatal(err.Error())
				//	}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(videoCmd)
}
