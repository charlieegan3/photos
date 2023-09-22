package medias

import (
	"bytes"
	"database/sql"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gobuffalo/plush"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/memblob"

	"github.com/charlieegan3/photos/internal/pkg/database"
	"github.com/charlieegan3/photos/internal/pkg/imageproxy"
	"github.com/charlieegan3/photos/internal/pkg/mediametadata"
	"github.com/charlieegan3/photos/internal/pkg/models"
	"github.com/charlieegan3/photos/internal/pkg/server/templating"
)

//go:embed templates/index.html.plush
var indexTemplate string

//go:embed templates/new.html.plush
var newTemplate string

//go:embed templates/show.html.plush
var showTemplate string

// gorilla decoder can be safely shared and caches data on structs used
var decoder = schema.NewDecoder()

// requiredThumbs is a list of the required thumbnail sizes
// Note: this must be ordered
var requiredThumbs = []int{2000, 1000, 500, 200}

func BuildIndexHandler(db *sql.DB, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		medias, err := database.AllMedias(db, true)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		posts, err := database.AllPosts(db, true, database.SelectOptions{})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		postMediaMap := make(map[int]bool)
		for _, p := range posts {
			postMediaMap[p.MediaID] = true
		}

		ctx := plush.NewContext()
		ctx.Set("medias", medias)
		ctx.Set("postMediaMap", postMediaMap)

		err = renderer(ctx, indexTemplate, w)
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

		id, ok := mux.Vars(r)["mediaID"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("media id is required"))
			return
		}

		intID, err := strconv.Atoi(id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to parse supplied ID"))
			return
		}

		medias, err := database.FindMediasByID(db, []int{intID})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(medias) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		devices, err := database.AllDevices(db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		deviceOptionMap := make(map[string]interface{})
		for _, d := range devices {
			deviceOptionMap[d.Name] = d.ID
		}

		lenses, err := database.AllLenses(db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		lensOptionMap := map[string]interface{}{
			"No Lens": 0,
		}
		for _, l := range lenses {
			lensOptionMap[l.Name] = l.ID
		}

		posts, err := database.FindPostsByMediaID(db, medias[0].ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
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
			w.Write([]byte(err.Error()))
			return
		}
	}
}

func BuildFormHandler(db *sql.DB, bucket *blob.Bucket, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	ir := imageproxy.Resizer{}
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		contentType, ok := r.Header["Content-Type"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Content-Type must be set"))
			return
		}

		id, ok := mux.Vars(r)["mediaID"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("media id is required"))
			return
		}

		intID, err := strconv.Atoi(id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to parse supplied ID"))
			return
		}

		existingMedias, err := database.FindMediasByID(db, []int{intID})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(existingMedias) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// handle delete, of database item and blob
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

			err = database.DeleteMedias(db, []models.Media{existingMedias[0]})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			mediaKey := fmt.Sprintf("media/%d.%s", existingMedias[0].ID, existingMedias[0].Kind)
			err = bucket.Delete(r.Context(), mediaKey)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			listOptions := &blob.ListOptions{
				Prefix: fmt.Sprintf("thumbs/%d-", existingMedias[0].ID),
			}
			iter := bucket.List(listOptions)
			for {
				obj, err := iter.Next(r.Context())
				if err == io.EOF {
					break
				}
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}

				err = bucket.Delete(r.Context(), obj.Key)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}
			}

			http.Redirect(w, r, "/admin/medias", http.StatusSeeOther)
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

		floatKeys := []string{"FNumber", "Latitude", "Longitude", "Altitude"}
		floatMap := make(map[string]float64)
		for _, key := range floatKeys {
			floatMap[key] = 0
			if val := r.PostForm.Get(key); val != "" {
				f, err := strconv.ParseFloat(val, 64)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(fmt.Sprintf("float value %v was invalid", val)))
					return
				}
				floatMap[key] = f
			}
		}

		var isoSpeed int
		if val := r.PostForm.Get("ISOSpeed"); val != "" {
			isoSpeed, err = strconv.Atoi(val)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("int value %v was invalid", val)))
				return
			}
		}

		var displayOffset int
		if val := r.PostForm.Get("DisplayOffset"); val != "" {
			displayOffset, err = strconv.Atoi(val)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("int value %v was invalid", val)))
				return
			}
		}

		var exposureTimeNumerator uint64
		if val := r.PostForm.Get("ExposureTimeNumerator"); val != "" {
			exposureTimeNumerator, err = strconv.ParseUint(val, 10, 32)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("exposureTimeNumerator int value %v was invalid", val)))
				return
			}
		}

		var exposureTimeDenominator uint64
		if val := r.PostForm.Get("ExposureTimeDenominator"); val != "" {
			exposureTimeDenominator, err = strconv.ParseUint(val, 10, 32)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("exposureTimeDenominator int value %v was invalid", val)))
				return
			}
		}

		var takenAt time.Time
		if val := r.PostForm.Get("TakenAt"); val != "" {
			takenAt, err = time.Parse("2006-01-02T15:04", val)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("time value %v was invalid", val)))
				return
			}
		}

		media := models.Media{
			ID:                      existingMedias[0].ID,
			Kind:                    existingMedias[0].Kind,
			Make:                    r.PostForm.Get("Make"),
			Model:                   r.PostForm.Get("Model"),
			Lens:                    r.PostForm.Get("Lens"),
			FocalLength:             r.PostForm.Get("FocalLength"),
			TakenAt:                 takenAt,
			ISOSpeed:                isoSpeed,
			ExposureTimeNumerator:   uint32(exposureTimeNumerator),
			ExposureTimeDenominator: uint32(exposureTimeDenominator),
			FNumber:                 floatMap["FNumber"],
			Latitude:                floatMap["Latitude"],
			Longitude:               floatMap["Longitude"],
			Altitude:                floatMap["Altitude"],
			DisplayOffset:           displayOffset,
			Width:                   existingMedias[0].Width,
			Height:                  existingMedias[0].Height,
		}

		media.DeviceID, err = strconv.ParseInt(r.Form.Get("DeviceID"), 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("failed to parse device ID"))
			return
		}

		// only handle the lens if set, it's optional
		rawLensID := r.Form.Get("LensID")
		if rawLensID != "" && rawLensID != "0" {
			media.LensID, err = strconv.ParseInt(rawLensID, 10, 0)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed to parse lens ID"))
				return
			}
		}

		updatedMedias, err := database.UpdateMedias(db, []models.Media{media})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(updatedMedias) != 1 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unexpected number of updatedMedias"))
			return
		}

		// only handle the file when it's present, file might not be submitted
		// every time the form is sent
		f, header, err := r.FormFile("File")
		if err == nil {
			lowerFilename := strings.ToLower(header.Filename)
			if !strings.HasSuffix(lowerFilename, ".jpg") &&
				!strings.HasSuffix(lowerFilename, ".jpeg") &&
				!strings.HasSuffix(lowerFilename, ".mp4") {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("media file must be jpg or mp4"))
				return
			}

			key := fmt.Sprintf("media/%d.%s", updatedMedias[0].ID, updatedMedias[0].Kind)

			bw, err := bucket.NewWriter(r.Context(), key, nil)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed initialize media storage"))
				return
			}

			_, err = io.Copy(bw, f)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed to save to media storage"))
				return
			}

			err = bw.Close()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed to close connection to media storage"))
				return
			}

			br, err := bucket.NewReader(r.Context(), key, nil)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed to read from media storage"))
				return
			}
			defer br.Close()

			imageBytes, err := io.ReadAll(br)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed to read from media storage"))
				return
			}

			for _, thumbSize := range requiredThumbs {
				var err error
				imageResizeString := fmt.Sprintf("%dx", thumbSize)
				if media.Width != 0 && media.Height != 0 {
					imageResizeString = fmt.Sprintf("%d,fit", thumbSize)
				}
				thumbMediaPath := fmt.Sprintf(
					"thumbs/media/%d-%s.%s",
					updatedMedias[0].ID,
					strings.Replace(imageResizeString, ",", "-", 1),
					updatedMedias[0].Kind,
				)

				// use the downsampled bytes from for the next thumb to save time
				imageBytes, err = ir.CreateThumbInBucket(
					r.Context(),
					bytes.NewReader(imageBytes),
					bucket,
					imageResizeString,
					thumbMediaPath,
				)
				if err != nil {
					w.Header().Set("Content-Type", "application/text")
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}
			}
		}

		http.Redirect(
			w,
			r,
			fmt.Sprintf("/admin/medias/%d", updatedMedias[0].ID),
			http.StatusSeeOther,
		)
	}
}

func BuildNewHandler(db *sql.DB, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		devices, err := database.AllDevices(db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		deviceOptionMap := make(map[string]interface{})
		for _, d := range devices {
			deviceOptionMap[d.Name] = d.ID
		}

		// device will be empty if there isn't a most recently used device in the current db
		device, err := database.MostRecentlyUsedDevice(db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		lenses, err := database.AllLenses(db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
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
			w.Write([]byte(err.Error()))
			return
		}
	}
}

func BuildCreateHandler(db *sql.DB, bucket *blob.Bucket, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	ir := imageproxy.Resizer{}

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
			w.Write([]byte(fmt.Sprintf("failed to parse multipart form: %s", err.Error())))
			return
		}

		media := models.Media{
			Make: r.Form.Get("Make"),
			// new uploads from 2022-08-10 will have this set to true since we can only now trust the time
			UTCCorrect: true,
		}

		// may be overridden if the EXIF data matches an existing device
		media.DeviceID, err = strconv.ParseInt(r.Form.Get("DeviceID"), 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("failed to parse device ID"))
			return
		}

		if r.Form.Get("LensID") != "" {
			media.LensID, err = strconv.ParseInt(r.Form.Get("LensID"), 10, 64)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("failed to parse lens ID"))
				return
			}
		}

		f, header, err := r.FormFile("File")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("failed to open uploaded file: %s", err)))
			return
		}

		lowerFilename := strings.ToLower(header.Filename)
		if !strings.HasSuffix(lowerFilename, ".jpg") &&
			!strings.HasSuffix(lowerFilename, ".jpeg") {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("media file must be jpg"))
			return
		}

		if parts := strings.Split(lowerFilename, "."); len(parts) > 1 {
			media.Kind = parts[len(parts)-1]
			if media.Kind == "jpeg" {
				media.Kind = "jpg"
			}
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("file must have name and extension"))
			return
		}

		fileBytes, err := io.ReadAll(f)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("failed to read uploaded file data: %s", err)))
			return
		}

		exifData, err := mediametadata.ExtractMetadata(fileBytes)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("failed to get exif data file: %s", err)))
			return
		}

		// location information is now required for all uploaded images.
		// images are now updated with https://github.com/charlieegan3/gpxif
		latValue, _ := exifData.Latitude.ToDecimal()
		longValue, _ := exifData.Longitude.ToDecimal()
		if latValue == 0 && longValue == 0 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("images must have a location set"))
			return
		}

		media.Make = exifData.Make
		media.Model = exifData.Model
		media.Lens = exifData.Lens
		media.FocalLength = exifData.FocalLength
		media.TakenAt = exifData.DateTime
		// TODO handle exif errors
		media.FNumber, err = exifData.FNumber.ToDecimal()
		media.ExposureTimeNumerator = exifData.ExposureTime.Numerator
		media.ExposureTimeDenominator = exifData.ExposureTime.Denominator
		media.ISOSpeed = int(exifData.ISOSpeed)
		media.Latitude, err = exifData.Latitude.ToDecimal()
		media.Longitude, err = exifData.Longitude.ToDecimal()
		media.Altitude, err = exifData.Altitude.ToDecimal()
		media.Width = exifData.Width
		media.Height = exifData.Height

		// if there's a match from the EXIF data, then use that to set the device ID
		modelMatchedDevice, err := database.FindDeviceByModelMatches(db, exifData.Model)
		if err == nil {
			if modelMatchedDevice != nil {
				media.DeviceID = modelMatchedDevice.ID
			}
		}

		lensMatchLens, err := database.FindLensByLensMatches(db, exifData.Lens)
		if err == nil {
			if lensMatchLens != nil {
				media.LensID = lensMatchLens.ID
			}
		}

		persistedMedias, err := database.CreateMedias(db, []models.Media{media})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(persistedMedias) != 1 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unexpected number of persistedMedias"))
			return
		}

		key := fmt.Sprintf("media/%d.%s", persistedMedias[0].ID, persistedMedias[0].Kind)

		bw, err := bucket.NewWriter(r.Context(), key, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed initialize media storage"))
			return
		}

		_, err = io.Copy(bw, bytes.NewReader(fileBytes))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to save to media storage"))
			return
		}

		err = bw.Close()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to close connection to media storage"))
			return
		}

		br, err := bucket.NewReader(r.Context(), key, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to read from media storage"))
			return
		}
		defer br.Close()

		imageBytes, err := io.ReadAll(br)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to read from media storage"))
			return
		}

		for _, thumbSize := range requiredThumbs {
			var err error
			imageResizeString := fmt.Sprintf("%dx", thumbSize)
			if media.Width != 0 && media.Height != 0 {
				imageResizeString = fmt.Sprintf("%d,fit", thumbSize)
			}
			thumbMediaPath := fmt.Sprintf(
				"thumbs/media/%d-%s.%s",
				persistedMedias[0].ID,
				strings.Replace(imageResizeString, ",", "-", 1),
				persistedMedias[0].Kind,
			)

			// use the downsampled bytes from for the next thumb to save time
			imageBytes, err = ir.CreateThumbInBucket(
				r.Context(),
				bytes.NewReader(imageBytes),
				bucket,
				imageResizeString,
				thumbMediaPath,
			)
			if err != nil {
				w.Header().Set("Content-Type", "application/text")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
		}

		http.Redirect(w, r, fmt.Sprintf("/admin/medias/%d", persistedMedias[0].ID), http.StatusSeeOther)
	}
}
