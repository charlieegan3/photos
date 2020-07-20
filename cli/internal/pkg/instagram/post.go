package instagram

import (
	"encoding/json"

	"github.com/charlieegan3/photos/internal/pkg/proxy"
	"github.com/charlieegan3/photos/internal/types"
	"github.com/pkg/errors"
)

// Post returns the latest posts on the profile
func Post(shortcode string) (types.CompletedPost, error) {
	headers := map[string]string{
		"Cookie": cookie,
	}

	_, body, err := proxy.GetURLViaProxy("https://www.instagram.com/p/"+shortcode+"/?__a=1", headers)
	if err != nil {
		return types.CompletedPost{}, errors.Wrap(err, "failed to get url via proxy")
	}

	var post types.RawPost
	if err := json.Unmarshal(body, &post); err != nil {
		return types.CompletedPost{}, errors.Wrap(err, "failed to parse response")
	}

	return post.ToCompletedPost(), nil
}
