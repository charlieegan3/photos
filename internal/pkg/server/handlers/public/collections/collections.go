package public

import (
	"database/sql"
	_ "embed"
	"net/http"
	"strconv"

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
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")

		collectionsRepo := database.NewCollectionRepository(db)
		collections, err := collectionsRepo.AllOrderedByPostCount(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		if len(collections) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		ctx := plush.NewContext()
		ctx.Set("collections", collections)

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

		rawID, ok := mux.Vars(r)["collectionID"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("collection id is required"))
			return
		}

		id, err := strconv.Atoi(rawID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("failed to parse collection ID"))
			return
		}

		collectionsRepo := database.NewCollectionRepository(db)
		collections, err := collectionsRepo.FindByIDs(r.Context(), []int64{int64(id)})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		if len(collections) != 1 {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("collection not found"))
			return
		}

		collection := collections[0]

		// Get posts in this collection (ordered by newest first)
		posts, err := collectionsRepo.Posts(r.Context(), id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		if len(posts) == 0 {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("no posts found in collection"))
			return
		}

		// Load locations for posts
		var locationIDs []int
		for i := range posts {
			locationIDs = append(locationIDs, posts[i].LocationID)
		}

		locations, err := database.FindLocationsByID(r.Context(), db, locationIDs)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		locationsByID := make(map[int]models.Location)
		for _, l := range locations {
			locationsByID[l.ID] = l
		}

		// Load medias for posts
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
		ctx.Set("collection", collection)
		ctx.Set("posts", posts)
		ctx.Set("locations", locationsByID)
		ctx.Set("medias", mediasByID)

		err = renderer(ctx, showTemplate, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
	}
}
