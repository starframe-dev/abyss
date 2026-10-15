# `internal/asciirender`

Package `asciirender` decodes raster images and converts them into a bounded grid of terminal cells. Each cell uses a character selected from a luminance ramp and an averaged foreground color; `Flatten` can produce a plain-text preview without ANSI styling.

## Public API

### Constants

| Constant | Value | Meaning |
| --- | ---: | --- |
| `MaxImageWidth` | `8192` | Maximum accepted source width. |
| `MaxImageHeight` | `8192` | Maximum accepted source height. |
| `MaxImagePixels` | `8 * 1024 * 1024` | Maximum source pixel count. |
| `MaxOutputColumns` | `4096` | Maximum output columns. |
| `MaxOutputCells` | `1 << 18` | Maximum total output cells. |

### `Cell`

A `Cell` represents one terminal character cell:

- `Char rune` is the selected density character.
- `FG css.Color` is the averaged source color for the cell.
- `BG css.Color` is the configured or default background color.
- `IsHalfRow bool` marks a top-half pixel paired with the next row's bottom half. The current quantizer does not set this field.

The color type comes from `internal/css`.

### `Options`

`Options` configures [`Quantize`](#quantize):

| Field | Behavior |
| --- | --- |
| `Columns int` | Number of output columns. Values at or below zero select the default of 60; values above the maximum are capped. |
| `Rows int` | Number of output rows. Zero or a negative value derives a row count from image aspect ratio; positive values are capped by the total-cell limit. |
| `BGColor css.Color` | Underlay color. An unset color selects RGB `(226, 232, 240)`. |

### `Palette`

`Palette` is the exported rune slice used for density mapping, ordered from light/sparse to dark/dense:

```text
 .,:;i1tfLCG08@#
```

`pickRune` maps low luminance to later, denser characters and high luminance to earlier, lighter characters. Since `Palette` is an exported mutable slice, callers can change the mapping; it should not be modified concurrently with rendering.

### `Decode`

```go
func Decode(data []byte) (image.Image, error)
```

Decodes PNG, JPEG, and GIF (registered by this package), as well as formats registered with Go's standard `image` package. It first reads the image configuration and rejects non-positive dimensions, dimensions exceeding the width or height limits, or pixel counts exceeding `MaxImagePixels`; only then does it decode the pixels. Decode/configuration errors are returned. Dimension-limit failures return an error describing the rejected dimensions.

### `Quantize`

```go
func Quantize(img image.Image, opts Options) [][]Cell
```

Converts an image to rows of terminal cells. A nil image or an image whose bounds have invalid or excessive dimensions returns `nil`. Otherwise the result has the selected number of rows and columns, with each cell covering the corresponding proportional rectangle of source pixels. The default row count approximates terminal cell proportions of two source-image vertical units per column, derived as `Columns * sourceHeight / sourceWidth / 2`; it is clamped to at least one. Output is bounded by `MaxOutputColumns` and `MaxOutputCells`.

Each cell's RGB foreground is the average of its sampled pixels. Pixels with alpha are composited over white before averaging. Luminance uses the ITU-R BT.601 coefficients (`0.299R + 0.587G + 0.114B`) and selects `Char` from `Palette`. The background is `Options.BGColor` when set, otherwise RGB `(226, 232, 240)`.

### `Flatten`

```go
func Flatten(grid [][]Cell) string
```

Writes each cell's character in row order and appends a newline after every row. A zero `Char` is represented as a space. The output contains no ANSI color codes; colors remain available on the cells for a caller to render separately.

## Implementation notes

- The dimension check uses division (`width <= MaxImagePixels/height`) to avoid multiplying dimensions and risking integer overflow.
- `Decode` performs configuration inspection before full decoding to enforce limits before pixel allocation.
- Quantization uses image bounds, including non-zero bound origins, when sampling.
- Source pixels are assigned to proportional cell regions using integer division. If a region contains no pixels, its sample is black with zero luminance.
- Internal helpers (`imageDimensionsAllowed`, `sampleCell`, `pickRune`, `luminance`, and `rgba`) implement validation, sampling, and color conversion and are not part of the exported API.
