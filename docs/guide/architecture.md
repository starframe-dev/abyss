# Architecture

Abyss has a custom HTML/CSS layout engine. It is not a full browser implementation.

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

The layout engine works in terminal cells of 10 × 20 px (`internal/cellmetrics`).

Abyss does not run pages in a sandbox. JavaScript runs in one JavaScriptCore context per tab.
