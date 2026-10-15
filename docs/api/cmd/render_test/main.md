# `cmd/render_test`

`cmd/render_test` is a development-only command that fetches a web page and renders it as ANSI-colored text art. It is intended for inspecting the ASCII image renderer and writes both colored and plain-text fixtures to `/tmp`.

## Usage

```sh
go run ./cmd/render_test <url> [width]
```

- `<url>` is required and is passed to the project fetcher.
- `[width]` is an optional integer rendering width; the default is `120`.
- If the URL is missing, the command prints usage to standard error and exits with status `2`.
- A page-fetch error is printed to standard error and exits with status `1`.

The optional width is parsed with `fmt.Sscanf`; its error is ignored. Consequently, an invalid width leaves the default or any value partially parsed by `Sscanf` in use.

## Processing behavior

The command performs these steps:

1. Initializes a process-local image cache and installs `render.ImageProvider`, which returns cached images or `nil` for a cache miss.
2. Fetches the page with `fetcher.Fetch` and parses the HTML into a DOM.
3. Finds `<img>` elements, resolves and deduplicates their sources, fetches or decodes supported image data, and stores successfully decoded images in the cache.
4. Computes CSS, builds a layout using the requested width, renders the layout, and converts the framebuffer to ANSI output.
5. Writes the ANSI output to `/tmp/img_fixture_ansi.txt` and a version with ANSI escape sequences removed and trailing spaces trimmed from each line to `/tmp/img_fixture_plain.txt`.
6. Prints the output paths and each non-empty plain-text line, prefixed with its zero-based line index.

Errors from image retrieval, image decoding, and fixture writes are intentionally ignored. Failed images therefore remain absent from the cache, and failed writes do not stop the command or alter its final status.

## Image handling details

Image source values are trimmed and empty values are skipped. `data:` sources are used as-is; other sources are passed to `resolveURL`. Duplicate resolved source strings are fetched or decoded only once per invocation.

Only data URLs with the exact prefix `data:image/png;base64,` use the dedicated base64 path. Other sources are loaded using `fetcher.FetchBytes` and decoded by `asciirender.Decode`. The cache is keyed by the resolved source string. An `sync.RWMutex` protects cache access, including reads performed by the renderer's provider callback.

`resolveURL` returns absolute `http://` and `https://` values unchanged. For a leading-slash path it concatenates the base URL and path; for other relative values it returns the base URL unchanged. This helper does not perform general URL resolution.

## Functions and state

All declarations are package-private; this command exposes no public Go API.

| Name | Role |
| --- | --- |
| `main()` | Parses arguments and orchestrates fetch, parse, image preparation, rendering, and fixture output. |
| `fetchImages(doc, baseURL)` | Collects image sources from the DOM, loads supported images, and populates the cache. |
| `resolveURL(base, rel)` | Applies the command's limited URL handling for image sources. |
| `decodeBase64PNG(b64)` | Base64-decodes a PNG data payload and delegates image decoding to `asciirender.Decode`. |
| `decodeBase64(s)` | Decodes standard-alphabet base64 using `encoding/base64`. |
| `stripANSI(s)` | Removes supported ANSI SGR escape sequences from a string for the plain-text fixture. |
| `imageCache` | Maps source strings to decoded `image.Image` values. |
| `imageCacheMu` | Read/write mutex guarding `imageCache`. |

`stripANSI` recognizes escape sequences beginning with ESC followed by `[` and numeric/semicolon parameters ending in `m`. Other ESC-prefixed sequences are skipped according to its byte-wise parser; it is not a general terminal-control-sequence sanitizer.
