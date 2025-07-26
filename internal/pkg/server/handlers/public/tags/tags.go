package public

import (
	"database/sql"
	_ "embed"
	"net/http"

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

func BuildIndexHandler(db *sql.DB, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		tags, err := database.AllTags(
			r.Context(),
			db,
			false,
			database.SelectOptions{
				SortField: "name",
			},
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		ctx := plush.NewContext()
		ctx.Set("tags", tags)

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

		tagName, ok := mux.Vars(r)["tagName"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("tagName is required"))
			return
		}

		tags, err := database.FindTagsByName(r.Context(), db, []string{tagName})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		if len(tags) == 0 {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("not found"))
			return
		}

		// TODO create a db function to get tags for post in SQL
		taggings, err := database.FindTaggingsByTagID(r.Context(), db, tags[0].ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		if len(taggings) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		var postIDs []int
		for _, t := range taggings {
			postIDs = append(postIDs, t.PostID)
		}

		posts, err := database.FindPostsByID(r.Context(), db, postIDs)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		var mediaIDs []int
		for i := range posts {
			mediaIDs = append(mediaIDs, posts[i].MediaID)
		}

		medias, err := database.FindMediasByID(r.Context(), db, mediaIDs)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		mediasByID := make(map[int]models.Media)
		for i := range medias {
			mediasByID[medias[i].ID] = medias[i]
		}

		ctx := plush.NewContext()
		ctx.Set("tagName", tagName)
		ctx.Set("medias", mediasByID)
		ctx.Set("posts", posts)

		err = renderer(ctx, showTemplate, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
	}
}
