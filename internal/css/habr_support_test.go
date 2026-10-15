package css

import (
	"reflect"
	"testing"
)

func TestParseStylesheet(t *testing.T) {
	tests := []struct {
		name string
		source string
		want []Rule
	}{
		{"plain and selectors", "a, b { color: red; }", []Rule{{Selectors: []string{"a", "b"}, Declarations: []Declaration{{Property:"color", Value:"red"}}}}},
		{"nested conditions", "@media screen { @supports (display: grid) { .x { color: blue; } } }", []Rule{{Selectors: []string{".x"}, Declarations: []Declaration{{Property:"color", Value:"blue"}}, MediaQuery:"screen", SupportsQuery:"(display: grid)"}}},
		{"unsupported and empty", "@import 'x'; .empty {} .ok { color: green }", []Rule{{Selectors: []string{".ok"}, Declarations: []Declaration{{Property:"color", Value:"green"}}}}},
	}
	for _, tt := range tests { t.Run(tt.name, func(t *testing.T) { if got:=ParseStylesheet(tt.source); !reflect.DeepEqual(got,tt.want) { t.Fatalf("ParseStylesheet() = %#v, want %#v",got,tt.want) } }) }
}

func TestResolveCustomProperties(t *testing.T) {
	tests := []struct { name, value string; props map[string]string; want string; ok bool }{
		{"substitute", "color: var(--color)", map[string]string{"--color":"red"}, "color: red", true},
		{"fallback", "var(--missing, blue)", nil, "blue", true},
		{"property beats fallback", "var(--x, blue)", map[string]string{"--x":"red"}, "red", true},
		{"unresolved", "var(--missing)", nil, "var(--missing)", false},
		{"malformed", "var(--x", map[string]string{"--x":"red"}, "var(--x", false},
		{"recursive", "var(--a)", map[string]string{"--a":"var(--b)", "--b":"ok"}, "ok", true},
	}
	for _, tt := range tests { t.Run(tt.name, func(t *testing.T) { got, ok:=ResolveCustomProperties(tt.value,tt.props); if got!=tt.want || ok!=tt.ok { t.Fatalf("ResolveCustomProperties() = %q, %v; want %q, %v",got,ok,tt.want,tt.ok) } }) }
}
