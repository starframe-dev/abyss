// Bridge between JavaScriptCore C API and Go.
// Uses JSClassCreate for proper private data support.

#include <JavaScriptCore/JavaScriptCore.h>
#include <stdlib.h>
#include <stdio.h>
#include "_cgo_export.h"

// === Helper functions ===

JSStringRef _str(const char* s) {
    return JSStringCreateWithUTF8CString(s);
}

void _releaseStr(JSStringRef s) {
    JSStringRelease(s);
}

char* _valueToString(JSContextRef ctx, JSValueRef value, JSValueRef* exception) {
    JSStringRef str = JSValueToStringCopy(ctx, value, exception);
    if (!str) return NULL;
    size_t len = JSStringGetMaximumUTF8CStringSize(str);
    char* buf = (char*)malloc(len);
    JSStringGetUTF8CString(str, buf, len);
    JSStringRelease(str);
    return buf;
}

void _setProperty(JSContextRef ctx, JSObjectRef obj, const char* name, JSValueRef value, JSValueRef* exception) {
    JSStringRef propName = _str(name);
    JSObjectSetProperty(ctx, obj, propName, value, kJSPropertyAttributeNone, exception);
    _releaseStr(propName);
}

JSValueRef _getProperty(JSContextRef ctx, JSObjectRef obj, const char* name, JSValueRef* exception) {
    JSStringRef propName = _str(name);
    JSValueRef value = JSObjectGetProperty(ctx, obj, propName, exception);
    _releaseStr(propName);
    return value;
}

JSValueRef _evaluate(JSContextRef ctx, const char* script, const char* sourceURL, JSValueRef* exception) {
    JSStringRef jsScript = _str(script);
    JSStringRef jsURL = _str(sourceURL);
    JSValueRef result = JSEvaluateScript(ctx, jsScript, NULL, jsURL, 1, exception);
    _releaseStr(jsScript);
    _releaseStr(jsURL);
    return result;
}

JSValueRef _makeNumber(JSContextRef ctx, double value) {
    return JSValueMakeNumber(ctx, value);
}

JSValueRef _makeString(JSContextRef ctx, const char* value) {
    JSStringRef str = _str(value);
    JSValueRef result = JSValueMakeString(ctx, str);
    _releaseStr(str);
    return result;
}

JSValueRef _makeBoolean(JSContextRef ctx, bool value) {
    return JSValueMakeBoolean(ctx, value);
}

JSValueRef _makeUndefined(JSContextRef ctx) {
    return JSValueMakeUndefined(ctx);
}

JSValueRef _makeNull(JSContextRef ctx) {
    return JSValueMakeNull(ctx);
}

JSType _getType(JSContextRef ctx, JSValueRef value) {
    return JSValueGetType(ctx, value);
}

JSValueRef _makeArray(JSContextRef ctx, JSValueRef* values, size_t count, JSValueRef* exception) {
    return JSObjectMakeArray(ctx, count, values, exception);
}

FILE* _getStderr(void) {
    return stderr;
}

void _fprintf_stderr(const char* msg) {
    fprintf(stderr, "%s", msg);
    fflush(stderr);
}

double _toNumber(JSContextRef ctx, JSValueRef value, JSValueRef* exception) {
    return JSValueToNumber(ctx, value, exception);
}

bool _toBoolean(JSContextRef ctx, JSValueRef value) {
    return JSValueToBoolean(ctx, value);
}

void _protectValue(JSContextRef ctx, JSValueRef value) {
    JSValueProtect(ctx, value);
}

void _unprotectValue(JSContextRef ctx, JSValueRef value) {
    JSValueUnprotect(ctx, value);
}

JSValueRef _callAsFunction(JSContextRef ctx, JSObjectRef func, JSObjectRef thisObj, size_t argCount, const JSValueRef args[], JSValueRef* exception) {
    return JSObjectCallAsFunction(ctx, func, thisObj, argCount, args, exception);
}

// === Go Bridge Functions ===

// Callback function for Go bridge functions
static JSValueRef bridgeCallAsFunction(
    JSContextRef ctx,
    JSObjectRef function,
    JSObjectRef thisObject,
    size_t argumentCount,
    const JSValueRef arguments[],
    JSValueRef* exception)
{
    // Get the function name stored as private data
    char* funcName = (char*)JSObjectGetPrivate(function);
    if (!funcName) {
        JSStringRef errStr = JSStringCreateWithUTF8CString("missing function name");
        *exception = JSValueMakeString(ctx, errStr);
        JSStringRelease(errStr);
        return NULL;
    }
    return goBridgeDispatch(funcName, ctx, function, thisObject, argumentCount, (void*)arguments, exception);
}

// Finalize function to free the name string
static void bridgeFinalize(JSObjectRef object) {
    char* name = (char*)JSObjectGetPrivate(object);
    if (name) {
        free(name);
    }
}

// Create a bridge function class once
static JSClassRef getBridgeClass() {
    static JSClassRef bridgeClass = NULL;
    if (bridgeClass == NULL) {
        JSClassDefinition def = kJSClassDefinitionEmpty;
        def.className = "GoBridgeFunction";
        def.callAsFunction = bridgeCallAsFunction;
        def.finalize = bridgeFinalize;
        bridgeClass = JSClassCreate(&def);
    }
    return bridgeClass;
}

// Make a Go bridge function
JSObjectRef makeGoFunction(JSContextRef ctx, const char* funcName)
{
    JSClassRef bridgeClass = getBridgeClass();
    // Copy the name string - will be freed by finalize
    char* nameCopy = strdup(funcName);
    JSObjectRef func = JSObjectMake(ctx, bridgeClass, nameCopy);
    return func;
}