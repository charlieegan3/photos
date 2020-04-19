package git

import (
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/pkg/errors"
)

// ListFiles will list all files in the git repo
func ListFiles(fs *billy.Filesystem) ([]string, error) {
	// fs.Remove("looted_json/2020-04-09-2283745206165792307.json")

	files, err := recursiveList(*fs, "/")
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
