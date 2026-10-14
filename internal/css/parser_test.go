//go:build legacy

package css

import (
    "testing"
)

func TestParseInline(t *testing.T) {
    tests := []struct {
        name      string
        input     string
        wantLen   int
        wantProps []string
        wantVals  []string
    }{
        {
            name:      "simple single declaration",
            input:     "color:red",
            wantLen:   1,
            wantProps: []string{"color"},
            wantVals:  []string{"red"},
        },
        {
            name:      "multiple declarations",
            input:     "color:red;font-size:12px",
            wantLen:   2,
            wantProps: []string{"color", "font-size"},
            wantVals:  []string{"red", "12px"},
        },
        {
            name:    "empty input",
            input:   "",
            wantLen: 0,
        },
        {
            name:      "whitespace around declarations",
            input:     "  color : red ; font-size : 12px  ",
            wantLen:   2,
            wantProps: []string{"color", "font-size"},
            wantVals:  []string{"red", "12px"},
        },
        {
            name:      "invalid declarations (no colon)",
            input:     "color red; font-size: 12px",
            wantLen:   1,
            wantProps: []string{"font-size"},
            wantVals:  []string{"12px"},
        },
        {
            name:      "empty declaration values",
            input:     "color:red;",
            wantLen:   1,
            wantProps: []string{"color"},
            wantVals:  []string{"red"},
        },
        {
            name:      "multiple semicolons",
            input:     "color:red;;font-size:12px",
            wantLen:   2,
            wantProps: []string{"color", "font-size"},
            wantVals:  []string{"red", "12px"},
        },
        {
            name:    "only whitespace",
            input:   "   ",
            wantLen: 0,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := ParseInline(tt.input)
            if len(got) != tt.wantLen {
                t.Errorf("ParseInline() = %d declarations; want %d", len(got), tt.wantLen)
                return
            }
            for i := 0; i < tt.wantLen; i++ {
                if got[i].Property != tt.wantProps[i] {
                    t.Errorf("ParseInline() property[%d] = %q; want %q", i, got[i].Property, tt.wantProps[i])
                }
                if got[i].Value != tt.wantVals[i] {
                    t.Errorf("ParseInline() value[%d] = %q; want %q", i, got[i].Value, tt.wantVals[i])
                }
            }
        })
    }
}

func TestParseStyleSheet(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantLen int
    }{
        {
            name:    "simple single rule",
            input:   "div { color: red; }",
            wantLen: 1,
        },
        {
            name:    "multiple rules",
            input:   "div { color: red; } p { font-size: 12px; }",
            wantLen: 2,
        },
        {
            name:    "selectors with spaces",
            input:   "div , span { color: red; }",
            wantLen: 1,
        },
        {
            name:    "empty input",
            input:   "",
            wantLen: 0,
        },
        {
            name:    "only braces",
            input:   "{}",
            wantLen: 0,
        },
        {
            name:    "complex selectors",
            input:   "div.class#id { color: red; } div.class { color: blue; }",
            wantLen: 2,
        },
        {
            name:    "nested braces in values",
            input:   "div { content: \"(100% - 10px) / 2\"; }",
            wantLen: 1,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := ParseStyleSheet(tt.input)
            if len(got) != tt.wantLen {
                t.Errorf("ParseStyleSheet() = %d rules; want %d", len(got), tt.wantLen)
                return
            }
            if len(got) > 0 {
                for i, rule := range got {
                    if len(rule.Selectors) == 0 {
                        t.Errorf("Rule[%d] has no selectors", i)
                    }
                }
            }
        })
    }
}

func TestSpecificity(t *testing.T) {
    tests := []struct {
        name  string
        input string
        want  int
    }{
        {
            name:  "element selector",
            input: "div",
            want:  1,
        },
        {
            name:  "element and class",
            input: "div.class",
            want:  11,
        },
        {
            name:  "element and id",
            input: "div#id",
            want:  101,
        },
        {
            name:  "class and id",
            input: ".class#id",
            want:  110,
        },
        {
            name:  "multiple classes",
            input: "div.class1.class2",
            want:  12,
        },
        {
            name:  "multiple ids",
            input: "div#id1#id2",
            want:  201,
        },
        {
            name:  "multiple elements",
            input: "div p span",
            want:  3,
        },
        {
            name:  "element with pseudo-class",
            input: "div:hover",
            want:  11,
        },
        {
            name:  "element with pseudo-element",
            input: "div::before",
            want:  10,
        },
        {
            name:  "attribute selector",
            input: "div[attr]",
            want:  11,
        },
        {
            name:  "descendant combinator",
            input: "div p",
            want:  2,
        },
        {
            name:  "child combinator",
            input: "div > p",
            want:  2,
        },
        {
            name:  "sibling combinator",
            input: "div + p",
            want:  2,
        },
        {
            name:  "universal selector",
            input: "*",
            want:  0,
        },
        {
            name:  "complex selector",
            input: "div.class#id[attr]:hover",
            want:  112,
        },
        {
            name:  "whitespace trimmed",
            input: "  div  ",
            want:  1,
        },
        {
            name:  "empty string",
            input: "",
            want:  0,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Specificity(tt.input)
            if got != tt.want {
                t.Errorf("Specificity(%q) = %d; want %d", tt.input, got, tt.want)
            }
        })
    }
}
