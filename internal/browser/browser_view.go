package browser

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
	"github.com/rivo/uniseg"
)

// View implements tea.Model.
func (m Model) View() string {
	var b strings.Builder

	// Tab bar
	tabStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Background(lipgloss.Color("#313244")).
		Foreground(lipgloss.Color("#cdd6f4"))

	activeTabStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Background(lipgloss.Color("#45475a")).
		Foreground(lipgloss.Color("#f5c2e7")).
		Bold(true)

	for i, t := range m.Tabs {
		label := safeTerminalText(t.URL)
		if label == "" {
			label = "New Tab"
		} else {
			// Shorten URL for display
			label = strings.TrimPrefix(label, "https://")
			label = strings.TrimPrefix(label, "http://")
			if runewidth.StringWidth(label) > 20 {
				label = truncateTerminalText(label, 19) + "…"
			}
		}
		if i == m.ActiveTab {
			b.WriteString(activeTabStyle.Render(label))
		} else {
			b.WriteString(tabStyle.Render(label))
		}
	}
	b.WriteString("\n")

	// Header bar
	headerStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#1e1e2e")).
		Foreground(lipgloss.Color("#cdd6f4")).
		Padding(0, 1)

	if m.Mode == ModeAddress {
		addressInput := m.Address
		addressInput.SetValue(safeTerminalText(addressInput.Value()))
		b.WriteString(headerStyle.Render("Abyss " + addressInput.View()))
	} else {
		addr := m.Address.Value()
		if addr == "" {
			addr = "Abyss"
		}
		// Show zoom level in header
		if m.ZoomFactor() != 1.0 {
			addr = fmt.Sprintf("%s [%.0f%%]", addr, m.ZoomFactor()*100)
		}
		b.WriteString(headerStyle.Render(safeTerminalText(addr)))
	}
	b.WriteString("\n")

	// Content area
	switch m.Mode {
	case ModeLoading:
		b.WriteString("\n" + center("Loading...", m.Width))
	case ModeError:
		errMsg := safeTerminalText(fmt.Sprintf("Error: %v", m.Err()))
		b.WriteString("\n" + center(errMsg, m.Width))
		b.WriteString("\n" + center("Press Ctrl+L to edit URL or Q to quit", m.Width))
	case ModeAddress:
		b.WriteString("\n" + center(safeTerminalText(m.Address.Value()), m.Width))
	case ModeView:
		lines := m.ANSILines()
		yoff := m.Viewport.YOffset
		h := m.Viewport.Height
		if h <= 0 {
			h = 24
		}
		if yoff >= len(lines) {
			yoff = len(lines) - h
			if yoff < 0 {
				yoff = 0
			}
		}
		end := yoff + h
		if end > len(lines) {
			end = len(lines)
		}
		visible := lines[yoff:end]

		// Apply horizontal scroll: clip each line by offsetX
		w := m.Width
		if w <= 0 {
			w = 80
		}
		offsetX := m.OffsetX()
		// Copy before editing: visible aliases tab.ANSILines, and writing into
		// it in place destroyed the original rows permanently, so scrolling
		// right and back left the page blank.
		out := make([]string, len(visible))
		for i, line := range visible {
			if offsetX == 0 && visibleWidth(line) <= w {
				out[i] = line
				continue
			}
			out[i] = clipANSI(line, offsetX, w)
		}
		result := strings.Join(out, "\n")
		b.WriteString(result)
	}

	return b.String()
}

func safeTerminalText(text string) string {
	return strings.Map(func(r rune) rune {
		if !unicode.IsPrint(r) && r != 0x200c && r != 0x200d {
			return '�'
		}
		return r
	}, text)
}

func truncateTerminalText(text string, maxWidth int) string {
	var out strings.Builder
	used := 0
	graphemes := uniseg.NewGraphemes(text)
	for graphemes.Next() {
		width := graphemes.Width()
		if width > 0 && used+width > maxWidth {
			break
		}
		out.WriteString(graphemes.Str())
		used += max(0, width)
	}
	return out.String()
}

func center(text string, width int) string {
	textWidth := runewidth.StringWidth(text)
	if textWidth >= width {
		return text
	}
	pad := (width - textWidth) / 2
	return strings.Repeat(" ", pad) + text
}
