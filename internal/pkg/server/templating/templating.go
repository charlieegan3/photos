package templating

import (
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"math"
	"strings"
	"time"

	"github.com/gobuffalo/plush"
	"github.com/gomarkdown/markdown"
	"github.com/pkg/errors"

	"github.com/charlieegan3/photos/internal/pkg/models"
)

//go:embed base.html
var baseTemplate string

//go:embed base.admin.html
var baseTemplateAdmin string

type PageRenderer func(*plush.Context, string, io.Writer) error

func BuildPageRenderFunc(showMenu bool, headContent string, intermediateTemplates ...string) PageRenderer {
	// list of all templates to run including intermediateTemplates
	templates := intermediateTemplates
	templates = append(templates, "base")

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

		ctx.Set("display_offset", func(media models.Media) string {
			x := 50
			y := 50

			isRotated := media.Orientation == 6 || media.Orientation == 8

			if isRotated {
				if media.Width < media.Height {
					x = media.DisplayOffset
					y = 0
				} else if media.Width > media.Height {
					y = media.DisplayOffset
					x = 0
				}
			} else {
				if media.Width > media.Height {
					x = media.DisplayOffset
					y = 0
				} else if media.Width < media.Height {
					y = media.DisplayOffset
					x = 0
				}
			}
			return fmt.Sprintf("%d%% %d%%", x, y)
		})

		ctx.Set("days_diff", func(t1, t2 time.Time) string {
			t1 = t1.Truncate(time.Hour * 24)
			t2 = t2.Truncate(time.Hour * 24)

			return fmt.Sprintf("%v days", math.Abs(math.Ceil(t2.Sub(t1).Hours()/24)))
		})

		ctx.Set("format_fnumber", func(f float64) string {
			return fmt.Sprintf("%.1f", f)
		})

		body, err := plush.Render(t, ctx)
		if err != nil {
			return errors.Wrap(err, "failed to evaluate provided template")
		}

		for _, chainTemplate := range templates {
			var templateContent string
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
				ShowMenu: showMenu,

				// comes from trusted author
				//nolint:gosec
				HeadContent: template.HTML(headContent),
				//nolint:gosec
				Body: template.HTML(body),
			})
			if err != nil {
				return errors.Wrap(err, "failed to parse base template")
			}

			body = bodyBuilder.String()
		}

		_, err = io.Copy(w, strings.NewReader(body))
		if err != nil {
			return fmt.Errorf("failed to write rendered template to writer: %w", err)
		}

		return nil
	}
}
