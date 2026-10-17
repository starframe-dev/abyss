package js

import "testing"

func TestEventListenersAndDispatch(t *testing.T) {
	if bad := checkEventListeners(); len(bad) != 0 {
		t.Fatal(bad)
	}
}

func TestTimers(t *testing.T) {
	if bad := checkTimers(); len(bad) != 0 {
		t.Fatal(bad)
	}
}
