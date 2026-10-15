package ascii

import (
	"strings"
	"testing"

	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/starframe-dev/abyss/internal/framebuffer"
	"github.com/starframe-dev/abyss/internal/layout"
)

// An absolutely positioned icon image must not paint its ASCII cells over the
// search field it sits on. The field's own ASCII icons stay visible.
func TestAbsoluteIconDoesNotPaintOverSearchField(t *testing.T) {
	doc := dom.Parse(`<body><div style="position:relative"><input name="q" size="57" style="height:30px"><img src="/tia.png" style="position:absolute;right:5px;top:4px" width="27" height="23"></div></body>`)
	box := layout.Build(doc, css.Compute(doc), 200)
	fb := framebuffer.New(200, 40)
	lines := Render(fb, box, 200)
	if strings.Contains(strings.Join(lines, "\n"), ",,") {
		t.Fatalf("absolute icon cells drawn over the field:\n%s", strings.Join(lines, "\n"))
	}
}
