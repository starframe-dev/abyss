# `internal/dom`

Package `dom` builds and manipulates a lightweight DOM-like tree from HTML. It uses `golang.org/x/net/html` for tokenization, then applies its own tree-building, selector delegation, and serialization behavior. This is a simplified representation, not a full browser DOM or HTML5 tree-construction implementation.

## Parsing

```go
func Parse(htmlStr string) *Node
func ParseFragment(htmlStr string) *Node
```

`Parse` returns a `DocumentNode` root; `ParseFragment` returns a `DocumentFragmentNode` root without introducing an element wrapper. Both initialize empty attribute and child collections and use the same bounded tokenizer.

Parsing preserves the safely constructed prefix when it reaches a limit. The exported limits are:

| Constant | Value | Meaning |
| --- | ---: | --- |
| `MaxDOMNodes` | 100000 | Maximum counted nodes, including the root |
| `MaxDOMDepth` | 512 | Maximum stack depth, including the root |

The tokenizer's maximum token buffer is 1 MiB. A tokenizer error (including malformed input or an oversized token) ends parsing and returns the tree built so far. Text whitespace-only tokens are generally discarded, except under a `pre` element or when the preceding child is an element. This retains whitespace separating adjacent elements while omitting much formatting whitespace.

The parser recognizes HTML void elements (`area`, `base`, `br`, `col`, `embed`, `hr`, `img`, `input`, `keygen`, `link`, `meta`, `param`, `source`, `track`, and `wbr`) and does not push them on the open-element stack. It also performs limited implied closure for paragraphs, list items, definition terms/descriptions, options, table cells/rows/sections, and headings. This recovery is intentionally narrower than browser HTML parsing. Attribute names are normalized to lowercase; values are stored as returned by the tokenizer.

## Data model

### `NodeType`

`NodeType` is an integer identifying the node kind:

| Constant | Value | Meaning |
| --- | ---: | --- |
| `ElementNode` | 1 | Element |
| `TextNode` | 3 | Text |
| `CommentNode` | 8 | Comment |
| `DocumentNode` | 9 | Document root |
| `DocumentFragmentNode` | 11 | Fragment root |

### `Node`

```go
type Node struct {
    Type       NodeType
    TagName    string
    Attributes map[string]string
    Children   []*Node
    Text       string
    Parent     *Node
}
```

Fields are exported and directly mutable. Elements created during parsing have a lowercase tokenizer tag name, an attribute map, and a child slice. Parent pointers are set for parsed children. Text nodes store their content in `Text`. Document and fragment roots have no parent. Callers modifying the tree directly are responsible for maintaining consistent child and parent relationships.

## Selectors and traversal

```go
func (n *Node) QueryAll(selector string) []*Node
func (n *Node) QuerySelector(selector string) *Node
func (n *Node) Matches(selector string) bool
func (n *Node) Closest(selector string) *Node
```

These methods delegate selector evaluation to `QueryAllSelector`, `QuerySelectorSelector`, and `MatchesSelector` in the package. `QueryAll` returns matching descendants; `QuerySelector` returns the first matching descendant or `nil`. `Closest` checks the receiver first, then walks parents, returning the nearest match or `nil`; it safely returns `nil` for a nil receiver. Supported selector syntax is determined by those selector helpers.

```go
func (n *Node) FindStyleTags() []*Node
func (n *Node) FindLinkStylesheets() []*Node
```

`FindStyleTags` selects `style` elements. `FindLinkStylesheets` selects `link[rel=stylesheet]` elements.

## Attributes and tree mutation

```go
func (n *Node) GetAttr(name string) string
func (n *Node) RemoveChild(child *Node) *Node
func (n *Node) AppendChild(child *Node) *Node
func (n *Node) Clone(deep bool) *Node
```

- `GetAttr` lowercases the requested name and returns its value, or `""` when the map is nil or the key is absent. An absent attribute and an explicitly empty value therefore have the same result.
- `RemoveChild` removes the first child whose pointer equals `child`, clears that child's parent, and returns it. It returns `nil` if the child is not present.
- `AppendChild` appends the supplied pointer, sets its parent to the receiver, and returns the child. It does not detach it from an earlier parent or check for cycles; callers must manage those invariants.
- `Clone(false)` copies the node's type, tag, text, and attributes but not its children or parent. `Clone(true)` recursively copies descendants and assigns their parent pointers to the corresponding clones. Attribute maps are independent in either case.

## Text and serialization

```go
func (n *Node) TextContent() string
func (n *Node) OuterHTML() string
```

`TextContent` concatenates descendant text in child order. A text node returns its own `Text` directly; other node kinds concatenate children and do not include the node's own `Text` field.

`OuterHTML` returns an empty string for a nil receiver. Document and fragment nodes serialize as the concatenation of their children, without a wrapper. Text and comment nodes are both serialized by escaping their `Text` as ordinary text; comment delimiters are not emitted. Elements serialize with their tag, attributes, content, and an end tag. Attribute iteration follows Go map iteration order, so serialized attribute order is unspecified. Text and attribute values escape `&`, `<`, `>`, and `"`.

An empty element recognized as void serializes as `<tag />`; the serializer's void-element list excludes `keygen`, even though the parser's list includes it. Non-void elements always receive a closing tag. Since parsing does not populate element `Text`, element content normally comes from children; a manually assigned non-empty `Text` is serialized before the children.

## Notes

The package exposes a compact mutable tree, not a standards-complete DOM. In particular, parsing recovery is selective, comments are not given HTML comment syntax by serialization, and serializer output is a representation rather than a lossless round trip of the source.