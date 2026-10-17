package js

import "testing"

// These bridge functions depend on JavaScriptCore values and are exercised through
// a live Context so their argument validation and returned JS values are covered.
func TestBridgeEvents(t *testing.T) {
	if bad := checkBridgeEventArgs(); len(bad) != 0 {
		t.Fatal(bad)
	}
}
