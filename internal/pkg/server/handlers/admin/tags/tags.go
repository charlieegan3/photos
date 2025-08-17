package tags

import (
	"database/sql"
	_ "embed"
	"errors"
	"net/http"
	"strings"

	"github.com/gobuffalo/plush"
	"github.com/gorilla/mux"

	"github.com/charlieegan3/photos/internal/pkg/database"
	"github.com/charlieegan3/photos/internal/pkg/models"
	"github.com/charlieegan3/photos/internal/pkg/server/templating"
)

//go:embed templates/index.html.plush
var indexTemplate string

//go:embed templates/show.html.plush
var showTemplate string

//go:embed templates/new.html.plush
var newTemplate string

func BuildIndexHandler(db *sql.DB, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")

		tags, err := database.AllTags(r.Context(), db, true, database.SelectOptions{SortField: "name"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		ctx := plush.NewContext()
		ctx.Set("tags", tags)

		err = renderer(ctx, indexTemplate, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
	}
}

func BuildGetHandler(db *sql.DB, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")

		name, ok := mux.Vars(r)["tagName"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("tag name is required"))
			return
		}

		tags, err := database.FindTagsByName(r.Context(), db, []string{name})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		if len(tags) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		ctx := plush.NewContext()
		ctx.Set("tag", tags[0])

		err = renderer(ctx, showTemplate, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
	}
}

func BuildNewHandler(renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")

		ctx := plush.NewContext()
		ctx.Set("tag", models.Tag{})

		w.WriteHeader(http.StatusOK)
		err := renderer(ctx, newTemplate, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
	}
}

func BuildCreateHandler(db *sql.DB, _ templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")

		if val, ok := r.Header["Content-Type"]; !ok || val[0] != "application/x-www-form-urlencoded" {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Content-Type must be 'multipart/form-data'"))
			return
		}

		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("failed to parse multipart form"))
			return
		}

		tag := models.Tag{
			Name: strings.TrimSpace(r.Form.Get("Name")),
		}
		if r.Form.Get("Hidden") != "" {
			tag.Hidden = true
		}

		persistedTags, err := database.CreateTags(r.Context(), db, []models.Tag{tag})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		if len(persistedTags) != 1 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("unexpected number of persistedTags"))
			return
		}

		http.Redirect(w, r, "/admin/tags/"+persistedTags[0].Name, http.StatusSeeOther)
	}
}

func BuildFormHandler(db *sql.DB, _ templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")

		contentType, ok := r.Header["Content-Type"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Content-Type must be set"))
			return
		}

		name, ok := mux.Vars(r)["tagName"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("tag slug is required"))
			return
		}

		existingTags, err := database.FindTagsByName(r.Context(), db, []string{name})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		if len(existingTags) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// handle delete
		if contentType[0] != "application/x-www-form-urlencoded" {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Content-Type must be 'application/x-www-form-urlencoded'"))
			return
		}

		err = r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("failed to parse delete form"))
			return
		}

		if r.Form.Get("_method") == http.MethodDelete {
			err = database.DeleteTags(r.Context(), db, []models.Tag{existingTags[0]})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
			http.Redirect(w, r, "/admin/tags", http.StatusSeeOther)
			return
		}

		if r.PostForm.Get("_method") != http.MethodPut {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("expected _method to be PUT or DELETE in form"))
			return
		}

		tag := models.Tag{
			ID:   existingTags[0].ID,
			Name: r.PostForm.Get("Name"),
		}
		if r.Form.Get("Hidden") != "" {
			tag.Hidden = true
		}
		if r.Form.Get("Hidden") == "false" {
			tag.Hidden = false
		}

		updatedTags, err := database.UpdateTags(r.Context(), db, []models.Tag{tag})
		var redirectTo string
		if err != nil {
			var tagConflictErr *database.TagNameConflictError
			if errors.As(err, &tagConflictErr) {
				conflictingTags, err := database.FindTagsByName(r.Context(), db, []string{tag.Name})
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(err.Error()))
					return
				}
				if len(conflictingTags) != 1 {
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte("conflicting tag was not found"))
					return
				}
				err = database.MergeTags(r.Context(), db, conflictingTags[0], tag)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(err.Error()))
					return
				}
				redirectTo = "/admin/tags/" + conflictingTags[0].Name
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
		} else {
			if len(updatedTags) != 1 {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("unexpected number of updatedTags"))
				return
			}

			redirectTo = "/admin/tags/" + updatedTags[0].Name
		}

		// also possible to update from the index
		if referrer := r.Form.Get("RedirectTo"); referrer != "" {
			redirectTo = referrer
		}

		http.Redirect(
			w,
			r,
			redirectTo,
			http.StatusSeeOther,
		)
	}
}
