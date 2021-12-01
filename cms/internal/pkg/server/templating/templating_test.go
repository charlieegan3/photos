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
    <link rel="stylesheet" href="/styles.css">
  </head>
  <body>
    <p>bar</p>
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
