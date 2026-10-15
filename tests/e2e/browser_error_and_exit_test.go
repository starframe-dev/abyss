package e2e

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/starframe-dev/cue-tty/pkg/cue"
)

func TestBrowserErrorAndCleanExit(t *testing.T) {
	app, err := cue.Launch("./bin/abyss",
		cue.WithSize(120, 40),
		cue.WithDir(projectRoot(t)),
		cue.WithEnv("TERM=xterm-256color", "HOME="+t.TempDir()),
	)
	if err != nil {
		t.Fatal(err)
	}

	page := app.Page()
	page.WaitStable(2 * time.Second)
	if err := page.Press("Ctrl+l"); err != nil {
		_ = app.Close()
		t.Fatalf("focus address bar: %v", err)
	}
	if err := page.Type("http://127.0.0.1:1/unreachable"); err != nil {
		_ = app.Close()
		t.Fatalf("type unreachable URL: %v", err)
	}
	if err := page.Press("Enter"); err != nil {
		_ = app.Close()
		t.Fatalf("submit unreachable URL: %v", err)
	}

	deadline := time.Now().Add(15 * time.Second)
	var errorScreen string
	for time.Now().Before(deadline) {
		screen, textErr := page.Text()
		output := string(app.RawLog())
		combined := screen + "\n" + output
		lower := strings.ToLower(combined)
		if textErr == nil && (strings.Contains(lower, "error") || strings.Contains(lower, "failed") || strings.Contains(lower, "connection refused")) {
			errorScreen = combined
			break
		}
		if bytes.Contains(app.RawLog(), []byte("ERR")) {
			errorScreen = combined
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if errorScreen == "" {
		_ = app.Close()
		t.Fatalf("expected an error to be displayed for unreachable URL; terminal output:\n%s", app.RawLog())
	}

	if err := page.Press("q"); err != nil {
		_ = app.Close()
		t.Fatalf("quit browser with q: %v", err)
	}
	if err := app.Close(); err != nil {
		t.Fatalf("close browser cleanly after q: %v", err)
	}
}
