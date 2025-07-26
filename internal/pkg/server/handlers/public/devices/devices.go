package public

import (
	"database/sql"
	_ "embed"
	"net/http"
	"strconv"

	"github.com/gobuffalo/plush"
	"github.com/gorilla/mux"
	"gocloud.dev/blob"

	"github.com/charlieegan3/photos/internal/pkg/database"
	"github.com/charlieegan3/photos/internal/pkg/models"
	"github.com/charlieegan3/photos/internal/pkg/server/handlers/shared"
	"github.com/charlieegan3/photos/internal/pkg/server/templating"
)

//go:embed templates/index.html.plush
var indexTemplate string

//go:embed templates/show.html.plush
var showTemplate string

func BuildIndexHandler(db *sql.DB, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")

		devices, err := database.AllDevices(r.Context(), db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
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
			_, _ = w.Write([]byte(err.Error()))
			return
		}
	}
}

func BuildShowHandler(db *sql.DB, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")

		rawID, ok := mux.Vars(r)["deviceID"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("device ID is required"))
			return
		}

		id, err := strconv.ParseInt(rawID, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("device ID was not integer"))
			return
		}

		devices, err := database.FindDevicesByID(r.Context(), db, []int64{id})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		if len(devices) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if len(devices) != 1 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("unexpected number of devices found"))
			return
		}

		posts, err := database.DevicePosts(r.Context(), db, devices[0].ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		mediaIDs := []int{}
		for i := range posts {
			mediaIDs = append(mediaIDs, posts[i].MediaID)
		}

		medias, err := database.FindMediasByID(r.Context(), db, mediaIDs)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		mediasByID := make(map[int]models.Media)
		for i := range medias {
			mediasByID[medias[i].ID] = medias[i]
		}

		ctx := plush.NewContext()
		ctx.Set("device", devices[0])
		ctx.Set("posts", posts)
		ctx.Set("medias", mediasByID)

		err = renderer(ctx, showTemplate, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
	}
}

func BuildIconHandler(db *sql.DB, bucket *blob.Bucket) func(http.ResponseWriter, *http.Request) {
	return shared.BuildIconHandler(
		db,
		bucket,
		"deviceID",
		database.FindDevicesByID,
		func(device models.Device) shared.IconEntity {
			return shared.DeviceIcon{
				ID:       device.ID,
				Slug:     device.Slug,
				IconKind: device.IconKind,
			}
		},
	)
}
