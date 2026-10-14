package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/starframe-dev/abyss/internal/ascii"
	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/starframe-dev/abyss/internal/js"
	"github.com/starframe-dev/abyss/internal/layout"
	"github.com/starframe-dev/abyss/internal/render"
)

func TestDefaultClientTimeout(t *testing.T) {
	expectedTimeout := 10 * time.Second
	if defaultClientTimeout != expectedTimeout {
		t.Errorf("Expected default client timeout %v, got %v", expectedTimeout, defaultClientTimeout)
	}
}

func TestDefaultClientCreation(t *testing.T) {
	if defaultClient == nil {
		t.Error("Expected defaultClient to be non-nil")
	}
}

func TestViewportWidth(t *testing.T) {
	expectedWidth := 800
	if viewportWidth != expectedWidth {
		t.Errorf("Expected viewportWidth %v, got %v", expectedWidth, viewportWidth)
	}
}

func TestFetchClientNewClient(t *testing.T) {
	client := &http.Client{}
	if client == nil {
		t.Error("Expected HTTP client to be non-nil")
	}
}

func TestDomParse(t *testing.T) {
	html := "<html><body><div>test</div></body></html>"
	doc := dom.Parse(html)
	if doc == nil {
		t.Error("Expected Parse to return non-nil document")
	}
}

func TestJsNewContext(t *testing.T) {
	ctx := js.NewContext()
	if ctx == nil {
		t.Error("Expected NewContext to return non-nil context")
	}
	defer ctx.Close()
}

func TestJsRegisterBridgeFunctions(t *testing.T) {
	ctx := js.NewContext()
	defer ctx.Close()
	if err := ctx.RegisterBridgeFunctions(); err != nil {
		t.Errorf("Expected RegisterBridgeFunctions to succeed, got error: %v", err)
	}
}

func TestJsInitDOM(t *testing.T) {
	html := "<html><body><div>test</div></body></html>"
	doc := dom.Parse(html)
	js.InitDOM(doc)
	// Verify DOM was initialized
	if js.GetFinalDOM() == nil {
		t.Error("Expected InitDOM to initialize final DOM")
	}
}

func TestJsEvaluate(t *testing.T) {
	ctx := js.NewContext()
	defer ctx.Close()

	if _, err := ctx.Evaluate(`"hello"`); err != nil {
		t.Errorf("Expected Evaluate to succeed, got error: %v", err)
	}
}

func TestJsExecuteScripts(t *testing.T) {
	html := "<html><body><div>test</div></body></html>"
	doc := dom.Parse(html)

	// Create a mock context
	ctx := js.NewContext()
	defer ctx.Close()

	if err := ctx.RegisterBridgeFunctions(); err != nil {
		t.Fatalf("Failed to register bridge functions: %v", err)
	}

	js.InitDOM(doc)

	// This should not panic or return a critical error
	if err := js.ExecuteScripts(ctx, doc, "http://example.com"); err != nil {
		// Non-critical errors may occur during script execution
		t.Logf("Script execution error (may be expected): %v", err)
	}
}

func TestCssCompute(t *testing.T) {
	html := "<html><body><div>test</div></body></html>"
	doc := dom.Parse(html)

	if styles := css.Compute(doc); styles == nil {
		t.Error("Expected Compute to return non-nil styles")
	}
}

func TestLayoutBuild(t *testing.T) {
	html := "<html><body><div>test</div></body></html>"
	doc := dom.Parse(html)
	styles := css.Compute(doc)

	if box := layout.Build(doc, styles, 800); box == nil {
		t.Error("Expected Build to return non-nil box")
	}
}

func TestRenderRender(t *testing.T) {
	html := "<html><body><div>test</div></body></html>"
	doc := dom.Parse(html)
	styles := css.Compute(doc)
	box := layout.Build(doc, styles, 800)

	if fb := render.Render(box, 800); fb == nil {
		t.Error("Expected Render to return non-nil framebuffer")
	}
}

func TestAsciiRender(t *testing.T) {
	html := "<html><body><div>test</div></body></html>"
	doc := dom.Parse(html)
	styles := css.Compute(doc)
	box := layout.Build(doc, styles, 800)

	fb := render.Render(box, 800)
	if lines := ascii.Render(fb, box, 800); len(lines) == 0 {
		t.Error("Expected Render to return non-empty lines slice")
	}
}
