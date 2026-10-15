# DOM Serialization

This module implements HTML and text serialization helpers for the JavaScript-facing DOM, plus replacement of a node from an HTML fragment. The functions are package-private; they are used internally by the `internal/js` package and are not exported Go API.

## Behavior

### Inner HTML and text content

`getInnerHTML(id NodeID) string` looks up the node under `nodeMu` and serializes each child in order. The selected node itself is not serialized. A missing ID produces an empty string.

`getTextContent(id NodeID) string` looks up the node under the same mutex and concatenates descendant text in tree order. For a text node, `collectText` returns its text verbatim; for other node types it recursively concatenates children. A missing ID produces an empty string. Text content is not HTML-escaped.

### HTML serialization

`serializeNode` serializes text nodes by escaping their text and element nodes as a start tag, attributes, children, and end tag. Attributes are emitted in lexicographically sorted key order, making output deterministic regardless of map iteration order. Attribute values are escaped and quoted only when non-empty; empty values are emitted as a bare attribute name.

Text and attribute escaping replaces `&`, `<`, `>`, and `"` with their corresponding HTML entities. Other characters are left unchanged. Element tag and attribute names are emitted as stored, without normalization or escaping. Nodes other than text and element nodes contribute no output.

An element in `voidElements` with no children is serialized with the ` />` ending and no closing tag. The recognized names are `area`, `base`, `br`, `col`, `embed`, `hr`, `img`, `input`, `link`, `meta`, `param`, `source`, `track`, and `wbr`. A void element that has children follows the ordinary paired-tag serialization path.

`outerHTMLString(nodeID NodeID) string` returns the node's `OuterHTML()` result. It briefly locks `nodeMu` to retrieve the node, then calls `OuterHTML` after releasing the lock. A missing or nil node returns an empty string.

### Replacing outer HTML

`setOuterHTML(nodeID NodeID, html string)` replaces the node with the children of `dom.ParseFragment(html)` at the same position in its parent. It does nothing if the node is missing, nil, detached (has no parent), parsing returns nil, or the node cannot be found in the parent's children.

On a successful replacement, the old node is detached, each inserted node is assigned the parent and passed to `assignIDsToNew`, and the CSS cache is invalidated with `css.InvalidateCache()`. An empty parsed fragment removes the target without inserting replacements.

## Synchronization and implementation notes

The node map and tree mutations use the package-level `nodeMu`. Inner HTML and text collection hold the mutex while recursively traversing. `outerHTMLString` and the initial lookup in `setOuterHTML` release the mutex before invoking node methods or parsing; `setOuterHTML` reacquires it for locating and modifying the parent’s child list. Parsing occurs before the mutation lock is acquired. Callers should account for these lock boundaries when coordinating concurrent DOM operations.

The serialization path builds output with `strings.Builder`; it does not perform browser-grade HTML validation or normalization. Escaping is intentionally limited to the four characters listed above.

## API reference

| Function | Visibility | Purpose |
| --- | --- | --- |
| `getInnerHTML(id NodeID) string` | Unexported | Serialize a node's children as HTML. |
| `getTextContent(id NodeID) string` | Unexported | Concatenate descendant text. |
| `collectText(n *dom.Node) string` | Unexported | Recursively collect text from a node. |
| `serializeNode(b *strings.Builder, n *dom.Node)` | Unexported | Append a node's HTML serialization. |
| `escapeHTML(s string) string` | Unexported | Escape the supported HTML-special characters. |
| `outerHTMLString(nodeID NodeID) string` | Unexported | Return a node's outer HTML, or empty string if absent. |
| `setOuterHTML(nodeID NodeID, html string)` | Unexported | Replace a node with a parsed HTML fragment. |
