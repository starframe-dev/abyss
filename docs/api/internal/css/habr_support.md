# CSS Habr Support

This package file adds stylesheet parsing for ordinary CSS rules and nested `@media` / `@supports` blocks, plus substitution of CSS custom-property references. It is part of the internal `css` package.

## Public API

### `ParseStylesheet(source string) []Rule`

Parses stylesheet text and returns the collected `Rule` values. Each ordinary rule with at least one parsed declaration becomes a `Rule` containing its selectors, declarations, and enclosing media/supports conditions.

- Selectors are split on commas and whitespace-trimmed.
- Nested media conditions are combined with ` and `; nested supports conditions are combined the same way.
- `@media` and `@supports` matching is case-insensitive. The condition text is retained from the input, trimmed.
- Other at-rules are skipped, including their contents.
- Malformed or incomplete blocks stop parsing at the point where a closing brace cannot be found.

Declarations are delegated to `ParseInline`, so declaration parsing behavior is determined by that function. Rules with no parsed declarations are omitted.

### `ResolveCustomProperties(value string, properties map[string]string) (string, bool)`

Substitutes `var(...)` references using the supplied property-name-to-value map. Returns the resulting string and a success flag.

- A known variable is replaced with its mapped value.
- A reference with a comma uses the text after the first comma as a fallback when the variable is absent; the fallback is trimmed.
- An absent variable without a fallback, or a `var(` without a closing `)`, returns `false` and leaves the unresolved reference in the returned value.
- Substitution proceeds iteratively, allowing replacement values to contain further references, with a maximum of 32 iterations. The result is successful only if no `var(` remains after that limit.

The implementation recognizes the first closing parenthesis after `var(` and does not parse nested function syntax or CSS comments. A fallback containing another `var(...)` can be resolved on a subsequent iteration.

## Implementation details

The unexported `parseStylesheetBlock` walks source bytes, skipping whitespace and semicolons between constructs. It locates each block header up to `{`, finds the matching closing brace, then either recurses into supported conditional at-rules or parses a declaration block. The current media and supports strings are passed recursively and copied onto emitted rules.

`matchingBrace` counts nested braces and ignores brace characters inside single- or double-quoted strings. A quote is considered escaped when immediately preceded by a backslash. It returns the index of the brace that closes the supplied opening brace, or `-1` if no match exists.

## Types and dependencies

The exported functions use the package's `Rule` type, defined elsewhere in the `css` package. Its relevant fields are `Selectors`, `Declarations`, `MediaQuery`, and `SupportsQuery`. Inline declaration parsing is also provided elsewhere in the package via `ParseInline`.
