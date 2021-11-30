package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/charlieegan3/photos/cms/internal/pkg/database"
	"github.com/charlieegan3/photos/cms/internal/pkg/models"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/gcsblob"
)

type rawLocation struct {
	ID   string  `json:"id"`
	Lat  float64 `json:"lat"`
	Long float64 `json:"long"`
	Name string  `json:"name"`
}
type rawPost struct {
	Caption    string `json:"caption"`
	Code       string `json:"code"`
	Dimensions struct {
		Height int64 `json:"height"`
		Width  int64 `json:"width"`
	} `json:"dimensions"`
	DisplayURL string `json:"display_url"`
	ID         string `json:"id"`
	IsVideo    bool   `json:"is_video"`
	Location   struct {
		AddressJSON   string `json:"address_json"`
		HasPublicPage bool   `json:"has_public_page"`
		ID            string `json:"id"`
		Name          string `json:"name"`
		Slug          string `json:"slug"`
	} `json:"location"`
	MediaURL  string   `json:"media_url"`
	PostURL   string   `json:"post_url"`
	Tags      []string `json:"tags"`
	Timestamp int64    `json:"timestamp"`
}

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "import existing data",
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

		basePathLocations := "../locations/"
		basePathPosts := "../completed_json/"

		// import locations
		locationMap := make(map[string]int)
		rawLocations, err := ioutil.ReadDir(basePathLocations)
		if err != nil {
			log.Fatalf("failed to list locations source: %s", err)
		}
		for _, locationFile := range rawLocations {
			jsonFile, err := os.Open(basePathLocations + locationFile.Name())
			if err != nil {
				log.Fatalf("failed open location file: %s", err)
			}
			jsonBytes, err := ioutil.ReadAll(jsonFile)
			if err != nil {
				log.Fatalf("failed to read location file data: %s", err)
			}
			err = jsonFile.Close()
			if err != nil {
				log.Fatalf("failed to close file: %s", err)
			}

			var rl rawLocation
			json.Unmarshal([]byte(jsonBytes), &rl)

			existingLocations, err := database.FindLocationsByName(db, rl.Name)
			if err != nil {
				log.Fatalf("failed to look up existing locations: %s", err)
			}

			if len(existingLocations) == 1 {
				locationMap[rl.ID] = existingLocations[0].ID
				continue
			}

			if l := len(existingLocations); l != 0 {
				log.Fatalf("unexpected number of existingLocations %d", l)
			}

			newLocations, err := database.CreateLocations(
				db,
				[]models.Location{
					{Name: rl.Name, Latitude: rl.Lat, Longitude: rl.Long},
				},
			)
			if err != nil {
				log.Fatalf("failed to create location: %s", err)
			}

			locationMap[rl.ID] = newLocations[0].ID
		}

		// import posts
		rawPosts, err := ioutil.ReadDir(basePathPosts)
		if err != nil {
			log.Fatalf("failed to list post source: %s", err)
		}
		for i, postFile := range rawPosts {
			log.Printf("%d/%d", i, len(rawPosts))
			jsonFile, err := os.Open(basePathPosts + postFile.Name())
			if err != nil {
				log.Fatalf("failed open post file: %s", err)
			}
			jsonBytes, err := ioutil.ReadAll(jsonFile)
			if err != nil {
				log.Fatalf("failed to read post file data: %s", err)
			}
			err = jsonFile.Close()
			if err != nil {
				log.Fatalf("failed to close file: %s", err)
			}

			var rp rawPost
			json.Unmarshal([]byte(jsonBytes), &rp)

			postTime := time.Unix(rp.Timestamp, 0)

			var device string
			for _, t := range rp.Tags {
				lower := strings.TrimPrefix(strings.ToLower(t), "#")
				switch lower {
				case "x100f":
					device = "Fujifilm X100F"
				case "teampixel":
					device = "Google Pixel 2"
				case "#sonyrx100iv":
					device = "Sony RX100IV "
				case "#rx100iv":
					device = "Sony RX100IV "
				case "shotoniphone":
					device = "Apple iPhone 11 Pro Max"
				case "shotoniphone11pro":
					device = "Apple iPhone 11 Pro Max"
				case "gopro":
					device = "GoPro HERO9 Black"
				case "samsunggalaxy":
					device = "Samsung Galaxy Note ii"
				}
			}

			gotiPhone6 := time.Date(2014, 10, 1, 12, 0, 0, 0, time.Local)
			gotSony := time.Date(2016, 2, 14, 12, 0, 0, 0, time.Local)
			gotPixel := time.Date(2017, 10, 30, 12, 0, 0, 0, time.Local)

			if device == "" {
				if postTime.Before(gotiPhone6) {
					device = "Samsung Galaxy Note ii"
				} else if postTime.Before(gotSony) {
					device = "Apple iPhone 6"
				} else if postTime.Before(gotPixel) {
					device = "Sony RX100IV"
				}
			}

			if device == "" {
				device = "Unknown"
			}

			existingDevices, err := database.FindDevicesByName(db, device)
			if err != nil {
				log.Fatalf("error finding devices: %s", err)
			}

			if len(existingDevices) == 0 {
				log.Fatalf("failed to find device: %s", device)
			}

			existingMedias, err := database.FindMediasByInstagramCode(db, rp.Code)
			if err != nil {
				log.Fatalf("failed to get existing medias: %s", err)
			}

			var postMediaID int
			if len(existingMedias) == 0 {
				kind := "jpg"
				if rp.IsVideo {
					kind = "mp4"
				}

				newMedias, err := database.CreateMedias(db, []models.Media{
					{
						DeviceID:      existingDevices[0].ID,
						InstagramCode: rp.Code,
						Kind:          kind,
					},
				})
				if err != nil {
					log.Fatalf("failed to get new medias: %s", err)
				}

				mediaURL := fmt.Sprintf("https://storage.googleapis.com/charlieegan3-photos/current/%s.%s", strings.TrimSuffix(postFile.Name(), ".json"), kind)
				resp, err := http.Get(mediaURL)
				if err != nil {
					log.Fatalf("failed to download media: %s", err)
				}

				if resp.StatusCode != http.StatusOK {
					log.Fatalf("failed to download media, not 200 status: %s", mediaURL)
				}

				bw, err := bucket.NewWriter(context.Background(), fmt.Sprintf("media/%d.%s", newMedias[0].ID, kind), nil)
				if err != nil {
					log.Fatalf("failed to open bucket: %s", err)
				}

				_, err = io.Copy(bw, resp.Body)
				if err != nil {
					log.Fatalf("failed to copy to bucket: %s", err)
				}

				err = bw.Close()
				if err != nil {
					log.Fatalf("failed to close bucket: %s", err)
				}

				err = resp.Body.Close()
				if err != nil {
					log.Fatalf("failed to close request body: %s", err)
				}

				postMediaID = newMedias[0].ID
			} else {
				postMediaID = existingMedias[0].ID
			}

			existingPosts, err := database.FindPostsByInstagramCode(db, rp.Code)
			if err != nil {
				log.Fatalf("failed to look up existing posts: %s", err)
			}

			if len(existingPosts) == 1 {
				continue
			}

			if l := len(existingPosts); l != 0 {
				log.Fatalf("unexpected number of existingPosts %d", l)
			}

			locationID, ok := locationMap[rp.Location.ID]
			if !ok {
				log.Printf("failed to find location for post: %s", rp.Code)
				locationID = 9
			}

			createdPosts, err := database.CreatePosts(
				db,
				[]models.Post{{
					MediaID:       postMediaID,
					Description:   regexp.MustCompile(`(?m)( ?(#\w+) ?)+$`).ReplaceAllString(rp.Caption, ""),
					LocationID:    locationID,
					PublishDate:   postTime,
					InstagramCode: rp.Code,
				}},
			)
			if err != nil {
				log.Fatalf("failed to create location: %s", err)
			}

			database.SetPostTags(db, createdPosts[0], computeTags(rp.Caption))
		}
	},
}

func computeTags(str string) []string {
	var re = regexp.MustCompile(`#\w+`)

	tagMap := make(map[string]bool)

	for _, match := range re.FindAllString(str, -1) {
		cleaned := strings.TrimPrefix(strings.ToLower(match), "#")
		tagMap[cleaned] = true
	}

	tags := []string{}
	for k := range tagMap {
		tags = append(tags, k)
	}

	return tags
}

func init() {
	rootCmd.AddCommand(importCmd)
}
