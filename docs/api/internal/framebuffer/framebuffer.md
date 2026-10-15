# `internal/framebuffer`

Package `framebuffer` provides a bounded RGB pixel buffer and raster operations used to render layout boxes. Pixels are stored as `css.Color` values from the internal CSS package.

## Public API

### Limits

| Constant | Value | Meaning |
| --- | ---: | --- |
| `MaxWidth` | 4096 | Maximum accepted framebuffer width. |
| `MaxHeight` | 4096 | Maximum accepted framebuffer height. |
| `MaxPixels` | 16,777,216 | Maximum number of allocated pixels. |

All three limits apply when constructing a buffer.

### `Framebuffer`

```go
type Framebuffer struct {
    Width  int
    Height int
    Pixels []css.Color
}
```

`Width` and `Height` describe the raster dimensions. `Pixels` is a row-major slice: the pixel at `(x, y)` occupies index `y*Width + x`. The fields and slice are exported, so callers can inspect or modify storage directly; direct modifications should preserve the expected dimensions and slice length.

### `New`

```go
func New(width, height int) *Framebuffer
```

Creates a zero-initialized buffer. If either dimension is non-positive, exceeds its respective limit, or the product would exceed `MaxPixels`, it returns a non-nil empty `Framebuffer` with zero dimensions and no allocated pixels. The product constraint is checked as `width > MaxPixels/height`, avoiding multiplication overflow after positive dimensions have been established.

### Pixel operations

```go
func (fb *Framebuffer) Set(x, y int, c css.Color)
func (fb *Framebuffer) Get(x, y int) css.Color
func (fb *Framebuffer) Clear(c css.Color)
```

- `Set` assigns a color only when the receiver is non-nil and the integer coordinates are inside the buffer. Out-of-bounds writes are ignored.
- `Get` returns the pixel at an in-bounds coordinate. A nil receiver or out-of-bounds coordinate returns `css.Transparent()`.
- `Clear` fills the entire pixel slice with the supplied color. A nil receiver is a no-op.

### Drawing operations

```go
func (fb *Framebuffer) FillRect(x, y, w, h float64, c css.Color)
func (fb *Framebuffer) DrawBorder(x, y, w, h, topW, rightW, bottomW, leftW float64,
    topC, rightC, bottomC, leftC css.Color)
func (fb *Framebuffer) DrawRoundedRect(x, y, w, h, rtl, rtr, rbr, rbl float64, c css.Color)
```

`FillRect` fills a floating-point rectangle, clipped to the framebuffer. Its covered integer pixel range is determined by flooring the start and ceiling the end, so any pixel cell intersecting the rectangle's coordinate span is included. Non-positive extents, NaN or infinite coordinates/extents, and ranges that do not intersect the buffer produce no writes. The clipping calculations constrain conversion to integer coordinates to a bounded range.

`DrawBorder` paints each positive-width edge using `FillRect`: top, right, bottom, then left. Consequently, corner overlap is resolved by the later edge in that order. Non-positive edge widths are skipped, and clipping is inherited from `FillRect`.

`DrawRoundedRect` first fills the full rectangle, then clips each requested corner by clearing pixels outside a circular arc. Corner radii are supplied in top-left, top-right, bottom-right, bottom-left order (`rtl`, `rtr`, `rbr`, `rbl`). Only positive radii invoke clipping. The implementation does not clamp radii to the rectangle dimensions; overlapping or oversized corner regions are processed sequentially, and transparent clearing may therefore affect pixels previously filled or clipped by another corner.

## Rasterization details

The internal `clippedRange` helper rejects non-finite starts/extents, non-positive extents, and non-finite end coordinates. It clamps the relevant range to `[0, maxSize]`, then uses floor for the lower bound and ceiling for the upper bound. Rounded corners use pixel-center sampling: each candidate pixel is tested at `(px+0.5, py+0.5)` against a circle of radius `r`; pixels whose squared distance exceeds `r*r` are set to transparent. The corner helper ignores nil receivers and invalid or non-positive radii/coordinates.

The package performs no color blending or anti-aliasing: drawing operations assign complete `css.Color` values, and corner clipping clears pixels to transparent.