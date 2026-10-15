package asciirender

import (
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// A logo-sized image with dark strokes must keep visible glyphs at several
// column counts, so the logo does not collapse to blank cells.
func TestLogoQuantizationKeepsVisibleGlyphs(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 272, 92))
	for y := 0; y < 92; y++ {
		for x := 0; x < 272; x++ {
			img.Set(x, y, color.White)
			if x%24 < 4 {
				img.Set(x, y, color.Black)
			}
		}
	}
	for _, cols := range []int{20, 40, 60} {
		grid := Quantize(img, Options{Columns: cols})
		var sb strings.Builder
		for _, row := range grid {
			for _, c := range row {
				sb.WriteRune(c.Char)
			}
		}
		visible := strings.Count(sb.String(), "") - 1 - strings.Count(sb.String(), " ")
		require.Greater(t, visible, 0, "logo at %d cols has no visible glyphs", cols)
	}
}
