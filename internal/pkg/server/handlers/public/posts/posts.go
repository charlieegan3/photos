package public

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/gobuffalo/plush"
	"github.com/gomarkdown/markdown"
	"github.com/gorilla/feeds"
	"github.com/gorilla/mux"

	"github.com/charlieegan3/photos/internal/pkg/database"
	"github.com/charlieegan3/photos/internal/pkg/models"
	"github.com/charlieegan3/photos/internal/pkg/server/templating"
)

//go:embed templates/index.html.plush
var indexTemplate string

//go:embed templates/searchForm.html.plush
var searchFormTemplate string

//go:embed templates/search.html.plush
var searchTemplate string

//go:embed templates/period.html.plush
var periodTemplate string

//go:embed templates/periodIndex.html.plush
var periodIndexTemplate string

//go:embed templates/periodMissing.html.plush
var periodMissingTemplate string

//go:embed templates/show.html.plush
var showTemplate string

//go:embed templates/show-wide.html.plush
var showWideTemplate string

//go:embed templates/on-this-day.html.plush
var onThisDayTemplate string

var pageSize uint = 42

func BuildIndexHandler(db *sql.DB, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		pageParam := r.URL.Query().Get("page")
		var page uint = 1
		if pageParam != "" {
			parsedPage, err := strconv.Atoi(pageParam)
			if err == nil {
				if parsedPage < 2 { // first page also strip param
					http.Redirect(w, r, "/", http.StatusSeeOther)
					return
				}
				page = uint(parsedPage)
			}
		}

		posts, err := database.AllPosts(
			db,
			false,
			database.SelectOptions{
				SortField:      "publish_date",
				SortDescending: true,
				Limit:          pageSize,
				Offset:         (page - 1) * pageSize,
			},
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		postsCount, err := database.CountPosts(
			r.Context(),
			db,
			false,
			database.SelectOptions{},
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		var mediaIDs []int
		for i := range posts {
			mediaIDs = append(mediaIDs, posts[i].MediaID)
		}

		medias, err := database.FindMediasByID(r.Context(), db, mediaIDs)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		mediasByID := make(map[int]models.Media)
		for i := range medias {
			mediasByID[medias[i].ID] = medias[i]
		}

		lastPage := postsCount/int64(pageSize) + 1
		if int64(page) > lastPage {
			http.Redirect(w, r, fmt.Sprintf("/?page=%d", lastPage), http.StatusSeeOther)
			return
		}

		ctx := plush.NewContext()
		ctx.Set("posts", posts)
		ctx.Set("medias", mediasByID)
		ctx.Set("page", int(page))
		ctx.Set("lastPage", int(lastPage))

		err = renderer(ctx, indexTemplate, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}
}

func BuildSearchHandler(db *sql.DB, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		queryParam := r.URL.Query().Get("query")
		if queryParam == "" {
			err := renderer(plush.NewContext(), searchFormTemplate, w)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			return
		}

		// this is cleaned here since it's also shown in the HTML, it's also cleaned in the database package
		safeQuery := regexp.MustCompile(`[^\w\s]+`).ReplaceAllString(queryParam, "")
		safeQuery = regexp.MustCompile(`[\s]+`).ReplaceAllString(safeQuery, " ")

		posts, err := database.SearchPosts(r.Context(), db, safeQuery)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		mediaIDs := make([]int, len(posts))
		mediasByID := make(map[int]models.Media)
		if len(posts) > 0 {
			for i := range posts {
				mediaIDs[i] = posts[i].MediaID
			}

			medias, err := database.FindMediasByID(r.Context(), db, mediaIDs)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			for i := range medias {
				mediasByID[medias[i].ID] = medias[i]
			}
		}

		ctx := plush.NewContext()
		ctx.Set("posts", posts)
		ctx.Set("query", safeQuery)
		ctx.Set("medias", mediasByID)

		err = renderer(ctx, searchTemplate, w)
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

		rawID, ok := mux.Vars(r)["postID"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("post ID is required"))
			return
		}

		id, err := strconv.Atoi(rawID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("postID was not integer"))
			return
		}

		posts, err := database.FindPostsByID(r.Context(), db, []int{id})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(posts) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// TODO create a db function to get tags for post in SQL
		taggings, err := database.FindTaggingsByPostID(db, posts[0].ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		var tagIDs []int
		for _, t := range taggings {
			tagIDs = append(tagIDs, t.TagID)
		}

		tags, err := database.FindTagsByID(r.Context(), db, tagIDs)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		medias, err := database.FindMediasByID(r.Context(), db, []int{posts[0].MediaID})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		if len(medias) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		locations, err := database.FindLocationsByID(r.Context(), db, []int{posts[0].LocationID})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		if len(locations) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		devices, err := database.FindDevicesByID(r.Context(), db, []int64{medias[0].DeviceID})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		if len(devices) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		lenses, err := database.FindLensesByID(r.Context(), db, []int64{medias[0].LensID})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		nextPosts, err := database.FindNextPost(db, posts[0], false)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		previousPosts, err := database.FindNextPost(db, posts[0], true)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		ctx := plush.NewContext()
		ctx.Set("post", posts[0])
		ctx.Set("nextPost", 0)
		if len(nextPosts) > 0 {
			ctx.Set("nextPost", nextPosts[0].ID)
		}
		ctx.Set("previousPost", 0)
		if len(previousPosts) > 0 {
			ctx.Set("previousPost", previousPosts[0].ID)
		}
		ctx.Set("media", medias[0])
		ctx.Set("device", devices[0])
		ctx.Set("lenses", lenses)
		ctx.Set("location", locations[0])
		ctx.Set("tags", tags)

		if medias[0].Width > medias[0].Height {
			err = renderer(ctx, showWideTemplate, w)
		} else {
			err = renderer(ctx, showTemplate, w)
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}
}

// BuildLegacyPostRedirect will send requests to old post IDs to the period pages as a best guess
func BuildLegacyPostRedirect() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		date, ok := mux.Vars(r)["date"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to parse date from legacy URL"))
			return
		}
		http.Redirect(w, r, "/posts/period/"+date, http.StatusMovedPermanently)
	}
}

// BuildLegacyPeriodRedirect will transform requests for /archive to /posts/period paths
func BuildLegacyPeriodRedirect() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		month, monthOk := mux.Vars(r)["month"]
		day, dayOk := mux.Vars(r)["day"]
		if !monthOk || !dayOk {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to parse legacy archive URL"))
			return
		}
		date, err := time.Parse("01-02", fmt.Sprintf("%s-%s", month, day))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to parse date in legacy URL"))
			return
		}
		http.Redirect(
			w,
			r,
			fmt.Sprintf("/posts/on-this-day/%s-%s", date.Format("January"), date.Format("2")),
			http.StatusMovedPermanently,
		)
	}
}

func BuildPeriodHandler(db *sql.DB, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		fromString, ok := mux.Vars(r)["from"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("from param required"))
			return
		}

		fromTime, err := time.Parse("2006-01-02", fromString)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid from date format"))
			return
		}

		toTime := fromTime.Add(24 * time.Hour)
		toString, ok := mux.Vars(r)["to"]
		if ok {
			toTime, err = time.Parse("2006-01-02", toString)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("invalid from date format"))
				return
			}
			toTime = toTime.Add(24 * time.Hour).Add(-time.Second)
		}

		showDates := true
		title := fmt.Sprintf("Posts from %v to %v", fromTime.Format("January 2, 2006"), toTime.Format("January 2, 2006"))
		timeFormat := "January 2, 2006"
		if fromTime.Year() == toTime.Year() {
			title = fmt.Sprintf("Posts from %v to %v, %d",
				fromTime.Format("January 2"), toTime.Format("January 2"), fromTime.Year())
			timeFormat = "January 2"

			if fromTime.Month() == toTime.Month() {
				title = fmt.Sprintf("Posts from %s %v-%v, %d",
					fromTime.Month(), fromTime.Format("2"), toTime.Format("2"), fromTime.Year())
			}
		}
		if fromTime.Add(24*time.Hour).After(toTime) || fromTime.Add(24*time.Hour).Equal(toTime) {
			title = fmt.Sprintf("Posts from %v", fromTime.Format("January 2, 2006"))
			showDates = false
		}

		posts, err := database.PostsInDateRange(r.Context(), db, fromTime, toTime)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(posts) == 0 {
			w.WriteHeader(http.StatusNotFound)
			err := renderer(plush.NewContext(), periodMissingTemplate, w)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			return
		}

		var locationIDs []int
		// TODO fetch these with posts
		for i := range posts {
			locationIDs = append(locationIDs, posts[i].LocationID)
		}

		locations, err := database.FindLocationsByID(r.Context(), db, locationIDs)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		locationsByID := make(map[int]models.Location)
		for _, l := range locations {
			locationsByID[l.ID] = l
		}

		// TODO fetch these with posts
		var mediaIDs []int
		for i := range posts {
			mediaIDs = append(mediaIDs, posts[i].MediaID)
		}

		medias, err := database.FindMediasByID(r.Context(), db, mediaIDs)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		mediasByID := make(map[int]models.Media)
		for i := range medias {
			mediasByID[medias[i].ID] = medias[i]
		}

		postGroupKeys := []string{}
		postGroups := make(map[string][]models.Post)
		for i := range posts {
			key := posts[i].PublishDate.Format(timeFormat)
			if _, ok := postGroups[key]; !ok {
				postGroups[key] = []models.Post{}
				postGroupKeys = append(postGroupKeys, key)
			}
			postGroups[key] = append(postGroups[key], posts[i])
		}

		ctx := plush.NewContext()
		ctx.Set("postGroupKeys", postGroupKeys)
		ctx.Set("postGroups", postGroups)
		ctx.Set("locations", locationsByID)
		ctx.Set("medias", mediasByID)
		ctx.Set("title", title)
		ctx.Set("showDates", showDates)

		err = renderer(ctx, periodTemplate, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}
}

func BuildPeriodIndexHandler(db *sql.DB, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		fromString := r.URL.Query().Get("from")
		if fromString == "" {
			trips, err := database.AllTrips(r.Context(), db)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			ctx := plush.NewContext()
			ctx.Set("trips", trips)

			err = renderer(ctx, periodIndexTemplate, w)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			return
		}

		toString := r.URL.Query().Get("to")
		if toString != "" {
			http.Redirect(w, r, fmt.Sprintf("/posts/period/%s-to-%s", fromString, toString), http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/posts/period/"+fromString, http.StatusSeeOther)
	}
}

func BuildLatestHandler(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		posts, err := database.AllPosts(
			db,
			false,
			database.SelectOptions{
				SortField:      "publish_date",
				SortDescending: true,
				Limit:          1,
			},
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(posts) < 1 {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("no post found"))
			return
		}

		// TODO get in same query
		locations, err := database.FindLocationsByID(r.Context(), db, []int{posts[0].LocationID})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(locations) < 1 {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("no location found"))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(struct {
			Location  string `json:"location"`
			URL       string `json:"url"`
			CreatedAt string `json:"created_at"`
		}{
			Location:  locations[0].Name,
			URL:       fmt.Sprintf("https://photos.charlieegan3.com/posts/%d", posts[0].ID),
			CreatedAt: posts[0].PublishDate.Format(time.RFC3339),
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}
}

func BuildRSSHandler(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")

		posts, err := database.AllPosts(
			db,
			false,
			database.SelectOptions{
				SortField:      "publish_date",
				SortDescending: true,
				Limit:          25,
			},
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(posts) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var mediaIDs []int
		for i := range posts {
			mediaIDs = append(mediaIDs, posts[i].MediaID)
		}

		medias, err := database.FindMediasByID(r.Context(), db, mediaIDs)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		if len(medias) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		mediaMap := make(map[int]models.Media)
		for i := range medias {
			mediaMap[medias[i].ID] = medias[i]
		}

		locations, err := database.FindLocationsByID(r.Context(), db, []int{posts[0].LocationID})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		if len(locations) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		locationMap := make(map[int]models.Location)
		for _, l := range locations {
			locationMap[l.ID] = l
		}

		devices, err := database.AllDevices(r.Context(), db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		if len(devices) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		deviceMap := make(map[int64]models.Device)
		for _, l := range devices {
			deviceMap[l.ID] = l
		}

		feed := &feeds.Feed{
			Title:       "photos.charlieegan3.com - All",
			Link:        &feeds.Link{Href: "https://photos.charlieegan3.com/rss.xml"},
			Description: "RSS feed of all photos",
			Author:      &feeds.Author{Name: "Charlie Egan", Email: "me@charlieegan3.com"},
		}

		var feedItems []*feeds.Item
		for i := range posts {
			md := fmt.Sprintf("%s\n\n%s\n\n%s",
				posts[i].Description,
				fmt.Sprintf("![post image](https://photos.charlieegan3.com/medias/%d/image.jpg?o=1000,fit)", posts[i].MediaID),
				"Taken on "+deviceMap[mediaMap[posts[i].MediaID].DeviceID].Name,
			)

			content := markdown.NormalizeNewlines([]byte(md))

			feedItems = append(feedItems,
				&feeds.Item{
					Id:          fmt.Sprintf("https://photos.charlieegan3.com/posts/%d", posts[i].ID),
					Title:       fmt.Sprintf("%s - %s", posts[i].PublishDate.Format("January 2, 2006"), locationMap[posts[i].LocationID].Name),
					Link:        &feeds.Link{Href: fmt.Sprintf("https://photos.charlieegan3.com/posts/%d", posts[i].ID)},
					Description: string(markdown.ToHTML(content, nil, nil)),
					Created:     posts[i].PublishDate,
				})
		}

		feed.Items = feedItems

		rssFeed := (&feeds.Rss{Feed: feed}).RssFeed()
		output, err := xml.MarshalIndent(rssFeed, "", "    ")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8" ?>
<rss version="2.0">
`))
		w.Write([]byte(output))

		w.Write([]byte("\n</rss>"))
	}
}

func BuildOnThisDayHandler(db *sql.DB, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		rawMonth, monthOk := mux.Vars(r)["month"]
		rawDay, dayOk := mux.Vars(r)["day"]
		day, dayErr := strconv.Atoi(rawDay)
		month, monthErr := time.Parse("January", rawMonth)
		if dayErr != nil || monthErr != nil || !monthOk || !dayOk {
			http.Redirect(
				w, r,
				fmt.Sprintf(
					"/posts/on-this-day/%s-%d",
					time.Now().Month().String(),
					time.Now().Day(),
				),
				http.StatusSeeOther,
			)
			return
		}

		posts, err := database.PostsOnThisDay(r.Context(), db, month.Month(), day)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(posts) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var mediaIDs []int
		for i := range posts {
			mediaIDs = append(mediaIDs, posts[i].MediaID)
		}

		medias, err := database.FindMediasByID(r.Context(), db, mediaIDs)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		mediasByID := make(map[int]models.Media)
		for i := range medias {
			mediasByID[medias[i].ID] = medias[i]
		}

		var locationIDs []int
		for i := range posts {
			locationIDs = append(locationIDs, posts[i].LocationID)
		}

		locations, err := database.FindLocationsByID(r.Context(), db, locationIDs)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		locationsByID := make(map[int]models.Location)
		for _, l := range locations {
			locationsByID[l.ID] = l
		}

		ctx := plush.NewContext()
		ctx.Set("posts", posts)
		ctx.Set("locations", locationsByID)
		ctx.Set("medias", mediasByID)
		ctx.Set("month", month.Month())
		ctx.Set("day", day)

		err = renderer(ctx, onThisDayTemplate, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}
}

func BuildRandomHandler(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		postID, err := database.RandomPostID(r.Context(), db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/posts/%d", postID), http.StatusSeeOther)
	}
}
