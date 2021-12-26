package public

import (
	"database/sql"
	_ "embed"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gobuffalo/plush"
	"github.com/gorilla/mux"

	"github.com/charlieegan3/photos/cms/internal/pkg/database"
	"github.com/charlieegan3/photos/cms/internal/pkg/server/templating"
)

//go:embed templates/index.html.plush
var indexTemplate string

//go:embed templates/show.html.plush
var showTemplate string

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
			db,
			false,
			database.SelectOptions{},
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		lastPage := postsCount/int64(pageSize) + 1
		if int64(page) > lastPage {
			http.Redirect(w, r, fmt.Sprintf("/?page=%d", lastPage), http.StatusSeeOther)
			return
		}

		ctx := plush.NewContext()
		ctx.Set("posts", posts)
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

		posts, err := database.FindPostsByID(db, id)
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

		tags, err := database.FindTagsByID(db, tagIDs)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		medias, err := database.FindMediasByID(db, posts[0].MediaID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		if len(medias) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		locations, err := database.FindLocationsByID(db, posts[0].LocationID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		if len(medias) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		devices, err := database.FindDevicesByID(db, medias[0].DeviceID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		if len(medias) == 0 {
			w.WriteHeader(http.StatusNotFound)
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
		ctx.Set("location", locations[0])
		ctx.Set("tags", tags)

		err = renderer(ctx, showTemplate, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}
}
