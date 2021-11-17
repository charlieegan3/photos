package templating

import (
	_ "embed"

	"github.com/gobuffalo/plush"
	"github.com/pkg/errors"
)

//go:embed base.html.plush
var baseTemplate string

func RenderPage(ctx *plush.Context, template string, bucketWebURL string) (string, error) {
	body, err := plush.Render(template, ctx)
	if err != nil {
		return "", errors.Wrap(err, "failed to evaluate provided template")
	}

	ctx = plush.NewContext()
	ctx.Set("body", body)

	return plush.Render(baseTemplate, ctx)
}
