package debug

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/charlieegan3/photos/internal/pkg/types"
	"github.com/charlieegan3/photos/internal/pkg/utils"
	"github.com/spf13/cobra"
)

var source = ""
var output = ""

// CreateDebugCmd initializes the command used by cobra
func CreateDebugCmd() *cobra.Command {
	debugCmd := cobra.Command{
		Use:   "debug",
		Short: "A brief description of your command",
		Run:   RunDebug,
	}

	debugCmd.Flags().StringVar(&source, "source", ".", "Source directory to read from")
	debugCmd.Flags().StringVar(&output, "output", "output", "Source directory to read from")

	return &debugCmd
}

// RunDebug performs the writing of the debug site content
func RunDebug(cmd *cobra.Command, args []string) {
	sourcePath := cmd.Flag("source").Value.String()
	outputPath := cmd.Flag("output").Value.String()
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		os.Mkdir(outputPath, os.ModePerm)
	}

	posts, err := loadPostsFromSource(sourcePath)
	if err != nil {
		log.Fatal(err)
	}

	err = writePosts(outputPath, posts)
	if err != nil {
		log.Fatal(err)
	}

	err = writeIndex(outputPath, posts)
	if err != nil {
		log.Fatal(err)
	}

	err = utils.CopyDir(filepath.Join(sourcePath, "static"), outputPath)
	if err != nil {
		log.Fatal(err)
	}
}

func writeIndex(outputPath string, posts []types.Post) error {
	var siteIndex []struct {
		ID      string `json:"id"`
		IsVideo bool   `json:"is_video"`
	}
	for _, post := range posts {
		item := struct {
			ID      string `json:"id"`
			IsVideo bool   `json:"is_video"`
		}{ID: post.FullID(), IsVideo: post.IsVideo}
		siteIndex = append(siteIndex, item)
	}
	sort.SliceStable(siteIndex, func(i, j int) bool {
		return siteIndex[j].ID < siteIndex[i].ID
	})

	jsonIndex, err := json.Marshal(siteIndex)
	if err != nil {
		return err
	}

	tmpfn := filepath.Join(outputPath, "index.json")
	if err := ioutil.WriteFile(tmpfn, jsonIndex, 0666); err != nil {
		return err
	}

	return nil
}

func writePosts(outputPath string, posts []types.Post) error {
	if _, err := os.Stat(filepath.Join(outputPath, "posts")); os.IsNotExist(err) {
		os.Mkdir(filepath.Join(outputPath, "posts"), os.ModePerm)
	}
	for _, v := range posts {
		jsonPost, err := json.Marshal(v)
		if err != nil {
			return err
		}

		tmpfn := filepath.Join(outputPath, "posts/"+v.FullID()+".json")
		if err := ioutil.WriteFile(tmpfn, jsonPost, 0666); err != nil {
			return err
		}
	}
	return nil
}

func loadPostsFromSource(source string) ([]types.Post, error) {
	var posts []types.Post
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(dir)
	postsPath := filepath.Join(source, "completed_json")
	files, err := ioutil.ReadDir(postsPath)
	if err != nil {
		return posts, err
	}

	for _, f := range files {
		content, err := ioutil.ReadFile(filepath.Join(postsPath, f.Name()))
		if err != nil {
			return posts, err
		}
		var post types.Post
		err = json.Unmarshal(content, &post)
		if err != nil {
			return posts, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}
