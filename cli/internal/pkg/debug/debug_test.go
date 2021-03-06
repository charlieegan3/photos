package debug

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/charlieegan3/photos/internal/pkg/types"
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
		ID            string  `json:"id"`
		IsVideo       bool    `json:"is_video"`
		LocationCount int     `json:"location_count"`
		Lat           float64 `json:"lat"`
		Long          float64 `json:"long"`
	}
	err = json.Unmarshal(indexContent, &index)
	if err != nil {
		t.Error(err)
	}

	// load an example post from the output
	postContent, err := ioutil.ReadFile(filepath.Join(dir, "posts/2019-04-24-2029394066281649921.json"))
	if err != nil {
		t.Error(err)
	}
	var post types.Post
	err = json.Unmarshal(postContent, &post)
	if err != nil {
		t.Error(err)
	}

	// load an example location from the output
	locationContent, err := ioutil.ReadFile(filepath.Join(dir, "locations/1020946646.json"))
	if err != nil {
		t.Error(err)
	}
	var location types.Location
	err = json.Unmarshal(locationContent, &location)
	if err != nil {
		t.Error(err)
	}

	// load an example tag from the output
	tagContent, err := ioutil.ReadFile(filepath.Join(dir, "tags/sunset.json"))
	if err != nil {
		t.Error(err)
	}
	var tag types.Tag
	err = json.Unmarshal(tagContent, &tag)
	if err != nil {
		t.Error(err)
	}

	locationIndexContent, err := ioutil.ReadFile(filepath.Join(dir, "locations.json"))
	if err != nil {
		t.Error(err)
	}
	var locationsIndex []types.LocationIndexItem
	err = json.Unmarshal(locationIndexContent, &locationsIndex)
	if err != nil {
		t.Error(err)
	}

	tagIndexContent, err := ioutil.ReadFile(filepath.Join(dir, "tags.json"))
	if err != nil {
		t.Error(err)
	}
	var tagsIndex []types.TagIndexItem
	err = json.Unmarshal(tagIndexContent, &tagsIndex)
	if err != nil {
		t.Error(err)
	}

	calendarContent, err := ioutil.ReadFile(filepath.Join(dir, "calendar.json"))
	if err != nil {
		t.Error(err)
	}
	var calendar map[string]int
	err = json.Unmarshal(calendarContent, &calendar)
	if err != nil {
		t.Error(err)
	}

	// assertions
	// check that the size of the index is correct
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
	// assert the order of the index is correct
	for i, v := range index {
		if v.ID != expected[i] {
			t.Errorf("unexpected id, got %v, want: %v", v.ID, expected[i])
		}
	}
	if index[0].LocationCount != 1 {
		t.Errorf("unexpected locationcount, got %v, want: %v", index[0].LocationCount, 1)
	}
	// check that a post has been saved correctly
	if post.ID != "2029394066281649921" {
		t.Errorf("unexpected post ID %v", post.ID)
	}
	// check that a location has been saved too
	if location.ID != "1020946646" {
		t.Errorf("unexpected location ID %v", location.ID)
	}
	// check that the saved location has posts
	if len(location.Posts) != 1 {
		t.Errorf("Location has incorrect number of posts")
	}
	// check the post is the right one
	if location.Posts[0].ID != "1501486602281729368" {
		t.Errorf("Post has wrong ID, %v", location.Posts[0].ID)
	}
	// check that the saved location has nearby locations
	if len(location.Nearby) != 1 {
		t.Errorf("Location has incorrect number of nearby locations")
	}
	// check the post is the right one
	if location.Nearby[0].ID != "1234" {
		t.Errorf("Location has wrong ID, %v", location.Nearby[0].ID)
	}
	// check that the tag has the right name
	if tag.Name != "sunset" {
		t.Errorf("unexpected tag name %v", location.ID)
	}
	// check that the tag has the right posts
	if len(tag.Posts) != 1 {
		t.Errorf("unexpected number of posts %v", len(tag.Posts))
	}
	// check that the tag's posts have lat lon
	if tag.Posts[0].Lat != 123.45 {
		t.Errorf("Incorrect Location Lat for tag post %v", tag.Posts[0].Lat)
	}
	// check that the tag's posts have a location count
	if tag.Posts[0].LocationCount != 1 {
		t.Errorf("Incorrect LocationCount for tag post %v", tag.Posts[0].LocationCount)
	}
	if tag.Posts[0].ID != "2029394066281649921" {
		t.Errorf("unexpected tag post ID %v", tag.Posts[0].ID)
	}
	// check the location index
	if locationsIndex[0].Name != "1 Blossom Street" {
		t.Errorf("unexpected location name in index %v", locationsIndex[0].Name)
	}
	if locationsIndex[0].Count != 1 {
		t.Errorf("unexpected count in index %v", locationsIndex[0].Count)
	}
	if locationsIndex[0].MostRecentPost != "2017-04-26-1501486602281729368" {
		t.Errorf("unexpected post in location index %v", locationsIndex[0].MostRecentPost)
	}
	if locationsIndex[1].Name != "Other Example Location" {
		t.Errorf("unexpected location name in index %v", locationsIndex[1].Name)
	}

	// check the tag index
	if tagsIndex[0].Name != "nofilter" {
		t.Errorf("unexpected tag name in index %v", tagsIndex[0].Name)
	}
	if tagsIndex[0].MostRecentPost != "2018-04-26-1765647895489858959" {
		t.Errorf("unexpected most recent post in index %v", tagsIndex[0].MostRecentPost)
	}
	// check that some tags don't exist
	for _, tag := range tagsIndex {
		if tag.Name == "tagtoexclude" {
			t.Error("excluded tag present in generated index")
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "tagtoexclude.json")); !os.IsNotExist(err) {
		t.Error("excluded tag file was generated")
	}

	// check the calendar date list
	if calendar["2019-04-24"] != 1 {
		t.Error("Incorrect count for date in calendar")
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
