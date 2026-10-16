package render

import (
	"image"
	"testing"

	"github.com/starframe-dev/abyss/internal/asciirender"
	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/starframe-dev/abyss/internal/framebuffer"
	"github.com/starframe-dev/abyss/internal/layout"
	"github.com/stretchr/testify/assert"
)

func TestRender_NilTree(t *testing.T) {
	fb := Render(nil, 80)
	assert.Equal(t, 1, fb.Width)
	assert.Equal(t, 1, fb.Height)
}

func TestRender_BasicBox(t *testing.T) {
	doc := dom.Parse(`<div style="width: 100px; height: 50px; background-color: red;"></div>`)
	box := layout.Build(doc, css.Compute(doc), 800)
	fb := Render(box, 800)
	assert.Equal(t, 800, fb.Width)
	assert.Greater(t, fb.Height, 0)
	assert.Equal(t, uint8(255), fb.Get(10, 10).R)
}

func TestRender_Border(t *testing.T) {
	doc := dom.Parse(`<div style="width: 50px; height: 30px; border: 2px solid blue;"></div>`)
	box := layout.Build(doc, css.Compute(doc), 800)
	c := Render(box, 800).Get(1, 1)
	assert.Equal(t, uint8(255), c.B)
}

func TestRender_NegativeCoordinates(t *testing.T) {
	box := &layout.LayoutBox{Margin: layout.Rect{X: -5, Y: -3, Width: 10, Height: 8}}
	fb := Render(box, 20)
	assert.Equal(t, 8, fb.Height)
	assert.Equal(t, 0.0, box.Margin.X)
	assert.Equal(t, 0.0, box.Margin.Y)
}

func TestRender_ImageQuantizationAndCache(t *testing.T) {
	old := ImageProvider
	defer func() { ImageProvider = old }()
	calls := 0
	ImageProvider = func(src string) image.Image {
		calls++
		return image.NewRGBA(image.Rect(0, 0, 16, 16))
	}
	doc := dom.Parse(`<img src="fixture" alt="sample">`)
	box := layout.Build(doc, css.Compute(doc), 80)
	var imgBox *layout.LayoutBox
	var find func(*layout.LayoutBox)
	find = func(b *layout.LayoutBox) {
		if b.Node != nil && b.Node.TagName == "img" {
			imgBox = b
		}
		for _, child := range b.Children {
			find(child)
		}
	}
	find(box)
	if imgBox == nil {
		t.Fatal("image layout box not found")
	}
	imgBox.Content.Width, imgBox.Content.Height = 16, 16
	imgBox.Border.Width, imgBox.Border.Height = 16, 16
	imgBox.TextLines = []layout.TextLine{{}}
	grid := quantizeImageBox(imgBox)
	assert.NotEmpty(t, grid)
	assert.Nil(t, imgBox.TextLines)
	assert.Same(t, &grid[0][0], &quantizeImageBox(imgBox)[0][0])
	assert.Equal(t, 1, calls)
}

func TestRender_ImageProviderMissingAndDrawASCII(t *testing.T) {
	old := ImageProvider
	defer func() { ImageProvider = old }()
	ImageProvider = nil
	box := &layout.LayoutBox{Node: &dom.Node{TagName: "img"}, Border: layout.Rect{Width: 2, Height: 2}}
	assert.Nil(t, quantizeImageBox(box))
	fb := framebuffer.New(4, 4)
	drawAsciiCells(fb, box, [][]asciirender.Cell{{{FG: css.RGB(255, 0, 0)}}})
	assert.Equal(t, uint8(255), fb.Get(0, 0).R)
}
