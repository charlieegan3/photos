package admin

import (
	_ "embed"
	"fmt"
	"net/http"

	"github.com/charlieegan3/photos/cms/internal/pkg/server/templating"
	"github.com/gobuffalo/plush"
)

//go:embed templates/admin/index.html.plush
var adminIndexTemplate string

func BuildAdminIndexHandler(bucketWebURL string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-a")

		ctx := plush.NewContext()

		s, err := plush.Render(adminIndexTemplate, ctx)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		body, err := templating.RenderPage(s, bucketWebURL)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		fmt.Fprintf(w, body)
	}
}
