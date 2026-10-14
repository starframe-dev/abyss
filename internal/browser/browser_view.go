package browser

import (
    "fmt"
    "strings"

    "github.com/charmbracelet/lipgloss"
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
        label := t.URL
        if label == "" {
            label = "New Tab"
        } else {
            // Shorten URL for display
            label = strings.TrimPrefix(label, "https://")
            label = strings.TrimPrefix(label, "http://")
            if len(label) > 20 {
                label = label[:20] + "…"
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
        b.WriteString(headerStyle.Render("Abyss " + m.Address.View()))
    } else {
        addr := m.Address.Value()
        if addr == "" {
            addr = "Abyss"
        }
        // Show zoom level in header
        if m.ZoomFactor() != 1.0 {
            addr = fmt.Sprintf("%s [%.0f%%]", addr, m.ZoomFactor()*100)
        }
        b.WriteString(headerStyle.Render(addr))
    }
    b.WriteString("\n")

    // Content area
    switch m.Mode {
    case ModeLoading:
        b.WriteString("\n" + center("Loading...", m.Width))
    case ModeError:
        errMsg := fmt.Sprintf("Error: %v", m.Err())
        b.WriteString("\n" + center(errMsg, m.Width))
        b.WriteString("\n" + center("Press Ctrl+L to edit URL or Q to quit", m.Width))
    case ModeAddress:
        b.WriteString("\n" + center(m.Address.Value(), m.Width))
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

func center(text string, width int) string {
    if len(text) >= width {
        return text
    }
    pad := (width - len(text)) / 2
    return strings.Repeat(" ", pad) + text
}
