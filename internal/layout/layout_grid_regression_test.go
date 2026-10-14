package layout

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
)

func TestGridFixedRowsDoNotExpandToFitItems(t *testing.T) {
	box := &LayoutBox{
		Content: Rect{Width: 100},
		Children: []*LayoutBox{
			{Node: &dom.Node{Type: dom.ElementNode, TagName: "div"}, Style: css.ComputedStyle{Height: css.Px(80)}},
			{Node: &dom.Node{Type: dom.ElementNode, TagName: "div"}, Style: css.ComputedStyle{Height: css.Px(30)}},
		},
		Style: css.ComputedStyle{
			GridTemplateColumns: []css.Length{css.Px(100)},
			GridTemplateRows:    []css.Length{css.Px(20), css.Px(20)},
		},
	}

	if got := layoutGrid(box, 100); got != 40 {
		t.Fatalf("grid height = %v, want fixed track sum 40", got)
	}
	if got := box.Children[1].Content.Y; got != 20 {
		t.Fatalf("second row y = %v, want 20 despite first item overflow", got)
	}
}

func TestGridImplicitRowsAreAutoSized(t *testing.T) {
	box := &LayoutBox{
		Content: Rect{Width: 100},
		Children: []*LayoutBox{
			{Node: &dom.Node{Type: dom.ElementNode, TagName: "div"}, Style: css.ComputedStyle{Height: css.Px(80)}},
			{Node: &dom.Node{Type: dom.ElementNode, TagName: "div"}, Style: css.ComputedStyle{Height: css.Px(30)}},
		},
		Style: css.ComputedStyle{
			GridTemplateColumns: []css.Length{css.Px(100)},
			GridTemplateRows:    []css.Length{css.Px(20)},
		},
	}

	if got := layoutGrid(box, 100); got != 50 {
		t.Fatalf("grid height = %v, want fixed 20px row plus 30px implicit auto row", got)
	}
	if got := box.Children[1].Content.Y; got != 20 {
		t.Fatalf("implicit row y = %v, want 20", got)
	}
}

func TestGridPercentRowsInAutoHeightUseIntrinsicSizing(t *testing.T) {
	box := &LayoutBox{
		Content:  Rect{Width: 100},
		Children: []*LayoutBox{{Node: &dom.Node{Type: dom.ElementNode, TagName: "div"}, Style: css.ComputedStyle{Height: css.Px(80)}}},
		Style: css.ComputedStyle{
			Height:              css.Auto(),
			GridTemplateColumns: []css.Length{css.Px(100)},
			GridTemplateRows:    []css.Length{css.Percent(50)},
		},
	}

	if got := layoutGrid(box, 100); got != 80 {
		t.Fatalf("auto-height percentage row = %v, want intrinsic 80", got)
	}
}

func TestGridWithoutExplicitRowsUsesIntrinsicAutoTracks(t *testing.T) {
	box := &LayoutBox{
		Content: Rect{Width: 100},
		Children: []*LayoutBox{
			{Node: &dom.Node{Type: dom.ElementNode, TagName: "div"}, Style: css.ComputedStyle{Height: css.Px(80)}},
			{Node: &dom.Node{Type: dom.ElementNode, TagName: "div"}, Style: css.ComputedStyle{Height: css.Px(30)}},
		},
		Style: css.ComputedStyle{GridTemplateColumns: []css.Length{css.Px(100)}},
	}

	if got := layoutGrid(box, 100); got != 110 {
		t.Fatalf("implicit auto grid height = %v, want 110", got)
	}
	if got := box.Children[1].Content.Y; got != 80 {
		t.Fatalf("second auto row y = %v, want 80", got)
	}
}
