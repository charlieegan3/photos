package sync

import (
	"log"
	"os"

	"github.com/charlieegan3/photos/internal/pkg/git"
	"github.com/spf13/cobra"
)

var source = ""
var output = ""

type fileSystemUpdate struct {
	Path    string
	Content string
}

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
	r, fs, err := git.Clone()
	if err != nil {
		log.Fatalf("failed to clone into filesystem: %v", err)
		os.Exit(1)
	}

	var updates []fileSystemUpdate

	lUpdates, err := lootedUpdates(&fs)
	if err != nil {
		log.Fatalf("failed to get looted json: %v", err)
		os.Exit(1)
	}
	updates = append(updates, lUpdates...)

	updatesMap := make(map[string]string)
	for _, v := range updates {
		updatesMap[v.Path] = v.Content
	}

	if len(updates) > 0 {
		err = git.WriteToPaths(r, fs, updatesMap)
		if err != nil {
			log.Fatalf("failed to write new data to git: %v", err)
			os.Exit(1)
		}
	} else {
		log.Println("skipping sync, there were no updates")
	}
}
