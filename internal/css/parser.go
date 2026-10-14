// CSS parsing utilities.
package css

import (
    "strings"
    "unicode"
)

// Declaration is a single CSS property-value pair.
type Declaration struct {
    Property string
    Value    string
}

// Rule is a CSS rule with selectors and declarations.
type Rule struct {
    Selectors     []string
    Declarations  []Declaration
}

// ParseInline parses an inline style attribute string.
func ParseInline(style string) []Declaration {
    var decls []Declaration
    parts := splitBySemicolon(style)
    for _, part := range parts {
        part = strings.TrimSpace(part)
        if part == "" {
            continue
        }
        colon := strings.Index(part, ":")
        if colon < 0 {
            continue
        }
        prop := strings.TrimSpace(part[:colon])
        val := strings.TrimSpace(part[colon+1:])
        decls = append(decls, Declaration{Property: prop, Value: val})
    }
    return decls
}

// ParseStyleSheet parses a CSS stylesheet string into rules.
func ParseStyleSheet(css string) []Rule {
    return parseRules(stripComments(css))
}

func parseRules(css string) []Rule {
    var rules []Rule
    for len(css) > 0 {
        brace := strings.Index(css, "{")
        if brace < 0 {
            break
        }

        endBrace := findClosingBrace(css, brace)
        if endBrace < 0 {
            break
        }

        prelude := strings.TrimSpace(css[:brace])
        if strings.HasPrefix(prelude, "@") {
            if isNestedRuleAtRule(prelude) {
                rules = append(rules, parseRules(css[brace+1:endBrace])...)
            }
        } else {
            selectors := strings.Split(prelude, ",")
            for i := range selectors {
                selectors[i] = strings.TrimSpace(selectors[i])
            }
            rules = append(rules, Rule{
                Selectors:    selectors,
                Declarations: parseDeclarations(css[brace+1 : endBrace]),
            })
        }

        css = css[endBrace+1:]
    }
    return rules
}

func isNestedRuleAtRule(prelude string) bool {
    name := strings.ToLower(strings.TrimSpace(strings.SplitN(prelude, " ", 2)[0]))
    switch name {
    case "@media", "@supports", "@container", "@layer", "@document":
        return true
    default:
        return false
    }
}

func stripComments(css string) string {
    var result strings.Builder
    for i := 0; i < len(css); i++ {
        if i+1 < len(css) && css[i] == '/' && css[i+1] == '*' {
            // Skip comment
            i += 2
            for i+1 < len(css) && !(css[i] == '*' && css[i+1] == '/') {
                i++
            }
            i++ // skip */
            continue
        }
        result.WriteByte(css[i])
    }
    return result.String()
}

func findClosingBrace(s string, start int) int {
    depth := 0
    for i := start; i < len(s); i++ {
        if s[i] == '{' {
            depth++
        } else if s[i] == '}' {
            depth--
            if depth == 0 {
                return i
            }
        }
    }
    return -1
}

func parseDeclarations(block string) []Declaration {
    var decls []Declaration
    parts := splitBySemicolon(block)
    for _, part := range parts {
        part = strings.TrimSpace(part)
        if part == "" {
            continue
        }
        colon := strings.Index(part, ":")
        if colon < 0 {
            continue
        }
        prop := strings.TrimSpace(part[:colon])
        val := strings.TrimSpace(part[colon+1:])
        decls = append(decls, Declaration{Property: prop, Value: val})
    }
    return decls
}

func splitBySemicolon(s string) []string {
    var parts []string
    var current strings.Builder
    parenDepth := 0
    for i := 0; i < len(s); i++ {
        ch := s[i]
        if ch == '(' {
            parenDepth++
        } else if ch == ')' {
            parenDepth--
        } else if ch == ';' && parenDepth == 0 {
            parts = append(parts, current.String())
            current.Reset()
            continue
        }
        current.WriteByte(ch)
    }
    if current.Len() > 0 {
        parts = append(parts, current.String())
    }
    return parts
}

// Specificity computes CSS selector specificity as (a, b, c) where:
// a = number of ID selectors
// b = number of class selectors + attribute selectors + pseudo-classes
// c = number of type selectors + pseudo-elements
func Specificity(selector string) int {
    a, b, c := 0, 0, 0
    selector = strings.TrimSpace(selector)

    // Simple parser: split by combinators
    tokens := tokenizeSelector(selector)
    for _, tok := range tokens {
        tok = strings.TrimSpace(tok)
        if tok == "" || tok == ">" || tok == "+" || tok == "~" {
            continue
        }
        // Handle compound selectors like div.class#id
        a += strings.Count(tok, "#")
        b += strings.Count(tok, ".")
        // Count pseudo-classes (not pseudo-elements)
        for i := 0; i < len(tok); i++ {
            if tok[i] == ':' && i+1 < len(tok) && tok[i+1] != ':' {
                b++
            }
            if tok[i] == '[' {
                b++
            }
        }
        // Type selector: first char is a letter and not preceded by # or .
        // Check each segment
        parts := splitByClassID(tok)
        for _, part := range parts {
            part = strings.TrimSpace(part)
            if part == "" {
                continue
            }
            if strings.HasPrefix(part, "#") || strings.HasPrefix(part, ".") || strings.HasPrefix(part, "[") {
                continue
            }
            if strings.HasPrefix(part, ":") {
                continue
            }
            if isValidTypeSelector(part) {
                c++
            }
        }
    }

    return a*100 + b*10 + c
}

func splitByClassID(s string) []string {
    var parts []string
    var current strings.Builder
    for i := 0; i < len(s); i++ {
        ch := s[i]
        if ch == '.' || ch == '#' {
            if current.Len() > 0 {
                parts = append(parts, current.String())
                current.Reset()
            }
            current.WriteByte(ch)
            continue
        }
        current.WriteByte(ch)
    }
    if current.Len() > 0 {
        parts = append(parts, current.String())
    }
    return parts
}

func tokenizeSelector(selector string) []string {
    var tokens []string
    var current strings.Builder
    inBrackets := false
    for i := 0; i < len(selector); i++ {
        ch := selector[i]
        if ch == '[' {
            inBrackets = true
            current.WriteByte(ch)
            continue
        }
        if ch == ']' {
            inBrackets = false
            current.WriteByte(ch)
            continue
        }
        if inBrackets {
            current.WriteByte(ch)
            continue
        }
        if ch == '>' || ch == '+' || ch == '~' {
            if current.Len() > 0 {
                tokens = append(tokens, current.String())
                current.Reset()
            }
            tokens = append(tokens, string(ch))
            continue
        }
        if ch == ' ' {
            if current.Len() > 0 {
                tokens = append(tokens, current.String())
                current.Reset()
            }
            continue
        }
        current.WriteByte(ch)
    }
    if current.Len() > 0 {
        tokens = append(tokens, current.String())
    }
    return tokens
}

func isValidTypeSelector(s string) bool {
    if s == "" {
        return false
    }
    for _, r := range s {
        if !unicode.IsLetter(r) && r != '-' && r != '_' {
            return false
        }
    }
    return true
}
