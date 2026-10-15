package js

/*
#include <JavaScriptCore/JavaScriptCore.h>
*/
import "C"
import (
	"testing"
	"time"
)

func TestEventListenersAndDispatch(t *testing.T) {
	ctx := NewContext()
	if ctx == nil {
		t.Fatal("NewContext returned nil")
	}
	defer ctx.Close()
	InitEventSystem(ctx.ctx)
	defer resetEventSystem()

	var count int
	fn := ctx.Function(func(args []Value) Value { count++; return Undefined() })
	AddEventListener(1, "click", fn.obj)
	dispatchEvent(1, "other")
	if count != 0 {
		t.Fatal("unexpected event invocation")
	}
	dispatchEvent(1, "click")
	if count != 1 {
		t.Fatalf("count=%d", count)
	}
	RemoveEventListener(1, "click", fn.obj)
	dispatchEvent(1, "click")
	if count != 1 {
		t.Fatalf("removed callback invoked: %d", count)
	}
	removeEventListener(2, "click", fn.obj)
}

func TestTimers(t *testing.T) {
	ctx := NewContext()
	if ctx == nil {
		t.Fatal("NewContext returned nil")
	}
	defer ctx.Close()
	InitEventSystem(ctx.ctx)
	defer resetEventSystem()

	var count int
	fn := ctx.Function(func(args []Value) Value { count++; return Undefined() })
	id := setTimeout(fn.obj, 0)
	if id == 0 {
		t.Fatal("invalid timer id")
	}
	time.Sleep(time.Millisecond)
	if !ProcessTimers() || count != 1 {
		t.Fatalf("ProcessTimers/count: %d", count)
	}
	if ProcessTimers() {
		t.Fatal("one-shot timer ran twice")
	}

	intervalID := setInterval(fn.obj, 0)
	time.Sleep(time.Millisecond)
	if !ProcessTimers() || count != 2 {
		t.Fatalf("interval did not run: %d", count)
	}
	clearTimer(intervalID)
	if ProcessTimers() {
		t.Fatal("cleared interval ran")
	}
}
