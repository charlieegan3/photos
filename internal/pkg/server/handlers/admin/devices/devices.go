package devices

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"io"
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
	"github.com/charlieegan3/photos/internal/pkg/server/handlers/shared"
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

		deviceRepo := database.NewDeviceRepository(db)
		devices, err := deviceRepo.All(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		ctx := plush.NewContext()
		ctx.Set("devices", devices)

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
		ctx.Set("device", models.Device{})

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

		rawID, ok := mux.Vars(r)["deviceID"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("device id is required"))
			return
		}

		id, err := strconv.ParseInt(rawID, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("failed to parse device ID"))
			return
		}

		deviceRepo := database.NewDeviceRepository(db)
		devices, err := deviceRepo.FindByIDs(r.Context(), []int64{id})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
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

		device := models.Device{
			Name:         strings.TrimSpace(r.Form.Get("Name")),
			ModelMatches: strings.TrimSpace(r.PostForm.Get("ModelMatches")),
		}

		f, header, err := r.FormFile("Icon")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("failed to read uploaded icon file"))
			return
		}

		lowerFilename := strings.ToLower(header.Filename)
		if parts := strings.Split(lowerFilename, "."); len(parts) > 0 {
			device.IconKind = parts[len(parts)-1]
		} else {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("icon file missing extension"))
			return
		}
		if device.IconKind != "jpg" && device.IconKind != "jpeg" && device.IconKind != "png" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "icon file must be jpg or png, got: %s", device.IconKind)
			return
		}

		deviceRepo := database.NewDeviceRepository(db)
		persistedDevices, err := deviceRepo.Create(r.Context(), []models.Device{device})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		if len(persistedDevices) != 1 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("unexpected number of persistedDevices"))
			return
		}

		key := fmt.Sprintf("device_icons/%s.%s", persistedDevices[0].Slug, persistedDevices[0].IconKind)

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

		http.Redirect(w, r, fmt.Sprintf("/admin/devices/%d", persistedDevices[0].ID), http.StatusSeeOther)
	}
}

func BuildDeleteHandler(
	db *sql.DB,
	bucket *blob.Bucket,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")

		err := shared.ValidateContentType(r, "application/x-www-form-urlencoded")
		if err != nil {
			shared.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		id, err := shared.ParseIDFromPath(r, "deviceID")
		if err != nil {
			shared.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		deviceRepo := database.NewDeviceRepository(db)
		existingDevices, err := deviceRepo.FindByIDs(r.Context(), []int64{id})
		if err != nil {
			shared.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if len(existingDevices) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		err = r.ParseForm()
		if err != nil {
			shared.WriteError(w, http.StatusInternalServerError, "failed to parse delete form")
			return
		}

		if r.Form.Get("_method") != http.MethodDelete {
			shared.WriteError(w, http.StatusBadRequest, "expected _method to be DELETE")
			return
		}

		err = deviceRepo.Delete(r.Context(), []models.Device{existingDevices[0]})
		if err != nil {
			shared.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		iconKey := fmt.Sprintf("device_icons/%s.%s", existingDevices[0].Slug, existingDevices[0].IconKind)
		err = bucket.Delete(r.Context(), iconKey)
		if err != nil {
			shared.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		http.Redirect(w, r, "/admin/devices", http.StatusSeeOther)
	}
}

func BuildUpdateHandler(
	db *sql.DB,
	bucket *blob.Bucket,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")

		id, err := shared.ParseIDFromPath(r, "deviceID")
		if err != nil {
			shared.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		deviceRepo := database.NewDeviceRepository(db)
		existingDevices, err := deviceRepo.FindByIDs(r.Context(), []int64{id})
		if err != nil {
			shared.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if len(existingDevices) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		err = r.ParseMultipartForm(32 << 20)
		if err != nil {
			shared.WriteError(w, http.StatusBadRequest, "failed to parse multipart form")
			return
		}

		if r.PostForm.Get("_method") != http.MethodPut {
			shared.WriteError(w, http.StatusBadRequest, "expected _method to be PUT")
			return
		}

		device := models.Device{
			ID:           existingDevices[0].ID,
			Name:         strings.TrimSpace(r.PostForm.Get("Name")),
			ModelMatches: strings.TrimSpace(r.PostForm.Get("ModelMatches")),
			IconKind:     existingDevices[0].IconKind,
		}

		_, header, err := r.FormFile("Icon")
		if err == nil {
			lowerFilename := strings.ToLower(header.Filename)
			if parts := strings.Split(lowerFilename, "."); len(parts) > 0 {
				device.IconKind = parts[len(parts)-1]
			} else {
				shared.WriteError(w, http.StatusBadRequest, "icon file missing extension")
				return
			}
			if device.IconKind != "jpg" && device.IconKind != "jpeg" && device.IconKind != "png" {
				shared.WriteError(w, http.StatusBadRequest, "icon file must be jpg or png, got: "+device.IconKind)
				return
			}
		}

		updatedDevices, err := deviceRepo.Update(r.Context(), []models.Device{device})
		if err != nil {
			shared.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if len(updatedDevices) != 1 {
			shared.WriteError(w, http.StatusInternalServerError, "unexpected number of updatedDevices")
			return
		}

		err = moveIconIfNeeded(r.Context(), bucket, existingDevices[0], updatedDevices[0])
		if err != nil {
			shared.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		err = saveIconIfProvided(r, bucket, updatedDevices[0], device.IconKind)
		if err != nil {
			shared.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		http.Redirect(
			w,
			r,
			fmt.Sprintf("/admin/devices/%d", updatedDevices[0].ID),
			http.StatusSeeOther,
		)
	}
}

func BuildFormHandler(
	db *sql.DB,
	bucket *blob.Bucket,
	_ templating.PageRenderer,
) func(http.ResponseWriter, *http.Request) {
	deleteHandler := BuildDeleteHandler(db, bucket)
	updateHandler := BuildUpdateHandler(db, bucket)

	return func(w http.ResponseWriter, r *http.Request) {
		contentType, ok := r.Header["Content-Type"]
		if !ok {
			shared.WriteError(w, http.StatusInternalServerError, "Content-Type must be set")
			return
		}

		switch {
		case contentType[0] == "application/x-www-form-urlencoded":
			deleteHandler(w, r)
		case strings.HasPrefix(contentType[0], "multipart/form-data"):
			updateHandler(w, r)
		default:
			shared.WriteError(w, http.StatusInternalServerError,
				"Content-Type must be 'multipart/form-data' or 'application/x-www-form-urlencoded'")
		}
	}
}

func moveIconIfNeeded(ctx context.Context, bucket *blob.Bucket, existing, updated models.Device) error {
	if existing.Slug == updated.Slug && existing.IconKind == updated.IconKind {
		return nil
	}

	existingIconKey := fmt.Sprintf("device_icons/%s.%s", existing.Slug, existing.IconKind)
	br, err := bucket.NewReader(ctx, existingIconKey, nil)
	if err != nil {
		return fmt.Errorf("failed initialize icon storage: %w", err)
	}
	defer br.Close()

	iconPath := fmt.Sprintf("device_icons/%s.%s", updated.Slug, updated.IconKind)
	bw, err := bucket.NewWriter(ctx, iconPath, nil)
	if err != nil {
		return fmt.Errorf("failed to open new writer for object: %w", err)
	}
	defer func() {
		closeErr := bw.Close()
		if closeErr != nil {
			err = fmt.Errorf("failed to close writer: %w", closeErr)
		}
	}()

	_, err = io.Copy(bw, br)
	if err != nil {
		return fmt.Errorf("failed to save to icon storage: %w", err)
	}

	err = bucket.Delete(ctx, existingIconKey)
	if err != nil {
		return fmt.Errorf("failed to delete old icon: %w", err)
	}

	return nil
}

func saveIconIfProvided(r *http.Request, bucket *blob.Bucket, device models.Device, iconKind string) error {
	f, _, err := r.FormFile("Icon")
	if err != nil {
		return nil
	}
	defer f.Close()

	iconPath := fmt.Sprintf("device_icons/%s.%s", device.Slug, iconKind)
	bw, err := bucket.NewWriter(r.Context(), iconPath, nil)
	if err != nil {
		return fmt.Errorf("failed initialize icon storage: %w", err)
	}
	defer func() {
		closeErr := bw.Close()
		if closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close writer: %w", closeErr)
		}
	}()

	_, err = io.Copy(bw, f)
	if err != nil {
		return fmt.Errorf("failed to save to icon storage: %w", err)
	}

	return nil
}
