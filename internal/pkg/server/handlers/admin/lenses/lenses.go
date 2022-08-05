package lenses

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

		lenses, err := database.AllLenses(db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		ctx := plush.NewContext()
		ctx.Set("lenses", lenses)

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
		ctx.Set("lens", models.Lens{})

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

		rawID, ok := mux.Vars(r)["lensID"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("lens id is required"))
			return
		}

		id, err := strconv.ParseInt(rawID, 10, 0)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to parse lens ID"))
			return
		}

		lenses, err := database.FindLensesByID(db, []int64{id})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
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

		lens := models.Lens{Name: strings.TrimSpace(r.Form.Get("Name"))}

		f, header, err := r.FormFile("Icon")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("failed to read uploaded icon file"))
			return
		}

		lowerFilename := strings.ToLower(header.Filename)
		if parts := strings.Split(lowerFilename, "."); len(parts) > 0 {
			if fe := parts[len(parts)-1]; fe != "png" {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("icon file must be png, got: %s", fe)))
				return
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("icon file missing extension"))
			return
		}

		persistedLenses, err := database.CreateLenses(db, []models.Lens{lens})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(persistedLenses) != 1 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unexpected number of persistedLenses"))
			return
		}

		key := fmt.Sprintf("lens_icons/%d.png", persistedLenses[0].ID)

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

		http.Redirect(w, r, fmt.Sprintf("/admin/lenses/%d", persistedLenses[0].ID), http.StatusSeeOther)
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

		rawID, ok := mux.Vars(r)["lensID"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("lens id is required"))
			return
		}

		id, err := strconv.ParseInt(rawID, 10, 0)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to parse lens ID"))
			return
		}

		existingLenses, err := database.FindLensesByID(db, []int64{id})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
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
				w.Write([]byte("failed to parse delete form"))
				return
			}

			if r.Form.Get("_method") != "DELETE" {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("expected _method to be DELETE"))
				return
			}

			err = database.DeleteLenses(db, []models.Lens{existingLenses[0]})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			iconKey := fmt.Sprintf("lens_icons/%d.png", existingLenses[0].ID)
			err = bucket.Delete(r.Context(), iconKey)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			http.Redirect(w, r, "/admin/lenses", http.StatusSeeOther)
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

		lens := models.Lens{
			ID:   existingLenses[0].ID,
			Name: strings.TrimSpace(r.PostForm.Get("Name")),
		}

		f, header, err := r.FormFile("Icon")
		if err == nil {
			lowerFilename := strings.ToLower(header.Filename)
			if parts := strings.Split(lowerFilename, "."); len(parts) > 0 {
				if fe := parts[len(parts)-1]; fe != "png" {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(fmt.Sprintf("icon file must be png, got: %s", fe)))
					return
				}
			} else {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("icon file missing extension"))
				return
			}
		}

		updatedLenses, err := database.UpdateLenses(db, []models.Lens{lens})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(updatedLenses) != 1 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unexpected number of updatedLenses"))
			return
		}

		// only handle the file when it's present, file might not be submitted
		// every time the form is sent
		f, header, err = r.FormFile("Icon")
		if err == nil {
			bw, err := bucket.NewWriter(r.Context(), fmt.Sprintf("lens_icons/%d.png", updatedLenses[0].ID), nil)
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
			fmt.Sprintf("/admin/lenses/%d", updatedLenses[0].ID),
			http.StatusSeeOther,
		)
	}
}