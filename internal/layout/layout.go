// Package layout computes positions and sizes of DOM nodes using CSS box model.
package layout

import (
	"strconv"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/starframe-dev/abyss/internal/asciirender"
	"github.com/starframe-dev/abyss/internal/cellmetrics"
	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
)

// ViewportWidth is the default viewport width in pixels.
const ViewportWidth = 800

// Rect represents a rectangle with position and size.
type Rect struct {
	X, Y, Width, Height float64
}

// TextLine is a line of text with its position.
type TextLine struct {
	Text string
	X, Y float64
}

// LayoutBox represents a laid-out element.
type LayoutBox struct {
	Node      *dom.Node
	Parent    *LayoutBox
	Style     css.ComputedStyle
	Content   Rect // content area (relative to parent content area)
	Padding   Rect // padding area
	Border    Rect // border area
	Margin    Rect // margin area
	Children  []*LayoutBox
	TextLines []TextLine
	// AsciiCells holds the pre-quantized ASCII grid for <img> boxes
	// populated by render.Render when the image has been decoded.
	AsciiCells [][]asciirender.Cell
}

// Build creates a layout tree from a DOM document and computed styles.
func Build(doc *dom.Node, styles map[*dom.Node]css.ComputedStyle, viewportWidth float64) *LayoutBox {
	root := buildNode(doc, styles)
	if root != nil {
		Layout(root, viewportWidth)
	}
	return root
}

func buildNode(node *dom.Node, styles map[*dom.Node]css.ComputedStyle) *LayoutBox {
	if node.Type == dom.TextNode {
		return nil
	}

	// Skip style, script, and head elements - they should not produce layout boxes
	if node.Type == dom.ElementNode && (node.TagName == "style" || node.TagName == "script" || node.TagName == "head") {
		return nil
	}

	style, ok := styles[node]
	if !ok {
		style = css.Default()
	}

	if style.Display == css.DisplayNone {
		return nil
	}

	box := &LayoutBox{
		Node:     node,
		Style:    style,
		Children: []*LayoutBox{},
	}

	hasElementChild := false
	for _, child := range node.Children {
		if child.Type == dom.ElementNode {
			hasElementChild = true
			break
		}
	}

	for _, child := range node.Children {
		if child.Type == dom.TextNode {
			// Text next to element children needs its own anonymous inline box,
			// otherwise it is dropped. Text-only containers are handled by layoutInlineFlow.
			if hasElementChild && strings.TrimSpace(child.Text) != "" {
				box.Children = append(box.Children, anonymousTextBox(box, child))
			}
			continue
		}
		childBox := buildNode(child, styles)
		if childBox != nil {
			childBox.Parent = box
			box.Children = append(box.Children, childBox)
		}
	}

	return box
}

// anonymousTextBox wraps a text node in a box that inherits the parent's text
// styling but none of its box model, so it does not repeat padding or borders.
func anonymousTextBox(parent *LayoutBox, text *dom.Node) *LayoutBox {
	def := css.Default()
	style := parent.Style
	style.Display = css.DisplayInline
	style.Width = def.Width
	style.Height = def.Height
	style.Margin = def.Margin
	style.Padding = def.Padding
	style.BorderTop = def.BorderTop
	style.BorderRight = def.BorderRight
	style.BorderBottom = def.BorderBottom
	style.BorderLeft = def.BorderLeft
	style.BackgroundColor = def.BackgroundColor
	return &LayoutBox{Node: text, Style: style, Children: []*LayoutBox{}, Parent: parent}
}

// Layout computes the geometry of the layout tree.
func Layout(box *LayoutBox, containerWidth float64) {
	layoutBox(box, containerWidth, 0, 0, true)
}

func layoutBox(box *LayoutBox, containerWidth, offsetX, offsetY float64, isRoot bool) {
	style := box.Style

	// Resolve width
	width := resolveWidth(style, containerWidth)
	if style.Width.IsAuto && box.Node.Type == dom.ElementNode && box.Node.TagName == "img" {
		// An image without CSS width takes its width attribute (CSS px) instead of the full container.
		if attr, err := strconv.ParseFloat(strings.TrimSpace(box.Node.GetAttr("width")), 64); err == nil && attr > 0 {
			width = attr
		}
	}

	// Compute box model
	marginLeft := style.Margin.Left.Resolve(containerWidth)
	marginRight := style.Margin.Right.Resolve(containerWidth)
	marginTop := style.Margin.Top.Resolve(containerWidth)
	marginBottom := style.Margin.Bottom.Resolve(containerWidth)
	paddingLeft := style.Padding.Left.Resolve(containerWidth)
	paddingRight := style.Padding.Right.Resolve(containerWidth)
	if style.Width.IsAuto && isButtonInput(box.Node) {
		// Buttons size to their label plus horizontal padding, not to the container.
		width = buttonInputWidth(box.Node) + paddingLeft + paddingRight
	}
	paddingTop := style.Padding.Top.Resolve(containerWidth)
	paddingBottom := style.Padding.Bottom.Resolve(containerWidth)
	borderLeft := style.BorderLeft.Width
	borderRight := style.BorderRight.Width
	borderTop := style.BorderTop.Width
	borderBottom := style.BorderBottom.Width

	// Content width
	contentWidth := width - paddingLeft - paddingRight - borderLeft - borderRight
	if contentWidth < 0 {
		contentWidth = 0
	}

	// Position. The offset is the margin-box origin. Build the box model
	// from the outside in so padding expands inward instead of overlapping
	// preceding siblings.
	marginX := offsetX
	marginY := offsetY
	borderX := marginX + marginLeft
	borderY := marginY + marginTop

	if isRoot {
		marginX = 0
		marginY = 0
		borderX = marginLeft
		borderY = marginTop
	}

	if style.Position == css.PositionAbsolute {
		// Find positioned ancestor
		ancestor := findPositionedAncestor(box)
		if ancestor != nil {
			refX, refY := ancestor.Content.X, ancestor.Content.Y
			if !style.Top.IsAuto {
				borderY = refY + style.Top.Resolve(ancestor.Content.Height)
			}
			if !style.Left.IsAuto {
				borderX = refX + style.Left.Resolve(ancestor.Content.Width)
			}
			if !style.Right.IsAuto && style.Left.IsAuto {
				borderX = refX + ancestor.Content.Width - width - style.Right.Resolve(ancestor.Content.Width)
			}
			if !style.Bottom.IsAuto && style.Top.IsAuto {
				borderHeight := resolveHeight(style, 0) + paddingTop + paddingBottom + borderTop + borderBottom
				borderY = refY + ancestor.Content.Height - borderHeight - style.Bottom.Resolve(ancestor.Content.Height)
			}
			marginX = borderX - marginLeft
			marginY = borderY - marginTop
		}
	}

	paddingX := borderX + borderLeft
	paddingY := borderY + borderTop
	contentX := paddingX + paddingLeft
	contentY := paddingY + paddingTop
	paddingWidth := contentWidth + paddingLeft + paddingRight
	borderWidth := paddingWidth + borderLeft + borderRight

	box.Content = Rect{X: contentX, Y: contentY, Width: contentWidth, Height: 0}
	box.Padding = Rect{X: paddingX, Y: paddingY, Width: paddingWidth, Height: 0}
	box.Border = Rect{X: borderX, Y: borderY, Width: borderWidth, Height: 0}
	box.Margin = Rect{
		X:      marginX,
		Y:      marginY,
		Width:  borderWidth + marginLeft + marginRight,
		Height: 0,
	}

	// Layout children
	var contentHeight float64
	if box.Node.Type == dom.ElementNode && box.Node.TagName == "img" {
		contentHeight = layoutImage(box, contentWidth)
	} else {
		switch {
		case box.Node.Type == dom.ElementNode && box.Node.TagName == "table":
			contentHeight = layoutTable(box, contentWidth)
		case style.Display == css.DisplayFlex:
			contentHeight = layoutFlex(box, contentWidth)
		case style.Display == css.DisplayGrid:
			contentHeight = layoutGrid(box, contentWidth)
		case style.Display == css.DisplayBlock || style.Display == css.DisplayInlineBlock:
			// Inline-block boxes lay out their content as blocks; siblings share a row via layoutBlockFlow.
			contentHeight = layoutBlockFlow(box, contentWidth)
		default:
			// Inline layout
			contentHeight = layoutInlineFlow(box, contentWidth)
		}
	}

	// Resolve height
	height := contentHeight
	if !style.Height.IsAuto {
		height = style.Height.Resolve(contentHeight)
	}

	box.Content.Height = height
	box.Padding.Height = height + paddingTop + paddingBottom
	box.Border.Height = box.Padding.Height + borderTop + borderBottom
	box.Margin.Height = box.Border.Height + marginTop + marginBottom
}

func resolveWidth(style css.ComputedStyle, containerWidth float64) float64 {
	if !style.Width.IsAuto {
		return style.Width.Resolve(containerWidth)
	}
	// Auto width: fill container minus margins
	marginLeft := style.Margin.Left.Resolve(containerWidth)
	marginRight := style.Margin.Right.Resolve(containerWidth)
	w := containerWidth - marginLeft - marginRight
	if w < 0 {
		w = 0
	}
	return w
}

func resolveHeight(style css.ComputedStyle, contentHeight float64) float64 {
	if !style.Height.IsAuto {
		return style.Height.Resolve(contentHeight)
	}
	return contentHeight
}

const defaultImageHeight = 48.0

func layoutImage(box *LayoutBox, contentWidth float64) float64 {
	label := strings.TrimSpace(box.Node.GetAttr("alt"))
	if label == "" {
		label = "image"
	}
	lines := wordWrap([]string{"[image]", label}, box, contentWidth)
	box.TextLines = lines
	height := float64(len(lines)) * cellmetrics.HeightPx
	if h, err := strconv.ParseFloat(strings.TrimSpace(box.Node.GetAttr("height")), 64); err == nil && h > 0 {
		// The height attribute (CSS px) sets the picture's real row count.
		height = h
	}
	if height < defaultImageHeight {
		height = defaultImageHeight
	}
	return height
}

func layoutBlockFlow(box *LayoutBox, contentWidth float64) float64 {
	var y float64
	children := box.Children
	for len(children) > 0 {
		// A run of consecutive buttons shares lines; other children stack as blocks.
		if isRowItem(children[0]) {
			n := 1
			for n < len(children) && isRowItem(children[n]) {
				n++
			}
			y += layoutButtonRow(box, children[:n], contentWidth, y)
			children = children[n:]
			continue
		}
		layoutBox(children[0], contentWidth, box.Content.X, box.Content.Y+y, false)
		y += children[0].Margin.Height
		children = children[1:]
	}
	// If no block children, process inline content
	if y == 0 {
		return layoutInlineFlow(box, contentWidth)
	}
	return y
}

func layoutInlineFlow(box *LayoutBox, contentWidth float64) float64 {
	// Skip text content inside <style> and <script> tags
	if inStyleOrScript(box.Node) {
		return 0
	}

	textParts := collectInlineText(box.Node)
	if box.Node.Type == dom.TextNode {
		textParts = []string{strings.TrimSpace(box.Node.Text)}
	}
	if box.Node.Type == dom.ElementNode && box.Node.TagName == "input" {
		if strings.EqualFold(strings.TrimSpace(box.Node.GetAttr("type")), "hidden") {
			return 0
		}
		textParts = []string{inputLabel(box.Node)}
		if textParts[0] == "" {
			// An empty field still occupies one line so it stays visible.
			box.Content.Height = cellmetrics.HeightPx
			return cellmetrics.HeightPx
		}
	}
	if box.Node.Type == dom.ElementNode && box.Node.TagName == "textarea" && strings.TrimSpace(strings.Join(textParts, "")) == "" {
		// An empty textarea still occupies one line so it stays visible.
		box.Content.Height = cellmetrics.HeightPx
		return cellmetrics.HeightPx
	}
	if len(textParts) == 0 {
		return 0
	}

	// Word wrap - concatenate directly, original HTML has correct spacing
	allText := strings.Join(textParts, "")
	words := strings.Fields(allText)
	if len(words) == 0 {
		return 0
	}

	lines := wordWrap(words, box, contentWidth)
	box.TextLines = lines
	return float64(len(lines)) * cellmetrics.HeightPx
}

// buttonOf returns the push button a child stands for: the child itself, or the
// only child of a plain wrapper such as <span>. It returns nil for anything else.
func buttonOf(child *LayoutBox) *LayoutBox {
	if child == nil || child.Node == nil {
		return nil
	}
	if isButtonInput(child.Node) {
		return child
	}
	if child.Node.Type == dom.ElementNode && child.Node.TagName == "span" {
		// Only one child may draw anything; scripts, styles and hidden inputs draw nothing.
		var only *LayoutBox
		for _, c := range child.Children {
			if !drawsNothing(c) {
				if only != nil {
					return nil
				}
				only = c
			}
		}
		if only != nil {
			return buttonOf(only)
		}
	}
	return nil
}

// drawsNothing reports whether a box adds no visible content to its parent.
func drawsNothing(b *LayoutBox) bool {
	if b == nil || b.Node == nil {
		return true
	}
	if b.Node.Type == dom.TextNode {
		return strings.TrimSpace(b.Node.Text) == ""
	}
	if b.Node.Type != dom.ElementNode {
		return false
	}
	switch b.Node.TagName {
	case "script", "style":
		return true
	case "input":
		return strings.EqualFold(strings.TrimSpace(b.Node.GetAttr("type")), "hidden")
	}
	return false
}

// isRowItem reports whether a block child sits on a shared line with its siblings:
// a push button (or a wrapper of one) or a text-only link, as in a browser.
func isRowItem(child *LayoutBox) bool {
	if child.Node == nil {
		return false
	}
	if child.Node.Type == dom.TextNode {
		return strings.TrimSpace(child.Node.Text) != ""
	}
	return buttonOf(child) != nil || isTextLink(child.Node)
}

// isTextLink reports whether n is an <a> whose content is only text.
func isTextLink(n *dom.Node) bool {
	if n == nil || n.Type != dom.ElementNode || n.TagName != "a" {
		return false
	}
	for _, c := range n.Children {
		if c.Type != dom.TextNode {
			return false
		}
	}
	return true
}

// rowItemWidth returns the width a row item takes on its line.
func rowItemWidth(child *LayoutBox, contentWidth float64) float64 {
	if btn := buttonOf(child); btn != nil {
		// Measure at the origin: a wrapper span without a width spans the container.
		layoutBox(child, contentWidth, 0, 0, false)
		return btn.Margin.Width
	}
	if child.Node.Type == dom.TextNode {
		// Keep one cell for each edge space so text next to links keeps its gap.
		width := float64(runewidth.StringWidth(strings.Join(strings.Fields(child.Node.Text), " ")))
		if text := child.Node.Text; text != strings.TrimLeft(text, " \t\r\n") {
			width++
		}
		if text := child.Node.Text; text != strings.TrimRight(text, " \t\r\n") {
			width++
		}
		return width * cellmetrics.WidthPx
	}
	return float64(runewidth.StringWidth(strings.Join(collectInlineText(child.Node), ""))) * cellmetrics.WidthPx
}

// placeRowItem lays a row item out at x,y with the given width.
func placeRowItem(child *LayoutBox, width, contentWidth, x, y float64) {
	if buttonOf(child) != nil {
		layoutBox(child, contentWidth, x, y, false)
		return
	}
	width = min(width, contentWidth)
	orig := child.Style.Width
	child.Style.Width = css.Px(width)
	layoutBox(child, contentWidth, x, y, false)
	child.Style.Width = orig
}

// rowAlign returns how a row of items is aligned inside box: <center>, align="center"
// on a cell or block, or the block's own text-align.
func rowAlign(box *LayoutBox) css.TextAlignType {
	if box.Node != nil && box.Node.Type == dom.ElementNode {
		if box.Node.TagName == "center" || strings.EqualFold(strings.TrimSpace(box.Node.GetAttr("align")), "center") {
			return css.TextAlignCenter
		}
	}
	return box.Style.TextAlign
}

// layoutButtonRow places row items left to right from yOrigin, wrapping to the next
// line when they run out of width. Each line is aligned by rowAlign. It returns the height used.
func layoutButtonRow(box *LayoutBox, items []*LayoutBox, contentWidth, yOrigin float64) float64 {
	widths := make([]float64, len(items))
	gaps := make([]float64, len(items))
	for i, child := range items {
		widths[i] = rowItemWidth(child, contentWidth)
		if i > 0 && isTextLink(items[i-1].Node) && isTextLink(child.Node) {
			// Browsers space adjacent links apart; the footer links rely on it.
			gaps[i] = 2 * cellmetrics.WidthPx
		}
	}
	align := rowAlign(box)
	var y float64
	for start := 0; start < len(items); {
		end := start
		var lineWidth float64
		for end < len(items) {
			step := widths[end]
			if end > start {
				step += gaps[end]
			}
			if end > start && lineWidth+step > contentWidth {
				break
			}
			lineWidth += step
			end++
		}
		var offset float64
		switch align {
		case css.TextAlignCenter:
			offset = (contentWidth - lineWidth) / 2
		case css.TextAlignRight:
			offset = contentWidth - lineWidth
		}
		offset = max(0, offset)
		var x, rowHeight float64
		for i := start; i < end; i++ {
			if i > start {
				x += gaps[i]
			}
			placeRowItem(items[i], widths[i], contentWidth, box.Content.X+offset+x, box.Content.Y+yOrigin+y)
			x += widths[i]
			rowHeight = max(rowHeight, items[i].Margin.Height)
		}
		y += rowHeight
		start = end
	}
	return y
}

// isButtonInput reports whether n is an input rendered as a push button.
func isButtonInput(n *dom.Node) bool {
	if n == nil || n.Type != dom.ElementNode || n.TagName != "input" {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(n.GetAttr("type"))) {
	case "submit", "button", "reset":
		return true
	}
	return false
}

// buttonInputWidth is the label width plus one cell of padding on each side.
func buttonInputWidth(n *dom.Node) float64 {
	return float64(runewidth.StringWidth(inputLabel(n))+2) * cellmetrics.WidthPx
}

// inputLabel returns the text shown for a form input: its value, falling back to placeholder.
// title is a hover tooltip and is never drawn as text.
func inputLabel(n *dom.Node) string {
	for _, attr := range []string{"value", "placeholder"} {
		if v := strings.TrimSpace(n.GetAttr(attr)); v != "" {
			return v
		}
	}
	return ""
}

func inStyleOrScript(n *dom.Node) bool {
	for p := n; p != nil; p = p.Parent {
		if p.Type == dom.ElementNode && (p.TagName == "style" || p.TagName == "script") {
			return true
		}
	}
	return false
}

func collectInlineText(n *dom.Node) []string {
	var textParts []string
	for _, child := range n.Children {
		if child.Type == dom.ElementNode && (child.TagName == "style" || child.TagName == "script") {
			// Skip style and script content
			continue
		}
		if child.Type == dom.TextNode {
			text := strings.TrimSpace(child.Text)
			if text != "" {
				textParts = append(textParts, text)
			}
		} else if child.Type == dom.ElementNode {
			// Inline elements: add their text, but skip script/style descendants
			text := collectVisibleText(child)
			if text != "" {
				textParts = append(textParts, text)
			}
		}
	}
	return textParts
}

func wordWrap(words []string, box *LayoutBox, contentWidth float64) []TextLine {
	lineHeight := float64(cellmetrics.HeightPx)
	contentCols := contentWidth / cellmetrics.WidthPx
	var lines []TextLine
	var currentLine string
	var x float64

	for _, word := range words {
		wordWidth := float64(runewidth.StringWidth(word))
		if currentLine != "" {
			if x+1+wordWidth > contentCols {
				lines = append(lines, TextLine{Text: currentLine, X: box.Content.X, Y: box.Content.Y + float64(len(lines))*lineHeight})
				currentLine = word
				x = wordWidth
				continue
			}
			currentLine += " "
			x++
		}
		currentLine += word
		x += wordWidth
	}
	if currentLine != "" {
		lines = append(lines, TextLine{Text: currentLine, X: box.Content.X, Y: box.Content.Y + float64(len(lines))*lineHeight})
	}
	return lines
}

func findPositionedAncestor(box *LayoutBox) *LayoutBox {
	current := box.Parent
	for current != nil {
		if current.Style.Position == css.PositionRelative || current.Style.Position == css.PositionAbsolute {
			return current
		}
		current = current.Parent
	}
	return nil
}

func estimateTextWidth(text string) float64 {
	return float64(runewidth.StringWidth(text)) * cellmetrics.WidthPx
}

// collectVisibleText returns text content skipping <script> and <style> descendants.
func collectVisibleText(n *dom.Node) string {
	if n.Type == dom.TextNode {
		return n.Text
	}
	if n.Type == dom.ElementNode && (n.TagName == "script" || n.TagName == "style") {
		return ""
	}
	var b strings.Builder
	for _, child := range n.Children {
		b.WriteString(collectVisibleText(child))
	}
	return b.String()
}
