package css

import "testing"

func TestParseDisplayWebkitFlexAliases(t *testing.T) {
	for _, v := range []string{"-webkit-flex", "-webkit-box"} {
		if got := parseDisplay(v); got != DisplayFlex {
			t.Fatalf("parseDisplay(%q) = %v, want DisplayFlex", v, got)
		}
	}
}
