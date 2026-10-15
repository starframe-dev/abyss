package e2e

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestHeadlessDumpWritesANSIAtRequestedWidth(t *testing.T) {
	const width = 64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!doctype html><html><head><style>body{background:#123456;color:#ffffff}</style></head><body><h1>Deterministic headless fixture</h1><p>ANSI output at requested width.</p></body></html>`))
	}))
	defer server.Close()

	outputPath := filepath.Join(t.TempDir(), "page.ansi")
	cmd := exec.Command("go", "run", "../..", "--headless-dump", outputPath, "--url", server.URL, "--width", "64")
	cmd.Dir = projectRoot(t)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("headless-dump command failed: %v\n%s", err, output)
	}

	dump, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read ANSI dump: %v", err)
	}
	content := string(dump)
	if !strings.Contains(content, "Deterministic headless fixture") {
		t.Fatalf("dump does not contain fixture headline:\n%s", content)
	}
	if !regexp.MustCompile("\\x1b\\[[0-9;]*m").Match(dump) {
		t.Fatalf("dump contains no ANSI color escape sequences:\n%s", content)
	}

	lines := strings.Split(strings.TrimRight(content, "\r\n"), "\n")
	if len(lines) == 0 {
		t.Fatal("dump is empty")
	}
	for i, line := range lines {
		// ANSI styling sequences do not occupy terminal columns.
		plain := regexp.MustCompile("\\x1b\\[[0-9;]*m").ReplaceAllString(line, "")
		if got := len([]rune(plain)); got != width {
			t.Errorf("line %d has width %d, want %d", i+1, got, width)
		}
	}
}
