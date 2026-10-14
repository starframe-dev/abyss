//go:build legacy

package css

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestDisplayType(t *testing.T) {
    assert.Equal(t, DisplayBlock, DisplayBlock)
    assert.Equal(t, DisplayInline, DisplayInline)
    assert.Equal(t, DisplayFlex, DisplayFlex)
    assert.Equal(t, DisplayGrid, DisplayGrid)
    assert.Equal(t, DisplayNone, DisplayNone)
}

func TestPositionType(t *testing.T) {
    assert.Equal(t, PositionStatic, PositionStatic)
    assert.Equal(t, PositionRelative, PositionRelative)
    assert.Equal(t, PositionAbsolute, PositionAbsolute)
}

func TestOverflowType(t *testing.T) {
    assert.Equal(t, OverflowVisible, OverflowVisible)
    assert.Equal(t, OverflowHidden, OverflowHidden)
}

func TestFlexDirection(t *testing.T) {
    assert.Equal(t, FlexDirectionRow, FlexDirectionRow)
    assert.Equal(t, FlexDirectionColumn, FlexDirectionColumn)
}

func TestJustifyContent(t *testing.T) {
    assert.Equal(t, JustifyFlexStart, JustifyFlexStart)
    assert.Equal(t, JustifyCenter, JustifyCenter)
    assert.Equal(t, JustifyFlexEnd, JustifyFlexEnd)
    assert.Equal(t, JustifySpaceBetween, JustifySpaceBetween)
    assert.Equal(t, JustifySpaceAround, JustifySpaceAround)
}

func TestAlignItems(t *testing.T) {
    assert.Equal(t, AlignFlexStart, AlignFlexStart)
    assert.Equal(t, AlignCenter, AlignCenter)
    assert.Equal(t, AlignFlexEnd, AlignFlexEnd)
    assert.Equal(t, AlignStretch, AlignStretch)
}

func TestLengthAuto(t *testing.T) {
    l := Auto()
    assert.True(t, l.IsAuto)
    assert.Equal(t, float64(0), l.Resolve(100))
    assert.Equal(t, "auto", l.String())
}

func TestLengthPx(t *testing.T) {
    l := Px(100.5)
    assert.False(t, l.IsAuto)
    assert.False(t, l.IsPercent)
    assert.Equal(t, 100.5, l.Resolve(100))
    assert.Equal(t, "100.5px", l.String())
}

func TestLengthPercent(t *testing.T) {
    l := Percent(50)
    assert.False(t, l.IsAuto)
    assert.True(t, l.IsPercent)
    assert.Equal(t, 50.0, l.Resolve(100))
    assert.Equal(t, "50.0%", l.String())
}

func TestLengthResolve(t *testing.T) {
    // Auto resolves to 0
    l1 := Auto()
    assert.Equal(t, 0.0, l1.Resolve(100))

    // Px resolves to px value
    l2 := Px(100)
    assert.Equal(t, 100.0, l2.Resolve(100))

    // Percent resolves to percent of reference
    l3 := Percent(50)
    assert.Equal(t, 100.0, l3.Resolve(200))
    assert.Equal(t, 50.0, l3.Resolve(100))
}

func TestLengthString(t *testing.T) {
    assert.Equal(t, "auto", Auto().String())
    assert.Equal(t, "100.0px", Px(100.0).String())
    assert.Equal(t, "100.5px", Px(100.5).String())
    assert.Equal(t, "50.0%", Percent(50.0).String())
}

func TestColorTransparent(t *testing.T) {
    c := Transparent()
    assert.False(t, c.IsSet)
    assert.Equal(t, "transparent", c.String())
}

func TestColorRGB(t *testing.T) {
    c := RGB(255, 128, 64)
    assert.True(t, c.IsSet)
    assert.Equal(t, uint8(255), c.R)
    assert.Equal(t, uint8(128), c.G)
    assert.Equal(t, uint8(64), c.B)
    assert.Equal(t, "#ff8040", c.String())
}

func TestColorHexShort(t *testing.T) {
    c := Hex("#abc")
    assert.Equal(t, uint8(170), c.R)
    assert.Equal(t, uint8(187), c.G)
    assert.Equal(t, uint8(204), c.B)
    assert.Equal(t, "#aabbcc", c.String())
}

func TestColorHexLong(t *testing.T) {
    c := Hex("#aabbcc")
    assert.Equal(t, uint8(170), c.R)
    assert.Equal(t, uint8(187), c.G)
    assert.Equal(t, uint8(204), c.B)
    assert.Equal(t, "#aabbcc", c.String())
}

func TestColorHexInvalid(t *testing.T) {
    c := Hex("#12345")
    assert.False(t, c.IsSet)
    assert.Equal(t, "transparent", c.String())
    assert.Equal(t, uint8(0), c.R)
    assert.Equal(t, uint8(0), c.G)
    assert.Equal(t, uint8(0), c.B)
}

func TestBorderStyle(t *testing.T) {
    assert.Equal(t, BorderNone, BorderNone)
    assert.Equal(t, BorderSolid, BorderSolid)
}

func TestBoxModelIsZero(t *testing.T) {
    // Zero box model
    b1 := BoxModel{
        Top:    Length{Px: 0},
        Right:  Length{Px: 0},
        Bottom: Length{Px: 0},
        Left:   Length{Px: 0},
    }
    assert.True(t, b1.IsZero())

    // Non-zero box model
    b2 := BoxModel{
        Top:    Length{Px: 10},
        Right:  Length{Px: 0},
        Bottom: Length{Px: 0},
        Left:   Length{Px: 0},
    }
    assert.False(t, b2.IsZero())

    // Auto is not zero
    b3 := BoxModel{
        Top:    Length{IsAuto: true},
        Right:  Length{Px: 0},
        Bottom: Length{Px: 0},
        Left:   Length{Px: 0},
    }
    assert.False(t, b3.IsZero())

    // Percent is not zero
    b4 := BoxModel{
        Top:    Length{IsPercent: true, Percent: 50},
        Right:  Length{Px: 0},
        Bottom: Length{Px: 0},
        Left:   Length{Px: 0},
    }
    assert.False(t, b4.IsZero())
}

func TestDefaultComputedStyle(t *testing.T) {
    s := Default()

    assert.Equal(t, DisplayBlock, s.Display)
    assert.Equal(t, PositionStatic, s.Position)
    assert.Equal(t, OverflowVisible, s.Overflow)
    assert.True(t, s.Width.IsAuto)
    assert.True(t, s.Height.IsAuto)
    assert.False(t, s.BackgroundColor.IsSet)
    assert.Equal(t, uint8(30), s.Color.R)
    assert.Equal(t, uint8(30), s.Color.G)
    assert.Equal(t, uint8(30), s.Color.B)
    assert.Equal(t, FlexDirectionRow, s.FlexDirection)
    assert.Equal(t, JustifyFlexStart, s.JustifyContent)
    assert.Equal(t, AlignStretch, s.AlignItems)
    assert.Equal(t, 0.0, s.FlexGrow)
    assert.Equal(t, 1.0, s.FlexShrink)
    assert.Equal(t, "0.0px", s.Gap.String())
}
