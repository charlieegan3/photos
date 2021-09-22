package instagram

import (
	"fmt"

	"github.com/Jeffail/gabs/v2"
	"github.com/charlieegan3/photos/internal/pkg/proxy"
	"github.com/charlieegan3/photos/internal/pkg/types"
	"github.com/gosimple/slug"
	"github.com/pkg/errors"
)

// Location returns a formatted location object for the given ID
func Location(id string) (types.Location, error) {
	headers := map[string]string{
		"Cookie": cookie,
	}

	_, body, err := proxy.GetURLViaProxy("https://www.instagram.com/explore/locations/"+id+"/?__a=1", headers)
	if err != nil {
		return types.Location{}, errors.Wrap(err, fmt.Sprintf("failed to get url via proxy: %s", string(body)))
	}

	jsonParsed, err := gabs.ParseJSON(body)
	if err != nil {
		return types.Location{}, errors.Wrap(err, "failed to parse json for location")
	}

	idValue, ok := jsonParsed.Path("native_location_data.location_info.location_id").Data().(string)
	if !ok {
		return types.Location{}, fmt.Errorf("failed to get id value from response")
	}
	nameValue, ok := jsonParsed.Path("native_location_data.location_info.name").Data().(string)
	if !ok {
		return types.Location{}, fmt.Errorf("failed to get name value from response")
	}
	latValue, ok := jsonParsed.Path("native_location_data.location_info.lat").Data().(float64)
	if !ok {
		return types.Location{}, fmt.Errorf("failed to get lat value from response")
	}
	longValue, ok := jsonParsed.Path("native_location_data.location_info.lng").Data().(float64)
	if !ok {
		return types.Location{}, fmt.Errorf("failed to get lng value from response")
	}

	return types.Location{
		ID:   idValue,
		Name: nameValue,
		Slug: slug.Make(fmt.Sprintf("%s-%s", nameValue, idValue)),
		Lat:  latValue,
		Long: longValue,
	}, nil
}
