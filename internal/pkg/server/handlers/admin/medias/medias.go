package medias

import (
	"bytes"
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gobuffalo/plush"
	"github.com/gorilla/mux"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/memblob"

	"github.com/charlieegan3/photos/internal/pkg/database"
	"github.com/charlieegan3/photos/internal/pkg/imageproxy"
	"github.com/charlieegan3/photos/internal/pkg/mediametadata"
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

// requiredThumbs is a list of the required thumbnail sizes
// Note: this must be ordered.
var requiredThumbs = []int{2000, 1000, 500, 200}

func BuildIndexHandler(db *sql.DB, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")

		medias, err := database.AllMedias(r.Context(), db, true)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		posts, err := database.AllPosts(r.Context(), db, true, database.SelectOptions{})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		postMediaMap := make(map[int]bool)
		for i := range posts {
			postMediaMap[posts[i].MediaID] = true
		}

		ctx := plush.NewContext()
		ctx.Set("medias", medias)
		ctx.Set("postMediaMap", postMediaMap)

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

		id, ok := mux.Vars(r)["mediaID"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("media id is required"))
			return
		}

		intID, err := strconv.Atoi(id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("failed to parse supplied ID"))
			return
		}

		medias, err := database.FindMediasByID(r.Context(), db, []int{intID})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		if len(medias) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		devices, err := database.AllDevices(r.Context(), db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		deviceOptionMap := make(map[string]interface{})
		for _, d := range devices {
			deviceOptionMap[d.Name] = d.ID
		}

		lenses, err := database.AllLenses(r.Context(), db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		lensOptionMap := map[string]interface{}{
			"No Lens": 0,
		}
		for _, l := range lenses {
			lensOptionMap[l.Name] = l.ID
		}

		posts, err := database.FindPostsByMediaID(r.Context(), db, medias[0].ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		ctx := plush.NewContext()
		ctx.Set("media", medias[0])
		ctx.Set("devices", deviceOptionMap)
		ctx.Set("lenses", lensOptionMap)
		ctx.Set("posts", posts)

		err = renderer(ctx, showTemplate, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
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

		id, ok := mux.Vars(r)["mediaID"]
		if !ok {
			shared.WriteError(w, http.StatusInternalServerError, "media id is required")
			return
		}

		intID, err := strconv.Atoi(id)
		if err != nil {
			shared.WriteError(w, http.StatusInternalServerError, "failed to parse supplied ID")
			return
		}

		existingMedias, err := database.FindMediasByID(r.Context(), db, []int{intID})
		if err != nil {
			shared.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if len(existingMedias) == 0 {
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

		err = database.DeleteMedias(r.Context(), db, []models.Media{existingMedias[0]})
		if err != nil {
			shared.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		err = deleteMediaFiles(r.Context(), bucket, existingMedias[0])
		if err != nil {
			shared.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		http.Redirect(w, r, "/admin/medias", http.StatusSeeOther)
	}
}

func BuildUpdateHandler(
	db *sql.DB,
	bucket *blob.Bucket,
) func(http.ResponseWriter, *http.Request) {
	ir := imageproxy.Resizer{}
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")

		id, ok := mux.Vars(r)["mediaID"]
		if !ok {
			shared.WriteError(w, http.StatusInternalServerError, "media id is required")
			return
		}

		intID, err := strconv.Atoi(id)
		if err != nil {
			shared.WriteError(w, http.StatusInternalServerError, "failed to parse supplied ID")
			return
		}

		existingMedias, err := database.FindMediasByID(r.Context(), db, []int{intID})
		if err != nil {
			shared.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if len(existingMedias) == 0 {
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

		media, err := parseMediaForm(r, existingMedias[0])
		if err != nil {
			shared.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		updatedMedias, err := database.UpdateMedias(r.Context(), db, []models.Media{media})
		if err != nil {
			shared.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if len(updatedMedias) != 1 {
			shared.WriteError(w, http.StatusInternalServerError, "unexpected number of updatedMedias")
			return
		}

		err = processMediaFileIfProvided(r, bucket, &ir, updatedMedias[0], media)
		if err != nil {
			shared.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		http.Redirect(
			w,
			r,
			fmt.Sprintf("/admin/medias/%d", updatedMedias[0].ID),
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

func deleteMediaFiles(ctx context.Context, bucket *blob.Bucket, media models.Media) error {
	mediaKey := fmt.Sprintf("media/%d.%s", media.ID, media.Kind)
	err := bucket.Delete(ctx, mediaKey)
	if err != nil {
		return fmt.Errorf("failed to delete media file: %w", err)
	}

	listOptions := &blob.ListOptions{
		Prefix: fmt.Sprintf("thumbs/%d-", media.ID),
	}
	iter := bucket.List(listOptions)
	for {
		obj, err := iter.Next(ctx)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to list thumbnails: %w", err)
		}

		err = bucket.Delete(ctx, obj.Key)
		if err != nil {
			return fmt.Errorf("failed to delete thumbnail: %w", err)
		}
	}

	return nil
}

func parseMediaForm(r *http.Request, existing models.Media) (models.Media, error) {
	floatKeys := []string{"FNumber", "Latitude", "Longitude", "Altitude"}
	floatMap := make(map[string]float64)
	for _, key := range floatKeys {
		floatMap[key] = 0
		if val := r.PostForm.Get(key); val != "" {
			f, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return models.Media{}, fmt.Errorf("float value %v was invalid", val)
			}
			floatMap[key] = f
		}
	}

	var isoSpeed int
	if val := r.PostForm.Get("ISOSpeed"); val != "" {
		var err error
		isoSpeed, err = strconv.Atoi(val)
		if err != nil {
			return models.Media{}, fmt.Errorf("int value %v was invalid", val)
		}
	}

	var displayOffset int
	if val := r.PostForm.Get("DisplayOffset"); val != "" {
		var err error
		displayOffset, err = strconv.Atoi(val)
		if err != nil {
			return models.Media{}, fmt.Errorf("int value %v was invalid", val)
		}
	}

	var orientation int
	if val := r.PostForm.Get("Orientation"); val != "" {
		var err error
		orientation, err = strconv.Atoi(val)
		if err != nil {
			return models.Media{}, fmt.Errorf("orientation value %v was invalid", val)
		}
		if orientation != 1 && orientation != 3 && orientation != 6 && orientation != 8 {
			return models.Media{}, fmt.Errorf("orientation must be 1, 3, 6, or 8, got %d", orientation)
		}
	} else {
		orientation = 1
	}

	var exposureTimeNumerator uint32
	if val := r.PostForm.Get("ExposureTimeNumerator"); val != "" {
		val64, err := strconv.ParseUint(val, 10, 32)
		if err != nil {
			return models.Media{}, fmt.Errorf("exposureTimeNumerator int value %v was invalid", val)
		}
		exposureTimeNumerator = uint32(val64)
	}

	var exposureTimeDenominator uint32
	if val := r.PostForm.Get("ExposureTimeDenominator"); val != "" {
		val64, err := strconv.ParseUint(val, 10, 32)
		if err != nil {
			return models.Media{}, fmt.Errorf("exposureTimeDenominator int value %v was invalid", val)
		}
		exposureTimeDenominator = uint32(val64)
	}

	var takenAt time.Time
	if val := r.PostForm.Get("TakenAt"); val != "" {
		var err error
		takenAt, err = time.Parse("2006-01-02T15:04", val)
		if err != nil {
			return models.Media{}, fmt.Errorf("time value %v was invalid", val)
		}
	}

	media := models.Media{
		ID:                      existing.ID,
		Kind:                    existing.Kind,
		Make:                    r.PostForm.Get("Make"),
		Model:                   r.PostForm.Get("Model"),
		Lens:                    r.PostForm.Get("Lens"),
		FocalLength:             r.PostForm.Get("FocalLength"),
		TakenAt:                 takenAt,
		ISOSpeed:                isoSpeed,
		ExposureTimeNumerator:   exposureTimeNumerator,
		ExposureTimeDenominator: exposureTimeDenominator,
		FNumber:                 floatMap["FNumber"],
		Latitude:                floatMap["Latitude"],
		Longitude:               floatMap["Longitude"],
		Altitude:                floatMap["Altitude"],
		Orientation:             orientation,
		DisplayOffset:           displayOffset,
		Width:                   existing.Width,
		Height:                  existing.Height,
	}

	deviceID, err := strconv.ParseInt(r.Form.Get("DeviceID"), 10, 64)
	if err != nil {
		return models.Media{}, fmt.Errorf("failed to parse device ID: %w", err)
	}
	media.DeviceID = deviceID

	rawLensID := r.Form.Get("LensID")
	if rawLensID != "" && rawLensID != "0" {
		lensID, err := strconv.ParseInt(rawLensID, 10, 0)
		if err != nil {
			return models.Media{}, fmt.Errorf("failed to parse lens ID: %w", err)
		}
		media.LensID = lensID
	}

	return media, nil
}

func processMediaFileIfProvided(
	r *http.Request,
	bucket *blob.Bucket,
	ir *imageproxy.Resizer,
	updated models.Media,
	media models.Media,
) error {
	f, header, err := r.FormFile("File")
	if err != nil {
		return nil
	}
	defer f.Close()

	lowerFilename := strings.ToLower(header.Filename)
	if !strings.HasSuffix(lowerFilename, ".jpg") &&
		!strings.HasSuffix(lowerFilename, ".jpeg") &&
		!strings.HasSuffix(lowerFilename, ".mp4") {
		return errors.New("media file must be jpg or mp4")
	}

	// Determine the file extension from the uploaded file
	fileKind := "jpg"
	if strings.HasSuffix(lowerFilename, ".mp4") {
		fileKind = "mp4"
	}

	key := fmt.Sprintf("media/%d.%s", updated.ID, fileKind)

	bw, err := bucket.NewWriter(r.Context(), key, nil)
	if err != nil {
		return fmt.Errorf("failed initialize media storage: %w", err)
	}

	_, err = io.Copy(bw, f)
	if err != nil {
		bw.Close()
		return fmt.Errorf("failed to save to media storage: %w", err)
	}

	// Close the writer before attempting to read
	err = bw.Close()
	if err != nil {
		return fmt.Errorf("failed to close media storage writer: %w", err)
	}

	br, err := bucket.NewReader(r.Context(), key, nil)
	if err != nil {
		return fmt.Errorf("failed to read from media storage: %w", err)
	}
	defer br.Close()

	imageBytes, err := io.ReadAll(br)
	if err != nil {
		return fmt.Errorf("failed to read from media storage: %w", err)
	}

	for _, thumbSize := range requiredThumbs {
		imageResizeString := fmt.Sprintf("%dx", thumbSize)
		if media.Width != 0 && media.Height != 0 {
			imageResizeString = fmt.Sprintf("%d,fit", thumbSize)
		}
		thumbMediaPath := fmt.Sprintf(
			"thumbs/media/%d-%s.%s",
			updated.ID,
			strings.Replace(imageResizeString, ",", "-", 1),
			fileKind,
		)

		imageBytes, err = ir.CreateThumbInBucket(
			r.Context(),
			bytes.NewReader(imageBytes),
			bucket,
			imageResizeString,
			thumbMediaPath,
		)
		if err != nil {
			return fmt.Errorf("failed to create thumbnail: %w", err)
		}
	}

	return nil
}

func BuildNewHandler(db *sql.DB, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")

		devices, err := database.AllDevices(r.Context(), db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		deviceOptionMap := make(map[string]interface{})
		for _, d := range devices {
			deviceOptionMap[d.Name] = d.ID
		}

		// device will be empty if there isn't a most recently used device in the current db
		device, err := database.MostRecentlyUsedDevice(r.Context(), db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		lenses, err := database.AllLenses(r.Context(), db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		lensOptionMap := map[string]interface{}{
			"No Lens": 0,
		}
		for _, l := range lenses {
			lensOptionMap[l.Name] = l.ID
		}

		ctx := plush.NewContext()
		ctx.Set("media", models.Media{DeviceID: device.ID})
		ctx.Set("devices", deviceOptionMap)
		ctx.Set("lenses", lensOptionMap)

		err = renderer(ctx, newTemplate, w)
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
	ir := imageproxy.Resizer{}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")

		if val, ok := r.Header["Content-Type"]; !ok || !strings.HasPrefix(val[0], "multipart/form-data") {
			shared.WriteError(w, http.StatusInternalServerError, "Content-Type must be 'multipart/form-data'")
			return
		}

		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			shared.WriteError(w, http.StatusBadRequest, "failed to parse multipart form: "+err.Error())
			return
		}

		media, err := parseCreateForm(r)
		if err != nil {
			shared.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		fileBytes, filename, err := processUploadedFile(r)
		if err != nil {
			shared.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		err = enrichMediaFromEXIF(&media, fileBytes, filename)
		if err != nil {
			shared.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		matchDeviceAndLens(r.Context(), db, &media)

		persistedMedias, err := database.CreateMedias(r.Context(), db, []models.Media{media})
		if err != nil {
			shared.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if len(persistedMedias) != 1 {
			shared.WriteError(w, http.StatusInternalServerError, "unexpected number of persistedMedias")
			return
		}

		err = saveMediaAndGenerateThumbs(r.Context(), bucket, &ir, persistedMedias[0], fileBytes)
		if err != nil {
			shared.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/admin/medias/%d", persistedMedias[0].ID), http.StatusSeeOther)
	}
}

func parseCreateForm(r *http.Request) (models.Media, error) {
	media := models.Media{
		Make:       r.Form.Get("Make"),
		UTCCorrect: true,
	}

	deviceID, err := strconv.ParseInt(r.Form.Get("DeviceID"), 10, 64)
	if err != nil {
		return models.Media{}, fmt.Errorf("failed to parse device ID: %w", err)
	}
	media.DeviceID = deviceID

	if r.Form.Get("LensID") != "" {
		lensID, err := strconv.ParseInt(r.Form.Get("LensID"), 10, 64)
		if err != nil {
			return models.Media{}, fmt.Errorf("failed to parse lens ID: %w", err)
		}
		media.LensID = lensID
	}

	return media, nil
}

func processUploadedFile(r *http.Request) ([]byte, string, error) {
	f, header, err := r.FormFile("File")
	if err != nil {
		return nil, "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer f.Close()

	lowerFilename := strings.ToLower(header.Filename)
	if !strings.HasSuffix(lowerFilename, ".jpg") && !strings.HasSuffix(lowerFilename, ".jpeg") {
		return nil, "", errors.New("media file must be jpg")
	}

	fileBytes, err := io.ReadAll(f)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read uploaded file data: %w", err)
	}

	return fileBytes, header.Filename, nil
}

func enrichMediaFromEXIF(media *models.Media, fileBytes []byte, filename string) error {
	exifData, err := mediametadata.ExtractMetadata(fileBytes)
	if err != nil {
		return fmt.Errorf("failed to get exif data file: %w", err)
	}

	latValue, _ := exifData.Latitude.ToDecimal()
	longValue, _ := exifData.Longitude.ToDecimal()
	if latValue == 0 && longValue == 0 {
		return errors.New("images must have a location set")
	}

	media.Make = exifData.Make
	media.Model = exifData.Model
	media.Lens = exifData.Lens
	media.FocalLength = exifData.FocalLength
	media.TakenAt = exifData.DateTime
	media.FNumber, _ = exifData.FNumber.ToDecimal()
	media.ExposureTimeNumerator = exifData.ExposureTime.Numerator
	media.ExposureTimeDenominator = exifData.ExposureTime.Denominator
	media.ISOSpeed = int(exifData.ISOSpeed)
	media.Latitude, _ = exifData.Latitude.ToDecimal()
	media.Longitude, _ = exifData.Longitude.ToDecimal()
	media.Altitude, _ = exifData.Altitude.ToDecimal()
	media.Orientation = int(exifData.Orientation)
	media.Width = exifData.Width
	media.Height = exifData.Height

	if len(strings.Split(strings.ToLower(filename), ".")) > 1 {
		parts := strings.Split(strings.ToLower(filename), ".")
		media.Kind = parts[len(parts)-1]
		if media.Kind == "jpeg" {
			media.Kind = "jpg"
		}
	} else {
		return errors.New("file must have name and extension")
	}

	return nil
}

func matchDeviceAndLens(ctx context.Context, db *sql.DB, media *models.Media) {
	deviceRepo := database.NewDeviceRepository(db)
	modelMatchedDevice, err := deviceRepo.FindByModelMatches(ctx, media.Model)
	if err == nil && modelMatchedDevice != nil {
		media.DeviceID = modelMatchedDevice.ID
	}

	lensMatchLens, err := database.FindLensByLensMatches(ctx, db, media.Lens)
	if err == nil && lensMatchLens != nil {
		media.LensID = lensMatchLens.ID
	}
}

func saveMediaAndGenerateThumbs(
	ctx context.Context,
	bucket *blob.Bucket,
	ir *imageproxy.Resizer,
	media models.Media,
	fileBytes []byte,
) error {
	key := fmt.Sprintf("media/%d.%s", media.ID, media.Kind)

	bw, err := bucket.NewWriter(ctx, key, nil)
	if err != nil {
		return fmt.Errorf("failed initialize media storage: %w", err)
	}

	_, err = io.Copy(bw, bytes.NewReader(fileBytes))
	if err != nil {
		bw.Close()
		return fmt.Errorf("failed to save to media storage: %w", err)
	}

	// Close the writer before attempting to read
	err = bw.Close()
	if err != nil {
		return fmt.Errorf("failed to close media storage writer: %w", err)
	}

	br, err := bucket.NewReader(ctx, key, nil)
	if err != nil {
		return fmt.Errorf("failed to read from media storage: %w", err)
	}
	defer br.Close()

	imageBytes, err := io.ReadAll(br)
	if err != nil {
		return fmt.Errorf("failed to read from media storage: %w", err)
	}

	for _, thumbSize := range requiredThumbs {
		imageResizeString := fmt.Sprintf("%dx", thumbSize)
		if media.Width != 0 && media.Height != 0 {
			imageResizeString = fmt.Sprintf("%d,fit", thumbSize)
		}
		thumbMediaPath := fmt.Sprintf(
			"thumbs/media/%d-%s.%s",
			media.ID,
			strings.Replace(imageResizeString, ",", "-", 1),
			media.Kind,
		)

		imageBytes, err = ir.CreateThumbInBucket(
			ctx,
			bytes.NewReader(imageBytes),
			bucket,
			imageResizeString,
			thumbMediaPath,
		)
		if err != nil {
			return fmt.Errorf("failed to create thumbnail: %w", err)
		}
	}

	return nil
}
