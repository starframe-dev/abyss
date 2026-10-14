package browser

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
	"github.com/rivo/uniseg"
)

type ansiSegment struct {
	style []string
	text  string
}

// splitANSI keeps only renderer-compatible 24-bit foreground/background SGR
// sequences. All other escape sequences and terminal controls are discarded.
func splitANSI(s string) []ansiSegment {
	// Raw C1 control bytes are invalid UTF-8. Strip them before rune decoding,
	// including their payloads, so CSI parameters cannot leak as visible text.
	s = stripRawC1Controls(s)
	s = strings.ToValidUTF8(s, "")
	var segs []ansiSegment
	var style []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() == 0 {
			return
		}
		segs = append(segs, ansiSegment{style: append([]string(nil), style...), text: cur.String()})
		cur.Reset()
	}
	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == 0x1b {
			flush()
			if i+1 >= len(runes) {
				break
			}
			switch runes[i+1] {
			case '[':
				j := i + 2
				for j < len(runes) && runes[j] >= 0x30 && runes[j] <= 0x3f {
					j++
				}
				if j == len(runes) {
					break
				}
				if runes[j] == 'm' && validRendererSGR(string(runes[i+2:j])) {
					params := string(runes[i+2 : j])
					if params == "0" || params == "" {
						style = nil
					} else {
						style = append(style, "\x1b["+params+"m")
					}
				}
				i = j
			case ']':
				i += 2
				for i < len(runes) && runes[i] != 0x07 && runes[i] != 0x9c {
					if runes[i] == 0x1b && i+1 < len(runes) && runes[i+1] == '\\' {
						i++
						break
					}
					i++
				}
			case 'P', 'X', '^', '_':
				i += 2
				for i < len(runes) && runes[i] != 0x9c {
					if runes[i] == 0x1b && i+1 < len(runes) && runes[i+1] == '\\' {
						i++
						break
					}
					i++
				}
			default:
				i++
			}
			continue
		}
		if r == 0x9b {
			flush()
			j := i + 1
			for j < len(runes) && !(runes[j] >= 0x40 && runes[j] <= 0x7e) {
				j++
			}
			i = j
			continue
		}
		if r == 0x90 || r == 0x98 || r == 0x9d || r == 0x9e || r == 0x9f {
			flush()
			for i+1 < len(runes) && runes[i+1] != 0x9c {
				i++
			}
			if i+1 < len(runes) {
				i++
			}
			continue
		}
		if unicode.IsControl(r) || (unicode.Is(unicode.Cf, r) && r != 0x200c && r != 0x200d) {
			continue
		}
		cur.WriteRune(r)
	}
	flush()
	return segs
}

func stripRawC1Controls(s string) string {
	var out strings.Builder
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if size != 1 || r != utf8.RuneError {
			out.WriteString(s[i : i+size])
			i += size
			continue
		}
		switch s[i] {
		case 0x9b:
			i++
			for i < len(s) && !(s[i] >= 0x40 && s[i] <= 0x7e) {
				i++
			}
			if i < len(s) {
				i++
			}
		case 0x90, 0x98, 0x9d, 0x9e, 0x9f:
			kind := s[i]
			i++
			for i < len(s) {
				if s[i] == 0x9c || (kind == 0x9d && s[i] == 0x07) {
					i++
					break
				}
				if s[i] == 0x1b && i+1 < len(s) && s[i+1] == '\\' {
					i += 2
					break
				}
				i++
			}
		default:
			i++
		}
	}
	return out.String()
}

func validRendererSGR(params string) bool {
	if params == "" || params == "0" {
		return true
	}
	p := strings.Split(params, ";")
	for i := 0; i < len(p); {
		if (p[i] != "38" && p[i] != "48") || i+4 >= len(p) || p[i+1] != "2" {
			return false
		}
		for j := i + 2; j <= i+4; j++ {
			v, err := strconv.Atoi(p[j])
			if err != nil || v < 0 || v > 255 {
				return false
			}
		}
		i += 5
	}
	return true
}

func visibleWidth(line string) int {
	width := 0
	for _, seg := range splitANSI(line) {
		width += runewidth.StringWidth(seg.text)
	}
	return width
}

func clipANSI(line string, start, width int) string {
	if width <= 0 {
		return ""
	}
	start = max(0, start)
	end := start + width
	var out strings.Builder
	column := 0
	previousIncluded := false
	for _, seg := range splitANSI(line) {
		var selected strings.Builder
		graphemes := uniseg.NewGraphemes(seg.text)
		for graphemes.Next() {
			cluster, cellWidth := graphemes.Str(), graphemes.Width()
			if cellWidth <= 0 {
				if previousIncluded {
					selected.WriteString(cluster)
				}
				continue
			}
			clusterEnd := column + cellWidth
			overlapStart, overlapEnd := max(start, column), min(end, clusterEnd)
			if overlapStart < overlapEnd {
				if overlapStart == column && overlapEnd == clusterEnd {
					selected.WriteString(cluster)
					previousIncluded = true
				} else {
					selected.WriteString(strings.Repeat(" ", overlapEnd-overlapStart))
					previousIncluded = false
				}
			} else {
				previousIncluded = false
			}
			column = clusterEnd
			if column >= end {
				break
			}
		}
		if selected.Len() > 0 {
			if len(seg.style) > 0 {
				out.WriteString(strings.Join(seg.style, ""))
			}
			out.WriteString(selected.String())
			if len(seg.style) > 0 {
				out.WriteString("\x1b[0m")
			}
		}
		if column >= end {
			break
		}
	}
	return out.String()
}
