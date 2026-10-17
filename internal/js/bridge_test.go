package js

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/dom"
)

func TestBridgeHandlersRegistered(t *testing.T) {
	want := []string{
		"createElement", "appendChild", "setInnerHTML", "setAttribute", "setTextContent",
		"createTextNode", "createDocumentFragment", "setTimeout", "setInterval",
		"clearTimeout", "clearInterval", "getBody", "getElementById", "getAttribute",
		"removeChild", "getParentNode", "getChildren", "getInnerHTML", "getTextContent",
		"getElementsByTagName", "getElementsByClassName", "querySelector", "querySelectorAll",
		"addEventListener", "removeEventListener", "dispatchEvent", "getComputedStyle",
		"matches", "closest", "insertBefore", "replaceChild", "cloneNode", "contains",
		"getBoundingClientRect", "previousElementSibling", "nextElementSibling",
		"firstElementChild", "lastElementChild", "insertAdjacentHTML", "insertAdjacentText",
		"insertAdjacentElement", "outerHTML", "setOuterHTML", "log",
	}
	if len(bridgeHandlers) != len(want) {
		t.Fatalf("bridge handler count = %d, want %d", len(bridgeHandlers), len(want))
	}
	for _, name := range want {
		if bridgeHandlers[name] == nil {
			t.Errorf("missing bridge handler %q", name)
		}
	}
}

// A nil node has no bridge ID, so the wrapper must be JS null, not a Go nil.
func TestNodeToJSValueNil(t *testing.T) {
	ctx := NewContext()
	defer ctx.Close()
	if got := ctx.GetType(nodeToJSValue(ctx, (*dom.Node)(nil))); got != "null" {
		t.Fatalf("nodeToJSValue(nil node) type = %q, want null", got)
	}
}
