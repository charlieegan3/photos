package instagram

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/charlieegan3/photos/internal/pkg/proxy"
	"github.com/charlieegan3/photos/internal/pkg/types"
	"github.com/pkg/errors"
)

var username, cookie string

func init() {
	username = os.Getenv("INSTAGRAM_USERNAME")

	if username == "" {
		log.Fatal("INSTAGRAM_USERNAME must be set")
		os.Exit(1)
	}

	cookie = os.Getenv("INSTAGRAM_COOKIE_STRING")

	if cookie == "" {
		log.Fatal("INSTAGRAM_COOKIE_STRING must be set (contains session id)")
		os.Exit(1)
	}

	bytes, err := base64.StdEncoding.DecodeString(cookie)
	if err != nil {
		log.Fatal("INSTAGRAM_COOKIE_STRING must be b64")
	}
	cookie = string(bytes)
}

// LatestPosts returns the latest posts on the profile
func LatestPosts() ([]types.LatestPost, error) {
	var posts []types.LatestPost

	headers := map[string]string{
		"Cookie": cookie,
	}

	_, body, err := proxy.GetURLViaProxy("https://www.instagram.com/"+username+"/?__a=1", headers)
	if err != nil {
		return posts, errors.Wrap(err, "failed to get url via proxy")
	}

	fmt.Println(string(body))

	var profile types.Profile
	if err := json.Unmarshal(body, &profile); err != nil {
		return posts, errors.Wrap(err, "failed to parse response")
	}

	for _, v := range profile.Graphql.User.EdgeOwnerToTimelineMedia.Edges {
		posts = append(posts, v.Node)
	}

	return posts, nil
}
