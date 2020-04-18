package sync

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/charlieegan3/photos/internal/pkg/git"
	"github.com/charlieegan3/photos/internal/pkg/proxy"
	"github.com/charlieegan3/photos/internal/types"
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
	latest, err := latestIDs()
	if err != nil {
		log.Fatalf("failed to get latest ids: %v", err)
		os.Exit(1)
	}

	for _, v := range latest {
		if !stringArrayContains(existing, v) {
			fmt.Println(v + " is new")
		}
	}
}

func latestIDs() ([]string, error) {
	resp, err := proxy.GetURLViaProxy("https://www.instagram.com/charlieegan3/?__a=1")
	if err != nil {
		return []string{}, errors.Wrap(err, "failed to get url via proxy")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	var profile types.Profile
	if err := json.Unmarshal(body, &profile); err != nil {
		return []string{}, errors.Wrap(err, "failed to parse response")
	}

	var ids []string
	for _, v := range profile.Graphql.User.EdgeOwnerToTimelineMedia.Edges {
		ids = append(ids, v.Node.ID)
	}

	return ids, nil
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
