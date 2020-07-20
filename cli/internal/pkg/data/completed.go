package data

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/charlieegan3/photos/internal/pkg/git"
	"github.com/charlieegan3/photos/internal/pkg/instagram"
	"github.com/charlieegan3/photos/internal/types"
	"github.com/go-git/go-billy/v5"
	"github.com/pkg/errors"
)

func completedUpdates(fs billy.Filesystem) (map[string]string, error) {
	log.Println("starting fetching of completed updates")

	updates := make(map[string]string)

	lootedIDs, err := existingIDs(fs, "looted_json")
	if err != nil {
		return updates, errors.Wrap(err, "failed to get looted ids")
	}
	completedIDs, err := existingIDs(fs, "completed_json")
	if err != nil {
		return updates, errors.Wrap(err, "failed to get completed ids")
	}

	for _, v := range lootedIDs {
		if !stringArrayContains(completedIDs, v) {
			post, err := loadLootedPostForID(fs, v)
			if err != nil {
				return updates, errors.Wrap(err, fmt.Sprintf("failed to load latest post for ID: %s", v))
			}
			completedPost, err := instagram.Post(post.Shortcode)
			if err != nil {
				return updates, errors.Wrap(err, fmt.Sprintf("failed to get completed post for ID: %s", v))
			}

			dateString := time.Unix(post.TakenAtTimestamp, 0).Format("2006-01-02")
			bytes, err := json.MarshalIndent(completedPost, "", "    ")
			if err != nil {
				return updates, errors.Wrap(err, "failed to generate json for completed post")
			}
			updates["completed_json/"+dateString+"-"+post.ID+".json"] = string(bytes) + "\n"
		}
	}

	log.Println("completed fetching of completed updates")

	return updates, nil
}

func loadLootedPostForID(fs billy.Filesystem, id string) (types.LatestPost, error) {
	var post types.LatestPost

	files, err := git.ListFiles(fs)
	if err != nil {
		return post, errors.Wrap(err, "failed to list files in repo")
	}

	var path string
	for _, v := range files {
		if strings.Contains(v, "looted_json") && strings.Contains(v, id) {
			path = v
			break
		}
	}

	file, err := fs.Open(path)
	if err != nil {
		return post, errors.Wrap(err, fmt.Sprintf("failed to open file: %v", path))
	}
	defer file.Close()

	bytes, _ := ioutil.ReadAll(file)
	err = json.Unmarshal(bytes, &post)
	if err != nil {
		return post, errors.Wrap(err, fmt.Sprintf("failed to parse file: %v", path))
	}

	return post, nil
}
