// Package dom provides a simple DOM tree built from HTML.
package dom

import (
    "strings"

    "golang.org/x/net/html"
)

// NodeType represents the type of a DOM node.
type NodeType int

const (
    ElementNode          NodeType = 1
    TextNode             NodeType = 3
    CommentNode          NodeType = 8
    DocumentNode         NodeType = 9
    DocumentFragmentNode NodeType = 11
)

// Node is a node in the DOM tree.
type Node struct {
    Type       NodeType
    TagName    string
    Attributes map[string]string
    Children   []*Node
    Text       string
    Parent     *Node
}

// htmlVoidElements is the set of HTML5 void elements that have no end tag
// (must never be pushed onto the stack). The x/net/html tokenizer does not
// treat these as self-closing when written without trailing slash.
var htmlVoidElements = map[string]bool{
    "area":   true,
    "base":   true,
    "br":     true,
    "col":    true,
    "embed":  true,
    "hr":     true,
    "img":    true,
    "input":  true,
    "keygen": true,
    "link":   true,
    "meta":   true,
    "param":  true,
    "source": true,
    "track":  true,
    "wbr":    true,
}

// Parse parses an HTML string and returns the document root.
func Parse(htmlStr string) *Node {
    doc := &Node{Type: DocumentNode, Attributes: map[string]string{}, Children: []*Node{}}
    tokenizer := html.NewTokenizer(strings.NewReader(htmlStr))

    stack := []*Node{doc}

    for {
        tt := tokenizer.Next()
        switch tt {
        case html.ErrorToken:
            return doc

        case html.TextToken:
            handleTextToken(stack, tokenizer)

        case html.StartTagToken, html.SelfClosingTagToken:
            stack = handleStartTag(stack, tokenizer, tt)

        case html.EndTagToken:
            if len(stack) > 1 {
                stack = stack[:len(stack)-1]
            }
        }
    }
}

func handleTextToken(stack []*Node, tokenizer *html.Tokenizer) {
    if len(stack) == 0 {
        return
    }
    parent := stack[len(stack)-1]
    text := string(tokenizer.Text())
    // Skip whitespace-only text nodes unless parent is a pre-like tag
    if strings.TrimSpace(text) == "" && parent.TagName != "pre" {
        return
    }
    parent.Children = append(parent.Children, &Node{
        Type:   TextNode,
        Text:   text,
        Parent: parent,
    })
}

func handleStartTag(stack []*Node, tokenizer *html.Tokenizer, tt html.TokenType) []*Node {
    if len(stack) == 0 {
        return stack
    }
    parent := stack[len(stack)-1]
    tag, hasAttr := tokenizer.TagName()
    tagStr := string(tag)
    node := &Node{
        Type:       ElementNode,
        TagName:    tagStr,
        Attributes: map[string]string{},
        Children:   []*Node{},
        Parent:     parent,
    }
    if hasAttr {
        for {
            key, val, moreAttr := tokenizer.TagAttr()
            node.Attributes[strings.ToLower(string(key))] = string(val)
            if !moreAttr {
                break
            }
        }
    }
    parent.Children = append(parent.Children, node)
    // Push onto stack only for non-void start tags.
    // (x/net/html tokenizer emits StartTagToken for HTML5 void elements
    //  written without trailing slash, e.g. <meta>, <link>.)
    if tt == html.StartTagToken && !htmlVoidElements[tagStr] {
        stack = append(stack, node)
    }
    return stack
}

// QueryAll returns all descendant nodes matching the selector.
func (n *Node) QueryAll(selector string) []*Node {
    return QueryAllSelector(n, selector)
}

// QuerySelector returns the first descendant node matching the selector, or nil.
func (n *Node) QuerySelector(selector string) *Node {
    return QuerySelectorSelector(n, selector)
}

// Matches checks if a node matches the CSS selector.
func (n *Node) Matches(selector string) bool {
    return MatchesSelector(n, selector)
}

// Closest returns the closest ancestor (including self) matching the selector, or nil.
func (n *Node) Closest(selector string) *Node {
    if n == nil {
        return nil
    }
    for cur := n; cur != nil; cur = cur.Parent {
        if cur.Matches(selector) {
            return cur
        }
    }
    return nil
}

func hasClass(classAttr, className string) bool {
    classes := strings.Fields(classAttr)
    for _, c := range classes {
        if c == className {
            return true
        }
    }
    return false
}

// GetAttr returns an attribute value, or empty string if not present.
func (n *Node) GetAttr(name string) string {
    if n.Attributes == nil {
        return ""
    }
    return n.Attributes[strings.ToLower(name)]
}

// RemoveChild removes a child node from this node's children.
func (n *Node) RemoveChild(child *Node) *Node {
    for i, c := range n.Children {
        if c == child {
            n.Children = append(n.Children[:i], n.Children[i+1:]...)
            child.Parent = nil
            return child
        }
    }
    return nil
}

// AppendChild adds a child node to this node's children.
func (n *Node) AppendChild(child *Node) *Node {
    n.Children = append(n.Children, child)
    child.Parent = n
    return child
}

// Clone creates a deep or shallow copy of the node.
func (n *Node) Clone(deep bool) *Node {
    clone := &Node{
        Type:       n.Type,
        TagName:    n.TagName,
        Text:       n.Text,
        Attributes: make(map[string]string, len(n.Attributes)),
    }
    for k, v := range n.Attributes {
        clone.Attributes[k] = v
    }
    if deep {
        for _, child := range n.Children {
            childClone := child.Clone(true)
            childClone.Parent = clone
            clone.Children = append(clone.Children, childClone)
        }
    }
    return clone
}

// OuterHTML returns the HTML serialization of the node.
func (n *Node) OuterHTML() string {
    if n == nil {
        return ""
    }
    if n.Type == TextNode || n.Type == CommentNode {
        return escapeHTML(n.Text)
    }
    if n.Type == DocumentFragmentNode || n.Type == DocumentNode {
        var b strings.Builder
        for _, child := range n.Children {
            b.WriteString(child.OuterHTML())
        }
        return b.String()
    }
    var b strings.Builder
    b.WriteString("<")
    b.WriteString(n.TagName)
    for k, v := range n.Attributes {
        b.WriteString(" ")
        b.WriteString(k)
        b.WriteString(`="`)
        b.WriteString(escapeHTML(v))
        b.WriteString(`"`)
    }
    if len(n.Children) == 0 && n.Text == "" {
        // Self-closing tags
        if isVoidElement(n.TagName) {
            b.WriteString(" />")
            return b.String()
        }
    }
    b.WriteString(">")
    if n.Text != "" {
        b.WriteString(escapeHTML(n.Text))
    }
    for _, child := range n.Children {
        b.WriteString(child.OuterHTML())
    }
    b.WriteString("</")
    b.WriteString(n.TagName)
    b.WriteString(">")
    return b.String()
}

func escapeHTML(s string) string {
    s = strings.ReplaceAll(s, "&", "&amp;")
    s = strings.ReplaceAll(s, "<", "&lt;")
    s = strings.ReplaceAll(s, ">", "&gt;")
    s = strings.ReplaceAll(s, `"`, "&quot;")
    return s
}

func isVoidElement(tag string) bool {
    switch strings.ToLower(tag) {
    case "area", "base", "br", "col", "embed", "hr", "img", "input",
        "link", "meta", "param", "source", "track", "wbr":
        return true
    }
    return false
}

// FindStyleTags returns all <style> elements in the document.
func (n *Node) FindStyleTags() []*Node {
    return n.QueryAll("style")
}

// FindLinkStylesheets returns all <link rel="stylesheet"> elements.
func (n *Node) FindLinkStylesheets() []*Node {
    return n.QueryAll("link[rel=stylesheet]")
}

// TextContent returns the concatenated text content of all descendant text nodes.
func (n *Node) TextContent() string {
    if n.Type == TextNode {
        return n.Text
    }
    var b strings.Builder
    for _, child := range n.Children {
        b.WriteString(child.TextContent())
    }
    return b.String()
}
