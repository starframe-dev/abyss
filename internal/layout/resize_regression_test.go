package layout

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/stretchr/testify/require"
)

// Layout must follow the viewport width: an auto-width block gets the full
// width at each size, so a resize re-layouts instead of reusing old geometry.
func TestLayoutFollowsViewportWidth(t *testing.T) {
	doc := dom.Parse(`<body><div>aaaa</div></body>`)
	styles := css.Compute(doc)
	for _, w := range []float64{80, 120, 200} {
		box := Build(doc, styles, w)
		div := findTag(box, "div")
		require.NotNil(t, div, "div not found at width %d", w)
		require.Equal(t, float64(w), div.Border.Width, "width %d", w)
	}
}
