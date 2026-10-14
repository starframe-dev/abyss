package js

/*
#include <JavaScriptCore/JavaScriptCore.h>
#include <stdlib.h>

// Declare functions from bridge.c
extern JSObjectRef makeGoFunction(JSContextRef ctx, const char* funcName);
extern JSType _getType(JSContextRef ctx, JSValueRef value);
extern char* _valueToString(JSContextRef ctx, JSValueRef value, JSValueRef* exception);
extern JSValueRef _makeArray(JSContextRef ctx, JSValueRef* values, size_t count, JSValueRef* exception);
extern JSValueRef _makeString(JSContextRef ctx, const char* value);
extern JSValueRef _makeBoolean(JSContextRef ctx, bool value);
extern void _fprintf_stderr(const char* msg);
extern void _protectValue(JSContextRef ctx, JSValueRef value);
extern void _unprotectValue(JSContextRef ctx, JSValueRef value);
extern JSValueRef _callAsFunction(JSContextRef ctx, JSObjectRef func, JSObjectRef thisObj, size_t argCount, const JSValueRef args[], JSValueRef* exception);
*/
import "C"

func setTimeoutJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
	if len(args) < 2 {
		return ctx.MakeNumber(0)
	}
	callback := C.JSValueToObject(ctx.ctx, args[0], nil)
	if callback == nil {
		return ctx.MakeNumber(0)
	}
	delay := 0
	if d, err := ctx.ToNumber(args[1]); err == nil {
		delay = int(d)
	}
	id := setTimeout(callback, delay)
	return ctx.MakeNumber(float64(id))
}

func setIntervalJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
	if len(args) < 2 {
		return ctx.MakeNumber(0)
	}
	callback := C.JSValueToObject(ctx.ctx, args[0], nil)
	if callback == nil {
		return ctx.MakeNumber(0)
	}
	delay := 0
	if d, err := ctx.ToNumber(args[1]); err == nil {
		delay = int(d)
	}
	id := setInterval(callback, delay)
	return ctx.MakeNumber(float64(id))
}

func clearTimeoutJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
	if len(args) < 1 {
		return ctx.MakeUndefined()
	}
	if id, err := ctx.ToNumber(args[0]); err == nil {
		clearTimer(int(id))
	}
	return ctx.MakeUndefined()
}

func clearIntervalJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
	if len(args) < 1 {
		return ctx.MakeUndefined()
	}
	if id, err := ctx.ToNumber(args[0]); err == nil {
		clearTimer(int(id))
	}
	return ctx.MakeUndefined()
}

func addEventListenerJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
	if len(args) < 3 {
		return ctx.MakeUndefined()
	}
	id := jsToNodeID(ctx, args[0])
	if id < 0 {
		return ctx.MakeUndefined()
	}
	eventType, _ := getArgString(ctx, args[1])
	callback := C.JSValueToObject(ctx.ctx, args[2], nil)
	if callback == nil {
		return ctx.MakeUndefined()
	}
	addEventListener(id, eventType, callback)
	return ctx.MakeUndefined()
}

func removeEventListenerJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
	if len(args) < 3 {
		return ctx.MakeUndefined()
	}
	id := jsToNodeID(ctx, args[0])
	if id < 0 {
		return ctx.MakeUndefined()
	}
	eventType, _ := getArgString(ctx, args[1])
	callback := C.JSValueToObject(ctx.ctx, args[2], nil)
	if callback == nil {
		return ctx.MakeUndefined()
	}
	removeEventListener(id, eventType, callback)
	return ctx.MakeUndefined()
}

func dispatchEventJS(ctx *Context, args []C.JSValueRef) C.JSValueRef {
	if len(args) < 2 {
		return ctx.MakeUndefined()
	}
	id := jsToNodeID(ctx, args[0])
	if id < 0 {
		return ctx.MakeUndefined()
	}
	eventType, _ := getArgString(ctx, args[1])
	dispatchEvent(id, eventType)
	return ctx.MakeUndefined()
}
