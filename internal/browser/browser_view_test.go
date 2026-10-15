package browser

import (
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/stretchr/testify/assert"
)

func TestCenter(t *testing.T) {
	for _, tt := range []struct {
		text string
		width int
		want string
	}{{"hello", 10, "  hello"}, {"exact", 5, "exact"}, {"long text", 2, "long text"}, {"", 4, "  "}, {"猫a", 7, "  猫a"}} {
		assert.Equal(t, tt.want, center(tt.text, tt.width))
	}
}

func TestTruncateTerminalText(t *testing.T) {
	for _, tt := range []struct { text string; width int; want string }{
		{"abcdef", 3, "abc"}, {"猫ab", 3, "猫a"}, {"👩‍💻x", 2, "👩‍💻"}, {"abc", 0, ""},
	} {
		assert.Equal(t, tt.want, truncateTerminalText(tt.text, tt.width))
	}
}

func TestViewModesAndLayout(t *testing.T) {
	m := New()
	m.Width = 40
	m.Height = 8
	m.Tabs = []Tab{{URL: "https://example.com", Mode: ModeAddress}, {URL: "", Mode: ModeView}}
	m.ActiveTab = 0
	m.Mode = ModeAddress
	m.Address.SetValue("https://example.com")
	assert.Contains(t, m.View(), "example.com")
	assert.Contains(t, m.View(), "New Tab")

	m.Mode = ModeLoading
	assert.Contains(t, m.View(), "Loading...")
	m.Mode = ModeError
	m.Tabs[0].Err = assert.AnError
	assert.Contains(t, m.View(), "Press Ctrl+L to edit URL or Q to quit")

	m.Mode = ModeView
	m.Tabs[0].ANSILines = []string{"first", "second", "third"}
	m.Viewport = viewport.New(40, 2)
	assert.Contains(t, m.View(), "first")
	m.Viewport.SetYOffset(99)
	assert.Contains(t, m.View(), "third")
}

func TestViewTabURLTruncationAndZoom(t *testing.T) {
	m := New()
	m.Tabs = []Tab{{URL: "https://verylongexample.com/a/long/path", ZoomFactor: 1.5}}
	m.ActiveTab = 0
	m.Mode = ModeView
	m.Width = 80
	m.Tabs[0].ANSILines = []string{"content"}
	assert.Contains(t, m.View(), "…")
	assert.Contains(t, m.View(), "150%")
}

func TestViewEmptyModel(t *testing.T) {
	m := Model{Mode: ModeAddress, Address: textinput.New(), Width: 20}
	assert.Contains(t, m.View(), "Abyss")
}
