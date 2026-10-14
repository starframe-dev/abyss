package main

import (
	"fmt"
	"time"

	"github.com/starframe-dev/cue-tty/pkg/cue"
)

func main() {
	app, err := cue.Launch("./bin/abyss",
		cue.WithSize(200, 60),
		cue.WithDir("."),
		cue.WithEnv("TERM=xterm-256color"),
	)
	if err != nil {
		fmt.Printf("Launch error: %v\n", err)
		return
	}
	defer app.Close()

	page := app.Page()
	time.Sleep(1 * time.Second)

	// Load example.com
	page.Type("example.com")
	time.Sleep(200 * time.Millisecond)
	page.Press("Enter")

	// Wait for page to load
	time.Sleep(5 * time.Second)

	// Dump page lines
	lines, _ := page.Lines()
	fmt.Printf("=== PAGE LINES (%d lines, 200x60) ===\n", len(lines))
	for i, line := range lines {
		trimmed := line
		for len(trimmed) > 0 && trimmed[len(trimmed)-1] == ' ' {
			trimmed = trimmed[:len(trimmed)-1]
		}
		if trimmed == "" {
			fmt.Printf("[%2d] (empty)\n", i)
		} else {
			fmt.Printf("[%2d] %q\n", i, trimmed)
		}
	}

	// Also dump raw log for ANSI content
	raw := app.RawLog()
	fmt.Printf("\n=== RAW LOG (last 2000 bytes) ===\n")
	start := len(raw) - 2000
	if start < 0 {
		start = 0
	}
	for i := start; i < len(raw); i++ {
		b := raw[i]
		if b == '\033' {
			fmt.Printf("\\033")
		} else if b == '\n' {
			fmt.Printf("\\n\n")
		} else if b == '\r' {
		} else if b >= 32 && b <= 126 {
			fmt.Printf("%c", b)
		} else {
			fmt.Printf("\\x%02x", b)
		}
	}
	fmt.Println()

	page.Press("q")
	time.Sleep(200 * time.Millisecond)
}
