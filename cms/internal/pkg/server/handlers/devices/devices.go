package devices

import (
	"database/sql"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gobuffalo/plush"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/gosimple/slug"

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

func BuildIndexHandler(db *sql.DB) func(http.ResponseWriter, *http.Request) {
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

		s, err := plush.Render(indexTemplate, ctx)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		body, err := templating.RenderPage(s)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		fmt.Fprintf(w, body)
	}
}

func BuildNewHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		ctx := plush.NewContext()
		ctx.Set("device", models.Device{})

		s, err := plush.Render(newTemplate, ctx)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.WriteHeader(http.StatusOK)

		body, err := templating.RenderPage(s)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		fmt.Fprintf(w, body)
	}
}

func BuildGetHandler(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		name, ok := mux.Vars(r)["deviceName"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("device name is required"))
			return
		}

		devices, err := database.FindDevicesByName(db, name)
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

		s, err := plush.Render(showTemplate, ctx)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		body, err := templating.RenderPage(s)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		fmt.Fprintf(w, body)
	}
}

func BuildCreateHandler(db *sql.DB, bucket *blob.Bucket, bucketWebURL string) func(http.ResponseWriter, *http.Request) {
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

		var device models.Device
		err = decoder.Decode(&device, r.PostForm)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		f, _, err := r.FormFile("Icon")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("failed to read uploaded icon file"))
			return
		}

		key := fmt.Sprintf("device_icons/%s.jpg", slug.Make(device.Name))

		bw, err := bucket.NewWriter(r.Context(), key, nil)
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

		// use the supplied base and key to generate a url for use in icon
		// display
		device.IconURL = fmt.Sprintf("%s%s", bucketWebURL, key)

		_, err = database.CreateDevices(db, []models.Device{device})
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/admin/devices/%s", device.Name), http.StatusSeeOther)
	}
}

func BuildFormHandler(db *sql.DB, bucket *blob.Bucket, bucketWebURL string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		contentType, ok := r.Header["Content-Type"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Content-Type must be set"))
			return
		}

		name, ok := mux.Vars(r)["deviceName"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("device name is required"))
			return
		}

		devices, err := database.FindDevicesByName(db, name)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(devices) == 0 {
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

			err = database.DeleteDevices(db, []models.Device{devices[0]})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			err = bucket.Delete(r.Context(), strings.TrimPrefix(devices[0].IconURL, bucketWebURL))
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

		device := models.Device{Name: r.PostForm.Get("Name")}

		f, _, err := r.FormFile("Icon")
		if err == nil {
			key := fmt.Sprintf("device_icons/%s.jpg", slug.Make(device.Name))

			bw, err := bucket.NewWriter(r.Context(), key, nil)
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

			// use the supplied base and key to generate a url for use in icon
			// display
			device.IconURL = fmt.Sprintf("%s%s", bucketWebURL, key)
		}

		device.ID = devices[0].ID
		if device.IconURL == "" {
			device.IconURL = devices[0].IconURL
		}

		if r.PostForm.Get("_method") != "PUT" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("expected _method to be PUT or DELETE in form"))
			return
		}

		_, err = database.UpdateDevices(db, []models.Device{device})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		http.Redirect(
			w,
			r,
			fmt.Sprintf("/admin/devices/%s", device.Name),
			http.StatusSeeOther,
		)
	}
}
