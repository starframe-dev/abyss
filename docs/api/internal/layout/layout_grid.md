# Grid Layout

This file implements the internal grid-layout path for the `layout` package. It assigns children to a row-major grid, computes track dimensions and child coordinates, and returns the resulting grid content height.

## API and scope

All functions in this file are unexported implementation helpers; there is no public Go API defined here. The implementation operates on `LayoutBox` and CSS `Length` values from the surrounding layout and `internal/css` packages.

| Function | Responsibility |
| --- | --- |
| `layoutGrid` | Resolves grid configuration, lays out children, and returns total row height including gaps. |
| `resolveGridColumnWidths` | Resolves fixed, percentage, fractional, and auto column tracks against available content width. |
| `finiteNonNegative` | Checks whether a floating-point value is finite and non-negative. |
| `resolveGridFractionalRows` | Allocates available height among fractional row tracks while resolving percentages. |
| `placeGridChildren` | Convenience wrapper that derives row growth behavior from row definitions. |
| `placeGridChildrenWithTrackSizing` | Places children and grows eligible row tracks to fit their children. |
| `gridOffsetX`, `gridOffsetY` | Compute the origin of a column or row by summing preceding tracks and gaps. |
| `sumRowHeights` | Sums row heights and the gaps between them. |

## Layout behavior

`layoutGrid` resolves `Gap` against the content width. If no columns are specified, it supplies one column equal to the content width. If no rows are specified, it supplies one auto-sized row. Rows are marked growable when they are auto or fractional tracks, or percentage tracks under an auto-height container. It resolves fractional row sizes using the container's resolved height, resolves column widths, lays out children, and returns the final row-height sum.

Children are assigned in source order, filling columns before advancing to the next row:

```text
column = child index modulo number of columns
row    = child index divided by number of columns
```

The row list may be extended as children require additional rows. Rows beyond explicitly configured tracks begin at zero height and are growable. Each child's horizontal position is the content origin plus the widths and gaps of preceding columns; its vertical position similarly uses preceding row heights and gaps.

Before laying out a child, its width style is temporarily replaced with the resolved pixel width of its grid column. The layout call receives that width and the computed position. The original width style is restored afterward. A growable row is enlarged when the child's margin-box height exceeds the current row height. Non-growable tracks retain their resolved size even if a child is taller.

## Track sizing details

### Columns

Column widths are resolved as follows:

- Percentage tracks resolve against `contentWidth` and count toward fixed width.
- Fractional tracks contribute their non-negative finite fraction as a weight.
- Auto tracks initially receive zero width.
- Other track values use their pixel value and count toward fixed width.

The remaining width is `max(0, contentWidth - fixed widths - inter-column gaps)`. Fractional columns divide that remainder in proportion to their weights. If there are no positive fractional weights, auto columns remain zero. Invalid (negative, NaN, or infinite) fraction weights are treated as zero.

### Rows

Fractional rows divide the non-negative space remaining after non-fractional resolved heights and row gaps, proportionally to finite non-negative fraction weights. Percentage rows are converted to pixel lengths against the available height. If there are no positive fractional weights, percentages are still resolved and other row definitions are retained unchanged.

A subtlety in `layoutGrid` is that `rowCanGrow` is computed from the original row definitions before fractional and percentage rows are resolved. This preserves the decision about whether content can enlarge those tracks after their dimensions have been converted to pixels.

## Constraints and implementation notes

- Grid placement is strictly sequential and row-major; this code does not process explicit child placement, spanning, alignment, or implicit placement rules beyond adding rows as needed.
- Column auto sizing is not content-based in this implementation; auto columns start at zero.
- Row growth uses each child's `Margin.Height` and only enlarges a track; it does not shrink tracks.
- Gap is subtracted from available track space during sizing and added between tracks when calculating positions and total height.
- `sumRowHeights` returns zero for no rows. For one or more rows, it adds exactly `len(rows)-1` gaps.
- The helper `placeGridChildren` treats only auto and fractional rows as growable. The main `layoutGrid` path supplies its own growth flags, additionally allowing percentage rows when the container height is auto.
