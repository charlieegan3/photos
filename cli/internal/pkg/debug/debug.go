package debug

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/charlieegan3/photos/internal/pkg/utils"
	"github.com/spf13/cobra"
)

var source = ""
var output = ""

type siteIndexItem struct {
	ID      string `json:"id"`
	IsVideo bool   `json:"is_video"`
}

// Post represents a completed / downloaded post json file
type Post struct {
	Caption    string `json:"caption"`
	Code       string `json:"code"`
	Dimensions struct {
		Height int64 `json:"height"`
		Width  int64 `json:"width"`
	} `json:"dimensions"`
	DisplayURL string `json:"display_url"`
	ID         string `json:"id"`
	IsVideo    bool   `json:"is_video"`
	Location   struct {
		HasPublicPage bool   `json:"has_public_page"`
		ID            string `json:"id"`
		Name          string `json:"name"`
		Slug          string `json:"slug"`
	} `json:"location"`
	MediaURL  string   `json:"media_url"`
	PostURL   string   `json:"post_url"`
	Tags      []string `json:"tags"`
	Timestamp int64    `json:"timestamp"`
}

// FullID computes the full id from the date and instagram ID
func (c *Post) FullID() string {
	time := time.Unix(c.Timestamp, 0)
	return time.Format("2006-01-02") + "-" + c.ID
}

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
	var siteIndex []siteIndexItem
	for _, post := range posts {
		siteIndex = append(siteIndex, siteIndexItem{ID: post.FullID(), IsVideo: post.IsVideo})
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

	err = copyStatic(sourcePath, outputPath)
	if err != nil {
		log.Fatal(err)
	}
}

func loadPostsFromSource(source string) ([]Post, error) {
	var posts []Post
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
		var post Post
		err = json.Unmarshal(content, &post)
		if err != nil {
			return posts, err
		}
		posts = append(posts, post)
	}

	return posts, nil
}

func copyStatic(sourcePath string, outputPath string) error {
	return utils.CopyDir(filepath.Join(sourcePath, "static"), outputPath)
}
