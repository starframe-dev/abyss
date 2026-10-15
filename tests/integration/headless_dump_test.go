package integration

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHeadlessDumpCLI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!doctype html><html><body><h1>Headless integration headline</h1><p>Rendered terminal content</p></body></html>`))
	}))
	defer server.Close()

	outputPath := filepath.Join(t.TempDir(), "dump.ansi")
	cmd := exec.Command("go", "run", "../..", "--headless-dump", outputPath, "--url", server.URL, "--width", "80")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "CLI failed: %s", output)

	dump, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	require.NotEmpty(t, dump)
	require.Contains(t, string(dump), "Headless integration headline")
	require.Contains(t, string(dump), "Rendered terminal content")
	require.Contains(t, string(output), "headless dump:")
}

func TestHeadlessDumpRequiresURL(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "dump.ansi")
	cmd := exec.Command("go", "run", "../..", "--headless-dump", outputPath)
	output, err := cmd.CombinedOutput()
	require.Error(t, err)
	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 2, exitErr.ExitCode(), string(output))
	require.True(t, strings.Contains(string(output), "--headless-dump requires --url"), string(output))
	_, statErr := os.Stat(outputPath)
	require.True(t, os.IsNotExist(statErr))
}
