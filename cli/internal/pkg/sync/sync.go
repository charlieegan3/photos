package sync

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/charlieegan3/photos/internal/pkg/git"
	"github.com/charlieegan3/photos/internal/pkg/instagram"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var source = ""
var output = ""

// CreateSyncCmd initializes the command used by cobra
func CreateSyncCmd() *cobra.Command {
	syncCmd := cobra.Command{
		Use:   "sync",
		Short: "Refreshes and saves profile data",
		Run:   RunSync,
	}

	return &syncCmd
}

// RunSync clones or pulls a repo into the path
func RunSync(cmd *cobra.Command, args []string) {
	existing, err := existingLootedIDs()
	if err != nil {
		log.Fatalf("failed to get existing ids: %v", err)
		os.Exit(1)
	}
	latestPosts, err := instagram.LatestPosts()
	if err != nil {
		log.Fatalf("failed to get latest ids: %v", err)
		os.Exit(1)
	}

	files := make(map[string]string)
	for _, v := range latestPosts {
		if !stringArrayContains(existing, v.ID) {
			fmt.Println(v.ID + " is new")
			dateString := time.Unix(v.TakenAtTimestamp, 0).Format("2006-01-02")
			bytes, err := json.MarshalIndent(v, "", "    ")
			if err != nil {
				log.Fatalf("failed to generate json for post: %v", err)
				os.Exit(1)
			}
			files["looted_json/"+dateString+"-"+v.ID+".json"] = string(bytes) + "\n"
		}
	}

	err = git.WriteToPaths(files)
	if err != nil {
		log.Fatalf("failed to write new data to git: %v", err)
		os.Exit(1)
	}
}

func existingLootedIDs() ([]string, error) {
	files, err := git.ListFiles()
	if err != nil {
		return []string{}, errors.Wrap(err, "failed to list files in repo")
	}

	var lootedIDs []string
	var rgx = regexp.MustCompile(`-([^\-]+).json$`)
	for _, v := range files {
		if strings.Contains(v, "looted_json") {
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
