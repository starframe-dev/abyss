# DOM selector API

This package implements a small CSS-like selector parser and matcher for the `dom.Node` tree. It supports selector groups, compound selectors, combinators, attribute selectors, and pseudo selectors as represented by its exported data types. It is not a complete browser CSS selector implementation.

## Public API

### `SelectorGroup`

```go
type SelectorGroup struct {
    Selectors []ComplexSelector
}
```

A comma-separated collection of complex selectors. A selector group matches when any member matches.

### `ComplexSelector`

```go
type ComplexSelector struct {
    Parts []SelectorPart
}
```

A sequence of simple selector parts. The first part has no preceding combinator; subsequent parts record how they relate to the preceding part.

### `SelectorPart`

```go
type SelectorPart struct {
    Tag        string
    ID         string
    Classes    []string
    Attributes []AttrSelector
    Pseudos    []PseudoSelector
    Combinator string
}
```

A compound selector. An empty `Tag` means no tag restriction; `"*"` denotes the universal selector. `Combinator` is one of `""`, `" "` (descendant), `">"` (child), `"+"` (adjacent sibling), or `"~"` (general sibling). IDs, classes, attributes, and pseudos are stored independently, and may appear in any order in the source.

### `AttrSelector`

```go
type AttrSelector struct {
    Name  string
    Op    string
    Value string
}
```

An attribute constraint. Supported operators are presence (`""`), exact (`"="`), whitespace-token (`"~="`), language-prefix (`"|="`), prefix (`"^="`), suffix (`"$="`), and substring (`"*="`). Parsed names are converted to lowercase.

### `PseudoSelector`

```go
type PseudoSelector struct {
    Name string
    Arg  string
}
```

A pseudo-class-like constraint, with an optional parenthesized argument. Double-colon pseudo-element syntax is parsed as a single-colon name for matching.

### `ParseSelectorGroup`

```go
func ParseSelectorGroup(s string) (*SelectorGroup, error)
```

Trims surrounding whitespace and parses comma-separated selectors. An empty or whitespace-only input returns an empty group without an error. Empty members from commas are skipped. Invalid selector syntax returns an error; callers should not assume a partially parsed group is returned.

### `MatchesSelector`

```go
func MatchesSelector(n *Node, selector string) bool
```

Returns whether `n` matches any selector in the input group. Nil nodes and nodes that are not `ElementNode` return false. Parse errors also return false, so this convenience API does not expose syntax errors.

### `QueryAllSelector`

```go
func QueryAllSelector(root *Node, selector string) []*Node
```

Returns matching descendant elements in tree traversal order. The root node itself is not tested. Invalid selector input returns `nil`; a valid selector with no matches returns an empty result slice.

### `QuerySelectorSelector`

```go
func QuerySelectorSelector(root *Node, selector string) *Node
```

Returns the first matching descendant as determined by `QueryAllSelector`, or `nil` if there is no match or parsing fails.

## Selector syntax and parsing behavior

- Commas divide selector groups, except inside brackets, parentheses, or quoted strings.
- A compound selector can include a tag/universal selector, `#id`, `.class`, `[attribute]` or `[attribute operator value]`, and `:pseudo` / `:pseudo(argument)` forms.
- Whitespace between compounds means descendant. Explicit `>`, `+`, and `~` combinators may have surrounding whitespace.
- Identifiers are limited to Unicode letters and digits plus `-` and `_`; escaping is not implemented.
- Attribute values may be quoted with single or double quotes, or supplied as an identifier. Attribute names are lowercased while values are retained as parsed.
- Parenthesis nesting in pseudo arguments is tracked. Commas in nested or quoted contexts do not split the outer group.

## Matching and implementation notes

Matching is performed by internal `matchesComplex` logic against the DOM nodes; parsed structures are passed through without a separate compiled representation. Querying recursively visits each element child, tests it against group members, appends it at most once even if several group members match, and then visits its descendants. Non-element children are skipped as candidates.

The parser is intentionally lightweight and does not implement the full CSS grammar: escapes, namespace selectors, flags on attribute selectors, and many browser-specific grammar rules are not provided. `MatchesSelector` and query helpers convert parse errors into false/empty results; use `ParseSelectorGroup` when the distinction between invalid syntax and no match matters.
