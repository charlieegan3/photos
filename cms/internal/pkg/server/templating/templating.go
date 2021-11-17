package templating

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/gobuffalo/plush"
	"github.com/pkg/errors"
)

//go:embed base.html.plush
var baseTemplate string

type PageRenderer func(*plush.Context, string) (string, error)

func BuildPageRenderFunc(bucketWebURL string) PageRenderer {
	return func(ctx *plush.Context, template string) (string, error) {
		// make the image_url helper function available to supplied nested
		// templates
		ctx.Set("image_url", func(s ...string) string {
			return fmt.Sprintf("%s%s", bucketWebURL, strings.Join(s, ""))
		})

		body, err := plush.Render(template, ctx)
		if err != nil {
			return "", errors.Wrap(err, "failed to evaluate provided template")
		}

		ctx = plush.NewContext()
		ctx.Set("body", body)

		return plush.Render(baseTemplate, ctx)
	}
}
