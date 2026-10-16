package browser

import tea "github.com/charmbracelet/bubbletea"

// Helper methods to access tab state through the active tab.
func (m Model) ANSILines() []string {
	if t := m.active(); t != nil {
		return t.ANSILines
	}
	return nil
}

func (m Model) ASCIILines() []string {
	if t := m.active(); t != nil {
		return t.ASCIILines
	}
	return nil
}

func (m Model) ZoomFactor() float64 {
	if t := m.active(); t != nil {
		return t.ZoomFactor
	}
	return 1.0
}

func (m Model) activeExternalCSS() []string {
	if tab := m.active(); tab != nil {
		return tab.ExternalCSS
	}
	return nil
}

func (m Model) OffsetX() int {
	if t := m.active(); t != nil {
		return t.OffsetX
	}
	return 0
}

func (m Model) Err() error {
	if t := m.active(); t != nil {
		return t.Err
	}
	return nil
}

func (m Model) activeRenderCmd() tea.Cmd {
	tab := m.active()
	if tab == nil {
		return nil
	}
	return reRenderCmd(m.Width, m.ZoomFactor(), tab.ID, tab.RequestID, tab.URL, tab.ExternalCSS)
}
