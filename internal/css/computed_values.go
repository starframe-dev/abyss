// CSS value parsers: convert raw property values into typed computed-style values.
package css

import (
	"math"
	"strconv"
	"strings"
)

// namedColors отображает именованные CSS-цвета в типы Color.
var namedColors = map[string]Color{
	"transparent": Transparent(),
	"black":       RGB(0, 0, 0),
	"white":       RGB(255, 255, 255),
	"red":         RGB(255, 0, 0),
	"green":       RGB(0, 128, 0),
	"blue":        RGB(0, 0, 255),
	"yellow":      RGB(255, 255, 0),
	"cyan":        RGB(0, 255, 255),
	"magenta":     RGB(255, 0, 255),
	"gray":        RGB(128, 128, 128),
	"lightgray":   RGB(211, 211, 211),
	"darkgray":    RGB(169, 169, 169),
	"orange":      RGB(255, 165, 0),
	"purple":      RGB(128, 0, 128),
	"pink":        RGB(255, 192, 203),
	"brown":       RGB(165, 42, 42),
}

// Hex parses a #RGB or #RRGGBB color.
func Hex(s string) Color {
	var r, g, b uint8
	if len(s) == 4 {
		r = uint8(hexToInt(string(s[1])) * 17)
		g = uint8(hexToInt(string(s[2])) * 17)
		b = uint8(hexToInt(string(s[3])) * 17)
	} else if len(s) == 7 {
		r = uint8(hexToInt(s[1:3]))
		g = uint8(hexToInt(s[3:5]))
		b = uint8(hexToInt(s[5:7]))
	}
	return Color{R: r, G: g, B: b, IsSet: true}
}

// parseDisplay парсит значение свойства display в тип DisplayType.
func parseDisplay(val string) DisplayType {
	switch val {
	case "block":
		return DisplayBlock
	case "inline":
		return DisplayInline
	case "flex":
		return DisplayFlex
	case "grid":
		return DisplayGrid
	case "none":
		return DisplayNone
	}
	return DisplayBlock
}

// parsePosition парсит значение свойства position в тип PositionType.
func parsePosition(val string) PositionType {
	switch val {
	case "static":
		return PositionStatic
	case "relative":
		return PositionRelative
	case "absolute":
		return PositionAbsolute
	}
	return PositionStatic
}

// parseOverflow парсит значение свойства overflow в тип OverflowType.
func parseOverflow(val string) OverflowType {
	switch val {
	case "hidden":
		return OverflowHidden
	case "visible":
		return OverflowVisible
	}
	return OverflowVisible
}

const maxCSSLength = 1e6

func parseBoundedFloat(value string) (float64, bool) {
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) {
		return 0, false
	}
	return max(-maxCSSLength, min(maxCSSLength, parsed)), true
}

// parseLength парсит CSS-длину (px, %, rem/em, fr, простое число) в тип Length.
func parseLength(val string) Length {
	val = strings.TrimSpace(val)
	if val == "auto" {
		return Auto()
	}
	if strings.HasSuffix(val, "px") {
		v, ok := parseBoundedFloat(strings.TrimSuffix(val, "px"))
		if !ok {
			return Auto()
		}
		return Px(v)
	}
	if strings.HasSuffix(val, "%") {
		v, ok := parseBoundedFloat(strings.TrimSuffix(val, "%"))
		if !ok {
			return Auto()
		}
		return Percent(v)
	}
	if strings.HasSuffix(val, "rem") || strings.HasSuffix(val, "em") {
		// Treat rem/em as px for simplicity (assuming 1rem = 16px)
		v, ok := parseBoundedFloat(strings.TrimSuffix(strings.TrimSuffix(val, "rem"), "em"))
		if !ok {
			return Auto()
		}
		return Px(max(-maxCSSLength, min(maxCSSLength, v*16)))
	}
	if strings.HasSuffix(val, "fr") {
		v, ok := parseBoundedFloat(strings.TrimSuffix(val, "fr"))
		if !ok || v < 0 {
			return Auto()
		}
		return Fr(v)
	}
	// Try plain number as px
	v, ok := parseBoundedFloat(val)
	if ok {
		return Px(v)
	}
	return Auto()
}

// parseBoxModel парсит 1-4 значения в shorthand-формате в тип BoxModel.
func parseBoxModel(val string) BoxModel {
	parts := strings.Fields(val)
	switch len(parts) {
	case 1:
		l := parseLength(parts[0])
		return BoxModel{Top: l, Right: l, Bottom: l, Left: l}
	case 2:
		v := parseLength(parts[0])
		h := parseLength(parts[1])
		return BoxModel{Top: v, Right: h, Bottom: v, Left: h}
	case 3:
		t := parseLength(parts[0])
		h := parseLength(parts[1])
		b := parseLength(parts[2])
		return BoxModel{Top: t, Right: h, Bottom: b, Left: h}
	case 4:
		return BoxModel{
			Top:    parseLength(parts[0]),
			Right:  parseLength(parts[1]),
			Bottom: parseLength(parts[2]),
			Left:   parseLength(parts[3]),
		}
	}
	return BoxModel{}
}

// parseBorder парсит shorthand-значение свойства border в тип Border.
func parseBorder(val string) Border {
	parts := strings.Fields(val)
	b := Border{Style: BorderNone}
	for _, part := range parts {
		part = strings.ToLower(part)
		switch part {
		case "none":
			b.Style = BorderNone
		case "solid":
			b.Style = BorderSolid
		default:
			if strings.HasSuffix(part, "px") || isNumber(part) {
				b.Width, _ = strconv.ParseFloat(strings.TrimSuffix(part, "px"), 64)
				if b.Width == 0 && isNumber(part) {
					b.Width, _ = strconv.ParseFloat(part, 64)
				}
			} else {
				b.Color = parseColor(part)
			}
		}
	}
	return b
}

// parseBorderWidth парсит ширину границы в значение float64.
func parseBorderWidth(val string) float64 {
	v, _ := strconv.ParseFloat(strings.TrimSuffix(val, "px"), 64)
	return v
}

// parseBorderStyle парсит значение стиля границы в тип BorderStyle.
func parseBorderStyle(val string) BorderStyle {
	switch val {
	case "none":
		return BorderNone
	case "solid":
		return BorderSolid
	}
	return BorderNone
}

// parseColor парсит CSS-цвет (именованный, hex, rgb()) в тип Color.
func parseColor(val string) Color {
	val = strings.ToLower(strings.TrimSpace(val))
	if c, ok := namedColors[val]; ok {
		return c
	}
	if strings.HasPrefix(val, "#") {
		return Hex(val)
	}
	if c, ok := parseRgb(val); ok {
		return c
	}
	return Transparent()
}

// parseRgb парсит значение rgb(r, g, b) в тип Color. Возвращает ok = false для неизвестных значений.
func parseRgb(val string) (Color, bool) {
	if !strings.HasPrefix(val, "rgb(") || !strings.HasSuffix(val, ")") {
		return Color{}, false
	}
	inner := val[4 : len(val)-1]
	parts := strings.Split(inner, ",")
	if len(parts) != 3 {
		return Color{}, false
	}
	channels := [3]uint8{}
	for i, part := range parts {
		value, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			return Color{}, false
		}
		channels[i] = uint8(max(0, min(255, value)))
	}
	return RGB(channels[0], channels[1], channels[2]), true
}

// parseInt парсит целое число, игнорируя ошибки.
func parseInt(val string) int {
	v, _ := strconv.Atoi(strings.TrimSpace(val))
	return v
}

// parseFloat парсит число с плавающей точкой, игнорируя ошибки.
func parseFloat(val string) float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(val), 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	return max(-maxCSSLength, min(maxCSSLength, v))
}

// parseFlexDirection парсит значение свойства flex-direction в тип FlexDirection.
func parseFlexDirection(val string) FlexDirection {
	switch val {
	case "row":
		return FlexDirectionRow
	case "column":
		return FlexDirectionColumn
	}
	return FlexDirectionRow
}

// parseJustifyContent парсит значение свойства justify-content в тип JustifyContent.
func parseJustifyContent(val string) JustifyContent {
	switch val {
	case "flex-start":
		return JustifyFlexStart
	case "center":
		return JustifyCenter
	case "flex-end":
		return JustifyFlexEnd
	case "space-between":
		return JustifySpaceBetween
	case "space-around":
		return JustifySpaceAround
	}
	return JustifyFlexStart
}

// parseAlignItems парсит значение свойства align-items в тип AlignItems.
func parseAlignItems(val string) AlignItems {
	switch val {
	case "flex-start":
		return AlignFlexStart
	case "center":
		return AlignCenter
	case "flex-end":
		return AlignFlexEnd
	case "stretch":
		return AlignStretch
	}
	return AlignStretch
}

// parseLengthList парсит список длин, поддерживая синтаксис repeat().
const maxGridTracks = 256

func parseLengthList(val string) []Length {
	var result []Length
	val = strings.TrimSpace(val)

	for len(val) > 0 {
		val = strings.TrimSpace(val)
		if val == "" {
			break
		}

		// Handle repeat() syntax: repeat(3, 1fr)
		if len(result) >= maxGridTracks {
			return result
		}
		if strings.HasPrefix(val, "repeat(") {
			end := strings.Index(val, ")")
			if end < 0 {
				break
			}
			inner := val[7:end]
			comma := strings.Index(inner, ",")
			if comma > 0 {
				count, err := strconv.Atoi(strings.TrimSpace(inner[:comma]))
				item := strings.TrimSpace(inner[comma+1:])
				if err == nil && count > 0 {
					count = min(count, maxGridTracks-len(result))
					for i := 0; i < count; i++ {
						result = append(result, parseLength(item))
					}
				}
			}
			val = val[end+1:]
			continue
		}

		// Find next separator (space or comma)
		sep := -1
		for i := 0; i < len(val); i++ {
			if val[i] == ' ' || val[i] == ',' {
				sep = i
				break
			}
		}
		if sep < 0 {
			result = append(result, parseLength(val))
			break
		}
		part := strings.TrimSpace(val[:sep])
		if part != "" {
			result = append(result, parseLength(part))
		}
		val = val[sep+1:]
	}

	if len(result) > maxGridTracks {
		result = result[:maxGridTracks]
	}
	return result
}

// isNumber проверяет, является ли строка валидным числом.
func isNumber(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// hexToInt парсит HEX-строку в целое число.
func hexToInt(s string) int {
	i := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			i = i*16 + int(c-'0')
		} else if c >= 'a' && c <= 'f' {
			i = i*16 + int(c-'a'+10)
		} else if c >= 'A' && c <= 'F' {
			i = i*16 + int(c-'A'+10)
		}
	}
	return i
}
