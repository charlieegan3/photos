package debug

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func TestRunDebug(t *testing.T) {
	dir, err := ioutil.TempDir(".", "test_output")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	command := CreateDebugCmd()
	command.Flags().Set("output", dir)
	command.Flags().Set("source", "../../../test")

	RunDebug(command, []string{})

	indexContent, err := ioutil.ReadFile(filepath.Join(dir, "index.json"))
	if err != nil {
		t.Error(err)
	}

	var index []struct {
		ID      string
		IsVideo bool
	}
	err = json.Unmarshal(indexContent, &index)
	if err != nil {
		t.Error(err)
	}

	postContent, err := ioutil.ReadFile(filepath.Join(dir, "posts/2019-04-24-2029394066281649921.json"))
	if err != nil {
		t.Error(err)
	}

	var post struct {
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
	err = json.Unmarshal(postContent, &post)
	if err != nil {
		t.Error(err)
	}

	// assertions
	if len(index) != 5 {
		t.Errorf("unexpected number of items %v", len(index))
	}
	expected := []string{
		"2019-04-24-2029394066281649921",
		"2018-04-26-1765647895489858959",
		"2017-04-26-1501486602281729368",
		"2017-03-25-1478214926051861624",
		"2013-09-11-542730817640294126",
	}
	for i, v := range index {
		if v.ID != expected[i] {
			t.Errorf("unexpected id, got %v, want: %v", v.ID, expected[i])
		}
	}
	if post.ID != "2029394066281649921" {
		t.Errorf("unexpected ID %v", post.ID)
	}
}

func TestMissingOutputFolder(t *testing.T) {
	dir, err := ioutil.TempDir(".", "test_output")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	command := CreateDebugCmd()
	command.Flags().Set("output", filepath.Join(dir, "missing_output"))
	command.Flags().Set("source", "../../../test")

	RunDebug(command, []string{})

	if _, err := os.Stat(filepath.Join(dir, "missing_output/index.json")); os.IsNotExist(err) {
		t.Error(err)
	}
}

func TestCopiesStaticFiles(t *testing.T) {
	dir, err := ioutil.TempDir(".", "test_output")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	command := CreateDebugCmd()
	command.Flags().Set("output", dir)
	command.Flags().Set("source", "../../../test")

	RunDebug(command, []string{})

	if _, err := os.Stat(filepath.Join(dir, "index.html")); os.IsNotExist(err) {
		t.Error(err)
	}
}
