package templating

import (
	"strings"
	"testing"

	"github.com/gobuffalo/plush"
	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/require"
)

func TestRenderPage(t *testing.T) {
	nestedTemplate := "<p><%= foo %></p>"

	expectedResult := `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <title>Photos</title>
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <link rel="stylesheet" href="/styles.css">
  </head>
  <body>
    <div class="pa2"><p>bar</p></div>
  </body>
</html>
`

	b := new(strings.Builder)

	ctx := plush.NewContext()
	ctx.Set("foo", "bar")

	renderFunc := BuildPageRenderFunc("http://", "")

	err := renderFunc(ctx, nestedTemplate, b)
	require.NoError(t, err)

	td.Cmp(t, expectedResult, b.String())
}
