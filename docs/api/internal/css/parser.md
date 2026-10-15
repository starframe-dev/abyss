# CSS Parser

Package `css` provides lightweight parsing helpers for inline styles and CSS stylesheets, plus a simple selector-specificity score. It is an intentionally limited parser rather than a complete CSS implementation.

## Public API

### `Declaration`

Represents one parsed property/value pair.

| Field | Type | Meaning |
| --- | --- | --- |
| `Property` | `string` | Trimmed property name. |
| `Value` | `string` | Trimmed value, excluding a recognized `!important` suffix. |
| `Important` | `bool` | Whether the value ended in `!important`, matched without regard to case. |

### `Rule`

Represents a selector list and its declarations, optionally nested under supported conditions.

| Field | Type | Meaning |
| --- | --- | --- |
| `Selectors` | `[]string` | Selectors split on commas and individually trimmed. |
| `Declarations` | `[]Declaration` | Declarations parsed from the rule block. |
| `MediaQuery` | `string` | Accumulated enclosing `@media` query, or empty. |
| `SupportsQuery` | `string` | Accumulated enclosing `@supports` query, or empty. |

### `ParseInline(style string) []Declaration`

Parses a style-attribute string into declarations. Semicolon-separated segments without a colon are ignored; the first colon divides property from value. Empty segments are skipped. Parentheses protect semicolons from splitting, so function values such as `url(...)` or `var(...)` remain intact.

### `ParseStyleSheet(css string) []Rule`

Removes block comments and parses ordinary rules, along with rules nested in `@media` and `@supports` blocks. Nested conditions are joined with ` and ` in their respective fields. Unsupported at-rules are skipped with their contents; malformed or incomplete blocks stop parsing at the point where a matching closing brace cannot be found. Recursion is limited to 32 nested at-rule levels.

### `Specificity(selector string) int`

Returns a weighted approximation of specificity: `100` per ID, `10` per class, attribute selector, or pseudo-class, and `1` per type selector or pseudo-element. This is a sortable score, not a full CSS specificity tuple, and does not implement all modern selector rules.

## Parsing behavior and limitations

- Declaration property and value whitespace is trimmed. A trailing case-insensitive `!important` is removed and reflected in `Important`.
- Stylesheet selectors are split literally on commas; commas inside functional pseudo-classes are not specially protected.
- Brace matching accounts for nested braces, quoted strings, backslash escapes, and parentheses. It does not parse the full CSS token grammar.
- Comment removal recognizes `/* ... */` without considering whether the markers occur inside quoted strings.
- Declaration splitting tracks parenthesis depth but not quotes or escaped characters. Colons are split at the first literal colon.
- Specificity tokenization handles whitespace and `>`, `+`, and `~` combinators, while preserving those characters inside attribute brackets. Escaped identifiers and functional pseudo-class specificity semantics are not fully modeled.
- Empty inputs and unparseable fragments produce empty results or are skipped rather than returning errors; the API exposes no diagnostics.

## Internal implementation

`parseRules` delegates to `parseRulesInMedia`, which scans for opening braces, finds the matching close brace, and recursively descends into supported conditional at-rules. `stripImportant`, `splitBySemicolon`, and the brace/selector helpers implement the parsing primitives. All parser state is local to each call; there is no package-level mutable state.
