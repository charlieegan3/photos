package git

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/pkg/errors"
)

// Clone downloads a copy of the repo for processing
func Clone() (git.Repository, billy.Filesystem, error) {
	fs := memfs.New()

	r, err := git.Clone(memory.NewStorage(), fs, &git.CloneOptions{
		URL:   repoURL,
		Depth: 1,
		Auth: &http.BasicAuth{
			Username: username,
			Password: accessToken,
		},
	})

	if err != nil {
		return *r, fs, errors.Wrap(err, "failed to clone repo")
	}

	return *r, fs, nil
}
