package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestMainHeadlessDumpRequiresURL(t *testing.T) {
	if os.Getenv("ABYSS_MAIN_HELPER") == "1" {
		os.Args = []string{"abyss", "-headless-dump", "unused.ansi"}
		main()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestMainHeadlessDumpRequiresURL$")
	cmd.Env = append(os.Environ(), "ABYSS_MAIN_HELPER=1")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected main to exit with status 2 when --url is missing")
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok || exitErr.ExitCode() != 2 {
		t.Fatalf("expected exit code 2, got error %v; output: %s", err, output)
	}
	if !strings.Contains(string(output), "--headless-dump requires --url") {
		t.Fatalf("expected missing URL message, got %q", output)
	}
}
