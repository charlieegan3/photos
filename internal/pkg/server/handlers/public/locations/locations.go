package public

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/charlieegan3/photos/cms/internal/pkg/server/templating"
	"github.com/gobuffalo/plush"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/charlieegan3/photos/cms/internal/pkg/database"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"gocloud.dev/blob"
)

//go:embed templates/index.html.plush
var indexTemplate string

//go:embed templates/show.html.plush
var showTemplate string

// HeadContent is appended to the head of the base template
const HeadContent = `
<link rel="stylesheet" type="text/css" href="https://unpkg.com/maplibre-gl@1.15.2/dist/maplibre-gl.css">
<script type="text/javascript" src="https://unpkg.com/maplibre-gl@1.15.2/dist/maplibre-gl.js"></script>
`

func mapURL(serverURL, apiKey string, latitude, longitude float64) (string, error) {
	mapURL, err := url.Parse(serverURL)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse mapURL")
	}
	values := url.Values{
		"style":       []string{"osm-bright-smooth"},
		"center":      []string{fmt.Sprintf("lonlat:%f,%f", longitude, latitude)},
		"zoom":        []string{"10.3497"},
		"width":       []string{"400"},
		"height":      []string{"400"},
		"scaleFactor": []string{"2"},
		"marker":      []string{fmt.Sprintf("lonlat:%f,%f;type:awesome;color:#e01401", longitude, latitude)},
		"apiKey":      []string{apiKey},
	}
	mapURL.RawQuery = values.Encode()

	return mapURL.String(), nil
}

func BuildGetHandler(db *sql.DB, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		rawID, ok := mux.Vars(r)["locationID"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("location id is required"))
			return
		}

		id, err := strconv.Atoi(rawID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to parse location ID"))
			return
		}

		locations, err := database.FindLocationsByID(db, []int{id})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(locations) != 1 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		posts, err := database.FindPostsByLocation(db, []int{id})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		ctx := plush.NewContext()
		ctx.Set("location", locations[0])
		ctx.Set("posts", posts)

		err = renderer(ctx, showTemplate, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}
}

func BuildIndexHandler(db *sql.DB, mapServerAPIKey string, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		locations, err := database.AllLocations(db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(locations) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		locationsJSON, err := json.Marshal(locations)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		ctx := plush.NewContext()
		ctx.Set("locations", string(locationsJSON))
		ctx.Set("api_key", mapServerAPIKey)

		err = renderer(ctx, indexTemplate, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}
}

func BuildMapHandler(db *sql.DB, bucket *blob.Bucket, mapServerURL, mapServerAPIKey string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Cache-Control", "public, max-age=604800")

		rawID, ok := mux.Vars(r)["locationID"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("location ID is required"))
			return
		}

		id, err := strconv.Atoi(rawID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("location ID was not integer"))
			return
		}

		locations, err := database.FindLocationsByID(db, []int{id})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(locations) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if len(locations) != 1 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unexpected number of locations found"))
			return
		}

		mapPath := fmt.Sprintf("location_maps/%d.jpg", locations[0].ID)

		exists, err := bucket.Exists(r.Context(), mapPath)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if !exists {
			mapImageURL, err := mapURL(mapServerURL, mapServerAPIKey, locations[0].Latitude, locations[0].Longitude)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			resp, err := http.Get(mapImageURL)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("upstream map server request failed: %s", err)))
				return
			}
			if resp.StatusCode != http.StatusOK {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("upstream map server request failed: %d", resp.StatusCode)))
				return
			}

			bw, err := bucket.NewWriter(r.Context(), mapPath, nil)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed to open bucket to stash map"))
				return
			}

			_, err = io.Copy(bw, resp.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed to copy map into bucket"))
				return
			}

			err = bw.Close()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed to close bucket after writing"))
				return
			}
		}

		br, err := bucket.NewReader(r.Context(), mapPath, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		// defer close here in case we have a 304 response
		defer br.Close()

		attrs, err := bucket.Attributes(r.Context(), mapPath)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Header().Set("ETag", attrs.ETag)

		// handle potential 304 response
		if ifNoneMatch := r.Header.Get("If-None-Match"); ifNoneMatch != "" {
			if ifNoneMatch == attrs.ETag {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}

		_, err = io.Copy(w, br)
		if err != nil {
			w.Header().Set("Content-Type", "application/text")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to copy map into response"))
			return
		}
	}
}
