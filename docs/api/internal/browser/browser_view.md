# Browser View Rendering

This file documents `internal/browser/browser_view.go`, which renders the browser model as a terminal interface. It uses Lipgloss for styling, `go-runewidth` for terminal-cell width calculations, and `uniseg` for grapheme-aware truncation.

## Public API

### `Model.View() string`

`View` implements the Bubble Tea `tea.Model` view method on `Model`. It returns the complete terminal representation as a string; it does not write directly to the terminal or mutate the model.

The output is composed in this order:

1. **Tab bar.** Each tab URL is sanitized for terminal display. An empty sanitized URL is shown as `New Tab`; otherwise the HTTP or HTTPS prefix is removed. Labels wider than 20 terminal cells are truncated to 19 cells and suffixed with `…`. The active tab uses a distinct background, foreground, and bold style.
2. **Header.** In `ModeAddress`, the address input is rendered after the `Abyss` label. In other modes, the current address is displayed, or `Abyss` if empty. When zoom differs from 100%, the header appends the zoom percentage rounded to a whole number.
3. **Content.** Content is selected by mode:
   - `ModeLoading`: centered `Loading...` message.
   - `ModeError`: a centered sanitized error message and a centered keyboard hint (`Ctrl+L` to edit the URL or `Q` to quit).
   - `ModeAddress`: the sanitized address input value, centered.
   - `ModeView`: the visible ANSI-rendered document rows, vertically selected using the viewport offset and height, then horizontally clipped according to the horizontal scroll offset.

The tab bar and header each end with a newline. Mode-specific content adds its own leading newlines as implemented; view mode emits the joined visible rows directly.

## Rendering details

### Terminal text safety

`safeTerminalText` replaces Unicode runes that are not printable with `�`, except zero-width non-joiner (`U+200C`) and zero-width joiner (`U+200D`), which are preserved. This is applied to tab labels, header text, address-mode content, and error text. The address input value is sanitized on a local copy of the input model before rendering.

### Width, truncation, and centering

Widths are terminal display-cell widths rather than byte counts. Tab truncation iterates grapheme clusters, so combining sequences are not split. `center` prepends half of the available horizontal space when the text is narrower than the requested width; it does not pad the right side. If the text is at least as wide as the requested width, it is returned unchanged.

### Viewport and horizontal clipping

In `ModeView`, a non-positive viewport height defaults to 24 rows, and a non-positive model width defaults to 80 columns. The vertical offset is clamped so the selected slice does not start beyond the end of the available ANSI lines; the slice is then bounded by the line count. Horizontal clipping is applied per row using the model's horizontal offset and `clipANSI`, preserving ANSI-aware clipping behavior. Rows are copied into a separate output slice before clipping because the visible rows may alias stored tab content; this prevents rendering from destructively changing the source lines and ensures horizontal scrolling can be reversed.

## Internal helpers

| Helper | Behavior |
|---|---|
| `safeTerminalText(text string) string` | Replaces non-printable runes as described above. |
| `truncateTerminalText(text string, maxWidth int) string` | Returns a grapheme-safe prefix whose accumulated display width does not exceed the maximum. Zero-width graphemes are retained; negative widths, if any, do not reduce the accumulated width. |
| `center(text string, width int) string` | Adds left padding to approximately center text based on terminal-cell width. |

These helpers are unexported and are implementation details of the browser package.
