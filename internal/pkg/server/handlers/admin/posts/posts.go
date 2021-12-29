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

	"github.com/charlieegan3/photos/cms/internal/pkg/database"
	"github.com/charlieegan3/photos/cms/internal/pkg/models"
	"github.com/charlieegan3/photos/cms/internal/pkg/server/templating"
)

//go:embed templates/index.html.plush
var indexTemplate string

//go:embed templates/show.html.plush
var showTemplate string

//go:embed templates/new.html.plush
var newTemplate string

type SelectableModel struct {
	Name string
	ID   int
}

func (sm SelectableModel) SelectLabel() string {
	return sm.Name
}

func (sm SelectableModel) SelectValue() interface{} {
	return sm.ID
}

func BuildIndexHandler(db *sql.DB, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		posts, err := database.AllPosts(db, true, database.SelectOptions{SortField: "publish_date", SortDescending: true})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		ctx := plush.NewContext()
		ctx.Set("posts", posts)

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

		medias, err := database.FindMediasByID(db, []int{posts[0].MediaID})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		locations, err := database.AllLocations(db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		allMedias, err := database.AllMedias(db, true)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		mediaMap := make(map[string]interface{})
		for _, m := range allMedias {
			mediaMap[fmt.Sprintf("%d-%s", m.ID, m.TakenAt)] = m.ID
		}

		var formLocations []SelectableModel
		for _, l := range locations {
			formLocations = append(formLocations, SelectableModel{Name: l.Name, ID: l.ID})
		}

		ctx := plush.NewContext()
		ctx.Set("post", posts[0])
		ctx.Set("media", medias[0])
		ctx.Set("locations", formLocations)
		ctx.Set("medias", mediaMap)
		ctx.Set("tags", tags)

		err = renderer(ctx, showTemplate, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
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
				w.Write([]byte("failed to parse supplied mediaID value"))
				return
			}
			newPost.MediaID = int(i)
		}

		timestamp := r.URL.Query().Get("timestamp")
		if timestamp != "" {
			i, err := strconv.ParseInt(timestamp, 10, 64)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed to parse supplied timestamp value"))
				return
			}
			newPost.PublishDate = time.Unix(i, 0)
		}

		locations, err := database.AllLocations(db)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		medias, err := database.AllMedias(db, true)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		mediaMap := make(map[string]interface{})
		for _, m := range medias {
			mediaMap[fmt.Sprintf("%d-%s", m.ID, m.TakenAt)] = m.ID
		}

		var formLocations []SelectableModel
		for _, l := range locations {
			formLocations = append(formLocations, SelectableModel{Name: l.Name, ID: l.ID})
		}

		ctx := plush.NewContext()
		ctx.Set("post", newPost)
		ctx.Set("locations", formLocations)
		ctx.Set("medias", mediaMap)

		err = renderer(ctx, newTemplate, w)
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

		post := models.Post{
			Description: r.Form.Get("Description"),
		}
		if r.Form.Get("IsDraft") != "" {
			post.IsDraft = true
		}
		if val := r.PostForm.Get("PublishDate"); val != "" {
			post.PublishDate, err = time.Parse("2006-01-02T15:04", val)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("time value '%v' was invalid", val)))
				return
			}
		}
		post.LocationID, err = strconv.Atoi(r.Form.Get("LocationID"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("failed to parse location ID"))
			return
		}
		post.MediaID, err = strconv.Atoi(r.Form.Get("MediaID"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("failed to parse media ID"))
			return
		}

		persistedPosts, err := database.CreatePosts(db, []models.Post{post})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(persistedPosts) != 1 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unexpected number of persistedPosts"))
			return
		}

		tags := strings.Fields(strings.ToLower(r.Form.Get("Tags")))
		err = database.SetPostTags(db, persistedPosts[0], tags)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/admin/posts/%d", persistedPosts[0].ID), http.StatusSeeOther)
	}
}

func BuildFormHandler(db *sql.DB, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		contentType, ok := r.Header["Content-Type"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Content-Type must be set"))
			return
		}

		rawID, ok := mux.Vars(r)["postID"]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("post slug is required"))
			return
		}

		id, err := strconv.Atoi(rawID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to parse post ID"))
			return
		}

		existingPosts, err := database.FindPostsByID(db, id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(existingPosts) == 0 {
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

		if r.Form.Get("_method") == "DELETE" {
			err = database.DeletePosts(db, []models.Post{existingPosts[0]})
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			http.Redirect(w, r, "/admin/posts", http.StatusSeeOther)
			return
		}

		if r.PostForm.Get("_method") != "PUT" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("expected _method to be PUT or DELETE in form"))
			return
		}

		post := models.Post{
			ID:          existingPosts[0].ID,
			Description: r.Form.Get("Description"),
		}
		if r.Form.Get("IsDraft") != "" {
			post.IsDraft = true
		}
		if val := r.PostForm.Get("PublishDate"); val != "" {
			post.PublishDate, err = time.Parse("2006-01-02T15:04", val)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(fmt.Sprintf("time value '%v' was invalid", val)))
				return
			}
		}
		post.LocationID, err = strconv.Atoi(r.Form.Get("LocationID"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("failed to parse location ID"))
			return
		}
		post.MediaID, err = strconv.Atoi(r.Form.Get("MediaID"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("failed to parse media ID"))
			return
		}

		updatedPosts, err := database.UpdatePosts(db, []models.Post{post})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if len(updatedPosts) != 1 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("unexpected number of updatedPosts"))
			return
		}

		tags := strings.Fields(strings.ToLower(r.Form.Get("Tags")))
		err = database.SetPostTags(db, updatedPosts[0], tags)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		http.Redirect(
			w,
			r,
			fmt.Sprintf("/admin/posts/%d", updatedPosts[0].ID),
			http.StatusSeeOther,
		)
	}
}
