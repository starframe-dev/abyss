package e2e

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/starframe-dev/cue-tty/pkg/cue"
)

func tailBytes(b []byte, n int) []byte {
	if len(b) <= n {
		return b
	}
	return b[len(b)-n:]
}

func TestBrowserScrolling(t *testing.T) {
	const sections = 80
	var document strings.Builder
	document.WriteString(`<!doctype html><html><body><h1>Scroll fixture top</h1>`)
	for i := 1; i <= sections; i++ {
		fmt.Fprintf(&document, `<p>Scroll section %02d: local page content.</p>`, i)
	}
	document.WriteString(`<p>Scroll fixture bottom</p></body></html>`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(document.String()))
	}))
	defer server.Close()

	app, err := cue.Launch("./bin/abyss",
		cue.WithSize(100, 24),
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

	waitFor := func(want string) string {
		t.Helper()
		deadline := time.Now().Add(15 * time.Second)
		for time.Now().Before(deadline) {
			screen, textErr := page.Text()
			if textErr == nil && strings.Contains(screen, want) {
				return screen
			}
			if bytes.Contains(app.RawLog(), []byte(want)) {
				return string(app.RawLog())
			}
			time.Sleep(200 * time.Millisecond)
		}
		screen, _ := page.Text()
		t.Fatalf("timed out waiting for %q; screen:\n%s\nraw log tail:\n%s", want, screen, tailBytes(app.RawLog(), 400))
		return ""
	}

	waitFor("Scroll fixture top")
	// Keys are paced: the viewport handles each one before the next arrives.
	pace := func() { time.Sleep(700 * time.Millisecond) }
	if err := page.Press("ArrowDown"); err != nil {
		t.Fatalf("scroll down with arrow key: %v", err)
	}
	pace()
	if err := page.Press("PageDown"); err != nil {
		t.Fatalf("scroll down with Page Down: %v", err)
	}
	waitFor("Scroll section 15")
	pace()
	// The document is longer than three pages, so keep paging until the last line is on screen.
	for attempt := 0; attempt < 20; attempt++ {
		screen, _ := page.Text()
		if strings.Contains(screen, "Scroll fixture bottom") {
			break
		}
		if err := page.Press("PageDown"); err != nil {
			t.Fatalf("scroll to bottom with Page Down: %v", err)
		}
		pace()
	}
	waitFor("Scroll fixture bottom")
	pace()

	if err := page.Press("PageUp"); err != nil {
		t.Fatalf("scroll up with Page Up: %v", err)
	}
	pace()
	if err := page.Press("ArrowUp"); err != nil {
		t.Fatalf("scroll up with arrow key: %v", err)
	}
	waitFor("Scroll section 60")
}
