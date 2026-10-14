# Abyss

TUI browser that renders a supported subset of web pages as **ANSI color art**.

Abyss has a custom HTML/CSS layout engine (not a full browser implementation): it computes
basic block, Flexbox, Grid, and absolute positioning, draws backgrounds and borders into a
bounded pixel framebuffer, then converts pixels into colored block characters (█▄▀) with
ANSI truecolor escape codes.

## Architecture

```
HTML + CSS → DOM Parser → CSS Engine (cascade, specificity)
    ↓
Layout Engine (box model, flexbox, grid, absolute positioning)
    ↓
Pixel Framebuffer (800px viewport, backgrounds, borders, border-radius)
    ↓
ANSI Color Art (2×1 pixel blocks → █▄▀ with truecolor codes)
    ↓
Bubble Tea TUI (viewport, scrolling)
```

## Usage

```bash
go run .
```

Or build and run:

```bash
go build -o abyss .
./abyss
```

For a non-interactive ANSI dump:

```bash
go run . --headless-dump /tmp/page.ansi --url https://example.com --width 120
```

## Controls

| Key | Action |
|-----|--------|
| `Ctrl+L` | Focus address bar |
| `Enter` | Load URL |
| `↑/↓` `PgUp/PgDn` | Scroll page |
| `q` / `Ctrl+C` | Quit |

## Supported CSS (subset)

- **Display**: `block`, `inline`, `flex`, `grid`, `none`
- **Position**: `static`, `relative`, `absolute`
- **Box Model**: `width`, `height`, `margin`, `padding`, `border`
- **Visual**: `background-color`, inherited `color`, `border-radius`, `z-index`, `overflow`
- **Flexbox**: basic row/column direction, grow/shrink, and justify-content distribution
- **Grid**: bounded `px`, percentage, `fr`, and repeat tracks; basic gaps

This is intentionally a subset; Grid `fr` row distribution needs a definite container
height, while auto-height rows use content sizes. Intrinsic sizing, browser-default styles
and complex CSS are not fully implemented.

## Project Structure

```
├── internal/
│   ├── dom/         # HTML → DOM tree
│   ├── css/         # CSS parser + computed styles (cascade, specificity)
│   ├── layout/      # CSS layout engine (box model, flex, grid, absolute)
│   ├── framebuffer/ # RGB pixel buffer
│   ├── render/      # Layout tree → pixel framebuffer
│   ├── ascii/       # Framebuffer → ANSI color art
│   ├── fetcher/     # HTTP page loading
│   └── browser/     # Bubble Tea TUI
├── main.go
└── SPEC.md
```

## Testing

```bash
go test ./... -v
```

### E2E TUI Tests (CueTTY)

Real PTY-based tests using [cue-tty](https://github.com/starframe-dev/cue-tty):

```bash
go test ./internal/browser -run TestAbyss_E2E -v
```

Tests use deterministic local HTTP fixtures rather than relying on live websites. They verify:
- Initial address bar render
- Page loading and block stacking
- Large viewport / truecolor output
- Error page and clean exit via `q`

## Resource limits and known limitations

Network responses, DOM size/depth, CSS tracks, image dimensions/cache, and page resource
counts are bounded. JavaScript executes through JavaScriptCore; currently its context is
shared across tabs, and synchronous scripts cannot be safely preempted. Do not use Abyss
for untrusted sites if a non-terminating script could block the process.
