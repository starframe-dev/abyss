package js

/*
#include <JavaScriptCore/JavaScriptCore.h>

extern JSValueRef _makeString(JSContextRef ctx, const char* value);
*/
import "C"
import (
    "github.com/starframe/abyss/internal/css"
    "github.com/starframe/abyss/internal/dom"
)

// setComputedStyles stores the computed styles for getComputedStyle bridge.
func setComputedStyles(styles map[*dom.Node]css.ComputedStyle) {
    computedStylesMu.Lock()
    defer computedStylesMu.Unlock()
    computedStyles = styles
}

// SetComputedStyles stores computed styles for the DOM bridge.
func SetComputedStyles(styles map[*dom.Node]css.ComputedStyle) {
    setComputedStyles(styles)
}

// setLayoutBoxes stores the layout boxes for getBoundingClientRect.
func setLayoutBoxes(boxes map[*dom.Node]Rect) {
    layoutBoxesMu.Lock()
    defer layoutBoxesMu.Unlock()
    layoutBoxes = boxes
}

// SetLayoutBoxes stores layout boxes for the DOM bridge.
func SetLayoutBoxes(boxes map[*dom.Node]Rect) {
    setLayoutBoxes(boxes)
}

// getBoundingClientRect returns the bounding rect of a node.
func getBoundingClientRect(nodeID NodeID) Rect {
    layoutBoxesMu.Lock()
    defer layoutBoxesMu.Unlock()

    if layoutBoxes == nil {
        return Rect{}
    }

    nodeMu.Lock()
    n, ok := nodes[nodeID]
    nodeMu.Unlock()
    if !ok || n == nil {
        return Rect{}
    }

    rect, ok := layoutBoxes[n]
    if !ok {
        return Rect{}
    }
    return rect
}

// getComputedStyle returns a JS object with computed CSS properties for a node.
func getComputedStyle(nodeID NodeID) C.JSValueRef {
    computedStylesMu.Lock()
    defer computedStylesMu.Unlock()

    if computedStyles == nil {
        return C.JSValueMakeNull(eventCtx)
    }

    nodeMu.Lock()
    n, ok := nodes[nodeID]
    nodeMu.Unlock()
    if !ok || n == nil {
        return C.JSValueMakeNull(eventCtx)
    }

    style, ok := computedStyles[n]
    if !ok {
        return C.JSValueMakeNull(eventCtx)
    }

    ctx := eventCtx
    obj := C.JSObjectMake(ctx, nil, nil)

    setStr := func(name, val string) {
        pn := C.JSStringCreateWithUTF8CString(C.CString(name))
        defer C.JSStringRelease(pn)
        C.JSObjectSetProperty(ctx, obj, pn, C._makeString(ctx, C.CString(val)), C.kJSPropertyAttributeNone, nil)
    }

    setNum := func(name string, val float64) {
        pn := C.JSStringCreateWithUTF8CString(C.CString(name))
        defer C.JSStringRelease(pn)
        C.JSObjectSetProperty(ctx, obj, pn, C.JSValueMakeNumber(ctx, C.double(val)), C.kJSPropertyAttributeNone, nil)
    }

    setStr("display", displayString(style.Display))
    setStr("position", positionString(style.Position))
    setStr("overflow", overflowString(style.Overflow))

    setNum("width", style.Width.Px)
    setStr("widthUnit", style.Width.String())
    setNum("height", style.Height.Px)
    setStr("heightUnit", style.Height.String())

    setStr("color", style.Color.String())
    setStr("backgroundColor", style.BackgroundColor.String())

    setNum("marginTop", style.Margin.Top.Px)
    setNum("marginRight", style.Margin.Right.Px)
    setNum("marginBottom", style.Margin.Bottom.Px)
    setNum("marginLeft", style.Margin.Left.Px)

    setNum("paddingTop", style.Padding.Top.Px)
    setNum("paddingRight", style.Padding.Right.Px)
    setNum("paddingBottom", style.Padding.Bottom.Px)
    setNum("paddingLeft", style.Padding.Left.Px)

    setNum("borderTopWidth", style.BorderTop.Width)
    setNum("borderRightWidth", style.BorderRight.Width)
    setNum("borderBottomWidth", style.BorderBottom.Width)
    setNum("borderLeftWidth", style.BorderLeft.Width)

    setStr("flexDirection", flexDirectionString(style.FlexDirection))
    setStr("justifyContent", justifyContentString(style.JustifyContent))
    setStr("alignItems", alignItemsString(style.AlignItems))

    setNum("flexGrow", style.FlexGrow)
    setNum("flexShrink", style.FlexShrink)

    setNum("zIndex", float64(style.ZIndex))
    setNum("gap", style.Gap.Px)

    return jsObjToVal(obj)
}

func displayString(d css.DisplayType) string {
    switch d {
    case css.DisplayInline:
        return "inline"
    case css.DisplayFlex:
        return "flex"
    case css.DisplayGrid:
        return "grid"
    case css.DisplayNone:
        return "none"
    default:
        return "block"
    }
}

func positionString(p css.PositionType) string {
    switch p {
    case css.PositionRelative:
        return "relative"
    case css.PositionAbsolute:
        return "absolute"
    default:
        return "static"
    }
}

func overflowString(o css.OverflowType) string {
    if o == css.OverflowHidden {
        return "hidden"
    }
    return "visible"
}

func flexDirectionString(fd css.FlexDirection) string {
    if fd == css.FlexDirectionColumn {
        return "column"
    }
    return "row"
}

func justifyContentString(j css.JustifyContent) string {
    switch j {
    case css.JustifyCenter:
        return "center"
    case css.JustifyFlexEnd:
        return "flex-end"
    case css.JustifySpaceBetween:
        return "space-between"
    case css.JustifySpaceAround:
        return "space-around"
    default:
        return "flex-start"
    }
}

func alignItemsString(a css.AlignItems) string {
    switch a {
    case css.AlignCenter:
        return "center"
    case css.AlignFlexEnd:
        return "flex-end"
    case css.AlignStretch:
        return "stretch"
    default:
        return "flex-start"
    }
}
