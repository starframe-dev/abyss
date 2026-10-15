# `cmd/debug_render`

`cmd/debug_render` is a standalone diagnostic executable that fetches a fixed web page, runs it through Abyss's parsing, JavaScript, styling, layout, and rendering pipeline, then writes text-rendered output to temporary files and prints a summary to standard output.

## Behavior

The program processes `https://example.com` using this sequence:

1. Fetches the page with `fetcher.Fetch`. A fetch failure is reported to standard error and terminates the process with exit status 1.
2. Parses the response with `dom.Parse`.
3. Creates a JavaScript context, arranges to close it on return, initializes the DOM bridge, registers bridge functions, and evaluates the DOM API script. Bridge registration or DOM API evaluation errors are reported to standard error and terminate with exit status 1.
4. Executes page scripts. A script execution error is printed to standard output, but does not stop the remaining pipeline.
5. Reads the final DOM, computes CSS styles, builds layout at width 800, renders a framebuffer at width 800, and converts it to ASCII lines.
6. Writes the full lines, joined with newline characters, to `/tmp/abyss_full_ansi.txt`. It also creates a simplified version in `/tmp/abyss_short_ansi.txt`: ANSI escape sequences are stripped, each line is truncated to 120 bytes, and trailing spaces are removed.
7. Prints the output paths, counts non-empty shortened lines, and prints each non-empty line with its zero-based line index.

The two output files are written with mode `0644`. Write errors from `os.WriteFile` are ignored, so the success message does not guarantee that either file was written successfully.

## Public API

This is a `main` package and exposes no package-level public API intended for import. Running the executable invokes `main`; it accepts no command-line arguments and uses a hard-coded URL, render width, and output paths.

### `stripANSI`

```go
func stripANSI(s string) string
```

Removes recognized ANSI SGR sequences of the form ESC `[` followed by digits and/or semicolons and then `m`. Other escape sequences are skipped according to the function's byte-by-byte scan; this is a small renderer-output helper, not a general terminal escape parser. The function is unexported.

## Implementation notes

- The JavaScript context is closed with `defer` after creation. Calls to `os.Exit` bypass deferred calls; the early fatal paths therefore terminate the process without running deferred cleanup.
- Short-line truncation uses Go byte lengths and slicing (`len` and `clean[:120]`), not Unicode rune counts. Non-ASCII text may consequently be cut in the middle of a UTF-8 encoded character.
- Full output retains ANSI styling; shortened output strips the recognized styling sequences before truncation.
- Script execution errors are non-fatal, unlike fetch, bridge registration, and DOM API evaluation errors.
