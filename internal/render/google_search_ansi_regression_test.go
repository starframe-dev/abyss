package render

import (
	"os"
	"strings"
	"testing"

	"github.com/starframe-dev/abyss/internal/ascii"
	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/starframe-dev/abyss/internal/layout"
	"github.com/stretchr/testify/require"
)

// The search field surface (#f1f3f4) must reach the ANSI output of the
// google.com-shaped fixture, not only the framebuffer.
func TestGoogleSearchSurfaceReachesANSI(t *testing.T) {
	raw, err := os.ReadFile("../layout/testdata/google_search_flex.html")
	require.NoError(t, err)
	doc := dom.Parse(string(raw))
	box := layout.Build(doc, css.Compute(doc), 200)
	fb := Render(box, 200)
	lines := ascii.RenderANSI(fb, box, 200)
	count := 0
	for _, l := range lines {
		count += strings.Count(l, "48;2;241;243;244")
	}
	require.Greater(t, count, 0, "search field surface missing from ANSI output")
}
