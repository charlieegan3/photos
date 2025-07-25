package public

import (
	"database/sql"
	_ "embed"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gobuffalo/plush"
	"github.com/gorilla/mux"

	"github.com/charlieegan3/photos/internal/pkg/database"
	"github.com/charlieegan3/photos/internal/pkg/models"
	"github.com/charlieegan3/photos/internal/pkg/server/templating"
)

//go:embed templates/show.html.plush
var showTemplate string

func BuildGetHandler(db *sql.DB, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		rawID, ok := mux.Vars(r)["tripID"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("trip id is required"))
			return
		}

		id, err := strconv.Atoi(rawID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to parse trip ID"))
			return
		}

		trips, err := database.FindTripsByID(db, []int{id})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(trips) != 1 {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("trip not found"))
			return
		}

		trip := trips[0]

		toTime := trip.EndDate.Add(24 * time.Hour).Add(-time.Second)

		timeFormat := "January 2"

		dateTitle := fmt.Sprintf(
			"%s - %s",
			trip.StartDate.Format(timeFormat),
			trip.EndDate.Format(timeFormat+", 2006"),
		)
		if trip.StartDate.Year() != trip.EndDate.Year() {
			dateTitle = fmt.Sprintf(
				"%s - %s",
				trip.StartDate.Format(timeFormat+", 2006"),
				trip.EndDate.Format(timeFormat+", 2006"),
			)
		}

		if trip.StartDate.Equal(trip.EndDate) {
			dateTitle = trip.StartDate.Format(timeFormat)
		}
		if trip.StartDate.Month() == trip.EndDate.Month() {
			dateTitle = fmt.Sprintf(
				"%s-%s, %s",
				trip.StartDate.Format("January 2"),
				trip.EndDate.Format("2"),
				trip.EndDate.Format("2006"),
			)
		}

		posts, err := database.PostsInDateRange(db, trip.StartDate, toTime)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(posts) == 0 {
			http.Redirect(w, r, "/posts/period", http.StatusFound)
			return
		}

		var locationIDs []int
		for _, p := range posts {
			locationIDs = append(locationIDs, p.LocationID)
		}

		locations, err := database.FindLocationsByID(db, locationIDs)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		locationsByID := make(map[int]models.Location)
		for _, l := range locations {
			locationsByID[l.ID] = l
		}

		var mediaIDs []int
		for _, p := range posts {
			mediaIDs = append(mediaIDs, p.MediaID)
		}

		medias, err := database.FindMediasByID(db, mediaIDs)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		mediasByID := make(map[int]models.Media)
		for _, m := range medias {
			mediasByID[m.ID] = m
		}

		postGroupKeys := []string{}
		postGroups := make(map[string][]models.Post)
		for _, p := range posts {
			key := p.PublishDate.Format(timeFormat)
			if _, ok := postGroups[key]; !ok {
				postGroups[key] = []models.Post{}
				postGroupKeys = append(postGroupKeys, key)
			}
			postGroups[key] = append(postGroups[key], p)
		}

		showDates := trip.StartDate.Add(24 * time.Hour).Before(toTime)

		if len(postGroupKeys) == 1 {
			dateTitle = postGroupKeys[0]
			showDates = false
		}

		ctx := plush.NewContext()
		ctx.Set("postGroupKeys", postGroupKeys)
		ctx.Set("postGroups", postGroups)
		ctx.Set("locations", locationsByID)
		ctx.Set("medias", mediasByID)
		ctx.Set("trip", trip)
		ctx.Set("timeFormat", timeFormat)
		ctx.Set("dateTitle", dateTitle)
		ctx.Set("showDates", showDates)

		err = renderer(ctx, showTemplate, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}
}
