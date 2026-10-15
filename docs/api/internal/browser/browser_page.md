# Browser page pipeline

`internal/browser/browser_page.go` implements page loading and rendering for the browser package. It fetches a document, parses and executes its scripts, fetches external stylesheets and images, computes styles and layout, and renders terminal-friendly output. Bubble Tea commands run the work asynchronously and return typed messages consumed by the application.

## Page loading

The internal entry points `loadPage`, `loadPageContext`, and `loadPageForTab` create a `tea.Cmd`. `loadTabPage` adds the tab and request identifiers to the resulting success or error message. The context-aware path checks cancellation after fetching and parsing and at later pipeline boundaries.

The load sequence is:

1. Fetch HTML through `fetcher.FetchContext` and parse it into a DOM.
2. Under the JavaScript lifecycle mutex, suspend the previously active runtime, discard any old runtime for this tab, create a context, initialize DOM bridges and the event system, and evaluate the DOM API script.
3. Set `document._url`, provide the JavaScript `window.getComputedStyle` bridge, execute document scripts, and dispatch a `DOMContentLoaded` event. Script execution errors are handled by the script executor; bridge or DOM API initialization errors become `PageErrorMsg` errors.
4. Capture the resulting runtime for the tab and release the lifecycle lock.
5. Fetch external CSS and images, compute styles at the viewport width, build layout boxes, update the JavaScript context's computed-style and layout data when that context is still current, and render both plain and ANSI output.
6. Return `PageLoadedMsg` containing rendered lines, URL, and external CSS.

If `ABYSS_DUMP_ANSI` is set, the ANSI lines are also written to that path; write failures are ignored and a diagnostic is printed to standard error after the write attempt.

`reRenderCmd` reuses a tab's JavaScript runtime and existing DOM/CSS data. It marks the runtime clean, recomputes styles and layout for the requested zoom-adjusted width (clamped to at least 10 cells), updates the runtime's layout information, and returns `PageRenderedMsg`. Missing, stale, or unusable contexts produce no message (`nil`).

## JavaScript runtime lifecycle

JavaScript contexts are process-level package state coordinated by `jsLifecycleMu`. At most one runtime is active in the underlying JS package at a time. Inactive runtimes are captured in `jsRuntimes`, keyed by tab ID, and activated when work for that tab resumes. Switching tabs captures the active runtime first. Closing a tab closes both its active context, if applicable, and its saved runtime. `processJSTimers` activates the requested tab, processes timers only when the current URL matches the expected URL, reports whether the runtime became dirty, and captures it again.

## External resources

### Stylesheets

`fetchExternalCSS` is the background-context wrapper around `fetchExternalCSSContext`. The loader resolves stylesheet links relative to the page URL, skips empty or invalid references and media strings longer than 256 bytes, and deduplicates identical URL-and-media pairs. It fetches at most 64 stylesheets with up to four workers, preserves document order in its result, and applies a total CSS body budget of 4 MiB. Fetch failures and over-budget bodies are omitted. A non-empty media attribute wraps the stylesheet in an `@media` block.

### Images

`fetchImages` wraps `fetchImagesContext` using a background context. Image sources are trimmed and resolved against the page URL; resolved URLs are written back to the DOM so the renderer's lookup key matches the cache key. Data URLs are decoded locally, while external images are fetched with at most two workers. At most 64 distinct external sources are considered, and the aggregate accepted response-body budget is 64 MiB. Images are decoded through `asciirender.Decode`; failed fetches or decodes are not cached.

The image cache is keyed by source URL and guarded by an RW mutex. It has a 64 MiB conservative decoded-image budget, estimating eight bytes per pixel and rejecting invalid or excessively large bounds. Insertion evicts oldest entries until the budget fits. Data URLs must be base64 encoded and are limited to 16 MiB. `render.ImageProvider` is initialized to `cachedImage`, which checks the cache and lazily decodes eligible data URLs.

## Public API

| Name | Behavior |
| --- | --- |
| `ResolveURL(base, rel string) string` | Resolves a relative reference against a base URL. If parsing the base fails, returns `rel`; if parsing the reference fails, returns an empty string. Otherwise returns `baseURL.ResolveReference(relURL).String()`. |
| `ResolveURLForTest(base, rel string) string` | Exported test helper delegating to `ResolveURL`. |
| `FetchExternalCSSForTest(doc *dom.Node, baseURL string) []string` | Exported test helper that fetches stylesheets using a background context. |

The file also uses `PageLoadedMsg`, `PageErrorMsg`, `PageRenderedMsg`, and `tea.Cmd` from the package/application dependencies; their definitions are outside this file. Loading and rendering command constructors are unexported.

## Limits and concurrency

| Resource | Limit |
| --- | ---: |
| Base64 data URL length | 16 MiB |
| Decoded image cache | 64 MiB (8 bytes per pixel estimate) |
| External image sources per page | 64 |
| Aggregate external image response bodies | 64 MiB |
| External stylesheets per page | 64 |
| Aggregate external CSS bodies | 4 MiB |
| Stylesheet media attribute | 256 bytes |
| Stylesheet fetch workers | 4 |
| Image fetch workers | 2 |

Worker results are collected and applied in source order. Context cancellation is passed through to resource fetches; page loading checks cancellation and closes the tab runtime on cancellation at its explicit cleanup points. Shared caches and JavaScript runtime state are protected by separate mutexes.
