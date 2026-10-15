package layout

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/stretchr/testify/require"
)

// Rows without text fields keep the original shrink behaviour: the minimum
// content width only applies to items that contain a field.
func TestFlexRowWithoutFieldsStillShrinks(t *testing.T) {
	doc := dom.Parse(`<body><div style="display:flex"><div>aaa</div><div>bbb</div><div>ccc</div><div>ddd</div><div>eee</div><div>fff</div><div>ggg</div><div>hhh</div></div></body>`)
	box := Build(doc, css.Compute(doc), 200)
	row := findTag(box, "div")
	require.NotNil(t, row, "flex row not found")
	var total float64
	for _, c := range row.Children {
		total += c.Border.Width
	}
	require.LessOrEqual(t, total, 200.0+0.5, "flex row overflowed: total=%v", total)
}
