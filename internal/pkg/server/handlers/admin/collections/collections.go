package collections

import (
	"database/sql"
	_ "embed"
	"fmt"
	"net/http"
	"strconv"

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
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		repo := database.NewCollectionRepository(db)
		collections, err := repo.AllOrderedByPostCount(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		ctx := plush.NewContext()
		ctx.Set("collections", collections)

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
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		rawID, ok := mux.Vars(r)["collectionID"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("collection id is required"))
			return
		}

		id, err := strconv.Atoi(rawID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("failed to parse collection ID"))
			return
		}

		repo := database.NewCollectionRepository(db)
		collections, err := repo.FindByIDs(r.Context(), []int64{int64(id)})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		if len(collections) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		ctx := plush.NewContext()
		ctx.Set("collection", collections[0])

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
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		collection := models.Collection{}

		ctx := plush.NewContext()
		ctx.Set("collection", collection)

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
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		if val, ok := r.Header["Content-Type"]; !ok || val[0] != "application/x-www-form-urlencoded" {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Content-Type must be 'application/x-www-form-urlencoded'"))
			return
		}

		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("failed to parse form"))
			return
		}

		collection := models.Collection{
			Title:       r.Form.Get("Title"),
			Description: r.Form.Get("Description"),
		}

		repo := database.NewCollectionRepository(db)
		persistedCollections, err := repo.Create(r.Context(), []models.Collection{collection})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		if len(persistedCollections) != 1 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("unexpected number of persistedCollections"))
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/admin/collections/%d", persistedCollections[0].ID), http.StatusSeeOther)
	}
}

func BuildFormHandler(db *sql.DB, _ templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		contentType, ok := r.Header["Content-Type"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Content-Type must be set"))
			return
		}

		rawID, ok := mux.Vars(r)["collectionID"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("collection id is required"))
			return
		}

		id, err := strconv.Atoi(rawID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("failed to parse collection ID"))
			return
		}

		repo := database.NewCollectionRepository(db)
		existingCollections, err := repo.FindByIDs(r.Context(), []int64{int64(id)})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		if len(existingCollections) == 0 {
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
			err = repo.Delete(r.Context(), []models.Collection{existingCollections[0]})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}

			http.Redirect(w, r, "/admin/collections", http.StatusSeeOther)
			return
		}

		if r.PostForm.Get("_method") != http.MethodPut {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("expected _method to be PUT or DELETE in form"))
			return
		}

		collection := models.Collection{
			ID:          existingCollections[0].ID,
			Title:       r.Form.Get("Title"),
			Description: r.Form.Get("Description"),
		}

		updatedCollections, err := repo.Update(r.Context(), []models.Collection{collection})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		if len(updatedCollections) != 1 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("unexpected number of updatedCollections"))
			return
		}

		http.Redirect(
			w,
			r,
			fmt.Sprintf("/admin/collections/%d", updatedCollections[0].ID),
			http.StatusSeeOther,
		)
	}
}
