package js

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/starframe-dev/abyss/internal/dom"
)

func TestGetFinalDOM(t *testing.T) {
	old := rootDoc
	defer func() { rootDoc = old }()
	doc := &dom.Node{Type: dom.DocumentNode}
	rootDoc = doc
	if got := GetFinalDOM(); got != doc {
		t.Fatalf("GetFinalDOM() = %p, want %p", got, doc)
	}
}

func TestFetchScript(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/app.js" {
			t.Errorf("request path = %q, want /app.js", r.URL.Path)
		}
		_, _ = w.Write([]byte("globalThis.testValue = 1"))
	}))
	defer server.Close()
	got, err := fetchScript(server.URL+"/page", "app.js")
	if err != nil {
		t.Fatal(err)
	}
	if got != "globalThis.testValue = 1" {
		t.Fatalf("fetchScript() = %q", got)
	}
	if _, err := fetchScript("://invalid", "app.js"); err == nil {
		t.Fatal("fetchScript() expected invalid base URL error")
	}
}

func TestFetchScriptContext(t *testing.T) {
	if _, err := fetchScriptContext(context.Background(), "://invalid", "app.js"); err == nil {
		t.Fatal("fetchScriptContext() expected invalid base URL error")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()
	if _, err := fetchScriptContext(ctx, server.URL, "app.js"); err == nil {
		t.Fatal("fetchScriptContext() expected canceled-context error")
	}
}

func TestExecuteScriptsContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	doc := &dom.Node{Type: dom.DocumentNode}
	if err := ExecuteScriptsContext(ctx, nil, doc, ""); err == nil {
		t.Fatal("ExecuteScriptsContext() expected canceled-context error")
	}
}

func TestExecuteScriptsEmptyDocument(t *testing.T) {
	doc := &dom.Node{Type: dom.DocumentNode}
	if err := ExecuteScripts(nil, doc, ""); err != nil {
		t.Fatalf("ExecuteScripts() error = %v", err)
	}
}
