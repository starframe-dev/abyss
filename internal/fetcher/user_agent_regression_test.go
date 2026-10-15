package fetcher

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Google serves the logo variant only to simple agents, so every request must
// send the terminal browser User-Agent rather than a desktop Safari string.
func TestFetchSendsTerminalBrowserUserAgent(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("User-Agent")
		w.Write([]byte("<html></html>"))
	}))
	defer srv.Close()

	if _, err := Fetch(srv.URL); err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if got != UserAgent {
		t.Fatalf("User-Agent = %q, want %q", got, UserAgent)
	}
}
