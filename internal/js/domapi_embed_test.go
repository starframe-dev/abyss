package js

import (
	"strings"
	"testing"
)

func TestDOMAPIEmbedded(t *testing.T) {
	for _, fragment := range domapiFragments {
		if !strings.Contains(domapiJSSource, fragment) {
			t.Errorf("embedded DOM API missing fragment %q", fragment)
		}
	}
	if !strings.HasPrefix(domapiJSSource, "(function()") {
		t.Fatalf("unexpected source wrapper: %q", domapiJSSource[:min(20, len(domapiJSSource))])
	}
}
