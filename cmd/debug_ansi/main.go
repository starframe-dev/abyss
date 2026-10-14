package main

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/starframe-dev/cue-tty/pkg/cue"
)

var ansiRegexp = regexp.MustCompile(`\033\[[0-9;]*[a-zA-Z]`)

func stripANSI(s string) string {
	return ansiRegexp.ReplaceAllString(s, "")
}

func main() {
	app, err := cue.Launch("./bin/abyss",
		cue.WithSize(120, 40),
		cue.WithDir("."),
		cue.WithEnv("TERM=xterm-256color"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: launch: %v\n", err)
		os.Exit(1)
	}
	defer app.Close()

	page := app.Page()
	page.WaitStable(2 * time.Second)

	// Check initial screen
	text, _ := page.Text()
	fmt.Printf("Page text (%d chars):\n%s\n", len(text), text[:min(len(text), 500)])

	rawLog := app.RawLog()
	cleanLog := stripANSI(string(rawLog))
	fmt.Printf("\nClean log (%d bytes):\n%s\n", len(cleanLog), cleanLog[:min(len(cleanLog), 500)])

	// Check for "Abyss" in clean log
	if bytes.Contains([]byte(cleanLog), []byte("Abyss")) {
		fmt.Println("OK: 'Abyss' found in clean log")
	} else {
		fmt.Println("FAIL: 'Abyss' NOT found in clean log")
	}

	// Check for "Enter URL" in clean log
	if bytes.Contains([]byte(cleanLog), []byte("Enter URL")) {
		fmt.Println("OK: 'Enter URL' found in clean log")
	} else {
		fmt.Println("FAIL: 'Enter URL' NOT found in clean log")
	}

	// Load example.com
	page.Type("example.com")
	page.WaitStable(300 * time.Millisecond)
	page.Press("Enter")

	deadline := time.Now().Add(15 * time.Second)
	found := false
	for time.Now().Before(deadline) {
		rawLog := app.RawLog()
		cleanLog := stripANSI(string(rawLog))
		if bytes.Contains([]byte(cleanLog), []byte("Example Domain")) {
			found = true
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if found {
		fmt.Println("OK: 'Example Domain' found in clean log")
	} else {
		fmt.Println("FAIL: 'Example Domain' NOT found in clean log")
	}

	page.Press("q")
	page.WaitStable(500 * time.Millisecond)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
