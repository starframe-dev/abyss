package css

import "testing"

func TestLengthConstructorsResolveAndString(t *testing.T) {
 tests := []struct{name string; got Length; resolve float64; text string}{
  {"auto", Auto(), 0, "auto"}, {"pixels", Px(12.5), 12.5, "12.50px"},
  {"percent", Percent(25), 50, "25.00%"}, {"fraction", Fr(2.5), 0, "2.50fr"},
 }
 for _, tt := range tests { t.Run(tt.name, func(t *testing.T) { if got:=tt.got.Resolve(200); got!=tt.resolve { t.Errorf("Resolve()=%v, want %v",got,tt.resolve) }; if got:=tt.got.String(); got!=tt.text { t.Errorf("String()=%q, want %q",got,tt.text) } }) }
}

func TestLengthPrecedence(t *testing.T) {
 l:=Length{IsAuto:true, IsPercent:true, Percent:50, IsFraction:true, Fraction:3, Px:9}
 if l.Resolve(100)!=0 || l.String()!="auto" { t.Fatalf("auto precedence failed: %#v",l) }
 l=Length{IsPercent:true,Percent:50,IsFraction:true,Fraction:3,Px:9}
 if l.Resolve(100)!=50 || l.String()!="50.00%" { t.Fatalf("percent precedence failed: %#v",l) }
}

func TestColorConstructorsAndString(t *testing.T) {
 if got:=Transparent(); got.IsSet || got.String()!="transparent" { t.Fatalf("Transparent()=%+v (%s)",got,got.String()) }
 c:=RGB(0, 15, 255)
 if !c.IsSet || c.R!=0 || c.G!=15 || c.B!=255 || c.String()!="#000fff" { t.Errorf("RGB/String()=%+v %q",c,c.String()) }
}

func TestBoxModelIsZero(t *testing.T) {
 if !(BoxModel{}).IsZero() { t.Fatal("zero box should be zero") }
 cases:=[]Length{Auto(),Percent(0),Fr(0),Px(1)}
 for _, side:=range cases { if (BoxModel{Top:side}).IsZero() { t.Errorf("box with side %+v should not be zero",side) } }
}

func TestDefault(t *testing.T) {
 s:=Default()
 if s.Display!=DisplayBlock || s.Position!=PositionStatic || s.Overflow!=OverflowVisible || !s.Width.IsAuto || !s.Height.IsAuto { t.Errorf("unexpected defaults: %+v",s) }
 if s.BackgroundColor!=Transparent() || s.Color!=RGB(30,30,30) || s.FlexDirection!=FlexDirectionRow || s.JustifyContent!=JustifyFlexStart || s.AlignItems!=AlignStretch || s.FlexGrow!=0 || s.FlexShrink!=1 || s.Gap!=Px(0) { t.Errorf("unexpected defaults: %+v",s) }
}
