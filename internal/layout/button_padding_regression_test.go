package layout

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/cellmetrics"
	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/stretchr/testify/require"
)

func TestButtonWithPaddingKeepsLabelOnOneLine(t *testing.T) {
	doc := dom.Parse(`<body><input type="submit" value="Поиск в Google" style="padding:0 15px"></body>`)
	box := Build(doc, css.Compute(doc), 200*cellmetrics.WidthPx)
	btn := findTag(box, "input")
	require.NotNil(t, btn)
	require.Len(t, btn.TextLines, 1, "label wrapped: %+v", btn.TextLines)
}
