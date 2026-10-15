package layout

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/cellmetrics"
	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/stretchr/testify/require"
)

// google.com's empty search field has a title tooltip. A browser shows the
// title only on hover, so it must not appear as text inside the field.
func TestEmptyInputTitleIsNotDrawnAsText(t *testing.T) {
	doc := dom.Parse(`<body><input name="q" title="Поиск в Google" value=""></body>`)
	box := Build(doc, css.Compute(doc), 200*cellmetrics.WidthPx)
	field := findTag(box, "input")
	require.NotNil(t, field)
	require.Empty(t, field.TextLines, "title must not be drawn as the field text")
}
