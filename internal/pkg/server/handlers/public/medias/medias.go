package public

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gocloud.dev/blob"
	"willnorris.com/go/imageproxy"

	"github.com/charlieegan3/photos/internal/pkg/database"
)

func BuildMediaHandler(db *sql.DB, bucket *blob.Bucket) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		rawID, ok := mux.Vars(r)["mediaID"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("media ID is required"))
			return
		}

		id, err := strconv.Atoi(rawID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("media ID was not integer"))
			return
		}

		medias, err := database.FindMediasByID(db, []int{id})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(medias) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if len(medias) != 1 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unexpected number of medias found"))
			return
		}

		w.Header().Set("Cache-Control", "public, max-age=604800")
		w.Header().Set("Content-Type", "image/jpeg")

		// TODO validate this matches a basic regex
		imageResizeString := r.URL.Query().Get("o")
		originalMediaPath := fmt.Sprintf("media/%d.%s", medias[0].ID, medias[0].Kind)
		thumbMediaPath := fmt.Sprintf("thumbs/media/%d-%s.jpg", medias[0].ID, imageResizeString)

		// if there are no options, serve the image from the media upload path
		if imageResizeString == "" {
			attrs, err := bucket.Attributes(r.Context(), originalMediaPath)
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

			br, err := bucket.NewReader(r.Context(), originalMediaPath, nil)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			_, err = io.Copy(w, br)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed to copy media into bucket"))
				return
			}

			err = br.Close()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed to close bucket reader"))
				return
			}

			return
		}

		exists, err := bucket.Exists(r.Context(), thumbMediaPath)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		if !exists {
			// create a reader to get the full size media from the bucket
			br, err := bucket.NewReader(r.Context(), originalMediaPath, nil)
			defer br.Close()
			if err != nil {
				br.Close()
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			// read the full size item
			buf := bytes.NewBuffer([]byte{})
			_, err = io.Copy(buf, br)
			if err != nil {
				w.Header().Set("Content-Type", "application/text")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("failed to copy media item into byte buffer for image processing: %s", err)))
				return
			}

			err = br.Close()
			if err != nil {
				w.Header().Set("Content-Type", "application/text")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed to close handle loading image from backing store"))
				return
			}

			// resize the image based on the current settings
			imageOptions := imageproxy.ParseOptions(imageResizeString)
			imageOptions.ScaleUp = false // don't attempt to make images larger if not possible

			imageBytes, err := imageproxy.Transform(buf.Bytes(), imageOptions)
			buf = bytes.NewBuffer(imageBytes)

			// create a writer for the new thumb
			bw, err := bucket.NewWriter(r.Context(), thumbMediaPath, nil)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed to open bucket to stash resized image"))
				return
			}
			_, err = io.Copy(bw, bytes.NewReader(imageBytes))
			if err != nil {
				w.Header().Set("Content-Type", "application/text")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("failed to copy media image into response: %s", err)))
				return
			}

			err = bw.Close()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed to close bucket after writing"))
				return
			}

			attrs, err := bucket.Attributes(r.Context(), thumbMediaPath)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			w.Header().Set("ETag", attrs.ETag)

			// return the resized image in the response
			_, err = io.Copy(w, bytes.NewReader(imageBytes))
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed to copy media into bucket"))
				return
			}
			return
		}

		// if there is a thumb, then return the contents in the response
		attrs, err := bucket.Attributes(r.Context(), thumbMediaPath)
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

		br, err := bucket.NewReader(r.Context(), thumbMediaPath, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		_, err = io.Copy(w, br)
		if err != nil {
			w.Header().Set("Content-Type", "application/text")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("failed to copy media image into response: %s", err)))
			return
		}

		err = br.Close()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to close bucket after reading"))
			return
		}
	}
}
