package layout

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/cellmetrics"
	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/stretchr/testify/require"
)

// The browser footer has two link groups on one row: the left group starts at
// the left edge and the right group ends at the right edge of the page.
func TestFooterLeftAndRightGroupsShareOneRow(t *testing.T) {
	doc := dom.Parse(`<body><div style="display:flex;justify-content:space-between"><div><a href="/a">Всё о Google</a> <a href="/b">Реклама</a></div><div><a href="/c">Конфиденциальность</a> <a href="/d">Условия</a></div></div></body>`)
	box := Build(doc, css.Compute(doc), 200*cellmetrics.WidthPx)
	links := findAllTags(box, "a")
	require.Len(t, links, 4)
	for _, l := range links[1:] {
		require.Equal(t, links[0].Content.Y, l.Content.Y, "footer links must share one row")
	}
	right := links[3]
	require.InDelta(t, 200*cellmetrics.WidthPx, right.Content.X+right.Content.Width, cellmetrics.WidthPx,
		"right group must end at the right edge")
}
