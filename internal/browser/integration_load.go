package browser

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
)

// LoadPageForIntegration loads a page synchronously and returns the resulting
// message (PageLoadedMsg or PageErrorMsg). It exists for the integration tests,
// which need the outcome without driving the Bubble Tea loop.
func LoadPageForIntegration(ctx context.Context, pageURL string, viewportWidth int) tea.Msg {
	return loadPageContext(ctx, pageURL, viewportWidth)()
}
