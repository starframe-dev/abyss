package browser

import (
	"context"
	"math"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Update implements tea.Model and applies one message to the browser state.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.ensureTabIDs()
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.QuitMsg:
		m.cancelLoads()
		closeJSCContext()
		return m, tea.Quit

	case tea.WindowSizeMsg:
		previousWidth := m.Width
		m.Width = msg.Width
		m.Height = msg.Height
		m.Address.Width = msg.Width - 4
		m.Viewport.Width = msg.Width
		m.Viewport.Height = msg.Height - 4
		if m.Mode == ModeView {
			m.Viewport.SetContent(strings.Join(m.ASCIILines(), "\n"))
			if previousWidth != msg.Width {
				if cmd := m.activeRenderCmd(); cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			saveSession(m.Tabs)
			m.cancelLoads()
			closeJSCContext()
			return m, tea.Quit
		}
		if (msg.String() == "tab" || msg.String() == "shift+tab") && len(m.Tabs) > 1 {
			m.saveActiveTab()
			if msg.String() == "tab" {
				m.ActiveTab = (m.ActiveTab + 1) % len(m.Tabs)
			} else {
				m.ActiveTab = (m.ActiveTab - 1 + len(m.Tabs)) % len(m.Tabs)
			}
			m.restoreActiveTab()
			return m, nil
		}

		switch m.Mode {
		case ModeAddress:
			switch msg.String() {
			case "enter":
				address := m.Address.Value()
				if address == "" {
					return m, nil
				}
				if !strings.HasPrefix(address, "http://") && !strings.HasPrefix(address, "https://") {
					address = "https://" + address
				}
				m.saveActiveTab()
				m.Mode = ModeLoading
				tab := m.active()
				if tab == nil {
					return m, nil
				}
				if tab.cancelLoad != nil {
					tab.cancelLoad()
				}
				loadCtx, cancel := context.WithCancel(context.Background())
				tab.cancelLoad = cancel
				tab.RequestID++
				tab.Mode = ModeLoading
				tab.ZoomFactor = 1
				tab.OffsetX = 0
				tab.URL = address
				m.Address.SetValue(address)
				return m, loadTabPage(loadCtx, address, m.Width, tab.ID, tab.RequestID)
			case "q":
				saveSession(m.Tabs)
				m.cancelLoads()
				closeJSCContext()
				return m, tea.Quit
			}
			var cmd tea.Cmd
			m.Address, cmd = m.Address.Update(msg)
			return m, cmd

		case ModeView:
			switch msg.String() {
			case "ctrl+l":
				m.saveActiveTab()
				m.Mode = ModeAddress
				m.Address.Focus()
				return m, textinput.Blink
			case "r":
				return m, m.activeRenderCmd()
			case "+", "=":
				if tab := m.active(); tab != nil {
					tab.ZoomFactor = math.Min(3, m.ZoomFactor()+0.5)
				}
				return m, m.activeRenderCmd()
			case "-", "_":
				if tab := m.active(); tab != nil {
					tab.ZoomFactor = math.Max(0.5, m.ZoomFactor()-0.5)
				}
				return m, m.activeRenderCmd()
			case "left":
				if tab := m.active(); tab != nil {
					tab.OffsetX = max(0, m.OffsetX()-10)
				}
			case "right":
				if tab := m.active(); tab != nil {
					tab.OffsetX += 10
				}
			case "t":
				m.saveActiveTab()
				m.Tabs = append(m.Tabs, Tab{Mode: ModeAddress})
				m.ActiveTab = len(m.Tabs) - 1
				m.Mode = ModeAddress
				m.Address.SetValue("")
				m.Address.Focus()
				m.Viewport.SetContent("")
				return m, textinput.Blink
			case "w":
				if len(m.Tabs) <= 1 {
					break
				}
				if closing := m.active(); closing != nil {
					if closing.cancelLoad != nil {
						closing.cancelLoad()
					}
					closeJSCContextForTab(closing.ID)
				}
				m.Tabs = append(m.Tabs[:m.ActiveTab], m.Tabs[m.ActiveTab+1:]...)
				if m.ActiveTab >= len(m.Tabs) {
					m.ActiveTab = len(m.Tabs) - 1
				}
				m.restoreActiveTab()
				return m, nil
			case "tab":
				if len(m.Tabs) > 0 {
					m.saveActiveTab()
					m.ActiveTab = (m.ActiveTab + 1) % len(m.Tabs)
					m.restoreActiveTab()
				}
				return m, nil
			case "shift+tab":
				if len(m.Tabs) > 0 {
					m.saveActiveTab()
					m.ActiveTab = (m.ActiveTab - 1 + len(m.Tabs)) % len(m.Tabs)
					m.restoreActiveTab()
				}
				return m, nil
			case "q":
				saveSession(m.Tabs)
				m.cancelLoads()
				closeJSCContext()
				return m, tea.Quit
			default:
				var cmd tea.Cmd
				m.Viewport, cmd = m.Viewport.Update(msg)
				return m, cmd
			}

		case ModeError:
			switch msg.String() {
			case "ctrl+l":
				m.Mode = ModeAddress
				m.Address.Focus()
				return m, textinput.Blink
			case "q":
				saveSession(m.Tabs)
				m.cancelLoads()
				closeJSCContext()
				return m, tea.Quit
			}
		}

	case tickMsg:
		if m.Mode == ModeLoading {
			cmds = append(cmds, TickCmd())
			break
		}
		if tab := m.active(); tab != nil && processJSTimers(tab.ID, tab.URL) {
			cmds = append(cmds, m.activeRenderCmd())
		}
		cmds = append(cmds, TickCmd())

	case PageRenderedMsg:
		tab := m.tabForResult(msg.TabID, msg.RequestID)
		if tab != nil {
			tab.ANSILines, tab.ASCIILines = msg.ANSI, msg.Lines
			if m.active() == tab {
				m.Viewport.SetContent(strings.Join(msg.Lines, "\n"))
			}
		}

	case PageLoadedMsg:
		tab := m.tabForResult(msg.TabID, msg.RequestID)
		if tab == nil {
			break
		}
		if tab.cancelLoad != nil {
			tab.cancelLoad()
		}
		tab.cancelLoad = nil
		tab.URL = msg.URL
		tab.ANSILines = msg.ANSI
		tab.ASCIILines = msg.Lines
		tab.ExternalCSS = msg.ExternalCSS
		tab.OffsetY = 0
		tab.Mode = ModeView
		if m.active() == tab {
			m.loadedPageContent = strings.Join(msg.Lines, "\n")
			m.Mode = ModeView
			m.Viewport.SetContent(strings.Join(msg.Lines, "\n"))
			m.Viewport.GotoTop()
			cmds = append(cmds, TickCmd())
		}

	case PageErrorMsg:
		tab := m.tabForResult(msg.TabID, msg.RequestID)
		if tab == nil {
			break
		}
		if tab.cancelLoad != nil {
			tab.cancelLoad()
		}
		tab.cancelLoad = nil
		tab.Err = msg.Err
		tab.Mode = ModeError
		if m.active() == tab {
			m.Mode = ModeError
		}
	}

	return m, tea.Batch(cmds...)
}
