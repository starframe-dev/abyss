# `cmd/external_test`

## Purpose

This executable is a manual end-to-end harness for loading and rendering a page that uses external JavaScript. It fetches a fixed local HTML URL, parses the document, initializes the JavaScript/DOM bridge, executes scripts, computes styles and layout, renders the result to an ASCII representation, and writes the lines to standard output.

## Public API

The package is `main` and exposes no library API. It is intended to be run as an executable. Its package-level constants and variable are unexported implementation details:

| Name | Kind | Value / role |
| --- | --- | --- |
| `viewportWidth` | constant | Rendering viewport width of 800 units. |
| `defaultClientTimeout` | constant | Duration of 10 seconds; declared but not used in this file. |
| `defaultClient` | variable | Alias of `fetcher.DefaultClient`; declared but not used in this file. |

## Runtime behavior

`main` performs these steps in order:

1. Fetches `http://localhost:9999/test_external_js.html` with `fetcher.Fetch`.
2. On fetch failure, prints `Fetch error: <error>` to standard error and exits with status 1.
3. Parses the response bytes into a DOM using `dom.Parse`.
4. Creates a JavaScript context and defers `Close` to release it on normal return.
5. Registers bridge functions. A failure is reported as `Bridge init error: <error>` and terminates the process with status 1.
6. Initializes the JavaScript DOM binding with `js.InitDOM` and evaluates the DOM API script. Evaluation errors are reported as `DOM API error: <error>` and terminate the process.
7. Executes document scripts using the original URL as the script base URL. Errors are reported as `Script error: <error>` and terminate the process.
8. Obtains the final DOM after script execution, computes CSS, builds layout for an 800-unit viewport, renders a framebuffer, and converts it to ASCII lines.
9. Prints each rendered line to standard output, followed by the sentinel line `=== END ===`.

## Implementation details

- The target URL and viewport width are fixed in source; the command does not accept command-line arguments or environment configuration.
- The pipeline is synchronous. Each stage consumes the result of the preceding stage, and processing stops at the first reported error.
- Fetching uses `fetcher.Fetch`; the local `defaultClient` alias and timeout constant do not configure this call in the current implementation.
- The executable depends on the internal `ascii`, `css`, `dom`, `fetcher`, `js`, `layout`, and `render` packages.
- Successful output consists of the ASCII rendering followed by `=== END ===`. Diagnostic errors go to standard error, while the rendered output goes to standard output.
