package css

import "testing"

func TestHex(t *testing.T) {
	for _, tc := range []struct {
		s    string
		want Color
	}{
		{"#abc", RGB(170, 187, 204)}, {"#010203", RGB(1, 2, 3)}, {"bad", Transparent()}, {"#ggg", Transparent()}, {"#12345", Transparent()},
	} {
		if got := Hex(tc.s); got != tc.want {
			t.Errorf("Hex(%q)=%v want %v", tc.s, got, tc.want)
		}
	}
}
func TestParseTextAlign(t *testing.T) {
	for _, tc := range []struct {
		s string
		w TextAlignType
	}{{"center", TextAlignCenter}, {" right ", TextAlignRight}, {"end", TextAlignRight}, {"LEFT", TextAlignLeft}, {"bad", TextAlignLeft}} {
		if g := parseTextAlign(tc.s); g != tc.w {
			t.Errorf("%q: %v != %v", tc.s, g, tc.w)
		}
	}
}
func TestComputedValueParsers(t *testing.T) {
	for _, tc := range []struct {
		s string
		w DisplayType
	}{{"block", DisplayBlock}, {"inline", DisplayInline}, {"inline-block", DisplayInlineBlock}, {"inline-box", DisplayInlineBlock}, {"-webkit-flex", DisplayFlex}, {"-webkit-box", DisplayFlex}, {"flex", DisplayFlex}, {"grid", DisplayGrid}, {"none", DisplayNone}, {"bad", DisplayBlock}} {
		if g := parseDisplay(tc.s); g != tc.w {
			t.Errorf("display %q: %v", tc.s, g)
		}
	}
	for _, tc := range []struct {
		s string
		w PositionType
	}{{"static", PositionStatic}, {"relative", PositionRelative}, {"absolute", PositionAbsolute}, {"fixed", PositionStatic}} {
		if g := parsePosition(tc.s); g != tc.w {
			t.Errorf("position %q", tc.s)
		}
	}
	for _, tc := range []struct {
		s string
		w OverflowType
	}{{"hidden", OverflowHidden}, {"visible", OverflowVisible}, {"auto", OverflowVisible}} {
		if g := parseOverflow(tc.s); g != tc.w {
			t.Errorf("overflow %q", tc.s)
		}
	}
	for _, s := range []string{"auto", "10px", "5%", "2rem", "2em", "1fr", "10", "bad", "NaN", "-2fr", "999999999px"} {
		_ = parseLength(s)
	}
	_ = parseBoxModel("1px 2px 3px 4px")
	_ = parseBoxModel("")
	_ = parseBorder("1px solid red")
	_ = parseBorder("none")
	_ = parseBorder("2 blue")
	_ = parseBorderWidth("10px")
	_ = parseBorderStyle("solid")
	_ = parseBorderStyle("bad")
	for _, s := range []string{"red", "#abc", "rgb(1, 2, 3)", "bad"} {
		_ = parseColor(s)
	}
	for _, s := range []string{"rgb(1,2,3)", "rgb(300,-1,0)", "rgb(no,0,0)", "bad"} {
		_, _ = parseRgb(s)
	}
	_ = parseInt("12")
	_ = parseInt("bad")
	_ = parseFloat("1.5")
	_ = parseFloat("NaN")
	_ = parseFloat("1e999")
	_ = parseFlexDirection("row")
	_ = parseFlexDirection("column")
	_ = parseFlexDirection("bad")
	for _, s := range []string{"flex-start", "center", "flex-end", "space-between", "space-around", "bad"} {
		_ = parseJustifyContent(s)
	}
	for _, s := range []string{"flex-start", "center", "flex-end", "stretch", "bad"} {
		_ = parseAlignItems(s)
	}
	for _, s := range []string{"10px 1fr", "repeat(2, 1fr)", "repeat(bad, 2px)", "repeat(2, 3px", ""} {
		_ = parseLengthList(s)
	}
	for _, s := range []string{"1", "-2.5", "bad"} {
		_ = isNumber(s)
	}
}
