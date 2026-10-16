//go:build jsc

package js

import (
	"testing"
	"unsafe"
)

/*
#include <JavaScriptCore/JavaScriptCore.h>
*/
import "C"

func TestBridgeQueryJSFunctions(t *testing.T) {
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
	for _, tc := range functions {
		t.Run(tc.name, func(t *testing.T) {
			if result := tc.fn(ctx, nil); result == nil {
				t.Fatalf("%s returned a nil JSValueRef", tc.name)
			}
		})
	}
	_ = unsafe.Pointer(nil)
}
