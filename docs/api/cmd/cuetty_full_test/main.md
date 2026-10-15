# `cmd/cuetty_full_test`

## Overview

This Go `main` package is a manual end-to-end smoke-test executable for the Abyss terminal application. It launches `./bin/abyss` in a simulated terminal using the `github.com/starframe-dev/cue-tty/pkg/cue` test-control API, drives keyboard input, inspects rendered terminal text, and exits non-zero when a required behavior fails.

Run it from the repository context where `./bin/abyss` exists and the current directory is the intended working directory. The terminal is configured as 120 columns by 40 rows with `TERM=xterm-256color`.

## Public API

There is no exported library API. The package exposes the Go program entry point only:

| Symbol | Description |
| --- | --- |
| `main()` | Launches the application, runs the sequential smoke-test scenarios, reports results to standard output, and terminates the process on failure. |

The executable depends on the `cue` package's `Launch`, launch options, page interaction methods, and expectation helpers. These are used here; this file does not define their implementations or additional public types.

## Test sequence

1. **Initial screen:** Launches Abyss and waits up to two seconds for a stable screen. Prints the rendered text and checks for `Abyss` and `Enter URL`.
2. **Navigation:** Types `example.com`, waits briefly, presses Enter, and waits up to 15 seconds for `Example Domain`. It then checks and prints the loaded page text.
3. **Scrolling:** Captures the screen, presses Down eight times, waits for stability, and fails if the rendered text is unchanged.
4. **Refresh:** Presses `r`, waits up to two seconds, and checks that `Example Domain` remains visible.
5. **Error page:** Presses Escape, types `not-a-valid-url-12345.xyz`, submits it, waits up to ten seconds for `Error:`, and verifies that marker.
6. **Quit:** Presses `q`, waits briefly, explicitly closes the application, and checks that closing succeeds.

Successful completion prints `ALL TESTS PASSED` after the individual scenario messages.

## Failure and resource handling

A launch failure, failed expectation, or failed quit/close check writes a `FAIL:` message to standard error and calls `os.Exit(1)`. A successful launch registers `defer app.Close()` as cleanup; the explicit close in the quit scenario is also checked and is followed by the deferred close when `main` returns. Because `os.Exit` terminates immediately, deferred functions are not run on failure paths that call it.

Calls that retrieve terminal text discard the returned error (`page.Text()` uses `text, _`). Consequently, this test treats the returned text as usable and does not report text-retrieval errors separately. Timing and waits are fixed rather than configurable, so this is a smoke test whose behavior depends on the executable being available and its network navigation completing within the specified timeout.
