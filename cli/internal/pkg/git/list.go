package git

import (
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/pkg/errors"
)

// ListFiles will list all files in the git repo
func ListFiles() ([]string, error) {
	fs := memfs.New()

	// clone repo with credentials
	_, err := git.Clone(memory.NewStorage(), fs, &git.CloneOptions{
		URL:   repoURL,
		Depth: 1,
		Auth: &http.BasicAuth{
			Username: username,
			Password: accessToken,
		},
	})
	if err != nil {
		return []string{}, errors.Wrap(err, "failed to clone repo")
	}

	files, err := recursiveList(fs, "/")
	if err != nil {
		return []string{}, errors.Wrap(err, "failed to recursively list files")
	}

	return files, nil
}

func recursiveList(fs billy.Filesystem, path string) ([]string, error) {
	infos, err := fs.ReadDir(path)
	if err != nil {
		return []string{}, errors.Wrap(err, "failed read dir")
	}

	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	var files []string
	for _, v := range infos {
		if v.IsDir() {
			dirFiles, err := recursiveList(fs, path+v.Name())
			if err != nil {
				return []string{}, errors.Wrap(err, "failed read dir")
			}
			files = append(files, dirFiles...)
		} else {
			files = append(files, path+v.Name())
		}
	}

	return files, nil
}
