package browser

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/starframe-dev/abyss/internal/js"
	"github.com/starframe-dev/abyss/internal/layout"
	"github.com/stretchr/testify/assert"
)

func TestBuildLayoutBoxMap(t *testing.T) {
	t.Run("returns empty map for nil root", func(t *testing.T) {
		result := buildLayoutBoxMap(nil)
		assert.Empty(t, result)
	})

	t.Run("handles single node", func(t *testing.T) {
		node := &dom.Node{Type: dom.ElementNode, TagName: "div"}
		box := &layout.LayoutBox{
			Node:    node,
			Content: layout.Rect{X: 100, Y: 200, Width: 50, Height: 30},
		}
		result := buildLayoutBoxMap(box)
		assert.Len(t, result, 1)
		assert.Equal(t, js.Rect{X: 100, Y: 200, Width: 50, Height: 30}, result[node])
	})

	t.Run("handles tree structure", func(t *testing.T) {
		root := &dom.Node{Type: dom.ElementNode, TagName: "div"}
		grandChild := &dom.Node{Type: dom.ElementNode, TagName: "div"}
		child := &dom.Node{Type: dom.ElementNode, TagName: "div"}

		rootChild := &layout.LayoutBox{
			Node:    child,
			Content: layout.Rect{X: 10, Y: 20, Width: 100, Height: 80},
			Children: []*layout.LayoutBox{
				{
					Node:    grandChild,
					Content: layout.Rect{X: 0, Y: 0, Width: 50, Height: 40},
				},
			},
		}

		rootChildBox := &layout.LayoutBox{
			Node:    root,
			Content: layout.Rect{X: 5, Y: 10, Width: 200, Height: 160},
			Children: []*layout.LayoutBox{
				rootChild,
			},
		}

		result := buildLayoutBoxMap(rootChildBox)
		assert.Len(t, result, 3)
		assert.Contains(t, result, root)
		assert.Contains(t, result, child)
	})
}
