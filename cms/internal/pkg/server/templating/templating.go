package templating

import (
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"net/url"
	"strings"

	"github.com/gobuffalo/plush"
	"github.com/pkg/errors"
)

//go:embed base.html.plush
var baseTemplate string

type PageRenderer func(*plush.Context, string, io.Writer) error

func BuildPageRenderFunc(bucketWebURL string, geoapifyAPIKey string) PageRenderer {
	return func(ctx *plush.Context, t string, w io.Writer) error {
		// make the media_url helper function available to supplied nested
		// templates
		ctx.Set("media_url", func(s ...string) string {
			return fmt.Sprintf("%s%s", bucketWebURL, strings.Join(s, ""))
		})

		ctx.Set("map", func(latitude, longitude float64) (template.HTML, error) {
			mapURL, err := url.Parse("https://maps.geoapify.com/v1/staticmap")
			if err != nil {
				return "", errors.Wrap(err, "failed to parse mapURL")
			}
			values := url.Values{
				"style":       []string{"osm-bright-smooth"},
				"center":      []string{fmt.Sprintf("lonlat:%f,%f", longitude, latitude)},
				"zoom":        []string{"10.3497"},
				"width":       []string{"400"},
				"height":      []string{"400"},
				"scaleFactor": []string{"2"},
				"marker":      []string{fmt.Sprintf("lonlat:%f,%f;type:awesome;color:#e01401", longitude, latitude)},
				"apiKey":      []string{geoapifyAPIKey},
			}
			mapURL.RawQuery = values.Encode()

			return template.HTML(fmt.Sprintf(`<img loading="lazy" class="w-100" src="%v"/>`, mapURL.String())), nil
		})

		ctx.Set("to_string", func(arg interface{}) string {
			return fmt.Sprintf("%v", arg)
		})

		ctx.Set("truncate", func(s string, length int) string {
			if len(s) < length {
				return s
			}
			return s[:length] + "..."
		})

		body, err := plush.Render(t, ctx)
		if err != nil {
			return errors.Wrap(err, "failed to evaluate provided template")
		}

		tmpl, err := template.New("base").Parse(baseTemplate)
		if err != nil {
			return errors.Wrap(err, "failed to parse base template")
		}

		return tmpl.Execute(w, struct{ Body template.HTML }{Body: template.HTML(body)})
	}
}
