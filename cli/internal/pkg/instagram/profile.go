package instagram

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/charlieegan3/photos/internal/pkg/proxy"
	"github.com/charlieegan3/photos/internal/types"
	"github.com/pkg/errors"
)

var username string

func init() {
	username = os.Getenv("INSTAGRAM_USERNAME")

	if username == "" {
		log.Fatal("INSTAGRAM_USERNAME must be set")
		os.Exit(1)
	}
}

// LatestPosts returns the latest posts on the profile
func LatestPosts() ([]types.LatestPost, error) {
	var posts []types.LatestPost

	resp, err := proxy.GetURLViaProxy("https://www.instagram.com/" + username + "/?__a=1")
	if err != nil {
		return posts, errors.Wrap(err, "failed to get url via proxy")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	var profile types.Profile
	if err := json.Unmarshal(body, &profile); err != nil {
		return posts, errors.Wrap(err, "failed to parse response")
	}

	for _, v := range profile.Graphql.User.EdgeOwnerToTimelineMedia.Edges {
		posts = append(posts, v.Node)
	}

	return posts, nil
}
