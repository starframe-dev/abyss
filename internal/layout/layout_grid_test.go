package layout

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/css"
	"github.com/stretchr/testify/assert"
)

func TestGridOffsetX(t *testing.T) {
	tests := []struct {
		name      string
		base      float64
		colWidths []float64
		gap       float64
		col       int
		want      float64
	}{
		{"zero columns", 0, []float64{10, 20, 30}, 5, 0, 0},
		{"first column", 0, []float64{10, 20, 30}, 5, 0, 0},
		{"first column with gap", 0, []float64{10, 20, 30}, 5, 0, 0},
		{"second column", 0, []float64{10, 20, 30}, 5, 1, 15},
		{"third column", 0, []float64{10, 20, 30}, 5, 2, 40},
		{"with base offset", 100, []float64{10, 20, 30}, 5, 1, 115},
		{"with base offset and gap", 100, []float64{10, 20, 30}, 5, 2, 140},
		{"negative gap", 0, []float64{10, 20, 30}, -5, 2, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, gridOffsetX(tt.base, tt.colWidths, tt.gap, tt.col))
		})
	}
}

func TestGridOffsetY(t *testing.T) {
	tests := []struct {
		name       string
		base       float64
		rowHeights []float64
		gap        float64
		row        int
		want       float64
	}{
		{"zero rows", 0, []float64{10, 20, 30}, 5, 0, 0},
		{"first row", 0, []float64{10, 20, 30}, 5, 0, 0},
		{"first row with gap", 0, []float64{10, 20, 30}, 5, 0, 0},
		{"second row", 0, []float64{10, 20, 30}, 5, 1, 15},
		{"third row", 0, []float64{10, 20, 30}, 5, 2, 40},
		{"with base offset", 100, []float64{10, 20, 30}, 5, 1, 115},
		{"with base offset and gap", 100, []float64{10, 20, 30}, 5, 2, 140},
		{"negative gap", 0, []float64{10, 20, 30}, -5, 2, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, gridOffsetY(tt.base, tt.rowHeights, tt.gap, tt.row))
		})
	}
}

func TestSumRowHeights(t *testing.T) {
	tests := []struct {
		name       string
		rowHeights []float64
		gap        float64
		want       float64
	}{
		{"empty", []float64{}, 5, 0},
		{"single row", []float64{10}, 5, 10},
		{"two rows", []float64{10, 20}, 5, 35},
		{"three rows", []float64{10, 20, 30}, 5, 70},
		{"negative gap", []float64{10, 20}, -5, 25},
		{"zero gap", []float64{10, 20}, 0, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, sumRowHeights(tt.rowHeights, tt.gap))
		})
	}
}

func TestResolveGridPercentageRowsUseContainerHeight(t *testing.T) {
	got := resolveGridFractionalRows([]css.Length{css.Percent(50), css.Px(20)}, 200, 0)
	assert.Equal(t, []css.Length{css.Px(100), css.Px(20)}, got)
}

func TestResolveGridFractionalRows(t *testing.T) {
	got := resolveGridFractionalRows([]css.Length{css.Px(40), css.Fr(1), css.Fr(3)}, 240, 10)
	assert.Equal(t, []css.Length{css.Px(40), css.Px(45), css.Px(135)}, got)
}

func TestResolveGridFractionalTracks(t *testing.T) {
	got := resolveGridColumnWidths([]css.Length{css.Px(100), css.Fr(1), css.Fr(2)}, 500, 10)
	assert.Equal(t, []float64{100, 380.0 / 3, 760.0 / 3}, got)
}

func TestResolveGridColumnWidths(t *testing.T) {
	tests := []struct {
		name         string
		columns      []css.Length
		contentWidth float64
		gap          float64
		want         []float64
	}{
		{"single px column", []css.Length{{Px: 100}}, 500, 0, []float64{100}},
		{"single auto column", []css.Length{{IsAuto: true, Px: 100}}, 500, 0, []float64{0}},
		{"single percent column", []css.Length{{IsPercent: true, Percent: 100}}, 500, 0, []float64{500}},
		{"mixed columns", []css.Length{{Px: 100}, {IsPercent: true, Percent: 50}, {IsAuto: true, Px: 100}}, 500, 0, []float64{100, 250, 0}},
		{"negative remaining", []css.Length{{Px: 100}, {IsAuto: true, Px: 100}}, 100, 0, []float64{100, 0}},
		{"gap included", []css.Length{{Px: 50}, {IsAuto: true, Px: 50}}, 300, 10, []float64{50, 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveGridColumnWidths(tt.columns, tt.contentWidth, tt.gap)
			assert.Equal(t, tt.want, got)
		})
	}
}
