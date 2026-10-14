package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/starframe-dev/abyss/internal/browser"
)

func main() {
	headless := flag.String("headless-dump", "", "if set to a file path, load --url and write the ANSI render to that path instead of running the TUI")
	url := flag.String("url", "", "URL to load in headless-dump mode")
	width := flag.Int("width", 200, "viewport width for headless-dump mode")
	flag.Parse()

	if *headless != "" {
		if *url == "" {
			fmt.Fprintln(os.Stderr, "--headless-dump requires --url")
			os.Exit(2)
		}
		browser.HeadlessDump(*url, *width, *headless)
		return
	}

	m := browser.New()
	p := tea.NewProgram(
		&m,
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
