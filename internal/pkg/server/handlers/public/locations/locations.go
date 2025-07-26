package public

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gobuffalo/plush"

	"github.com/charlieegan3/photos/internal/pkg/models"
	"github.com/charlieegan3/photos/internal/pkg/server/templating"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"gocloud.dev/blob"

	"github.com/charlieegan3/photos/internal/pkg/database"
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

		locations, err := database.FindLocationsByID(r.Context(), db, []int{id})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(locations) != 1 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		posts, err := database.FindPostsByLocation(r.Context(), db, []int{id})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		var mediaIDs []int
		mediasByID := make(map[int]models.Media)
		if len(posts) == 0 {
			for i := range posts {
				mediaIDs = append(mediaIDs, posts[i].MediaID)
			}

			medias, err := database.FindMediasByID(r.Context(), db, mediaIDs)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			for i := range medias {
				mediasByID[medias[i].ID] = medias[i]
			}
		}

		ctx := plush.NewContext()
		ctx.Set("location", locations[0])
		ctx.Set("posts", posts)
		ctx.Set("medias", mediasByID)

		err = renderer(ctx, showTemplate, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}
}

func BuildIndexHandler(
	db *sql.DB,
	mapServerAPIKey string,
	renderer templating.PageRenderer,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		locations, err := database.AllLocations(r.Context(), db)
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

func BuildMapHandler(
	db *sql.DB,
	bucket *blob.Bucket,
	mapServerURL, mapServerAPIKey string,
) func(http.ResponseWriter, *http.Request) {
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

		locations, err := database.FindLocationsByID(r.Context(), db, []int{id})
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

			client := &http.Client{}

			req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, mapImageURL, nil)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "failed to create request to upstream map server: %s", err)
				return
			}

			resp, err := client.Do(req)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "upstream map server request failed: %s", err)
				return
			}
			if resp.StatusCode != http.StatusOK {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "upstream map server request failed: %d", resp.StatusCode)
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
			fmt.Fprintf(w, "failed to copy map image into response: %s", err)
			return
		}
	}
}
