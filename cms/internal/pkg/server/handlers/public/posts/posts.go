package public

import (
	"database/sql"
	_ "embed"
	"fmt"
	"net/http"
	"strconv"

	"github.com/charlieegan3/photos/cms/internal/pkg/database"
	"github.com/charlieegan3/photos/cms/internal/pkg/server/templating"
	"github.com/gobuffalo/plush"
)

//go:embed templates/index.html.plush
var indexTemplate string

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
