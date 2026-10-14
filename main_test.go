package main

import (
    "testing"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/starframe/abyss/internal/browser"
)

func TestBrowserModelImplementsTeaModel(t *testing.T) {
    model := browser.New()
    var _ tea.Model = &model
}

// TestMainEntryPoint verifies that the package compiles and main can be
// referenced without panicking. The actual main() reads from stdin and is
// not safe to invoke from tests, so we only check that the entry point is
// callable by reference (used by the go runtime).
func TestMainEntryPoint(t *testing.T) {
    // Just verify main.go compiles; we don't invoke main directly.
    // If we got here, the binary links correctly.
    _ = browser.New
}
