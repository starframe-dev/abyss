package css

import "testing"

func TestLengthConstructorsResolveAndString(t *testing.T) {
	tests := []struct {
		name        string
		got         Length
		wantResolve float64
		wantString  string
	}{
		{"auto", Auto(), 0, "auto"},
		{"pixels", Px(12.5), 12.5, "12.5px"},
		{"percent", Percent(25), 50, "25%"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.got.Resolve(200); got != tt.wantResolve {
				t.Errorf("Resolve() = %v, want %v", got, tt.wantResolve)
			}
			if got := tt.got.String(); got != tt.wantString {
				t.Errorf("String() = %q, want %q", got, tt.wantString)
			}
		})
	}
}

func TestColors(t *testing.T) {
	if c := Transparent(); c.IsSet || c.String() != "" {
		t.Errorf("Transparent() = %#v, String() = %q", c, c.String())
	}
	if c := RGB(1, 128, 255); !c.IsSet || c.String() != "#0180ff" {
		t.Errorf("RGB() = %#v, String() = %q", c, c.String())
	}
	for _, tt := range []struct{ input, want string }{{"#abc", "#aabbcc"}, {"aabbcc", "#aabbcc"}, {"#12FG00", ""}, {"#12345", ""}, {"", ""}} {
		c := Hex(tt.input)
		if got := c.String(); got != tt.want {
			t.Errorf("Hex(%q).String() = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestBoxModelIsZero(t *testing.T) {
	if !(BoxModel{}).IsZero() {
		t.Fatal("zero BoxModel should be zero")
	}
	for _, side := range []Length{Px(1), Auto(), Percent(0)} {
		b := BoxModel{Top: side}
		if b.IsZero() {
			t.Errorf("BoxModel with side %#v should not be zero", side)
		}
	}
}

func TestDefault(t *testing.T) {
	s := Default()
	if s.Display != DisplayBlock || s.Position != PositionStatic || s.Overflow != OverflowVisible {
		t.Errorf("unexpected display/position/overflow defaults: %#v", s)
	}
	if !s.Width.IsAuto || !s.Height.IsAuto {
		t.Error("width and height should default to auto")
	}
	if s.BackgroundColor.IsSet || s.Color != (RGB(30, 30, 30)) {
		t.Error("unexpected color defaults")
	}
	if s.FlexDirection != FlexDirectionRow || s.JustifyContent != JustifyFlexStart || s.AlignItems != AlignStretch {
		t.Error("unexpected flex defaults")
	}
	if s.FlexGrow != 0 || s.FlexShrink != 1 || s.Gap != Px(0) {
		t.Error("unexpected grow, shrink, or gap defaults")
	}
}
