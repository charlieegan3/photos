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
    <div class="center ph2-l pv3-l mw8">
      <p>bar</p>
    </div>
    
    <a class="absolute bottom-0 tc w-100 pa2 db bt b--silver hover-bg-light-gray" href="/menu">Menu</a>
    
  </body>
</html>
`

	b := new(strings.Builder)

	ctx := plush.NewContext()
	ctx.Set("foo", "bar")

	renderFunc := BuildPageRenderFunc(true)

	err := renderFunc(ctx, nestedTemplate, b)
	require.NoError(t, err)

	td.Cmp(t, b.String(), expectedResult)
}
