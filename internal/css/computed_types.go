package css

import "fmt"

// DisplayType represents the CSS display property.
type DisplayType int

const (
	DisplayBlock DisplayType = iota
	DisplayInline
	DisplayInlineBlock
	DisplayFlex
	DisplayGrid
	DisplayNone
)

// PositionType represents the CSS position property.
type PositionType int

const (
	PositionStatic PositionType = iota
	PositionRelative
	PositionAbsolute
)

// TextAlignType represents the text-align property. Only the block's own value
// is used; inheritance is not modelled.
type TextAlignType int

const (
	TextAlignLeft TextAlignType = iota
	TextAlignCenter
	TextAlignRight
)

// OverflowType represents the CSS overflow property.
type OverflowType int

const (
	OverflowVisible OverflowType = iota
	OverflowHidden
)

// FlexDirection represents flex-direction.
type FlexDirection int

const (
	FlexDirectionRow FlexDirection = iota
	FlexDirectionColumn
)

// JustifyContent represents justify-content.
type JustifyContent int

const (
	JustifyFlexStart JustifyContent = iota
	JustifyCenter
	JustifyFlexEnd
	JustifySpaceBetween
	JustifySpaceAround
)

// AlignItems represents align-items.
type AlignItems int

const (
	AlignFlexStart AlignItems = iota
	AlignCenter
	AlignFlexEnd
	AlignStretch
)

// BorderStyle represents the border style.
type BorderStyle int

const (
	BorderNone BorderStyle = iota
	BorderSolid
)

// Length is a CSS length, percentage, auto track, or fractional grid track.
type Length struct {
	IsAuto     bool
	Px         float64
	IsPercent  bool
	Percent    float64
	IsFraction bool
	Fraction   float64
}

// Auto returns Length{IsAuto: true}.
func Auto() Length {
	return Length{IsAuto: true}
}

// Px returns a pixel length.
func Px(v float64) Length {
	return Length{Px: v}
}

// Percent returns a percentage length.
func Percent(v float64) Length {
	return Length{IsPercent: true, Percent: v}
}

// Fr returns a fractional grid track size.
func Fr(v float64) Length {
	return Length{IsFraction: true, Fraction: v}
}

// Resolve resolves the length against a reference (for percents).
func (l Length) Resolve(ref float64) float64 {
	if l.IsAuto {
		return 0
	}
	if l.IsPercent {
		return l.Percent * ref / 100
	}
	if l.IsFraction {
		return 0
	}
	return l.Px
}

// String returns the CSS serialization of the length.
func (l Length) String() string {
	if l.IsAuto {
		return "auto"
	}
	if l.IsPercent {
		return fmt.Sprintf("%.2f%%", l.Percent)
	}
	if l.IsFraction {
		return fmt.Sprintf("%.2ffr", l.Fraction)
	}
	return fmt.Sprintf("%.2fpx", l.Px)
}

// Color is an RGB color.
type Color struct {
	R, G, B uint8
	IsSet   bool
}

// Transparent returns an unset/transparent color.
func Transparent() Color {
	return Color{}
}

// RGB returns a set RGB color.
func RGB(r, g, b uint8) Color {
	return Color{R: r, G: g, B: b, IsSet: true}
}

// String returns the CSS serialization of the color.
func (c Color) String() string {
	if !c.IsSet {
		return "transparent"
	}
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}

// Border is a CSS border (width, style, color).
type Border struct {
	Width float64
	Style BorderStyle
	Color Color
}

// BoxModel is top/right/bottom/left lengths.
type BoxModel struct {
	Top, Right, Bottom, Left Length
}

// IsZero reports whether all sides are zero px.
func (b BoxModel) IsZero() bool {
	return isZeroLength(b.Top) && isZeroLength(b.Right) && isZeroLength(b.Bottom) && isZeroLength(b.Left)
}

func isZeroLength(length Length) bool {
	return !length.IsAuto && !length.IsPercent && !length.IsFraction && length.Px == 0
}

// ComputedStyle holds all computed CSS properties for a node.
type ComputedStyle struct {
	Display                 DisplayType
	Position                PositionType
	Overflow                OverflowType
	Width                   Length
	Height                  Length
	MinWidth                Length
	MaxWidth                Length
	Top                     Length
	Left                    Length
	Right                   Length
	Bottom                  Length
	Margin                  BoxModel
	Padding                 BoxModel
	BorderTop               Border
	BorderRight             Border
	BorderBottom            Border
	BorderLeft              Border
	BackgroundColor         Color
	Color                   Color
	ZIndex                  int
	TextAlign               TextAlignType
	FlexDirection           FlexDirection
	JustifyContent          JustifyContent
	AlignItems              AlignItems
	FlexGrow                float64
	FlexShrink              float64
	GridTemplateColumns     []Length
	GridTemplateRows        []Length
	Gap                     Length
	BorderRadiusTopLeft     Length
	BorderRadiusTopRight    Length
	BorderRadiusBottomRight Length
	BorderRadiusBottomLeft  Length
}

// Default returns the default computed style.
func Default() ComputedStyle {
	return ComputedStyle{
		Display:         DisplayBlock,
		Position:        PositionStatic,
		Overflow:        OverflowVisible,
		Width:           Auto(),
		Height:          Auto(),
		BackgroundColor: Transparent(),
		Color:           RGB(30, 30, 30),
		FlexDirection:   FlexDirectionRow,
		JustifyContent:  JustifyFlexStart,
		AlignItems:      AlignStretch,
		FlexGrow:        0,
		FlexShrink:      1,
		Gap:             Px(0),
	}
}
