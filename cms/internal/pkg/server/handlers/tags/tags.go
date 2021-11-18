package tags

import (
	"database/sql"
	_ "embed"
	"fmt"
	"net/http"

	"github.com/charlieegan3/photos/cms/internal/pkg/database"
	"github.com/charlieegan3/photos/cms/internal/pkg/models"
	"github.com/charlieegan3/photos/cms/internal/pkg/server/templating"
	"github.com/gobuffalo/plush"
	"github.com/gorilla/mux"
)

//go:embed templates/index.html.plush
var indexTemplate string

//go:embed templates/show.html.plush
var showTemplate string

//go:embed templates/new.html.plush
var newTemplate string

func BuildIndexHandler(db *sql.DB, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		tags, err := database.AllTags(db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		ctx := plush.NewContext()
		ctx.Set("tags", tags)

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

		name, ok := mux.Vars(r)["tagName"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("tag name is required"))
			return
		}

		tags, err := database.FindTagsByName(db, name)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(tags) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		ctx := plush.NewContext()
		ctx.Set("tag", tags[0])

		body, err := renderer(ctx, showTemplate)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		fmt.Fprintf(w, body)
	}
}

func BuildNewHandler(renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		ctx := plush.NewContext()
		ctx.Set("tag", models.Tag{})

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

func BuildCreateHandler(db *sql.DB, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		if val, ok := r.Header["Content-Type"]; !ok || val[0] != "application/x-www-form-urlencoded" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Content-Type must be 'multipart/form-data'"))
			return
		}

		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("failed to parse multipart form"))
			return
		}

		tag := models.Tag{Name: r.Form.Get("Name")}
		if r.Form.Get("Hidden") != "" {
			tag.Hidden = true
		}

		persistedTags, err := database.CreateTags(db, []models.Tag{tag})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(persistedTags) != 1 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unexpected number of persistedTags"))
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/admin/tags/%s", persistedTags[0].Name), http.StatusSeeOther)
	}
}

func BuildFormHandler(db *sql.DB, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		contentType, ok := r.Header["Content-Type"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Content-Type must be set"))
			return
		}

		name, ok := mux.Vars(r)["tagName"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("tag slug is required"))
			return
		}

		existingTags, err := database.FindTagsByName(db, name)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(existingTags) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// handle delete
		if contentType[0] != "application/x-www-form-urlencoded" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Content-Type must be 'application/x-www-form-urlencoded'"))
			return
		}

		err = r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to parse delete form"))
			return
		}

		if r.Form.Get("_method") == "DELETE" {
			err = database.DeleteTags(db, []models.Tag{existingTags[0]})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			http.Redirect(w, r, "/admin/tags", http.StatusSeeOther)
			return
		}

		if r.PostForm.Get("_method") != "PUT" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("expected _method to be PUT or DELETE in form"))
			return
		}

		tag := models.Tag{
			ID:   existingTags[0].ID,
			Name: r.PostForm.Get("Name"),
		}
		if r.Form.Get("Hidden") != "" {
			tag.Hidden = true
		}

		updatedTags, err := database.UpdateTags(db, []models.Tag{tag})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(updatedTags) != 1 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unexpected number of updatedTags"))
			return
		}

		http.Redirect(
			w,
			r,
			fmt.Sprintf("/admin/tags/%s", updatedTags[0].Name),
			http.StatusSeeOther,
		)
	}
}
