// Package layout computes positions and sizes of DOM nodes using CSS box model.
package layout

import (
	"sort"
	"strconv"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/starframe-dev/abyss/internal/cellmetrics"
	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
)

// flexFitted holds the fitted widths of the row being laid out; it is set only
// for the duration of layoutFlexRow.
var flexFitted map[*LayoutBox]float64

func layoutFlex(box *LayoutBox, contentWidth float64) float64 {
	if len(box.Children) == 0 {
		return 0
	}
	if box.Style.FlexDirection == css.FlexDirectionColumn {
		return layoutFlexColumn(box, contentWidth)
	}
	children := box.Children
	box.Children = orderedFlexChildren(children)
	defer func() { box.Children = children }()
	if box.Style.FlexWrap {
		return layoutFlexWrapped(box, contentWidth)
	}
	return layoutFlexRow(box, contentWidth)
}

// orderedFlexChildren sorts children by CSS order. Children with equal order keep
// their DOM order, and the slice is returned unchanged when no child sets order.
func orderedFlexChildren(children []*LayoutBox) []*LayoutBox {
	for _, c := range children {
		if c.Style.Order != 0 {
			sorted := append([]*LayoutBox(nil), children...)
			sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Style.Order < sorted[j].Style.Order })
			return sorted
		}
	}
	return children
}

// layoutFlexWrapped breaks a row into lines whose item basis fits contentWidth,
// lays out each line, and stacks the lines vertically (flex-wrap: wrap).
func layoutFlexWrapped(box *LayoutBox, contentWidth float64) float64 {
	gap := box.Style.Gap.Resolve(contentWidth)
	children := box.Children
	top := box.Content.Y
	var total float64
	lines := 0
	var line []*LayoutBox
	var used float64
	flush := func() {
		if len(line) == 0 {
			return
		}
		box.Children = line
		box.Content.Y = top + total
		h := layoutFlexRow(box, contentWidth)
		if lines > 0 {
			total += gap
		}
		total += h
		lines++
		line = nil
		used = 0
	}
	for _, child := range children {
		basis := flexBasis(child, contentWidth)
		if len(line) > 0 && used+gap+basis > contentWidth {
			flush()
		}
		if len(line) > 0 {
			used += gap
		}
		line = append(line, child)
		used += basis
	}
	flush()
	box.Children = children
	box.Content.Y = top
	return total
}

// layoutFlexRow lays out box.Children on a single row.
func layoutFlexRow(box *LayoutBox, contentWidth float64) float64 {
	style := box.Style
	flexDir := style.FlexDirection
	gap := style.Gap.Resolve(contentWidth)

	numChildren := len(box.Children)
	totalGap := gap * float64(numChildren-1)

	availableSpace := flexAvailableSpace(flexDir, contentWidth, totalGap)
	totalBasis, totalGrow, _ := flexTotals(box, contentWidth)
	remaining := flexRemaining(flexDir, availableSpace, totalBasis)
	totalShrinkWeight := flexShrinkWeight(box, contentWidth)

	pos, distributedGap := flexJustification(style.JustifyContent, remaining, numChildren, totalGrow)
	gap += distributedGap
	flexFitted = flexFittedWidths(box, contentWidth, gap)
	defer func() { flexFitted = nil }()
	for _, child := range box.Children {
		pos = layoutFlexChild(box, child, flexDir, gap, contentWidth, pos, remaining, totalGrow, totalShrinkWeight)
	}

	return flexContainerSize(box, flexDir, pos, gap)
}

func layoutFlexColumn(box *LayoutBox, contentWidth float64) float64 {
	gap := box.Style.Gap.Resolve(contentWidth)
	heights := make([]float64, len(box.Children))
	var contentHeight float64
	for i, child := range box.Children {
		layoutBox(child, contentWidth, box.Content.X, box.Content.Y+contentHeight, false)
		heights[i] = child.Margin.Height
		contentHeight += heights[i]
		if i > 0 {
			contentHeight += gap
		}
	}
	if box.Style.Height.IsAuto {
		return contentHeight
	}
	remaining := max(0, box.Style.Height.Resolve(contentWidth)-contentHeight)
	var totalGrow float64
	for _, child := range box.Children {
		totalGrow += child.Style.FlexGrow
	}
	growth := make([]float64, len(box.Children))
	if totalGrow > 0 && remaining > 0 {
		for i, child := range box.Children {
			growth[i] = remaining * child.Style.FlexGrow / totalGrow
			heights[i] += growth[i]
		}
		contentHeight += remaining
		remaining = 0
	}
	offset, extraGap := flexJustification(box.Style.JustifyContent, remaining, len(box.Children), totalGrow)
	gap += extraGap
	pos := offset
	for i, child := range box.Children {
		if growth[i] > 0 {
			originalHeight := child.Style.Height
			child.Style.Height = css.Px(child.Content.Height + growth[i])
			layoutBox(child, contentWidth, box.Content.X, box.Content.Y+pos, false)
			child.Style.Height = originalHeight
		} else {
			layoutBox(child, contentWidth, box.Content.X, box.Content.Y+pos, false)
		}
		pos += heights[i] + gap
	}
	return contentHeight
}

// flexFittedWidths returns the width of each row item with its min-width applied,
// scaled down so that the items together fit the row. Only items with min-width
// are shrunk; the rest keep their width. It is empty when nothing needs shrinking.
func flexFittedWidths(box *LayoutBox, contentWidth, gap float64) map[*LayoutBox]float64 {
	available := contentWidth - gap*float64(len(box.Children)-1)
	var total float64
	var minSum float64
	widths := make(map[*LayoutBox]float64, len(box.Children))
	for _, child := range box.Children {
		w := flexBasis(child, contentWidth)
		if mw := child.Style.MinWidth; !mw.IsAuto {
			w = max(w, mw.Resolve(contentWidth))
			minSum += mw.Resolve(contentWidth)
		}
		widths[child] = w
		total += w
	}
	if total <= available || minSum == 0 {
		return nil
	}
	// Give back the excess from items with min-width, in proportion to their width.
	excess := total - available
	var shrinkable float64
	for child, w := range widths {
		if !child.Style.MinWidth.IsAuto {
			shrinkable += w - min(w, child.Style.MinWidth.Resolve(contentWidth))
		}
	}
	if shrinkable <= 0 {
		return nil
	}
	for child, w := range widths {
		if child.Style.MinWidth.IsAuto {
			continue
		}
		floor := min(w, child.Style.MinWidth.Resolve(contentWidth))
		share := (w - floor) / shrinkable
		widths[child] = max(floor, w-excess*share)
	}
	return widths
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

// flexBasis returns the width a flex item starts from: its declared width, else
// its content width when it holds only text and links, else a default of 100.
func flexBasis(child *LayoutBox, contentWidth float64) float64 {
	if b := child.Style.Width.Resolve(contentWidth); b != 0 {
		return b
	}
	if w := inlineContentWidth(child); w > 0 {
		return w
	}
	return 100
}

// inlineContentWidth measures an element whose children are only text or text
// links, as a browser does for an auto-width flex item. Consecutive links are
// separated by a gap of 2 cells. It returns 0 for any other content.
func inlineContentWidth(child *LayoutBox) float64 {
	if child.Node == nil || child.Node.Type != dom.ElementNode || len(child.Children) == 0 {
		return 0
	}
	var width float64
	prevLink := false
	for _, c := range child.Children {
		if c.Node == nil {
			return 0
		}
		link := isTextLink(c.Node)
		if c.Node.Type != dom.TextNode && !link {
			return 0
		}
		if link && prevLink {
			width += 2 * cellmetrics.WidthPx
		}
		if text := strings.Join(collectInlineText(c.Node), " "); strings.TrimSpace(text) != "" {
			width += float64(runewidth.StringWidth(strings.Join(strings.Fields(text), " "))) * cellmetrics.WidthPx
		}
		prevLink = link
	}
	return width
}

func flexTotals(box *LayoutBox, contentWidth float64) (basis, grow, shrink float64) {
	for _, child := range box.Children {
		b := flexBasis(child, contentWidth)
		basis += b
		grow += child.Style.FlexGrow
		shrink += child.Style.FlexShrink
	}
	return
}

// flexMinContentWidth returns the widest default field (textarea cols or input
// size, 20 columns when unset) nested inside box. It is the minimum content
// width a flex item keeps while shrinking.
func flexMinContentWidth(box *LayoutBox) float64 {
	var widest float64
	var walk func(b *LayoutBox)
	walk = func(b *LayoutBox) {
		if b.Node != nil && b.Node.Type == dom.ElementNode {
			switch b.Node.TagName {
			case "textarea":
				widest = max(widest, fieldColumns(b.Node.GetAttr("cols")))
			case "input":
				if strings.EqualFold(strings.TrimSpace(b.Node.GetAttr("type")), "hidden") {
					break
				}
				if w := b.Style.Width; !w.IsAuto && !w.IsPercent && w.Px > 0 {
					// An explicit CSS width (google.com's .Ucigb, 458px) replaces the size default.
					widest = max(widest, w.Px/cellmetrics.WidthPx)
				} else {
					widest = max(widest, fieldColumns(b.Node.GetAttr("size")))
				}
			}
		}
		for _, c := range b.Children {
			walk(c)
		}
	}
	walk(box)
	return widest
}

const defaultFieldColumns = 20

func fieldColumns(attr string) float64 {
	n, err := strconv.Atoi(strings.TrimSpace(attr))
	if err != nil || n <= 0 {
		return defaultFieldColumns
	}
	return float64(n)
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
		basis := flexBasis(child, contentWidth)
		if child.Style.FlexShrink > 0 {
			total += basis * child.Style.FlexShrink
		}
	}
	return total
}

func layoutFlexChild(box, child *LayoutBox, flexDir css.FlexDirection, gap, contentWidth, pos, remaining, totalGrow, totalShrinkWeight float64) float64 {
	basis := flexBasis(child, contentWidth)

	// Apply flex grow/shrink
	childWidth := basis
	if totalGrow > 0 && remaining > 0 {
		childWidth += remaining * child.Style.FlexGrow / totalGrow
	} else if remaining < 0 && totalShrinkWeight > 0 && child.Style.FlexShrink > 0 {
		childWidth = max(0, basis+remaining*(basis*child.Style.FlexShrink)/totalShrinkWeight)
	}

	if flexDir == css.FlexDirectionRow {
		// min-width:auto keeps a flex item at least as wide as its minimum
		// content, so a shrunken row cannot squeeze a text field to nothing.
		childWidth = max(childWidth, flexMinContentWidth(child))
		if mw := child.Style.MinWidth; !mw.IsAuto {
			childWidth = max(childWidth, mw.Resolve(contentWidth))
		}
		if fit, ok := flexFitted[child]; ok {
			childWidth = fit
		}
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
