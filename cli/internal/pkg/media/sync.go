package media

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/charlieegan3/photos/internal/pkg/git"
	"github.com/charlieegan3/photos/internal/pkg/instagram"
	"github.com/charlieegan3/photos/internal/pkg/types"
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
	log.Println("starting sync of media")

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("failed to init GCS storage client: %s", err)
	}
	bkt := client.Bucket(bucket)

	_, fs, err := git.Clone()
	if err != nil {
		log.Fatalf("failed to clone into filesystem: %v", err)
		os.Exit(1)
	}

	completedPosts, err := listCompletedPosts(fs)
	if err != nil {
		log.Fatalf("failed to list all completed post jsons: %s", err)
	}

	media, err := listAllMedia(ctx, bkt)
	if err != nil {
		log.Fatalf("failed to list all current images: %s", err)
	}

	log.Printf("found data for  %d posts", len(completedPosts))
	log.Printf("found media for %d posts", len(media))

	if len(media) != len(completedPosts) {
		log.Printf("completedPosts - media: %v", arrayDifference(completedPosts, media))
		log.Printf("media - completedPosts: %v", arrayDifference(media, completedPosts))
	}

	missing := findMissingMedia(completedPosts, media)

	for _, imageIdentifier := range missing {
		file, err := fs.Open(fmt.Sprintf("completed_json/%s.json", imageIdentifier))
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

		// refresh the data before fetching
		completedPost, err = instagram.Post(completedPost.Code)
		if err != nil {
			log.Fatal(err)
		}

		// set the object filename
		filename := fmt.Sprintf("%s.jpg", imageIdentifier)
		if completedPost.IsVideo {
			filename = fmt.Sprintf("%s.mp4", imageIdentifier)
		}

		// fetch image
		resp, err := http.Get(completedPost.MediaURL)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		// upload to gcs
		obj := bkt.Object(fmt.Sprintf("current/%s", filename))
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("uploading %s to GCS", filename)

		wc := obj.NewWriter(ctx)
		defer wc.Close()
		_, err = io.Copy(wc, resp.Body)
		if err != nil {
			log.Fatal(err)
		}

		err = wc.Close()
		if err != nil {
			log.Fatal(err)
		}

		// all images are public
		err = obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader)
		if err != nil {
			log.Fatal(err)
		}
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

func arrayDifference(a, b []string) []string {
	target := map[string]bool{}
	for _, x := range b {
		target[x] = true
	}

	result := []string{}
	for _, x := range a {
		if _, ok := target[x]; !ok {
			result = append(result, x)
		}
	}

	return result
}
