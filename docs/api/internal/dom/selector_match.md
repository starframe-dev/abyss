# Selector Matching

`internal/dom/selector_match.go` implements the matching side of the DOM selector engine. It evaluates parsed selector structures against `Node` values, including simple selectors, combinators, attribute operators, and a defined set of pseudo-classes.

## API

All functions and types used in this file are unexported; the file does not add exported Go API. The matcher operates on selector and node types declared elsewhere in package `dom`:

- `Node` represents a DOM node, with `Type`, `TagName`, `Attributes`, `Children`, `Parent`, and text content used by the matcher.
- `ComplexSelector` supplies an ordered `Parts` slice.
- `SelectorPart` holds a tag, ID, class list, attribute selectors, pseudo-classes, and a combinator.
- `AttrSelector` carries an attribute name, operator, and comparison value.
- `PseudoSelector` carries the pseudo-class name and optional argument.

The package-level selector API (such as `MatchesSelector`) is defined elsewhere. This implementation calls `MatchesSelector` to evaluate selector arguments for `:not()`, `:is()`, and `:where()`.

## Matching behavior

### Complex selectors and combinators

`matchesComplex` rejects a selector with no parts. Otherwise it starts matching at the final part, then `matchPart` walks backward through the selector according to that part's combinator:

| Combinator | Meaning |
| --- | --- |
| ` ` or empty | An ancestor must match the preceding part. Any ancestor may satisfy it. |
| `>` | The direct parent must match. |
| `+` | The immediately preceding element sibling must match. |
| `~` | Any preceding element sibling may match. |

A part must first pass `matchesSimple`. The first part succeeds once its own simple selector matches. Non-element nodes cannot satisfy simple selectors.

Sibling lookup scans the parent's `Children` in reverse and ignores non-element nodes. It relies on the node being present in its parent's child slice.

### Simple selectors

`matchesSimple` applies all populated constraints conjunctively:

- **Tag:** `*` and an empty tag impose no tag restriction. Other tag names compare case-insensitively.
- **ID:** compares with the `id` attribute using exact string equality.
- **Classes:** each requested class must occur as a whitespace-delimited token in the `class` attribute.
- **Attributes:** every attribute selector must match.
- **Pseudo-classes:** every pseudo-class must match.

### Attribute operators

`matchesAttr` implements these operators:

| Operator | Behavior |
| --- | --- |
| empty | Attribute exists, regardless of its value. |
| `=` | Exact value equality. |
| `~=` | Attribute value contains the requested value as a whitespace-delimited token. |
| `|=` | Exact equality or a value beginning with the requested value followed by `-`. |
| `^=` | Value has the requested prefix. |
| `$=` | Value has the requested suffix. |
| `*=` | Value contains the requested substring. |

An unknown operator does not match. String comparisons are case-sensitive. For the existence operator, an attribute present with an empty value still matches.

### Supported pseudo-classes

Pseudo-class names are normalized to lowercase before lookup. The matcher recognizes:

- Structural: `:first-child`, `:last-child`, `:only-child`, `:first-of-type`, `:last-of-type`, `:only-of-type`, `:nth-child()`, `:nth-of-type()`, `:nth-last-child()`, and `:nth-last-of-type()`.
- Selector arguments: `:not()`, `:is()`, and `:where()`, delegated to `MatchesSelector` with the pseudo argument.
- Content/state: `:empty`, `:checked`, `:disabled`, `:enabled`, and `:selected`.
- `:focus`, which always returns false because this matcher has no focus-state model.
- `:root`, which matches nodes whose type is `DocumentNode`.

Structural positions and `nth` expression evaluation are delegated to helpers such as `childIndex`, `lastChildIndex`, and `matchNth` defined elsewhere in the package. This file does not define their expression grammar.

State pseudo-classes use `GetAttr`: checked, disabled, and selected match when the returned value is non-empty; enabled matches when the disabled value is empty. Consequently, behavior follows `GetAttr`'s missing-versus-empty handling rather than a separate browser form-control state model.

`:empty` considers a node non-empty if it has an element child or a text child containing non-whitespace text. Other child node types and whitespace-only text do not make it non-empty.

## Implementation details

Pseudo-class functions are held in a package-level map initialized lazily with `sync.Once`. Initialization is safe for concurrent calls; after initialization matching performs a lookup by normalized pseudo name. Unrecognized pseudo-classes return false.

Complex matching is recursive and may search multiple ancestors or preceding siblings for descendant and general-sibling combinators. The matcher evaluates the in-memory `Node` tree and does not mutate it.
