package layout

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/cellmetrics"
	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/stretchr/testify/require"
)

// google.com wraps each submit button in a <span class="lsbb">. The pair must
// still share one row when the wrappers are the only children.
func TestWrappedSubmitButtonsShareOneRow(t *testing.T) {
	doc := dom.Parse(`<body><div><span><input type="submit" value="Поиск в Google"></span><span><input type="submit" value="Мне повезёт!"></span></div></body>`)
	box := Build(doc, css.Compute(doc), 200*cellmetrics.WidthPx)
	buttons := findAllTags(box, "input")
	require.Len(t, buttons, 2)
	require.Equal(t, buttons[0].Content.Y, buttons[1].Content.Y, "wrapped buttons must share one row")
}
