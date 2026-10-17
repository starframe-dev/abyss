package js

import "testing"

// DOM bridge functions depend on the JavaScriptCore runtime; these smoke tests
// exercise each exported-to-JS callback through valid and missing arguments.
func TestDOMBridgeCallbacks(t *testing.T) {
	if bad := checkDOMBridgeCallbacks(); len(bad) != 0 {
		t.Fatalf("callbacks returned nil JS value: %v", bad)
	}
}
