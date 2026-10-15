// Package fetcher loads bounded web resources over HTTP.
package fetcher

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

var ErrResponseTooLarge = errors.New("response exceeds configured size limit")

type ResponseReadError struct {
	BytesRead int64
	Err       error
}

func (e *ResponseReadError) Error() string { return e.Err.Error() }
func (e *ResponseReadError) Unwrap() error { return e.Err }

const (
	MaxHTMLBytes  int64 = 8 << 20
	MaxImageBytes int64 = 16 << 20
	MaxCSSBytes   int64 = 2 << 20
	MaxJSBytes    int64 = 4 << 20
)

// UserAgent identifies the terminal browser. Google serves a page with the
// logo (id="hplogo") to simple agents, while full Chrome/Safari agents get a
// script-heavy variant without it.
const UserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"

// DefaultClient is the shared HTTP client with a reasonable timeout.
var DefaultClient = &http.Client{Timeout: 10 * time.Second}

func get(ctx context.Context, resourceURL string, limit int64, accept string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, resourceURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("fetch %s: create request: %w", resourceURL, err)
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", accept)
	resp, err := DefaultClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("fetch %s: %w", resourceURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("fetch %s: %s", resourceURL, resp.Status)
	}
	if resp.ContentLength > limit {
		return nil, "", fmt.Errorf("%w: fetch %s: response exceeds %d bytes", ErrResponseTooLarge, resourceURL, limit)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, "", &ResponseReadError{BytesRead: int64(len(body)), Err: fmt.Errorf("fetch %s: read body: %w", resourceURL, err)}
	}
	if int64(len(body)) > limit {
		return nil, "", fmt.Errorf("%w: fetch %s: response exceeds %d bytes", ErrResponseTooLarge, resourceURL, limit)
	}
	return body, resp.Header.Get("Content-Type"), nil
}

// Fetch downloads bounded HTML using a background context.
func Fetch(url string) (string, error) { return FetchContext(context.Background(), url) }

// FetchContext downloads bounded HTML and supports cancellation.
func FetchContext(ctx context.Context, url string) (string, error) {
	body, _, err := get(ctx, url, MaxHTMLBytes, "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	return string(body), err
}

// FetchBytes downloads a bounded binary resource using a background context.
func FetchBytes(url string) ([]byte, string, error) {
	return FetchBytesContext(context.Background(), url)
}

// FetchBytesContext downloads a bounded binary resource and supports cancellation.
func FetchBytesContext(ctx context.Context, url string) ([]byte, string, error) {
	return get(ctx, url, MaxImageBytes, "image/avif,image/webp,image/png,image/jpeg,image/*;q=0.8,*/*;q=0.5")
}

// FetchCSS downloads bounded CSS and supports cancellation.
func FetchCSS(ctx context.Context, url string) (string, error) {
	body, _, err := get(ctx, url, MaxCSSBytes, "text/css,*/*;q=0.1")
	return string(body), err
}

// FetchJS downloads bounded JavaScript and supports cancellation.
func FetchJS(ctx context.Context, url string) (string, error) {
	body, _, err := get(ctx, url, MaxJSBytes, "application/javascript,text/javascript,*/*;q=0.1")
	return string(body), err
}
