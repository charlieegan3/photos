package handlers

import (
	"database/sql"
	_ "embed"
	"fmt"
	"net/http"

	"github.com/gobuffalo/plush"
	"github.com/gorilla/schema"

	"github.com/charlieegan3/photos/cms/internal/pkg/database"
	"github.com/charlieegan3/photos/cms/internal/pkg/models"
)

//go:embed devices/templates/index.html.plush
var indexTemplate string

//go:embed devices/templates/new.html.plush
var newTemplate string

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

		fmt.Fprintf(w, s)
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

		fmt.Fprintf(w, s)
	}
}

func BuildCreateHandler(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

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
