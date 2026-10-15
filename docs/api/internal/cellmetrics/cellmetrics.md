# `internal/cellmetrics`

The `cellmetrics` package defines the shared dimensions of a terminal cell in CSS pixels. Layout, painting, and ASCII sampling can use these constants to stay aligned.

## Public API

| Constant | Value | Meaning |
| --- | ---: | --- |
| `WidthPx` | `10` | Width of one terminal cell, in CSS pixels. |
| `HeightPx` | `20` | Height of one terminal row, in CSS pixels. |

Both constants are untyped integer constants and can be used in compile-time expressions as well as ordinary arithmetic.

## Behavior and implementation

The package has no runtime behavior or mutable state. Its dimensions provide a common basis for mapping terminal coordinates to CSS-pixel dimensions:

- A viewport with `N` columns is `N * WidthPx` pixels wide.
- Text rows are separated vertically by `HeightPx` pixels.

The values are defined centrally so consumers that calculate layout, paint cells, or sample ASCII can share the same metrics.
