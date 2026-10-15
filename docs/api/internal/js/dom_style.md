# DOM style bridge

`internal/js/dom_style.go` connects computed CSS and layout results to the JavaScriptCore-backed DOM bridge. It stores snapshots keyed by Go DOM nodes, resolves bridge node IDs to those nodes, and exposes geometry and computed style data to JavaScript.

## Public API

### `SetComputedStyles(styles map[*dom.Node]css.ComputedStyle)`

Replaces the computed-style snapshot used by the DOM bridge. The map is keyed by node pointers. The assignment is protected by `computedStylesMu`; the function does not merge with a previous snapshot or copy the map.

### `SetLayoutBoxes(boxes map[*dom.Node]Rect)`

Replaces the layout-box snapshot used by the `getBoundingClientRect` bridge. The map is keyed by node pointers and assigned under `layoutBoxesMu`, without merging or copying.

Both setters accept `nil`. A nil snapshot behaves like a missing result: computed-style lookup returns JavaScript `null`, and geometry lookup returns a zero `Rect`.

## Bridge behavior

### Bounding rectangle lookup

The internal `getBoundingClientRect(nodeID NodeID) Rect` acquires the layout snapshot lock, then resolves `nodeID` through the global node registry under `nodeMu`. It returns the associated rectangle only if the snapshot, node ID, node pointer, and map entry are all present. Every missing case returns `Rect{}`. This helper returns a Go value and does not construct a JavaScript object itself.

### Computed style lookup

The internal `getComputedStyle(nodeID NodeID) C.JSValueRef` resolves the node in the same manner and looks up its `css.ComputedStyle` in the current snapshot. If the snapshot, node, or style entry is missing, it returns JavaScript `null` using `eventCtx`.

For a successful lookup it creates a plain JavaScript object in `eventCtx` and populates these properties:

| Property | JavaScript value | Source |
| --- | --- | --- |
| `display` | String | Display enum converted to CSS-like text |
| `position` | String | Position enum converted to CSS-like text |
| `overflow` | String | Overflow enum converted to text |
| `width` | Number | `Width.Px` |
| `widthUnit` | String | `Width.String()` |
| `height` | Number | `Height.Px` |
| `heightUnit` | String | `Height.String()` |
| `color` | String | `Color.String()` |
| `backgroundColor` | String | `BackgroundColor.String()` |
| `marginTop`, `marginRight`, `marginBottom`, `marginLeft` | Number | Corresponding margin `Px` values |
| `paddingTop`, `paddingRight`, `paddingBottom`, `paddingLeft` | Number | Corresponding padding `Px` values |
| `borderTopWidth`, `borderRightWidth`, `borderBottomWidth`, `borderLeftWidth` | Number | Corresponding border widths |
| `flexDirection` | String | Flex direction converted to text |
| `justifyContent` | String | Justification converted to text |
| `alignItems` | String | Alignment converted to text |
| `flexGrow`, `flexShrink` | Number | Corresponding computed values |
| `zIndex` | Number | Integer converted to `float64` |
| `gap` | Number | `Gap.Px` |

The constructed object is returned through `jsObjToVal`. Numeric dimensions and spacing are exposed as pixel numbers; width and height additionally carry their original formatted unit strings in `widthUnit` and `heightUnit`.

## Enum-to-string mappings

The conversion helpers are internal and provide the bridge's supported textual representation:

- Display: `inline`, `inline-block`, `flex`, `grid`, or `none`; unrecognized/default values become `block`.
- Position: `relative` or `absolute`; other values become `static`.
- Overflow: `hidden` for `css.OverflowHidden`; otherwise `visible`.
- Flex direction: `column` for `css.FlexDirectionColumn`; otherwise `row`.
- Justification: `center`, `flex-end`, `space-between`, or `space-around`; other values become `flex-start`.
- Alignment: `center`, `flex-end`, or `stretch`; other values become `flex-start`.

## Implementation notes

The snapshot maps and node registry are shared bridge state. Each setter protects its map replacement with the corresponding mutex. Lookups hold the snapshot mutex while checking the map; they briefly acquire `nodeMu` to resolve the ID, then release it before accessing the snapshot entry. Callers should therefore publish coherent snapshots and avoid mutating a map after passing it to a setter, since the setter stores the map reference rather than making a defensive copy.

JavaScript property names and string values are created through JavaScriptCore C APIs, and properties use `kJSPropertyAttributeNone`. This file depends on the package's existing `eventCtx`, node registry, `NodeID`, `Rect`, and JavaScript value conversion helper; it does not define those types or perform CSS computation or layout.