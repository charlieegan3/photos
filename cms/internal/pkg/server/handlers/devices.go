package handlers

import (
	"database/sql"
	_ "embed"
	"fmt"
	"net/http"

	"github.com/gobuffalo/plush"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"

	"github.com/charlieegan3/photos/cms/internal/pkg/database"
	"github.com/charlieegan3/photos/cms/internal/pkg/models"
	"github.com/charlieegan3/photos/cms/internal/pkg/server/templating"
)

//go:embed devices/templates/index.html.plush
var indexTemplate string

//go:embed devices/templates/new.html.plush
var newTemplate string

//go:embed devices/templates/show.html.plush
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

func BuildCreateHandler(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		if val, ok := r.Header["Content-Type"]; !ok || val[0] != "application/x-www-form-urlencoded" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Content-Type must be 'x-www-form-urlencoded'"))
			return
		}

		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		var device models.Device

		err = decoder.Decode(&device, r.PostForm)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		_, err = database.CreateDevices(db, []models.Device{device})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		http.Redirect(w, r, "/admin/devices", http.StatusSeeOther)
	}
}

func BuildFormHandler(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		if val, ok := r.Header["Content-Type"]; !ok || val[0] != "application/x-www-form-urlencoded" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Content-Type must be 'x-www-form-urlencoded'"))
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

		device := devices[0]

		err = r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if r.PostForm.Get("_method") == "DELETE" {
			err = database.DeleteDevices(db, []models.Device{device})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			http.Redirect(w, r, "/admin/devices", http.StatusSeeOther)
			return
		}

		if r.PostForm.Get("_method") != "PUT" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("expected _method to be PUT or DELETE in form"))
			return
		}

		err = decoder.Decode(&device, r.PostForm)
		if err != nil {
			// all keys are processed and then the errors all returned, so a
			// single error of this type is ok for _method
			if _, ok := err.(*schema.UnknownKeyError); ok {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
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
