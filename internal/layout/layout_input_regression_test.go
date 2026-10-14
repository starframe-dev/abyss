package layout

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/stretchr/testify/assert"
)

func TestInputValueIsLaidOut(t *testing.T) {
	doc := dom.Parse(`<div><input name="q" value="golang"></div>`)
	box := Build(doc, css.Compute(doc), 200)
	input := findTag(box, "input")
	if input == nil {
		t.Fatal("input box not found")
	}
	if assert.NotEmpty(t, input.TextLines) {
		assert.Equal(t, "golang", input.TextLines[0].Text)
	}
}

func TestHiddenInputIsNotLaidOut(t *testing.T) {
	doc := dom.Parse(`<div><input type="hidden" name="token" value="secret123"></div>`)
	box := Build(doc, css.Compute(doc), 200)
	input := findTag(box, "input")
	if input == nil {
		t.Fatal("input box not found")
	}
	assert.Empty(t, input.TextLines)
}

func TestInputFallsBackToPlaceholder(t *testing.T) {
	doc := dom.Parse(`<div><input placeholder="Search"></div>`)
	box := Build(doc, css.Compute(doc), 200)
	input := findTag(box, "input")
	if input == nil {
		t.Fatal("input box not found")
	}
	if assert.NotEmpty(t, input.TextLines) {
		assert.Equal(t, "Search", input.TextLines[0].Text)
	}
}

func TestMixedTextNextToBlockChildIsKept(t *testing.T) {
	doc := dom.Parse(`<div>Hello<p>block</p>tail</div>`)
	box := Build(doc, css.Compute(doc), 200)
	div := findTag(box, "div")
	if div == nil {
		t.Fatal("div not found")
	}
	var got []string
	var collect func(b *LayoutBox)
	collect = func(b *LayoutBox) {
		for _, l := range b.TextLines {
			got = append(got, l.Text)
		}
		for _, c := range b.Children {
			collect(c)
		}
	}
	collect(div)
	assert.Contains(t, got, "Hello")
	assert.Contains(t, got, "tail")
}
