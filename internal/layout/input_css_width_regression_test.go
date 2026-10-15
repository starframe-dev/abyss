package layout

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/cellmetrics"
	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/stretchr/testify/require"
)

// google.com sets the search field width in CSS (.Ucigb{width:458px}) on top of
// size="57". Inside a flex row the explicit CSS width must win over the size
// default, otherwise the field overflows its box and the right icons drift out.
func TestInputExplicitCSSWidthWinsOverSizeInFlexRow(t *testing.T) {
	doc := dom.Parse(`<body><style>.Ucigb{width:458px}</style><div style="display:flex"><input class="lst Ucigb" name="q" size="57" style="height:30px"></div></body>`)
	box := Build(doc, css.Compute(doc), 200*cellmetrics.WidthPx)
	field := findTag(box, "input")
	require.NotNil(t, field)
	require.Equal(t, 458.0, field.Border.Width)
}
