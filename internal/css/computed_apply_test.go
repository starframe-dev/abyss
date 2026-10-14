package css

import "testing"

func TestApplyBoxModel(t *testing.T) {
	s := &ComputedStyle{}
	applyBoxModel(s, "margin", "10px")
	applyBoxModel(s, "margin-top", "5px")
	applyBoxModel(s, "margin-right", "5px")
	applyBoxModel(s, "margin-bottom", "5px")
	applyBoxModel(s, "margin-left", "5px")
	applyBoxModel(s, "padding", "8px")
	applyBoxModel(s, "padding-top", "4px")
	applyBoxModel(s, "padding-right", "4px")
	applyBoxModel(s, "padding-bottom", "4px")
	applyBoxModel(s, "padding-left", "4px")
}

func TestApplyBorder(t *testing.T) {
	s := &ComputedStyle{}
	applyBorder(s, "border", "1px solid #000")
	applyBorder(s, "border-top", "2px solid red")
	applyBorder(s, "border-right", "2px solid green")
	applyBorder(s, "border-bottom", "2px solid blue")
	applyBorder(s, "border-left", "2px solid yellow")
	applyBorder(s, "border-top-color", "red")
	applyBorder(s, "border-right-color", "green")
	applyBorder(s, "border-bottom-color", "blue")
	applyBorder(s, "border-left-color", "yellow")
	applyBorder(s, "border-radius", "5px")
}

func TestApplyDimensions(t *testing.T) {
	s := &ComputedStyle{}
	applyDimensions(s, "width", "100px")
	applyDimensions(s, "height", "50px")
	applyDimensions(s, "min-width", "10px")
	applyDimensions(s, "min-height", "10px")
	applyDimensions(s, "max-width", "200px")
	applyDimensions(s, "max-height", "200px")
}

func TestApplyOffset(t *testing.T) {
	s := &ComputedStyle{}
	applyOffset(s, "top", "10px")
	applyOffset(s, "right", "20px")
	applyOffset(s, "bottom", "30px")
	applyOffset(s, "left", "40px")
}

func TestApplyLayout(t *testing.T) {
	s := &ComputedStyle{}
	applyLayout(s, "display", "flex")
	applyLayout(s, "position", "absolute")
	applyLayout(s, "overflow", "hidden")
	applyLayout(s, "z-index", "10")
}

func TestResolveBoxModel(t *testing.T) {
	s := &ComputedStyle{}
	_ = s.Margin
	_ = s.Padding
}
