package layout

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/cellmetrics"
	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/stretchr/testify/require"
)

// The footer reads "Конфиденциальность - Условия": the spaces around the dash
// are part of the text, so the dash must sit one cell away from each link.
func TestSpaceAfterLinkKeepsGapBeforeDash(t *testing.T) {
	doc := dom.Parse(`<body><p><a href="/a">Конфиденциальность</a> - <a href="/b">Условия</a></p></body>`)
	box := Build(doc, css.Compute(doc), 200*cellmetrics.WidthPx)
	links := findAllTags(box, "a")
	require.Len(t, links, 2)
	var dash *LayoutBox
	for _, c := range findTag(box, "p").Children {
		if c.Node != nil && c.Node.Type == dom.TextNode {
			dash = c
		}
	}
	require.NotNil(t, dash, "the text between links must be laid out")
	first := links[0]
	gap := dash.Content.X - (first.Content.X + first.Content.Width)
	require.GreaterOrEqual(t, gap, float64(cellmetrics.WidthPx), "a space must separate the first link from the dash")
}
