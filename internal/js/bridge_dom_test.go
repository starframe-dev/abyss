//go:build jsc

package js

/*
#include <JavaScriptCore/JavaScriptCore.h>
*/
import "C"
import "testing"

// DOM bridge functions depend on the JavaScriptCore runtime; these smoke tests
// exercise each exported-to-JS callback through valid and missing arguments.
func TestDOMBridgeCallbacks(t *testing.T) {
	ctx := NewContext()
	if ctx == nil {
		t.Fatal("NewContext returned nil")
	}
	defer ctx.Release()

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
	for _, cb := range callbacks {
		t.Run(cb.name, func(t *testing.T) {
			if cb.fn(ctx, nil) == nil {
				t.Fatal("callback returned nil JS value")
			}
		})
	}
}
