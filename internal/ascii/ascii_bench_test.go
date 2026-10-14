package ascii

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/css"
)

func BenchmarkANSIConversion(b *testing.B) {
	grid := make([][]Cell, 80)
	for y := range grid {
		grid[y] = make([]Cell, 120)
		for x := range grid[y] {
			grid[y][x] = Cell{Char: 'x', FG: css.RGB(uint8(x), uint8(y), 120), BG: css.RGB(20, 30, 40)}
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = renderANSI(grid, 120)
	}
}
