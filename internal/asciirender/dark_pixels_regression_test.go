package asciirender

import (
	"image"
	"image/color"
	"testing"

	"github.com/stretchr/testify/require"
)

// A thin dark stroke on white must still produce non-blank cells after
// area-averaged quantization, not vanish at small column counts.
func TestQuantizeKeepsThinDarkStroke(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 200, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 200; x++ {
			img.Set(x, y, color.White)
		}
	}
	for y := 0; y < 100; y++ {
		for x := 96; x < 104; x++ {
			img.Set(x, y, color.Black)
		}
	}
	grid := Quantize(img, Options{Columns: 20})
	require.NotEmpty(t, grid)
	var nonBlank int
	for _, row := range grid {
		for _, c := range row {
			if c.Char != ' ' {
				nonBlank++
			}
		}
	}
	require.Greater(t, nonBlank, 0, "dark stroke disappeared after quantization")
}
