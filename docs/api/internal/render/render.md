# `internal/render`

Package `render` draws a layout tree into a pixel framebuffer. It paints box backgrounds, borders, and raster images converted to colored ASCII cells. The resulting framebuffer can then be sampled or otherwise consumed by the caller.

## Public API

### `ImageProvider`

```go
var ImageProvider func(src string) image.Image
```

Optional package-level callback used to resolve an image source URL. It returns a decoded `image.Image`, or `nil` when unavailable. Callers are expected to configure it during startup. If it is unset, has no source URL, or returns `nil`, image quantization is skipped.

### `Render`

```go
func Render(tree *layout.LayoutBox, viewportWidth int) *framebuffer.Framebuffer
```

Renders `tree` and returns a new framebuffer whose width is `viewportWidth` and whose height is determined from the tree's margin-box extents. A nil tree returns a 1-by-1 framebuffer.

Rendering proceeds as follows:

1. Collect the root and all descendants in preorder.
2. Find negative minimum margin coordinates and shift all layout geometry and text-line coordinates so these minima are nonnegative.
3. Determine the maximum vertical extent and allocate the framebuffer.
4. Clear it with the root background color when set, otherwise white (`RGB(255, 255, 255)`).
5. Stable-sort boxes by ascending z-index and draw each box. Equal z-index boxes retain preorder, so descendants paint after their ancestors.

The input layout tree is modified in place by the coordinate shift. Image drawing may also populate `AsciiCells` and clear image placeholder `TextLines` on layout boxes.

## Drawing behavior

Each box is processed in paint order:

- A set background color paints a rounded rectangle using the box border bounds and the four resolved corner radii.
- Otherwise, `input` and `textarea` nodes with positive border height receive a light gray rounded surface (`RGB(241, 243, 244)`), with radius equal to half the border height.
- Otherwise, `img` nodes are quantized and drawn as ASCII cells when image data is available. If unavailable, non-absolutely-positioned images receive a neutral rounded placeholder (`RGB(226, 232, 240)`). Absolutely positioned images without quantized content are omitted so the placeholder cannot obscure underlying content.
- Configured borders are painted after the background or image, using each side's width and color. Border drawing is selected when any side's style is not `BorderNone`.

## Image quantization and cells

For an image box, the renderer first reuses `LayoutBox.AsciiCells` if already populated. Otherwise it obtains the node's `src`, asks `ImageProvider` for the decoded image, and quantizes it through `asciirender.Quantize`.

The grid dimensions come from the content dimensions divided by `cellmetrics.WidthPx` and `cellmetrics.HeightPx`, respectively, with each dimension clamped to at least one cell. The quantizer background is the box background when set, or `RGB(226, 232, 240)` otherwise. The resulting grid is cached on the box. Since layout may have seeded `TextLines` with an image alt-text placeholder, those lines are cleared when real ASCII cells are generated to prevent the placeholder from overlaying the image.

`drawAsciiCells` maps grid cells across the box's border rectangle. For cells with a set foreground color, it fills the corresponding framebuffer slice; cells without a foreground color do not paint a slice. Painting is clipped on the right edge to framebuffer width.

## Implementation notes

- The framebuffer height is computed recursively from each box's `Margin.Y + Margin.Height` and descendant results. The implementation does not apply a separate viewport-height limit.
- Negative offsets are calculated from margin X and Y positions, starting with zero, and applied recursively to content, padding, border, margin, and text-line coordinates.
- `collectBoxes`, `calculateOffset`, `calculateHeight`, `shiftTree`, `drawBox`, `quantizeImageBox`, and `drawAsciiCells` are package-private helpers; they are not part of the exported API.
- `ImageProvider` is shared package state. Configure it before rendering rather than changing it concurrently with calls to `Render`.
