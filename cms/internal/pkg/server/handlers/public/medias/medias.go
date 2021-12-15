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

		medias, err := database.FindMediasByID(db, id)
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

		br, err := bucket.NewReader(r.Context(), fmt.Sprintf("media/%d.%s", medias[0].ID, medias[0].Kind), nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		// TODO handle other media kinds
		w.Header().Set("Content-Type", "image/jpeg")
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
