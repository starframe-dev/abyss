# `cmd/debug_ansi`

## Purpose

This standalone Go command is a manual integration/debugging harness for the Abyss terminal application. It launches `./bin/abyss` in a simulated terminal, inspects the screen and captured output, checks that expected prompts and page content appear, and then sends the quit key.

The command reports check results to standard output. A launch error is reported to standard error and causes exit status 1. Content-check failures are printed as `FAIL` messages but do not themselves change the process exit status.

## Execution flow

1. Launch `./bin/abyss` with a 120-by-40 terminal, the current directory as its working directory, and `TERM=xterm-256color`.
2. Defer closing the launched application.
3. Obtain its page and wait up to the requested stability interval of two seconds.
4. Read page text and print its character count and at most the first 500 characters. The error returned by `page.Text()` is ignored.
5. Read the application's raw log, strip recognized ANSI escape sequences, and print the cleaned byte count and at most the first 500 bytes.
6. Check the cleaned log for `Abyss` and `Enter URL`, reporting each result.
7. Type `example.com`, wait 300 milliseconds, and press Enter.
8. For up to 15 seconds, poll the raw log every 200 milliseconds for `Example Domain`, stripping ANSI sequences on each poll. Report whether the text was found.
9. Press `q` and wait 500 milliseconds for the page to stabilize; deferred cleanup closes the application when `main` returns.

## Internal helpers and state

| Name | Kind | Description |
| --- | --- | --- |
| `ansiRegexp` | package variable (`*regexp.Regexp`) | Compiled expression matching CSI-style ANSI sequences of the form ESC, `[`, zero or more digits/semicolons, then one ASCII letter. |
| `stripANSI(s string) string` | unexported function | Removes all substrings matching `ansiRegexp`. It does not claim to handle every ANSI/VT escape-sequence form. |
| `min(a, b int) int` | unexported function | Returns the smaller integer; used to cap preview slices at 500 without exceeding the available string length. |
| `main()` | entry point | Runs the launch, inspection, input, polling, and shutdown sequence described above. |

This `main` package exposes no public Go API. Its imports use the standard `bytes`, `fmt`, `os`, `regexp`, and `time` packages and `github.com/starframe-dev/cue-tty/pkg/cue` for terminal application control.

## Operational details

- The executable path is relative (`./bin/abyss`) and the working directory is explicitly set to `.`; execution therefore depends on the caller's current directory and on the binary being present there.
- `cue.Launch` failure is handled explicitly: the command prints `FAIL: launch: ...` and exits with status 1. Once launch succeeds, `app.Close()` is deferred.
- Output preview lengths use Go string byte lengths/slicing for the log and string indexing for page text. The labels use `chars` for page text and `bytes` for the cleaned log, but Go's `len` counts bytes in both cases.
- The initial presence checks are diagnostic only. The later `Example Domain` check is bounded by a deadline and uses periodic polling rather than waiting indefinitely.
- The command is a debug harness, not a general-purpose ANSI parser or a formal test: it does not return structured results or fail the process on missing expected content.
