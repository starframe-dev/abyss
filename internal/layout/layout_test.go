package layout

import (
	"strings"
	"testing"

	"github.com/starframe-dev/abyss/internal/cellmetrics"
	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/stretchr/testify/assert"
)

func TestGridFractionRowsUseIntrinsicHeightsWhenContainerAuto(t *testing.T) {
	doc := dom.Parse(`<main style="display:grid;grid-template-rows:1fr 2fr;gap:4px"><div style="height:20px">first</div><div style="height:30px">second</div></main>`)
	box := Build(doc, css.Compute(doc), 100)
	grid := findTag(box, "main")
	if grid == nil || len(grid.Children) != 2 {
		t.Fatalf("grid children = %#v", grid)
	}
	first, second := grid.Children[0], grid.Children[1]
	if first.Content.Height != 20 || second.Content.Height != 30 {
		t.Fatalf("row item heights = %v, %v", first.Content.Height, second.Content.Height)
	}
	if second.Content.Y-first.Content.Y != first.Margin.Height+4 {
		t.Fatalf("second row y = %v, want %v", second.Content.Y, first.Content.Y+first.Margin.Height+4)
	}
	if grid.Content.Height != 54 {
		t.Fatalf("auto grid height = %v, want 54", grid.Content.Height)
	}
}

func TestWordWrapUsesDisplayColumnsForCJK(t *testing.T) {
	box := &LayoutBox{Content: Rect{X: 3, Y: 4}}
	lines := wordWrap([]string{"猫", "a"}, box, 4*cellmetrics.WidthPx)
	if len(lines) != 1 || lines[0].Text != "猫 a" {
		t.Fatalf("wrapped lines = %#v, want one line containing 猫 a", lines)
	}

	lines = wordWrap([]string{"ab", "cd", "ef"}, box, 5*cellmetrics.WidthPx)
	if len(lines) != 2 || lines[0].Text != "ab cd" || lines[1].Text != "ef" {
		t.Fatalf("ASCII wrapping = %#v", lines)
	}
}

func BenchmarkBuildLayout(b *testing.B) {
	html := `<main style="display:grid;grid-template-columns:repeat(3,1fr);gap:8px">` + strings.Repeat(`<article><h2>Title</h2><p>Text with кириллица</p></article>`, 40) + `</main>`
	doc := dom.Parse(html)
	styles := css.ComputeWithExternal(doc, []string{`article { display:flex; gap:4px; } h2 { color: red; }`})
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Build(doc, styles, 800)
	}
}

// findTag recursively finds the first LayoutBox with the given tag name.
func findTag(box *LayoutBox, tag string) *LayoutBox {
	if box.Node != nil && box.Node.TagName == tag {
		return box
	}
	for _, child := range box.Children {
		if found := findTag(child, tag); found != nil {
			return found
		}
	}
	return nil
}

// findAllTags recursively finds all LayoutBoxes with the given tag name.
func findAllTags(box *LayoutBox, tag string) []*LayoutBox {
	var result []*LayoutBox
	if box.Node != nil && box.Node.TagName == tag {
		result = append(result, box)
	}
	for _, child := range box.Children {
		result = append(result, findAllTags(child, tag)...)
	}
	return result
}

func TestBuild_Basic(t *testing.T) {
	html := `<div style="width: 200px; height: 100px;"></div>`
	doc := dom.Parse(html)
	styles := css.Compute(doc)
	box := Build(doc, styles, 800)

	assert.NotNil(t, box)
	assert.Equal(t, float64(800), box.Content.Width)
}

func TestLayout_BlockFlow(t *testing.T) {
	html := `
    <div style="width: 400px;">
        <div style="height: 50px;"></div>
        <div style="height: 30px;"></div>
    </div>`
	doc := dom.Parse(html)
	styles := css.Compute(doc)
	box := Build(doc, styles, 400)

	root := findTag(box, "div")
	assert.NotNil(t, root)
	assert.Equal(t, 400.0, root.Content.Width)

	children := findAllTags(root, "div")
	// Exclude root itself
	var childBoxes []*LayoutBox
	for _, c := range children {
		if c != root {
			childBoxes = append(childBoxes, c)
		}
	}
	assert.Len(t, childBoxes, 2)
	assert.Equal(t, 50.0, childBoxes[0].Content.Height)
	assert.Equal(t, 30.0, childBoxes[1].Content.Height)
	assert.True(t, childBoxes[1].Content.Y > childBoxes[0].Content.Y)
}

func TestLayout_FlexRow(t *testing.T) {
	html := `
    <div style="display: flex; width: 300px;">
        <div style="width: 100px;"></div>
        <div style="width: 100px;"></div>
    </div>`
	doc := dom.Parse(html)
	styles := css.Compute(doc)
	box := Build(doc, styles, 300)

	root := findTag(box, "div")
	assert.NotNil(t, root)
	assert.Equal(t, 300.0, root.Content.Width)

	children := findAllTags(root, "div")
	var childBoxes []*LayoutBox
	for _, c := range children {
		if c != root {
			childBoxes = append(childBoxes, c)
		}
	}
	assert.Len(t, childBoxes, 2)
	assert.Equal(t, 100.0, childBoxes[0].Content.Width)
	assert.Equal(t, 100.0, childBoxes[1].Content.Width)
	assert.True(t, childBoxes[1].Content.X > childBoxes[0].Content.X)
}

func TestLayout_FlexColumn(t *testing.T) {
	html := `
    <div style="display: flex; flex-direction: column; width: 300px;">
        <div style="height: 50px;"></div>
        <div style="height: 30px;"></div>
    </div>`
	doc := dom.Parse(html)
	styles := css.Compute(doc)
	box := Build(doc, styles, 300)

	root := findTag(box, "div")
	assert.NotNil(t, root)
	children := findAllTags(root, "div")
	var childBoxes []*LayoutBox
	for _, c := range children {
		if c != root {
			childBoxes = append(childBoxes, c)
		}
	}
	assert.Len(t, childBoxes, 2)
	assert.True(t, childBoxes[1].Content.Y > childBoxes[0].Content.Y)
}

func TestLayout_Grid(t *testing.T) {
	html := `
    <div style="display: grid; grid-template-columns: 100px 100px; width: 300px;">
        <div style="height: 50px;"></div>
        <div style="height: 50px;"></div>
        <div style="height: 50px;"></div>
    </div>`
	doc := dom.Parse(html)
	styles := css.Compute(doc)
	box := Build(doc, styles, 300)

	root := findTag(box, "div")
	assert.NotNil(t, root)
	children := findAllTags(root, "div")
	var childBoxes []*LayoutBox
	for _, c := range children {
		if c != root {
			childBoxes = append(childBoxes, c)
		}
	}
	assert.Len(t, childBoxes, 3)
	assert.True(t, childBoxes[1].Content.X > childBoxes[0].Content.X)
	assert.True(t, childBoxes[2].Content.Y > childBoxes[0].Content.Y)
}

func TestLayout_Absolute(t *testing.T) {
	html := `
    <div style="position: relative; width: 400px; height: 400px;">
        <div style="position: absolute; top: 10px; left: 20px; width: 50px; height: 50px;"></div>
    </div>`
	doc := dom.Parse(html)
	styles := css.Compute(doc)
	box := Build(doc, styles, 400)

	parent := findTag(box, "div")
	assert.NotNil(t, parent)
	children := findAllTags(parent, "div")
	var child *LayoutBox
	for _, c := range children {
		if c != parent {
			child = c
			break
		}
	}
	assert.NotNil(t, child)
	assert.Equal(t, 10.0+parent.Content.Y, child.Content.Y)
	assert.Equal(t, 20.0+parent.Content.X, child.Content.X)
}

func TestLayout_Padding(t *testing.T) {
	html := `<div style="padding: 20px; width: 200px;"></div>`
	doc := dom.Parse(html)
	styles := css.Compute(doc)
	box := Build(doc, styles, 200)

	root := findTag(box, "div")
	assert.NotNil(t, root)
	assert.Equal(t, 160.0, root.Content.Width)
	assert.Equal(t, 200.0, root.Border.Width)
}

func TestLayout_PaddedSiblingDoesNotOverlapPreviousSibling(t *testing.T) {
	html := `<main style="width: 100px;">
        <h1 style="height: 2px; background-color: #312e81;">Heading</h1>
        <div style="padding: 4px; background-color: #ef4444;">Panel</div>
    </main>`
	doc := dom.Parse(html)
	styles := css.Compute(doc)
	box := Build(doc, styles, 100)

	heading := findTag(box, "h1")
	panel := findTag(box, "div")
	assert.NotNil(t, heading)
	assert.NotNil(t, panel)
	assert.GreaterOrEqual(t, panel.Border.Y, heading.Margin.Y+heading.Margin.Height)
	assert.Equal(t, panel.Border.Y+4, panel.Content.Y)
}

func TestLayout_Margin(t *testing.T) {
	html := `
    <div style="width: 200px;">
        <div style="margin: 10px; height: 50px;"></div>
    </div>`
	doc := dom.Parse(html)
	styles := css.Compute(doc)
	box := Build(doc, styles, 200)

	parent := findTag(box, "div")
	assert.NotNil(t, parent)
	children := findAllTags(parent, "div")
	var child *LayoutBox
	for _, c := range children {
		if c != parent {
			child = c
			break
		}
	}
	assert.NotNil(t, child)
	assert.Equal(t, 10.0, child.Content.X-parent.Content.X)
}

func TestLayout_InlineText(t *testing.T) {
	html := `<p>Hello world this is a test paragraph</p>`
	doc := dom.Parse(html)
	styles := css.Compute(doc)
	box := Build(doc, styles, 800)

	root := findTag(box, "p")
	assert.NotNil(t, root)
	assert.Greater(t, len(root.TextLines), 0)
	assert.Contains(t, root.TextLines[0].Text, "Hello")
}

func TestLayout_ImagePlaceholder(t *testing.T) {
	doc := dom.Parse(`<div><img alt="Article preview" src="https://example.com/image.png"></div>`)
	styles := css.Compute(doc)
	box := Build(doc, styles, 200*cellmetrics.WidthPx)

	image := findTag(box, "img")
	assert.NotNil(t, image)
	assert.Greater(t, image.Content.Height, 0.0)
	assert.Contains(t, image.TextLines[0].Text, "Article preview")
}
