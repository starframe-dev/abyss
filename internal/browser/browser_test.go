package browser

import (
    "encoding/json"
    "os"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
    m := New()
    assert.Equal(t, ModeAddress, m.Mode)
    assert.NotNil(t, m.Address)
    assert.NotNil(t, m.Viewport)
    assert.NotNil(t, m.Tabs)
    assert.Equal(t, 0, m.ActiveTab)
    assert.Greater(t, len(m.Tabs), 0)
}

func TestActive(t *testing.T) {
    m := New()
    // Test when there's an active tab
    tab := m.active()
    assert.NotNil(t, tab)
    assert.Equal(t, 0, m.ActiveTab)

    // Test when active tab is out of bounds
    m.ActiveTab = 100
    tab = m.active()
    assert.Nil(t, tab)

    m.ActiveTab = -1
    tab = m.active()
    assert.Nil(t, tab)
}

func TestSaveActiveTab(t *testing.T) {
    m := New()
    m.Address.SetValue("https://example.com")
    m.Mode = ModeView
    m.Tabs[0].ANSILines = []string{"line1"}
    m.Tabs[0].ASCIILines = []string{"asciiline1"}
    m.Tabs[0].ZoomFactor = 2.0
    m.Tabs[0].OffsetX = 5

    m.saveActiveTab()
    tab := m.active()
    assert.Equal(t, "https://example.com", tab.URL)
    assert.Equal(t, ModeView, tab.Mode)
    assert.Equal(t, 2.0, tab.ZoomFactor)
    assert.Equal(t, 5, tab.OffsetX)
}

func TestRestoreActiveTab(t *testing.T) {
    m := New()
    tab := Tab{
        URL:        "https://example.com",
        Mode:       ModeView,
        ANSILines:  []string{"ansi_line1", "ansi_line2"},
        ASCIILines: []string{"asciiline1\nasciiline2"},
        ZoomFactor: 2.0,
        OffsetX:    5,
    }
    m.Tabs = []Tab{tab}
    m.ActiveTab = 0

    m.restoreActiveTab()
    assert.Equal(t, "https://example.com", m.Address.Value())
    assert.Equal(t, ModeView, m.Mode)
    assert.Contains(t, m.Viewport.View(), "asciiline1")
}

func TestInit(t *testing.T) {
    m := New()
    cmd := m.Init()
    assert.NotNil(t, cmd)
}

func TestSaveSession_EmptyTabs(t *testing.T) {
    m := New()
    saveSession(m.Tabs)
    // Should not panic with empty tabs
}

func TestSaveSession_NoURLs(t *testing.T) {
    m := New()
    m.Tabs = []Tab{{}, {}}
    m.Tabs[0].URL = ""
    m.Tabs[1].URL = ""
    saveSession(m.Tabs)
    // Should not panic with tabs that have no URLs
}

func TestSaveSession_ValidURLs(t *testing.T) {
    m := New()
    m.Tabs = []Tab{{}, {}}
    m.Tabs[0].URL = "https://example.com"
    m.Tabs[1].URL = "https://test.com"
    saveSession(m.Tabs)
    // Should write to session file
}

func TestSaveSession_SingleTabWithURL(t *testing.T) {
    m := New()
    m.Tabs[0].URL = "https://single.com"
    saveSession(m.Tabs)
    // Should write to session file
}

func TestSaveSession_NonEmptyTabs(t *testing.T) {
    m := New()
    m.Tabs = append(m.Tabs, Tab{URL: "https://test.com"})
    saveSession(m.Tabs)
    // Should write to session file
}

func TestLoadSession_NoFile(t *testing.T) {
    // Remove any existing session file so we test the not-found path.
    os.Remove(sessionFile)
    urls := loadSession()
    assert.Nil(t, urls)
}

func TestLoadSession_EmptyFile(t *testing.T) {
    // Create a session file with empty content
    os.WriteFile(sessionFile, []byte(""), 0644)
    defer os.Remove(sessionFile)

    urls := loadSession()
    assert.Nil(t, urls)
}

func TestLoadSession_InvalidJSON(t *testing.T) {
    // Create a session file with invalid JSON
    os.WriteFile(sessionFile, []byte("invalid json"), 0644)
    defer os.Remove(sessionFile)

    urls := loadSession()
    assert.Nil(t, urls)
}

func TestLoadSession_ValidSession(t *testing.T) {
    data := sessionData{URLs: []string{"https://example.com", "https://test.com"}}
    b, _ := json.Marshal(data)
    os.WriteFile(sessionFile, b, 0644)
    defer os.Remove(sessionFile)

    urls := loadSession()
    assert.Equal(t, []string{"https://example.com", "https://test.com"}, urls)
}

func TestLoadSession_SingleURL(t *testing.T) {
    data := sessionData{URLs: []string{"https://example.com"}}
    b, _ := json.Marshal(data)
    os.WriteFile(sessionFile, b, 0644)
    defer os.Remove(sessionFile)

    urls := loadSession()
    assert.Equal(t, []string{"https://example.com"}, urls)
}

func TestANSILines(t *testing.T) {
    m := New()
    lines := m.ANSILines()
    assert.Nil(t, lines)

    m.Tabs = []Tab{{ANSILines: []string{"<ansi>line</ansi>"}}}
    m.ActiveTab = 0
    lines = m.ANSILines()
    assert.Equal(t, []string{"<ansi>line</ansi>"}, lines)
}

func TestASCIILines(t *testing.T) {
    m := New()
    lines := m.ASCIILines()
    assert.Nil(t, lines)

    m.Tabs = []Tab{{ASCIILines: []string{"line1\nline2"}}}
    m.ActiveTab = 0
    lines = m.ASCIILines()
    assert.Equal(t, []string{"line1\nline2"}, lines)
}

func TestZoomFactor(t *testing.T) {
    m := New()
    // No tabs means ZoomFactor returns the default 1.0.
    m.Tabs = nil
    factor := m.ZoomFactor()
    assert.Equal(t, 1.0, factor)

    m.Tabs = []Tab{{ZoomFactor: 2.5}}
    m.ActiveTab = 0
    factor = m.ZoomFactor()
    assert.Equal(t, 2.5, factor)
}

func TestOffsetX(t *testing.T) {
    m := New()
    offset := m.OffsetX()
    assert.Equal(t, 0, offset)

    m.Tabs = []Tab{{OffsetX: 10}}
    m.ActiveTab = 0
    offset = m.OffsetX()
    assert.Equal(t, 10, offset)
}

func TestErr(t *testing.T) {
    m := New()
    err := m.Err()
    assert.Nil(t, err)

    // Replace the default tab with one that has an error.
    m.Tabs = []Tab{{Err: assert.AnError}}
    m.ActiveTab = 0
    err = m.Err()
    assert.NotNil(t, err)
}

func TestActive_OutOfBoundsNegative(t *testing.T) {
    m := New()
    m.ActiveTab = -1
    tab := m.active()
    assert.Nil(t, tab)
}

func TestActive_OutOfBoundsPositive(t *testing.T) {
    m := New()
    m.ActiveTab = 999
    tab := m.active()
    assert.Nil(t, tab)
}

func TestActive_ValidIndex(t *testing.T) {
    m := New()
    m.ActiveTab = 0
    tab := m.active()
    assert.NotNil(t, tab)
}

func TestSaveActiveTab_NoTabs(t *testing.T) {
    m := New()
    m.Tabs = nil
    m.ActiveTab = 0
    tab := m.active()
    assert.Nil(t, tab)
    m.saveActiveTab()
    // Should not panic when there are no tabs.
}

func TestRestoreActiveTab_NoTabs(t *testing.T) {
    m := New()
    m.Tabs = nil
    m.ActiveTab = 0
    tab := m.active()
    assert.Nil(t, tab)
    m.restoreActiveTab()
    // Should not panic when there are no tabs.
}

func TestActive_NilTabs(t *testing.T) {
    m := New()
    m.Tabs = nil
    m.ActiveTab = 0
    tab := m.active()
    assert.Nil(t, tab)
}

func TestNew_MultipleTabs(t *testing.T) {
    m := New()
    assert.GreaterOrEqual(t, len(m.Tabs), 1)
    assert.NotNil(t, m.Tabs[0])
}

func TestSaveActiveTab_State(t *testing.T) {
    m := New()
    m.Address.SetValue("https://test.com")
    m.Mode = ModeError
    m.Tabs[0].ANSILines = []string{"<ansi>error</ansi>"}
    m.Tabs[0].ASCIILines = []string{"error"}
    m.Tabs[0].ZoomFactor = 0.5
    m.Tabs[0].OffsetX = -3

    m.saveActiveTab()
    tab := m.active()
    assert.Equal(t, ModeError, tab.Mode)
}

func TestRestoreActiveTab_ViewContent(t *testing.T) {
    m := New()
    tab := Tab{
        URL:        "https://example.com",
        ASCIILines: []string{"page content line 1", "page content line 2"},
    }
    m.Tabs = []Tab{tab}
    m.ActiveTab = 0

    m.restoreActiveTab()
    assert.Contains(t, m.Viewport.View(), "page content line 1")
}

func TestInit_ReturnsCmd(t *testing.T) {
    m := New()
    cmd := m.Init()
    // Init should always return a command
    assert.NotNil(t, cmd)
}

func TestLoadSession_ReturnsNilOnReadError(t *testing.T) {
    // When ReadFile returns error, loadSession should return nil.
    os.Remove(sessionFile)
    urls := loadSession()
    assert.Nil(t, urls)
}

func TestLoadSession_ReturnsNilOnUnmarshalError(t *testing.T) {
    // When Unmarshal returns error, loadSession should return nil
    os.WriteFile(sessionFile, []byte("not json"), 0644)
    defer os.Remove(sessionFile)

    urls := loadSession()
    assert.Nil(t, urls)
}

func TestTickCmd(t *testing.T) {
    cmd := TickCmd()
    assert.NotNil(t, cmd)
}

func TestNew_ReturnsModel(t *testing.T) {
    m := New()
    assert.NotNil(t, m)
    assert.Equal(t, ModeAddress, m.Mode)
}

func TestActive_ReturnsNilWhenNoTabs(t *testing.T) {
    m := New()
    m.Tabs = nil
    m.ActiveTab = 0
    tab := m.active()
    assert.Nil(t, tab)
}

func TestSaveActiveTab_DoesNotModifyTabWhenNil(t *testing.T) {
    m := New()
    m.Tabs = nil
    m.ActiveTab = 0
    m.saveActiveTab()
    assert.Nil(t, m.active())
}
