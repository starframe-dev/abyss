# ANSI clipping utilities

This file documents the unexported ANSI parsing and clipping helpers in the `browser` package (`internal/browser/ansi_clip.go`). They are implementation details, not a public Go API; callers outside the package cannot access them.

## Behavior

The helpers turn terminal-oriented text into renderer-safe visible text and support measuring and clipping it by terminal cell columns. Only selected 24-bit foreground and background SGR styling is retained. Other escape sequences and terminal controls are discarded.

### `splitANSI`

```go
func splitANSI(s string) []ansiSegment
```

Splits input into text segments carrying the active style sequence(s). It:

- Removes raw C1 control encodings and their payloads before UTF-8 decoding, then replaces invalid UTF-8 with nothing.
- Recognizes ESC-prefixed CSI sequences and preserves only valid SGR sequences for 24-bit foreground (`38;2;r;g;b`) and background (`48;2;r;g;b`) colors. Each color component must be an integer from 0 through 255. Multiple color specifications can be combined.
- Clears accumulated style on SGR `0` or an empty SGR parameter list. Unsupported or malformed SGR sequences are discarded without changing the current style.
- Discards OSC, DCS, SOS, PM, APC, other escape sequences, C1 sequences, Unicode control characters, and format characters, except zero-width non-joiner (`U+200C`) and zero-width joiner (`U+200D`).
- Emits a segment when it has text; each segment receives a copy of the current style list.

### `ansiSegment`

```go
type ansiSegment struct {
    style []string
    text  string
}
```

An internal segment consisting of the accumulated renderer-compatible SGR strings and the corresponding plain text. It is not exported.

### `visibleWidth`

```go
func visibleWidth(line string) int
```

Returns the sum of `go-runewidth` cell widths of the text in parsed segments. ANSI styling and discarded controls contribute no width.

### `clipANSI`

```go
func clipANSI(line string, start, width int) string
```

Returns the portion occupying the requested cell-column interval, `[start, start+width)`. A non-positive width returns an empty string; negative starts are clamped to zero. The end is computed from the clamped start plus width.

Clipping operates on grapheme clusters (`uniseg`) so multi-rune clusters are not split. Clusters fully within the interval are preserved. If the interval intersects only part of a cluster's cell width, the overlap is represented by spaces, since a grapheme cannot be partially emitted. Zero-width clusters are included only when attached to a previously included full cluster. Style sequences are emitted for non-empty selected text, followed by `\x1b[0m` to avoid leaking style state into subsequent rendering. An entirely empty selection emits no styling.

## Implementation notes

- `stripRawC1Controls` scans the original byte string so invalid UTF-8 bytes representing C1 controls can be removed together with their control payloads before rune conversion. It handles CSI, OSC (including BEL termination), and string controls terminated by ST (`U+009C` or ESC backslash).
- `validRendererSGR` accepts only empty/reset parameters or semicolon-separated groups of exactly `38;2;r;g;b` and/or `48;2;r;g;b`; it rejects all other SGR forms.
- `clipANSI` advances its column count by grapheme cell width, not byte length, rune count, or the width of escape sequences.
