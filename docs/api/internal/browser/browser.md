# `internal/browser`

Package `browser` implements the terminal browser UI with Bubble Tea. It models address entry, loading, page viewing, errors, tabs, viewport state, and asynchronous page-load results.

## Modes

`Mode` is an integer enum describing a tab or the overall UI state:

| Constant | Meaning |
| --- | --- |
| `ModeAddress` | Address-entry state; also the initial state for a new tab. |
| `ModeLoading` | A page request is in progress. |
| `ModeView` | Loaded page is being viewed. |
| `ModeError` | A page load or display error is active. |

## Public types

### `Tab`

Stores the state associated with one browser tab.

| Field | Purpose |
| --- | --- |
| `ID`, `RequestID` (`uint64`) | Tab identity and current request generation, used to associate asynchronous results with the request that produced them. |
| `URL` (`string`) | Address associated with the tab. |
| `Mode` (`Mode`) | Tab's UI state. |
| `ANSILines`, `ASCIILines` (`[]string`) | Rendered page lines in ANSI and plain ASCII forms. |
| `ExternalCSS` (`[]string`) | External stylesheet references associated with the page. |
| `Err` (`error`) | Error associated with the tab. |
| `ZoomFactor` (`float64`), `OffsetX`, `OffsetY` (`int`) | Per-tab zoom and viewport position. |

The load-cancellation function is private and is used internally to cancel outstanding work.

### `Model`

The main Bubble Tea model. It exposes the current `Mode`, address input (`Address`), page `Viewport`, `Tabs`, `ActiveTab` index, and terminal `Width` and `Height`. Other fields hold internal tab-ID sequencing and rendered page content.

`Model` provides `Init() tea.Cmd` as the Bubble Tea initialization hook. It batches the text-input blink command and a periodic tick command.

### Asynchronous messages

- `PageLoadedMsg` carries `TabID`, `RequestID`, rendered `Lines` and `ANSI` lines, final `URL`, and `ExternalCSS` when a load completes.
- `PageRenderedMsg` carries updated `Lines` and `ANSI` for a tab/request without replacing navigation metadata.
- `PageErrorMsg` carries the tab and request identifiers plus the load error.

Results are matched against both tab ID and request ID; stale results do not resolve to a current tab. A zero tab ID is rejected by the internal lookup.

## Public functions and methods

### `New() Model`

Creates a model with an 80-by-24 viewport, an address input focused with placeholder `Enter URL...` and width 50, and one address-mode tab. If a saved session contains URLs, those URLs instead initialize address-mode tabs and the last restored tab becomes active. Tab IDs are assigned to tabs that do not already have one.

### `TickCmd() tea.Cmd`

Returns a Bubble Tea command that emits an internal tick message every 500 milliseconds.

### `(*Model).GetActiveTab() *Tab`

Returns a pointer to the active tab, or `nil` when `ActiveTab` is outside the tab slice.

### `(*Model).SaveActiveTab()`

Copies the current address, mode, ANSI and ASCII lines, zoom, horizontal offset, and viewport vertical offset into the active tab. Does nothing if there is no valid active tab.

### `(*Model).RestoreActiveTab()`

Restores the active tab's mode, address, and ASCII viewport content. It clamps the vertical offset to the available content range. Does nothing if there is no valid active tab.

## Session persistence and implementation notes

Session storage is private and located at `$HOME/.abyss/session.json`, with a JSON object containing a `urls` array. Saving omits empty URLs and does nothing when there are no tabs or no non-empty URLs. Directory creation and file writing use permissions `0755` and `0644`; persistence and JSON errors are deliberately ignored. Loading returns no URLs if the file cannot be read or decoded.

Tab IDs are monotonically assigned within the model when missing, while existing IDs advance the next-ID counter as needed. Each tab may hold a cancellation function for in-flight loading; internal cancellation invokes and clears these functions. Active-tab lookup is bounds-checked. Viewport restoration joins ASCII lines with newlines and limits the saved vertical offset to the current content bounds.
