# `cmd/manual`

`cmd/manual` is a standalone Go executable that runs a built-in integration demonstration of Abyss's HTML, JavaScript, CSS, layout, rendering, and ASCII-output pipeline. It is a manual/demo command rather than a reusable library package; the package exposes no exported API.

## Execution flow

When run, `main`:

1. Writes the embedded `testPage` HTML to `abyss_test_page.html` in the operating system's temporary directory, reports its path, and schedules its removal with `defer`. A write failure panics.
2. Parses the page with `dom.Parse` and prints the number of children under the parsed document root.
3. Creates a JavaScript context and schedules `Close` on return. It initializes the JS DOM bridge against the parsed document and directs JS logging to standard error.
4. Registers bridge functions and evaluates the DOM API bootstrap script. Either failure is reported to standard output and terminates the process with exit status 1.
5. Executes scripts in the document. A script-execution error is reported but does not immediately terminate the command.
6. Retrieves the resulting DOM. When non-nil, it prints a summary of the final tree.
7. Computes CSS, builds layout at width 800, renders an 800-wide framebuffer, converts it to ANSI lines, and prints dimensions and the output lines.
8. Sleeps for two seconds so the output remains observable.

## Embedded test page

The `testPage` constant is a self-contained HTML fixture containing inline CSS and JavaScript. Its styles exercise flex and grid layouts, spacing, borders, colors, and rounded corners. The script exercises the bridge's logging function and DOM operations including element creation, text content, class names and lists, inline styles, attributes, parent/child access, `innerHTML`, `document.write`, ID and tag-name queries, and child removal. It also logs selected results and catches errors around several API checks. The body supplies colored blocks, grid cells, and cards for visual rendering.

## Functions

### `main()`

The executable entry point and orchestration described above. It uses standard output for progress, errors, DOM summaries, and rendered lines; JavaScript logging is routed to standard error. Rendering uses a fixed viewport width of 800. The command has no flags or arguments.

### `printDOMTree(root *dom.Node)`

Prints each element child of `root` with its tag and child count. For an element named `html`, it additionally delegates to `printHTMLBody`. Non-element root children are ignored. This is a diagnostic helper, not an exported package API.

### `printHTMLBody(htmlEl *dom.Node)`

Finds direct `body` element children of the supplied HTML node and prints each body's direct children. Element output includes tag name, `class` attribute, and `TextContent()`. Text nodes are trimmed and truncated to 40 characters with an ellipsis when longer; other node types are omitted. This helper assumes the DOM node and its attribute map are usable according to the `internal/dom` package's representation.

## Dependencies and boundaries

The command composes internal packages directly: `internal/dom` parses and represents the document; `internal/js` initializes and executes the embedded script; `internal/css` computes styles; `internal/layout` builds boxes; `internal/render` creates the framebuffer; and `internal/ascii` converts the result into terminal-oriented lines. These are internal project dependencies, not public APIs of `cmd/manual`.

The temporary HTML file is written with permission mode `0644` and removed on normal function return. JavaScript context cleanup is also deferred. Bridge/bootstrap failures use `os.Exit(1)`, which ends the process immediately and therefore does not run deferred cleanup registered in `main`; script execution errors, by contrast, allow the later rendering stages to proceed.
