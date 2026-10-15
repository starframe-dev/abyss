# Table Layout (`internal/layout/layout_table.go`)

This file implements the layout pass for table rows and cells. It is part of the internal layout package and exposes no exported Go identifiers; its functions are package-private helpers used by the layout engine.

## Behavior

`layoutTable` lays out each discovered table row in document order. Cells in a row are positioned horizontally, beginning at the table content area's left edge. Rows are stacked vertically, with the next row starting after the tallest cell in the current row.

Only cell boxes receive geometry through this routine. Table row and section boxes (such as `thead`, `tbody`, and `tfoot`) are traversal/grouping nodes and are not themselves laid out here. The function returns the total height occupied by the rows, which is zero when there are no cells to lay out.

Cell widths are determined per row:

- A non-auto width is resolved against the supplied table content width.
- Auto-width cells share the nonnegative remainder of the table width equally.
- If fixed-width cells consume more than the table width, auto cells receive zero width.
- If there are no auto-width cells, any unallocated table width remains unused.

The algorithm computes widths independently for each row; it does not align column widths across rows or account for column spans.

## Internal API

| Function | Purpose |
| --- | --- |
| `layoutTable(box *LayoutBox, contentWidth float64) float64` | Lays out rows and cells and returns the total row height. |
| `tableRows(box *LayoutBox) []*LayoutBox` | Recursively collects `tr` element boxes under the supplied box. |
| `rowCells(row *LayoutBox) []*LayoutBox` | Collects direct `td` and `th` element children in document order. |
| `tableColumnWidths(cells []*LayoutBox, tableWidth float64) []float64` | Resolves each cell width and distributes remaining width among auto cells. |

These are unexported implementation details, not a public package API.

## Implementation details

### Row and cell discovery

`tableRows` performs a depth-first traversal. When it encounters an element whose tag name is `tr`, it appends that box and stops descending into that row. This allows rows nested in table section containers to be found without treating cell descendants as additional rows. A box without a DOM node, or a non-element node, is traversed but is not considered a row.

`rowCells` examines only the row's immediate children. It includes boxes backed by element nodes named `td` or `th`, preserving child order; nested cells are not collected.

Rows without direct cells are skipped by `layoutTable` and do not increase the vertical offset.

### Width calculation and layout

For each row, `tableColumnWidths` allocates a result entry per cell. Fixed widths are resolved with `css` width resolution against `tableWidth`. It then divides `max(0, tableWidth-fixed)` evenly among auto-width cells when any exist.

Before calling `layoutBox`, `layoutTable` temporarily replaces each cell's `Style.Width` with its computed pixel width. It supplies the cell's width and its position (`box.Content.X + x`, `box.Content.Y + y`) to the layout routine, with the final boolean argument set to `false`, then restores the original style width. The horizontal offset advances by the resolved width. Row height is the maximum `Margin.Height` among its cells, and the accumulated row height determines the next row's vertical position.

The temporary width change is restored after each call, so the calculated layout geometry does not permanently rewrite the cell's declared width style.

## Dependencies and assumptions

- Uses `internal/dom` to identify element nodes and tag names (`tr`, `td`, and `th`).
- Uses `internal/css` to create pixel widths for the temporary cell layout style.
- Relies on the surrounding layout package's `LayoutBox`, `layoutBox`, and `max` definitions.
- Assumes table content width is provided by the caller and that each cell's width style can be resolved against that width.
