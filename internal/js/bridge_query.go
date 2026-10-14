package js

/*
#include <JavaScriptCore/JavaScriptCore.h>
#include <stdlib.h>

// Declare functions from bridge.c
extern JSObjectRef makeGoFunction(JSContextRef ctx, const char* funcName);
extern JSType _getType(JSContextRef ctx, JSValueRef value);

extern char* _valueToString(JSContextRef ctx, JSValueRef value, JSValueRef* exception);
extern JSValueRef _makeArray(JSContextRef ctx, JSValueRef* values, size_t count, JSValueRef* exception);
extern void _fprintf_stderr(const char* msg);
extern void _protectValue(JSContextRef ctx, JSValueRef value);
extern void _unprotectValue(JSContextRef ctx, JSValueRef value);
extern JSValueRef _callAsFunction(JSContextRef ctx, JSObjectRef func, JSObjectRef thisObj, size_t argCount, const JSValueRef args[], JSValueRef* exception);
extern JSValueRef _makeString(JSContextRef ctx, const char* value);
extern JSValueRef _makeBoolean(JSContextRef ctx, bool value);
*/
import "C"

func getBodyJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
    body := getBody()
    if body == nil {
        return ctx.MakeNull()
    }
    return nodeToJS(ctx, body)
}

func getElementByIdJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
    if len(args) < 1 {
        return ctx.MakeNull()
    }
    id, _ := getArgString(ctx, args[0])
    node := getElementByID(id)
    if node == nil {
        return ctx.MakeNull()
    }
    return nodeToJS(ctx, node)
}

func querySelectorJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
    if len(args) < 2 {
        return ctx.MakeNull()
    }
    id := jsToNodeID(ctx, args[0])
    if id < 0 {
        return ctx.MakeNull()
    }
    sel, _ := getArgString(ctx, args[1])
    node := querySelector(id, sel)
    if node == nil {
        return ctx.MakeNull()
    }
    return nodeToJS(ctx, node)
}

func querySelectorAllJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
    if len(args) < 2 {
        return ctx.MakeUndefined()
    }
    id := jsToNodeID(ctx, args[0])
    if id < 0 {
        return ctx.MakeUndefined()
    }
    sel, _ := getArgString(ctx, args[1])
    nodes := querySelectorAll(id, sel)
    childValues := make([]C.JSValueRef, len(nodes))
    for i, n := range nodes {
        childValues[i] = nodeToJS(ctx, n)
    }
    var exc C.JSValueRef
    if len(childValues) == 0 {
        return C._makeArray(ctx.ctx, nil, 0, &exc)
    }
    return C._makeArray(ctx.ctx, &childValues[0], C.size_t(len(childValues)), &exc)
}

func getAttributeJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
    if len(args) < 2 {
        return ctx.MakeNull()
    }
    id := jsToNodeID(ctx, args[0])
    if id < 0 {
        return ctx.MakeNull()
    }
    name, _ := getArgString(ctx, args[1])
    val := getAttribute(id, name)
    return ctx.MakeString(val)
}

func getParentNodeJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
    if len(args) < 1 {
        return ctx.MakeNull()
    }
    id := jsToNodeID(ctx, args[0])
    if id < 0 {
        return ctx.MakeNull()
    }
    parent := getParentNode(id)
    if parent == nil {
        return ctx.MakeNull()
    }
    return nodeToJS(ctx, parent)
}

func getChildrenJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
    if len(args) < 1 {
        return ctx.MakeUndefined()
    }
    id := jsToNodeID(ctx, args[0])
    if id < 0 {
        return ctx.MakeUndefined()
    }
    children := getChildren(id)
    childValues := make([]C.JSValueRef, len(children))
    for i, child := range children {
        childValues[i] = nodeToJS(ctx, child)
    }
    var exc C.JSValueRef
    if len(childValues) == 0 {
        return C._makeArray(ctx.ctx, nil, 0, &exc)
    }
    return C._makeArray(ctx.ctx, &childValues[0], C.size_t(len(childValues)), &exc)
}

func getElementsByTagNameJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
    if len(args) < 1 {
        return ctx.MakeUndefined()
    }
    tagName, _ := getArgString(ctx, args[0])
    nodes := getElementsByTagName(tagName)
    childValues := make([]C.JSValueRef, len(nodes))
    for i, n := range nodes {
        childValues[i] = nodeToJS(ctx, n)
    }
    var exc C.JSValueRef
    if len(childValues) == 0 {
        return C._makeArray(ctx.ctx, nil, 0, &exc)
    }
    return C._makeArray(ctx.ctx, &childValues[0], C.size_t(len(childValues)), &exc)
}

func getElementsByClassNameJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
    if len(args) < 1 {
        return ctx.MakeUndefined()
    }
    className, _ := getArgString(ctx, args[0])
    nodes := getElementsByClassName(className)
    childValues := make([]C.JSValueRef, len(nodes))
    for i, n := range nodes {
        childValues[i] = nodeToJS(ctx, n)
    }
    var exc C.JSValueRef
    if len(childValues) == 0 {
        return C._makeArray(ctx.ctx, nil, 0, &exc)
    }
    return C._makeArray(ctx.ctx, &childValues[0], C.size_t(len(childValues)), &exc)
}

func matchesJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
    if len(args) < 2 {
        return ctx.MakeBoolean(false)
    }
    id := jsToNodeID(ctx, args[0])
    if id < 0 {
        return ctx.MakeBoolean(false)
    }
    selector, err := getArgString(ctx, args[1])
    if err != nil {
        return ctx.MakeBoolean(false)
    }
    return ctx.MakeBoolean(matches(id, selector))
}

func closestJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
    if len(args) < 2 {
        return ctx.MakeNull()
    }
    id := jsToNodeID(ctx, args[0])
    if id < 0 {
        return ctx.MakeNull()
    }
    selector, err := getArgString(ctx, args[1])
    if err != nil {
        return ctx.MakeNull()
    }
    node := closest(id, selector)
    if node == nil {
        return ctx.MakeNull()
    }
    return nodeToJS(ctx, node)
}

func previousElementSiblingJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
    if len(args) < 1 {
        return ctx.MakeNull()
    }
    id := jsToNodeID(ctx, args[0])
    if id < 0 {
        return ctx.MakeNull()
    }
    node := previousElementSibling(id)
    if node == nil {
        return ctx.MakeNull()
    }
    return nodeToJS(ctx, node)
}

func nextElementSiblingJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
    if len(args) < 1 {
        return ctx.MakeNull()
    }
    id := jsToNodeID(ctx, args[0])
    if id < 0 {
        return ctx.MakeNull()
    }
    node := nextElementSibling(id)
    if node == nil {
        return ctx.MakeNull()
    }
    return nodeToJS(ctx, node)
}

func firstElementChildJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
    if len(args) < 1 {
        return ctx.MakeNull()
    }
    id := jsToNodeID(ctx, args[0])
    if id < 0 {
        return ctx.MakeNull()
    }
    node := firstElementChild(id)
    if node == nil {
        return ctx.MakeNull()
    }
    return nodeToJS(ctx, node)
}

func lastElementChildJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
    if len(args) < 1 {
        return ctx.MakeNull()
    }
    id := jsToNodeID(ctx, args[0])
    if id < 0 {
        return ctx.MakeNull()
    }
    node := lastElementChild(id)
    if node == nil {
        return ctx.MakeNull()
    }
    return nodeToJS(ctx, node)
}
