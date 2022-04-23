package public

import (
	"bytes"
	"database/sql"
	_ "embed"
	"fmt"
	"github.com/charlieegan3/photos/cms/internal/pkg/server/templating"
	"github.com/gobuffalo/plush"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gocloud.dev/blob"
	"willnorris.com/go/imageproxy"

	"github.com/charlieegan3/photos/cms/internal/pkg/database"
)

//go:embed templates/index.html.plush
var indexTemplate string

//go:embed templates/show.html.plush
var showTemplate string

func BuildIndexHandler(db *sql.DB, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		devices, err := database.AllDevices(db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(devices) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		ctx := plush.NewContext()
		ctx.Set("devices", devices)

		err = renderer(ctx, indexTemplate, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}
}

func BuildShowHandler(db *sql.DB, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		rawID, ok := mux.Vars(r)["deviceID"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("device ID is required"))
			return
		}

		id, err := strconv.ParseInt(rawID, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("device ID was not integer"))
			return
		}

		devices, err := database.FindDevicesByID(db, []int64{id})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(devices) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if len(devices) != 1 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unexpected number of devices found"))
			return
		}

		posts, err := database.DevicePosts(db, devices[0].ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		ctx := plush.NewContext()
		ctx.Set("device", devices[0])
		ctx.Set("posts", posts)

		err = renderer(ctx, showTemplate, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}
}

func BuildIconHandler(db *sql.DB, bucket *blob.Bucket) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		rawID, ok := mux.Vars(r)["deviceID"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("device ID is required"))
			return
		}

		id, err := strconv.ParseInt(rawID, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("device ID was not integer"))
			return
		}

		devices, err := database.FindDevicesByID(db, []int64{id})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(devices) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if len(devices) != 1 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unexpected number of devices found"))
			return
		}

		w.Header().Set("Cache-Control", "public, max-age=604800")
		w.Header().Set("Content-Type", "image/jpeg")

		// TODO validate this matches a basic regex
		imageResizeString := r.URL.Query().Get("o")
		originalIconPath := fmt.Sprintf("device_icons/%s.%s", devices[0].Slug, devices[0].IconKind)
		thumbIconPath := fmt.Sprintf("thumbs/device_icons/%d-%s.%s", devices[0].ID, imageResizeString, devices[0].IconKind)

		// if there are no options, serve the image from the media upload path
		if imageResizeString == "" {
			attrs, err := bucket.Attributes(r.Context(), originalIconPath)
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

			br, err := bucket.NewReader(r.Context(), originalIconPath, nil)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			_, err = io.Copy(w, br)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed to copy map into bucket"))
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

		exists, err := bucket.Exists(r.Context(), thumbIconPath)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if !exists {
			// create a reader to get the full size media from the bucket
			br, err := bucket.NewReader(r.Context(), originalIconPath, nil)
			if err != nil {
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
				w.Write([]byte("failed to copy media item into byte buffer for image processing"))
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
			bw, err := bucket.NewWriter(r.Context(), thumbIconPath, nil)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed to open bucket to stash resized image"))
				return
			}
			_, err = io.Copy(bw, bytes.NewReader(imageBytes))
			if err != nil {
				w.Header().Set("Content-Type", "application/text")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed to copy media item into response"))
				return
			}

			err = bw.Close()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed to close bucket after writing"))
				return
			}

			attrs, err := bucket.Attributes(r.Context(), thumbIconPath)
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
				w.Write([]byte("failed to copy map into bucket"))
				return
			}
			return
		}

		// if there is a thumb, then return the contents in the response
		attrs, err := bucket.Attributes(r.Context(), thumbIconPath)
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

		br, err := bucket.NewReader(r.Context(), thumbIconPath, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		_, err = io.Copy(w, br)
		if err != nil {
			w.Header().Set("Content-Type", "application/text")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to copy thumbnail into response"))
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
