package js

import "testing"

func TestBridgeQueryJSFunctions(t *testing.T) {
	if bad := checkQueryCallbacks(); len(bad) != 0 {
		t.Fatalf("query callbacks returned nil JSValueRef: %v", bad)
	}
}
