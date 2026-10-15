# Getting Started

Abyss is a TUI browser. It renders a supported subset of HTML and CSS into a terminal as ANSI color art.

## Run

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
