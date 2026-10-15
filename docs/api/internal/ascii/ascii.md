# `internal/ascii`

Package `ascii` converts framebuffer and layout content into terminal-friendly text. It provides both plain-text output and ANSI true-color output, while preserving Unicode grapheme clusters and terminal cell widths.

## Public API

### `Cell`

Represents one terminal cell. It is also the intermediate representation used while creating output.

| Field | Type | Meaning |
|---|---|---|
| `Char` | `rune` | Leading rune displayed in the cell. A zero rune is treated as a space by the renderers. |
| `Extra` | `string` | Remaining runes in the same grapheme cluster, such as combining marks or joiners. |
| `Width` | `int` | Terminal-column width of the cluster. A non-positive value causes the renderer to calculate a width. |
| `Continuation` | `bool` | Marks a cell occupied by the second column of a wide grapheme; it is skipped during output. |
| `FG` | `css.Color` | Foreground color. |
| `BG` | `css.Color` | Background color. |

### `Render`

```go
func Render(fb *framebuffer.Framebuffer, tree *layout.LayoutBox, viewportWidth int) []string
```

Converts framebuffer and layout content to plain-text lines. The viewport width is clamped to zero or greater. A nil framebuffer or one whose pixel width is zero returns a one-element slice containing an empty string. Otherwise, the framebuffer is sampled into terminal cells, layout text and ASCII-cell overlays are applied, and rows are rendered up to the requested width. Plain output does not emit ANSI color sequences.

### `RenderANSI`

```go
func RenderANSI(fb *framebuffer.Framebuffer, tree *layout.LayoutBox, viewportWidth int) []string
```

Uses the same grid construction and overlays as `Render`, but emits ANSI escape sequences for set foreground and background colors. Lines are padded with spaces to the requested viewport width and end with an ANSI reset sequence. The same empty-framebuffer behavior and width clamping apply.

## Rendering behavior

The framebuffer is collapsed to a grid using the shared `cellmetrics` dimensions. Each terminal cell samples the framebuffer pixel at the center of that cell. The grid height rounds up to include a partial final pixel row, while its width uses only complete terminal-cell widths. A set sampled color is represented as a space with that color as its background; unset colors remain transparent.

Layout overlays are applied in two stages:

1. Pre-quantized ASCII cells from layout boxes are copied into the grid. Boxes with absolute positioning are skipped, so icons positioned over fields do not obscure the field in terminal output. Out-of-grid cells are ignored.
2. Text lines are collected recursively from the layout tree and placed according to pixel coordinates converted to terminal rows and columns. Text uses the box's style color, or dark gray (`RGB(30, 30, 30)`) when unset. Existing cell backgrounds are retained. Text that extends beyond a row is clipped; two-column graphemes mark their second cell as a continuation.

Text is processed as Unicode grapheme clusters, not individual bytes. A cluster's width comes from `uniseg`; a missing or non-positive cell width is calculated with `go-runewidth`. Non-printable characters are replaced with `�`. Cluster tails are restricted to printable runes plus selected zero-width joiners and variation selectors, preventing unsafe control characters from entering output.

Plain rendering emits each row without padding to the viewport width. When the next grapheme would exceed the width, it emits one space and stops that row. ANSI rendering likewise avoids splitting wide graphemes, substitutes a space if a grapheme cannot fit, and pads the remaining width. ANSI color changes use 24-bit foreground (`38;2`) and background (`48;2`) parameters; color state is reset at line end.

## Implementation notes

- The grid is an internal rendering representation; `makeGrid`, overlay helpers, and serialization helpers are not exported.
- `colorsEqual` compares both the set/unset state and RGB components.
- ANSI sequences are generated with `strings.Builder`; byte-valued RGB components are formatted directly as decimal digits.
- The package relies on `framebuffer`, `layout`, `cellmetrics`, and `css` types for source pixels, geometry, cell dimensions, and colors.
