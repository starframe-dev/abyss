// Package browser implements the TUI using Bubble Tea.
package browser

import (
    "encoding/json"
    "math"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/charmbracelet/bubbles/textinput"
    "github.com/charmbracelet/bubbles/viewport"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/starframe/abyss/internal/js"
)

// Mode represents the current UI mode.
type Mode int

const (
    ModeAddress Mode = iota
    ModeLoading
    ModeView
    ModeError
)

// Tab stores the state of a single browser tab.
type Tab struct {
    URL           string
    Mode          Mode
    ANSILines     []string
    ASCIILines    []string
    Err           error
    ZoomFactor    float64
    OffsetX       int
}

// Model is the main Bubble Tea model for the browser.
type Model struct {
    Mode               Mode
    Address            textinput.Model
    Viewport           viewport.Model
    Tabs               []Tab
    ActiveTab          int
    Width              int
    Height             int
    loadedPageContent  string
}

// PageLoadedMsg is sent when a page finishes loading.
type PageLoadedMsg struct {
    Lines []string
    ANSI  []string
    URL   string
}

// PageErrorMsg is sent when loading fails.
type PageErrorMsg struct {
    Err error
}

// tickMsg is sent periodically to check for timer-based DOM updates.
type tickMsg struct{}

// TickCmd returns a command that waits for the next tick.
func TickCmd() tea.Cmd {
    return tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg {
        return tickMsg{}
    })
}

// sessionFile is the path to the session file for tab restoration.
var sessionFile = filepath.Join(os.Getenv("HOME"), ".abyss", "session.json")

// sessionData stores URLs for tab restoration.
type sessionData struct {
    URLs []string `json:"urls"`
}

// saveSession writes the current tab URLs to the session file.
func saveSession(tabs []Tab) {
    if len(tabs) == 0 {
        return
    }
    var urls []string
    for _, t := range tabs {
        if t.URL != "" {
            urls = append(urls, t.URL)
        }
    }
    if len(urls) == 0 {
        return
    }
    data := sessionData{URLs: urls}
    dir := filepath.Dir(sessionFile)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return
    }
    b, err := json.Marshal(data)
    if err != nil {
        return
    }
    _ = os.WriteFile(sessionFile, b, 0644)
}

// loadSession reads the session file and returns the saved URLs.
func loadSession() []string {
    b, err := os.ReadFile(sessionFile)
    if err != nil {
        return nil
    }
    var data sessionData
    if err := json.Unmarshal(b, &data); err != nil {
        return nil
    }
    return data.URLs
}

// New creates a new browser model.
func New() Model {
    ti := textinput.New()
    ti.Placeholder = "Enter URL..."
    ti.Focus()
    ti.Width = 50

    vp := viewport.New(80, 24)

    m := Model{
        Mode:      ModeAddress,
        Address:   ti,
        Viewport:  vp,
        Tabs:      []Tab{{Mode: ModeAddress}},
        ActiveTab: 0,
        Width:     80,
        Height:    24,
    }

    // Restore session: show saved URLs in tab bar, don't fill address bar
    urls := loadSession()
    if len(urls) > 0 {
        m.Tabs = nil
        for _, u := range urls {
            m.Tabs = append(m.Tabs, Tab{URL: u, Mode: ModeAddress})
        }
        m.ActiveTab = len(m.Tabs) - 1
    }

    return m
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
    return tea.Batch(textinput.Blink, TickCmd())
}

// active returns a pointer to the current active tab.
func (m *Model) active() *Tab {
    if m.ActiveTab < 0 || m.ActiveTab >= len(m.Tabs) {
        return nil
    }
    return &m.Tabs[m.ActiveTab]
}

// GetActiveTab returns a pointer to the current active tab (public API).
func (m *Model) GetActiveTab() *Tab {
    return m.active()
}

// SaveActiveTab copies the current view state into the active tab.
func (m *Model) SaveActiveTab() {
    t := m.active()
    if t == nil {
        return
    }
    t.URL = m.Address.Value()
    t.Mode = m.Mode
    t.ANSILines = m.ANSILines()
    t.ASCIILines = m.ASCIILines()
    t.ZoomFactor = m.ZoomFactor()
    t.OffsetX = m.OffsetX()
}

// saveActiveTab is a thin alias for SaveActiveTab used by internal callers.
func (m *Model) saveActiveTab() { m.SaveActiveTab() }

// RestoreActiveTab restores view state from the active tab.
func (m *Model) RestoreActiveTab() {
    t := m.active()
    if t == nil {
        return
    }
    m.Mode = t.Mode
    m.Address.SetValue(t.URL)
    m.Viewport.SetContent(strings.Join(t.ASCIILines, "\n"))
    m.Viewport.GotoTop()
}

// restoreActiveTab is a thin alias for RestoreActiveTab used by internal callers.
func (m *Model) restoreActiveTab() { m.RestoreActiveTab() }

// Update implements tea.Model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd

    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.Width = msg.Width
        m.Height = msg.Height
        m.Address.Width = msg.Width - 4
        m.Viewport.Width = msg.Width
        m.Viewport.Height = msg.Height - 4
        if m.Mode == ModeView {
            m.Viewport.SetContent(strings.Join(m.ASCIILines(), "\n"))
        }

    case tea.KeyMsg:
        if msg.String() == "ctrl+c" {
            saveSession(m.Tabs)
            js.CloseCurrentContext()
            return m, tea.Quit
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
                if tab := m.active(); tab != nil {
                    tab.Mode = ModeLoading
                    tab.ZoomFactor = 1.0
                    tab.OffsetX = 0
                    tab.URL = address
                }
                m.Address.SetValue(address)
                return m, loadPage(address, m.Width)
            case "q":
                saveSession(m.Tabs)
                js.CloseCurrentContext()
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
                return m, reRenderCmd(m.Width, m.ZoomFactor())
            case "+", "=":
                if tab := m.active(); tab != nil {
                    tab.ZoomFactor = math.Min(3.0, m.ZoomFactor()+0.5)
                }
                return m, reRenderCmd(m.Width, m.ZoomFactor())
            case "-", "_":
                if tab := m.active(); tab != nil {
                    tab.ZoomFactor = math.Max(0.5, m.ZoomFactor()-0.5)
                }
                return m, reRenderCmd(m.Width, m.ZoomFactor())
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
                js.CloseCurrentContext()
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
                js.CloseCurrentContext()
                return m, tea.Quit
            }
        }

    case tickMsg:
        if m.Mode == ModeLoading {
            cmds = append(cmds, TickCmd())
            break
        }
        if context, _ := js.GetCurrentContext(); context != nil {
            js.ProcessTimers()
            if js.IsDirty() {
                cmds = append(cmds, reRenderCmd(m.Width, m.ZoomFactor()))
            }
        }
        cmds = append(cmds, TickCmd())

    case PageLoadedMsg:
        if tab := m.active(); tab != nil {
            tab.URL = msg.URL
            tab.ANSILines = msg.ANSI
            tab.ASCIILines = msg.Lines
            tab.Mode = ModeView
        }
        m.loadedPageContent = strings.Join(msg.Lines, "\n")
        m.Mode = ModeView
        m.Viewport.SetContent(strings.Join(msg.Lines, "\n"))
        m.Viewport.GotoTop()
        cmds = append(cmds, TickCmd())

    case PageErrorMsg:
        if tab := m.active(); tab != nil {
            tab.Err = msg.Err
            tab.Mode = ModeError
        }
        m.Mode = ModeError
    }

    return m, tea.Batch(cmds...)
}

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
