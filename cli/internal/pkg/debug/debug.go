package debug

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charlieegan3/photos/internal/pkg/types"
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
	locations, err := loadLocationsFromSource(sourcePath, posts)
	if err != nil {
		log.Fatal(err)
	}
	tags, err := loadTagsFromPosts(posts)
	if err != nil {
		log.Fatal(err)
	}

	err = writePosts(outputPath, posts)
	if err != nil {
		log.Fatal(err)
	}

	err = writeLocations(outputPath, locations)
	if err != nil {
		log.Fatal(err)
	}

	err = writeTags(outputPath, tags)
	if err != nil {
		log.Fatal(err)
	}

	err = writeIndex(outputPath, posts)
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
		}{ID: post.FullID, IsVideo: post.IsVideo}
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

		tmpfn := filepath.Join(outputPath, "posts/"+v.FullID+".json")
		if err := ioutil.WriteFile(tmpfn, jsonPost, 0666); err != nil {
			return err
		}
	}
	return nil
}

func writeLocations(outputPath string, locations []types.Location) error {
	if _, err := os.Stat(filepath.Join(outputPath, "locations")); os.IsNotExist(err) {
		os.Mkdir(filepath.Join(outputPath, "locations"), os.ModePerm)
	}

	var locationsIndex []types.LocationIndexItem
	for _, v := range locations {
		jsonLocation, err := json.Marshal(v)
		if err != nil {
			return err
		}

		tmpfn := filepath.Join(outputPath, "locations/"+v.ID+".json")
		if err := ioutil.WriteFile(tmpfn, jsonLocation, 0666); err != nil {
			return err
		}

		var mostRecentPost types.Post
		if len(v.Posts) > 0 {
			mostRecentPost = v.Posts[0]
		}

		locationsIndex = append(locationsIndex, types.LocationIndexItem{
			ID:             v.ID,
			Name:           v.Name,
			Lat:            v.Lat,
			Long:           v.Long,
			MostRecentPost: mostRecentPost.FullID,
			Count:          len(v.Posts),
		})
	}

	sort.SliceStable(locationsIndex, func(i, j int) bool {
		return locationsIndex[j].Count < locationsIndex[i].Count
	})

	jsonLocationsIndex, err := json.Marshal(locationsIndex)
	if err != nil {
		return err
	}

	tmpfn := filepath.Join(outputPath, "locations.json")
	if err := ioutil.WriteFile(tmpfn, jsonLocationsIndex, 0666); err != nil {
		return err
	}

	return nil
}

func writeTags(outputPath string, tags []types.Tag) error {
	if _, err := os.Stat(filepath.Join(outputPath, "tags")); os.IsNotExist(err) {
		os.Mkdir(filepath.Join(outputPath, "tags"), os.ModePerm)
	}

	var tagsIndex []types.TagIndexItem
	for _, v := range tags {
		jsonTag, err := json.Marshal(v)
		if err != nil {
			return err
		}

		tmpfn := filepath.Join(outputPath, "tags/"+v.Name+".json")
		if err := ioutil.WriteFile(tmpfn, jsonTag, 0666); err != nil {
			return err
		}

		var mostRecentPost types.Post
		if len(v.Posts) > 0 {
			mostRecentPost = v.Posts[0]
		}

		tagsIndex = append(tagsIndex, types.TagIndexItem{
			Name:           v.Name,
			MostRecentPost: mostRecentPost.FullID,
			Count:          len(v.Posts),
		})
	}

	sort.SliceStable(tagsIndex, func(i, j int) bool {
		return tagsIndex[j].Count < tagsIndex[i].Count
	})

	jsonTagsIndex, err := json.Marshal(tagsIndex)
	if err != nil {
		return err
	}

	tmpfn := filepath.Join(outputPath, "tags.json")
	if err := ioutil.WriteFile(tmpfn, jsonTagsIndex, 0666); err != nil {
		return err
	}
	return nil
}

func loadPostsFromSource(source string) ([]types.Post, error) {
	var posts []types.Post
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
		post.FullID = strings.Split(f.Name(), ".")[0]
		posts = append(posts, post)
	}

	return posts, nil
}

func loadLocationsFromSource(source string, posts []types.Post) ([]types.Location, error) {
	var locations []types.Location
	locationsPath := filepath.Join(source, "locations")
	files, err := ioutil.ReadDir(locationsPath)
	if err != nil {
		return locations, err
	}

	for _, f := range files {
		content, err := ioutil.ReadFile(filepath.Join(locationsPath, f.Name()))
		if err != nil {
			return locations, err
		}
		var location types.Location
		err = json.Unmarshal(content, &location)
		if err != nil {
			return locations, err
		}

		// populate the location with the matching posts
		location.SetPosts(posts)

		locations = append(locations, location)
	}

	for i := range locations {
		locations[i].SetNearby(locations, 5)
	}

	return locations, nil
}

func loadTagsFromPosts(posts []types.Post) ([]types.Tag, error) {
	var tags []types.Tag

	excludedTagsPath := filepath.Join(source, "excluded_tags")
	excludedTagsContent, err := ioutil.ReadFile(excludedTagsPath)
	if err != nil {
		return tags, err
	}
	excludedTags := strings.Split(string(excludedTagsContent), "\n")

	tagMap := make(map[string][]types.Post)
	for _, v := range posts {
		for _, stringTag := range v.Tags {
			tagMap[stringTag] = append(tagMap[stringTag], v)
		}
	}

	for tag, taggedPosts := range tagMap {
		sort.SliceStable(taggedPosts, func(i, j int) bool {
			return taggedPosts[j].FullID < taggedPosts[i].FullID
		})
		found := false
		for _, excluded := range excludedTags {
			if excluded == tag[1:] {
				found = true
				break
			}
		}
		if !found {
			tags = append(tags, types.Tag{
				Name:  tag[1:],
				Posts: taggedPosts,
			})
		}
	}

	sort.SliceStable(tags, func(i, j int) bool {
		return len(tags[j].Posts) < len(tags[i].Posts)
	})

	return tags, nil
}
