package css

import (
	"fmt"
	"strconv"
	"strings"
)

// DisplayType represents a CSS display value.
type DisplayType int

const (
	DisplayBlock DisplayType = iota
	DisplayInline
	DisplayFlex
	DisplayGrid
	DisplayNone
)

// PositionType represents a CSS position value.
type PositionType int

const (
	PositionStatic PositionType = iota
	PositionRelative
	PositionAbsolute
)

// OverflowType represents a CSS overflow value.
type OverflowType int

const (
	OverflowVisible OverflowType = iota
	OverflowHidden
)

// FlexDirection represents the flex-direction property.
type FlexDirection int

const (
	FlexDirectionRow FlexDirection = iota
	FlexDirectionColumn
)

// JustifyContent represents the justify-content property.
type JustifyContent int

const (
	JustifyFlexStart JustifyContent = iota
	JustifyCenter
	JustifyFlexEnd
	JustifySpaceBetween
	JustifySpaceAround
)

// AlignItems represents the align-items property.
type AlignItems int

const (
	AlignFlexStart AlignItems = iota
	AlignCenter
	AlignFlexEnd
	AlignStretch
)

// BorderStyle represents a supported CSS border style.
type BorderStyle int

const (
	BorderNone BorderStyle = iota
	BorderSolid
)

// Length represents a CSS length value in pixels, percent, or auto.
type Length struct {
	IsAuto    bool
	Px        float64
	IsPercent bool
	Percent   float64
}

// Auto returns an automatic length.
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

// Resolve computes a pixel value using ref as the percentage base.
func (l Length) Resolve(ref float64) float64 {
	if l.IsAuto {
		return 0
	}
	if l.IsPercent {
		return ref * l.Percent / 100
	}
	return l.Px
}

// String returns the CSS representation of the length.
func (l Length) String() string {
	if l.IsAuto {
		return "auto"
	}
	if l.IsPercent {
		return strconv.FormatFloat(l.Percent, 'f', -1, 64) + "%"
	}
	return strconv.FormatFloat(l.Px, 'f', -1, 64) + "px"
}

// Color represents an RGB color. IsSet distinguishes a set color from transparent.
type Color struct {
	R, G, B uint8
	IsSet   bool
}

// Transparent returns a transparent (unset) color.
func Transparent() Color {
	return Color{}
}

// RGB constructs a set RGB color.
func RGB(r, g, b uint8) Color {
	return Color{R: r, G: g, B: b, IsSet: true}
}

// Hex parses a #RGB or #RRGGBB color. Invalid input returns transparent.
func Hex(s string) Color {
	s = strings.TrimPrefix(s, "#")
	if len(s) == 3 {
		var expanded strings.Builder
		expanded.Grow(6)
		for _, r := range s {
			expanded.WriteRune(r)
			expanded.WriteRune(r)
		}
		s = expanded.String()
	}
	if len(s) != 6 {
		return Transparent()
	}
	parsed := [3]uint8{}
	for i := range parsed {
		value, err := strconv.ParseUint(s[i*2:i*2+2], 16, 8)
		if err != nil {
			return Transparent()
		}
		parsed[i] = uint8(value)
	}
	return RGB(parsed[0], parsed[1], parsed[2])
}

// String returns the lowercase hex representation, or an empty string if unset.
func (c Color) String() string {
	if !c.IsSet {
		return ""
	}
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}

// Border represents a CSS border.
type Border struct {
	Width float64
	Style BorderStyle
	Color Color
}

// BoxModel represents top, right, bottom, and left values.
type BoxModel struct {
	Top, Right, Bottom, Left Length
}

// IsZero reports whether every side is a non-auto, non-percent zero length.
func (b BoxModel) IsZero() bool {
	for _, side := range [...]Length{b.Top, b.Right, b.Bottom, b.Left} {
		if side.Px != 0 || side.IsAuto || side.IsPercent {
			return false
		}
	}
	return true
}

// ComputedStyle contains computed CSS properties for a node.
type ComputedStyle struct {
	Display                 DisplayType
	Position                PositionType
	Overflow                OverflowType
	Width, Height           Length
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

// Default returns a computed style initialized with CSS defaults supported here.
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
