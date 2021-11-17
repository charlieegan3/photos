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

func BuildIndexHandler(db *sql.DB, bucketWebURL string) func(http.ResponseWriter, *http.Request) {
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

		body, err := templating.RenderPage(s, bucketWebURL)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		fmt.Fprintf(w, body)
	}
}

func BuildNewHandler(bucketWebURL string) func(http.ResponseWriter, *http.Request) {
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

		body, err := templating.RenderPage(s, bucketWebURL)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		fmt.Fprintf(w, body)
	}
}

func BuildGetHandler(db *sql.DB, bucketWebURL string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		slug, ok := mux.Vars(r)["deviceSlug"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("device slug is required"))
			return
		}

		devices, err := database.FindDevicesBySlug(db, slug)
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
		ctx.Set("image_url", func(s ...string) string {
			return fmt.Sprintf("%s%s", bucketWebURL, strings.Join(s, ""))
		})

		s, err := plush.Render(showTemplate, ctx)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		body, err := templating.RenderPage(s, bucketWebURL)
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

		device := models.Device{Name: r.Form.Get("Name")}

		f, header, err := r.FormFile("Icon")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("failed to read uploaded icon file"))
			return
		}

		lowerFilename := strings.ToLower(header.Filename)
		if !strings.HasSuffix(lowerFilename, ".jpg") && !strings.HasSuffix(lowerFilename, ".jpeg") {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("icon file must be jpg"))
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

		key := fmt.Sprintf("device_icons/%s.jpg", persistedDevices[0].Slug)

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

		http.Redirect(w, r, fmt.Sprintf("/admin/devices/%s", persistedDevices[0].Slug), http.StatusSeeOther)
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

		slug, ok := mux.Vars(r)["deviceSlug"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("device slug is required"))
			return
		}

		existingDevices, err := database.FindDevicesBySlug(db, slug)
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

			iconKey := fmt.Sprintf("device_icons/%s.jpg", existingDevices[0].Slug)
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
			ID:   existingDevices[0].ID,
			Name: r.PostForm.Get("Name"),
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

		// move the icon, if the slug has changed
		if existingDevices[0].Slug != updatedDevices[0].Slug {
			existingIconKey := fmt.Sprintf("device_icons/%s.jpg", existingDevices[0].Slug)
			br, err := bucket.NewReader(r.Context(), existingIconKey, nil)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed initialize icon storage"))
				return
			}

			bw, err := bucket.NewWriter(r.Context(), fmt.Sprintf("device_icons/%s.jpg", updatedDevices[0].Slug), nil)
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
		f, header, err := r.FormFile("Icon")
		if err == nil {
			lowerFilename := strings.ToLower(header.Filename)
			if !strings.HasSuffix(lowerFilename, ".jpg") && !strings.HasSuffix(lowerFilename, ".jpeg") {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("icon file must be jpg"))
				return
			}

			bw, err := bucket.NewWriter(r.Context(), fmt.Sprintf("device_icons/%s.jpg", updatedDevices[0].Slug), nil)
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
			fmt.Sprintf("/admin/devices/%s", updatedDevices[0].Slug),
			http.StatusSeeOther,
		)
	}
}
