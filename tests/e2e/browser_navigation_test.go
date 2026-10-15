package e2e

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/starframe-dev/cue-tty/pkg/cue"
)

func projectRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate project root")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "../.."))
}

func TestBrowserNavigation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!doctype html><html><body><h1>Local navigation fixture</h1><p>The page rendered successfully.</p></body></html>`))
	}))
	defer server.Close()

	app, err := cue.Launch("./bin/abyss",
		cue.WithSize(120, 40),
		cue.WithDir(projectRoot(t)),
		cue.WithEnv("TERM=xterm-256color", "HOME="+t.TempDir()),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	page := app.Page()
	page.WaitStable(2 * time.Second)
	if err := page.Press("Ctrl+l"); err != nil {
		t.Fatalf("focus address bar: %v", err)
	}
	if err := page.Type(server.URL); err != nil {
		t.Fatalf("type fixture URL: %v", err)
	}
	if err := page.Press("Enter"); err != nil {
		t.Fatalf("submit fixture URL: %v", err)
	}

	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		screen, textErr := page.Text()
		if textErr == nil && strings.Contains(screen, "Local navigation fixture") {
			return
		}
		if bytes.Contains(app.RawLog(), []byte("Local navigation fixture")) {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatalf("expected fixture page to render; terminal output:\n%s", app.RawLog())
}
