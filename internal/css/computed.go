// Computed style computation: cascade, specificity, and value parsing.
package css

import (
    "strings"
    "sync"

    "github.com/starframe/abyss/internal/dom"
)

// cached is a global cache of computed styles keyed by document.
var cached struct {
    sync.Mutex
    doc    *dom.Node
    styles map[*dom.Node]ComputedStyle
}

// Compute computes computed styles for all nodes in the document.
// Results are cached and reused if the document hasn't changed.
func Compute(doc *dom.Node) map[*dom.Node]ComputedStyle {
    return computeWithExtra(doc, nil)
}

// ComputeWithExternal computes styles with additional external CSS strings.
func ComputeWithExternal(doc *dom.Node, externalCSS []string) map[*dom.Node]ComputedStyle {
    return computeWithExtra(doc, externalCSS)
}

func computeWithExtra(doc *dom.Node, externalCSS []string) map[*dom.Node]ComputedStyle {
    cached.Lock()
    defer cached.Unlock()

    // Return cached result if document is the same and no external CSS
    if cached.doc == doc && cached.styles != nil && len(externalCSS) == 0 {
        return cached.styles
    }

    styles := make(map[*dom.Node]ComputedStyle)

    // Collect all rules from <style> tags
    var allRules []Rule
    for _, styleNode := range doc.FindStyleTags() {
        cssText := styleNode.TextContent()
        rules := ParseStyleSheet(cssText)
        allRules = append(allRules, rules...)
    }

    // Add rules from external CSS
    for _, cssText := range externalCSS {
        rules := ParseStyleSheet(cssText)
        allRules = append(allRules, rules...)
    }

    // Compute styles recursively
    computeNode(doc, allRules, styles, nil)

    // Cache for next call
    cached.doc = doc
    cached.styles = styles

    return styles
}

// InvalidateCache clears the computed style cache.
// Should be called when DOM is modified via JS.
func InvalidateCache() {
    cached.Lock()
    defer cached.Unlock()
    cached.doc = nil
    cached.styles = nil
}

func computeNode(node *dom.Node, rules []Rule, styles map[*dom.Node]ComputedStyle, inheritedVars map[string]string) {
    style := Default()
    variables := cloneVariables(inheritedVars)

    // Apply matching rules from stylesheet (sorted by specificity)
    type match struct {
        decls       []Declaration
        specificity int
    }
    var matches []match

    for _, rule := range rules {
        for _, sel := range rule.Selectors {
            if nodeMatchesSelector(node, sel) {
                matches = append(matches, match{
                    decls:       rule.Declarations,
                    specificity: Specificity(sel),
                })
            }
        }
    }

    // Sort by specificity (lower first, so higher override)
    for i := 0; i < len(matches)-1; i++ {
        for j := i + 1; j < len(matches); j++ {
            if matches[i].specificity > matches[j].specificity {
                matches[i], matches[j] = matches[j], matches[i]
            }
        }
    }

    // Apply stylesheet rules
    for _, m := range matches {
        for _, decl := range m.decls {
            if isCustomProperty(decl.Property) {
                variables[strings.ToLower(strings.TrimSpace(decl.Property))] = resolveVariables(decl.Value, variables)
                continue
            }
            applyDeclaration(&style, decl.Property, resolveVariables(decl.Value, variables))
        }
    }

    // Apply inline styles (highest specificity)
    if inline := node.GetAttr("style"); inline != "" {
        decls := ParseInline(inline)
        for _, decl := range decls {
            if isCustomProperty(decl.Property) {
                variables[strings.ToLower(strings.TrimSpace(decl.Property))] = resolveVariables(decl.Value, variables)
                continue
            }
            applyDeclaration(&style, decl.Property, resolveVariables(decl.Value, variables))
        }
    }

    styles[node] = style

    // Recurse to children
    for _, child := range node.Children {
        computeNode(child, rules, styles, variables)
    }
}

func cloneVariables(source map[string]string) map[string]string {
    result := make(map[string]string, len(source))
    for name, value := range source {
        result[name] = value
    }
    return result
}

func isCustomProperty(property string) bool {
    return strings.HasPrefix(strings.TrimSpace(strings.ToLower(property)), "--")
}

func resolveVariables(value string, variables map[string]string) string {
    for {
        start := strings.Index(value, "var(")
        if start < 0 {
            return value
        }
        depth := 0
        end := -1
        for i := start + len("var("); i < len(value); i++ {
            switch value[i] {
            case '(':
                depth++
            case ')':
                if depth == 0 {
                    end = i
                } else {
                    depth--
                }
            }
            if end >= 0 {
                break
            }
        }
        if end < 0 {
            return value
        }
        contents := value[start+len("var("):end]
        parts := strings.SplitN(contents, ",", 2)
        name := strings.ToLower(strings.TrimSpace(parts[0]))
        replacement, ok := variables[name]
        if !ok && len(parts) == 2 {
            replacement = strings.TrimSpace(parts[1])
            ok = true
        }
        if !ok {
            return value[:start] + value[end+1:]
        }
        value = value[:start] + resolveVariables(replacement, variables) + value[end+1:]
    }
}

func nodeMatchesSelector(node *dom.Node, selector string) bool {
    if node == nil || node.Type != dom.ElementNode {
        return false
    }

    selector = strings.TrimSpace(selector)
    if selector == ":root" {
        return node.Parent != nil && node.Parent.Type == dom.DocumentNode
    }
    if selector == "" {
        return true
    }
    return node.Matches(selector)
}

func applyDeclaration(style *ComputedStyle, prop, val string) {
    prop = strings.ToLower(strings.TrimSpace(prop))
    val = strings.TrimSpace(val)

    switch {
    case prop == "margin" || strings.HasPrefix(prop, "margin-"),
        prop == "padding" || strings.HasPrefix(prop, "padding-"):
        applyBoxModel(style, prop, val)
    case prop == "border" || strings.HasPrefix(prop, "border-"):
        applyBorder(style, prop, val)
    case prop == "width", prop == "height", prop == "min-width", prop == "max-width",
        prop == "gap", prop == "grid-template-columns", prop == "grid-template-rows":
        applyDimensions(style, prop, val)
    case prop == "top", prop == "left", prop == "right", prop == "bottom":
        applyOffset(style, prop, val)
    case prop == "color" || prop == "background-color":
        applyColor(style, prop, val)
    default:
        applyLayout(style, prop, val)
    }
}

// applyBoxModel sets margin and padding properties.
func applyBoxModel(style *ComputedStyle, prop, val string) {
    switch prop {
    case "margin":
        style.Margin = parseBoxModel(val)
    case "margin-top":
        style.Margin.Top = parseLength(val)
    case "margin-right":
        style.Margin.Right = parseLength(val)
    case "margin-bottom":
        style.Margin.Bottom = parseLength(val)
    case "margin-left":
        style.Margin.Left = parseLength(val)
    case "padding":
        style.Padding = parseBoxModel(val)
    case "padding-top":
        style.Padding.Top = parseLength(val)
    case "padding-right":
        style.Padding.Right = parseLength(val)
    case "padding-bottom":
        style.Padding.Bottom = parseLength(val)
    case "padding-left":
        style.Padding.Left = parseLength(val)
    }
}

// applyBorder sets border edge, width, color, style and radius properties.
func applyBorder(style *ComputedStyle, prop, val string) {
    switch prop {
    case "border":
        b := parseBorder(val)
        style.BorderTop = b
        style.BorderRight = b
        style.BorderBottom = b
        style.BorderLeft = b
    case "border-top":
        style.BorderTop = parseBorder(val)
    case "border-right":
        style.BorderRight = parseBorder(val)
    case "border-bottom":
        style.BorderBottom = parseBorder(val)
    case "border-left":
        style.BorderLeft = parseBorder(val)
    case "border-width":
        w := parseBorderWidth(val)
        style.BorderTop.Width = w
        style.BorderRight.Width = w
        style.BorderBottom.Width = w
        style.BorderLeft.Width = w
    case "border-color":
        c := parseColor(val)
        style.BorderTop.Color = c
        style.BorderRight.Color = c
        style.BorderBottom.Color = c
        style.BorderLeft.Color = c
    case "border-style":
        s := parseBorderStyle(val)
        style.BorderTop.Style = s
        style.BorderRight.Style = s
        style.BorderBottom.Style = s
        style.BorderLeft.Style = s
    case "border-radius":
        r := parseLength(val)
        style.BorderRadiusTopLeft = r
        style.BorderRadiusTopRight = r
        style.BorderRadiusBottomRight = r
        style.BorderRadiusBottomLeft = r
    case "border-top-left-radius":
        style.BorderRadiusTopLeft = parseLength(val)
    case "border-top-right-radius":
        style.BorderRadiusTopRight = parseLength(val)
    case "border-bottom-right-radius":
        style.BorderRadiusBottomRight = parseLength(val)
    case "border-bottom-left-radius":
        style.BorderRadiusBottomLeft = parseLength(val)
    }
}

// applyDimensions sets width/height/gap and grid-template properties.
func applyDimensions(style *ComputedStyle, prop, val string) {
    switch prop {
    case "width":
        style.Width = parseLength(val)
    case "height":
        style.Height = parseLength(val)
    case "min-width":
        style.MinWidth = parseLength(val)
    case "max-width":
        style.MaxWidth = parseLength(val)
    case "gap":
        style.Gap = parseLength(val)
    case "grid-template-columns":
        style.GridTemplateColumns = parseLengthList(val)
    case "grid-template-rows":
        style.GridTemplateRows = parseLengthList(val)
    }
}

// applyOffset sets the top/left/right/bottom offset properties.
func applyOffset(style *ComputedStyle, prop, val string) {
    switch prop {
    case "top":
        style.Top = parseLength(val)
    case "left":
        style.Left = parseLength(val)
    case "right":
        style.Right = parseLength(val)
    case "bottom":
        style.Bottom = parseLength(val)
    }
}

// applyColor sets foreground and background colors.
func applyColor(style *ComputedStyle, prop, val string) {
    switch prop {
    case "color":
        style.Color = parseColor(val)
    case "background-color":
        style.BackgroundColor = parseColor(val)
    }
}

// applyLayout sets the remaining layout properties.
func applyLayout(style *ComputedStyle, prop, val string) {
    switch prop {
    case "display":
        style.Display = parseDisplay(val)
    case "position":
        style.Position = parsePosition(val)
    case "overflow":
        style.Overflow = parseOverflow(val)
    case "z-index":
        style.ZIndex = parseInt(val)
    case "flex-direction":
        style.FlexDirection = parseFlexDirection(val)
    case "justify-content":
        style.JustifyContent = parseJustifyContent(val)
    case "align-items":
        style.AlignItems = parseAlignItems(val)
    case "flex-grow":
        style.FlexGrow = parseFloat(val)
    case "flex-shrink":
        style.FlexShrink = parseFloat(val)
    }
}
