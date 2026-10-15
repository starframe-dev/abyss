package render

import (
	"strings"
	"testing"

	"github.com/starframe-dev/abyss/internal/ascii"
	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/starframe-dev/abyss/internal/layout"
)

// The browser search field shows a magnifier on the left and keyboard, voice and
// lens icons on the right. In the terminal they are drawn as ASCII art inside the
// field, so the field reads as a search box instead of an empty pill.
func TestSearchFieldDrawsAsciiIconsInside(t *testing.T) {
	doc := dom.Parse(`<body><div style="position:relative"><input name="q" size="57" style="height:30px"></div></body>`)
	box := layout.Build(doc, css.Compute(doc), 200)
	fb := Render(box, 200)
	lines := ascii.Render(fb, box, 200)
	found := false
	for _, line := range lines {
		if strings.Contains(line, "(o") && strings.Contains(line, "[|]") {
			found = true
		}
	}
	if !found {
		t.Fatalf("search field lacks ASCII icons (left magnifier \"(o\", right \"[|]\"); lines:\n%s", strings.Join(lines, "\n"))
	}
}
