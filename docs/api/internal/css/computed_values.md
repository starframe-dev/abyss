# Computed CSS Values

`internal/css/computed_values.go` contains conversion helpers that turn CSS-like text values into the package's typed values (`Length`, `Color`, `Border`, and layout enums). Most helpers are unexported and are intended for use inside the `css` package. The exported API in this file is `Hex`.

## Public API

### `Hex(s string) Color`

Parses a hexadecimal color in `#RGB` or `#RRGGBB` form. Short-form nibbles are expanded (for example, `#abc` becomes channels 170, 187, and 204). Invalid length, missing `#`, or invalid hexadecimal digits produce `Transparent()` rather than an error.

## Parsing behavior

All other parsers below are package-private. They generally choose a default or sentinel value on unsupported input instead of returning an error. Callers should therefore not treat a returned value as proof that the input was valid.

### Lengths and numbers

- `parseLength` trims surrounding whitespace and recognizes `auto`, `px`, `%`, `rem`, `em`, `fr`, and unitless numbers (treated as pixels).
- `rem` and `em` are simplified to pixels using a fixed factor of 16; this parser does not resolve inherited or computed font sizes.
- `fr` values must be non-negative. Invalid values and unsupported units return `Auto()`.
- Numeric length inputs reject parse failures, NaN, and infinities. Finite values are clamped to ±1,000,000 (`maxCSSLength`). The converted `rem`/`em` result is clamped as well.
- `parseFloat` applies the same finite-number and clamp policy, returning zero for invalid input. `parseInt` trims whitespace and returns zero when conversion fails.
- `parseBorderWidth` strips a trailing `px` and parses a float, but does not apply the bounded-number checks used by `parseLength`.

### Box model

`parseBoxModel` accepts one through four whitespace-separated lengths using the CSS shorthand mapping:

| Values | Top | Right | Bottom | Left |
| --- | --- | --- | --- | --- |
| 1 | 1st | 1st | 1st | 1st |
| 2 | 1st | 2nd | 1st | 2nd |
| 3 | 1st | 2nd | 3rd | 2nd |
| 4 | 1st | 2nd | 3rd | 4th |

Any other number of tokens returns a zero-value `BoxModel`.

### Colors

`parseColor` trims whitespace and lowercases input. It recognizes the built-in names in `namedColors`, hexadecimal forms accepted by `Hex`, and comma-separated integer `rgb(r, g, b)` syntax. Unknown or malformed colors become transparent. The supported named-color table is deliberately limited; it is not the complete CSS color-name set.

`parseRgb` requires exactly three integer channels and the literal `rgb(` prefix and `)` suffix. Each channel is clamped to 0–255. It reports parse success internally; `parseColor` converts failure into transparent.

### Borders

`parseBorder` splits a shorthand value on whitespace. It recognizes `none` and `solid` as styles, numeric or `px` tokens as widths, and treats other tokens as colors. The initial style is `BorderNone`; unrecognized color tokens follow `parseColor`'s transparent fallback. This is a small supported subset of CSS border syntax, not a complete CSS tokenizer.

`parseBorderStyle` recognizes only `none` and `solid`; every other value defaults to `BorderNone`.

### Layout and alignment keywords

The keyword parsers return enum values defined elsewhere in the `css` package and use these fallbacks:

| Parser | Recognized values | Fallback |
| --- | --- | --- |
| `parseTextAlign` | `center`; `right`, `end` | left |
| `parseDisplay` | `block`, `inline`, `inline-block`, `inline-box`, `flex`, `-webkit-flex`, `-webkit-box`, `grid`, `none` | block |
| `parsePosition` | `static`, `relative`, `absolute` | static |
| `parseOverflow` | `hidden`, `visible` | visible |
| `parseFlexDirection` | `row`, `column` | row |
| `parseJustifyContent` | `flex-start`, `center`, `flex-end`, `space-between`, `space-around` | flex-start |
| `parseAlignItems` | `flex-start`, `center`, `flex-end`, `stretch` | stretch |

Only `parseTextAlign` normalizes case and surrounding whitespace. The other keyword parsers use exact string matching.

### Grid track lists

`parseLengthList` parses whitespace- or comma-separated length tokens and supports the simple form `repeat(count, item)`, expanding the item `count` times. The result is limited to `maxGridTracks` (256) entries. Repeat counts must be positive integers; malformed or non-positive counts add no entries. An unterminated `repeat(` stops parsing the remaining input. Each resulting track is parsed through `parseLength`, so invalid tracks become `Auto()`.

## Implementation notes

- These routines intentionally provide deterministic fallback values rather than error values; validation of full CSS grammar is outside their scope.
- `namedColors` and all parser helpers are unexported implementation details. The enum and value constructors used by them are defined elsewhere in the package.
- `isNumber` is a small `strconv.ParseFloat`-based predicate used by border parsing; unlike `parseFloat`, it does not reject NaN or infinity explicitly.
