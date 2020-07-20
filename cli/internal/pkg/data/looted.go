package data

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/charlieegan3/photos/internal/pkg/git"
	"github.com/charlieegan3/photos/internal/pkg/instagram"
	"github.com/go-git/go-billy/v5"
	"github.com/pkg/errors"
)

func lootedUpdates(fs billy.Filesystem) (map[string]string, error) {
	log.Println("starting fetching of looted updates")

	updates := make(map[string]string)

	existing, err := existingIDs(fs, "looted_json")
	if err != nil {
		return updates, errors.Wrap(err, "failed to get looted ids")
	}
	latestPosts, err := instagram.LatestPosts()
	if err != nil {
		return updates, errors.Wrap(err, "failed to get latest posts from instagram")
	}

	for _, v := range latestPosts {
		if !stringArrayContains(existing, v.ID) {
			fmt.Println(v.ID + " is new")
			dateString := time.Unix(v.TakenAtTimestamp, 0).Format("2006-01-02")
			bytes, err := json.MarshalIndent(v, "", "    ")
			if err != nil {
				return updates, errors.Wrap(err, "failed to generate json for post")
			}
			updates["looted_json/"+dateString+"-"+v.ID+".json"] = string(bytes) + "\n"
		}
	}

	log.Println("completed fetching of looted updates")

	return updates, nil
}

func existingIDs(fs billy.Filesystem, dir string) ([]string, error) {
	files, err := git.ListFiles(fs)
	if err != nil {
		return []string{}, errors.Wrap(err, "failed to list files in repo")
	}

	var lootedIDs []string
	var rgx = regexp.MustCompile(`-([^\-]+).json$`)
	for _, v := range files {
		if strings.Contains(v, dir) {
			matches := rgx.FindStringSubmatch(v)
			if len(matches) < 2 {
				return []string{}, errors.Errorf("%v is not a valided looted json path", v)
			}
			lootedIDs = append(lootedIDs, matches[1])
		}
	}
	return lootedIDs, nil
}

func stringArrayContains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
