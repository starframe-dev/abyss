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

func TestMainConfiguration(t *testing.T) {
	if defaultClientTimeout != 10*time.Second {
		t.Errorf("defaultClientTimeout = %v, want %v", defaultClientTimeout, 10*time.Second)
	}
	if viewportWidth != 800 {
		t.Errorf("viewportWidth = %d, want 800", viewportWidth)
	}
	if defaultClient == nil {
		t.Error("defaultClient is nil")
	}
}

func TestPipelineComponents(t *testing.T) {
	client := &http.Client{}
	if client == nil {
		t.Fatal("expected HTTP client")
	}

	doc := dom.Parse("<html><body><div>test</div></body></html>")
	if doc == nil {
		t.Fatal("dom.Parse returned nil")
	}

	ctx := js.NewContext()
	if ctx == nil {
		t.Fatal("js.NewContext returned nil")
	}
	defer ctx.Close()
	if err := ctx.RegisterBridgeFunctions(); err != nil {
		t.Fatalf("RegisterBridgeFunctions: %v", err)
	}
	js.InitDOM(doc)
	if js.GetFinalDOM() == nil {
		t.Fatal("InitDOM did not initialize DOM")
	}
	if _, err := ctx.Evaluate(`"hello"`); err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if err := js.ExecuteScripts(ctx, doc, "http://example.com"); err != nil {
		t.Fatalf("ExecuteScripts: %v", err)
	}

	styles := css.Compute(doc)
	if styles == nil {
		t.Fatal("css.Compute returned nil")
	}
	box := layout.Build(doc, styles, viewportWidth)
	if box == nil {
		t.Fatal("layout.Build returned nil")
	}
	fb := render.Render(box, viewportWidth)
	if fb == nil {
		t.Fatal("render.Render returned nil")
	}
	if lines := ascii.Render(fb, box, viewportWidth); len(lines) == 0 {
		t.Error("ascii.Render returned no lines")
	}
}
