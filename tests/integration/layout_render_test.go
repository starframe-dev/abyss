package integration

import (
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"
	"github.com/starframe-dev/abyss/internal/ascii"
	"github.com/starframe-dev/abyss/internal/cellmetrics"
	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/starframe-dev/abyss/internal/layout"
	"github.com/starframe-dev/abyss/internal/render"
	"github.com/stretchr/testify/require"
)

func TestLayoutRenderRepresentativePagesAlignAndClip(t *testing.T) {
	const columns = 24
	const viewport = columns * cellmetrics.WidthPx
	pages := []struct {
		name string
		html string
		css  string
		want []string
	}{
		{
			name: "flex nested blocks and wide text",
			html: `<html><body><main><section class="flex"><div>左側</div><div>right pane</div></section><article><p>Nested block content wraps cleanly</p><p>outside-visible-marker</p></article></main></body></html>`,
			css:  `.flex { display: flex; } .flex > div { width: 100px; } article { padding: 2px; }`,
			want: []string{"左側", "right pane", "Nested block", "outside-visible-marker"},
		},
		{
			name: "grid table ansi and viewport clipping",
			html: `<html><body><div class="grid"><div>grid-A</div><div>grid-B</div></div><table><tr><td>row-one</td><td>行二</td></tr></table><p class="styled">ANSI styled text</p><p>` + strings.Repeat("W", 90) + `</p></body></html>`,
			css:  `.grid { display: grid; grid-template-columns: 1fr 1fr; } .styled { color: #ff0000; background-color: #eeeeee; }`,
			want: []string{"grid-A", "grid-B", "row-one", "行二", "ANSI styled text"},
		},
	}

	for _, page := range pages {
		t.Run(page.name, func(t *testing.T) {
			doc := dom.Parse(page.html)
			require.NotNil(t, doc)
			styles := css.ComputeWithExternalAtWidth(doc, []string{page.css}, viewport)
			tree := layout.Build(doc, styles, viewport)
			require.NotNil(t, tree)
			fb := render.Render(tree, viewport)
			plain := ascii.Render(fb, tree, columns)
			ansi := ascii.RenderANSI(fb, tree, columns)
			require.NotEmpty(t, plain)
			require.Equal(t, len(plain), len(ansi))
			for row := range plain {
				require.Equal(t, columns, runewidth.StringWidth(plain[row]), "plain row %d", row)
				require.Equal(t, columns, runewidth.StringWidth(ansi[row]), "ANSI row %d", row)
				for _, line := range []string{plain[row], ansi[row]} {
					for _, marker := range page.want {
						if strings.Contains(line, marker) {
							// A complete marker must fit in the viewport; clipping never
							// expands a rendered row beyond its terminal cell budget.
							require.LessOrEqual(t, runewidth.StringWidth(line), columns)
						}
					}
				}
			}
			joined := strings.Join(plain, "\n")
			for _, marker := range page.want {
				require.Contains(t, joined, marker)
			}
			long := strings.Repeat("W", 90)
			if strings.Contains(page.html, long) {
				require.NotContains(t, joined, long, "content wider than viewport must wrap or clip")
			}
		})
	}
}
