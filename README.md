# Abyss

TUI browser that renders web pages as **ASCII color art**.

Unlike text-based browsers (Lynx, w3m), Abyss has a real **CSS layout engine**:
it computes flexbox/grid/absolute positioning, draws backgrounds, borders, and border-radius
into an 800px pixel framebuffer, then converts pixels into colored block characters (█▄▀)
with ANSI truecolor escape codes.

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

## Controls

| Key | Action |
|-----|--------|
| `Ctrl+L` | Focus address bar |
| `Enter` | Load URL |
| `↑/↓` `PgUp/PgDn` | Scroll page |
| `q` / `Ctrl+C` | Quit |

## Supported CSS

- **Display**: `block`, `inline`, `flex`, `grid`, `none`
- **Position**: `static`, `relative`, `absolute`
- **Box Model**: `width`, `height`, `margin`, `padding`, `border`
- **Visual**: `background-color`, `color`, `border-radius`, `z-index`, `overflow`
- **Flexbox**: `flex-direction`, `justify-content`, `align-items`, `flex-grow`
- **Grid**: `grid-template-columns`, `grid-template-rows`, `gap`

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
└── specs/SPEC.md
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

Tests verify:
- Initial address bar render
- Loading `example.com` and showing "Example Domain"
- Scrolling
- Error page for invalid URL
- Clean exit via `q`
