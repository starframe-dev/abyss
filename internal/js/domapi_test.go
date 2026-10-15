package js

import "testing"

func TestDOMAPIScript(t *testing.T) {
	got := DOMAPIScript()
	if got == "" {
		t.Fatal("DOMAPIScript() returned an empty script")
	}
	if got != domapiJSSource {
		t.Fatal("DOMAPIScript() did not return the DOM API source")
	}
}
