package data

import (
	"fmt"
	"log"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/charlieegan3/photos/internal/pkg/git"
	"github.com/spf13/cobra"
)

var source = ""
var output = ""

type fileSystemUpdate struct {
	Path    string
	Content string
}

// CreateSyncCmd initializes the command used by cobra to get profile data
func CreateSyncCmd() *cobra.Command {
	backoffConfig := backoff.NewExponentialBackOff()
	backoffConfig.MaxElapsedTime = 3 * time.Minute

	run := func(cmd *cobra.Command, args []string) {
		err := backoff.Retry(func() error {
			return RunSync(cmd, args)
		}, backoffConfig)

		if err != nil {
			log.Fatalf("failed after backoff: %s", err)
		}
	}

	syncCmd := cobra.Command{
		Use:   "data",
		Short: "Refreshes and saves profile data",
		Run:   run,
	}

	return &syncCmd
}

// RunSync clones or pulls a repo into the path
func RunSync(cmd *cobra.Command, args []string) error {
	log.Println("starting sync of data")

	r, fs, err := git.Clone()
	if err != nil {
		return fmt.Errorf("failed to clone into filesystem: %v", err)
	}

	var updateCount int

	updates, err := lootedUpdates(fs)
	if err != nil {
		return fmt.Errorf("failed to get looted json: %v", err)
	}
	updateCount += len(updates)
	err = git.WriteToPaths(r, fs, updates)
	if err != nil {
		return fmt.Errorf("failed to write new data to git: %v", err)
	}

	updates, err = completedUpdates(fs)
	if err != nil {
		return fmt.Errorf("failed to get completed json: %v", err)
	}
	updateCount += len(updates)
	err = git.WriteToPaths(r, fs, updates)
	if err != nil {
		return fmt.Errorf("failed to write new data to git: %v", err)
	}

	if updateCount > 0 {
		err = git.CommitAndUpdate(r)
		if err != nil {
			return fmt.Errorf("failed update git state: %v", err)
		}
	} else {
		log.Println("skipping sync, there were no updates")
	}
	return nil
}
