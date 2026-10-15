package layout

import (
	"math"
	"testing"

	"github.com/starframe-dev/abyss/internal/cellmetrics"
	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/stretchr/testify/require"
)

// google.com's footer is a centered div holding several text links. A browser
// sets consecutive links on one line; each link must not take its own row.
func TestConsecutiveTextLinksShareOneRow(t *testing.T) {
	doc := dom.Parse(`<body><div style="text-align:center"><a href="/a">Реклама</a><a href="/b">Всё о Google</a></div></body>`)
	box := Build(doc, css.Compute(doc), 200*cellmetrics.WidthPx)
	links := findAllTags(box, "a")
	require.Len(t, links, 2)
	require.Equal(t, links[0].Content.Y, links[1].Content.Y, "consecutive links must share one row")
	gap := links[1].Content.X - (links[0].Content.X + links[0].Content.Width)
	require.GreaterOrEqual(t, gap, 2.0*cellmetrics.WidthPx, "adjacent links need a gap")
}

// A paragraph's text between links is part of the same line as the links.
func TestTextBetweenLinksStaysOnOneLine(t *testing.T) {
	doc := dom.Parse(`<body><p>© 2026 - <a href="/a">Конфиденциальность</a> - <a href="/b">Условия</a></p></body>`)
	box := Build(doc, css.Compute(doc), 200*cellmetrics.WidthPx)
	links := findAllTags(box, "a")
	require.Len(t, links, 2)
	require.Equal(t, links[0].Content.Y, links[1].Content.Y, "links in one paragraph must share one line")
	copyright := findTag(box, "p")
	require.NotNil(t, copyright)
	require.Equal(t, copyright.Content.Y, links[0].Content.Y, "paragraph text must share the link line")
}

// text-align:center centers a row of buttons: the gap on each side is equal.
func TestCenteredButtonRowHasEqualSideGaps(t *testing.T) {
	doc := dom.Parse(`<body><div style="text-align:center"><input type="submit" value="Поиск в Google"><input type="submit" value="Мне повезёт!"></div></body>`)
	box := Build(doc, css.Compute(doc), 200*cellmetrics.WidthPx)
	div := findTag(box, "div")
	buttons := findAllTags(box, "input")
	require.Len(t, buttons, 2)
	require.Equal(t, buttons[0].Content.Y, buttons[1].Content.Y, "buttons must share one row")
	left := buttons[0].Content.X - div.Content.X
	last := buttons[1]
	right := (div.Content.X + div.Content.Width) - (last.Content.X + last.Content.Width)
	require.True(t, math.Abs(left-right) <= cellmetrics.WidthPx, "row not centered: left=%v right=%v", left, right)
}
