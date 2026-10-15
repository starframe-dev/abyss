# Computed CSS Types

`internal/css/computed_types.go` defines the typed values used to represent CSS properties after computation. It provides enums for supported property values, a compact `Length` representation with helpers for construction, resolution, and serialization, RGB colors and border/box-model structures, and `ComputedStyle` with its default initializer.

## Property value types

The following integer types encode the supported values. Their constants use `iota`, so the numeric values are implementation details; use the named constants rather than relying on their ordinals.

| Type | Constants |
| --- | --- |
| `DisplayType` | `DisplayBlock`, `DisplayInline`, `DisplayInlineBlock`, `DisplayFlex`, `DisplayGrid`, `DisplayNone` |
| `PositionType` | `PositionStatic`, `PositionRelative`, `PositionAbsolute` |
| `TextAlignType` | `TextAlignLeft`, `TextAlignCenter`, `TextAlignRight` |
| `OverflowType` | `OverflowVisible`, `OverflowHidden` |
| `FlexDirection` | `FlexDirectionRow`, `FlexDirectionColumn` |
| `JustifyContent` | `JustifyFlexStart`, `JustifyCenter`, `JustifyFlexEnd`, `JustifySpaceBetween`, `JustifySpaceAround` |
| `AlignItems` | `AlignFlexStart`, `AlignCenter`, `AlignFlexEnd`, `AlignStretch` |
| `BorderStyle` | `BorderNone`, `BorderSolid` |

`TextAlignType` represents only the block's own text alignment; inheritance is not modeled here.

## Length

`Length` represents a pixel value, percentage, `auto`, or fractional grid-track size:

```go
type Length struct {
    IsAuto     bool
    Px         float64
    IsPercent  bool
    Percent    float64
    IsFraction bool
    Fraction   float64
}
```

Construct values with `Auto()`, `Px(v)`, `Percent(v)`, or `Fr(v)`. These helpers set the corresponding discriminator and value. The fields remain exported, so callers can also construct `Length` directly; the representation does not enforce that only one discriminator is set. When multiple flags are true, methods give precedence to `IsAuto`, then `IsPercent`, then `IsFraction`, and otherwise treat the value as pixels.

- `Resolve(ref)` returns `0` for `auto` and fractional tracks, computes percentages as `Percent * ref / 100`, and returns `Px` for pixel values. Fractional tracks are not resolved by this method because they require grid-track allocation.
- `String()` serializes as `auto`, or as a value with two decimal places and the appropriate suffix (`%`, `fr`, or `px`).

## Colors, borders, and box model

`Color` stores RGB channels (`uint8`) and an `IsSet` flag. `Transparent()` returns the zero/unset color; `RGB(r, g, b)` returns a set color. `Color.String()` formats unset colors as `transparent` and set colors as lowercase six-digit hexadecimal (`#rrggbb`).

`Border` groups a pixel `Width`, `BorderStyle`, and `Color`. `BoxModel` stores `Top`, `Right`, `Bottom`, and `Left` values as `Length`. `BoxModel.IsZero()` is true only when every side is a pixel length equal to zero. `auto`, percentage, and fractional values are not considered zero, even if their numeric field is zero.

## ComputedStyle

`ComputedStyle` is the aggregate of computed properties for one node. Its exported fields are:

| Field | Meaning |
| --- | --- |
| `Display`, `Position`, `Overflow` | Display, positioning, and overflow modes |
| `Width`, `Height`, `MinWidth`, `MaxWidth` | Size constraints |
| `Top`, `Left`, `Right`, `Bottom` | Position offsets |
| `Margin`, `Padding` | Four-sided box-model lengths |
| `BorderTop`, `BorderRight`, `BorderBottom`, `BorderLeft` | Per-side borders |
| `BackgroundColor`, `Color` | Background and foreground colors |
| `ZIndex` | Stacking order value |
| `TextAlign` | Text alignment |
| `FlexDirection`, `JustifyContent`, `AlignItems` | Flex layout configuration |
| `FlexGrow`, `FlexShrink`, `FlexWrap`, `Order` | Flex item configuration |
| `GridTemplateColumns`, `GridTemplateRows` | Grid track lengths |
| `Gap` | Layout gap |
| `BorderRadiusTopLeft`, `BorderRadiusTopRight`, `BorderRadiusBottomRight`, `BorderRadiusBottomLeft` | Per-corner radii |

`Default()` creates a new style with these explicit defaults: block display, static positioning, visible overflow, auto width and height, transparent background, foreground RGB `(30, 30, 30)`, row flex direction, flex-start justification, stretch alignment, flex grow `0`, flex shrink `1`, and a zero-pixel gap. Other fields retain their Go zero values: for example, the remaining lengths are pixel zero, enums take their first constant, booleans are false, numeric values are zero, colors are unset, and grid slices are nil. This is a baseline data value; this file does not implement CSS parsing, cascade, inheritance, or layout allocation.
