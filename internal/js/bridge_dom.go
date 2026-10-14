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
import (
    "unsafe"

    "github.com/starframe/abyss/internal/css"
)

func createElementJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
    if len(args) < 1 {
        return C.JSValueMakeUndefined(ctx.ctx)
    }
    tagName, _ := getArgString(ctx, args[0])
    if tagName == "" {
        tagName = "div"
    }
    el := createElement(tagName)
    return nodeToJS(ctx, el)
}

func appendChildJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
    if len(args) < 2 {
        return ctx.MakeUndefined()
    }
    parentID := jsToNodeID(ctx, args[0])
    childID := jsToNodeID(ctx, args[1])
    if parentID < 0 || childID < 0 {
        return ctx.MakeUndefined()
    }
    appendChild(parentID, childID)
    css.InvalidateCache()
    SetDirty()
    return args[1]
}

func setInnerHTMLJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
    if len(args) < 2 {
        return ctx.MakeUndefined()
    }
    id := jsToNodeID(ctx, args[0])
    if id < 0 {
        return ctx.MakeUndefined()
    }
    html, _ := getArgString(ctx, args[1])
    setInnerHTML(id, html)
    css.InvalidateCache()
    SetDirty()
    return ctx.MakeUndefined()
}

func setAttributeJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
    if len(args) < 3 {
        return ctx.MakeUndefined()
    }
    id := jsToNodeID(ctx, args[0])
    if id < 0 {
        return ctx.MakeUndefined()
    }
    name, _ := getArgString(ctx, args[1])
    value, _ := getArgString(ctx, args[2])
    setAttribute(id, name, value)
    css.InvalidateCache()
    SetDirty()
    return ctx.MakeUndefined()
}

func setTextContentJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
    if len(args) < 2 {
        return ctx.MakeUndefined()
    }
    id := jsToNodeID(ctx, args[0])
    if id < 0 {
        return ctx.MakeUndefined()
    }
    text, _ := getArgString(ctx, args[1])
    setTextContent(id, text)
    SetDirty()
    return ctx.MakeUndefined()
}

func createTextNodeJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
    if len(args) < 1 {
        return ctx.MakeUndefined()
    }
    text, _ := getArgString(ctx, args[0])
    tn := createTextNode(text)
    return nodeToJS(ctx, tn)
}

func createDocumentFragmentJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
    frag := createDocumentFragment()
    return nodeToJS(ctx, frag)
}

func getInnerHTMLJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
    if len(args) < 1 {
        return ctx.MakeString("")
    }
    id := jsToNodeID(ctx, args[0])
    if id < 0 {
        return ctx.MakeString("")
    }
    html := getInnerHTML(id)
    return ctx.MakeString(html)
}

func getTextContentJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
    if len(args) < 1 {
        return ctx.MakeString("")
    }
    id := jsToNodeID(ctx, args[0])
    if id < 0 {
        return ctx.MakeString("")
    }
    text := getTextContent(id)
    return ctx.MakeString(text)
}

func removeChildJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
    if len(args) < 2 {
        return ctx.MakeUndefined()
    }
    parentID := jsToNodeID(ctx, args[0])
    childID := jsToNodeID(ctx, args[1])
    if parentID < 0 || childID < 0 {
        return ctx.MakeUndefined()
    }
    removeChild(parentID, childID)
    css.InvalidateCache()
    SetDirty()
    return args[1]
}

func insertBeforeJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
    if len(args) < 3 {
        return ctx.MakeUndefined()
    }
    parentID := jsToNodeID(ctx, args[0])
    newNodeID := jsToNodeID(ctx, args[1])
    refNodeID := jsToNodeID(ctx, args[2])
    if parentID < 0 || newNodeID < 0 {
        return ctx.MakeUndefined()
    }
    insertBefore(parentID, newNodeID, refNodeID)
    SetDirty()
    return args[1] // return the new node
}

func replaceChildJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
    if len(args) < 3 {
        return ctx.MakeUndefined()
    }
    parentID := jsToNodeID(ctx, args[0])
    newNodeID := jsToNodeID(ctx, args[1])
    oldNodeID := jsToNodeID(ctx, args[2])
    if parentID < 0 || newNodeID < 0 || oldNodeID < 0 {
        return ctx.MakeUndefined()
    }
    replaceChild(parentID, newNodeID, oldNodeID)
    SetDirty()
    return args[2] // return the removed node
}

func cloneNodeJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
    if len(args) < 1 {
        return ctx.MakeNull()
    }
    id := jsToNodeID(ctx, args[0])
    if id < 0 {
        return ctx.MakeNull()
    }
    deep := false
    if len(args) >= 2 {
        deep = bool(C.JSValueToBoolean(ctx.ctx, args[1]))
    }
    node := cloneNode(id, deep)
    if node == nil {
        return ctx.MakeNull()
    }
    return nodeToJS(ctx, node)
}

func containsJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
    if len(args) < 2 {
        return ctx.MakeBoolean(false)
    }
    parentID := jsToNodeID(ctx, args[0])
    childID := jsToNodeID(ctx, args[1])
    if parentID < 0 || childID < 0 {
        return ctx.MakeBoolean(false)
    }
    return ctx.MakeBoolean(contains(parentID, childID))
}

func insertAdjacentHTMLJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
    if len(args) < 3 {
        return ctx.MakeUndefined()
    }
    id := jsToNodeID(ctx, args[0])
    if id < 0 {
        return ctx.MakeUndefined()
    }
    position, _ := getArgString(ctx, args[1])
    html, _ := getArgString(ctx, args[2])
    InsertAdjacentHTML(id, position, html)
    SetDirty()
    return ctx.MakeUndefined()
}

func insertAdjacentTextJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
    if len(args) < 3 {
        return ctx.MakeUndefined()
    }
    id := jsToNodeID(ctx, args[0])
    if id < 0 {
        return ctx.MakeUndefined()
    }
    position, _ := getArgString(ctx, args[1])
    text, _ := getArgString(ctx, args[2])
    insertAdjacentText(id, position, text)
    SetDirty()
    return ctx.MakeUndefined()
}

func insertAdjacentElementJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
    if len(args) < 3 {
        return ctx.MakeNull()
    }
    id := jsToNodeID(ctx, args[0])
    elemID := jsToNodeID(ctx, args[1])
    if id < 0 || elemID < 0 {
        return ctx.MakeNull()
    }
    position, _ := getArgString(ctx, args[2])
    insertAdjacentElement(id, elemID, position)
    SetDirty()
    return args[1]
}

func outerHTMLJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
    if len(args) < 1 {
        return ctx.MakeString("")
    }
    id := jsToNodeID(ctx, args[0])
    if id < 0 {
        return ctx.MakeString("")
    }
    html := outerHTMLString(id)
    cstr := C.CString(html)
    defer C.free(unsafe.Pointer(cstr))
    return C._makeString(ctx.ctx, cstr)
}

func setOuterHTMLJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
    if len(args) < 2 {
        return ctx.MakeUndefined()
    }
    id := jsToNodeID(ctx, args[0])
    if id < 0 {
        return ctx.MakeUndefined()
    }
    html, _ := getArgString(ctx, args[1])
    setOuterHTML(id, html)
    SetDirty()
    return ctx.MakeUndefined()
}
