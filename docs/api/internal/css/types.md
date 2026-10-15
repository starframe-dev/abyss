# CSS Types

Package `css` defines the value types used to represent the subset of CSS properties supported by the application. It provides enumerations for discrete property values, value objects for lengths and colors, box and border models, and a `ComputedStyle` structure with a constructor for supported defaults.

## Discrete property types

All property types below are integer-backed enums. Their constants use `iota`, so the numeric values start at zero and follow the listed order. They describe the supported values; conversion to CSS text is not provided by these types.

| Type | Constants, in declaration order | Property |
|---|---|---|
| `DisplayType` | `DisplayBlock`, `DisplayInline`, `DisplayFlex`, `DisplayGrid`, `DisplayNone` | `display` |
| `PositionType` | `PositionStatic`, `PositionRelative`, `PositionAbsolute` | `position` |
| `OverflowType` | `OverflowVisible`, `OverflowHidden` | `overflow` |
| `FlexDirection` | `FlexDirectionRow`, `FlexDirectionColumn` | `flex-direction` |
| `JustifyContent` | `JustifyFlexStart`, `JustifyCenter`, `JustifyFlexEnd`, `JustifySpaceBetween`, `JustifySpaceAround` | `justify-content` |
| `AlignItems` | `AlignFlexStart`, `AlignCenter`, `AlignFlexEnd`, `AlignStretch` | `align-items` |
| `BorderStyle` | `BorderNone`, `BorderSolid` | border style |

## Length

`Length` represents a pixel length, percentage, or automatic length:

```go
type Length struct {
    IsAuto    bool
    Px        float64
    IsPercent bool
    Percent   float64
}
```

The fields are exported and can be constructed directly. The helper constructors establish the conventional representation:

- `Auto() Length` marks the value as automatic.
- `Px(v float64) Length` stores a pixel value in `Px`.
- `Percent(v float64) Length` marks the value as a percentage and stores its numeric amount in `Percent`.

`Resolve(ref float64) float64` returns `0` for auto, `ref * Percent / 100` for a percentage, or `Px` otherwise. `String() string` returns `auto`, a percentage followed by `%`, or a pixel value followed by `px`. Numeric formatting uses `strconv.FormatFloat` with the shortest round-trippable decimal representation.

Because the struct is public, callers can set conflicting flags or values. `IsAuto` takes precedence over `IsPercent` in both methods; otherwise `IsPercent` takes precedence over `Px`.

## Color

`Color` stores 8-bit red, green, and blue channels and an `IsSet` flag. `IsSet` distinguishes an explicit color (including black) from transparent/unset.

- `Transparent() Color` returns the zero, unset color.
- `RGB(r, g, b uint8) Color` returns a set RGB color.
- `Hex(s string) Color` parses `#RGB` or `#RRGGBB`. The leading `#` is optional. Three-digit input is expanded by duplicating each character. Invalid length or invalid hexadecimal digits return `Transparent()`; parsing is case-insensitive through base-16 parsing.
- `String() string` returns an empty string when unset, otherwise lowercase `#rrggbb` with two hexadecimal digits per channel.

## Border and box model

`Border` groups a floating-point `Width`, a `BorderStyle`, and a `Color`.

`BoxModel` holds `Top`, `Right`, `Bottom`, and `Left` `Length` values. `IsZero() bool` is true only when every side has `Px == 0` and neither `IsAuto` nor `IsPercent` set. Any nonzero pixel value or flagged auto/percentage side makes it false.

## ComputedStyle

`ComputedStyle` is the aggregate of supported computed properties. Its exported fields are:

- Layout: `Display`, `Position`, `Overflow`, `Width`, `Height`, `MinWidth`, `MaxWidth`, `Top`, `Left`, `Right`, and `Bottom`.
- Box spacing: `Margin` and `Padding`, each a `BoxModel`.
- Borders: `BorderTop`, `BorderRight`, `BorderBottom`, and `BorderLeft`.
- Colors and stacking: `BackgroundColor`, `Color`, and integer `ZIndex`.
- Flex layout: `FlexDirection`, `JustifyContent`, `AlignItems`, `FlexGrow`, and `FlexShrink`.
- Grid layout: `GridTemplateColumns` and `GridTemplateRows` as slices of `Length`.
- Spacing and corners: `Gap` and the four corner radius lengths `BorderRadiusTopLeft`, `BorderRadiusTopRight`, `BorderRadiusBottomRight`, and `BorderRadiusBottomLeft`.

`Default() ComputedStyle` returns the package's baseline supported defaults:

| Field | Default |
|---|---|
| `Display` | `DisplayBlock` |
| `Position` | `PositionStatic` |
| `Overflow` | `OverflowVisible` |
| `Width`, `Height` | `Auto()` |
| `BackgroundColor` | `Transparent()` |
| `Color` | `RGB(30, 30, 30)` |
| `FlexDirection` | `FlexDirectionRow` |
| `JustifyContent` | `JustifyFlexStart` |
| `AlignItems` | `AlignStretch` |
| `FlexGrow`, `FlexShrink` | `0`, `1` |
| `Gap` | `Px(0)` |
| Other fields | Go zero values (including zero lengths, unset borders/colors, nil grid slices, and `ZIndex == 0`) |

The returned value is an ordinary struct and can be modified by callers. No validation, cascading, inheritance, or CSS parsing is performed by `Default` or these data types.
