// Package fetcher loads web pages over HTTP.
package fetcher

import (
    "fmt"
    "io"
    "net/http"
    "time"
)

// DefaultClient is the shared HTTP client with a reasonable timeout.
var DefaultClient = &http.Client{
    Timeout: 10 * time.Second,
}

// Fetch downloads the body of url and returns it as a string.
func Fetch(url string) (string, error) {
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return "", fmt.Errorf("fetch %s: create request: %w", url, err)
    }
    req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.0 Safari/605.1.15")
    req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
    req.Header.Set("Accept-Language", "en-US,en;q=0.9")

    resp, err := DefaultClient.Do(req)
    if err != nil {
        return "", fmt.Errorf("fetch %s: %w", url, err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("fetch %s: %s", url, resp.Status)
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", fmt.Errorf("fetch %s: read body: %w", url, err)
    }

    return string(body), nil
}

// FetchBytes downloads the raw body of url. Used for binary resources
// like images.
func FetchBytes(url string) ([]byte, string, error) {
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, "", fmt.Errorf("fetch %s: create request: %w", url, err)
    }
    req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.0 Safari/605.1.15")
    req.Header.Set("Accept", "image/avif,image/webp,image/png,image/jpeg,image/*;q=0.8,*/*;q=0.5")

    resp, err := DefaultClient.Do(req)
    if err != nil {
        return nil, "", fmt.Errorf("fetch %s: %w", url, err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, "", fmt.Errorf("fetch %s: %s", url, resp.Status)
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, "", fmt.Errorf("fetch %s: read body: %w", url, err)
    }

    return body, resp.Header.Get("Content-Type"), nil
}
