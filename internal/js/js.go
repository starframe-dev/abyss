// Package js provides JavaScript execution via JavaScriptCore (JSC) framework on macOS.
package js

/*
#cgo LDFLAGS: -framework JavaScriptCore
#cgo CFLAGS: -x objective-c

#include <JavaScriptCore/JavaScriptCore.h>
#include <stdlib.h>

// Declare functions from bridge.c
extern JSObjectRef makeGoFunction(JSContextRef ctx, void* goFuncID);
extern JSValueRef _evaluate(JSContextRef ctx, const char* script, const char* sourceURL, JSValueRef* exception);
extern char* _valueToString(JSContextRef ctx, JSValueRef value, JSValueRef* exception);
extern void _setProperty(JSContextRef ctx, JSObjectRef obj, const char* name, JSValueRef value, JSValueRef* exception);
extern JSValueRef _getProperty(JSContextRef ctx, JSObjectRef obj, const char* name, JSValueRef* exception);
extern JSValueRef _makeNumber(JSContextRef ctx, double value);
extern JSValueRef _makeString(JSContextRef ctx, const char* value);
extern JSValueRef _makeBoolean(JSContextRef ctx, bool value);
extern JSValueRef _makeUndefined(JSContextRef ctx);
extern JSValueRef _makeNull(JSContextRef ctx);
extern JSType _getType(JSContextRef ctx, JSValueRef value);
extern double _toNumber(JSContextRef ctx, JSValueRef value, JSValueRef* exception);
extern bool _toBoolean(JSContextRef ctx, JSValueRef value);
*/
import "C"
import (
    "fmt"
    "unsafe"
)

// Context wraps a JavaScriptCore global execution context.
type Context struct {
    ctx C.JSGlobalContextRef
}

// NewContext creates a new JavaScript execution context.
func NewContext() *Context {
    ctx := C.JSGlobalContextCreate(nil)
    return &Context{ctx: ctx}
}

// Close releases the JavaScript context.
func (c *Context) Close() {
    if c.ctx != nil {
        C.JSGlobalContextRelease(c.ctx)
        c.ctx = nil
    }
}

// GetContextRef returns the underlying C JSGlobalContextRef.
func (c *Context) GetContextRef() C.JSGlobalContextRef {
    return c.ctx
}

// Evaluate evaluates a JavaScript string and returns the result as a string.
func (c *Context) Evaluate(script string) (string, error) {
    cScript := C.CString(script)
    defer C.free(unsafe.Pointer(cScript))

    cURL := C.CString("script.js")
    defer C.free(unsafe.Pointer(cURL))

    var exception C.JSValueRef
    result := C._evaluate(c.ctx, cScript, cURL, &exception)

    if exception != nil {
        errStr := C._valueToString(c.ctx, exception, nil)
        if errStr != nil {
            err := C.GoString(errStr)
            C.free(unsafe.Pointer(errStr))
            return "", fmt.Errorf("js error: %s", err)
        }
        return "", fmt.Errorf("unknown js error")
    }

    if result == nil {
        return "", nil
    }

    str := C._valueToString(c.ctx, result, nil)
    if str != nil {
        defer C.free(unsafe.Pointer(str))
        return C.GoString(str), nil
    }

    return "", nil
}

// EvaluateRaw evaluates JavaScript and returns the raw JSValueRef.
func (c *Context) EvaluateRaw(script string) (C.JSValueRef, error) {
    cScript := C.CString(script)
    defer C.free(unsafe.Pointer(cScript))

    cURL := C.CString("script.js")
    defer C.free(unsafe.Pointer(cURL))

    var exception C.JSValueRef
    result := C._evaluate(c.ctx, cScript, cURL, &exception)

    if exception != nil {
        errStr := C._valueToString(c.ctx, exception, nil)
        if errStr != nil {
            err := C.GoString(errStr)
            C.free(unsafe.Pointer(errStr))
            return nil, fmt.Errorf("js error: %s", err)
        }
        return nil, fmt.Errorf("unknown js error")
    }

    return result, nil
}

// GlobalObject returns the global object.
func (c *Context) GlobalObject() C.JSObjectRef {
    return C.JSContextGetGlobalObject(c.ctx)
}

// SetProperty sets a property on the global object.
func (c *Context) SetProperty(name string, value C.JSValueRef) error {
    cName := C.CString(name)
    defer C.free(unsafe.Pointer(cName))

    var exception C.JSValueRef
    C._setProperty(c.ctx, c.GlobalObject(), cName, value, &exception)

    if exception != nil {
        errStr := C._valueToString(c.ctx, exception, nil)
        if errStr != nil {
            err := C.GoString(errStr)
            C.free(unsafe.Pointer(errStr))
            return fmt.Errorf("set property error: %s", err)
        }
        return fmt.Errorf("unknown set property error")
    }
    return nil
}

// SetPropertyOn sets a property on a specific object.
func (c *Context) SetPropertyOn(obj C.JSObjectRef, name string, value C.JSValueRef) error {
    cName := C.CString(name)
    defer C.free(unsafe.Pointer(cName))

    var exception C.JSValueRef
    C._setProperty(c.ctx, obj, cName, value, &exception)

    if exception != nil {
        errStr := C._valueToString(c.ctx, exception, nil)
        if errStr != nil {
            err := C.GoString(errStr)
            C.free(unsafe.Pointer(errStr))
            return fmt.Errorf("set property error: %s", err)
        }
        return fmt.Errorf("unknown set property error")
    }
    return nil
}

// GetProperty gets a property from the global object.
func (c *Context) GetProperty(name string) (C.JSValueRef, error) {
    cName := C.CString(name)
    defer C.free(unsafe.Pointer(cName))

    var exception C.JSValueRef
    value := C._getProperty(c.ctx, c.GlobalObject(), cName, &exception)

    if exception != nil {
        errStr := C._valueToString(c.ctx, exception, nil)
        if errStr != nil {
            err := C.GoString(errStr)
            C.free(unsafe.Pointer(errStr))
            return nil, fmt.Errorf("get property error: %s", err)
        }
        return nil, fmt.Errorf("unknown get property error")
    }
    return value, nil
}

// ValueToString converts a JSValueRef to a Go string.
func (c *Context) ValueToString(value C.JSValueRef) (string, error) {
    var exception C.JSValueRef
    str := C._valueToString(c.ctx, value, &exception)
    if str != nil {
        defer C.free(unsafe.Pointer(str))
        return C.GoString(str), nil
    }
    if exception != nil {
        errStr := C._valueToString(c.ctx, exception, nil)
        if errStr != nil {
            err := C.GoString(errStr)
            C.free(unsafe.Pointer(errStr))
            return "", fmt.Errorf("value to string error: %s", err)
        }
    }
    return "", nil
}

// MakeString creates a JS string value.
func (c *Context) MakeString(value string) C.JSValueRef {
    cValue := C.CString(value)
    defer C.free(unsafe.Pointer(cValue))
    return C._makeString(c.ctx, cValue)
}

// MakeNumber creates a JS number value.
func (c *Context) MakeNumber(value float64) C.JSValueRef {
    return C._makeNumber(c.ctx, C.double(value))
}

// MakeBoolean creates a JS boolean value.
func (c *Context) MakeBoolean(value bool) C.JSValueRef {
    return C._makeBoolean(c.ctx, C.bool(value))
}

// MakeUndefined creates an undefined JS value.
func (c *Context) MakeUndefined() C.JSValueRef {
    return C._makeUndefined(c.ctx)
}

// MakeNull creates a null JS value.
func (c *Context) MakeNull() C.JSValueRef {
    return C._makeNull(c.ctx)
}

// ToNumber converts a JS value to a Go float64.
func (c *Context) ToNumber(value C.JSValueRef) (float64, error) {
    var exception C.JSValueRef
    num := C._toNumber(c.ctx, value, &exception)
    if exception != nil {
        errStr := C._valueToString(c.ctx, exception, nil)
        if errStr != nil {
            err := C.GoString(errStr)
            C.free(unsafe.Pointer(errStr))
            return 0, fmt.Errorf("to number error: %s", err)
        }
        return 0, fmt.Errorf("unknown to number error")
    }
    return float64(num), nil
}

// ToBoolean converts a JS value to a Go bool.
func (c *Context) ToBoolean(value C.JSValueRef) bool {
    return bool(C._toBoolean(c.ctx, value))
}

// GetType returns the JavaScript type of a value.
func (c *Context) GetType(value C.JSValueRef) string {
    switch C._getType(c.ctx, value) {
    case C.kJSTypeUndefined:
        return "undefined"
    case C.kJSTypeNull:
        return "null"
    case C.kJSTypeBoolean:
        return "boolean"
    case C.kJSTypeNumber:
        return "number"
    case C.kJSTypeString:
        return "string"
    case C.kJSTypeObject:
        return "object"
    default:
        return "unknown"
    }
}
