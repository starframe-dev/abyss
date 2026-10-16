package css

import "testing"

func TestColors(t *testing.T) {
	if c := Transparent(); c.IsSet || c.String() != "transparent" {
		t.Errorf("Transparent() = %#v, String() = %q", c, c.String())
	}
	if c := RGB(1, 128, 255); !c.IsSet || c.String() != "#0180ff" {
		t.Errorf("RGB() = %#v, String() = %q", c, c.String())
	}
	// Hex requires a leading '#'; anything invalid resolves to transparent.
	for _, tt := range []struct{ input, want string }{{"#abc", "#aabbcc"}, {"#12FG00", "transparent"}, {"#12345", "transparent"}, {"", "transparent"}} {
		c := Hex(tt.input)
		if got := c.String(); got != tt.want {
			t.Errorf("Hex(%q).String() = %q, want %q", tt.input, got, tt.want)
		}
	}
}
