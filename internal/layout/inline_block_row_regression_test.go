package layout

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/cellmetrics"
	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/stretchr/testify/require"
)

// google.com wraps each submit button in <span style="display:inline-block">
// around a block <span> that also holds a script. The wrappers must share one row.
func TestInlineBlockButtonWrappersShareOneRow(t *testing.T) {
	doc := dom.Parse(`<body><div>` +
		`<span style="display:inline-block"><span style="display:block"><input type="submit" value="Поиск в Google"><script>var x=1;</script></span></span>` +
		`<span style="display:inline-block"><span style="display:block"><input type="submit" value="Мне повезёт!"><script>var y=2;</script></span></span>` +
		`</div></body>`)
	box := Build(doc, css.Compute(doc), 200*cellmetrics.WidthPx)
	buttons := findAllTags(box, "input")
	require.Len(t, buttons, 2)
	require.Equal(t, buttons[0].Content.Y, buttons[1].Content.Y, "inline-block buttons must share one row")
}

// A centered table cell must center its button row inside the cell.
func TestButtonRowInCenteredCellIsCentered(t *testing.T) {
	doc := dom.Parse(`<body><table style="width:200px"><tr><td align="center">` +
		`<span style="display:inline-block"><span style="display:block"><input type="submit" value="Поиск в Google"></span></span>` +
		`<span style="display:inline-block"><span style="display:block"><input type="submit" value="Мне повезёт!"></span></span>` +
		`</td></tr></table></body>`)
	box := Build(doc, css.Compute(doc), 200*cellmetrics.WidthPx)
	buttons := findAllTags(box, "input")
	require.Len(t, buttons, 2)
	cell := findTag(box, "td")
	require.NotNil(t, cell)
	left := buttons[0].Content.X - cell.Content.X
	right := (cell.Content.X + cell.Content.Width) - (buttons[1].Content.X + buttons[1].Content.Width)
	require.Greater(t, left, 0.0, "row must not hug the left edge")
	require.InDelta(t, left, right, cellmetrics.WidthPx, "row must be centered in the cell")
}

// google.com keeps the buttons in a table cell next to a <br>. A run of
// consecutive buttons must still share a row when other children are present.
func TestButtonRunInTableCellWithBreakSharesRow(t *testing.T) {
	doc := dom.Parse(`<body><table><tr><td align="center"><br>` +
		`<span style="display:inline-block"><span style="display:block"><input type="submit" value="Поиск в Google"></span></span>` +
		`<span style="display:inline-block"><span style="display:block"><input type="submit" value="Мне повезёт!"></span></span>` +
		`</td></tr></table></body>`)
	box := Build(doc, css.Compute(doc), 200*cellmetrics.WidthPx)
	buttons := findAllTags(box, "input")
	require.Len(t, buttons, 2)
	require.Equal(t, buttons[0].Content.Y, buttons[1].Content.Y, "buttons in a table cell must share one row")
}
