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

		w.Header().Set("Content-Type", "image/jpeg")

		buf := bytes.NewBuffer([]byte{})

		// TODO handle other media kinds
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

		imageResizeString := r.URL.Query().Get("o")
		if imageResizeString == "" {
			imageResizeString = "0x0" // do nothing if no imageOptions set
		}

		imageOptions := imageproxy.ParseOptions(imageResizeString)
		imageOptions.ScaleUp = false // don't attempt to make images larger if not possible

		imageBytes, err := imageproxy.Transform(buf.Bytes(), imageOptions)
		buf = bytes.NewBuffer(imageBytes)

		_, err = io.Copy(w, buf)
		if err != nil {
			w.Header().Set("Content-Type", "application/text")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to copy media item into response"))
			return
		}
	}
}
