# `internal/browser/headless_dump.go`

## Overview

This file provides a headless entry point for rendering a fetched web page to an ANSI-art dump. It runs the browser's page-loading and rendering stages without starting the terminal UI, making it useful for deterministic end-to-end checks that do not require a PTY.

## Public API

### `HeadlessDump(pageURL string, width int, outputPath string)`

Fetches `pageURL`, processes the document and its resources, renders it at the requested terminal width, and writes the resulting ANSI text to `outputPath`.

| Argument | Meaning |
| --- | --- |
| `pageURL` | URL of the page to fetch. |
| `width` | Output width in terminal cells. The rendering pipeline converts this to pixels using `cellmetrics.WidthPx`. |
| `outputPath` | Destination file for the rendered ANSI dump. |

The function has no return value. Fetch and output-file errors are printed to standard error and terminate the process with exit status 1 via `os.Exit`; callers cannot handle these errors as returned values. On success, a status message containing the dump byte count and destination path is printed to standard error.

## Processing sequence

1. Reset the image cache and set `render.ImageProvider` to the browser's cached image provider.
2. Fetch the page HTML with `fetcher.Fetch`.
3. Parse the HTML into a DOM and fetch images referenced by the document.
4. Fetch external CSS and compute styles at `width * cellmetrics.WidthPx` pixels.
5. Build the layout using the computed styles and the same pixel width.
6. Render the layout to a framebuffer at that width.
7. Convert the framebuffer and layout to ANSI lines with `ascii.RenderANSI`.
8. Join the lines with newline characters and write the result with permissions `0644` (subject to the process umask).

The function does not start or interact with the TUI. Rendering setup uses package-level `render.ImageProvider`, so this operation does mutate shared package state; concurrent calls or other rendering work in the same process should account for that shared state.

## Output and errors

The output file contains the joined ANSI lines and does not receive an extra newline beyond those present between lines. File creation and replacement follow `os.WriteFile` behavior. Fetch failures are reported with the `headless dump: fetch:` prefix; write failures use `headless dump: write:`. Both failure paths exit immediately with status 1. Successful completion reports `headless dump: <bytes> bytes -> <path>` on standard error.
