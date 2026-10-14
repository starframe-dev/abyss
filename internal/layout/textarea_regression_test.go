package layout

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/stretchr/testify/assert"
)

// An empty textarea (google.com search box since 2026) must reserve a visible
// line, like an empty text input.
func TestEmptyTextareaReservesVisibleBox(t *testing.T) {
	doc := dom.Parse(`<div><textarea name="q" rows="1"></textarea></div>`)
	box := Build(doc, css.Compute(doc), 200)
	ta := findTag(box, "textarea")
	if ta == nil {
		t.Fatal("textarea box not found")
	}
	assert.Greater(t, ta.Border.Height, 0.0, "empty textarea must have height")
}
