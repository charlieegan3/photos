package git

import (
	"fmt"
	"io"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	"github.com/pkg/errors"
)

// WriteToPaths will write file content to paths in a given repo
func WriteToPaths(r git.Repository, fs billy.Filesystem, updates map[string]string) error {
	for k, v := range updates {
		file, err := fs.Create(k)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to create file: %v", k))
		}

		if _, err = io.Copy(file, strings.NewReader(v)); err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to write contents for file: %s", k))
		}
	}

	return nil
}
