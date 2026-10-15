package integration

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/starframe-dev/abyss/internal/browser"
	"github.com/stretchr/testify/require"
)

func updateBrowser(t *testing.T, m *browser.Model, msg tea.Msg) tea.Cmd {
	t.Helper()
	next, cmd := m.Update(msg)
	updated, ok := next.(browser.Model)
	require.True(t, ok)
	*m = updated
	return cmd
}

func typeURL(t *testing.T, m *browser.Model, value string) {
	t.Helper()
	for _, r := range value {
		updateBrowser(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
}

func TestBrowserNavigationAndStaleResults(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<html><body><h1>first integration page</h1></body></html>"))
	}))
	defer server.Close()

	m := browser.New()
	updateBrowser(t, &m, tea.WindowSizeMsg{Width: 80, Height: 24})
	typeURL(t, &m, server.URL)
	cmd := updateBrowser(t, &m, tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)
	require.Equal(t, browser.ModeLoading, m.Mode)
	firstTab := *m.GetActiveTab()
	require.Equal(t, server.URL, firstTab.URL)
	firstID, firstRequest := firstTab.ID, firstTab.RequestID

	// Start a newer navigation in the same tab. A late completion for the old
	// request must be ignored, even though its tab ID still matches.
	updateBrowser(t, &m, tea.KeyMsg{Type: tea.KeyCtrlL})
	m.Address.SetValue(server.URL + "/new")
	updateBrowser(t, &m, tea.KeyMsg{Type: tea.KeyEnter})
	second := *m.GetActiveTab()
	require.Greater(t, second.RequestID, firstRequest)
	updateBrowser(t, &m, browser.PageLoadedMsg{TabID: firstID, RequestID: firstRequest, URL: server.URL, Lines: []string{"stale page"}})
	require.Equal(t, browser.ModeLoading, m.Mode)
	require.NotContains(t, strings.Join(m.ASCIILines(), "\n"), "stale page")

	updateBrowser(t, &m, browser.PageLoadedMsg{TabID: second.ID, RequestID: second.RequestID, URL: server.URL + "/new", Lines: []string{"new page"}, ANSI: []string{"new page"}})
	require.Equal(t, browser.ModeView, m.Mode)
	require.Contains(t, strings.Join(m.ASCIILines(), "\n"), "new page")

	// New-tab creation and cyclic switching preserve each tab's navigation data.
	updateBrowser(t, &m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	require.Len(t, m.Tabs, 2)
	require.Equal(t, 1, m.ActiveTab)
	require.Equal(t, browser.ModeAddress, m.Mode)
	typeURL(t, &m, server.URL)
	updateBrowser(t, &m, tea.KeyMsg{Type: tea.KeyEnter})
	third := *m.GetActiveTab()
	updateBrowser(t, &m, browser.PageLoadedMsg{TabID: third.ID, RequestID: third.RequestID, URL: server.URL, Lines: []string{"tab two"}, ANSI: []string{"tab two"}})
	updateBrowser(t, &m, tea.KeyMsg{Type: tea.KeyTab})
	require.Equal(t, 0, m.ActiveTab)
	require.Contains(t, strings.Join(m.ASCIILines(), "\n"), "new page")
	updateBrowser(t, &m, tea.KeyMsg{Type: tea.KeyShiftTab})
	require.Equal(t, 1, m.ActiveTab)
	require.Contains(t, strings.Join(m.ASCIILines(), "\n"), "tab two")

	// Reload requests a render command, and load errors remain scoped to their
	// matching navigation request.
	updateBrowser(t, &m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	updateBrowser(t, &m, browser.PageErrorMsg{TabID: third.ID, RequestID: third.RequestID, Err: errors.New("load failed")})
	require.Equal(t, browser.ModeError, m.Mode)
	require.Error(t, m.GetActiveTab().Err)
	updateBrowser(t, &m, tea.KeyMsg{Type: tea.KeyCtrlL})
	require.Equal(t, browser.ModeAddress, m.Mode)
}
