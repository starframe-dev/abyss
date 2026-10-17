package js

/*
#include <JavaScriptCore/JavaScriptCore.h>
*/
import "C"

import (
	"errors"
	"strconv"
	"time"
)

// The checks below need JavaScriptCore types. go test does not allow import "C"
// in _test.go files, so they live here; the *_test.go files call them. They are
// not used by the browser at runtime.

// checkDOMBridgeCallbacks calls each DOM bridge callback with no arguments and
// returns the names of those that return a nil JS value.
func checkDOMBridgeCallbacks() []string {
	ctx := NewContext()
	if ctx == nil {
		return []string{"NewContext returned nil"}
	}
	defer ctx.Close()

	callbacks := []struct {
		name string
		fn   func(*Context, []C.JSValueRef) C.JSValueRef
	}{
		{"createElementJS", createElementJS}, {"appendChildJS", appendChildJS},
		{"setInnerHTMLJS", setInnerHTMLJS}, {"setAttributeJS", setAttributeJS},
		{"setTextContentJS", setTextContentJS}, {"createTextNodeJS", createTextNodeJS},
		{"createDocumentFragmentJS", createDocumentFragmentJS}, {"getInnerHTMLJS", getInnerHTMLJS},
		{"getTextContentJS", getTextContentJS}, {"removeChildJS", removeChildJS},
		{"insertBeforeJS", insertBeforeJS}, {"replaceChildJS", replaceChildJS},
		{"cloneNodeJS", cloneNodeJS}, {"containsJS", containsJS},
		{"insertAdjacentHTMLJS", insertAdjacentHTMLJS}, {"insertAdjacentTextJS", insertAdjacentTextJS},
		{"insertAdjacentElementJS", insertAdjacentElementJS}, {"outerHTMLJS", outerHTMLJS},
		{"setOuterHTMLJS", setOuterHTMLJS},
	}
	var bad []string
	for _, cb := range callbacks {
		if cb.fn(ctx, nil) == nil {
			bad = append(bad, cb.name)
		}
	}
	return bad
}

// checkQueryCallbacks runs each query callback in its own context group and
// returns the names of those that return a nil JS value.
func checkQueryCallbacks() []string {
	global := C.JSContextGroupCreate()
	defer C.JSContextGroupRelease(global)
	ctx := &Context{ctx: C.JSGlobalContextCreateInGroup(global, nil)}
	defer C.JSGlobalContextRelease(ctx.ctx)

	functions := []struct {
		name string
		fn   func(*Context, []C.JSValueRef) C.JSValueRef
	}{
		{"getBodyJS", getBodyJS}, {"getElementByIdJS", getElementByIdJS},
		{"querySelectorJS", querySelectorJS}, {"querySelectorAllJS", querySelectorAllJS},
		{"getAttributeJS", getAttributeJS}, {"getParentNodeJS", getParentNodeJS},
		{"getChildrenJS", getChildrenJS}, {"getElementsByTagNameJS", getElementsByTagNameJS},
		{"getElementsByClassNameJS", getElementsByClassNameJS}, {"matchesJS", matchesJS},
		{"closestJS", closestJS}, {"previousElementSiblingJS", previousElementSiblingJS},
		{"nextElementSiblingJS", nextElementSiblingJS}, {"firstElementChildJS", firstElementChildJS},
		{"lastElementChildJS", lastElementChildJS},
	}
	var bad []string
	for _, tc := range functions {
		if tc.fn(ctx, nil) == nil {
			bad = append(bad, tc.name)
		}
	}
	return bad
}

// newCountingCallback returns a JS function that increments globalThis.__calls.
func newCountingCallback(ctx *Context) (C.JSObjectRef, error) {
	v, err := ctx.EvaluateRaw("(function(){ globalThis.__calls = (globalThis.__calls || 0) + 1; })")
	if err != nil {
		return nil, err
	}
	obj := C.JSValueToObject(C.JSContextRef(ctx.ctx), v, nil)
	if obj == nil {
		return nil, errors.New("callback is not an object")
	}
	return obj, nil
}

// callCount reads globalThis.__calls from the context.
func callCount(ctx *Context) int {
	s, err := ctx.Evaluate("String(globalThis.__calls || 0)")
	if err != nil {
		return -1
	}
	n, _ := strconv.Atoi(s)
	return n
}

// checkEventListeners exercises add, dispatch and remove of DOM event listeners.
func checkEventListeners() []string {
	ctx := NewContext()
	if ctx == nil {
		return []string{"NewContext returned nil"}
	}
	defer ctx.Close()
	InitEventSystem(ctx.ctx)
	defer resetEventSystem()

	fn, err := newCountingCallback(ctx)
	if err != nil {
		return []string{err.Error()}
	}
	var bad []string
	AddEventListener(1, "click", fn)
	dispatchEvent(1, "other")
	if n := callCount(ctx); n != 0 {
		bad = append(bad, "unexpected event invocation: "+strconv.Itoa(n))
	}
	dispatchEvent(1, "click")
	if n := callCount(ctx); n != 1 {
		bad = append(bad, "click count: "+strconv.Itoa(n))
	}
	RemoveEventListener(1, "click", fn)
	dispatchEvent(1, "click")
	if n := callCount(ctx); n != 1 {
		bad = append(bad, "removed callback invoked: "+strconv.Itoa(n))
	}
	removeEventListener(2, "click", fn)
	return bad
}

// checkTimers exercises one-shot and interval timers through ProcessTimers.
func checkTimers() []string {
	ctx := NewContext()
	if ctx == nil {
		return []string{"NewContext returned nil"}
	}
	defer ctx.Close()
	InitEventSystem(ctx.ctx)
	defer resetEventSystem()

	fn, err := newCountingCallback(ctx)
	if err != nil {
		return []string{err.Error()}
	}
	var bad []string
	if id := setTimeout(fn, 0); id == 0 {
		return []string{"invalid timer id"}
	}
	time.Sleep(time.Millisecond)
	if !ProcessTimers() || callCount(ctx) != 1 {
		bad = append(bad, "one-shot timer did not run once")
	}
	if ProcessTimers() {
		bad = append(bad, "one-shot timer ran twice")
	}

	intervalID := setInterval(fn, 0)
	time.Sleep(time.Millisecond)
	if !ProcessTimers() || callCount(ctx) != 2 {
		bad = append(bad, "interval did not run")
	}
	clearTimer(intervalID)
	if ProcessTimers() {
		bad = append(bad, "cleared interval ran")
	}
	return bad
}

// checkBridgeEventArgs checks argument handling of the timer and event bridge
// callbacks and returns one message per failed expectation.
func checkBridgeEventArgs() []string {
	ctx := NewContext()
	defer ctx.Close()

	var bad []string
	undefined := ctx.MakeUndefined()
	zero := ctx.MakeNumber(0)
	callback, err := ctx.EvaluateRaw("(function(){})")
	if err != nil {
		return []string{"evaluate callback: " + err.Error()}
	}

	isNumber := func(name string, v C.JSValueRef) {
		if _, err := ctx.ToNumber(v); err != nil {
			bad = append(bad, name+" did not return a number: "+err.Error())
		}
	}
	isNumber("setTimeoutJS", setTimeoutJS(ctx, nil))
	isNumber("short setTimeoutJS", setTimeoutJS(ctx, []C.JSValueRef{zero}))
	isNumber("setTimeoutJS with callback", setTimeoutJS(ctx, []C.JSValueRef{callback, zero}))
	isNumber("setIntervalJS", setIntervalJS(ctx, nil))
	isNumber("setIntervalJS with callback", setIntervalJS(ctx, []C.JSValueRef{callback, zero}))

	for _, fn := range []func(*Context, []C.JSValueRef) C.JSValueRef{clearTimeoutJS, clearIntervalJS, addEventListenerJS, removeEventListenerJS, dispatchEventJS} {
		if got := fn(ctx, nil); got != undefined {
			bad = append(bad, "callback with no args did not return undefined")
		}
	}
	return bad
}
