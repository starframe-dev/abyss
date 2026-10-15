package layout

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/cellmetrics"
	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/stretchr/testify/require"
)

// google.com's second button wrapper holds a <script> and a hidden <input>
// beside the button. Those add no box, so the wrapper still counts as a button.
func TestButtonWrapperIgnoresScriptAndHiddenInput(t *testing.T) {
	doc := dom.Parse(`<body><table><tr><td align="center">` +
		`<span style="display:inline-block"><span style="display:block"><input type="submit" value="Поиск в Google"></span></span>` +
		`<span style="display:inline-block"><span style="display:block"><input type="submit" value="Мне повезёт!"><script>var x=1;</script><input type="hidden" name="h" value="1"></span></span>` +
		`</td></tr></table></body>`)
	box := Build(doc, css.Compute(doc), 200*cellmetrics.WidthPx)
	buttons := findAllTags(box, "input")
	visible := []*LayoutBox{}
	for _, b := range buttons {
		if b.Node.GetAttr("type") == "submit" {
			visible = append(visible, b)
		}
	}
	require.Len(t, visible, 2)
	require.Equal(t, visible[0].Content.Y, visible[1].Content.Y, "buttons with script or hidden input beside them must share one row")
}
