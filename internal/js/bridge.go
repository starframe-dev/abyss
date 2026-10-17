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
	"fmt"
	"io"
	"unsafe"

	"github.com/starframe-dev/abyss/internal/dom"
)

// LogWriter is the output writer for console.log bridge.
var LogWriter io.Writer = io.Discard

// bridgeFunc is a single go→JS bridge handler.
type bridgeFunc func(ctx *Context, args []C.JSValueRef) C.JSValueRef

// bridgeHandlers maps a DOM API name to its Go implementation.
var bridgeHandlers = map[string]bridgeFunc{
	"createElement":          createElementJS,
	"appendChild":            appendChildJS,
	"setInnerHTML":           setInnerHTMLJS,
	"setAttribute":           setAttributeJS,
	"setTextContent":         setTextContentJS,
	"createTextNode":         createTextNodeJS,
	"createDocumentFragment": createDocumentFragmentJS,
	"documentQuerySelector":  documentQuerySelectorJS,
	"setTimeout":             setTimeoutJS,
	"setInterval":            setIntervalJS,
	"clearTimeout":           clearTimeoutJS,
	"clearInterval":          clearIntervalJS,
	"getBody":                getBodyJS,
	"getElementById":         getElementByIdJS,
	"getAttribute":           getAttributeJS,
	"removeChild":            removeChildJS,
	"getParentNode":          getParentNodeJS,
	"getChildren":            getChildrenJS,
	"getInnerHTML":           getInnerHTMLJS,
	"getTextContent":         getTextContentJS,
	"getElementsByTagName":   getElementsByTagNameJS,
	"getElementsByClassName": getElementsByClassNameJS,
	"querySelector":          querySelectorJS,
	"querySelectorAll":       querySelectorAllJS,
	"addEventListener":       addEventListenerJS,
	"removeEventListener":    removeEventListenerJS,
	"dispatchEvent":          dispatchEventJS,
	"getComputedStyle":       getComputedStyleJS,
	"matches":                matchesJS,
	"closest":                closestJS,
	"insertBefore":           insertBeforeJS,
	"replaceChild":           replaceChildJS,
	"cloneNode":              cloneNodeJS,
	"contains":               containsJS,
	"getBoundingClientRect":  getBoundingClientRectJS,
	"previousElementSibling": previousElementSiblingJS,
	"nextElementSibling":     nextElementSiblingJS,
	"firstElementChild":      firstElementChildJS,
	"lastElementChild":       lastElementChildJS,
	"insertAdjacentHTML":     insertAdjacentHTMLJS,
	"insertAdjacentText":     insertAdjacentTextJS,
	"insertAdjacentElement":  insertAdjacentElementJS,
	"outerHTML":              outerHTMLJS,
	"setOuterHTML":           setOuterHTMLJS,
	"log":                    logJS,
}

//export goBridgeDispatch
func goBridgeDispatch(funcName *C.char, ctx C.JSContextRef, function C.JSObjectRef, thisObject C.JSObjectRef, argumentCount C.size_t, arguments unsafe.Pointer, exceptionPtr unsafe.Pointer) C.JSValueRef {
	name := C.GoString(funcName)
	gCtx := &Context{ctx: C.JSGlobalContextRef(ctx)}

	argCount := int(argumentCount)
	args := make([]C.JSValueRef, argCount)
	if argCount > 0 {
		argSlice := (*[1 << 24]C.JSValueRef)(arguments)[:argCount:argCount]
		copy(args, argSlice)
	}

	if fn, ok := bridgeHandlers[name]; ok {
		return fn(gCtx, args)
	}

	errStr := C.CString(fmt.Sprintf("unknown function: %s", name))
	defer C.free(unsafe.Pointer(errStr))
	jsErrStr := C.JSStringCreateWithUTF8CString(errStr)
	defer C.JSStringRelease(jsErrStr)
	*(*C.JSValueRef)(exceptionPtr) = C.JSValueMakeString(ctx, jsErrStr)
	return nil
}

// Helper: get a Go string from a JSValueRef argument
func getArgString(ctx *Context, val C.JSValueRef) (string, error) {
	var exception C.JSValueRef
	str := C._valueToString(ctx.ctx, val, &exception)
	if str != nil {
		defer C.free(unsafe.Pointer(str))
		return C.GoString(str), nil
	}
	return "", fmt.Errorf("arg error")
}

// Helper: get NodeID from a JSValueRef
func jsToNodeID(ctx *Context, val C.JSValueRef) NodeID {
	if C._getType(ctx.ctx, val) != C.kJSTypeObject {
		return -1
	}
	obj := C.JSValueToObject(ctx.ctx, val, nil)
	if obj == nil {
		return -1
	}
	propName := C.JSStringCreateWithUTF8CString(C.CString("__go_id"))
	defer C.JSStringRelease(propName)
	var exc C.JSValueRef
	idVal := C.JSObjectGetProperty(ctx.ctx, obj, propName, &exc)
	if exc != nil || idVal == nil {
		return -1
	}
	num, err := ctx.ToNumber(idVal)
	if err != nil {
		return -1
	}
	return NodeID(num)
}

// Helper: make a JS wrapper for a node
func nodeToJSValue(ctx *Context, n *dom.Node) C.JSValueRef {
	return nodeToJS(ctx, n)
}

// RegisterBridgeFunctions registers the bridge functions on the global object.
func (c *Context) RegisterBridgeFunctions() error {
	bridgeNames := []string{
		"createElement",
		"appendChild",
		"setInnerHTML",
		"setAttribute",
		"setTextContent",
		"createTextNode",
		"createDocumentFragment",
		"setTimeout",
		"setInterval",
		"clearTimeout",
		"clearInterval",
		"getBody",
		"documentQuerySelector",
		"getElementById",
		"getAttribute",
		"removeChild",
		"getParentNode",
		"getChildren",
		"getInnerHTML",
		"getTextContent",
		"getElementsByTagName",
		"getElementsByClassName",
		"querySelector",
		"querySelectorAll",
		"addEventListener",
		"removeEventListener",
		"dispatchEvent",
		"getComputedStyle",
		"matches",
		"closest",
		"insertBefore",
		"replaceChild",
		"cloneNode",
		"contains",
		"getBoundingClientRect",
		"previousElementSibling",
		"nextElementSibling",
		"firstElementChild",
		"lastElementChild",
		"insertAdjacentHTML",
		"insertAdjacentText",
		"insertAdjacentElement",
		"outerHTML",
		"setOuterHTML",
		"log",
	}
	for _, name := range bridgeNames {
		fn := C.makeGoFunction(c.ctx, C.CString(name))
		jsName := "__go_" + name
		if err := c.SetProperty(jsName, jsObjToVal(fn)); err != nil {
			return err
		}
	}
	return nil
}

// jsObjToVal converts a JSObjectRef to JSValueRef.
func jsObjToVal(obj C.JSObjectRef) C.JSValueRef {
	return *(*C.JSValueRef)(unsafe.Pointer(&obj))
}
