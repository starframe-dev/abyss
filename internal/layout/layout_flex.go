// Package layout computes positions and sizes of DOM nodes using CSS box model.
package layout

import (
	"github.com/starframe-dev/abyss/internal/css"
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
	totalShrinkWeight := flexShrinkWeight(box, contentWidth)

	pos, distributedGap := flexJustification(style.JustifyContent, remaining, numChildren, totalGrow)
	gap += distributedGap
	for _, child := range box.Children {
		pos = layoutFlexChild(box, child, flexDir, gap, contentWidth, pos, remaining, totalGrow, totalShrinkWeight)
	}

	return flexContainerSize(box, flexDir, pos, gap)
}

func flexJustification(justify css.JustifyContent, remaining float64, children int, totalGrow float64) (offset, extraGap float64) {
	if remaining <= 0 || children == 0 || totalGrow > 0 {
		return 0, 0
	}
	switch justify {
	case css.JustifyCenter:
		return remaining / 2, 0
	case css.JustifyFlexEnd:
		return remaining, 0
	case css.JustifySpaceBetween:
		if children > 1 {
			return 0, remaining / float64(children-1)
		}
	case css.JustifySpaceAround:
		extraGap = remaining / float64(children)
		return extraGap / 2, extraGap
	}
	return 0, 0
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

func flexShrinkWeight(box *LayoutBox, contentWidth float64) float64 {
	var total float64
	for _, child := range box.Children {
		basis := child.Style.Width.Resolve(contentWidth)
		if basis == 0 {
			basis = 100
		}
		if child.Style.FlexShrink > 0 {
			total += basis * child.Style.FlexShrink
		}
	}
	return total
}

func layoutFlexChild(box, child *LayoutBox, flexDir css.FlexDirection, gap, contentWidth, pos, remaining, totalGrow, totalShrinkWeight float64) float64 {
	basis := child.Style.Width.Resolve(contentWidth)
	if basis == 0 {
		basis = 100
	}

	// Apply flex grow/shrink
	childWidth := basis
	if totalGrow > 0 && remaining > 0 {
		childWidth += remaining * child.Style.FlexGrow / totalGrow
	} else if remaining < 0 && totalShrinkWeight > 0 && child.Style.FlexShrink > 0 {
		childWidth = max(0, basis+remaining*(basis*child.Style.FlexShrink)/totalShrinkWeight)
	}

	if flexDir == css.FlexDirectionRow {
		origWidth := child.Style.Width
		child.Style.Width = css.Px(childWidth)
		layoutBox(child, childWidth, box.Content.X+pos, box.Content.Y, false)
		child.Style.Width = origWidth
		return pos + childWidth + gap
	}

	layoutBox(child, contentWidth, box.Content.X, box.Content.Y+pos, false)
	return pos + child.Margin.Height + gap
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
