package browser

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHeadlessDump(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte("<!doctype html><html><body><p>Headless dump test</p></body></html>"))
	}))
	defer server.Close()

	output := filepath.Join(t.TempDir(), "dump.ansi")
	HeadlessDump(server.URL, 24, output)

	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty dump")
	}
	if !strings.Contains(string(data), "Headless dump test") {
		t.Fatalf("dump does not contain page text: %q", data)
	}
}
