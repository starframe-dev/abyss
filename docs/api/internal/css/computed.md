# Computed CSS styles

`internal/css/computed.go` computes the supported subset of CSS computed styles for DOM nodes. It gathers stylesheet rules, matches them to elements, applies the cascade, resolves custom properties, and returns a style map keyed by node pointers.

## Public API

| Function | Behavior |
| --- | --- |
| `Compute(doc *dom.Node) map[*dom.Node]ComputedStyle` | Computes styles at width `0`, with no external CSS. |
| `ComputeAtWidth(doc *dom.Node, viewportWidth float64) map[*dom.Node]ComputedStyle` | Computes styles for the supplied CSS-pixel viewport width. |
| `ComputeWithExternal(doc *dom.Node, externalCSS []string) map[*dom.Node]ComputedStyle` | Computes styles with additional CSS source strings at width `0`. |
| `ComputeWithExternalAtWidth(doc *dom.Node, externalCSS []string, viewportWidth float64) map[*dom.Node]ComputedStyle` | Combines external CSS and a viewport width. |
| `InvalidateCache()` | Clears the package's computed-style cache. Call after DOM changes made by JavaScript. |

The returned map associates each visited DOM node pointer with its `ComputedStyle`. This file defines no exported types; `ComputedStyle` and the rule/parser types are defined elsewhere in the `css` package.

## Computation and cascade

The implementation collects CSS from every `<style>` element found by `FindStyleTags`. A non-empty `media` attribute wraps that element's text in an `@media` rule before parsing. Each external CSS string is parsed and appended after embedded rules.

The DOM is traversed recursively. Each node starts with `Default()`; the parent's computed `Color` is copied when available. Custom properties are cloned from the parent so descendants inherit them without mutating the parent's variable map. Matching rules are screened by media and supports conditions, then each matching selector contributes the rule's declarations. Inline declarations parsed from the node's `style` attribute are included as well.

Declarations are stably sorted from lower to higher priority: non-`!important` before important, non-inline before inline, then lower selector specificity before higher specificity. Stable ordering preserves source order for ties. Custom properties are collected in this cascade order before ordinary declarations are applied, allowing `var()` to reference a custom property declared later in source order. Custom property names are normalized to lowercase and surrounding whitespace is removed. Ordinary declarations pass through `applyDeclaration`, which dispatches supported properties to their parsers and updates the relevant fields in `ComputedStyle`.

Supported property groups include box-model margins and padding, borders and radii, dimensions and grid track lists, positional offsets, foreground/background colors, and layout properties such as display, position, overflow, z-index, text alignment, and selected flex settings. Unsupported properties are ignored by the dispatch functions.

## Selector and conditional handling

Only element nodes match selectors. `:root` is handled specially and matches an element whose parent is a document node. An empty selector is treated as matching any element; other selectors are delegated to `dom.Node.Matches`.

Media queries are limited to 256 characters. Empty queries match; comma-separated alternatives use OR semantics. The implementation recognizes `all`, `screen`, optional `only`, `screen and ...`, `all and ...`, and parenthesized `min-width`/`max-width` features expressed in `px`. `print`, `not ...`, malformed features, unsupported feature names, or unparsable values do not match. All parsed width constraints in an alternative must pass.

Supports conditions also have a 256-character limit. The evaluator handles `or`, `and`, `not`, optional outer parentheses, and property/value tests for a limited set of `display` values (`block`, `inline`, `flex`, `grid`, `none`) and `position` values (`static`, `relative`, `absolute`). Unknown or malformed tests fail.

## Custom properties

`var()` references are resolved recursively, case-insensitively by variable name, and support a comma-separated fallback. A missing or cyclic variable uses its fallback when present; otherwise resolution fails and the entire ordinary property value becomes an empty string. Resolution is bounded by `maxVariableResolutionDepth` (64); malformed parentheses and excessive recursion likewise fail. Values without a `var(` token are returned unchanged.

## Cache and concurrency

A package-level mutex protects a single cached document pointer, style map, and viewport width. A result is reused only when the document pointer and width match, the cache is populated, and no external CSS was supplied. External-CSS computations are always recomputed, but their result is still stored in the cache and may consequently be reused by a later no-external-CSS call for the same document and width. DOM mutations are not detected automatically; callers must use `InvalidateCache` to prevent stale results. The cache retains only one document/result at a time.

The mutex serializes cache checks and the full computation. The implementation assumes a non-nil document and does not report errors; parsing and unsupported values are handled by the existing parser behavior rather than returned as errors.
