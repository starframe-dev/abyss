package render

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/ascii"
	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/starframe-dev/abyss/internal/layout"
	"github.com/stretchr/testify/assert"
)

func TestRender_BasicBox(t *testing.T) {
	html := `<div style="width: 100px; height: 50px; background-color: red;"></div>`
	doc := dom.Parse(html)
	styles := css.Compute(doc)
	box := layout.Build(doc, styles, 800)
	fb := Render(box, 800)

	assert.Equal(t, 800, fb.Width)
	assert.Greater(t, fb.Height, 0)

	// Check that background is red
	c := fb.Get(10, 10)
	assert.True(t, c.IsSet)
	assert.Equal(t, uint8(255), c.R)
}

func TestRender_Border(t *testing.T) {
	html := `<div style="width: 50px; height: 30px; border: 2px solid blue;"></div>`
	doc := dom.Parse(html)
	styles := css.Compute(doc)
	box := layout.Build(doc, styles, 800)
	fb := Render(box, 800)

	// Border pixel should be blue
	c := fb.Get(1, 1)
	assert.True(t, c.IsSet)
	assert.Equal(t, uint8(0), c.R)
	assert.Equal(t, uint8(0), c.G)
	assert.Equal(t, uint8(255), c.B)
}

func TestRender_TransparentBackground(t *testing.T) {
	html := `<div style="width: 50px; height: 30px; background-color: transparent;"></div>`
	doc := dom.Parse(html)
	styles := css.Compute(doc)
	box := layout.Build(doc, styles, 800)
	fb := Render(box, 800)

	// Transparent bg should not overwrite the white page background
	c := fb.Get(10, 10)
	assert.Equal(t, css.RGB(255, 255, 255), c)
}

func TestRender_FlexBox(t *testing.T) {
	html := `
    <div style="display: flex; width: 200px;">
        <div style="width: 100px; height: 50px; background-color: red;"></div>
        <div style="width: 100px; height: 50px; background-color: blue;"></div>
    </div>`
	doc := dom.Parse(html)
	styles := css.Compute(doc)
	box := layout.Build(doc, styles, 800)
	fb := Render(box, 800)

	// Left half should be red, right half blue
	red := fb.Get(10, 10)
	blue := fb.Get(110, 10)
	assert.True(t, red.IsSet)
	assert.True(t, blue.IsSet)
	assert.Equal(t, uint8(255), red.R)
	assert.Equal(t, uint8(255), blue.B)
}

func TestRender_EqualZIndexPreservesDOMPaintOrder(t *testing.T) {
	html := `<body style="width: 100px; height: 100px; background-color: #ef4444;">
        <h1 style="width: 80px; height: 20px; background-color: #312e81;">Heading</h1>
        <div style="height: 1px;"></div><div style="height: 1px;"></div>
        <div style="height: 1px;"></div><div style="height: 1px;"></div>
        <div style="height: 1px;"></div><div style="height: 1px;"></div>
        <div style="height: 1px;"></div><div style="height: 1px;"></div>
        <div style="height: 1px;"></div><div style="height: 1px;"></div>
        <div style="height: 1px;"></div><div style="height: 1px;"></div>
        <div style="height: 1px;"></div><div style="height: 1px;"></div>
    </body>`
	doc := dom.Parse(html)
	styles := css.Compute(doc)
	box := layout.Build(doc, styles, 100)
	fb := Render(box, 100)

	// The child heading must paint over its parent when both use z-index: auto/0.
	// An unstable z-index sort can reorder equal items and let the parent erase it.
	pixel := fb.Get(5, 5)
	assert.Equal(t, css.RGB(49, 46, 129), pixel)
}

func TestRender_ASCII(t *testing.T) {
	html := `<div style="width: 10px; height: 10px; background-color: #ff0000;"></div>`
	doc := dom.Parse(html)
	styles := css.Compute(doc)
	box := layout.Build(doc, styles, 800)
	fb := Render(box, 800)
	lines := ascii.Render(fb, box, 800)

	assert.Greater(t, len(lines), 0)
}

func TestRender_TextOverlay(t *testing.T) {
	html := `<p style="color: white;">Hello</p>`
	doc := dom.Parse(html)
	styles := css.Compute(doc)
	box := layout.Build(doc, styles, 800)
	fb := Render(box, 800)
	lines := ascii.Render(fb, box, 800)

	found := false
	for _, line := range lines {
		if len(line) > 0 {
			found = true
			break
		}
	}
	assert.True(t, found)
}
