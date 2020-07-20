package git

import (
	"fmt"
	"log"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/pkg/errors"
)

// CommitAndUpdate saves the fs state and pushes to the repo
func CommitAndUpdate(r git.Repository) error {
	log.Println("starting update of repo")

	wt, err := r.Worktree()
	if err != nil {
		return errors.Wrap(err, "failed load repo worktree")
	}

	_, err = wt.Add(".")
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to stage addition for files"))
	}

	_, err = wt.Commit("Add files", &git.CommitOptions{
		Author: &object.Signature{Name: "Robot", Email: "robot@charlieegan3.com", When: time.Now()},
	})
	if err != nil {
		return errors.Wrap(err, "failed to write commit")
	}

	err = r.Push(&git.PushOptions{
		Auth: &http.BasicAuth{
			Username: username,
			Password: accessToken,
		},
	})
	if err != nil {
		return errors.Wrap(err, "failed to push changes")
	}

	log.Println("completed update of repo")

	return nil
}
