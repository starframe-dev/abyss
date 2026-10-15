package layout

import (
	"os"
	"testing"

	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/stretchr/testify/require"
)

// The google.com search textarea is width:100% inside flex ancestors whose
// auto basis collapses when a sibling takes the row. The field must keep a
// usable width (at least the default 20 columns) instead of 0 or ~1 px.
func TestGoogleSearchTextareaKeepsWidthInFlexChain(t *testing.T) {
	raw, err := os.ReadFile("testdata/google_search_flex.html")
	require.NoError(t, err)
	doc := dom.Parse(string(raw))
	box := Build(doc, css.Compute(doc), 200)
	ta := findTag(box, "textarea")
	require.NotNil(t, ta, "textarea box not found")
	require.GreaterOrEqual(t, ta.Border.Width, 20.0, "textarea width = %v, want >= 20", ta.Border.Width)
}
