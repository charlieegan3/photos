package menu

import (
	"database/sql"
	_ "embed"
	"net/http"

	"github.com/charlieegan3/photos/internal/pkg/server/templating"
	"github.com/gobuffalo/plush"
)

//go:embed templates/index.html.plush
var indexTemplate string

func BuildIndexHandler(db *sql.DB, renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		ctx := plush.NewContext()

		err = renderer(ctx, indexTemplate, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	}
}
