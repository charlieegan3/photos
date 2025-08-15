package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/charlieegan3/photos/internal/pkg/database"
)

func BuildRedirectHandler(path string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, path, http.StatusSeeOther)
	}
}

func BuildMediaRedirectHelperHandler(db *sql.DB) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		postIDStr := vars["postID"]

		postID, err := strconv.Atoi(postIDStr)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		posts, err := database.FindPostsByID(r.Context(), db, []int{postID})
		if err != nil || len(posts) == 0 {
			http.NotFound(w, r)
			return
		}

		post := posts[0]
		redirectURL := fmt.Sprintf("/admin/medias/%d", post.MediaID)
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
	}
}
