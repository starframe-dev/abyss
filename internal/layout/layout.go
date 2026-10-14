// Package layout computes positions and sizes of DOM nodes using CSS box model.
package layout

import (
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/starframe-dev/abyss/internal/asciirender"
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

	// Compute box model
	marginLeft := style.Margin.Left.Resolve(containerWidth)
	marginRight := style.Margin.Right.Resolve(containerWidth)
	marginTop := style.Margin.Top.Resolve(containerWidth)
	marginBottom := style.Margin.Bottom.Resolve(containerWidth)
	paddingLeft := style.Padding.Left.Resolve(containerWidth)
	paddingRight := style.Padding.Right.Resolve(containerWidth)
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
		switch style.Display {
		case css.DisplayFlex:
			contentHeight = layoutFlex(box, contentWidth)
		case css.DisplayGrid:
			contentHeight = layoutGrid(box, contentWidth)
		case css.DisplayBlock:
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
	height := float64(len(lines)) * 2.0
	if height < defaultImageHeight {
		height = defaultImageHeight
	}
	return height
}

func layoutBlockFlow(box *LayoutBox, contentWidth float64) float64 {
	var y float64
	for _, child := range box.Children {
		layoutBox(child, contentWidth, box.Content.X, box.Content.Y+y, false)
		y += child.Margin.Height
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
			box.Content.Height = 2
			return 2
		}
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
	return float64(len(lines)) * 2.0
}

// inputLabel returns the text shown for a form input: its value, falling back to placeholder or title.
func inputLabel(n *dom.Node) string {
	for _, attr := range []string{"value", "placeholder", "title"} {
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
	lineHeight := 2.0 // 1 terminal row = 2 CSS pixels (via █▄▀ block chars)
	var lines []TextLine
	var currentLine string
	var x float64

	for _, word := range words {
		wordWidth := float64(runewidth.StringWidth(word))
		if currentLine != "" {
			if x+1+wordWidth > contentWidth {
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
	return float64(runewidth.StringWidth(text))
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
