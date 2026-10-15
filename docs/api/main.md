# Command-line entry point (`main.go`)

This file defines the `main` package entry point for Abyss. It provides two execution modes: an interactive terminal UI (the default) and a headless ANSI-render dump mode.

## Command-line flags

| Flag | Default | Description |
| --- | --- | --- |
| `--headless-dump` | Empty | Output file path. When non-empty, enables headless mode and writes the ANSI rendering to that path. |
| `--url` | Empty | URL to load in headless mode. Required when `--headless-dump` is supplied. |
| `--width` | `200` | Viewport width used by headless mode. |

Flags are parsed using Go's standard `flag` package. There are no exported declarations in this file; the command-line behavior is exposed through the executable rather than a Go library API.

## Execution behavior

### Headless dump mode

When `--headless-dump` is non-empty, the program checks that `--url` is also non-empty. If it is missing, the program writes `--headless-dump requires --url` to standard error and exits with status `2`. Otherwise it calls:

```go
browser.HeadlessDump(url, width, outputPath)
```

The entry point then returns without creating or running the interactive UI. Loading, rendering, output-file handling, and any error behavior of the dump operation are implemented by `internal/browser`.

### Interactive mode

When `--headless-dump` is empty, the program creates a browser model with `browser.New()` and runs it using Bubble Tea with the alternate screen enabled. If `tea.Program.Run` returns an error, the program writes `Error: <error>` to standard error and exits with status `1`. A successful run returns normally.

## Implementation notes

- Headless mode is selected solely by whether `--headless-dump` is a non-empty string; `--url` and `--width` do not independently enable it.
- The width flag is passed through directly to `browser.HeadlessDump`; this file does not validate or normalize it.
- The command has no returned Go values. Failures handled here use process exit codes: `2` for the missing headless URL and `1` for an interactive Bubble Tea run error.
