# Browser Update

`internal/browser/browser_update.go` implements message handling for the terminal browser's Bubble Tea model. It updates model and tab state in response to input, window changes, timer events, and asynchronous page-load/render results.

## Public API

| Symbol | Description |
| --- | --- |
| `(*Model).Update(msg tea.Msg) (tea.Model, tea.Cmd)` | Bubble Tea `Update` implementation. Applies one message and returns the updated model plus a command, if any. |

The method has a pointer receiver but returns the model as the `tea.Model` interface. Its behavior relies on `Model`, tab helpers, message types, and commands defined elsewhere in the `browser` package.

## Message handling

### Quit and window sizing

- A `tea.QuitMsg` cancels outstanding loads, closes the JavaScript context, and returns `tea.Quit`.
- A `tea.WindowSizeMsg` updates model dimensions, sets address-input width to `width - 4`, and viewport dimensions to `width` by `height - 4`.
- While viewing a page, a resize refreshes viewport content from the active tab's ANSI lines. If the width changed, it also schedules the active render command when one is available.

### Keyboard input

`ctrl+c` saves the session, cancels loads, closes the JavaScript context, and quits in any mode. When multiple tabs exist, `tab` and `shift+tab` save the current tab, move cyclically through tabs, restore the newly active tab, and return without a command.

#### Address mode

- `enter` does nothing for an empty address. Otherwise, it adds `https://` unless the value already begins with `http://` or `https://`, saves active-tab state, and starts a new cancellable load.
- Starting a load cancels any previous load for that tab, creates a context with cancellation, increments the request ID, sets loading state and URL, resets zoom to `1` and horizontal offset to `0`, and updates the address field. The returned command loads the page using the current width, tab ID, and request ID.
- `q` saves the session, cancels loads, closes the JavaScript context, and quits.
- Other keys are delegated to the address text input.

#### View mode

| Key | Behavior |
| --- | --- |
| `ctrl+l` | Save active-tab state, switch to address mode, focus the address input, and start its blink command. |
| `r` | Return the active page render command. |
| `+` or `=` | Increase zoom by `0.5`, capped at `3`, then request a render. |
| `-` or `_` | Decrease zoom by `0.5`, floored at `0.5`, then request a render. |
| `left` | Move horizontal offset left by `10`, not below zero. |
| `right` | Move horizontal offset right by `10`. |
| `t` | Save current tab state, create and activate a blank address-mode tab, clear and focus the address field and viewport, and start input blinking. |
| `w` | If more than one tab exists, cancel and close the active tab's JavaScript context, remove that tab, select a valid remaining index, and restore it. The last tab cannot be closed through this key. |
| `tab` / `shift+tab` | Cycle through tabs and restore state. |
| `q` | Save session, cancel loads, close the JavaScript context, and quit. |
| Other keys | Delegate to the viewport and return its command. |

Zoom and horizontal navigation update the active tab when one exists. A render command may be nil if the active tab cannot render.

#### Error mode

`ctrl+l` returns to address mode, focuses the address field, and starts blinking. `q` saves the session, cancels loads, closes the JavaScript context, and quits. Other keys do not trigger a mode-specific action.

### Timer events

On `tickMsg`, loading mode schedules another tick and skips timer processing. Otherwise, if the active tab has JavaScript timers to process, the model schedules a render; it then schedules the next tick. Commands accumulated in this handler are combined with `tea.Batch`.

### Asynchronous page results

Results are matched to a tab using both `TabID` and `RequestID`; unmatched or stale results are ignored.

- `PageRenderedMsg` replaces the matched tab's ANSI and ASCII lines. If that tab is active, its lines are also joined with newline separators and displayed in the viewport.
- `PageLoadedMsg` cancels and clears the tab's load cancellation function, records the resulting URL, ANSI/ASCII lines, and external CSS, resets vertical offset, and marks the tab as viewing. For the active tab it also updates the model mode and loaded-page content, sets viewport content, returns the viewport to the top, and schedules a tick.
- `PageErrorMsg` cancels and clears the matched load, stores the error, and marks the tab as error state. The model's mode changes only when the errored tab is active.

## State and implementation details

- `ensureTabIDs()` runs at the beginning of every update, before message dispatch.
- Per-tab state is saved and restored around tab switches so the displayed model follows the active tab.
- Loads use `context.WithCancel`; replacing or closing a tab cancels its current load. Quit paths cancel outstanding loads and close JavaScript context resources.
- Request IDs prevent late results from an earlier navigation from overwriting newer tab state.
- Most message cases return immediately when they have a dedicated command or no further work. Remaining accumulated commands are returned as `tea.Batch(cmds...)`.
- Address normalization is a simple prefix check; only lowercase `http://` and `https://` are recognized as already-schemed values.
