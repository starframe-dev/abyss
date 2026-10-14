package layout

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/stretchr/testify/assert"
)

// An empty text input still occupies visible space in a browser. Without a
// minimum box it collapses to zero height and disappears from the page.
func TestEmptyTextInputReservesVisibleBox(t *testing.T) {
	doc := dom.Parse(`<div><input name="q" size="57"></div>`)
	box := Build(doc, css.Compute(doc), 200)
	input := findTag(box, "input")
	if input == nil {
		t.Fatal("input box not found")
	}
	assert.Greater(t, input.Border.Height, 0.0, "empty input must have height")
	assert.Greater(t, input.Border.Width, 0.0, "empty input must have width")
}
