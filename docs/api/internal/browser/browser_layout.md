# Browser Layout Box Mapping

This file documents `internal/browser/browser_layout.go`, which converts a layout tree into a lookup from DOM nodes to absolute content rectangles.

## Behavior

The mapper starts at the root layout box with a zero `(x, y)` offset and recursively visits its descendants. For each box with a non-nil `Node`, it:

1. Adds the box's `Content.X` and `Content.Y` to the accumulated parent offset.
2. Stores a `js.Rect` for that node, using the resulting absolute position and the content width and height.
3. Visits each child with the computed absolute position as the child's offset.

A nil box, or a box whose `Node` is nil, is skipped together with its descendants. The returned map is initialized even if the root is nil. Since map keys are node pointers, repeated occurrences of the same node pointer overwrite the earlier rectangle; traversal order follows each box's `Children` order.

The positions are absolute within the coordinate system represented by the layout tree. The calculation uses `Content` geometry, not an outer border or margin rectangle.

## API

### `BuildLayoutBoxMapForTest`

```go
func BuildLayoutBoxMapForTest(root *layout.LayoutBox) map[*dom.Node]js.Rect
```

Exported test-facing wrapper around the package-private `buildLayoutBoxMap`. It returns the node-to-rectangle mapping for the supplied layout tree. The API exposes pointers to `dom.Node` as map keys and `js.Rect` values.

## Implementation details

- `buildLayoutBoxMap` allocates the result map and delegates traversal to `buildLayoutBoxRecursive` with offsets `(0, 0)`.
- `buildLayoutBoxRecursive` accumulates offsets recursively; each child's content coordinates are treated as relative to its parent content origin.
- Nil handling is performed before reading box fields, preventing dereferences of nil boxes and avoiding traversal through a nil-node box.
- The recursive helper is unexported. The only exported symbol in this file is explicitly named as a test helper.

## Dependencies

The implementation uses:

- `internal/layout` for `LayoutBox` and its `Content` and `Children` fields.
- `internal/dom` for node pointers.
- `internal/js` for the `Rect` result type.
