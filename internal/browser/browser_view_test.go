package browser

import (
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/stretchr/testify/assert"
)

func TestCenter(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		width    int
		expected string
	}{
		// center() pads with spaces on the left only (left-aligned padding).
		{"text shorter than width", "hello", 10, "  hello"},
		{"text equal to width", "exact", 5, "exact"},
		{"text longer than width", "thisisalongtext", 5, "thisisalongtext"},
		{"empty text", "", 10, "     "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := center(tt.text, tt.width)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCenterUsesTerminalDisplayWidth(t *testing.T) {
	assert.Equal(t, "  猫a", center("猫a", 7))
	assert.Equal(t, "猫a", truncateTerminalText("猫ab", 3))
}

func TestView_EmptyModel(t *testing.T) {
	m := Model{
		Mode:      ModeAddress,
		Address:   textinput.New(),
		Viewport:  viewport.New(80, 24),
		Tabs:      []Tab{{Mode: ModeAddress}},
		ActiveTab: 0,
		Width:     80,
		Height:    24,
	}

	result := m.View()
	assert.Contains(t, result, "Abyss")
}

func TestView_ModeLoading(t *testing.T) {
	m := New()
	m.Mode = ModeLoading
	m.ActiveTab = 0
	m.Width = 80
	m.Height = 24

	result := m.View()
	assert.Contains(t, result, "Loading...")
}

func TestView_ModeError(t *testing.T) {
	m := New()
	m.Mode = ModeError
	m.Tabs[0].Err = assert.AnError
	m.ActiveTab = 0
	m.Width = 80
	m.Height = 24

	result := m.View()
	assert.Contains(t, result, "Error:")
	assert.Contains(t, result, "Press Ctrl+L to edit URL or Q to quit")
}

func TestView_ModeView(t *testing.T) {
	m := New()
	m.Mode = ModeView
	m.Tabs[0].ANSILines = []string{"<ansi>line 1</ansi>", "<ansi>line 2</ansi>"}
	m.Tabs[0].ASCIILines = []string{"asciiline 1", "asciiline 2"}
	m.ActiveTab = 0
	m.Width = 80
	m.Height = 24

	result := m.View()
	assert.Contains(t, result, "line 1")
	assert.Contains(t, result, "line 2")
}

func TestView_ModeAddress(t *testing.T) {
	m := New()
	m.Address.SetValue("https://example.com")
	m.Mode = ModeAddress
	m.ActiveTab = 0
	m.Width = 80
	m.Height = 24

	result := m.View()
	assert.Contains(t, result, "example.com")
}

func TestView_Tabs(t *testing.T) {
	m := New()
	m.Tabs = []Tab{
		{URL: "https://example.com", Mode: ModeAddress},
		{URL: "https://test.com", Mode: ModeAddress},
	}
	m.ActiveTab = 0
	m.Width = 80
	m.Height = 24

	result := m.View()
	assert.Contains(t, result, "example.com")
	assert.Contains(t, result, "test.com")
}

func TestView_EmptyTabLabels(t *testing.T) {
	m := New()
	m.Tabs = []Tab{
		{URL: "", Mode: ModeAddress},
	}
	m.ActiveTab = 0
	m.Width = 80
	m.Height = 24

	result := m.View()
	assert.Contains(t, result, "New Tab")
}

func TestView_ShortURLs(t *testing.T) {
	m := New()
	m.Tabs = []Tab{
		{URL: "https://verylongexample.com/path/to/somewhere/very/deep", Mode: ModeAddress},
	}
	m.ActiveTab = 0
	m.Width = 80
	m.Height = 24

	result := m.View()
	assert.Contains(t, result, "verylongexample.com")
}

func TestView_ZoomFactor(t *testing.T) {
	m := New()
	m.Tabs[0].ZoomFactor = 2.0
	m.Mode = ModeView
	m.ActiveTab = 0
	m.Width = 80
	m.Height = 24

	result := m.View()
	assert.Contains(t, result, "Abyss [200%]")
}

func TestView_ZoomFactor_OnePointZero(t *testing.T) {
	m := New()
	m.Tabs[0].ZoomFactor = 1.0
	m.Mode = ModeView
	m.ActiveTab = 0
	m.Width = 80
	m.Height = 24

	result := m.View()
	assert.NotContains(t, result, "[100%]")
}

func TestView_OffsetX(t *testing.T) {
	m := New()
	m.Tabs[0].OffsetX = 10
	m.Mode = ModeView
	m.Tabs[0].ASCIILines = []string{"content line 1", "content line 2"}
	m.ActiveTab = 0
	m.Width = 80
	m.Height = 24

	result := m.View()
	assert.NotContains(t, result, "content")
}

func TestView_OffsetX_Zero(t *testing.T) {
	m := New()
	m.Tabs[0].OffsetX = 0
	m.Mode = ModeView
	m.Tabs[0].ASCIILines = []string{"content line 1", "content line 2"}
	m.ActiveTab = 0
	m.Width = 80
	m.Height = 24

	result := m.View()
	// Verify View renders content with zero offset.
	assert.NotEmpty(t, result)
}

func TestView_EmptyTabsSlice(t *testing.T) {
	m := Model{
		Mode:      ModeAddress,
		Address:   textinput.New(),
		Viewport:  viewport.New(80, 24),
		Tabs:      []Tab{},
		ActiveTab: 0,
		Width:     80,
		Height:    24,
	}

	result := m.View()
	assert.Contains(t, result, "Abyss")
}

func TestView_NegativeActiveTab(t *testing.T) {
	m := New()
	m.Tabs = []Tab{{URL: "https://example.com", Mode: ModeAddress}}
	m.ActiveTab = -1
	m.Width = 80
	m.Height = 24

	result := m.View()
	assert.Contains(t, result, "example.com")
}

func TestView_LargeActiveTab(t *testing.T) {
	m := New()
	m.Tabs = []Tab{{URL: "https://example.com", Mode: ModeAddress}}
	m.ActiveTab = 100
	m.Width = 80
	m.Height = 24

	result := m.View()
	assert.Contains(t, result, "example.com")
}

func TestView_MixedModes(t *testing.T) {
	m := New()
	m.Tabs = []Tab{
		{URL: "https://example.com", Mode: ModeView},
		{URL: "https://test.com", Mode: ModeAddress},
	}
	m.ActiveTab = 0
	m.Width = 80
	m.Height = 24

	result := m.View()
	assert.Contains(t, result, "example.com")
	assert.Contains(t, result, "test.com")
}

func TestView_EmptyLines(t *testing.T) {
	m := New()
	m.Tabs[0].ASCIILines = []string{""}
	m.Mode = ModeView
	m.ActiveTab = 0
	m.Width = 80
	m.Height = 24

	result := m.View()
	// Just verify View returns a non-empty string in this state.
	assert.NotEmpty(t, result)
}

func TestView_ActiveTabNil(t *testing.T) {
	m := New()
	m.Tabs = []Tab{{URL: "https://example.com", Mode: ModeView, ANSILines: []string{"<ansi>line</ansi>"}, ASCIILines: []string{"line"}}}
	m.ActiveTab = 0
	m.Width = 80
	m.Height = 24

	result := m.View()
	assert.Contains(t, result, "example.com")
}

func TestView_ActiveTabOutOfBounds(t *testing.T) {
	m := New()
	m.Tabs = []Tab{{URL: "https://example.com", Mode: ModeView, ANSILines: []string{"<ansi>line</ansi>"}, ASCIILines: []string{"line"}}}
	m.ActiveTab = 1000
	m.Width = 80
	m.Height = 24

	result := m.View()
	assert.Contains(t, result, "example.com")
}

func TestView_EmptyModelNoTabs(t *testing.T) {
	m := Model{
		Mode:      ModeAddress,
		Address:   textinput.New(),
		Viewport:  viewport.New(80, 24),
		Tabs:      nil,
		ActiveTab: 0,
		Width:     80,
		Height:    24,
	}

	result := m.View()
	assert.Contains(t, result, "Abyss")
}

func TestView_ErrorMessage(t *testing.T) {
	m := New()
	m.Tabs[0].Err = assert.AnError
	m.Mode = ModeError
	m.ActiveTab = 0
	m.Width = 80
	m.Height = 24

	result := m.View()
	assert.Contains(t, result, "Error:")
}

func TestView_MultipleTabs(t *testing.T) {
	m := New()
	m.Tabs = []Tab{
		{URL: "https://example1.com", Mode: ModeView, ANSILines: []string{"<ansi>line1</ansi>"}, ASCIILines: []string{"asciiline1"}},
		{URL: "https://example2.com", Mode: ModeView, ANSILines: []string{"<ansi>line2</ansi>"}, ASCIILines: []string{"asciiline2"}},
		{URL: "https://example3.com", Mode: ModeView, ANSILines: []string{"<ansi>line3</ansi>"}, ASCIILines: []string{"asciiline3"}},
	}
	m.ActiveTab = 1
	m.Width = 80
	m.Height = 24

	result := m.View()
	assert.Contains(t, result, "example1.com")
	assert.Contains(t, result, "example2.com")
	assert.Contains(t, result, "example3.com")
}

func TestView_TabsWithNilANSILines(t *testing.T) {
	m := New()
	m.Tabs = []Tab{
		{URL: "https://example.com", Mode: ModeView, ANSILines: nil, ASCIILines: []string{"line"}},
	}
	m.ActiveTab = 0
	m.Width = 80
	m.Height = 24

	result := m.View()
	// Just verify the view renders something.
	assert.NotEmpty(t, result)
}

func TestView_ActiveTabNilPointer(t *testing.T) {
	m := New()
	m.Tabs = []Tab{{URL: "https://example.com", Mode: ModeView, ANSILines: []string{"<ansi>line</ansi>"}, ASCIILines: []string{"line"}}}
	m.ActiveTab = 0
	m.Width = 80
	m.Height = 24

	result := m.View()
	assert.Contains(t, result, "example.com")
}

func TestView_ActiveTabOutOfBoundsNegative(t *testing.T) {
	m := New()
	m.Tabs = []Tab{{URL: "https://example.com", Mode: ModeView, ANSILines: []string{"<ansi>line</ansi>"}, ASCIILines: []string{"line"}}}
	m.ActiveTab = -1
	m.Width = 80
	m.Height = 24

	result := m.View()
	assert.Contains(t, result, "example.com")
}

func TestView_ActiveTabOutOfBoundsPositive(t *testing.T) {
	m := New()
	m.Tabs = []Tab{{URL: "https://example.com", Mode: ModeView, ANSILines: []string{"<ansi>line</ansi>"}, ASCIILines: []string{"line"}}}
	m.ActiveTab = 1000
	m.Width = 80
	m.Height = 24

	result := m.View()
	assert.Contains(t, result, "example.com")
}

func TestView_EmptyModelNoTabsWithNil(t *testing.T) {
	m := Model{
		Mode:      ModeAddress,
		Address:   textinput.New(),
		Viewport:  viewport.New(80, 24),
		Tabs:      nil,
		ActiveTab: 0,
		Width:     80,
		Height:    24,
	}

	result := m.View()
	assert.Contains(t, result, "Abyss")
}

func TestView_EmptyModelNoTabsWithEmptySlice(t *testing.T) {
	m := Model{
		Mode:      ModeAddress,
		Address:   textinput.New(),
		Viewport:  viewport.New(80, 24),
		Tabs:      []Tab{},
		ActiveTab: 0,
		Width:     80,
		Height:    24,
	}

	result := m.View()
	assert.Contains(t, result, "Abyss")
}

func TestView_EmptyLinesWithZeroWidth(t *testing.T) {
	m := New()
	m.Tabs[0].ASCIILines = []string{"a"}
	m.Mode = ModeView
	m.ActiveTab = 0
	m.Width = 0
	m.Height = 24

	result := m.View()
	assert.Contains(t, result, "a")
}

func TestView_EmptyLinesWithNegativeWidth(t *testing.T) {
	m := New()
	m.Tabs[0].ASCIILines = []string{"a"}
	m.Mode = ModeView
	m.ActiveTab = 0
	m.Width = -1
	m.Height = 24

	result := m.View()
	assert.Contains(t, result, "a")
}

func TestView_EmptyLinesWithZeroHeight(t *testing.T) {
	m := New()
	m.Tabs[0].ASCIILines = []string{"a", "b", "c", "d", "e"}
	m.Mode = ModeView
	m.ActiveTab = 0
	m.Width = 80
	m.Height = 0

	result := m.View()
	assert.Contains(t, result, "a")
}
