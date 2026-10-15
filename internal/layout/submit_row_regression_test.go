package layout

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/cellmetrics"
	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/stretchr/testify/require"
)

// google.com puts its two submit buttons side by side in one centered
// container. Both are inputs and must share a row.
func TestSubmitButtonsShareOneRow(t *testing.T) {
	doc := dom.Parse(`<body><div><input type="submit" value="Поиск в Google"><input type="submit" value="Мне повезёт!"></div></body>`)
	box := Build(doc, css.Compute(doc), 200*cellmetrics.WidthPx)
	buttons := findAllTags(box, "input")
	require.Len(t, buttons, 2)
	require.Equal(t, buttons[0].Content.Y, buttons[1].Content.Y, "buttons must share one row")
}
