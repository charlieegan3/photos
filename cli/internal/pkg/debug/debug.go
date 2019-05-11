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

func RunDebug(cmd *cobra.Command, args []string) {
	sourcePath := cmd.Flag("source").Value.String()
	outputPath := cmd.Flag("output").Value.String()

	posts, err := loadPostsFromSource(sourcePath)
	if err != nil {
		log.Fatal(err)
	}
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
		log.Fatal(err)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		os.Mkdir(outputPath, os.ModePerm)
	}

	tmpfn := filepath.Join(outputPath, "index.json")
	if err := ioutil.WriteFile(tmpfn, jsonIndex, 0666); err != nil {
		log.Fatal(err)
	}

	err = utils.CopyDir(filepath.Join(sourcePath, "static"), outputPath)
	if err != nil {
		log.Fatal(err)
	}
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
