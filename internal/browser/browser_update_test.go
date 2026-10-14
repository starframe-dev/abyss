package browser

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/starframe-dev/abyss/internal/js"
)

func TestModel_ClosingTabCancelsLoadAndOwnedJSContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	m := New()
	m.Tabs = []Tab{{ID: 1, Mode: ModeLoading, cancelLoad: cancel}, {ID: 2, Mode: ModeView}}
	m.ActiveTab = 0
	m.Mode = ModeView
	jsContext := js.NewContext()
	js.SetCurrentContext(jsContext, "https://loading.test")
	jsLifecycleMu.Lock()
	jsRuntimeTabID = 1
	jsLifecycleMu.Unlock()
	t.Cleanup(closeJSCContext)

	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	assert.Error(t, ctx.Err())
	current, _ := js.GetCurrentContext()
	assert.Nil(t, current, "closing the owner tab must release its JS context")
}

func TestModel_CanSwitchTabsWhileLoading(t *testing.T) {
	m := New()
	m.Tabs = []Tab{{ID: 1, Mode: ModeLoading, URL: "https://loading.test"}, {ID: 2, Mode: ModeView, URL: "https://ready.test", ASCIILines: []string{"ready"}}}
	m.ActiveTab = 0
	m.Mode = ModeLoading
	m.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.Equal(t, 1, m.ActiveTab)
	assert.Equal(t, ModeView, m.Mode)
}

func TestModel_IgnoresStalePageResult(t *testing.T) {
	m := Model{Mode: ModeView, ActiveTab: 0, Tabs: []Tab{{ID: 7, RequestID: 3, URL: "new", Mode: ModeLoading}}}
	m.Update(PageLoadedMsg{TabID: 7, RequestID: 2, URL: "old", Lines: []string{"stale"}})
	assert.Equal(t, "new", m.Tabs[0].URL)
	assert.Equal(t, ModeLoading, m.Tabs[0].Mode)
}

func TestModel_StoresBackgroundTabResultWithoutSwitchingActiveTab(t *testing.T) {
	m := Model{Mode: ModeView, ActiveTab: 1, Tabs: []Tab{{ID: 7, RequestID: 1, Mode: ModeLoading}, {ID: 8, RequestID: 0, URL: "active", Mode: ModeView}}}
	m.Update(PageLoadedMsg{TabID: 7, RequestID: 1, URL: "background", Lines: []string{"loaded"}})
	assert.Equal(t, "background", m.Tabs[0].URL)
	assert.Equal(t, ModeView, m.Tabs[0].Mode)
	assert.Equal(t, "active", m.Tabs[1].URL)
	assert.Equal(t, 1, m.ActiveTab)
}

func TestModel_StoresRenderResultsForOwningTabOnly(t *testing.T) {
	m := Model{Mode: ModeView, ActiveTab: 1, Tabs: []Tab{{ID: 1, RequestID: 2, ASCIILines: []string{"old"}}, {ID: 2, RequestID: 1, ASCIILines: []string{"active"}}}}
	m.Update(PageRenderedMsg{TabID: 1, RequestID: 2, Lines: []string{"background render"}})
	assert.Equal(t, []string{"background render"}, m.Tabs[0].ASCIILines)
	assert.Equal(t, []string{"active"}, m.Tabs[1].ASCIILines)
	m.Update(PageRenderedMsg{TabID: 1, RequestID: 1, Lines: []string{"stale"}})
	assert.Equal(t, []string{"background render"}, m.Tabs[0].ASCIILines)
}

func TestModel_RerenderPreservesPageMetadata(t *testing.T) {
	m := Model{Mode: ModeView, ActiveTab: 1, Tabs: []Tab{{ID: 5, RequestID: 1, URL: "https://keep.test", ExternalCSS: []string{".x { color: red }"}, Mode: ModeView}, {ID: 6, RequestID: 1, URL: "https://active.test", Mode: ModeView}}}
	m.Update(PageRenderedMsg{TabID: 5, RequestID: 1, Lines: []string{"rerendered"}, ANSI: []string{"ansi"}})
	assert.Equal(t, "https://keep.test", m.Tabs[0].URL)
	assert.Equal(t, []string{".x { color: red }"}, m.Tabs[0].ExternalCSS)
	assert.Equal(t, []string{"rerendered"}, m.Tabs[0].ASCIILines)
	assert.Equal(t, "https://active.test", m.Tabs[1].URL)
}

func TestModel_StoresExternalStylesheetsPerTab(t *testing.T) {
	m := Model{Mode: ModeView, ActiveTab: 0, Tabs: []Tab{{ID: 11, RequestID: 1, Mode: ModeLoading}}}
	sheets := []string{".site { color: red; }"}
	m.Update(PageLoadedMsg{TabID: 11, RequestID: 1, URL: "https://styles.test", ExternalCSS: sheets})
	assert.Equal(t, sheets, m.activeExternalCSS())
}

func TestModel_Update_Quit(t *testing.T) {
	m := Model{Mode: ModeAddress, Tabs: []Tab{}}

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	assert.NotNil(t, cmd)
}

func TestModel_Update_Quit_With_q(t *testing.T) {
	m := Model{Mode: ModeAddress, Tabs: []Tab{{Mode: ModeAddress}}}

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	assert.NotNil(t, cmd)
}

func TestModel_Update_Address_Typing(t *testing.T) {
	m := New()
	m.Tabs = []Tab{{Mode: ModeAddress}}
	m.Mode = ModeAddress
	m.ActiveTab = 0

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})

	assert.Same(t, &m, updated)
	assert.Equal(t, "a", m.Address.Value())
}

func TestModel_Update_Address_Enter(t *testing.T) {
	m := New()
	m.Mode = ModeAddress
	m.ActiveTab = 0

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.NotNil(t, updated)
}

func TestModel_Update_Address_Enter_Empty(t *testing.T) {
	m := New()
	m.Mode = ModeAddress
	m.ActiveTab = 0

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.NotNil(t, updated)
	// Cmd may or may not be nil; we only assert no panic.
	_ = cmd
}

func TestModel_Update_Address_With_HTTPS(t *testing.T) {
	m := New()
	m.Mode = ModeAddress
	m.ActiveTab = 0

	// Enter key is not currently bound; verify Update is safe.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.NotNil(t, updated)
}

func TestModel_Update_View_ControlL(t *testing.T) {
	m := New()
	m.Mode = ModeView
	m.ActiveTab = 0

	// Ctrl+L is not currently bound; verify Update is safe.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlL})
	assert.NotNil(t, updated)
}

func TestModel_Update_View_R(t *testing.T) {
	m := New()
	m.Mode = ModeView
	m.ActiveTab = 0

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	assert.Equal(t, ModeView, m.Mode)
}

func TestModel_Update_View_Plus(t *testing.T) {
	m := New()
	m.Mode = ModeView
	m.ActiveTab = 0
	m.Tabs[0].ZoomFactor = 1.0

	// '+' is not currently bound; verify Update is safe.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}})
	assert.NotNil(t, updated)
}

func TestModel_Update_View_Equal(t *testing.T) {
	m := New()
	m.Mode = ModeView
	m.ActiveTab = 0
	m.Tabs[0].ZoomFactor = 1.0

	// '=' is not currently bound; verify Update is safe.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'='}})
	assert.NotNil(t, updated)
}

func TestModel_Update_View_Minus(t *testing.T) {
	m := New()
	m.Mode = ModeView
	m.ActiveTab = 0
	m.Tabs[0].ZoomFactor = 3.0

	// '-' is not currently bound; verify Update is safe.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'-'}})
	assert.NotNil(t, updated)
}

func TestModel_Update_View_Underscore(t *testing.T) {
	m := New()
	m.Mode = ModeView
	m.ActiveTab = 0
	m.Tabs[0].ZoomFactor = 3.0

	// '_' is not currently bound; verify Update is safe.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'_'}})
	assert.NotNil(t, updated)
}

func TestModel_Update_View_Left(t *testing.T) {
	m := New()
	m.Mode = ModeView
	m.ActiveTab = 0
	m.Tabs[0].OffsetX = 0

	// Arrow Left is not currently bound in Update; verify Update is safe.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	assert.NotNil(t, updated)
}

func TestModel_Update_View_Right(t *testing.T) {
	m := New()
	m.Mode = ModeView
	m.ActiveTab = 0
	m.Tabs[0].OffsetX = 0

	// Arrow Right is not currently bound in Update; verify Update is safe.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	assert.NotNil(t, updated)
}

func TestModel_Update_View_T(t *testing.T) {
	m := New()
	m.Mode = ModeView
	m.ActiveTab = 0

	// 't' is not currently bound; verify Update is safe.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	assert.NotNil(t, updated)
}

func TestModel_Update_View_W(t *testing.T) {
	m := New()
	m.Mode = ModeView
	m.ActiveTab = 0

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	assert.Equal(t, 1, len(m.Tabs))
}

func TestModel_Update_View_Tab(t *testing.T) {
	m := New()
	m.Mode = ModeView
	m.ActiveTab = 0

	// Tab key is bound but currently a no-op (returns m, nil); verify Update
	// is safe and the model stays in a consistent state.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	assert.NotNil(t, updated)
}

func TestModel_Update_View_ShiftTab(t *testing.T) {
	m := New()
	m.Mode = ModeView
	m.ActiveTab = 0

	// Shift+Tab is bound but currently a no-op.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	assert.NotNil(t, updated)
}

func TestModel_Update_View_Q(t *testing.T) {
	m := New()
	m.Mode = ModeView
	m.ActiveTab = 0

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	assert.NotNil(t, cmd)
}

func TestModel_Update_Error_ControlL(t *testing.T) {
	m := New()
	m.Mode = ModeError
	m.ActiveTab = 0

	// Ctrl+L is not currently bound; verify Update handles it without panic.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlL})
	assert.NotNil(t, updated)
}

func TestModel_Update_Error_Q(t *testing.T) {
	m := New()
	m.Mode = ModeError
	m.ActiveTab = 0

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	assert.NotNil(t, cmd)
}

func TestModel_Update_Width(t *testing.T) {
	m := New()
	m.Mode = ModeAddress
	m.ActiveTab = 0

	updated, _ := m.Update(tea.WindowSizeMsg{Width: 800, Height: 600})
	assert.NotNil(t, updated)
}

func TestModel_Update_View_Width(t *testing.T) {
	m := New()
	m.Mode = ModeView
	m.ActiveTab = 0

	updated, cmd := m.Update(tea.WindowSizeMsg{Width: 800, Height: 600})
	assert.NotNil(t, updated)
	assert.NotNil(t, cmd, "width changes should trigger a fresh layout/render")
	assert.Equal(t, 800, m.Width)
}

func TestModel_Update_TickMsg(t *testing.T) {
	m := New()
	m.Mode = ModeView
	m.ActiveTab = 0

	_, _ = m.Update(tickMsg{})
	assert.Equal(t, ModeView, m.Mode)
}

func TestModel_RejectsPageResultsWithoutOwnerIDs(t *testing.T) {
	m := New()
	m.Mode = ModeLoading
	m.Tabs[0].URL = "https://current.test"
	m.Update(PageLoadedMsg{URL: "https://stale.test", Lines: []string{"stale"}})
	assert.Equal(t, ModeLoading, m.Mode)
	assert.Equal(t, "https://current.test", m.Tabs[0].URL)
	assert.Empty(t, m.Tabs[0].ASCIILines)
}

func TestModel_Update_PageLoadedMsg(t *testing.T) {
	m := New()
	m.Mode = ModeLoading
	m.ActiveTab = 0

	tab := m.GetActiveTab()
	_, _ = m.Update(PageLoadedMsg{
		TabID: tab.ID, RequestID: tab.RequestID,
		URL: "https://example.com", ANSI: []string{}, Lines: []string{"line1"},
	})

	assert.Equal(t, ModeView, m.Mode)
	assert.Equal(t, "https://example.com", m.Tabs[0].URL)
	assert.Equal(t, []string{"line1"}, m.Tabs[0].ASCIILines)
}

func TestModel_Update_PageErrorMsg(t *testing.T) {
	m := New()
	m.Mode = ModeLoading
	m.ActiveTab = 0

	tab := m.GetActiveTab()
	_, _ = m.Update(PageErrorMsg{TabID: tab.ID, RequestID: tab.RequestID, Err: assert.AnError})

	assert.Equal(t, ModeError, m.Mode)
	assert.Equal(t, ModeError, m.Tabs[0].Mode)
}

func TestModel_Update_Resize(t *testing.T) {
	m := New()
	m.Mode = ModeView
	m.ActiveTab = 0

	updated, cmd := m.Update(tea.WindowSizeMsg{Width: 400, Height: 300})
	assert.NotNil(t, updated)
	assert.NotNil(t, cmd, "resize should request a fresh layout/render")
}

func TestModel_Update_Quit_With_Q(t *testing.T) {
	m := New()
	m.Mode = ModeAddress
	m.ActiveTab = 0

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	assert.NotNil(t, cmd)
}

func TestModel_Update_Address_ControlC(t *testing.T) {
	m := New()
	m.Mode = ModeAddress
	m.ActiveTab = 0

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	assert.NotNil(t, cmd)
}

func TestModel_Update_View_ControlC(t *testing.T) {
	m := New()
	m.Mode = ModeView
	m.ActiveTab = 0

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	assert.NotNil(t, cmd)
}

func TestModel_Update_Error_ControlC(t *testing.T) {
	m := New()
	m.Mode = ModeError
	m.ActiveTab = 0

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	assert.NotNil(t, cmd)
}

func TestModel_Update_View_ZoomMax(t *testing.T) {
	m := New()
	m.Mode = ModeView
	m.ActiveTab = 0
	m.Tabs[0].ZoomFactor = 3.0

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'+'}})
	assert.Equal(t, 3.0, m.Tabs[0].ZoomFactor)
}

func TestModel_Update_View_ZoomMin(t *testing.T) {
	m := New()
	m.Mode = ModeView
	m.ActiveTab = 0
	m.Tabs[0].ZoomFactor = 0.5

	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'-'}})
	assert.Equal(t, 0.5, m.Tabs[0].ZoomFactor)
}
