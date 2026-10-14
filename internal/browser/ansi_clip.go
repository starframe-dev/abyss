package browser

import (
    "strings"
    "unicode"
)

// ansiSegment is a run of printable cells that share one SGR state.
type ansiSegment struct {
    style []string
    text  string
}

// splitANSI breaks an ANSI string into printable segments, attaching the SGR
// sequences that are active for each segment. Unhandled control sequences are
// consumed without emitting cells so they cannot corrupt the clipped output.
func splitANSI(s string) []ansiSegment {
    var segs []ansiSegment
    var style []string
    var cur strings.Builder

    flush := func() {
        if cur.Len() == 0 {
            return
        }
        st := make([]string, len(style))
        copy(st, style)
        segs = append(segs, ansiSegment{style: st, text: cur.String()})
        cur.Reset()
    }

    runes := []rune(s)
    for i := 0; i < len(runes); i++ {
        if runes[i] != 0x1b || i+1 >= len(runes) || runes[i+1] != '[' {
            cur.WriteRune(runes[i])
            continue
        }
        j := i + 2
        for j < len(runes) && !unicode.IsLetter(runes[j]) && runes[j] != '~' {
            j++
        }
        if j >= len(runes) {
            // Dangling introducer: drop it so it cannot leak into the output.
            break
        }
        params := string(runes[i+2 : j])
        final := runes[j]
        flush()
        if final == 'm' {
            if params == "" || params == "0" {
                style = nil
            } else {
                style = append(style, "\x1b["+params+"m")
            }
        }
        i = j
    }
    flush()
    return segs
}

// visibleWidth returns the number of printable columns an ANSI line occupies,
// ignoring escape sequences.
func visibleWidth(line string) int {
    n := 0
    for _, seg := range splitANSI(line) {
        n += len([]rune(seg.text))
    }
    return n
}

// clipANSI returns the printable columns [start, start+width) of an ANSI line,
// re-emitting complete SGR sequences for every kept segment.
//
// Slicing the raw runes (the previous approach) cut escape codes in half, so a
// truecolor image line - where each cell costs ~20 runes - was truncated after
// a handful of columns and left a dangling ESC that terminals interpreted as
// erase-to-end-of-line, swallowing the rest of the picture.
func clipANSI(line string, start, width int) string {
    if width <= 0 {
        return ""
    }
    end := start + width
    var b strings.Builder
    col := 0
    for _, seg := range splitANSI(line) {
        runes := []rune(seg.text)
        segEnd := col + len(runes)
        if segEnd > start && col < end {
            lo, hi := max(0, start-col), min(len(runes), end-col)
            if len(seg.style) > 0 {
                b.WriteString(strings.Join(seg.style, ""))
                b.WriteString(string(runes[lo:hi]))
                b.WriteString("\x1b[0m")
            } else {
                b.WriteString(string(runes[lo:hi]))
            }
        }
        col = segEnd
        if col >= end {
            break
        }
    }
    return b.String()
}
