package js

import (
	"testing"
)

// These bridge functions depend on JavaScriptCore values and are exercised through
// a live Context so their argument validation and returned JS values are covered.
func TestBridgeEvents(t *testing.T) {
	ctx := NewContext()
	defer ctx.Close()

	undefined := ctx.MakeUndefined()
	zero := ctx.MakeNumber(0)
	callback := ctx.EvaluateScript("(function(){})")

	if got := ctx.ToNumber(setTimeoutJS(ctx, nil)); got == nil {
		t.Fatal("setTimeoutJS did not return a number")
	}
	if got := ctx.ToNumber(setTimeoutJS(ctx, []C.JSValueRef{zero})); got == nil {
		t.Fatal("short setTimeoutJS did not return a number")
	}
	if got := ctx.ToNumber(setTimeoutJS(ctx, []C.JSValueRef{callback, zero})); got == nil {
		t.Fatal("setTimeoutJS did not return a number")
	}
	if got := ctx.ToNumber(setIntervalJS(ctx, nil)); got == nil {
		t.Fatal("setIntervalJS did not return a number")
	}
	if got := ctx.ToNumber(setIntervalJS(ctx, []C.JSValueRef{callback, zero})); got == nil {
		t.Fatal("setIntervalJS did not return a number")
	}

	for _, fn := range []func(*Context, []C.JSValueRef) C.JSValueRef{clearTimeoutJS, clearIntervalJS, addEventListenerJS, removeEventListenerJS, dispatchEventJS} {
		if got := fn(ctx, nil); got != undefined {
			t.Fatalf("%T with no args did not return undefined", fn)
		}
	}
}
