package browser

import (
    "github.com/starframe/abyss/internal/dom"
    "github.com/starframe/abyss/internal/js"
    "github.com/starframe/abyss/internal/layout"
)

// buildLayoutBoxMap traverses a layout tree and builds a map of node -> absolute bounding rect.
func buildLayoutBoxMap(root *layout.LayoutBox) map[*dom.Node]js.Rect {
    result := map[*dom.Node]js.Rect{}
    buildLayoutBoxRecursive(root, 0, 0, result)
    return result
}

func buildLayoutBoxRecursive(box *layout.LayoutBox, offsetX, offsetY float64, result map[*dom.Node]js.Rect) {
    if box == nil || box.Node == nil {
        return
    }

    absX := box.Content.X + offsetX
    absY := box.Content.Y + offsetY
    result[box.Node] = js.Rect{X: absX, Y: absY, Width: box.Content.Width, Height: box.Content.Height}

    for _, child := range box.Children {
        buildLayoutBoxRecursive(child, absX, absY, result)
    }
}

// BuildLayoutBoxMapForTest exposes buildLayoutBoxMap for testing.
func BuildLayoutBoxMapForTest(root *layout.LayoutBox) map[*dom.Node]js.Rect {
    return buildLayoutBoxMap(root)
}
