package admin

import (
	_ "embed"
	"net/http"

	"github.com/charlieegan3/photos/internal/pkg/server/templating"
	"github.com/gobuffalo/plush"
)

//go:embed templates/admin/index.html.plush
var adminIndexTemplate string

func BuildAdminIndexHandler(renderer templating.PageRenderer) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		ctx := plush.NewContext()
		err := renderer(ctx, adminIndexTemplate, w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
	}
}
