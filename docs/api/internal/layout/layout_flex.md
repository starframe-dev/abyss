# Flex layout

This file implements the flex-container layout path used by the `layout` package. It lays out a `LayoutBox`'s children as a row or column, with support for CSS ordering, gaps, wrapping, growth, shrinking, and justification.

## API

All declarations in this file are unexported implementation helpers; it adds no public Go API. The principal entry point is `layoutFlex(box, contentWidth)`, called by the package's layout machinery. It returns the computed content height as a `float64`.

| Helper | Purpose |
| --- | --- |
| `orderedFlexChildren` | Applies stable CSS `order` sorting if any child has nonzero order. |
| `layoutFlexWrapped` | Partitions row items into lines and lays out each line. |
| `layoutFlexRow` | Places children on one row and returns the row/container height. |
| `layoutFlexColumn` | Places children vertically and distributes available height. |
| `flexBasis`, `inlineContentWidth` | Determine the starting width of a flex item. |
| `flexTotals`, `flexRemaining`, `flexShrinkWeight` | Aggregate basis and flex factors and calculate free space. |
| `flexJustification` | Computes leading offset or additional inter-item gap. |
| `flexMinContentWidth`, `fieldColumns` | Preserve minimum widths for nested form controls during row shrinking. |
| `layoutFlexChild`, `flexContainerSize` | Size/place an individual item and derive container height. |

## Behavior

`layoutFlex` returns zero for an empty child list. A column direction delegates directly to column layout. Otherwise it temporarily orders children using stable ascending `Style.Order` (preserving DOM order for ties), then chooses wrapped or single-row layout. The original child slice is restored after row processing, including wrapped processing.

### Rows and wrapping

A row resolves the container's `Gap` against `contentWidth`. Its initial available space is content width minus the gaps between children. Item bases are summed, and remaining space is distributed according to flex factors. Positive free space is apportioned by `FlexGrow`; negative free space is apportioned by each item's basis multiplied by `FlexShrink`, with widths clamped to zero before minimum-content constraints.

For `FlexWrap`, items are accumulated into lines using their flex basis and the resolved gap. A line is flushed before adding an item that would exceed `contentWidth`; each line is laid out as a row, and line heights are stacked with the same resolved gap. The wrapping decision is based on basis widths, before growth or shrinkage.

Each row item is laid out at the container content origin plus its running horizontal position. Its width is temporarily set to the computed pixel width for `layoutBox`, then restored. The item's minimum width is at least `flexMinContentWidth`.

### Columns

Column layout initially lays out children in sequence using their natural margin heights, adding the resolved gap between items. If the container height is automatic, that accumulated height is returned. For a specified height, positive remaining space is distributed proportionally to child `FlexGrow` values when any are present. Otherwise, `JustifyContent` can place the sequence at the start, center, or end, or add space between/around children. Column layout does not implement flex shrinking; its remaining space calculation is zero for column direction.

When growth is assigned, each growing child's height is temporarily replaced with a pixel value based on its content height plus its share, and then restored after relayout. The returned height is the accumulated natural height plus distributed growth; justification offsets and extra gaps position children within the specified height.

### Justification

`flexJustification` returns no offset or extra gap if remaining space is nonpositive, there are no children, or the total grow factor is positive. Otherwise:

- `center` offsets by half the remaining space.
- `flex-end` offsets by all remaining space.
- `space-between` divides it between multiple children.
- `space-around` assigns an equal extra gap per child and half that gap at either end.

Other values leave items at the start.

## Width basis and minimum content

`flexBasis` uses the child's resolved declared width when nonzero. Otherwise, it measures eligible inline content; if that measurement is unavailable or zero, it uses a default basis of 100 layout units.

`inlineContentWidth` only measures element nodes with children consisting exclusively of text nodes or text links. Text is collected recursively, whitespace is normalized, and display width is measured with `go-runewidth` and converted using `cellmetrics.WidthPx`. Consecutive link children include a two-cell separator. Any other child type makes the measurement unavailable (zero).

For row items, `flexMinContentWidth` walks descendants and keeps the widest eligible form field as a minimum. A `textarea` uses its `cols` attribute; a non-hidden `input` uses an explicit positive, non-percent CSS pixel width when present, otherwise its `size` attribute. Invalid, missing, or nonpositive column/size values default to 20 columns. Pixel widths are converted to cell columns using `cellmetrics.WidthPx`. Hidden inputs are excluded.

## Implementation notes and limitations

- CSS order sorting copies the slice only when at least one child has a nonzero order; stable sorting preserves the DOM sequence among equal order values.
- Row growth/shrink calculations use the flex basis, not measured post-layout dimensions. Shrinking is weighted by basis and `FlexShrink`.
- Wrapped rows use the same row algorithm independently, so growth and justification are resolved per line.
- The column path uses the container's resolved gap and height while passing `contentWidth` through to children; it does not calculate a separate column flex basis.
- Temporary changes to child widths/heights and to the container's child list are restored after layout, preventing these calculations from becoming persistent style/tree mutations.
