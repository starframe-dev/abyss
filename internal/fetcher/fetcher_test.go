package fetcher

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestFetch_Success(t *testing.T) {
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("<html><body>Hello</body></html>"))
    }))
    defer ts.Close()

    body, err := Fetch(ts.URL)
    require.NoError(t, err)
    assert.Contains(t, body, "Hello")
}

func TestFetch_NotFound(t *testing.T) {
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusNotFound)
    }))
    defer ts.Close()

    _, err := Fetch(ts.URL)
    require.Error(t, err)
    assert.Contains(t, err.Error(), "404")
}

func TestFetch_NetworkError(t *testing.T) {
    _, err := Fetch("http://127.0.0.1:1")
    require.Error(t, err)
}

func TestFetchBytes_BinaryBody(t *testing.T) {
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "image/png")
        _, _ = w.Write([]byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a})
    }))
    defer ts.Close()

    body, contentType, err := FetchBytes(ts.URL + "/test.png")
    require.NoError(t, err)
    assert.Equal(t, 8, len(body))
    assert.Equal(t, "image/png", contentType)
}

func TestFetchBytes_RejectsNonOK(t *testing.T) {
    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusNotFound)
    }))
    defer ts.Close()

    _, _, err := FetchBytes(ts.URL)
    require.Error(t, err)
}
