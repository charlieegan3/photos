package locations

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/charlieegan3/photos/internal/pkg/git"
	"github.com/charlieegan3/photos/internal/pkg/instagram"
	"github.com/charlieegan3/photos/internal/pkg/types"
	"github.com/go-git/go-billy/v5"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
}

// CreateSyncCmd builds a command to check all image files have been saved to GCS
func CreateSyncCmd() *cobra.Command {
	syncCmd := cobra.Command{
		Use:   "locations",
		Short: "Ensures locations are present in repo",
		Run:   RunSync,
	}

	return &syncCmd
}

// RunSync downloads the latest data and checks that all repo images are in GCS
func RunSync(cmd *cobra.Command, args []string) {
	log.Println("starting sync of locations")

	r, fs, err := git.Clone()
	if err != nil {
		log.Fatalf("failed to clone into filesystem: %v", err)
		os.Exit(1)
	}

	updates, err := locationUpdates(fs)
	if err != nil {
		log.Fatalf("failed to get updated locations: %s", err)
	}

	err = git.WriteToPaths(r, fs, updates)
	if err != nil {
		log.Fatalf("failed to write new data to git: %v", err)
		os.Exit(1)
	}

	if len(updates) > 0 {
		err = git.CommitAndUpdate(r)
		if err != nil {
			log.Fatalf("failed update git state: %v", err)
			os.Exit(1)
		}
	} else {
		log.Println("skipping sync, there were no updates")
	}
}

func locationUpdates(fs billy.Filesystem) (map[string]string, error) {
	updates := make(map[string]string)

	usedLocations, err := getAllUsedLocations(fs)
	if err != nil {
		log.Fatalf("failed to get locations used in post jsons: %s", err)
	}
	savedLocations, err := listSavedLocationIDs(fs)
	if err != nil {
		log.Fatalf("failed to get locations currently saved: %s", err)
	}

	for _, usedLocation := range usedLocations {
		missing := true
		for _, savedLocation := range savedLocations {
			if usedLocation == savedLocation {
				missing = false
				break
			}

		}
		if missing {
			location, err := instagram.Location(usedLocation)
			if err != nil {
				log.Fatalf("failed to get location %s: %s", usedLocation, err)
			}
			bytes, err := json.MarshalIndent(location, "", "    ")
			if err != nil {
				return updates, errors.Wrap(err, "failed to generate json for location")
			}
			updates["locations/"+location.ID+".json"] = string(bytes) + "\n"
		}
	}

	return updates, nil
}

func getAllUsedLocations(fs billy.Filesystem) ([]string, error) {
	completedPosts, err := getCompletedPosts(fs)
	if err != nil {
		return []string{}, errors.Wrap(err, "failed to load completed posts from file system: ")
	}

	var locations []string
	for _, completedPost := range completedPosts {
		if completedPost.Location.ID == "" {
			log.Printf("image missing location: %s", completedPost.Code)
		} else {
			locations = append(locations, completedPost.Location.ID)
		}
	}

	return arrayUnique(locations), nil
}

func listSavedLocationIDs(fs billy.Filesystem) ([]string, error) {
	ids := []string{}

	files, err := fs.ReadDir("locations")
	if err != nil {
		return ids, errors.Wrap(err, "failed to list location jsons")
	}
	for _, v := range files {
		ids = append(ids, strings.TrimSuffix(v.Name(), ".json"))
	}

	return ids, nil
}

func getCompletedPosts(fs billy.Filesystem) ([]types.CompletedPost, error) {
	posts := []types.CompletedPost{}

	files, err := fs.ReadDir("completed_json")
	if err != nil {
		return posts, errors.Wrap(err, "failed to list completed post jsons")
	}
	for _, v := range files {
		var post types.CompletedPost
		file, err := fs.Open(fmt.Sprintf("completed_json/%s", v.Name()))
		if err != nil {
			return posts, errors.Wrap(err, "failed to read completedPost file")
		}

		jsonBlob, err := ioutil.ReadAll(file)
		file.Close()
		err = json.Unmarshal(jsonBlob, &post)
		if err != nil {
			return posts, errors.Wrap(err, "failed to unmarshal json data in completed post")
		}
		posts = append(posts, post)

	}

	return posts, nil
}

func arrayUnique(strs []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range strs {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
