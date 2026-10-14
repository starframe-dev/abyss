package main

import (
	"fmt"
	"os"
	"time"

	"github.com/starframe-dev/cue-tty/pkg/cue"
)

func main() {
	// Test 1: Initial screen
	fmt.Println("=== Test 1: Initial screen ===")
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

	text, _ := page.Text()
	fmt.Printf("Initial screen (%d chars):\n%s\n", len(text), text)

	if err := cue.Expect(page).ToContain("Abyss"); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: no 'Abyss' in initial screen: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("PASS: Abyss title visible")

	if err := cue.Expect(page).ToContain("Enter URL"); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: no 'Enter URL' in initial screen: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("PASS: Address bar visible")

	// Test 2: Load example.com
	fmt.Println("\n=== Test 2: Load example.com ===")
	page.Type("example.com")
	page.WaitStable(300 * time.Millisecond)
	page.Press("Enter")
	page.WaitFor("Example Domain", 15*time.Second)

	text, _ = page.Text()
	fmt.Printf("Page loaded (%d chars):\n%s\n", len(text), text)

	if err := cue.Expect(page).ToContain("Example Domain"); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: no 'Example Domain': %v\n", err)
		os.Exit(1)
	}
	fmt.Println("PASS: Example Domain visible")

	// Test 3: Scrolling
	fmt.Println("\n=== Test 3: Scrolling ===")
	before, _ := page.Text()
	page.Press("Down")
	page.Press("Down")
	page.Press("Down")
	page.Press("Down")
	page.Press("Down")
	page.Press("Down")
	page.Press("Down")
	page.Press("Down")
	page.WaitStable(300 * time.Millisecond)
	after, _ := page.Text()
	if before == after {
		fmt.Fprintf(os.Stderr, "FAIL: screen didn't change after scrolling\n")
		os.Exit(1)
	}
	fmt.Println("PASS: Scrolling works")

	// Test 4: Manual refresh with 'r' key
	fmt.Println("\n=== Test 4: Manual refresh (r key) ===")
	page.Press("r")
	page.WaitStable(2 * time.Second)

	text, _ = page.Text()
	fmt.Printf("After refresh (%d chars):\n%s\n", len(text), text)

	if err := cue.Expect(page).ToContain("Example Domain"); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: no 'Example Domain' after refresh: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("PASS: Manual refresh works")

	// Test 5: Error page
	fmt.Println("\n=== Test 5: Error page ===")
	page.Press("Escape")
	page.WaitStable(200 * time.Millisecond)
	page.Type("not-a-valid-url-12345.xyz")
	page.WaitStable(100 * time.Millisecond)
	page.Press("Enter")
	page.WaitFor("Error:", 10*time.Second)

	if err := cue.Expect(page).ToContain("Error:"); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: no 'Error:' on error page: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("PASS: Error page works")

	// Test 6: Quit
	fmt.Println("\n=== Test 6: Quit ===")
	page.Press("q")
	page.WaitStable(500 * time.Millisecond)

	if err := app.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: exit error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("PASS: Clean exit")

	fmt.Println("\n✅ ALL TESTS PASSED")
}
