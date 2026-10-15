# `internal/layout`

Package `layout` builds a geometric layout tree from a DOM tree and computed CSS styles. It implements a deliberately compact CSS box-model and text-flow layout for the renderer, including block, inline, flex, grid, table, image, and selected form-control behavior.

## Public API

### `ViewportWidth`

```go
const ViewportWidth = 800
```

Default viewport width in CSS pixels. `Build` accepts an explicit width; callers should use this constant when they need the package's default.

### `Rect`

```go
type Rect struct {
    X, Y, Width, Height float64
}
```

A rectangle in CSS-pixel coordinates. The rectangles on a `LayoutBox` describe the content, padding, border, and margin areas respectively. `Content` coordinates are absolute layout coordinates despite the comment's historical wording that calls them relative to the parent content area.

### `TextLine`

```go
type TextLine struct {
    Text string
    X, Y float64
}
```

A text fragment positioned for rendering. Text measurement and wrapping use terminal-cell widths from `go-runewidth` and pixel dimensions from `internal/cellmetrics`.

### `LayoutBox`

```go
type LayoutBox struct {
    Node      *dom.Node
    Parent    *LayoutBox
    Style     css.ComputedStyle
    Content   Rect
    Padding   Rect
    Border    Rect
    Margin    Rect
    Children  []*LayoutBox
    TextLines []TextLine
    AsciiCells [][]asciirender.Cell
}
```

One laid-out DOM node. `Node` and `Style` retain the source node and its computed style; `Parent` and `Children` form the layout tree. The four rectangles describe box-model geometry. `TextLines` contains positioned text generated during layout. `AsciiCells` is reserved for decoded image content and is populated by `render.Render`, not by this package's layout routines.

### `Build`

```go
func Build(doc *dom.Node, styles map[*dom.Node]css.ComputedStyle, viewportWidth float64) *LayoutBox
```

Creates a layout tree and computes its geometry using the supplied viewport width. Styles are looked up by DOM-node pointer. Missing entries use `css.Default()`. Text nodes are not generally independent boxes; non-whitespace text adjacent to element children is retained through anonymous inline boxes. `style`, `script`, and `head` elements, and nodes with `display: none`, are omitted. Returns `nil` if the document root is omitted; otherwise returns the laid-out root.

### `Layout`

```go
func Layout(box *LayoutBox, containerWidth float64)
```

Computes geometry in place for an existing layout tree. The root begins at the origin, and `containerWidth` is the available width used to resolve auto and percentage dimensions. This function does not return an error; callers are responsible for passing a valid box and suitable dimensions.

## Layout behavior

### Tree construction and box model

The builder preserves element nodes that are not filtered, assigns parent pointers, and initializes each box with an empty child slice. Text-only containers are laid out from their descendants' inline text. For mixed element/text content, meaningful sibling text is wrapped in an anonymous inline box that inherits text styling but resets width, height, margin, padding, borders, and background to defaults, preventing the parent box model from being applied repeatedly.

For each box, layout resolves width against the containing width. Auto width fills the container less horizontal margins, clamped to zero. Images with auto CSS width may instead use a positive numeric `width` attribute. Auto-width push-button inputs (`submit`, `button`, or `reset`) size to their label plus horizontal padding. Content width is the resolved width minus padding and borders, also clamped to zero. Rectangles are built outward from the border position; root positioning starts from the origin. Margins and padding resolve against the containing width.

Height is derived from content layout unless a non-auto CSS height is specified. The resulting padding, border, and margin heights include their respective vertical padding, borders, and margins. Absolute positioning uses the nearest relative or absolute ancestor. Explicit `left`/`top` take precedence; `right` and `bottom` position the box when the corresponding `left`/`top` is auto.

### Content flow

Children are dispatched according to node and display type:

- Images use image placeholder/text layout.
- Tables delegate to the table layout routine.
- Flex and grid boxes delegate to their corresponding routines.
- Block and inline-block boxes use block flow, with inline content used when there are no block-flow children.
- Other boxes use inline flow.

Block flow stacks ordinary children vertically. Runs of button-like controls, text-only links, and visible text nodes share rows and wrap when the available width is exhausted. Whitespace-only nodes between row items do not split the run. Rows honor center alignment on `<center>` or `align="center"`, and otherwise the box's text alignment; adjacent text links receive an explicit gap.

Inline flow joins visible text, splits it into words, and wraps using terminal display widths. Each line is positioned at the box's content origin with a vertical step of `cellmetrics.HeightPx`. Script and style descendants are excluded. Hidden inputs contribute no content. Inputs draw their value, falling back to placeholder; empty inputs and textareas retain one visible line. Search inputs (including `name="q"`) with no label receive simple ASCII search/control icons.

Images display a wrapped `[image]` label based on their `alt` attribute, defaulting to `image`. A positive numeric height attribute sets the image content height, subject to a minimum of 48 CSS pixels; otherwise height is based on wrapped lines.

## Important implementation details

- Layout is renderer-oriented rather than a complete browser CSS implementation. It supports the cases explicitly implemented here and relies on `css.ComputedStyle` for parsed style values.
- Width and text calculations use `runewidth.StringWidth`, so wide Unicode characters consume multiple terminal cells.
- Percentage margins and padding are resolved against the containing width, including vertical values.
- `TextLines` are generated during layout, while actual ASCII image cells are attached later by the rendering layer.
- Internal helpers implement row classification, form labels, text collection, wrapping, and positioned-ancestor lookup; they are not exported API.
