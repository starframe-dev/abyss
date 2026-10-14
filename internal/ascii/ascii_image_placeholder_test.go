package ascii

import (
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/starframe-dev/abyss/internal/layout"
	"github.com/starframe-dev/abyss/internal/render"
)

// TestRenderANSI_ImagePlaceholderDoesNotErasePicture is a regression for the
// bug where the "[image] alt" placeholder produced by layoutImage was painted
// on top of the decoded ASCII art, erasing its first columns.
func TestRenderANSI_ImagePlaceholderDoesNotErasePicture(t *testing.T) {
	const src = "https://example.test/red.png"

	prev := render.ImageProvider
	render.ImageProvider = func(s string) image.Image {
		if s != src {
			return nil
		}
		img := image.NewRGBA(image.Rect(0, 0, 8, 4))
		for y := 0; y < 4; y++ {
			for x := 0; x < 8; x++ {
				img.Set(x, y, color.RGBA{R: 220, G: 30, B: 30, A: 255})
			}
		}
		return img
	}
	defer func() { render.ImageProvider = prev }()

	doc := dom.Parse(`<html><body><img src="` + src + `" alt="red block"/></body></html>`)
	styles := css.Compute(doc)
	box := layout.Build(doc, styles, 120)
	fb := render.Render(box, 120)
	lines := RenderANSI(fb, box, 120)
	joined := strings.Join(lines, "\n")

	if strings.Contains(joined, "[image]") {
		t.Errorf("ANSI output must not contain the [image] placeholder once the picture is decoded")
	}
	// Abyss emits FG and BG as a single combined SGR sequence, so match the
	// foreground prefix rather than a full "...m" terminator.
	if !strings.Contains(joined, "\x1b[38;2;220;30;30;") {
		t.Errorf("ANSI output lost the image truecolor foreground; got:\n%s", joined)
	}
}

// TestRenderANSI_PlaceholderRemainsWhenImageMissing verifies we did not simply
// delete the fallback: undecodable images still show their alt text.
func TestRenderANSI_PlaceholderRemainsWhenImageMissing(t *testing.T) {
	prev := render.ImageProvider
	render.ImageProvider = func(string) image.Image { return nil }
	defer func() { render.ImageProvider = prev }()

	doc := dom.Parse(`<html><body><img src="https://example.test/none.png" alt="alt stays"/></body></html>`)
	styles := css.Compute(doc)
	box := layout.Build(doc, styles, 120)
	fb := render.Render(box, 120)
	joined := strings.Join(RenderANSI(fb, box, 120), "\n")

	if !strings.Contains(joined, "alt stays") {
		t.Errorf("placeholder alt text must remain when no image could be decoded")
	}
}
