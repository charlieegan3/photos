package medias

import (
	"bytes"
	"database/sql"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gobuffalo/plush"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/memblob"

	"github.com/charlieegan3/photos/cms/internal/pkg/database"
	"github.com/charlieegan3/photos/cms/internal/pkg/mediametadata"
	"github.com/charlieegan3/photos/cms/internal/pkg/models"
	"github.com/charlieegan3/photos/cms/internal/pkg/server/templating"
)

//go:embed templates/index.html.plush
var indexTemplate string

//go:embed templates/new.html.plush
var newTemplate string

//go:embed templates/show.html.plush
var showTemplate string

// gorilla decoder can be safely shared and caches data on structs used
var decoder = schema.NewDecoder()

func BuildIndexHandler(db *sql.DB, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		medias, err := database.AllMedias(db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		ctx := plush.NewContext()
		ctx.Set("medias", medias)

		body, err := renderer(ctx, indexTemplate)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		fmt.Fprintf(w, body)
	}
}

func BuildGetHandler(db *sql.DB, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		id, ok := mux.Vars(r)["mediaID"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("media id is required"))
			return
		}

		intID, err := strconv.Atoi(id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to parse supplied ID"))
			return
		}

		medias, err := database.FindMediasByID(db, intID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(medias) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		ctx := plush.NewContext()
		ctx.Set("media", medias[0])

		body, err := renderer(ctx, showTemplate)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		fmt.Fprintf(w, body)
	}
}

func BuildFormHandler(db *sql.DB, bucket *blob.Bucket, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		contentType, ok := r.Header["Content-Type"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Content-Type must be set"))
			return
		}

		id, ok := mux.Vars(r)["mediaID"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("media id is required"))
			return
		}

		intID, err := strconv.Atoi(id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to parse supplied ID"))
			return
		}

		existingMedias, err := database.FindMediasByID(db, intID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(existingMedias) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// handle delete
		if contentType[0] == "application/x-www-form-urlencoded" {
			err := r.ParseForm()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed to parse delete form"))
				return
			}

			if r.Form.Get("_method") != "DELETE" {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("expected _method to be DELETE"))
				return
			}

			err = database.DeleteMedias(db, []models.Media{existingMedias[0]})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			iconKey := fmt.Sprintf("media/%d.jpg", existingMedias[0].ID)
			err = bucket.Delete(r.Context(), iconKey)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			http.Redirect(w, r, "/admin/medias", http.StatusSeeOther)
			return
		}

		if !strings.HasPrefix(contentType[0], "multipart/form-data") {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Content-Type must be 'multipart/form-data'"))
			return
		}

		err = r.ParseMultipartForm(32 << 20)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("failed to parse multipart form"))
			return
		}

		if r.PostForm.Get("_method") != "PUT" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("expected _method to be PUT or DELETE in form"))
			return
		}

		floatKeys := []string{"FNumber", "ShutterSpeed", "Latitude", "Longitude", "Altitude"}
		floatMap := make(map[string]float64)
		for _, key := range floatKeys {
			floatMap[key] = 0
			if val := r.PostForm.Get(key); val != "" {
				f, err := strconv.ParseFloat(val, 64)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(fmt.Sprintf("float value %v was invalid", val)))
					return
				}
				floatMap[key] = f
			}
		}

		var isoSpeed int
		if val := r.PostForm.Get("ISOSpeed"); val != "" {
			isoSpeed, err = strconv.Atoi(val)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("int value %v was invalid", val)))
				return
			}
		}

		var takenAt time.Time
		if val := r.PostForm.Get("TakenAt"); val != "" {
			takenAt, err = time.Parse("2006-01-02T15:04", val)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("time value %v was invalid", val)))
				return
			}
		}

		media := models.Media{
			ID:           existingMedias[0].ID,
			Make:         r.PostForm.Get("Make"),
			Model:        r.PostForm.Get("Model"),
			TakenAt:      takenAt,
			ISOSpeed:     isoSpeed,
			FNumber:      floatMap["FNumber"],
			ShutterSpeed: floatMap["ShutterSpeed"],
			Latitude:     floatMap["Latitude"],
			Longitude:    floatMap["Longitude"],
			Altitude:     floatMap["Altitude"],
		}

		updatedMedias, err := database.UpdateMedias(db, []models.Media{media})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(updatedMedias) != 1 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unexpected number of updatedMedias"))
			return
		}

		// move the icon, if the id has changed
		if existingMedias[0].ID != updatedMedias[0].ID {
			existingFileKey := fmt.Sprintf("media/%d.jpg", existingMedias[0].ID)
			br, err := bucket.NewReader(r.Context(), existingFileKey, nil)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed initialize icon storage"))
				return
			}

			bw, err := bucket.NewWriter(r.Context(), fmt.Sprintf("media/%d.jpg", updatedMedias[0].ID), nil)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed initialize icon storage"))
				return
			}

			_, err = io.Copy(bw, br)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed to save to icon storage"))
				return
			}

			err = bucket.Delete(r.Context(), existingFileKey)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			err = bw.Close()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed to close connection to icon storage"))
				return
			}
			err = br.Close()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed to close connection to icon storage"))
				return
			}
		}

		// only handle the file when it's present, file might not be submitted
		// every time the form is sent
		f, header, err := r.FormFile("File")
		if err == nil {
			lowerFilename := strings.ToLower(header.Filename)
			if !strings.HasSuffix(lowerFilename, ".jpg") && !strings.HasSuffix(lowerFilename, ".jpeg") {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("icon file must be jpg"))
				return
			}

			bw, err := bucket.NewWriter(r.Context(), fmt.Sprintf("media/%d.jpg", updatedMedias[0].ID), nil)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed initialize icon storage"))
				return
			}

			_, err = io.Copy(bw, f)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed to save to icon storage"))
				return
			}

			err = bw.Close()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed to close connection to icon storage"))
				return
			}
		}

		http.Redirect(
			w,
			r,
			fmt.Sprintf("/admin/medias/%d", updatedMedias[0].ID),
			http.StatusSeeOther,
		)
	}
}

func BuildNewHandler(renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		ctx := plush.NewContext()
		ctx.Set("media", models.Media{})

		body, err := renderer(ctx, newTemplate)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, body)
	}
}

func BuildCreateHandler(db *sql.DB, bucket *blob.Bucket, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		if val, ok := r.Header["Content-Type"]; !ok || !strings.HasPrefix(val[0], "multipart/form-data") {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Content-Type must be 'multipart/form-data'"))
			return
		}

		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("failed to parse multipart form"))
			return
		}

		media := models.Media{Make: r.Form.Get("Make")}

		f, header, err := r.FormFile("File")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("failed to open uploaded file: %s", err)))
			return
		}

		lowerFilename := strings.ToLower(header.Filename)
		if !strings.HasSuffix(lowerFilename, ".jpg") && !strings.HasSuffix(lowerFilename, ".jpeg") {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("icon file must be jpg"))
			return
		}

		fileBytes, err := io.ReadAll(f)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("failed to read uploaded file data: %s", err)))
			return
		}

		exifData, err := mediametadata.ExtractMetadata(fileBytes)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("failed to get exif data file: %s", err)))
			return
		}

		media.Make = exifData.Make
		media.Model = exifData.Model
		media.TakenAt = exifData.DateTime
		media.FNumber, err = exifData.FNumber.ToDecimal()
		media.ShutterSpeed, err = exifData.ShutterSpeed.ToDecimal()
		media.ISOSpeed = int(exifData.ISOSpeed)
		media.Latitude, err = exifData.Latitude.ToDecimal()
		media.Longitude, err = exifData.Longitude.ToDecimal()
		media.Altitude, err = exifData.Altitude.ToDecimal()

		persistedMedias, err := database.CreateMedias(db, []models.Media{media})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(persistedMedias) != 1 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unexpected number of persistedMedias"))
			return
		}

		key := fmt.Sprintf("media/%d.jpg", persistedMedias[0].ID)

		bw, err := bucket.NewWriter(r.Context(), key, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed initialize icon storage"))
			return
		}

		_, err = io.Copy(bw, bytes.NewReader(fileBytes))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to save to icon storage"))
			return
		}

		err = bw.Close()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to close connection to icon storage"))
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/admin/medias/%d", persistedMedias[0].ID), http.StatusSeeOther)
	}
}
