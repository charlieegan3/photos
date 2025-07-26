package public

import (
	"bytes"
	"database/sql"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gobuffalo/plush"
	"github.com/gorilla/mux"
	"gocloud.dev/blob"
	"willnorris.com/go/imageproxy"

	"github.com/charlieegan3/photos/internal/pkg/database"
	"github.com/charlieegan3/photos/internal/pkg/models"
	"github.com/charlieegan3/photos/internal/pkg/server/templating"
)

//go:embed templates/index.html.plush
var indexTemplate string

//go:embed templates/show.html.plush
var showTemplate string

func BuildIndexHandler(db *sql.DB, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		lenses, err := database.AllLenses(r.Context(), db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(lenses) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		ctx := plush.NewContext()
		ctx.Set("lenses", lenses)

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

		rawID, ok := mux.Vars(r)["lensID"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("lens ID is required"))
			return
		}

		id, err := strconv.ParseInt(rawID, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("lens ID was not integer"))
			return
		}

		lenses, err := database.FindLensesByID(r.Context(), db, []int64{id})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(lenses) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if len(lenses) != 1 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unexpected number of lenses found"))
			return
		}

		posts, err := database.LensPosts(r.Context(), db, lenses[0].ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		mediaIDs := []int{}
		for i := range posts {
			mediaIDs = append(mediaIDs, posts[i].MediaID)
		}

		medias, err := database.FindMediasByID(r.Context(), db, mediaIDs)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		mediasByID := make(map[int]models.Media)
		for i := range medias {
			mediasByID[medias[i].ID] = medias[i]
		}

		ctx := plush.NewContext()
		ctx.Set("lens", lenses[0])
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

func BuildIconHandler(db *sql.DB, bucket *blob.Bucket) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		rawID, ok := mux.Vars(r)["lensID"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("lens ID is required"))
			return
		}

		id, err := strconv.ParseInt(rawID, 10, 0)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("lens ID was not integer"))
			return
		}

		lenses, err := database.FindLensesByID(r.Context(), db, []int64{id})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(lenses) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if len(lenses) != 1 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unexpected number of lenses found"))
			return
		}

		w.Header().Set("Cache-Control", "public, max-age=604800")
		w.Header().Set("Content-Type", "image/jpeg")

		// TODO validate this matches a basic regex
		imageResizeString := r.URL.Query().Get("o")
		originalIconPath := fmt.Sprintf("lens_icons/%d.png", lenses[0].ID)
		thumbIconPath := fmt.Sprintf("thumbs/lens_icons/%d-%s.png", lenses[0].ID, imageResizeString)

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
			defer br.Close()

			_, err = io.Copy(w, br)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed to copy lens icon into bucket"))
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
			defer br.Close()

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
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed to transform image"))
				return
			}

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
				fmt.Fprintf(w, "failed to copy lens image into response: %s", err)
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
				w.Write([]byte("failed to copy lens icon into bucket"))
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
		defer br.Close()

		_, err = io.Copy(w, br)
		if err != nil {
			w.Header().Set("Content-Type", "application/text")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "failed to copy lens image into response: %s", err)
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
