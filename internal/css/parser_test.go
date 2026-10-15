package css

import "testing"

func TestParseInline(t *testing.T) {
 got := ParseInline(" color : red ; background: url(data:x;y); bad; color: blue !IMPORTANT ")
 want := []Declaration{{Property:"color", Value:"red"}, {Property:"background", Value:"url(data:x;y)"}, {Property:"color", Value:"blue", Important:true}}
 if len(got)!=len(want) { t.Fatalf("got %#v",got) }; for i:=range want { if got[i]!=want[i] { t.Errorf("[%d] got %#v want %#v",i,got[i],want[i]) } }
}

func TestParseStyleSheet(t *testing.T) {
 got:=ParseStyleSheet("/*x*/ div, span { color: red !important; content: '}'; } @media screen { p {x: y} @supports (a:b) { i {z: q} } } @font-face {a:b}")
 if len(got)!=3 { t.Fatalf("got %#v",got) }
 if got[0].Selectors[0]!="div" || got[0].Selectors[1]!="span" || len(got[0].Declarations)!=2 || !got[0].Declarations[0].Important { t.Errorf("rule: %#v",got[0]) }
 if got[1].MediaQuery!="screen" || got[1].Selectors[0]!="p" { t.Errorf("media: %#v",got[1]) }
 if got[2].MediaQuery!="screen" || got[2].SupportsQuery!="(a:b)" { t.Errorf("nested: %#v",got[2]) }
 if len(ParseStyleSheet("div {x:y"))!=0 { t.Error("malformed rule accepted") }
}

func TestSpecificity(t *testing.T) {
 tests:=map[string]int{"div.class#id[attr]:hover":112,"div::before":10,"div > p":2,"*":0,"[x]":10,"svg|a":0}
 for selector,want:=range tests { if got:=Specificity(selector); got!=want { t.Errorf("Specificity(%q)=%d want %d",selector,got,want) } }
}
