package devices

import (
	"database/sql"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gobuffalo/plush"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"

	//"gocloud.dev/blob"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/memblob"

	"github.com/charlieegan3/photos/cms/internal/pkg/database"
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

		devices, err := database.AllDevices(db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
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

func BuildNewHandler(renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		ctx := plush.NewContext()
		ctx.Set("device", models.Device{})

		w.WriteHeader(http.StatusOK)
		err := renderer(ctx, newTemplate, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}
}

func BuildGetHandler(db *sql.DB, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		rawID, ok := mux.Vars(r)["deviceID"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("device id is required"))
			return
		}

		id, err := strconv.Atoi(rawID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to parse device ID"))
			return
		}

		devices, err := database.FindDevicesByID(db, []int{id})
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
		ctx.Set("device", devices[0])

		err = renderer(ctx, showTemplate, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
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

		device := models.Device{Name: strings.TrimSpace(r.Form.Get("Name"))}

		f, header, err := r.FormFile("Icon")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("failed to read uploaded icon file"))
			return
		}

		lowerFilename := strings.ToLower(header.Filename)
		if parts := strings.Split(lowerFilename, "."); len(parts) > 0 {
			device.IconKind = parts[len(parts)-1]
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("icon file missing extension"))
			return
		}
		if device.IconKind != "jpg" && device.IconKind != "jepg" && device.IconKind != "png" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("icon file must be jpg or png, got: %s", device.IconKind)))
			return
		}

		persistedDevices, err := database.CreateDevices(db, []models.Device{device})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(persistedDevices) != 1 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unexpected number of persistedDevices"))
			return
		}

		key := fmt.Sprintf("device_icons/%s.%s", persistedDevices[0].Slug, persistedDevices[0].IconKind)

		bw, err := bucket.NewWriter(r.Context(), key, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("failed to initialize icon storage: %s", err)))
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

		http.Redirect(w, r, fmt.Sprintf("/admin/devices/%d", persistedDevices[0].ID), http.StatusSeeOther)
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

		rawID, ok := mux.Vars(r)["deviceID"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("device id is required"))
			return
		}

		id, err := strconv.Atoi(rawID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to parse device ID"))
			return
		}

		existingDevices, err := database.FindDevicesByID(db, []int{id})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(existingDevices) == 0 {
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

			err = database.DeleteDevices(db, []models.Device{existingDevices[0]})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			iconKey := fmt.Sprintf("device_icons/%s.%s", existingDevices[0].Slug, existingDevices[0].IconKind)
			err = bucket.Delete(r.Context(), iconKey)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			http.Redirect(w, r, "/admin/devices", http.StatusSeeOther)
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

		device := models.Device{
			ID:       existingDevices[0].ID,
			Name:     strings.TrimSpace(r.PostForm.Get("Name")),
			IconKind: existingDevices[0].IconKind,
		}

		f, header, err := r.FormFile("Icon")
		if err == nil {
			lowerFilename := strings.ToLower(header.Filename)
			if parts := strings.Split(lowerFilename, "."); len(parts) > 0 {
				device.IconKind = parts[len(parts)-1]
			} else {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("icon file missing extension"))
				return
			}
			if device.IconKind != "jpg" && device.IconKind != "jepg" && device.IconKind != "png" {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("icon file must be jpg or png, got: %s", device.IconKind)))
				return
			}
		}

		updatedDevices, err := database.UpdateDevices(db, []models.Device{device})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(updatedDevices) != 1 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unexpected number of updatedDevices"))
			return
		}

		// move the icon, if the slug or iconkind has changed
		if existingDevices[0].Slug != updatedDevices[0].Slug || existingDevices[0].IconKind != updatedDevices[0].IconKind {
			existingIconKey := fmt.Sprintf("device_icons/%s.%s", existingDevices[0].Slug, existingDevices[0].IconKind)
			br, err := bucket.NewReader(r.Context(), existingIconKey, nil)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("failed initialize icon storage: %s", err)))
				return
			}

			bw, err := bucket.NewWriter(r.Context(), fmt.Sprintf("device_icons/%s.%s", updatedDevices[0].Slug, updatedDevices[0].IconKind), nil)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(fmt.Sprintf("failed to open new writer for object: %s", err)))
				return
			}

			_, err = io.Copy(bw, br)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed to save to icon storage"))
				return
			}

			err = bucket.Delete(r.Context(), existingIconKey)
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
		f, header, err = r.FormFile("Icon")
		if err == nil {
			bw, err := bucket.NewWriter(r.Context(), fmt.Sprintf("device_icons/%s.%s", updatedDevices[0].Slug, device.IconKind), nil)
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
			fmt.Sprintf("/admin/devices/%d", updatedDevices[0].ID),
			http.StatusSeeOther,
		)
	}
}
