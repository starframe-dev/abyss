// Package browser implements the TUI using Bubble Tea.
package browser

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
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
	ID          uint64
	RequestID   uint64
	cancelLoad  context.CancelFunc
	URL         string
	Mode        Mode
	ANSILines   []string
	ASCIILines  []string
	ExternalCSS []string
	Err         error
	ZoomFactor  float64
	OffsetX     int
	OffsetY     int
}

// Model is the main Bubble Tea model for the browser.
type Model struct {
	Mode              Mode
	Address           textinput.Model
	Viewport          viewport.Model
	Tabs              []Tab
	ActiveTab         int
	Width             int
	Height            int
	nextTabID         uint64
	loadedPageContent string
}

// PageLoadedMsg is sent when a page finishes loading.
type PageLoadedMsg struct {
	TabID       uint64
	RequestID   uint64
	Lines       []string
	ANSI        []string
	URL         string
	ExternalCSS []string
}

// PageRenderedMsg updates rendered content without resetting navigation metadata.
type PageRenderedMsg struct {
	TabID     uint64
	RequestID uint64
	Lines     []string
	ANSI      []string
}

// PageErrorMsg is sent when loading fails.
type PageErrorMsg struct {
	TabID     uint64
	RequestID uint64
	Err       error
}

type tickMsg struct{}

// TickCmd returns a command that waits for the next tick.
func TickCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

var sessionFile = filepath.Join(os.Getenv("HOME"), ".abyss", "session.json")

type sessionData struct {
	URLs []string `json:"urls"`
}

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
	data, err := json.Marshal(sessionData{URLs: urls})
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(sessionFile), 0755); err != nil {
		return
	}
	_ = os.WriteFile(sessionFile, data, 0644)
}

func loadSession() []string {
	data, err := os.ReadFile(sessionFile)
	if err != nil {
		return nil
	}
	var session sessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return nil
	}
	return session.URLs
}

// New creates a new browser model.
func New() Model {
	ti := textinput.New()
	ti.Placeholder = "Enter URL..."
	ti.Focus()
	ti.Width = 50
	m := Model{
		Mode: ModeAddress, Address: ti, Viewport: viewport.New(80, 24),
		Tabs: []Tab{{Mode: ModeAddress}}, ActiveTab: 0, Width: 80, Height: 24,
	}
	if urls := loadSession(); len(urls) > 0 {
		m.Tabs = nil
		for _, url := range urls {
			m.Tabs = append(m.Tabs, Tab{URL: url, Mode: ModeAddress})
		}
		m.ActiveTab = len(m.Tabs) - 1
	}
	m.ensureTabIDs()
	return m
}

func (m *Model) ensureTabIDs() {
	for i := range m.Tabs {
		if m.Tabs[i].ID == 0 {
			m.nextTabID++
			m.Tabs[i].ID = m.nextTabID
		} else if m.Tabs[i].ID > m.nextTabID {
			m.nextTabID = m.Tabs[i].ID
		}
	}
}

func (m *Model) cancelLoads() {
	for i := range m.Tabs {
		if m.Tabs[i].cancelLoad != nil {
			m.Tabs[i].cancelLoad()
			m.Tabs[i].cancelLoad = nil
		}
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, TickCmd())
}

func (m *Model) tabForResult(tabID, requestID uint64) *Tab {
	if tabID == 0 {
		return nil
	}
	for i := range m.Tabs {
		tab := &m.Tabs[i]
		if tab.ID == tabID && tab.RequestID == requestID {
			return tab
		}
	}
	return nil
}

func (m *Model) active() *Tab {
	if m.ActiveTab < 0 || m.ActiveTab >= len(m.Tabs) {
		return nil
	}
	return &m.Tabs[m.ActiveTab]
}

// GetActiveTab returns a pointer to the current active tab.
func (m *Model) GetActiveTab() *Tab { return m.active() }

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
	t.OffsetY = m.Viewport.YOffset
}

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
	m.Viewport.YOffset = min(max(0, t.OffsetY), max(0, len(t.ASCIILines)-m.Viewport.Height))
}

func (m *Model) restoreActiveTab() { m.RestoreActiveTab() }
