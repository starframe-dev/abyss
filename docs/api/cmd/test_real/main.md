# `cmd/test_real`

`cmd/test_real` is a standalone Go executable that exercises the project's real-page processing pipeline against two hard-coded URLs. For each page, it fetches HTML, parses and runs the DOM-related JavaScript processing, computes styles and layout, renders a framebuffer, converts it to ASCII lines, and prints timing and output summaries to standard output.

## Command behavior

The executable processes these URLs sequentially:

- `https://en.wikipedia.org/wiki/ASCII_art`
- `https://news.ycombinator.com`

For each URL, it performs the following steps:

1. Starts a timer and calls `fetcher.Fetch`.
2. On a fetch error, prints `[URL] FETCH ERROR: error` and continues to the next URL. No further work is done for that page.
3. Records fetch duration and parses the HTML with `dom.Parse`.
4. Creates a JavaScript context, initializes its DOM integration, registers bridge functions, evaluates the DOM API script, and executes scripts found in the document with the URL as context.
5. Closes the JavaScript context, then obtains the final DOM using `js.GetFinalDOM`.
6. Computes CSS styles, builds a layout at width 800, renders it at width 800, and converts the rendered result into ASCII lines.
7. Prints the URL, HTML byte count, fetch/parse/total elapsed times, and total and non-empty line counts. It then prints up to the first 20 non-empty lines, with their zero-based line indexes.

The parse duration is measured from the original start time, so it includes fetching as well as parsing. Total duration is likewise measured from that start and includes the pipeline through ASCII conversion and the non-empty-line count calculation. The displayed sample is printed after the total duration is recorded.

## Public API

This package declares no exported library API. It is a `main` package intended to be run as a command; its entry point is the unexported `main` function.

## Output details

- HTML size is reported using `len(html)`, in bytes.
- A line is considered non-empty when `strings.TrimSpace(line)` is not empty.
- Displayed sample lines have trailing spaces removed with `strings.TrimRight(line, " ")`; their leading whitespace is retained.
- At most 20 non-empty lines are displayed for each successfully processed page.
- Fetch failures are reported and do not prevent the second URL from being attempted.

## Implementation notes

The pipeline is wired directly in `main`; this file does not expose configuration or command-line flags. Both the requested page list and rendering width are fixed in source. The JavaScript context is closed before retrieving the final DOM, matching the current integration's lifecycle. Errors are handled explicitly for fetching; subsequent processing functions in this file are called without local error handling.
