package css

import (
	"testing"
)

func TestParseDisplay(t *testing.T) {
	tests := []struct {
		input    string
		expected DisplayType
	}{
		{"block", DisplayBlock},
		{"inline", DisplayInline},
		{"flex", DisplayFlex},
		{"grid", DisplayGrid},
		{"none", DisplayNone},
		{"invalid", DisplayBlock},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseDisplay(tt.input)
			if result != tt.expected {
				t.Errorf("parseDisplay(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParsePosition(t *testing.T) {
	tests := []struct {
		input    string
		expected PositionType
	}{
		{"static", PositionStatic},
		{"relative", PositionRelative},
		{"absolute", PositionAbsolute},
		{"fixed", PositionStatic},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parsePosition(tt.input)
			if result != tt.expected {
				t.Errorf("parsePosition(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseOverflow(t *testing.T) {
	tests := []struct {
		input    string
		expected OverflowType
	}{
		{"hidden", OverflowHidden},
		{"visible", OverflowVisible},
		{"auto", OverflowVisible},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseOverflow(tt.input)
			if result != tt.expected {
				t.Errorf("parseOverflow(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseLength(t *testing.T) {
	tests := []struct {
		input    string
		expected Length
	}{
		{"10px", Px(10)},
		{"5%", Percent(5)},
		{"1rem", Px(16)},
		{"1em", Px(16)},
		{"1fr", Fr(1)},
		{"auto", Auto()},
		{"10", Px(10)},
		{"", Auto()},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseLength(tt.input)
			if result != tt.expected {
				t.Errorf("parseLength(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseBoxModel(t *testing.T) {
	tests := []struct {
		input    string
		expected BoxModel
	}{
		{"10px", BoxModel{Top: Px(10), Right: Px(10), Bottom: Px(10), Left: Px(10)}},
		{"10px 20px", BoxModel{Top: Px(10), Right: Px(20), Bottom: Px(10), Left: Px(20)}},
		{"10px 20px 30px", BoxModel{Top: Px(10), Right: Px(20), Bottom: Px(30), Left: Px(20)}},
		{"10px 20px 30px 40px", BoxModel{Top: Px(10), Right: Px(20), Bottom: Px(30), Left: Px(40)}},
		{"", BoxModel{}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseBoxModel(tt.input)
			if result != tt.expected {
				t.Errorf("parseBoxModel(%q) = %+v, want %+v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseBorder(t *testing.T) {
	tests := []struct {
		input string
		want  Border
	}{
		{"1px solid red", Border{Width: 1, Style: BorderSolid, Color: Hex("#ff0000")}},
		{"2px red", Border{Width: 2, Style: BorderNone, Color: Hex("#ff0000")}},
		{"solid", Border{Style: BorderSolid}},
		{"none", Border{Style: BorderNone}},
		{"1px", Border{Width: 1, Style: BorderNone}},
		{"red", Border{Style: BorderNone, Color: Hex("#ff0000")}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseBorder(tt.input)
			if result != tt.want {
				t.Errorf("parseBorder(%q) = %+v, want %+v", tt.input, result, tt.want)
			}
		})
	}
}

func TestParseBorderWidth(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"10px", 10},
		{"20px", 20},
		{"5", 5},
		{"", 0},
		{"invalid", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseBorderWidth(tt.input)
			if result != tt.want {
				t.Errorf("parseBorderWidth(%q) = %v, want %v", tt.input, result, tt.want)
			}
		})
	}
}

func TestParseBorderStyle(t *testing.T) {
	tests := []struct {
		input    string
		expected BorderStyle
	}{
		{"none", BorderNone},
		{"solid", BorderSolid},
		{"invalid", BorderNone},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseBorderStyle(tt.input)
			if result != tt.expected {
				t.Errorf("parseBorderStyle(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseColor(t *testing.T) {
	tests := []struct {
		input string
		want  Color
	}{
		{"red", RGB(255, 0, 0)},
		{"#ff0000", Hex("#ff0000")},
		{"#fff", Hex("#fff")},
		{"rgb(255, 0, 0)", RGB(255, 0, 0)},
		{"transparent", Transparent()},
		{"invalid", Transparent()},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseColor(tt.input)
			if result != tt.want {
				t.Errorf("parseColor(%q) = %v, want %v", tt.input, result, tt.want)
			}
		})
	}
}

func TestParseRgb(t *testing.T) {
	tests := []struct {
		input string
		want  Color
		ok    bool
	}{
		{"rgb(255, 0, 0)", RGB(255, 0, 0), true},
		{"rgb(100, 200, 50)", RGB(100, 200, 50), true},
		{"rgb(10,20,30)", RGB(10, 20, 30), true},
		{"rgb()", Color{}, false},
		{"rgb(10)", Color{}, false},
		{"invalid", Color{}, false},
		{"rgb(256, 0, 0)", RGB(255, 0, 0), true},
		{"rgb(-5, 0, 300)", RGB(0, 0, 255), true},
		{"rgb(nope, 0, 0)", Color{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, ok := parseRgb(tt.input)
			if ok != tt.ok {
				t.Errorf("parseRgb(%q) ok = %v, want %v", tt.input, ok, tt.ok)
			}
			if !tt.ok && result != tt.want {
				t.Errorf("parseRgb(%q) = %v, want %v", tt.input, result, tt.want)
			}
			if tt.ok && result != tt.want {
				t.Errorf("parseRgb(%q) = %v, want %v", tt.input, result, tt.want)
			}
		})
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"10", 10},
		{"-5", -5},
		{"invalid", 0},
		{"  42  ", 42},
		{"", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseInt(tt.input)
			if result != tt.want {
				t.Errorf("parseInt(%q) = %v, want %v", tt.input, result, tt.want)
			}
		})
	}
}

func TestParseFloat(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"3.14", 3.14},
		{"-2.5", -2.5},
		{"invalid", 0},
		{"  1.5  ", 1.5},
		{"", 0},
		{"NaN", 0},
		{"+Inf", 0},
		{"999999999", maxCSSLength},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseFloat(tt.input)
			if result != tt.want {
				t.Errorf("parseFloat(%q) = %v, want %v", tt.input, result, tt.want)
			}
		})
	}
}

func TestParseFlexDirection(t *testing.T) {
	tests := []struct {
		input    string
		expected FlexDirection
	}{
		{"row", FlexDirectionRow},
		{"column", FlexDirectionColumn},
		{"invalid", FlexDirectionRow},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseFlexDirection(tt.input)
			if result != tt.expected {
				t.Errorf("parseFlexDirection(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseJustifyContent(t *testing.T) {
	tests := []struct {
		input    string
		expected JustifyContent
	}{
		{"flex-start", JustifyFlexStart},
		{"center", JustifyCenter},
		{"flex-end", JustifyFlexEnd},
		{"space-between", JustifySpaceBetween},
		{"space-around", JustifySpaceAround},
		{"invalid", JustifyFlexStart},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseJustifyContent(tt.input)
			if result != tt.expected {
				t.Errorf("parseJustifyContent(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseAlignItems(t *testing.T) {
	tests := []struct {
		input    string
		expected AlignItems
	}{
		{"flex-start", AlignFlexStart},
		{"center", AlignCenter},
		{"flex-end", AlignFlexEnd},
		{"stretch", AlignStretch},
		{"invalid", AlignStretch},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseAlignItems(tt.input)
			if result != tt.expected {
				t.Errorf("parseAlignItems(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseLengthBoundsExtremeValues(t *testing.T) {
	if got := parseLength("999999999px"); got.Px != maxCSSLength {
		t.Fatalf("length bound = %v, want %v", got.Px, maxCSSLength)
	}
	if got := parseLength("1e999px"); !got.IsAuto {
		t.Fatalf("non-finite CSS value should be rejected: %#v", got)
	}
}

func TestParseLengthListBoundsRepeatExpansion(t *testing.T) {
	got := parseLengthList("repeat(999999999, 1fr)")
	if len(got) != maxGridTracks {
		t.Fatalf("expanded %d tracks, want limit %d", len(got), maxGridTracks)
	}
}

func TestParseLengthList(t *testing.T) {
	tests := []struct {
		input string
		want  []Length
	}{
		{"10px", []Length{Px(10)}},
		{"10px 20px 30px", []Length{Px(10), Px(20), Px(30)}},
		{"10px,20px,30px", []Length{Px(10), Px(20), Px(30)}},
		{"repeat(3, 1fr)", []Length{Fr(1), Fr(1), Fr(1)}},
		{"repeat(2, 10px)", []Length{Px(10), Px(10)}},
		{"1fr repeat(2, 2fr)", []Length{Fr(1), Fr(2), Fr(2)}},
		{"", []Length{}},
		{"invalid", []Length{Auto()}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseLengthList(tt.input)
			if len(result) != len(tt.want) {
				t.Errorf("parseLengthList(%q) = %v, want len %d", tt.input, result, len(tt.want))
			}
			for i := range result {
				if result[i] != tt.want[i] {
					t.Errorf("parseLengthList[%d] = %v, want %v", i, result[i], tt.want[i])
				}
			}
		})
	}
}

func TestIsNumber(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"10", true},
		{"-5", true},
		{"3.14", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isNumber(tt.input)
			if result != tt.want {
				t.Errorf("isNumber(%q) = %v, want %v", tt.input, result, tt.want)
			}
		})
	}
}

func TestNamedColors(t *testing.T) {
	if len(namedColors) < 2 {
		t.Fatal("namedColors should be populated")
	}

	// Check a few named colors exist
	if _, ok := namedColors["red"]; !ok {
		t.Error("namedColors should contain 'red'")
	}
	if _, ok := namedColors["blue"]; !ok {
		t.Error("namedColors should contain 'blue'")
	}
	if _, ok := namedColors["transparent"]; !ok {
		t.Error("namedColors should contain 'transparent'")
	}
}
