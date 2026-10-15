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
