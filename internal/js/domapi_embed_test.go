package js

import (
	"strings"
	"testing"
)

func TestDOMAPIEmbedded(t *testing.T) {
	for _, fragment := range domapiFragments {
		data, err := domapiFiles.ReadFile("domapi/" + fragment + ".js")
		if err != nil {
			t.Fatalf("read fragment %q: %v", fragment, err)
		}
		if !strings.Contains(domapiJSSource, string(data)) {
			t.Errorf("embedded DOM API missing content of fragment %q", fragment)
		}
	}
	if !strings.HasPrefix(domapiJSSource, "(function()") {
		t.Fatalf("unexpected source wrapper: %q", domapiJSSource[:min(20, len(domapiJSSource))])
	}
}
