package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/starframe-dev/abyss/internal/browser"
)

func TestJavaScriptDOMMutationsAreRendered(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!doctype html><html><body>
			<h1 id="title">Before</h1>
			<p class="item" data-state="old">Original text</p>
			<div id="target"></div>
			<button id="action">Click</button>
			<script>
				const title = document.querySelector('#title');
				title.textContent = 'Script updated title';
				const item = document.querySelector('.item');
				item.setAttribute('data-state', 'ready');
				item.textContent = 'Script updated content';
				item.style.color = 'red';
				document.querySelector('#target').innerHTML = '<strong>Inserted by script</strong>';
				let clicks = 0;
				document.querySelector('#action').addEventListener('click', function () { clicks++; });
				document.querySelector('#action').click();
				document.querySelector('#action').textContent = 'Handled ' + clicks + ' click';
			</script>
		</body></html>`))
	}))
	defer server.Close()

	msg := browser.LoadPageForIntegration(context.Background(), server.URL, 80)
	loaded, ok := msg.(browser.PageLoadedMsg)
	if !ok {
		if failure, ok := msg.(browser.PageErrorMsg); ok {
			t.Fatalf("load page: %v", failure.Err)
		}
		t.Fatalf("expected PageLoadedMsg, got %T", msg)
	}
	if len(loaded.Lines) == 0 {
		t.Fatal("rendered page has no lines")
	}
	page := strings.Join(loaded.Lines, "\n")
	for _, want := range []string{"Script updated title", "Script updated content", "Inserted by script", "Handled 1 click"} {
		if !strings.Contains(page, want) {
			t.Errorf("rendered page does not contain %q:\n%s", want, page)
		}
	}
	_ = tea.Quit
}
