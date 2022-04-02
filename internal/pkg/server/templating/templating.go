package templating

import (
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"strings"

	"github.com/gobuffalo/plush"
	"github.com/gomarkdown/markdown"
	"github.com/pkg/errors"
)

//go:embed base.html
var baseTemplate string

//go:embed base.admin.html
var baseTemplateAdmin string

type PageRenderer func(*plush.Context, string, io.Writer) error

func BuildPageRenderFunc(showMenu bool, headContent string, intermediateTemplates ...string) PageRenderer {
	// list of all templates to run including intermediateTemplates
	templates := append(intermediateTemplates, "base")

	return func(ctx *plush.Context, t string, w io.Writer) error {
		ctx.Set("to_string", func(arg interface{}) string {
			return fmt.Sprintf("%v", arg)
		})

		ctx.Set("markdown", func(md string) string {
			d := markdown.NormalizeNewlines([]byte(md))
			return string(markdown.ToHTML(d, nil, nil))
		})

		ctx.Set("truncate", func(s string, length int, elipsis bool) string {
			if len(s) < length {
				return s
			}
			if elipsis {
				return s[:length] + "..."
			}
			return s[:length]
		})

		body, err := plush.Render(t, ctx)
		if err != nil {
			return errors.Wrap(err, "failed to evaluate provided template")
		}

		for _, chainTemplate := range templates {
			templateContent := ""
			switch chainTemplate {
			case "base":
				templateContent = baseTemplate
			case "admin":
				templateContent = baseTemplateAdmin
			default:
				return fmt.Errorf("unknown template: %s", chainTemplate)
			}

			tmpl, err := template.New("base").Parse(templateContent)
			if err != nil {
				return errors.Wrap(err, "failed to parse base template")
			}

			var bodyBuilder strings.Builder
			err = tmpl.Execute(&bodyBuilder, struct {
				ShowMenu    bool
				HeadContent template.HTML
				Body        template.HTML
			}{
				ShowMenu:    showMenu,
				HeadContent: template.HTML(headContent),
				Body:        template.HTML(body),
			})
			if err != nil {
				return errors.Wrap(err, "failed to parse base template")
			}

			body = bodyBuilder.String()
		}

		_, err = io.Copy(w, strings.NewReader(body))
		return err
	}
}
