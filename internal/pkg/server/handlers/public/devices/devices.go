package public

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gocloud.dev/blob"

	"github.com/charlieegan3/photos/cms/internal/pkg/database"
)

func BuildIconHandler(db *sql.DB, bucket *blob.Bucket) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		rawID, ok := mux.Vars(r)["deviceID"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("device ID is required"))
			return
		}

		id, err := strconv.Atoi(rawID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("device ID was not integer"))
			return
		}

		devices, err := database.FindDevicesByID(db, id)
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

		br, err := bucket.NewReader(r.Context(), fmt.Sprintf("device_icons/%s.%s", devices[0].Slug, devices[0].IconKind), nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Header().Set("Content-Type", "image/jpeg")
		if devices[0].IconKind == "png" {
			w.Header().Set("Content-Type", "image/png")
		}
		_, err = io.Copy(w, br)
		if err != nil {
			w.Header().Set("Content-Type", "application/text")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to copy media item into response"))
			return
		}

		err = br.Close()
		if err != nil {
			w.Header().Set("Content-Type", "application/text")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to close handle loading image from backing store"))
			return
		}
	}
}
