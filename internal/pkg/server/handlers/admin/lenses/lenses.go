package lenses

import (
	"database/sql"
	_ "embed"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/gobuffalo/plush"
	"github.com/gorilla/mux"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/memblob"

	"github.com/charlieegan3/photos/internal/pkg/database"
	"github.com/charlieegan3/photos/internal/pkg/models"
	"github.com/charlieegan3/photos/internal/pkg/server/templating"
)

//go:embed templates/index.html.plush
var indexTemplate string

//go:embed templates/new.html.plush
var newTemplate string

//go:embed templates/show.html.plush
var showTemplate string

func BuildIndexHandler(db *sql.DB, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")

		lenses, err := database.AllLenses(r.Context(), db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		ctx := plush.NewContext()
		ctx.Set("lenses", lenses)

		err = renderer(ctx, indexTemplate, w)
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
		ctx.Set("lens", models.Lens{})

		w.WriteHeader(http.StatusOK)
		err := renderer(ctx, newTemplate, w)
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

		rawID, ok := mux.Vars(r)["lensID"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("lens id is required"))
			return
		}

		id, err := strconv.ParseInt(rawID, 10, 0)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("failed to parse lens ID"))
			return
		}

		lenses, err := database.FindLensesByID(r.Context(), db, []int64{id})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		if len(lenses) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		ctx := plush.NewContext()
		ctx.Set("lens", lenses[0])

		err = renderer(ctx, showTemplate, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
	}
}

func BuildCreateHandler(
	db *sql.DB,
	bucket *blob.Bucket,
	_ templating.PageRenderer,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")

		if val, ok := r.Header["Content-Type"]; !ok || !strings.HasPrefix(val[0], "multipart/form-data") {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Content-Type must be 'multipart/form-data'"))
			return
		}

		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("failed to parse multipart form"))
			return
		}

		lens := models.Lens{
			Name:        strings.TrimSpace(r.Form.Get("Name")),
			LensMatches: strings.TrimSpace(r.Form.Get("LensMatches")),
		}

		f, header, err := r.FormFile("Icon")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("failed to read uploaded icon file"))
			return
		}

		lowerFilename := strings.ToLower(header.Filename)
		if parts := strings.Split(lowerFilename, "."); len(parts) > 0 {
			if fe := parts[len(parts)-1]; fe != "png" {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "icon file must be png, got: %s", fe)
				return
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("icon file missing extension"))
			return
		}

		persistedLenses, err := database.CreateLenses(r.Context(), db, []models.Lens{lens})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		if len(persistedLenses) != 1 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("unexpected number of persistedLenses"))
			return
		}

		key := fmt.Sprintf("lens_icons/%d.png", persistedLenses[0].ID)

		bw, err := bucket.NewWriter(r.Context(), key, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "failed to initialize icon storage: %s", err)
			return
		}

		_, err = io.Copy(bw, f)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("failed to save to icon storage"))
			return
		}

		err = bw.Close()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("failed to close connection to icon storage"))
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/admin/lenses/%d", persistedLenses[0].ID), http.StatusSeeOther)
	}
}

func BuildFormHandler(
	db *sql.DB,
	bucket *blob.Bucket,
	_ templating.PageRenderer,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")

		contentType, ok := r.Header["Content-Type"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Content-Type must be set"))
			return
		}

		rawID, ok := mux.Vars(r)["lensID"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("lens id is required"))
			return
		}

		id, err := strconv.ParseInt(rawID, 10, 0)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("failed to parse lens ID"))
			return
		}

		existingLenses, err := database.FindLensesByID(r.Context(), db, []int64{id})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		if len(existingLenses) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// handle delete
		if contentType[0] == "application/x-www-form-urlencoded" {
			err := r.ParseForm()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("failed to parse delete form"))
				return
			}

			if r.Form.Get("_method") != http.MethodDelete {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("expected _method to be DELETE"))
				return
			}

			err = database.DeleteLenses(r.Context(), db, []models.Lens{existingLenses[0]})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}

			iconKey := fmt.Sprintf("lens_icons/%d.png", existingLenses[0].ID)
			err = bucket.Delete(r.Context(), iconKey)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}

			http.Redirect(w, r, "/admin/lenses", http.StatusSeeOther)
			return
		}

		if !strings.HasPrefix(contentType[0], "multipart/form-data") {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Content-Type must be 'multipart/form-data'"))
			return
		}

		err = r.ParseMultipartForm(32 << 20)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("failed to parse multipart form"))
			return
		}

		if r.PostForm.Get("_method") != http.MethodPut {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("expected _method to be PUT or DELETE in form"))
			return
		}

		lens := models.Lens{
			ID:          existingLenses[0].ID,
			Name:        strings.TrimSpace(r.PostForm.Get("Name")),
			LensMatches: strings.TrimSpace(r.PostForm.Get("LensMatches")),
		}

		_, header, err := r.FormFile("Icon")
		if err == nil {
			lowerFilename := strings.ToLower(header.Filename)
			if parts := strings.Split(lowerFilename, "."); len(parts) > 0 {
				if fe := parts[len(parts)-1]; fe != "png" {
					w.WriteHeader(http.StatusBadRequest)
					fmt.Fprintf(w, "icon file must be png, got: %s", fe)
					return
				}
			} else {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("icon file missing extension"))
				return
			}
		}

		updatedLenses, err := database.UpdateLenses(r.Context(), db, []models.Lens{lens})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		if len(updatedLenses) != 1 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("unexpected number of updatedLenses"))
			return
		}

		// only handle the file when it's present, file might not be submitted
		// every time the form is sent
		var f multipart.File
		f, _, err = r.FormFile("Icon")
		if err == nil {
			bw, err := bucket.NewWriter(r.Context(), fmt.Sprintf("lens_icons/%d.png", updatedLenses[0].ID), nil)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("failed initialize icon storage"))
				return
			}

			_, err = io.Copy(bw, f)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("failed to save to icon storage"))
				return
			}

			err = bw.Close()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("failed to close connection to icon storage"))
				return
			}
		}

		http.Redirect(
			w,
			r,
			fmt.Sprintf("/admin/lenses/%d", updatedLenses[0].ID),
			http.StatusSeeOther,
		)
	}
}
