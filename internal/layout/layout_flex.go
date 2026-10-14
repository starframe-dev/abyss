// Package layout computes positions and sizes of DOM nodes using CSS box model.
package layout

import (
    "github.com/starframe/abyss/internal/css"
)

func layoutFlex(box *LayoutBox, contentWidth float64) float64 {
    style := box.Style
    flexDir := style.FlexDirection
    gap := style.Gap.Resolve(contentWidth)

    numChildren := len(box.Children)
    if numChildren == 0 {
        return 0
    }
    totalGap := gap * float64(numChildren-1)

    availableSpace := flexAvailableSpace(flexDir, contentWidth, totalGap)
    totalBasis, totalGrow, _ := flexTotals(box, contentWidth)
    remaining := flexRemaining(flexDir, availableSpace, totalBasis)

    var pos float64
    for _, child := range box.Children {
        pos = layoutFlexChild(box, child, flexDir, gap, contentWidth, pos, remaining, totalGrow)
    }

    return flexContainerSize(box, flexDir, pos, gap)
}

func flexAvailableSpace(flexDir css.FlexDirection, contentWidth, totalGap float64) float64 {
    if flexDir == css.FlexDirectionRow {
        return contentWidth - totalGap
    }
    return 100000 // Large number for column height
}

func flexTotals(box *LayoutBox, contentWidth float64) (basis, grow, shrink float64) {
    for _, child := range box.Children {
        b := child.Style.Width.Resolve(contentWidth)
        if b == 0 {
            b = 100 // default basis
        }
        basis += b
        grow += child.Style.FlexGrow
        shrink += child.Style.FlexShrink
    }
    return
}

func flexRemaining(flexDir css.FlexDirection, availableSpace, totalBasis float64) float64 {
    if flexDir == css.FlexDirectionRow {
        return availableSpace - totalBasis
    }
    return 0
}

func layoutFlexChild(box, child *LayoutBox, flexDir css.FlexDirection, gap, contentWidth, pos, remaining, totalGrow float64) float64 {
    basis := child.Style.Width.Resolve(contentWidth)
    if basis == 0 {
        basis = 100
    }

    // Apply flex grow/shrink
    childWidth := basis
    if totalGrow > 0 && remaining > 0 {
        childWidth += remaining * child.Style.FlexGrow / totalGrow
    }

    // Temporarily set width for layout
    origWidth := child.Style.Width
    child.Style.Width = css.Px(childWidth)

    if flexDir == css.FlexDirectionRow {
        layoutBox(child, childWidth, box.Content.X+pos, box.Content.Y, false)
        pos += childWidth + gap
    } else {
        layoutBox(child, contentWidth, box.Content.X, box.Content.Y+pos, false)
        pos += child.Margin.Height + gap
    }

    child.Style.Width = origWidth
    return pos
}

func flexContainerSize(box *LayoutBox, flexDir css.FlexDirection, pos, gap float64) float64 {
    if flexDir == css.FlexDirectionRow {
        var maxHeight float64
        for _, child := range box.Children {
            if child.Margin.Height > maxHeight {
                maxHeight = child.Margin.Height
            }
        }
        return maxHeight
    }
    return pos - gap
}
