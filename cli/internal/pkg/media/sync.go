package media

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/charlieegan3/photos/internal/types"
	"github.com/go-git/go-billy/osfs"
	"github.com/go-git/go-billy/v5"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
)

var bucket string

func init() {
	bucket = os.Getenv("GOOGLE_BUCKET")
	if bucket == "" {
		log.Fatal("GOOGLE_BUCKET must be set")
		os.Exit(1)
	}
	googleJSON := os.Getenv("GOOGLE_JSON")
	if googleJSON == "" {
		log.Fatal("GOOGLE_JSON must be set")
		os.Exit(1)
	}

	content := []byte(googleJSON)
	tmpfile, err := ioutil.TempFile("", "google.*.json")
	if err != nil {
		log.Fatal(err)
	}
	if _, err := tmpfile.Write(content); err != nil {
		tmpfile.Close()
		log.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}

	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", tmpfile.Name())
}

// CreateSyncCmd builds a command to check all image files have been saved to GCS
func CreateSyncCmd() *cobra.Command {
	syncCmd := cobra.Command{
		Use:   "media",
		Short: "Ensures images present in GCS",
		Run:   RunSync,
	}

	return &syncCmd
}

// RunSync downloads the latest data and checks that all repo images are in GCS
func RunSync(cmd *cobra.Command, args []string) {
	// ctx := context.Background()
	// client, err := storage.NewClient(ctx)
	// if err != nil {
	// 	log.Fatalf("failed to init GCS storage client: %s", err)
	// }
	// bkt := client.Bucket(bucket)

	// fs := osfs.New("../")
	//
	// completedPosts, err := listCompletedPosts(fs)
	// if err != nil {
	// 	log.Fatalf("failed to list all completed post jsons: %s", err)
	// }
	//
	// media, err := listAllMedia(ctx, bkt)
	// if err != nil {
	// 	log.Fatalf("failed to list all current images: %s", err)
	// }
	//
	// log.Printf("found data for  %d posts", len(completedPosts))
	// log.Printf("found media for %d posts", len(media))

	missing := []string{
		"2020-07-14-2353286799696139549",
		"2020-07-19-2356320147083470928",
		"2020-07-19-2356572732247851630",
		"2020-07-19-2356650898144376574",
		"2020-07-19-2356726266012516573",
		"2020-07-19-2356932250588919820",
	}

	fs := osfs.New("../")

	for _, v := range missing {
		file, err := fs.Open(fmt.Sprintf("completed_json/%s.json", v))
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		b, err := ioutil.ReadAll(file)

		var completedPost types.CompletedPost
		err = json.Unmarshal(b, &completedPost)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(completedPost)
	}
}

func findMissingMedia(completedPosts, media []string) []string {
	missing := []string{}

	for _, i := range completedPosts {
		present := false
		for _, j := range media {
			if i == j {
				present = true
				break
			}
		}
		if !present {
			missing = append(missing, i)
		}
	}

	return missing
}

func listAllMedia(ctx context.Context, bkt *storage.BucketHandle) ([]string, error) {
	log.Printf("listing bucket: %s", bucket)
	dir := "current/"
	query := &storage.Query{Prefix: dir}

	var names []string
	it := bkt.Objects(ctx, query)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		basename := strings.TrimPrefix(attrs.Name, dir)
		parts := strings.Split(basename, ".")

		if len(parts) != 2 {
			log.Printf("unexpected filename in bucket %s", attrs.Name)
			continue
		}

		names = append(names, parts[0])
	}

	return names, nil
}
func listCompletedPosts(fs billy.Filesystem) ([]string, error) {
	posts := []string{}

	files, err := fs.ReadDir("completed_json")
	if err != nil {
		return posts, errors.Wrap(err, "failed to list completed post jsons")
	}
	for _, v := range files {
		posts = append(posts, strings.TrimSuffix(v.Name(), ".json"))
	}

	return posts, nil
}
