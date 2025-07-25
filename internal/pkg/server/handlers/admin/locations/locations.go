package locations

import (
	"database/sql"
	_ "embed"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gobuffalo/plush"
	"github.com/gorilla/mux"
	"gocloud.dev/blob"

	"github.com/charlieegan3/photos/internal/pkg/database"
	"github.com/charlieegan3/photos/internal/pkg/geoapify"
	"github.com/charlieegan3/photos/internal/pkg/models"
	"github.com/charlieegan3/photos/internal/pkg/server/templating"
)

//go:embed templates/index.html.plush
var indexTemplate string

//go:embed templates/show.html.plush
var showTemplate string

//go:embed templates/new.html.plush
var newTemplate string

//go:embed templates/select.html.plush
var selectTemplate string

//go:embed templates/lookup.html.plush
var lookupTemplate string

//go:embed templates/lookupForm.html.plush
var lookupFormTemplate string

func BuildIndexHandler(db *sql.DB, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		locations, err := database.AllLocations(db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		ctx := plush.NewContext()
		ctx.Set("locations", locations)

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

		rawID, ok := mux.Vars(r)["locationID"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("location id is required"))
			return
		}

		id, err := strconv.Atoi(rawID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to parse location ID"))
			return
		}

		locations, err := database.FindLocationsByID(db, []int{id})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(locations) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		ctx := plush.NewContext()
		ctx.Set("location", locations[0])

		err = renderer(ctx, showTemplate, w)
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

		var location models.Location

		latitude, ok := r.URL.Query()["lat"]
		if ok {
			if s, err := strconv.ParseFloat(latitude[0], 64); err == nil {
				location.Latitude = s
			}
		}
		longitude, ok := r.URL.Query()["long"]
		if ok {
			if s, err := strconv.ParseFloat(longitude[0], 64); err == nil {
				location.Longitude = s
			}
		}

		// this is only set when coming from lookup
		location.Name = r.URL.Query().Get("name")

		ctx := plush.NewContext()
		ctx.Set("location", location)

		err := renderer(ctx, newTemplate, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
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

		location := models.Location{
			Name: r.Form.Get("Name"),
		}

		latitudeString := r.Form.Get("Latitude")
		if s, err := strconv.ParseFloat(latitudeString, 64); err == nil {
			location.Latitude = s
		}

		longitudeString := r.Form.Get("Longitude")
		if s, err := strconv.ParseFloat(longitudeString, 64); err == nil {
			location.Longitude = s
		}

		persistedLocations, err := database.CreateLocations(db, []models.Location{location})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(persistedLocations) != 1 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unexpected number of persistedLocations"))
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/admin/locations/%d", persistedLocations[0].ID), http.StatusSeeOther)
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

		rawID, ok := mux.Vars(r)["locationID"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("location id is required"))
			return
		}

		id, err := strconv.Atoi(rawID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to parse location ID"))
			return
		}

		existingLocations, err := database.FindLocationsByID(db, []int{id})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(existingLocations) == 0 {
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

		if r.Form.Get("_method") == http.MethodDelete {
			mapKey := fmt.Sprintf("location_maps/%d.jpg", existingLocations[0].ID)
			exists, err := bucket.Exists(r.Context(), mapKey)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			if exists {
				err = bucket.Delete(r.Context(), mapKey)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}
			}
			err = database.DeleteLocations(db, []models.Location{existingLocations[0]})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			http.Redirect(w, r, "/admin/locations", http.StatusSeeOther)
			return
		}

		if r.PostForm.Get("_method") != http.MethodPut {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("expected _method to be PUT or DELETE in form"))
			return
		}

		name := r.PostForm.Get("Name")

		// handle the case where there's an existing location with the new name
		if name != existingLocations[0].Name {
			newLocationID, err := database.MergeLocations(db, name, existingLocations[0].Name)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			if newLocationID != 0 {
				mapKey := fmt.Sprintf("location_maps/%d.jpg", existingLocations[0].ID)
				exists, err := bucket.Exists(r.Context(), mapKey)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}
				if exists {
					err = bucket.Delete(r.Context(), mapKey)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(err.Error()))
						return
					}
				}
				http.Redirect(w, r, fmt.Sprintf("/admin/locations/%d", newLocationID), http.StatusSeeOther)
				return
			}
		}

		location := models.Location{
			ID:   existingLocations[0].ID,
			Name: name,
		}

		latitudeString := r.Form.Get("Latitude")
		if s, err := strconv.ParseFloat(latitudeString, 64); err == nil {
			location.Latitude = s
		}

		longitudeString := r.Form.Get("Longitude")
		if s, err := strconv.ParseFloat(longitudeString, 64); err == nil {
			location.Longitude = s
		}

		updatedLocations, err := database.UpdateLocations(db, []models.Location{location})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(updatedLocations) != 1 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unexpected number of updatedLocations"))
			return
		}

		http.Redirect(
			w,
			r,
			fmt.Sprintf("/admin/locations/%d", updatedLocations[0].ID),
			http.StatusSeeOther,
		)
	}
}

func BuildSelectHandler(db *sql.DB, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		redirectToRaw := r.URL.Query().Get("redirectTo")
		if redirectToRaw == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing redirectTo param"))
			return
		}

		params := r.URL.Query()
		params.Del("redirectTo")

		redirectToURL, err := url.QueryUnescape(redirectToRaw)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		redirectToURL += "?" + params.Encode()

		mediaIDRaw := r.URL.Query().Get("mediaID")
		if mediaIDRaw == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing mediaID param"))
			return
		}

		mediaID, err := strconv.ParseInt(mediaIDRaw, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid mediaID"))
			return
		}

		medias, err := database.FindMediasByID(db, []int{int(mediaID)})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(medias) != 1 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unexpected number of medias"))
			return
		}

		media := medias[0]

		nearbyLocations := []models.Location{}
		nearbyLocations, err = database.NearbyLocations(db, media.Latitude, media.Longitude)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to get points at that timestamp"))
			return
		}

		locations, err := database.AllLocations(db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		ctx := plush.NewContext()
		ctx.Set("locations", locations)
		ctx.Set("nearbyLocations", nearbyLocations)
		ctx.Set("lat", media.Latitude)
		ctx.Set("long", media.Longitude)
		ctx.Set("redirectTo", redirectToURL)

		err = renderer(ctx, selectTemplate, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}
}

func BuildLookupHandler(mapServerAPIKey string, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		query := r.URL.Query().Get("query")
		if query == "" {
			err := renderer(plush.NewContext(), lookupFormTemplate, w)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
			}
			return
		}

		client, err := geoapify.NewClient("https://api.geoapify.com", mapServerAPIKey)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		features, err := client.GeocodingSearch(query)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		ctx := plush.NewContext()
		ctx.Set("features", features)

		err = renderer(ctx, lookupTemplate, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
		return
	}
}
