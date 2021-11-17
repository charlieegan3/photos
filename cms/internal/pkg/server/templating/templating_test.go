package templating

import (
	"testing"

	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/require"
)

func TestRenderPage(t *testing.T) {
	body := "<p>hello</p>"

	expectedResult := `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <title>Photos</title>
  </head>
  <body>
    <p>hello</p>
  </body>
</html>
`

	result, err := RenderPage(body, "http://...")
	require.NoError(t, err)

	td.Cmp(t, expectedResult, result)
}
