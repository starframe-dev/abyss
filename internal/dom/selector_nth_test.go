//go:build legacy

package dom

import (
    "testing"
)

func TestChildIndex(t *testing.T) {
    tests := []struct {
        name    string
        n       *Node
        ofType  bool
        wantIdx int
        wantErr bool
    }{
        {
            name:    "nil node",
            n:       nil,
            ofType:  false,
            wantIdx: 0,
        },
        {
            name:    "node without parent",
            n:       &Node{Type: ElementNode, TagName: "div"},
            ofType:  false,
            wantIdx: 0,
        },
        {
            name: "first element child",
            n: &Node{Type: ElementNode, TagName: "div", Parent: &Node{Children: []*Node{
                {Type: ElementNode, TagName: "div"},
                {Type: ElementNode, TagName: "div"},
                {Type: ElementNode, TagName: "div"},
            }}},
            ofType:  false,
            wantIdx: 1,
        },
        {
            name: "second element child",
            n: &Node{Type: ElementNode, TagName: "div", Parent: &Node{Children: []*Node{
                {Type: ElementNode, TagName: "div"},
                {Type: ElementNode, TagName: "div"},
                {Type: ElementNode, TagName: "div"},
            }}},
            ofType:  false,
            wantIdx: 2,
        },
        {
            name: "third element child",
            n: &Node{Type: ElementNode, TagName: "div", Parent: &Node{Children: []*Node{
                {Type: ElementNode, TagName: "div"},
                {Type: ElementNode, TagName: "div"},
                {Type: ElementNode, TagName: "div"},
            }}},
            ofType:  false,
            wantIdx: 3,
        },
        {
            name: "mixed types, ofType false",
            n: &Node{Type: ElementNode, TagName: "span", Parent: &Node{Children: []*Node{
                {Type: ElementNode, TagName: "div"},
                {Type: ElementNode, TagName: "span"},
                {Type: ElementNode, TagName: "div"},
            }}},
            ofType:  false,
            wantIdx: 2,
        },
        {
            name: "mixed types, ofType true same tag",
            n: &Node{Type: ElementNode, TagName: "span", Parent: &Node{Children: []*Node{
                {Type: ElementNode, TagName: "div"},
                {Type: ElementNode, TagName: "span"},
                {Type: ElementNode, TagName: "div"},
            }}},
            ofType:  true,
            wantIdx: 1,
        },
        {
            name: "mixed types, ofType true different tag",
            n: &Node{Type: ElementNode, TagName: "div", Parent: &Node{Children: []*Node{
                {Type: ElementNode, TagName: "div"},
                {Type: ElementNode, TagName: "span"},
                {Type: ElementNode, TagName: "div"},
            }}},
            ofType:  true,
            wantIdx: 1,
        },
        {
            name: "mixed types case insensitive",
            n: &Node{Type: ElementNode, TagName: "SPAN", Parent: &Node{Children: []*Node{
                {Type: ElementNode, TagName: "div"},
                {Type: ElementNode, TagName: "span"},
                {Type: ElementNode, TagName: "div"},
            }}},
            ofType:  true,
            wantIdx: 1,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := childIndex(tt.n, tt.ofType)
            if got != tt.wantIdx {
                t.Errorf("childIndex() = %d, want %d", got, tt.wantIdx)
            }
        })
    }
}

func TestLastChildIndex(t *testing.T) {
    tests := []struct {
        name    string
        n       *Node
        ofType  bool
        wantIdx int
    }{
        {
            name:    "nil node",
            n:       nil,
            ofType:  false,
            wantIdx: 0,
        },
        {
            name:    "node without parent",
            n:       &Node{Type: ElementNode, TagName: "div"},
            ofType:  false,
            wantIdx: 0,
        },
        {
            name: "last element child",
            n: &Node{Type: ElementNode, TagName: "div", Parent: &Node{Children: []*Node{
                {Type: ElementNode, TagName: "div"},
                {Type: ElementNode, TagName: "div"},
                {Type: ElementNode, TagName: "div"},
            }}},
            ofType:  false,
            wantIdx: 3,
        },
        {
            name: "second to last element child",
            n: &Node{Type: ElementNode, TagName: "div", Parent: &Node{Children: []*Node{
                {Type: ElementNode, TagName: "div"},
                {Type: ElementNode, TagName: "div"},
                {Type: ElementNode, TagName: "div"},
            }}},
            ofType:  false,
            wantIdx: 2,
        },
        {
            name: "first element child from last",
            n: &Node{Type: ElementNode, TagName: "div", Parent: &Node{Children: []*Node{
                {Type: ElementNode, TagName: "div"},
                {Type: ElementNode, TagName: "div"},
                {Type: ElementNode, TagName: "div"},
            }}},
            ofType:  false,
            wantIdx: 1,
        },
        {
            name: "mixed types, ofType false",
            n: &Node{Type: ElementNode, TagName: "span", Parent: &Node{Children: []*Node{
                {Type: ElementNode, TagName: "div"},
                {Type: ElementNode, TagName: "span"},
                {Type: ElementNode, TagName: "div"},
            }}},
            ofType:  false,
            wantIdx: 2,
        },
        {
            name: "mixed types, ofType true same tag",
            n: &Node{Type: ElementNode, TagName: "span", Parent: &Node{Children: []*Node{
                {Type: ElementNode, TagName: "div"},
                {Type: ElementNode, TagName: "span"},
                {Type: ElementNode, TagName: "div"},
            }}},
            ofType:  true,
            wantIdx: 1,
        },
        {
            name: "mixed types, ofType true different tag",
            n: &Node{Type: ElementNode, TagName: "div", Parent: &Node{Children: []*Node{
                {Type: ElementNode, TagName: "div"},
                {Type: ElementNode, TagName: "span"},
                {Type: ElementNode, TagName: "div"},
            }}},
            ofType:  true,
            wantIdx: 2,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := lastChildIndex(tt.n, tt.ofType)
            if got != tt.wantIdx {
                t.Errorf("lastChildIndex() = %d, want %d", got, tt.wantIdx)
            }
        })
    }
}

func TestMatchNth(t *testing.T) {
    tests := []struct {
        name string
        arg  string
        idx  int
        want bool
    }{
        {
            name: "arg empty",
            arg:  "",
            idx:  5,
            want: false,
        },
        {
            name: "arg whitespace",
            arg:  "  ",
            idx:  5,
            want: false,
        },
        {
            name: "arg zero matches first only",
            arg:  "0",
            idx:  0,
            want: true,
        },
        {
            name: "arg zero does not match second",
            arg:  "0",
            idx:  1,
            want: false,
        },
        {
            name: "arg even matches even indices",
            arg:  "even",
            idx:  0,
            want: true,
        },
        {
            name: "arg even matches even indices",
            arg:  "even",
            idx:  2,
            want: true,
        },
        {
            name: "arg even matches even indices",
            arg:  "even",
            idx:  4,
            want: true,
        },
        {
            name: "arg even does not match odd indices",
            arg:  "even",
            idx:  1,
            want: false,
        },
        {
            name: "arg odd matches odd indices",
            arg:  "odd",
            idx:  1,
            want: true,
        },
        {
            name: "arg odd matches odd indices",
            arg:  "odd",
            idx:  3,
            want: true,
        },
        {
            name: "arg odd does not match even indices",
            arg:  "odd",
            idx:  0,
            want: false,
        },
        {
            name: "arg plain number matches",
            arg:  "2",
            idx:  2,
            want: true,
        },
        {
            name: "arg plain number does not match",
            arg:  "2",
            idx:  3,
            want: false,
        },
        {
            name: "arg plain number invalid",
            arg:  "xyz",
            idx:  5,
            want: false,
        },
        {
            name: "arg 2n+1 matches",
            arg:  "2n+1",
            idx:  1,
            want: true,
        },
        {
            name: "arg 2n+1 matches",
            arg:  "2n+1",
            idx:  3,
            want: true,
        },
        {
            name: "arg 2n+1 matches",
            arg:  "2n+1",
            idx:  5,
            want: true,
        },
        {
            name: "arg 2n+1 does not match",
            arg:  "2n+1",
            idx:  0,
            want: false,
        },
        {
            name: "arg 2n+1 does not match",
            arg:  "2n+1",
            idx:  2,
            want: false,
        },
        {
            name: "arg 3n matches",
            arg:  "3n",
            idx:  0,
            want: true,
        },
        {
            name: "arg 3n matches",
            arg:  "3n",
            idx:  3,
            want: true,
        },
        {
            name: "arg 3n matches",
            arg:  "3n",
            idx:  6,
            want: true,
        },
        {
            name: "arg 3n does not match",
            arg:  "3n",
            idx:  1,
            want: false,
        },
        {
            name: "arg -n+3 matches",
            arg:  "-n+3",
            idx:  3,
            want: true,
        },
        {
            name: "arg -n+3 matches",
            arg:  "-n+3",
            idx:  4,
            want: true,
        },
        {
            name: "arg -n+3 matches",
            arg:  "-n+3",
            idx:  5,
            want: true,
        },
        {
            name: "arg -n+3 does not match",
            arg:  "-n+3",
            idx:  0,
            want: false,
        },
        {
            name: "arg -n+3 does not match",
            arg:  "-n+3",
            idx:  2,
            want: false,
        },
        {
            name: "arg n matches",
            arg:  "n",
            idx:  0,
            want: true,
        },
        {
            name: "arg n matches",
            arg:  "n",
            idx:  1,
            want: true,
        },
        {
            name: "arg n matches",
            arg:  "n",
            idx:  2,
            want: true,
        },
        {
            name: "arg n with spaces",
            arg:  "  2n + 1  ",
            idx:  1,
            want: true,
        },
        {
            name: "arg 2n-1 matches",
            arg:  "2n-1",
            idx:  1,
            want: true,
        },
        {
            name: "arg 2n-1 matches",
            arg:  "2n-1",
            idx:  3,
            want: true,
        },
        {
            name: "arg 2n-1 matches",
            arg:  "2n-1",
            idx:  5,
            want: true,
        },
        {
            name: "arg 2n-1 does not match",
            arg:  "2n-1",
            idx:  0,
            want: false,
        },
        {
            name: "arg 5n+2 matches",
            arg:  "5n+2",
            idx:  2,
            want: true,
        },
        {
            name: "arg 5n+2 matches",
            arg:  "5n+2",
            idx:  7,
            want: true,
        },
        {
            name: "arg -2n+5 matches",
            arg:  "-2n+5",
            idx:  5,
            want: true,
        },
        {
            name: "arg -2n+5 matches",
            arg:  "-2n+5",
            idx:  3,
            want: true,
        },
        {
            name: "arg -2n+5 does not match",
            arg:  "-2n+5",
            idx:  1,
            want: false,
        },
        {
            name: "arg 0n matches",
            arg:  "0n",
            idx:  0,
            want: true,
        },
        {
            name: "arg 0n matches",
            arg:  "0n",
            idx:  5,
            want: true,
        },
        {
            name: "arg 0n matches",
            arg:  "0n",
            idx:  100,
            want: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := matchNth(tt.arg, tt.idx)
            if got != tt.want {
                t.Errorf("matchNth(%q, %d) = %v, want %v", tt.arg, tt.idx, got, tt.want)
            }
        })
    }
}

func TestParseNth(t *testing.T) {
    tests := []struct {
        name   string
        s      string
        wantA  int
        wantB  int
        wantOk bool
    }{
        {
            name:   "empty string",
            s:      "",
            wantA:  0,
            wantB:  0,
            wantOk: false,
        },
        {
            name:   "zero",
            s:      "0",
            wantA:  0,
            wantB:  0,
            wantOk: true,
        },
        {
            name:   "2n+1",
            s:      "2n+1",
            wantA:  2,
            wantB:  1,
            wantOk: true,
        },
        {
            name:   "2n",
            s:      "2n",
            wantA:  2,
            wantB:  0,
            wantOk: true,
        },
        {
            name:   "n",
            s:      "n",
            wantA:  1,
            wantB:  0,
            wantOk: true,
        },
        {
            name:   "-n+3",
            s:      "-n+3",
            wantA:  -1,
            wantB:  3,
            wantOk: true,
        },
        {
            name:   "-2n+5",
            s:      "-2n+5",
            wantA:  -2,
            wantB:  5,
            wantOk: true,
        },
        {
            name:   "3n-2",
            s:      "3n-2",
            wantA:  3,
            wantB:  -2,
            wantOk: true,
        },
        {
            name:   "single n",
            s:      " n ",
            wantA:  1,
            wantB:  0,
            wantOk: true,
        },
        {
            name:   "single 2n",
            s:      " 2n ",
            wantA:  2,
            wantB:  0,
            wantOk: true,
        },
        {
            name:   "invalid formula",
            s:      "2x+1",
            wantA:  0,
            wantB:  0,
            wantOk: false,
        },
        {
            name:   "just plus",
            s:      "+",
            wantA:  0,
            wantB:  0,
            wantOk: false,
        },
        {
            name:   "just minus",
            s:      "-",
            wantA:  -1,
            wantB:  0,
            wantOk: true,
        },
        {
            name:   "5n",
            s:      "5n",
            wantA:  5,
            wantB:  0,
            wantOk: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            a, b, ok := parseNth(tt.s)
            if a != tt.wantA {
                t.Errorf("parseNth(%q) got a=%d, want %d", tt.s, a, tt.wantA)
            }
            if b != tt.wantB {
                t.Errorf("parseNth(%q) got b=%d, want %d", tt.s, b, tt.wantB)
            }
            if ok != tt.wantOk {
                t.Errorf("parseNth(%q) got ok=%v, want %v", tt.s, ok, tt.wantOk)
            }
        })
    }
}
