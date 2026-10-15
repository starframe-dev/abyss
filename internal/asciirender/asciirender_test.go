package asciirender

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"

	"github.com/starframe-dev/abyss/internal/css"
)

func TestImageDimensionsAllowed(t *testing.T) {
	for _, tc := range []struct { w, h int; want bool }{
		{1, 1, true}, {3840, 2160, true}, {8192, MaxImagePixels / 8192, true},
		{0, 1, false}, {1, 0, false}, {-1, 1, false}, {MaxImageWidth + 1, 1, false},
		{8192, MaxImagePixels/8192 + 1, false},
	} { if got := imageDimensionsAllowed(tc.w, tc.h); got != tc.want { t.Errorf("imageDimensionsAllowed(%d,%d)=%v, want %v", tc.w,tc.h,got,tc.want) } }
}

func TestDecode(t *testing.T) {
	if _, err := Decode(nil); err == nil { t.Fatal("Decode(nil) expected error") }
	img := image.NewRGBA(image.Rect(0,0,2,3)); img.Set(0,0,color.RGBA{R:255,A:255})
	var b bytes.Buffer
	if err := png.Encode(&b,img); err != nil { t.Fatal(err) }
	decoded, err := Decode(b.Bytes()); if err != nil { t.Fatal(err) }
	if decoded.Bounds().Dx()!=2 || decoded.Bounds().Dy()!=3 { t.Fatalf("bounds=%v",decoded.Bounds()) }
}

func solid(w,h int,c color.Color) *image.RGBA {
	img:=image.NewRGBA(image.Rect(0,0,w,h)); for y:=0;y<h;y++ { for x:=0;x<w;x++ { img.Set(x,y,c) } }; return img
}

func TestQuantizeAndOptions(t *testing.T) {
	if Quantize(nil,Options{})!=nil { t.Fatal("nil image must return nil") }
	if Quantize(solid(MaxImageWidth,MaxImagePixels/MaxImageWidth+1,color.Black),Options{})!=nil { t.Fatal("oversized image must return nil") }
	grid:=Quantize(solid(8,8,color.Black),Options{Columns:4,Rows:2})
	if len(grid)!=2 || len(grid[0])!=4 { t.Fatalf("grid shape %dx%d",len(grid),len(grid[0])) }
	for _,row:=range grid { for _,c:=range row { if c.Char!=Palette[len(Palette)-1] { t.Errorf("black char=%q",c.Char) }; if c.BG != css.RGB(226,232,240) { t.Errorf("default background=%v",c.BG) } } }
	custom:=css.RGB(1,2,3); grid=Quantize(solid(4,4,color.White),Options{Columns:0,Rows:-1,BGColor:custom})
	if len(grid)==0 || len(grid[0])!=60 || grid[0][0].BG!=custom { t.Fatalf("default options/custom background not applied") }
	if grid[0][0].Char!=Palette[0] { t.Errorf("white char=%q",grid[0][0].Char) }
	bounded:=Quantize(solid(1,1,color.White),Options{Columns:MaxOutputColumns+1,Rows:MaxOutputCells})
	if len(bounded)!=MaxOutputCells/MaxOutputColumns || len(bounded[0])!=MaxOutputColumns { t.Fatalf("output dimensions=%dx%d",len(bounded),len(bounded[0])) }
}

func TestSamplingAndCharacterHelpers(t *testing.T) {
	if got:=luminance(255,255,255); got!=255 { t.Errorf("white luminance=%d",got) }
	if got:=luminance(0,0,0); got!=0 { t.Errorf("black luminance=%d",got) }
	if pickRune(0)!=Palette[len(Palette)-1] || pickRune(255)!=Palette[0] || pickRune(1000)!=Palette[0] { t.Fatal("pickRune boundary mapping incorrect") }
	r,g,b,l:=sampleCell(solid(1,1,color.RGBA{0,0,0,0}),image.Rect(0,0,1,1),0,0,1,1)
	if r!=255 || g!=255 || b!=255 || l!=255 { t.Fatalf("transparent sample=(%d,%d,%d,%d)",r,g,b,l) }
	r,g,b,_,=sampleCell(solid(1,1,color.RGBA{10,20,30,255}),image.Rect(0,0,1,1),0,0,1,1)
	if r!=10 || g!=20 || b!=30 { t.Fatalf("sample RGB=(%d,%d,%d)",r,g,b) }
	r,g,b,l=sampleCell(solid(1,1,color.Black),image.Rect(0,0,1,1),0,0,2,1)
	if r|g|b|l!=0 { t.Fatalf("empty sample=(%d,%d,%d,%d)",r,g,b,l) }
	r,g,b,a:=rgba(color.RGBA{R:12,G:34,B:56,A:78}); if r!=12||g!=34||b!=56||a!=78 { t.Fatalf("rgba=(%d,%d,%d,%d)",r,g,b,a) }
}

func TestFlatten(t *testing.T) {
	got:=Flatten([][]Cell{{{Char:'a'},{Char:0}},{{Char:'界'}}})
	if got!="a \n界\n" { t.Fatalf("Flatten=%q",got) }
	if got=Flatten(nil); got!="" { t.Fatalf("Flatten(nil)=%q",got) }
}

func TestPalette(t *testing.T) {
	if len(Palette)<2 || strings.ContainsRune(string(Palette), '\n') { t.Fatalf("unexpected palette %q",string(Palette)) }
}
