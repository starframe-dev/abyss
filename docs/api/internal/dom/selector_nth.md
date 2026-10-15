# Nth Selector Helpers

This file implements internal helpers used to determine an element's position among its parent's children and to match that position against an `an+b`-style expression. All functions are unexported and form no public Go API of their own.

## Behavior

### `childIndex`

```go
childIndex(n *Node, ofType bool) int
```

Returns the 1-based position of `n` among its parent's element children. When `ofType` is true, only element children whose tag names match `n.TagName` are counted; matching is case-insensitive. Non-element children are always ignored. Returns `0` if `n` is nil, has no parent, or is not found in the parent's children.

### `lastChildIndex`

```go
lastChildIndex(n *Node, ofType bool) int
```

Uses the same filtering and failure behavior as `childIndex`, but scans from the end of the parent's child list. The returned position is therefore counted backward from the last matching element child.

### `matchNth`

```go
matchNth(arg string, idx int) bool
```

Trims surrounding whitespace and lowercases `arg`, then checks whether `idx` matches the expression. Supported forms are:

- `even` and `odd` (based on remainder modulo 2)
- An integer, matched exactly
- An `an+b` expression, such as `2n+1`, `3n`, `-n+3`, or `n`

Invalid or empty expressions return false. A zero coefficient (`a == 0`) matches only `b`. For nonzero coefficients the implementation requires the index to be on the appropriate side of `b` (`idx >= b` for positive `a`, `idx <= b` for negative `a`) and then checks divisibility of `idx-b` by `a`.

The literal expression `0` is handled specially by `matchNth` and matches only index `0`. This differs from ordinary CSS element positions, which are 1-based; callers should account for that behavior. `even` and `odd` likewise evaluate the supplied index directly.

### `parseNth`

```go
parseNth(s string) (a, b int, ok bool)
```

Parses the formula into its coefficient and offset so `matchNth` can evaluate it. Spaces are removed before parsing. The parser accepts a missing coefficient as `1`, a coefficient of `-` as `-1`, and an omitted offset as `0`. It also recognizes `0` as a valid formula `(0, 0, true)`. Inputs without `n`, or with invalid integer components, fail with `ok == false`.

## Implementation details and constraints

- Position helpers traverse `n.Parent.Children` directly and compare nodes by pointer identity.
- Only nodes with `Type == ElementNode` contribute to positions.
- `ofType` filtering uses `strings.EqualFold` on tag names.
- `matchNth` first attempts formula parsing and falls back to parsing a plain integer when the input is not a valid formula.
- Arithmetic uses Go `int` values; the helpers do not perform overflow checks or validate that `idx` is positive.
