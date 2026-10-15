package js

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
)

func TestStyleStringConversions(t *testing.T) {
	if displayString(css.DisplayInline) != "inline" || displayString(css.DisplayInlineBlock) != "inline-block" || displayString(css.DisplayFlex) != "flex" || displayString(css.DisplayGrid) != "grid" || displayString(css.DisplayNone) != "none" || displayString(css.DisplayType(255)) != "block" { t.Fatal("display conversion") }
	if positionString(css.PositionRelative) != "relative" || positionString(css.PositionAbsolute) != "absolute" || positionString(css.PositionType(255)) != "static" { t.Fatal("position conversion") }
	if overflowString(css.OverflowHidden) != "hidden" || overflowString(css.OverflowType(255)) != "visible" { t.Fatal("overflow conversion") }
	if flexDirectionString(css.FlexDirectionColumn) != "column" || flexDirectionString(css.FlexDirection(255)) != "row" { t.Fatal("flex direction conversion") }
	if justifyContentString(css.JustifyCenter) != "center" || justifyContentString(css.JustifyFlexEnd) != "flex-end" || justifyContentString(css.JustifySpaceBetween) != "space-between" || justifyContentString(css.JustifySpaceAround) != "space-around" || justifyContentString(css.JustifyContent(255)) != "flex-start" { t.Fatal("justify conversion") }
	if alignItemsString(css.AlignCenter) != "center" || alignItemsString(css.AlignFlexEnd) != "flex-end" || alignItemsString(css.AlignStretch) != "stretch" || alignItemsString(css.AlignItems(255)) != "flex-start" { t.Fatal("align conversion") }
}

func TestComputedStylesAndLayoutBoxes(t *testing.T) {
	defer resetDOMState()
	InitDOM(dom.Parse("<html><body><div id='target'></div></body></html>"))
	node := getElementByID("target")
	id := getNodeID(node)
	SetLayoutBoxes(map[*dom.Node]Rect{node: {X: 1, Y: 2, Width: 30, Height: 40}})
	if got := getBoundingClientRect(id); got != (Rect{X: 1, Y: 2, Width: 30, Height: 40}) { t.Fatalf("rectangle = %+v", got) }
	if got := getBoundingClientRect(-100); got != (Rect{}) { t.Fatalf("invalid rectangle = %+v", got) }
	SetLayoutBoxes(nil)
	if getBoundingClientRect(id) != (Rect{}) { t.Fatal("nil layout should yield zero rectangle") }
	SetComputedStyles(nil)
	if getComputedStyle(id) != nil { t.Fatal("nil styles should yield null") }
}
