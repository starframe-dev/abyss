# `internal/fetcher`

The `fetcher` package retrieves web resources over HTTP with resource-specific response-size limits. It exposes helpers for HTML, binary resources, CSS, and JavaScript. Requests use a shared HTTP client and set a browser-style user agent and an `Accept` header appropriate to the requested resource.

## Public API

### Size limits

| Constant | Maximum body size |
| --- | ---: |
| `MaxHTMLBytes` | 8 MiB |
| `MaxImageBytes` | 16 MiB |
| `MaxCSSBytes` | 2 MiB |
| `MaxJSBytes` | 4 MiB |

These limits are enforced both against a known `Content-Length` and while reading the response body. A body larger than the applicable limit is rejected with an error matching `ErrResponseTooLarge`.

### `Fetch` and `FetchContext`

```go
func Fetch(url string) (string, error)
func FetchContext(ctx context.Context, url string) (string, error)
```

Fetch HTML and return its body as a string. `Fetch` uses `context.Background`; `FetchContext` accepts a caller-provided context, allowing cancellation and deadlines. The request advertises HTML, XHTML, and XML media types.

### `FetchBytes` and `FetchBytesContext`

```go
func FetchBytes(url string) ([]byte, string, error)
func FetchBytesContext(ctx context.Context, url string) ([]byte, string, error)
```

Fetch a binary resource, returning the body bytes and the response's `Content-Type` header value. `FetchBytes` uses `context.Background`; `FetchBytesContext` supports caller cancellation and deadlines. These functions use the `MaxImageBytes` limit and advertise common image formats in the `Accept` header.

### `FetchCSS`

```go
func FetchCSS(ctx context.Context, url string) (string, error)
```

Fetch CSS as a string, with the `MaxCSSBytes` limit. The context is required; there is no background-context convenience wrapper.

### `FetchJS`

```go
func FetchJS(ctx context.Context, url string) (string, error)
```

Fetch JavaScript as a string, with the `MaxJSBytes` limit. The context is required; there is no background-context convenience wrapper.

### Request configuration

```go
const UserAgent = "..."
var DefaultClient = &http.Client{Timeout: 10 * time.Second}
```

`UserAgent` is sent on every request. `DefaultClient` is the shared client used for all requests and has a ten-second timeout by default. Since it is an exported variable, callers can replace or configure it; concurrent mutation should be avoided.

## Errors and response handling

- Requests are made with HTTP `GET` and the supplied context.
- Only `200 OK` is accepted. Other status codes produce an error containing the URL and status text.
- Request creation and transport failures are wrapped with URL context.
- Oversized responses return an error wrapping `ErrResponseTooLarge`, so callers can detect it with `errors.Is`.
- Body read failures return a `*ResponseReadError`, which preserves the underlying error for `errors.Is`/`errors.As` and records the number of bytes read.
- The response body is closed after processing.
- On success, the internal fetch operation returns the response body and `Content-Type`; string-oriented helpers discard the content type.

## Types

### `ResponseReadError`

```go
type ResponseReadError struct {
    BytesRead int64
    Err       error
}
```

Describes a response-body read failure. `Error` returns the underlying error's message, and `Unwrap` exposes that error for standard error-chain inspection. `BytesRead` is the length of the partial body returned by the read operation.

### `ErrResponseTooLarge`

Sentinel error for responses exceeding the selected byte limit. Errors returned for this condition wrap the sentinel.

## Implementation details

The package's shared internal `get` function constructs a context-bound request, applies the common headers, performs the request with `DefaultClient`, validates status and declared size, and reads at most `limit + 1` bytes. Reading one byte beyond the limit allows the implementation to detect oversized responses even when the server omits or understates `Content-Length`. Response bodies are converted directly to strings by the text helpers; no character-set decoding or content-type validation is performed.
