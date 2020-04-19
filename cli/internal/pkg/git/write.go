package git

import (
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/pkg/errors"
)

// WriteToPaths will write file content to paths in a given repo
func WriteToPaths(r git.Repository, fs billy.Filesystem, updates map[string]string) error {
	log.Printf("started writing of %d updates to repo\n", len(updates))

	wt, err := r.Worktree()
	if err != nil {
		return errors.Wrap(err, "failed load repo worktree")
	}

	for k, v := range updates {
		file, err := fs.Create(k)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to create file: %v", k))
		}

		if _, err = io.Copy(file, strings.NewReader(v)); err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to write contents for file: %s", k))
		}

		_, err = wt.Add(k)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to stage addition for file: %s", k))
		}
	}

	_, err = wt.Commit("add file", &git.CommitOptions{
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

	log.Printf("%d updates written to repo\n", len(updates))

	return nil
}
