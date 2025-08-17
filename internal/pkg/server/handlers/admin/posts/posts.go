package posts

import (
	"database/sql"
	_ "embed"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

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

//go:embed templates/new.html.plush
var newTemplate string

type selectableModel struct {
	Name string
	ID   int
}

func (sm selectableModel) SelectLabel() string {
	return sm.Name
}

func (sm selectableModel) SelectValue() interface{} {
	return sm.ID
}

func uniqueStrings(input []string) []string {
	keys := make(map[string]bool)
	var result []string
	for _, item := range input {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}
	return result
}

func BuildIndexHandler(db *sql.DB, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		posts, err := database.AllPosts(
			r.Context(), db, true,
			database.SelectOptions{SortField: "publish_date", SortDescending: true},
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		ctx := plush.NewContext()
		ctx.Set("posts", posts)

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

		rawID, ok := mux.Vars(r)["postID"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("post ID is required"))
			return
		}

		id, err := strconv.Atoi(rawID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("postID was not integer"))
			return
		}

		posts, err := database.FindPostsByID(r.Context(), db, []int{id})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
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
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		var tagIDs []int
		for _, t := range taggings {
			tagIDs = append(tagIDs, t.TagID)
		}

		tags, err := database.FindTagsByID(r.Context(), db, tagIDs)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		medias, err := database.FindMediasByID(r.Context(), db, []int{posts[0].MediaID})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		locations, err := database.AllLocations(r.Context(), db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		allMedias, err := database.AllMedias(r.Context(), db, true)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		var formMedias []selectableModel
		for i := range allMedias {
			m := &allMedias[i]
			formMedias = append(formMedias, selectableModel{
				Name: fmt.Sprintf("%d - %s", m.ID, m.TakenAt.Format("2006-01-02 15:04:05")),
				ID:   m.ID,
			})
		}

		var formLocations []selectableModel
		for _, l := range locations {
			formLocations = append(formLocations, selectableModel{Name: l.Name, ID: l.ID})
		}

		// Load collections for form
		collectionsRepo := database.NewCollectionRepository(db)
		allCollections, err := collectionsRepo.AllOrderedByPostCount(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		// Load existing post-collection relationships
		postCollectionsRepo := database.NewPostCollectionRepository(db)
		postCollections, err := postCollectionsRepo.FindByPostID(r.Context(), posts[0].ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		// Create map of collection IDs that this post belongs to
		postCollectionIDs := make(map[int]bool)
		for _, pc := range postCollections {
			postCollectionIDs[pc.CollectionID] = true
		}

		ctx := plush.NewContext()
		ctx.Set("post", posts[0])
		ctx.Set("media", medias[0])
		ctx.Set("locations", formLocations)
		ctx.Set("medias", formMedias)
		ctx.Set("tags", tags)
		ctx.Set("collections", allCollections)
		ctx.Set("postCollectionIDs", postCollectionIDs)

		err = renderer(ctx, showTemplate, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
	}
}

func BuildNewHandler(db *sql.DB, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		newPost := models.Post{}

		mediaID := r.URL.Query().Get("mediaID")
		if mediaID != "" {
			i, err := strconv.ParseInt(mediaID, 10, 64)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("failed to parse supplied mediaID value"))
				return
			}
			newPost.MediaID = int(i)
		}

		locationID := r.URL.Query().Get("locationID")
		if locationID != "" {
			i, err := strconv.ParseInt(locationID, 10, 64)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("failed to parse supplied locationID value"))
				return
			}
			newPost.LocationID = int(i)
		}

		timestamp := r.URL.Query().Get("timestamp")
		if timestamp != "" {
			i, err := strconv.ParseInt(timestamp, 10, 64)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("failed to parse supplied timestamp value"))
				return
			}
			newPost.PublishDate = time.Unix(i, 0)
		}

		locations, err := database.AllLocations(r.Context(), db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		medias, err := database.AllMedias(r.Context(), db, true)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		var formMedias []selectableModel
		for i := range medias {
			m := &medias[i]
			formMedias = append(formMedias, selectableModel{
				Name: fmt.Sprintf("%d - %s", m.ID, m.TakenAt.Format("2006-01-02 15:04:05")),
				ID:   m.ID,
			})
		}

		var formLocations []selectableModel
		for _, l := range locations {
			formLocations = append(formLocations, selectableModel{Name: l.Name, ID: l.ID})
		}

		ctx := plush.NewContext()
		ctx.Set("post", newPost)
		ctx.Set("locations", formLocations)
		ctx.Set("medias", formMedias)

		err = renderer(ctx, newTemplate, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
	}
}

func BuildCreateHandler(db *sql.DB, _ templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		if val, ok := r.Header["Content-Type"]; !ok || val[0] != "application/x-www-form-urlencoded" {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Content-Type must be 'multipart/form-data'"))
			return
		}

		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("failed to parse multipart form"))
			return
		}

		post := models.Post{
			Description: r.Form.Get("Description"),
		}
		if r.Form.Get("IsDraft") != "" {
			post.IsDraft = true
		}
		if r.Form.Get("IsFavourite") != "" {
			post.IsFavourite = true
		}
		dateVal := r.PostForm.Get("PublishDate")
		timeVal := r.PostForm.Get("PublishTime")
		if dateVal != "" && timeVal != "" {
			dateTimeStr := dateVal + "T" + timeVal + ":00Z"
			post.PublishDate, err = time.Parse("2006-01-02T15:04:05Z", dateTimeStr)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "date/time values '%v %v' were invalid", dateVal, timeVal)
				return
			}
		}
		post.LocationID, err = strconv.Atoi(r.Form.Get("LocationID"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("failed to parse location ID"))
			return
		}
		post.MediaID, err = strconv.Atoi(r.Form.Get("MediaID"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("failed to parse media ID"))
			return
		}

		persistedPosts, err := database.CreatePosts(r.Context(), db, []models.Post{post})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		if len(persistedPosts) != 1 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("unexpected number of persistedPosts"))
			return
		}

		tags := uniqueStrings(strings.Fields(strings.ToLower(r.Form.Get("Tags"))))
		err = database.SetPostTags(r.Context(), db, persistedPosts[0], tags)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/admin/posts/%d", persistedPosts[0].ID), http.StatusSeeOther)
	}
}

func BuildFormHandler(db *sql.DB, _ templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		contentType, ok := r.Header["Content-Type"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Content-Type must be set"))
			return
		}

		rawID, ok := mux.Vars(r)["postID"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("post slug is required"))
			return
		}

		id, err := strconv.Atoi(rawID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("failed to parse post ID"))
			return
		}

		existingPosts, err := database.FindPostsByID(r.Context(), db, []int{id})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		if len(existingPosts) == 0 {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// handle delete
		if contentType[0] != "application/x-www-form-urlencoded" {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Content-Type must be 'application/x-www-form-urlencoded'"))
			return
		}

		err = r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("failed to parse delete form"))
			return
		}

		if r.Form.Get("_method") == http.MethodDelete {
			err = database.DeletePosts(r.Context(), db, []models.Post{existingPosts[0]})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
			http.Redirect(w, r, "/admin/posts", http.StatusSeeOther)
			return
		}

		if r.PostForm.Get("_method") != http.MethodPut {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("expected _method to be PUT or DELETE in form"))
			return
		}

		post := models.Post{
			ID:          existingPosts[0].ID,
			Description: r.Form.Get("Description"),
			PublishDate: existingPosts[0].PublishDate, // Preserve existing PublishDate
		}
		if r.Form.Get("IsDraft") != "" {
			post.IsDraft = true
		}
		if r.Form.Get("IsFavourite") != "" {
			post.IsFavourite = true
		}
		dateVal := r.PostForm.Get("PublishDate")
		timeVal := r.PostForm.Get("PublishTime")
		if dateVal != "" && timeVal != "" {
			dateTimeStr := dateVal + "T" + timeVal + ":00Z"
			post.PublishDate, err = time.Parse("2006-01-02T15:04:05Z", dateTimeStr)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(w, "date/time values '%v %v' were invalid", dateVal, timeVal)
				return
			}
		}
		post.LocationID, err = strconv.Atoi(r.Form.Get("LocationID"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("failed to parse location ID"))
			return
		}
		post.MediaID, err = strconv.Atoi(r.Form.Get("MediaID"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("failed to parse media ID"))
			return
		}

		updatedPosts, err := database.UpdatePosts(r.Context(), db, []models.Post{post})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		if len(updatedPosts) != 1 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("unexpected number of updatedPosts"))
			return
		}

		tags := uniqueStrings(strings.Fields(strings.ToLower(r.Form.Get("Tags"))))
		err = database.SetPostTags(r.Context(), db, updatedPosts[0], tags)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		// Handle collections
		collectionValues := r.Form["Collections"]
		var selectedCollectionIDs []int
		for _, collectionIDStr := range collectionValues {
			collectionID, err := strconv.Atoi(collectionIDStr)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("failed to parse collection ID"))
				return
			}
			selectedCollectionIDs = append(selectedCollectionIDs, collectionID)
		}

		// Remove existing post-collection relationships
		postCollectionsRepo := database.NewPostCollectionRepository(db)
		err = postCollectionsRepo.DeleteByPostID(r.Context(), updatedPosts[0].ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		// Create new post-collection relationships
		if len(selectedCollectionIDs) > 0 {
			var postCollections []models.PostCollection
			for _, collectionID := range selectedCollectionIDs {
				postCollections = append(postCollections, models.PostCollection{
					PostID:       updatedPosts[0].ID,
					CollectionID: collectionID,
				})
			}

			_, err = postCollectionsRepo.CreateWithConflictHandling(r.Context(), postCollections)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(err.Error()))
				return
			}
		}

		http.Redirect(
			w,
			r,
			fmt.Sprintf("/admin/posts/%d", updatedPosts[0].ID),
			http.StatusSeeOther,
		)
	}
}
