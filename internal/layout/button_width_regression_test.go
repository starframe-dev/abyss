package layout

import (
	"testing"
	"unicode/utf8"

	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/stretchr/testify/require"
)

// A submit button without CSS width sizes to its label plus padding, not to
// the full container width.
func TestSubmitButtonSizesToLabel(t *testing.T) {
	doc := dom.Parse(`<body><input type="submit" name="btnK" value="Поиск"></body>`)
	box := Build(doc, css.Compute(doc), 200)
	btn := findTag(box, "input")
	require.NotNil(t, btn, "button box not found")
	require.Equal(t, float64(utf8.RuneCountInString("Поиск")+2), btn.Border.Width, "button width = %v", btn.Border.Width)
}
