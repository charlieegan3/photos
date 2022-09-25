package public

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"github.com/gorilla/mux"
	"gocloud.dev/blob"
	"io"
	"net/http"
	"strconv"
	"sync"
	"willnorris.com/go/imageproxy"

	"github.com/charlieegan3/photos/internal/pkg/database"
)

func BuildMediaHandler(db *sql.DB, bucket *blob.Bucket) func(http.ResponseWriter, *http.Request) {
	ir := imageResizer{}

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

		// TODO validate this matches a basic regex
		imageResizeString := r.URL.Query().Get("o")
		originalMediaPath := fmt.Sprintf("media/%d.%s", medias[0].ID, medias[0].Kind)
		thumbMediaPath := fmt.Sprintf("thumbs/media/%d-%s.jpg", medias[0].ID, imageResizeString)

		// if there are no options, serve the image from the media upload path
		if imageResizeString == "" {
			serveImageFromBucket(w, r, bucket, originalMediaPath)
			return
		}

		exists, err := bucket.Exists(r.Context(), thumbMediaPath)
		if err != nil {
			w.Header().Set("Content-Type", "application/text")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		if !exists {
			err := ir.ResizeInBucket(r.Context(), bucket, originalMediaPath, imageResizeString, thumbMediaPath)
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

type imageResizer struct {
	mu sync.Mutex
}

// ResizeInBucket resizes an image in a bucket and saves it to a new path. One at a time to reduce load on the server.
// When a batch of images are uploaded, there are many that need to be resized. This function is a throttler to ensure
// that only one image is resized at a time and the RAM usage is contained. Original images are quite large and quickly
// consume all available RAM leading to OOMKills.
func (ir *imageResizer) ResizeInBucket(
	ctx context.Context,
	bucket *blob.Bucket,
	originalMediaPath string,
	imageResizeString string,
	thumbMediaPath string,
) error {
	ir.mu.Lock()
	defer ir.mu.Unlock()

	// create a reader to get the full size media from the bucket
	br, err := bucket.NewReader(ctx, originalMediaPath, nil)
	if err != nil {
		return fmt.Errorf("failed to create reader for original media: %w", err)
	}
	defer br.Close()

	// read the full size item
	buf := bytes.NewBuffer([]byte{})
	_, err = io.Copy(buf, br)
	if err != nil {
		fmt.Errorf("failed to copy original media into buffer: %w", err)
	}

	err = br.Close()
	if err != nil {
		return fmt.Errorf("failed to close bucket reader: %w", err)
	}

	// resize the image based on the current settings
	imageOptions := imageproxy.ParseOptions(imageResizeString)
	imageOptions.ScaleUp = false // don't attempt to make images larger if not possible

	imageBytes, err := imageproxy.Transform(buf.Bytes(), imageOptions)
	buf = bytes.NewBuffer(imageBytes)

	// create a writer for the new thumb
	bw, err := bucket.NewWriter(ctx, thumbMediaPath, nil)
	if err != nil {
		return fmt.Errorf("failed to create writer for thumb media: %w", err)
	}
	defer bw.Close()

	_, err = io.Copy(bw, bytes.NewReader(imageBytes))
	if err != nil {
		return fmt.Errorf("failed to copy thumb media into bucket: %w", err)
	}
	err = bw.Close()
	if err != nil {
		return fmt.Errorf("failed to close bucket writer: %w", err)
	}

	return nil
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
		w.Write([]byte("failed to copy media into bucket"))
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
