package shared

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"gocloud.dev/blob"

	"github.com/charlieegan3/photos/internal/pkg/imageproxy"
)

type IconEntity interface {
	GetID() int64
	GetIconPath() string
	GetThumbPath(resizeString string) string
}

type DeviceIcon struct {
	ID       int64
	Slug     string
	IconKind string
}

func (d DeviceIcon) GetID() int64 {
	return d.ID
}

func (d DeviceIcon) GetIconPath() string {
	return fmt.Sprintf("device_icons/%s.%s", d.Slug, d.IconKind)
}

func (d DeviceIcon) GetThumbPath(resizeString string) string {
	return fmt.Sprintf("thumbs/device_icons/%d-%s.%s", d.ID, resizeString, d.IconKind)
}

type LensIcon struct {
	ID int64
}

func (l LensIcon) GetID() int64 {
	return l.ID
}

func (l LensIcon) GetIconPath() string {
	return fmt.Sprintf("lens_icons/%d.png", l.ID)
}

func (l LensIcon) GetThumbPath(resizeString string) string {
	return fmt.Sprintf("thumbs/lens_icons/%d-%s.png", l.ID, resizeString)
}

type EntityFinder[T any] func(context.Context, *sql.DB, []int64) ([]T, error)

func BuildIconHandler[T any](
	db *sql.DB,
	bucket *blob.Bucket,
	paramName string,
	finder EntityFinder[T],
	entityToIcon func(T) IconEntity,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")

		id, err := ParseIDFromPath(r, paramName)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, paramName+" ID is required")
			return
		}

		entities, err := finder(r.Context(), db, []int64{id})
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if len(entities) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if len(entities) != 1 {
			WriteError(w, http.StatusInternalServerError, "unexpected number of entities found")
			return
		}

		w.Header().Set("Cache-Control", "public, max-age=604800")

		icon := entityToIcon(entities[0])

		// Determine content type based on icon kind
		iconPath := icon.GetIconPath()
		var contentType string
		if len(iconPath) >= 3 {
			ext := iconPath[len(iconPath)-3:]
			switch ext {
			case "png":
				contentType = "image/png"
			case "jpg", "peg": // handles both .jpg and .jpeg
				contentType = "image/jpeg"
			default:
				contentType = "image/jpeg" // fallback
			}
		} else {
			contentType = "image/jpeg" // fallback
		}
		w.Header().Set("Content-Type", contentType)
		imageResizeString := r.URL.Query().Get("o")
		originalIconPath := icon.GetIconPath()
		thumbIconPath := icon.GetThumbPath(imageResizeString)

		if imageResizeString == "" {
			attrs, err := bucket.Attributes(r.Context(), originalIconPath)
			if err != nil {
				WriteError(w, http.StatusInternalServerError, err.Error())
				return
			}

			reader, err := bucket.NewReader(r.Context(), originalIconPath, nil)
			if err != nil {
				WriteError(w, http.StatusInternalServerError, err.Error())
				return
			}
			defer reader.Close()

			w.Header().Set("Content-Length", strconv.FormatInt(attrs.Size, 10))
			w.Header().Set("ETag", attrs.ETag)

			_, err = io.Copy(w, reader)
			if err != nil {
				WriteError(w, http.StatusInternalServerError, err.Error())
				return
			}
			return
		}

		ir := imageproxy.Resizer{}
		err = ir.ResizeInBucket(r.Context(), bucket, originalIconPath, imageResizeString, thumbIconPath)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		reader, err := bucket.NewReader(r.Context(), thumbIconPath, nil)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		defer reader.Close()

		_, err = io.Copy(w, reader)
		if err != nil {
			WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
}
