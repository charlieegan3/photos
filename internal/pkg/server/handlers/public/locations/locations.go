package public

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"gocloud.dev/blob"
	"gocloud.dev/gcerrors"

	"github.com/charlieegan3/photos/cms/internal/pkg/database"
)

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

func BuildMapHandler(db *sql.DB, bucket *blob.Bucket, mapServerURL, mapServerAPIKey string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

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

		locations, err := database.FindLocationsByID(db, id)
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

		br, err := bucket.NewReader(r.Context(), fmt.Sprintf("location_maps/%d.jpg", locations[0].ID), nil)
		if gcerrors.Code(err) == gcerrors.NotFound {
			mapImageURL, err := mapURL(mapServerURL, mapServerAPIKey, locations[0].Latitude, locations[0].Longitude)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			resp, err := http.Get(mapImageURL)
			if resp.StatusCode != http.StatusOK {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("upstream map server request failed: %d", resp.StatusCode)))
				return
			}
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("upstream map server request failed: %s", err)))
				return
			}

			bw, err := bucket.NewWriter(r.Context(), fmt.Sprintf("location_maps/%d.jpg", locations[0].ID), nil)
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

			br, err = bucket.NewReader(r.Context(), fmt.Sprintf("location_maps/%d.jpg", locations[0].ID), nil)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
		}

		w.Header().Set("Content-Type", "image/jpeg")
		_, err = io.Copy(w, br)
		if err != nil {
			w.Header().Set("Content-Type", "application/text")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to copy map into response"))
			return
		}

		err = br.Close()
		if err != nil {
			w.Header().Set("Content-Type", "application/text")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to close map image source"))
			return
		}
	}
}
