package instagram

import (
	"encoding/json"

	"github.com/charlieegan3/photos/internal/pkg/proxy"
	"github.com/charlieegan3/photos/internal/pkg/types"
	"github.com/pkg/errors"
)

// Location returns a formatted location object for the given ID
func Location(id string) (types.Location, error) {
	headers := map[string]string{
		"Cookie": cookie,
	}

	_, body, err := proxy.GetURLViaProxy("https://www.instagram.com/explore/locations/"+id+"/?__a=1", headers)
	if err != nil {
		return types.Location{}, errors.Wrap(err, "failed to get url via proxy")
	}

	var rawLocation types.RawLocation
	err = json.Unmarshal(body, &rawLocation)
	if err != nil {
		return types.Location{}, errors.Wrap(err, "failed to parse response")
	}
	location, err := rawLocation.ToLocation()
	if err != nil {
		return types.Location{}, errors.Wrap(err, "failed to format as completed post")
	}

	return location, nil
}
