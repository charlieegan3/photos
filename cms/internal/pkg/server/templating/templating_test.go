package templating

import (
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
  </head>
  <body>
    <p>bar</p>
  </body>
</html>
`

	ctx := plush.NewContext()
	ctx.Set("foo", "bar")

	result, err := RenderPage(ctx, nestedTemplate, "http://...")
	require.NoError(t, err)

	td.Cmp(t, expectedResult, result)
}
