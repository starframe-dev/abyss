//go:build legacy

package dom

import (
    "testing"
)

func TestMatchesComplex(t *testing.T) {
    tests := []struct {
        name      string
        node      *Node
        sel       ComplexSelector
        wantMatch bool
    }{
        {
            name:      "simple tag match",
            node:      createDivElement("div", map[string]string{"class": "foo"}),
            sel:       ComplexSelector{Parts: []SelectorPart{{Tag: "div"}}},
            wantMatch: true,
        },
        {
            name:      "simple tag no match",
            node:      createDivElement("div", map[string]string{"class": "foo"}),
            sel:       ComplexSelector{Parts: []SelectorPart{{Tag: "span"}}},
            wantMatch: false,
        },
        {
            name: "complex selector descendant combinator",
            node: createDivElement("div", map[string]string{"class": "foo"}, []*Node{
                createDivElement("span", map[string]string{"class": "bar"}),
            }),
            sel: ComplexSelector{Parts: []SelectorPart{
                {Tag: "div"},
                {Combinator: " ", Tag: "span"},
            }},
            wantMatch: true,
        },
        {
            name: "complex selector child combinator",
            node: createDivElement("div", map[string]string{"class": "foo"}, []*Node{
                createDivElement("span", map[string]string{"class": "bar"}),
            }),
            sel: ComplexSelector{Parts: []SelectorPart{
                {Tag: "div"},
                {Combinator: ">", Tag: "span"},
            }},
            wantMatch: true,
        },
        {
            name: "complex selector adjacent sibling",
            node: createDivElement("div", map[string]string{"class": "foo"}, []*Node{
                createDivElement("span", map[string]string{"class": "bar"}),
            }),
            sel: ComplexSelector{Parts: []SelectorPart{
                {Tag: "div"},
                {Combinator: "+", Tag: "span"},
            }},
            wantMatch: true,
        },
        {
            name: "complex selector general sibling",
            node: createDivElement("div", map[string]string{"class": "foo"}, []*Node{
                createDivElement("span", map[string]string{"class": "bar"}),
            }),
            sel: ComplexSelector{Parts: []SelectorPart{
                {Tag: "div"},
                {Combinator: "~", Tag: "span"},
            }},
            wantMatch: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := matchesComplex(tt.node, tt.sel)
            if got != tt.wantMatch {
                t.Errorf("matchesComplex(%v, %v) = %v, want %v", tt.node, tt.sel, got, tt.wantMatch)
            }
        })
    }
}

func TestMatchPart(t *testing.T) {
    tests := []struct {
        name      string
        node      *Node
        parts     []SelectorPart
        idx       int
        wantMatch bool
    }{
        {
            name: "first part match",
            node: createDivElement("div", map[string]string{"class": "foo"}),
            parts: []SelectorPart{
                {Tag: "div"},
            },
            idx:       0,
            wantMatch: true,
        },
        {
            name: "first part no match",
            node: createDivElement("div", map[string]string{"class": "foo"}),
            parts: []SelectorPart{
                {Tag: "span"},
            },
            idx:       0,
            wantMatch: false,
        },
        {
            name: "descendant combinator match",
            node: createDivElement("div", map[string]string{"class": "foo"}, []*Node{
                createDivElement("span", map[string]string{"class": "bar"}),
            }),
            parts: []SelectorPart{
                {Tag: "div"},
                {Combinator: " ", Tag: "span"},
            },
            idx:       1,
            wantMatch: true,
        },
        {
            name: "descendant combinator no match",
            node: createDivElement("div", map[string]string{"class": "foo"}, []*Node{
                createDivElement("span", map[string]string{"class": "other"}),
            }),
            parts: []SelectorPart{
                {Tag: "div"},
                {Combinator: " ", Tag: "span.class"},
            },
            idx:       1,
            wantMatch: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := matchPart(tt.node, tt.parts, tt.idx)
            if got != tt.wantMatch {
                t.Errorf("matchPart(%v, %v, %d) = %v, want %v", tt.node, tt.parts, tt.idx, got, tt.wantMatch)
            }
        })
    }
}

func TestPreviousElementSibling(t *testing.T) {
    tests := []struct {
        name       string
        node       *Node
        wantResult *Node
    }{
        {
            name: "has previous sibling",
            node: createDivElement("div", map[string]string{"class": "foo"}, []*Node{
                createDivElement("span", map[string]string{"class": "bar"}),
            }),
            wantResult: createDivElement("span", map[string]string{"class": "bar"}),
        },
        {
            name: "no previous sibling",
            node: createDivElement("div", map[string]string{"class": "foo"}, []*Node{
                createDivElement("span", map[string]string{"class": "bar"}),
            }),
            wantResult: nil,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := previousElementSibling(tt.node)
            if (got != nil) != (tt.wantResult != nil) {
                t.Errorf("previousElementSibling(%v) = %v, want %v", tt.node, got, tt.wantResult)
            }
        })
    }
}

func TestMatchesSimple(t *testing.T) {
    tests := []struct {
        name      string
        node      *Node
        part      SelectorPart
        wantMatch bool
    }{
        {
            name:      "tag match",
            node:      createDivElement("div", map[string]string{"class": "foo"}),
            part:      SelectorPart{Tag: "div"},
            wantMatch: true,
        },
        {
            name:      "tag no match",
            node:      createDivElement("div", map[string]string{"class": "foo"}),
            part:      SelectorPart{Tag: "span"},
            wantMatch: false,
        },
        {
            name:      "id match",
            node:      createDivElement("div", map[string]string{"id": "myid", "class": "foo"}),
            part:      SelectorPart{ID: "myid"},
            wantMatch: true,
        },
        {
            name:      "id no match",
            node:      createDivElement("div", map[string]string{"id": "other", "class": "foo"}),
            part:      SelectorPart{ID: "myid"},
            wantMatch: false,
        },
        {
            name:      "class match",
            node:      createDivElement("div", map[string]string{"class": "foo bar"}),
            part:      SelectorPart{Classes: []string{"foo"}},
            wantMatch: true,
        },
        {
            name:      "class no match",
            node:      createDivElement("div", map[string]string{"class": "foo bar"}),
            part:      SelectorPart{Classes: []string{"baz"}},
            wantMatch: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := matchesSimple(tt.node, tt.part)
            if got != tt.wantMatch {
                t.Errorf("matchesSimple(%v, %v) = %v, want %v", tt.node, tt.part, got, tt.wantMatch)
            }
        })
    }
}

func TestMatchesAttr(t *testing.T) {
    tests := []struct {
        name      string
        node      *Node
        attr      AttrSelector
        wantMatch bool
    }{
        {
            name:      "attribute exists",
            node:      createDivElement("div", map[string]string{"data-foo": "bar"}),
            attr:      AttrSelector{Name: "data-foo"},
            wantMatch: true,
        },
        {
            name:      "attribute does not exist",
            node:      createDivElement("div", map[string]string{"class": "foo"}),
            attr:      AttrSelector{Name: "data-foo"},
            wantMatch: false,
        },
        {
            name:      "exact match",
            node:      createDivElement("div", map[string]string{"data-foo": "bar"}),
            attr:      AttrSelector{Name: "data-foo", Op: "=", Value: "bar"},
            wantMatch: true,
        },
        {
            name:      "exact no match",
            node:      createDivElement("div", map[string]string{"data-foo": "baz"}),
            attr:      AttrSelector{Name: "data-foo", Op: "=", Value: "bar"},
            wantMatch: false,
        },
        {
            name:      "space-separated values match",
            node:      createDivElement("div", map[string]string{"data-foo": "bar baz"}),
            attr:      AttrSelector{Name: "data-foo", Op: "~=", Value: "baz"},
            wantMatch: true,
        },
        {
            name:      "space-separated values no match",
            node:      createDivElement("div", map[string]string{"data-foo": "bar baz"}),
            attr:      AttrSelector{Name: "data-foo", Op: "~=", Value: "qux"},
            wantMatch: false,
        },
        {
            name:      "starts with match",
            node:      createDivElement("div", map[string]string{"data-foo": "bar"}),
            attr:      AttrSelector{Name: "data-foo", Op: "^=", Value: "b"},
            wantMatch: true,
        },
        {
            name:      "ends with match",
            node:      createDivElement("div", map[string]string{"data-foo": "baz"}),
            attr:      AttrSelector{Name: "data-foo", Op: "$=", Value: "z"},
            wantMatch: true,
        },
        {
            name:      "contains match",
            node:      createDivElement("div", map[string]string{"data-foo": "barbaz"}),
            attr:      AttrSelector{Name: "data-foo", Op: "*=", Value: "baz"},
            wantMatch: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := matchesAttr(tt.node, tt.attr)
            if got != tt.wantMatch {
                t.Errorf("matchesAttr(%v, %v) = %v, want %v", tt.node, tt.attr, got, tt.wantMatch)
            }
        })
    }
}

func TestMatchesPseudo(t *testing.T) {
    tests := []struct {
        name      string
        node      *Node
        pseudo    PseudoSelector
        wantMatch bool
    }{
        {
            name: "first-child",
            node: createDivElement("div", map[string]string{"class": "foo"}, []*Node{
                createDivElement("span", map[string]string{"class": "bar"}),
                createDivElement("span", map[string]string{"class": "baz"}),
            }),
            pseudo:    PseudoSelector{Name: "first-child"},
            wantMatch: true,
        },
        {
            name: "last-child",
            node: createDivElement("div", map[string]string{"class": "foo"}, []*Node{
                createDivElement("span", map[string]string{"class": "bar"}),
                createDivElement("span", map[string]string{"class": "baz"}),
            }),
            pseudo:    PseudoSelector{Name: "last-child"},
            wantMatch: true,
        },
        {
            name: "only-child",
            node: createDivElement("div", map[string]string{"class": "foo"}, []*Node{
                createDivElement("span", map[string]string{"class": "bar"}),
            }),
            pseudo:    PseudoSelector{Name: "only-child"},
            wantMatch: true,
        },
        {
            name:      "empty",
            node:      createDivElement("div", map[string]string{"class": "foo"}),
            pseudo:    PseudoSelector{Name: "empty"},
            wantMatch: true,
        },
        {
            name: "checked",
            node: createDivElement("input", map[string]string{
                "type":    "checkbox",
                "checked": "",
            }),
            pseudo:    PseudoSelector{Name: "checked"},
            wantMatch: true,
        },
        {
            name: "disabled",
            node: createDivElement("input", map[string]string{
                "type":     "button",
                "disabled": "",
            }),
            pseudo:    PseudoSelector{Name: "disabled"},
            wantMatch: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := matchesPseudo(tt.node, tt.pseudo)
            if got != tt.wantMatch {
                t.Errorf("matchesPseudo(%v, %v) = %v, want %v", tt.node, tt.pseudo, got, tt.wantMatch)
            }
        })
    }
}

func TestIsEmptyNode(t *testing.T) {
    tests := []struct {
        name      string
        node      *Node
        wantEmpty bool
    }{
        {
            name:      "empty element",
            node:      createDivElement("div", map[string]string{"class": "foo"}),
            wantEmpty: true,
        },
        {
            name: "non-empty element",
            node: createDivElement("div", map[string]string{"class": "foo"}, []*Node{
                createDivElement("span", map[string]string{}),
            }),
            wantEmpty: false,
        },
        {
            name: "element with text",
            node: createDivElement("div", map[string]string{"class": "foo"}, []*Node{
                createTextNode("text"),
            }),
            wantEmpty: false,
        },
        {
            name: "element with whitespace text",
            node: createDivElement("div", map[string]string{"class": "foo"}, []*Node{
                createTextNode("   "),
            }),
            wantEmpty: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := isEmptyNode(tt.node)
            if got != tt.wantEmpty {
                t.Errorf("isEmptyNode(%v) = %v, want %v", tt.node, got, tt.wantEmpty)
            }
        })
    }
}

// Helper functions for tests
func createDivElement(tag string, attrs map[string]string, childLists ...[]*Node) *Node {
    node := &Node{
        Type:       ElementNode,
        TagName:    tag,
        Attributes: make(map[string]string),
    }
    for k, v := range attrs {
        node.Attributes[k] = v
    }
    if len(childLists) > 0 {
        node.Children = childLists[0]
        for _, child := range node.Children {
            child.Parent = node
        }
    }
    return node
}

func createTextNode(text string) *Node {
    return &Node{
        Type: TextNode,
        Text: text,
    }
}
