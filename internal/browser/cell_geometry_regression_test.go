package browser

import (
	"strings"
	"testing"

	"github.com/starframe-dev/abyss/internal/ascii"
	"github.com/starframe-dev/abyss/internal/cellmetrics"
	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/starframe-dev/abyss/internal/layout"
	"github.com/starframe-dev/abyss/internal/render"
)

// One terminal cell is 10 px wide and 20 px high. A 200-column viewport is
// therefore 2000 px wide, and a logo block of 88 px plus a 160 px header must
// place the search form inside the first 56 rows.
func TestGoogleShapedPageShowsSearchOnFirstScreen(t *testing.T) {
	const cols = 200
	const html = `<html><body>` +
		`<div style="height:160px"></div>` +
		`<div style="height:88px"></div>` +
		`<form><textarea name="q"></textarea>` +
		`<input type="submit" value="Поиск в Google">` +
		`<input type="button" value="Мне повезёт!"></form>` +
		`</body></html>`

	doc := dom.Parse(html)
	viewport := cols * cellmetrics.WidthPx
	box := layout.Build(doc, css.Compute(doc), float64(viewport))
	fb := render.Render(box, viewport)
	lines := ascii.Render(fb, box, cols)

	for _, label := range []string{"Поиск в Google", "Мне повезёт!"} {
		row := -1
		for i, line := range lines {
			if strings.Contains(line, label) {
				row = i
				break
			}
		}
		if row < 0 || row >= 56 {
			t.Fatalf("%q at row %d, want inside first 56 rows", label, row)
		}
	}
}
