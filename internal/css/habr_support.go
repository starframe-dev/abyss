package css

import (
	"strings"
)

// ParseStylesheet parses rules in a stylesheet, including rules nested in
// @media and @supports blocks. Unsupported at-rules are ignored.
func ParseStylesheet(source string) []Rule {
	return parseStylesheetBlock(source, "", "")
}

func parseStylesheetBlock(source, media, supports string) []Rule {
	var rules []Rule
	for i := 0; i < len(source); {
		for i < len(source) && (source[i] == ' ' || source[i] == '\n' || source[i] == '\r' || source[i] == '\t' || source[i] == ';') { i++ }
		if i >= len(source) { break }
		start := i
		for i < len(source) && source[i] != '{' && source[i] != '}' { i++ }
		if i >= len(source) || source[i] == '}' { break }
		header := strings.TrimSpace(source[start:i])
		end := matchingBrace(source, i)
		if end < 0 { break }
		body := source[i+1:end]
		i = end+1
		if strings.HasPrefix(header, "@") {
			lower := strings.ToLower(header)
			switch {
			case strings.HasPrefix(lower, "@media"):
				query := strings.TrimSpace(header[len("@media"):]); if media != "" { query = media + " and " + query }; rules = append(rules, parseStylesheetBlock(body, query, supports)...)
			case strings.HasPrefix(lower, "@supports"):
				query := strings.TrimSpace(header[len("@supports"):]); if supports != "" { query = supports + " and " + query }; rules = append(rules, parseStylesheetBlock(body, media, query)...)
			}
			continue
		}
		decls := ParseInline(body)
		if len(decls) == 0 { continue }
		selectors := strings.Split(header, ",")
		for j := range selectors { selectors[j] = strings.TrimSpace(selectors[j]) }
		rules = append(rules, Rule{Selectors: selectors, Declarations: decls, MediaQuery: media, SupportsQuery: supports})
	}
	return rules
}

func matchingBrace(s string, open int) int {
	depth := 0
	quote := byte(0)
	for i := open; i < len(s); i++ {
		c := s[i]
		if quote != 0 { if c == quote && (i == 0 || s[i-1] != '\\') { quote = 0 }; continue }
		if c == '\'' || c == '"' { quote = c; continue }
		if c == '{' { depth++ } else if c == '}' { depth--; if depth == 0 { return i } }
	}
	return -1
}

// ResolveCustomProperties substitutes known var() references; unresolved
// variables leave the original value intact so callers can safely ignore it.
func ResolveCustomProperties(value string, properties map[string]string) (string, bool) {
	for n := 0; n < 32; n++ {
		start := strings.Index(value, "var(")
		if start < 0 { return value, true }
		end := strings.Index(value[start+4:], ")")
		if end < 0 { return value, false }
		end += start+4
		name := strings.TrimSpace(value[start+4:end])
		if comma := strings.Index(name, ","); comma >= 0 {
			fallback := strings.TrimSpace(name[comma+1:]); name = strings.TrimSpace(name[:comma])
			if replacement, ok := properties[name]; ok { value = value[:start]+replacement+value[end+1:] } else { value = value[:start]+fallback+value[end+1:] }
		} else {
			replacement, ok := properties[name]; if !ok { return value, false }; value = value[:start]+replacement+value[end+1:]
		}
	}
	return value, !strings.Contains(value, "var(")
}
