package data

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
		Use:   "data",
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

	var updateCount int

	updates, err := lootedUpdates(fs)
	if err != nil {
		log.Fatalf("failed to get looted json: %v", err)
		os.Exit(1)
	}
	updateCount += len(updates)
	err = git.WriteToPaths(r, fs, updates)
	if err != nil {
		log.Fatalf("failed to write new data to git: %v", err)
		os.Exit(1)
	}

	updates, err = completedUpdates(fs)
	if err != nil {
		log.Fatalf("failed to get completed json: %v", err)
		os.Exit(1)
	}
	updateCount += len(updates)
	err = git.WriteToPaths(r, fs, updates)
	if err != nil {
		log.Fatalf("failed to write new data to git: %v", err)
		os.Exit(1)
	}

	if updateCount > 0 {
		err = git.CommitAndUpdate(r)
		if err != nil {
			log.Fatalf("failed update git state: %v", err)
			os.Exit(1)
		}
	} else {
		log.Println("skipping sync, there were no updates")
	}
}
