//go:build legacy

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
		{"first column", 0, []float64{10, 20, 30}, 5, 1, 10},
		{"first column with gap", 0, []float64{10, 20, 30}, 5, 1, 10},
		{"second column", 0, []float64{10, 20, 30}, 5, 2, 35},
		{"third column", 0, []float64{10, 20, 30}, 5, 3, 60},
		{"with base offset", 100, []float64{10, 20, 30}, 5, 1, 105},
		{"with base offset and gap", 100, []float64{10, 20, 30}, 5, 2, 140},
		{"negative gap", 0, []float64{10, 20, 30}, -5, 2, 25},
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
		{"first row", 0, []float64{10, 20, 30}, 5, 1, 10},
		{"first row with gap", 0, []float64{10, 20, 30}, 5, 1, 10},
		{"second row", 0, []float64{10, 20, 30}, 5, 2, 35},
		{"third row", 0, []float64{10, 20, 30}, 5, 3, 60},
		{"with base offset", 100, []float64{10, 20, 30}, 5, 1, 105},
		{"with base offset and gap", 100, []float64{10, 20, 30}, 5, 2, 140},
		{"negative gap", 0, []float64{10, 20, 30}, -5, 2, 25},
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
		{"single row", []float64{10}, 5, 15},
		{"two rows", []float64{10, 20}, 5, 30},
		{"three rows", []float64{10, 20, 30}, 5, 40},
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
		{"single percent column", []css.Length{{IsPercent: true, Px: 100}}, 500, 0, []float64{500}},
		{"mixed columns", []css.Length{{Px: 100}, {IsPercent: true, Px: 50}, {IsAuto: true, Px: 100}}, 500, 0, []float64{100, 400, 0}},
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

func TestLayoutGrid(t *testing.T) {
	tests := []struct {
		name           string
		box            *LayoutBox
		contentWidth   float64
		want           float64
		wantColWidths  []float64
		wantRowHeights []float64
	}{
		{"simple 2x2 grid", func() *LayoutBox {
			box := &LayoutBox{
				Content: Rect{X: 0, Y: 0, Width: 400, Height: 200},
				Children: []*LayoutBox{
					{Style: css.ComputedStyle{Width: css.Px(100), Height: css.Px(50)}},
					{Style: css.ComputedStyle{Width: css.Px(100), Height: css.Px(80)}},
					{Style: css.ComputedStyle{Width: css.Px(100), Height: css.Px(50)}},
					{Style: css.ComputedStyle{Width: css.Px(100), Height: css.Px(80)}},
				},
				Style: css.ComputedStyle{
					Width:               css.Px(400),
					Height:              css.Px(200),
					Gap:                 css.Px(10),
					GridTemplateColumns: []css.Length{css.Px(100), css.Px(100), css.Px(100), css.Px(100)},
					GridTemplateRows:    []css.Length{css.Px(50), css.Px(50), css.Px(50), css.Px(50)},
				},
			}
			return box
		}(), 400, 200, []float64{100, 100, 100, 100}, []float64{80, 80, 80, 80}},
		{"single column grid", func() *LayoutBox {
			box := &LayoutBox{
				Content: Rect{X: 0, Y: 0, Width: 300, Height: 100},
				Children: []*LayoutBox{
					{Style: css.ComputedStyle{Width: css.Px(100), Height: css.Px(50)}},
					{Style: css.ComputedStyle{Width: css.Px(100), Height: css.Px(80)}},
					{Style: css.ComputedStyle{Width: css.Px(100), Height: css.Px(60)}},
				},
				Style: css.ComputedStyle{
					Width:               css.Px(300),
					Height:              css.Px(100),
					Gap:                 css.Px(10),
					GridTemplateColumns: []css.Length{css.Px(100), css.Px(100), css.Px(100)},
					GridTemplateRows:    []css.Length{css.Px(50), css.Px(50), css.Px(50)},
				},
			}
			return box
		}(), 300, 100, nil, nil},
		{"empty children", func() *LayoutBox {
			box := &LayoutBox{
				Content:  Rect{X: 0, Y: 0, Width: 300, Height: 100},
				Children: []*LayoutBox{},
				Style: css.ComputedStyle{
					Width:               css.Px(300),
					Height:              css.Px(100),
					Gap:                 css.Px(0),
					GridTemplateColumns: []css.Length{css.Px(100), css.Px(100), css.Px(100)},
					GridTemplateRows:    []css.Length{css.Px(50), css.Px(50), css.Px(50)},
				},
			}
			return box
		}(), 300, 0, nil, nil},
		{"auto columns", func() *LayoutBox {
			box := &LayoutBox{
				Content: Rect{X: 0, Y: 0, Width: 400, Height: 200},
				Children: []*LayoutBox{
					{Style: css.ComputedStyle{Width: css.Px(100), Height: css.Px(50)}},
					{Style: css.ComputedStyle{Width: css.Px(100), Height: css.Px(80)}},
					{Style: css.ComputedStyle{Width: css.Px(100), Height: css.Px(50)}},
					{Style: css.ComputedStyle{Width: css.Px(100), Height: css.Px(80)}},
				},
				Style: css.ComputedStyle{
					Width:               css.Px(400),
					Height:              css.Px(200),
					Gap:                 css.Px(10),
					GridTemplateColumns: []css.Length{{IsAuto: true, Px: 100}, {IsAuto: true, Px: 100}, {IsAuto: true, Px: 100}, {IsAuto: true, Px: 100}},
					GridTemplateRows:    []css.Length{css.Px(50), css.Px(50), css.Px(50), css.Px(50)},
				},
			}
			return box
		}(), 400, 200, nil, nil},
		{"percent columns", func() *LayoutBox {
			box := &LayoutBox{
				Content: Rect{X: 0, Y: 0, Width: 400, Height: 200},
				Children: []*LayoutBox{
					{Style: css.ComputedStyle{Width: css.Px(100), Height: css.Px(50)}},
					{Style: css.ComputedStyle{Width: css.Px(100), Height: css.Px(80)}},
					{Style: css.ComputedStyle{Width: css.Px(100), Height: css.Px(50)}},
					{Style: css.ComputedStyle{Width: css.Px(100), Height: css.Px(80)}},
				},
				Style: css.ComputedStyle{
					Width:               css.Px(400),
					Height:              css.Px(200),
					Gap:                 css.Px(10),
					GridTemplateColumns: []css.Length{css.Px(25), css.Px(25), css.Px(25), css.Px(25)},
					GridTemplateRows:    []css.Length{css.Px(50), css.Px(50), css.Px(50), css.Px(50)},
				},
			}
			return box
		}(), 400, 200, nil, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := layoutGrid(tt.box, tt.contentWidth)
			assert.Equal(t, tt.want, got)

			if tt.wantColWidths != nil {
				// Verify column widths are set correctly
				for i, child := range tt.box.Children {
					// Column index
					col := i % len(tt.box.Style.GridTemplateColumns)
					assert.Equal(t, tt.wantColWidths[col], child.Style.Width)
				}
			}
		})
	}
}

func TestPlaceGridChildren(t *testing.T) {
	tests := []struct {
		name         string
		box          *LayoutBox
		columns      []css.Length
		rows         []css.Length
		gap          float64
		contentWidth float64
		wantHeights  []float64
	}{
		{"2x2 grid", func() *LayoutBox {
			box := &LayoutBox{
				Content: Rect{X: 0, Y: 0, Width: 400, Height: 200},
				Children: []*LayoutBox{
					{Style: css.ComputedStyle{Width: css.Px(100), Height: css.Px(50)}},
					{Style: css.ComputedStyle{Width: css.Px(100), Height: css.Px(80)}},
					{Style: css.ComputedStyle{Width: css.Px(100), Height: css.Px(50)}},
					{Style: css.ComputedStyle{Width: css.Px(100), Height: css.Px(80)}},
				},
				Style: css.ComputedStyle{
					Width:               css.Px(400),
					Height:              css.Px(200),
					Gap:                 css.Px(10),
					GridTemplateColumns: []css.Length{css.Px(100), css.Px(100)},
					GridTemplateRows:    []css.Length{css.Px(50), css.Px(50)},
				},
			}
			return box
		}(), []css.Length{css.Px(100), css.Px(100)}, []css.Length{css.Px(50), css.Px(50)}, 10, 400, []float64{80, 80}},
		{"3x2 grid", func() *LayoutBox {
			box := &LayoutBox{
				Content: Rect{X: 0, Y: 0, Width: 400, Height: 200},
				Children: []*LayoutBox{
					{Style: css.ComputedStyle{Width: css.Px(100), Height: css.Px(50)}},
					{Style: css.ComputedStyle{Width: css.Px(100), Height: css.Px(80)}},
					{Style: css.ComputedStyle{Width: css.Px(100), Height: css.Px(50)}},
					{Style: css.ComputedStyle{Width: css.Px(100), Height: css.Px(80)}},
					{Style: css.ComputedStyle{Width: css.Px(100), Height: css.Px(50)}},
					{Style: css.ComputedStyle{Width: css.Px(100), Height: css.Px(80)}},
					{Style: css.ComputedStyle{Width: css.Px(100), Height: css.Px(50)}},
					{Style: css.ComputedStyle{Width: css.Px(100), Height: css.Px(80)}},
				},
				Style: css.ComputedStyle{
					Width:               css.Px(400),
					Height:              css.Px(200),
					Gap:                 css.Px(10),
					GridTemplateColumns: []css.Length{css.Px(100), css.Px(100), css.Px(100)},
					GridTemplateRows:    []css.Length{css.Px(50), css.Px(50)},
				},
			}
			return box
		}(), []css.Length{css.Px(100), css.Px(100), css.Px(100)}, []css.Length{css.Px(50), css.Px(50)}, 10, 400, []float64{80, 80}},
		{"tall child increases row height", func() *LayoutBox {
			box := &LayoutBox{
				Content: Rect{X: 0, Y: 0, Width: 400, Height: 200},
				Children: []*LayoutBox{
					{Style: css.ComputedStyle{Width: css.Px(100), Height: css.Px(100)}},
					{Style: css.ComputedStyle{Width: css.Px(100), Height: css.Px(50)}},
					{Style: css.ComputedStyle{Width: css.Px(100), Height: css.Px(100)}},
				},
				Style: css.ComputedStyle{
					Width:               css.Px(400),
					Height:              css.Px(200),
					Gap:                 css.Px(10),
					GridTemplateColumns: []css.Length{css.Px(100), css.Px(100), css.Px(100)},
					GridTemplateRows:    []css.Length{css.Px(50), css.Px(50), css.Px(50)},
				},
			}
			return box
		}(), []css.Length{css.Px(100), css.Px(100), css.Px(100)}, []css.Length{css.Px(50), css.Px(50), css.Px(50)}, 10, 400, []float64{110, 110}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotHeights := placeGridChildren(tt.box, tt.columns, tt.rows, []float64{100, 100, 100}, tt.gap, tt.contentWidth)
			assert.Equal(t, tt.wantHeights, gotHeights)
		})
	}
}
