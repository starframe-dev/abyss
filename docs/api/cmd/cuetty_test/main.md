# `cmd/cuetty_test`

`cmd/cuetty_test` is a standalone manual integration-test program for exercising a terminal application through the `github.com/starframe-dev/cue-tty/pkg/cue` API. It starts `./bin/abyss`, drives its page as if a user were interacting with it, and prints both the rendered terminal lines and a readable representation of the captured raw output.

## Public API

This directory contains the `main` package and exposes no reusable exported API. Its entry point is `main()`.

## Execution flow

1. Launch `./bin/abyss` with a 200-column by 60-row terminal, the current directory as its working directory, and `TERM=xterm-256color`.
2. If launch fails, print `Launch error: <error>` to standard output and return. On success, defer `app.Close()` so the launched application is closed when `main` returns.
3. Obtain the application's page and wait one second for startup.
4. Type `example.com`, wait 200 milliseconds, and press Enter.
5. Wait five seconds for the page to load, then retrieve and print its lines.
6. Retrieve the application's raw log and print at most its final 2,000 bytes in an escaped, human-readable form.
7. Press `q` and wait 200 milliseconds before returning; the deferred close then runs.

## Output details

### Page lines

The program prints a heading containing the number of returned lines and the configured terminal dimensions. Each line is indexed with a two-character, space-padded index. Trailing ASCII spaces are removed before display. A line that becomes empty is printed as `(empty)`; other lines are printed with Go's `%q` formatting, which quotes and escapes the string.

The error returned by `page.Lines()` is intentionally ignored in this test program. Consequently, a retrieval error is not reported explicitly and the returned lines are still processed.

### Raw log

The raw log section is limited to the final 2,000 bytes (or the whole log when it is shorter). Bytes are displayed as follows:

- ESC is written as `\\033`.
- Newline is written as `\\n` followed by an actual line break.
- Carriage return is omitted.
- Printable ASCII bytes (32 through 126) are emitted directly.
- All other bytes are written as lowercase, two-digit hexadecimal escapes such as `\\x1b`.

The log is sliced by byte index, not by Unicode character, so the displayed tail may begin in the middle of a multibyte encoding. The implementation assumes `app.RawLog()` returns a byte slice suitable for `len` and indexing.

## Runtime assumptions and limitations

The program expects `./bin/abyss` to exist relative to its working directory and to be launchable by the cue-tty package. Its interaction sequence relies on fixed sleeps rather than readiness checks, so startup or page-load timing differences can affect the result. It is an executable diagnostic example rather than an automated assertion-based test: it prints observations but does not compare them with expected output or return a failure status for page-content mismatches.
