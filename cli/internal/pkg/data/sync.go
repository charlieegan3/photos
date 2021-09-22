package data

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/charlieegan3/photos/internal/pkg/git"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/spf13/cobra"
)

var source = ""
var output = ""

type fileSystemUpdate struct {
	Path    string
	Content string
}

var syncLocal bool

// CreateSyncCmd initializes the command used by cobra to get profile data
func CreateSyncCmd() *cobra.Command {
	backoffConfig := backoff.NewExponentialBackOff()
	backoffConfig.MaxElapsedTime = 3 * time.Minute

	run := func(cmd *cobra.Command, args []string) {
		var err error
		if syncLocal {
			err = RunSyncLocal(cmd, args)
		} else {
			err = backoff.Retry(func() error {
				err := RunSync(cmd, args)
				if err != nil {
					log.Printf("retrying due to error: %s", err)
				}

				return err
			}, backoffConfig)
		}

		if err != nil {
			log.Fatalf("failed to sync data: %s", err)
		}
	}

	syncCmd := cobra.Command{
		Use:   "data",
		Short: "Refreshes and saves profile data",
		Run:   run,
	}
	syncCmd.Flags().BoolVarP(&syncLocal, "local", "l", false, "if set, only sync using the local dir")

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

// RunSyncLocal syncs media using only the local directory
func RunSyncLocal(cmd *cobra.Command, args []string) error {
	log.Println("starting sync of data")

	var rootPath = ".."

	fs := osfs.New(rootPath)

	updates, err := lootedUpdates(fs)
	if err != nil {
		return fmt.Errorf("failed to get looted json: %v", err)
	}

	if len(updates) > 0 {
		for path, v := range updates {
			ioutil.WriteFile(fmt.Sprintf("%s/%s", rootPath, path), []byte(v), 0644)
		}
		log.Println("there are updates to commit")
	} else {
		log.Println("there are no updates")
	}

	updates, err = completedUpdates(fs)
	if err != nil {
		return fmt.Errorf("failed to get completed json: %v", err)
	}
	if len(updates) > 0 {
		for path, v := range updates {
			ioutil.WriteFile(fmt.Sprintf("%s/%s", rootPath, path), []byte(v), 0644)
		}
		log.Println("there are updates to commit")
	} else {
		log.Println("there are no updates")
	}

	return nil
}
