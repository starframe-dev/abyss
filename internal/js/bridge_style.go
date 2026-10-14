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
	"strings"
	"unicode"
	"unsafe"
)

func getComputedStyleJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
	if len(args) < 1 {
		return ctx.MakeNull()
	}
	id := jsToNodeID(ctx, args[0])
	if id < 0 {
		return ctx.MakeNull()
	}
	return getComputedStyle(id)
}

func getBoundingClientRectJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
	if len(args) < 1 {
		return ctx.MakeNull()
	}
	id := jsToNodeID(ctx, args[0])
	if id < 0 {
		return ctx.MakeNull()
	}
	rect := getBoundingClientRect(id)

	// Create JS object with x, y, width, height, top, left, bottom, right.
	obj := C.JSObjectMake(ctx.ctx, nil, nil)
	setNum := func(name string, val float64) {
		cname := C.CString(name)
		defer C.free(unsafe.Pointer(cname))
		pn := C.JSStringCreateWithUTF8CString(cname)
		defer C.JSStringRelease(pn)
		C.JSObjectSetProperty(ctx.ctx, obj, pn, C.JSValueMakeNumber(ctx.ctx, C.double(val)), C.kJSPropertyAttributeNone, nil)
	}
	setNum("x", rect.X)
	setNum("y", rect.Y)
	setNum("width", rect.Width)
	setNum("height", rect.Height)
	setNum("top", rect.Y)
	setNum("left", rect.X)
	setNum("bottom", rect.Y+rect.Height)
	setNum("right", rect.X+rect.Width)

	return jsObjToVal(obj)
}

func logJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
	// Convert each argument and strip terminal controls before writing to stderr.
	var parts []string
	for _, arg := range args {
		s, err := getArgString(ctx, arg)
		if err != nil {
			parts = append(parts, "<error>")
		} else {
			parts = append(parts, s)
		}
	}
	var safe strings.Builder
	for _, r := range "[JS] " + strings.Join(parts, " ") + "\n" {
		if r == '\n' || r == '\t' || (!unicode.IsControl(r) && r != 0x7f) {
			safe.WriteRune(r)
		} else {
			safe.WriteRune('�')
		}
	}
	msg := C.CString(safe.String())
	defer C.free(unsafe.Pointer(msg))
	C._fprintf_stderr(msg)
	return ctx.MakeUndefined()
}
