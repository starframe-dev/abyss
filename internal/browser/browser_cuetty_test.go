package browser

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
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

var ansiRegexp = regexp.MustCompile(`\033\[[0-9;]*[a-zA-Z]`)
var trueColorRegexp = regexp.MustCompile(`\033\[(?:38|48);2;[0-9;]+m`)

func stripANSI(b []byte) []byte {
	return ansiRegexp.ReplaceAll(b, []byte{})
}

func submitFixtureURL(page *cue.Page, pageURL string) error {
	var screen string
	for attempt := 0; attempt < 2; attempt++ {
		var err error
		screen, err = page.Text()
		if err != nil {
			return fmt.Errorf("read address screen: %w", err)
		}
		if !strings.Contains(screen, pageURL) {
			for range len([]rune(pageURL)) + 4 {
				if err := page.Press("Backspace"); err != nil {
					return fmt.Errorf("clear address: %w", err)
				}
			}
			if err := page.Type(pageURL); err != nil {
				return fmt.Errorf("type fixture URL: %w", err)
			}
			page.WaitStable(500 * time.Millisecond)
			screen, err = page.Text()
			if err != nil || !strings.Contains(screen, pageURL) {
				continue
			}
		}
		if err := page.Press("Enter"); err != nil {
			return fmt.Errorf("submit fixture URL: %w", err)
		}
		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			page.WaitStable(100 * time.Millisecond)
			screen, err = page.Text()
			if err != nil {
				return fmt.Errorf("read post-submit screen: %w", err)
			}
			if strings.Contains(screen, "Loading...") || strings.Contains(screen, "Example Domain") || strings.Contains(screen, "Error:") {
				return nil
			}
			if strings.TrimSpace(screen) == "" {
				continue
			}
			if !strings.Contains(screen, "Abyss >") {
				return fmt.Errorf("unexpected screen after URL submission: %q", screen)
			}
		}
	}
	return fmt.Errorf("address did not begin loading after retry: %q", screen)
}

// fixturePageHTML mirrors the structure that the E2E assertions depend on:
// a heading, a paragraph ending with "operations.", and a trailing paragraph
// with "Learn more". Served locally so tests do not depend on live sites.
const fixturePageHTML = `<!doctype html>
<html>
<head><title>Example Domain</title></head>
<body>
<div>
<h1>Example Domain</h1>
<p>This domain is for use in illustrative examples in documents. It is not a service for other operations.</p>
<p><a href="#more">Learn more</a></p>
</div>
</body>
</html>`

// startFixtureServer serves fixturePageHTML on a random local port.
func startFixtureServer(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(fixturePageHTML))
	}))
	t.Cleanup(server.Close)
	return server
}

func TestAbyss_E2E_LoadPage(t *testing.T) {
	server := startFixtureServer(t)

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

	// Test 1: Type URL of the local fixture server and ensure submission starts a load.
	if err := submitFixtureURL(page, server.URL); err != nil {
		t.Fatal(err)
	}

	// Wait for page to load by checking raw log with ANSI stripped
	deadline := time.Now().Add(15 * time.Second)
	found := false
	for time.Now().Before(deadline) {
		rawLog := app.RawLog()
		cleanLog := stripANSI(rawLog)
		if bytes.Contains(cleanLog, []byte("Example Domain")) {
			found = true
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if !found {
		t.Logf("RAW LOG:\n%s", app.RawLog())
		app.Close()
		t.Fatal("expected 'Example Domain' in raw log")
	}

	// Test 2: Quit works
	page.Press("q")
	page.WaitStable(500 * time.Millisecond)

	if err := app.Close(); err != nil {
		t.Fatalf("expected clean exit, got: %v", err)
	}
}

// TestAbyss_E2E_LargeViewport_BlockStacking regression-tests the bug where
// HTML5 void elements (<meta>, <link>) without trailing slash would swallow
// all following siblings as their children, causing the entire page to render
// as a single concatenated text line. With proper stacking, the body's block
// children (h1, p, p) must each appear on their own line in the viewport.
func TestAbyss_E2E_LargeViewport_BlockStacking(t *testing.T) {
	server := startFixtureServer(t)

	app, err := cue.Launch("./bin/abyss",
		cue.WithSize(200, 60),
		cue.WithDir(projectRoot(t)),
		cue.WithEnv("TERM=xterm-256color", "HOME="+t.TempDir()),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	page := app.Page()
	page.WaitStable(2 * time.Second)

	page.Type(server.URL)
	page.WaitStable(300 * time.Millisecond)
	page.Press("Enter")

	// Wait for content to render.
	deadline := time.Now().Add(15 * time.Second)
	ready := false
	for time.Now().Before(deadline) {
		cleanLog := stripANSI(app.RawLog())
		if bytes.Contains(cleanLog, []byte("Learn more")) {
			ready = true
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if !ready {
		t.Logf("RAW LOG:\n%s", app.RawLog())
		t.Fatal("expected 'Learn more' in raw log (page didn't render fully)")
	}

	// Assertion: h1 and the trailing <p> containing "Learn more" must appear
	// on DIFFERENT lines in the rendered output. Before the fix they were
	// merged into a single concatenated line by a single TextLine inside <html>.
	cleanLog := stripANSI(app.RawLog())
	h1Idx := bytes.Index(cleanLog, []byte("Example Domain"))
	h1SecondIdx := -1
	if h1Idx >= 0 {
		h1SecondIdx = bytes.Index(cleanLog[h1Idx+1:], []byte("Example Domain"))
		if h1SecondIdx >= 0 {
			h1SecondIdx += h1Idx + 1
		}
	}
	// The old bug produced "Example DomainExample Domain" — exact duplicate
	// with no separator, indicating two text sources were merged.
	if h1SecondIdx >= 0 {
		between := cleanLog[h1Idx+len("Example Domain") : h1SecondIdx]
		if len(bytes.TrimSpace(between)) == 0 {
			t.Fatalf("regression: 'Example Domain' appears twice with no whitespace between — block stacking broken. between=%q", between)
		}
	}

	// "Learn more" must appear on its own line (after a newline), not glued
	// to the preceding paragraph text.
	learnIdx := bytes.Index(cleanLog, []byte("Learn more"))
	if learnIdx < 0 {
		t.Fatal("expected 'Learn more' not found")
	}
	// Walk backwards to find the start of this line
	lineStart := learnIdx
	for lineStart > 0 && cleanLog[lineStart-1] != '\n' {
		lineStart--
	}
	line := cleanLog[lineStart : learnIdx+len("Learn more")]
	if bytes.Contains(line, []byte("operations.")) {
		t.Fatalf("regression: 'Learn more' is on the same line as the previous paragraph. line=%q", line)
	}

	page.Press("q")
	page.WaitStable(300 * time.Millisecond)
}

// TestAbyss_E2E_ExtraLargeViewport_TrueColor validates the complete rendering
// path (HTTP -> DOM -> CSS -> layout -> ANSI -> PTY) at 320x120. The test uses
// a local deterministic server and checks visible content plus exact RGB codes.
func TestAbyss_E2E_ExtraLargeViewport_TrueColor(t *testing.T) {
	const colorfulHTML = `<!doctype html>
<html>
<head>
<style>
body { background-color: #1e293b; color: #ffffff; margin: 0; padding: 8px; }
h1 { color: #fbbf24; background-color: #312e81; }
a { color: #60a5fa; }
.red { color: #ffffff; background-color: #ef4444; padding: 4px; }
.green { color: #1f2937; background-color: #10b981; padding: 4px; }
.purple { color: #fef3c7; background-color: #9333ea; padding: 4px; }
</style>
</head>
<body>
<h1>320x120 True Color Test</h1>
<p>Visit <a href="#target">the blue link</a>.</p>
<div class="red">Red panel</div>
<div class="green">Green panel</div>
<div class="purple">Purple panel</div>
</body>
</html>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(colorfulHTML))
	}))
	defer server.Close()

	app, err := cue.Launch("./bin/abyss",
		cue.WithSize(320, 120),
		cue.WithDir(projectRoot(t)),
		cue.WithEnv("TERM=xterm-truecolor", "HOME="+t.TempDir()),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer app.Close()

	page := app.Page()
	page.WaitStable(2 * time.Second)
	page.Type(server.URL)
	page.WaitStable(300 * time.Millisecond)
	page.Press("Enter")

	deadline := time.Now().Add(15 * time.Second)
	ready := false
	for time.Now().Before(deadline) {
		cleanLog := stripANSI(app.RawLog())
		if bytes.Contains(cleanLog, []byte("320x120 True Color Test")) &&
			bytes.Contains(cleanLog, []byte("Purple panel")) {
			ready = true
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if !ready {
		t.Logf("RAW LOG:\n%s", app.RawLog())
		t.Fatal("color test page did not render completely at 320x120")
	}

	rawLog := app.RawLog()
	expectedColors := map[string][]byte{
		"yellow heading foreground": []byte("38;2;251;191;36"),
		"indigo heading background": []byte("48;2;49;46;129"),
		"blue link foreground":      []byte("38;2;96;165;250"),
		"red panel background":      []byte("48;2;239;68;68"),
		"green panel background":    []byte("48;2;16;185;129"),
		"purple panel background":   []byte("48;2;147;51;234"),
	}
	for name, code := range expectedColors {
		if !bytes.Contains(rawLog, code) {
			t.Errorf("missing %s ANSI code %q", name, code)
		}
	}
	if t.Failed() {
		t.Logf("true-color codes found: %q", trueColorRegexp.FindAll(rawLog, -1))
	}

	artifactDir := os.Getenv("ABYSS_TEST_ARTIFACT_DIR")
	if artifactDir == "" {
		artifactDir = t.TempDir()
	}
	if err := page.SaveArtifact(artifactDir, "320x120-true-color"); err != nil {
		t.Errorf("save cue-tty artifact: %v", err)
	}
	if err := os.WriteFile(filepath.Join(artifactDir, "320x120-true-color.ansi"), rawLog, 0o644); err != nil {
		t.Errorf("save raw ANSI artifact: %v", err)
	}
	t.Logf("320x120 color artifacts: %s", artifactDir)
}

func TestAbyss_E2E_ErrorPage(t *testing.T) {
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
	page.WaitStable(300 * time.Millisecond)

	page.Type("not-a-valid-url-12345.xyz")
	page.WaitStable(100 * time.Millisecond)
	page.Press("Enter")

	// Wait for error by checking raw log with ANSI stripped
	deadline := time.Now().Add(10 * time.Second)
	found := false
	for time.Now().Before(deadline) {
		rawLog := app.RawLog()
		cleanLog := stripANSI(rawLog)
		if bytes.Contains(cleanLog, []byte("Error:")) {
			found = true
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if !found {
		t.Fatal("expected 'Error:' in raw log")
	}

	page.Press("q")
	page.WaitStable(200 * time.Millisecond)
}
