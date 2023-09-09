package public

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"gocloud.dev/blob"

	"github.com/charlieegan3/photos/internal/pkg/database"
	"github.com/charlieegan3/photos/internal/pkg/imageproxy"
)

func BuildMediaHandler(db *sql.DB, bucket *blob.Bucket) func(http.ResponseWriter, *http.Request) {
	ir := imageproxy.Resizer{}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		rawID, ok := mux.Vars(r)["mediaID"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/text")
			w.Write([]byte("media ID is required"))
			return
		}

		id, err := strconv.Atoi(rawID)
		if err != nil {
			w.Header().Set("Content-Type", "application/text")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("media ID was not integer"))
			return
		}

		medias, err := database.FindMediasByID(db, []int{id})
		if err != nil {
			w.Header().Set("Content-Type", "application/text")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(medias) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if len(medias) != 1 {
			w.Header().Set("Content-Type", "application/text")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unexpected number of medias found"))
			return
		}

		w.Header().Set("Cache-Control", "public, max-age=604800")
		w.Header().Set("Content-Type", "image/jpeg")

		originalMediaPath := fmt.Sprintf("media/%d.%s", medias[0].ID, medias[0].Kind)

		imageResizeString := r.URL.Query().Get("o")
		// if there are no options, serve the image from the media upload path
		if imageResizeString == "" {
			serveImageFromBucket(w, r, bucket, originalMediaPath)
			return
		}

		// if the image is an old, '0x0' image with no size information. Replace
		// requests for 'fit' with 'x' to use the old existing thumbs
		if medias[0].Width == 0 || medias[0].Height == 0 {
			imageResizeString = strings.Replace(imageResizeString, ",fit", "x", 1)
		} else {
			imageResizeString = strings.Replace(imageResizeString, ",", "-", 1)
		}

		thumbMediaPath := fmt.Sprintf("thumbs/media/%d-%s.jpg", medias[0].ID, imageResizeString)

		exists, err := bucket.Exists(r.Context(), thumbMediaPath)
		if err != nil {
			w.Header().Set("Content-Type", "application/text")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		if !exists {
			err := ir.ResizeInBucket(
				r.Context(),
				bucket,
				originalMediaPath,
				strings.Replace(imageResizeString, "-", ",", 1),
				thumbMediaPath,
			)
			if err != nil {
				w.Header().Set("Content-Type", "application/text")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
		}

		serveImageFromBucket(w, r, bucket, thumbMediaPath)
	}
}

func serveImageFromBucket(w http.ResponseWriter, r *http.Request, bucket *blob.Bucket, path string) {
	attrs, err := bucket.Attributes(r.Context(), path)
	if err != nil {
		w.Header().Set("Content-Type", "application/text")
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

	br, err := bucket.NewReader(r.Context(), path, nil)
	if err != nil {
		w.Header().Set("Content-Type", "application/text")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	defer br.Close()

	_, err = io.Copy(w, br)
	if err != nil {
		w.Header().Set("Content-Type", "application/text")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("failed to copy image to response: %s", err.Error())))
		return
	}

	err = br.Close()
	if err != nil {
		w.Header().Set("Content-Type", "application/text")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to close bucket reader"))
		return
	}
}
