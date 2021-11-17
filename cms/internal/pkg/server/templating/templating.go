package templating

import (
	_ "embed"

	"github.com/gobuffalo/plush"
)

//go:embed base.html.plush
var baseTemplate string

func RenderPage(body string, bucketWebURL string) (string, error) {
	ctx := plush.NewContext()
	ctx.Set("body", body)

	return plush.Render(baseTemplate, ctx)
}
