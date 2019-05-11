// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/spf13/cobra"
)

type siteIndexItem struct {
	ID      string `json:"id"`
	IsVideo bool   `json:"is_video"`
}

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

func (c *Post) FullID() string {
	time := time.Unix(c.Timestamp, 0)
	return time.Format("2006-01-02") + "-" + c.ID
}

var source = ""
var output = ""

func init() {
	command := createDebugCmd()
	siteCmd.AddCommand(command)
}

func createDebugCmd() *cobra.Command {
	debugCmd := cobra.Command{
		Use:   "debug",
		Short: "A brief description of your command",
		Run:   runDebug,
	}

	debugCmd.Flags().StringVar(&source, "source", ".", "Source directory to read from")
	debugCmd.Flags().StringVar(&output, "output", "output", "Source directory to read from")

	return &debugCmd
}

func runDebug(cmd *cobra.Command, args []string) {
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
