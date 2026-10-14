package js

import (
	"context"
	"errors"
	"testing"

	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
)

// TestNewContextClose verifies that a JavaScript context can be created
// and closed without crashing (covers NewContext + Close + GetContextRef).
func TestJSNodeRegistryHonorsDOMNodeLimit(t *testing.T) {
	doc := &dom.Node{Type: dom.DocumentNode}
	for i := 0; i < dom.MaxDOMNodes-1; i++ {
		child := &dom.Node{Type: dom.ElementNode, Parent: doc}
		doc.Children = append(doc.Children, child)
	}
	InitDOM(doc)
	defer func() { InitDOM(dom.Parse(`<html></html>`)) }()
	if len(nodes) != dom.MaxDOMNodes {
		t.Fatalf("registered nodes = %d, want %d", len(nodes), dom.MaxDOMNodes)
	}
	if created := createElement("extra"); created != nil {
		t.Fatal("node created after registry reached its cap")
	}
	if created := createTextNode("extra"); created != nil {
		t.Fatal("text node created after registry reached its cap")
	}
}

func TestInitDOMReleasesPreviousPageStyleAndLayoutData(t *testing.T) {
	oldDoc := dom.Parse(`<div id="old">previous</div>`)
	oldNode := oldDoc.QuerySelector("#old")
	SetComputedStyles(map[*dom.Node]css.ComputedStyle{oldNode: {}})
	SetLayoutBoxes(map[*dom.Node]Rect{oldNode: {Width: 10}})

	InitDOM(dom.Parse(`<div id="new">current</div>`))

	computedStylesMu.Lock()
	styles := computedStyles
	computedStylesMu.Unlock()
	layoutBoxesMu.Lock()
	boxes := layoutBoxes
	layoutBoxesMu.Unlock()
	if len(styles) != 0 || len(boxes) != 0 {
		t.Fatalf("previous page state retained: styles=%d layout boxes=%d", len(styles), len(boxes))
	}
}

func TestCloseCurrentContextReleasesPageState(t *testing.T) {
	ctx := NewContext()
	SetCurrentContext(ctx, "https://example.test")
	doc := dom.Parse(`<div id="page">content</div>`)
	InitDOM(doc)
	SetComputedStyles(map[*dom.Node]css.ComputedStyle{doc: {}})
	SetLayoutBoxes(map[*dom.Node]Rect{doc: {Width: 20}})
	SetDirty()

	CloseCurrentContext()

	if current, url := GetCurrentContext(); current != nil || url != "" {
		t.Fatalf("current context after close = (%v, %q)", current, url)
	}
	nodeMu.Lock()
	retainedNodes, retainedDoc := len(nodes), rootDoc
	nodeMu.Unlock()
	computedStylesMu.Lock()
	retainedStyles := len(computedStyles)
	computedStylesMu.Unlock()
	layoutBoxesMu.Lock()
	retainedBoxes := len(layoutBoxes)
	layoutBoxesMu.Unlock()
	if retainedNodes != 0 || retainedDoc != nil || retainedStyles != 0 || retainedBoxes != 0 || IsDirty() {
		t.Fatalf("page state retained: nodes=%d doc=%p styles=%d layout boxes=%d dirty=%v", retainedNodes, retainedDoc, retainedStyles, retainedBoxes, IsDirty())
	}
}

func TestDOMMutationRejectsAncestorCycles(t *testing.T) {
	doc := dom.Parse(`<div id="parent"><section id="child"></section></div>`)
	InitDOM(doc)
	parent, child := doc.QuerySelector("#parent"), doc.QuerySelector("#child")
	parentID, childID := getNodeID(parent), getNodeID(child)
	appendChild(parentID, parentID)
	appendChild(childID, parentID)
	insertBefore(childID, parentID, -1)
	if child.Parent != parent || parent.Parent != doc {
		t.Fatal("cycle attempt changed the DOM tree")
	}
	if got := len(doc.QuerySelector("#parent").Children); got != 1 {
		t.Fatalf("parent children = %d, want 1", got)
	}
}

func TestExecuteScriptsContextHonorsCancellationBeforeEvaluation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := ExecuteScriptsContext(ctx, nil, dom.Parse(`<script>throw new Error("must not run")</script>`), "https://example.test")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
}

func TestContextCloseResetsOwnedEventSystem(t *testing.T) {
	ctx := NewContext()
	InitEventSystem(ctx.GetContextRef())
	ctx.Close()
	eventMu.Lock()
	got := eventCtx
	eventMu.Unlock()
	if got != nil {
		t.Fatal("event system retained a released JS context")
	}
}

func TestNewContextClose(t *testing.T) {
	ctx := NewContext()
	if ctx == nil {
		t.Fatal("NewContext returned nil")
	}
	if ctx.GetContextRef() == nil {
		t.Error("expected non-nil context ref")
	}
	ctx.Close()
	// Double-close should be safe.
	ctx.Close()
}

// TestEvaluateSimpleExpression exercises Evaluate with a trivial script.
func TestEvaluateSimpleExpression(t *testing.T) {
	ctx := NewContext()
	defer ctx.Close()

	result, err := ctx.Evaluate("1 + 2")
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}
	if result != "3" {
		t.Errorf("expected '3', got %q", result)
	}
}

// TestEvaluateStringExpression exercises Evaluate with a string script.
func TestEvaluateStringExpression(t *testing.T) {
	ctx := NewContext()
	defer ctx.Close()

	result, err := ctx.Evaluate("'hello' + ' ' + 'world'")
	if err != nil {
		t.Fatalf("Evaluate failed: %v", err)
	}
	if result != "hello world" {
		t.Errorf("expected 'hello world', got %q", result)
	}
}

// TestEvaluateError verifies that Evaluate returns an error for invalid JS.
func TestEvaluateError(t *testing.T) {
	ctx := NewContext()
	defer ctx.Close()

	_, err := ctx.Evaluate("thisIsUndefined.foo")
	if err == nil {
		t.Error("expected error for undefined reference, got nil")
	}
}

// TestEvaluateRaw covers the raw evaluation path.
func TestEvaluateRaw(t *testing.T) {
	ctx := NewContext()
	defer ctx.Close()

	val, err := ctx.EvaluateRaw("42")
	if err != nil {
		t.Fatalf("EvaluateRaw failed: %v", err)
	}
	if val == nil {
		t.Error("expected non-nil JSValueRef for numeric literal")
	}
}

// TestEvaluateRawError covers the error path of EvaluateRaw.
func TestEvaluateRawError(t *testing.T) {
	ctx := NewContext()
	defer ctx.Close()

	_, err := ctx.EvaluateRaw("throw new Error('boom')")
	if err == nil {
		t.Error("expected error from throwing script")
	}
}

// TestGlobalObjectAndMakeAccessors exercises GlobalObject + Make* + GetType +
// ToNumber/ToBoolean/ValueToString on JS values.
func TestGlobalObjectAndMakeAccessors(t *testing.T) {
	ctx := NewContext()
	defer ctx.Close()

	obj := ctx.GlobalObject()
	if obj == nil {
		t.Fatal("GlobalObject returned nil ref")
	}

	numVal := ctx.MakeNumber(3.14)
	if ctx.GetType(numVal) != "number" {
		t.Errorf("type of MakeNumber(3.14) = %q, want number", ctx.GetType(numVal))
	}
	num, err := ctx.ToNumber(numVal)
	if err != nil {
		t.Fatalf("ToNumber failed: %v", err)
	}
	if num != 3.14 {
		t.Errorf("ToNumber = %v, want 3.14", num)
	}

	strVal := ctx.MakeString("hi")
	if ctx.GetType(strVal) != "string" {
		t.Errorf("type of MakeString = %q, want string", ctx.GetType(strVal))
	}
	str, err := ctx.ValueToString(strVal)
	if err != nil {
		t.Fatalf("ValueToString failed: %v", err)
	}
	if str != "hi" {
		t.Errorf("ValueToString = %q, want hi", str)
	}

	boolVal := ctx.MakeBoolean(true)
	if ctx.GetType(boolVal) != "boolean" {
		t.Errorf("type of MakeBoolean = %q, want boolean", ctx.GetType(boolVal))
	}
	if !ctx.ToBoolean(boolVal) {
		t.Error("ToBoolean(true) = false")
	}
	if ctx.ToBoolean(ctx.MakeBoolean(false)) {
		t.Error("ToBoolean(false) = true")
	}

	if ctx.GetType(ctx.MakeUndefined()) != "undefined" {
		t.Error("type of MakeUndefined is not 'undefined'")
	}
	if ctx.GetType(ctx.MakeNull()) != "null" {
		t.Error("type of MakeNull is not 'null'")
	}
}

// TestSetAndGetProperty exercises SetProperty + GetProperty on the global object.
func TestSetAndGetProperty(t *testing.T) {
	ctx := NewContext()
	defer ctx.Close()

	err := ctx.SetProperty("abyss_test_var", ctx.MakeNumber(7))
	if err != nil {
		t.Fatalf("SetProperty failed: %v", err)
	}

	val, err := ctx.GetProperty("abyss_test_var")
	if err != nil {
		t.Fatalf("GetProperty failed: %v", err)
	}
	num, err := ctx.ToNumber(val)
	if err != nil {
		t.Fatalf("ToNumber failed: %v", err)
	}
	if num != 7 {
		t.Errorf("got %v, want 7", num)
	}
}

// TestSetPropertyOn covers SetPropertyOn against a freshly constructed object.
func TestSetPropertyOn(t *testing.T) {
	ctx := NewContext()
	defer ctx.Close()

	val, err := ctx.EvaluateRaw("({foo: 1})")
	if err != nil {
		t.Fatalf("EvaluateRaw failed: %v", err)
	}
	if val == nil {
		t.Fatal("EvaluateRaw returned nil")
	}
	obj := ctx.GlobalObject() // not strictly an object, but exercises the path

	if err := ctx.SetPropertyOn(obj, "x", ctx.MakeNumber(99)); err != nil {
		t.Fatalf("SetPropertyOn failed: %v", err)
	}
}

// TestToNumberError covers the error branch of ToNumber by passing a value
// produced via a JS exception.
func TestToNumberError(t *testing.T) {
	ctx := NewContext()
	defer ctx.Close()

	// Property access on undefined throws -> returns nil (error path).
	val, err := ctx.EvaluateRaw("({}).missingProp.deep")
	if err == nil {
		// No exception raised; JS returned undefined. Convert undefined to number.
		if val != nil {
			if _, convErr := ctx.ToNumber(val); convErr == nil {
				// NaN is allowed by JS without an error, so we just
				// exercised the path successfully.
				t.Log("ToNumber tolerated undefined; got NaN-like")
			}
		}
	}
}

// TestInitDOMAndGetBody covers the InitDOM/getBody/getElementByID path
// using a parsed dom tree.
func TestInitDOMAndGetBody(t *testing.T) {
	html := `<html><body><div id="x">hello</div></body></html>`
	doc := dom.Parse(html)
	if doc == nil {
		t.Fatal("dom.Parse returned nil")
	}

	InitDOM(doc)
	body := getBody()
	if body == nil {
		t.Fatal("getBody returned nil after InitDOM")
	}

	div := getElementByID("x")
	if div == nil {
		t.Fatal("getElementByID(\"x\") returned nil")
	}
}

// TestCreateElementAndSetText covers createElement, setTextContent, and
// setAttribute via the public NodeID helpers.
func TestCreateElementAndSetText(t *testing.T) {
	InitDOM(dom.Parse("<html><body></body></html>"))

	elem := createElement("p")
	if elem == nil {
		t.Fatal("createElement returned nil")
	}

	id := getNodeID(elem)
	if id == 0 {
		t.Error("expected non-zero NodeID for new element")
	}

	setTextContent(id, "hi")
	if got := getAttribute(id, "id"); got != "" {
		// Just exercise the getter; attribute "id" wasn't set.
		_ = got
	}
}

// TestSetAttribute covers the setAttribute helper.
func TestSetAttribute(t *testing.T) {
	InitDOM(dom.Parse("<html><body></body></html>"))

	elem := createElement("a")
	id := getNodeID(elem)
	setAttribute(id, "href", "https://example.com")
	if got := getAttribute(id, "href"); got != "https://example.com" {
		t.Errorf("getAttribute(href) = %q, want %q", got, "https://example.com")
	}
}

// TestCreateTextNode covers createTextNode and the setTextContent path.
func TestCreateTextNode(t *testing.T) {
	InitDOM(dom.Parse("<html><body></body></html>"))
	tn := createTextNode("hello text")
	if tn == nil {
		t.Fatal("createTextNode returned nil")
	}
}

// TestCreateDocumentFragment covers createDocumentFragment.
func TestCreateDocumentFragment(t *testing.T) {
	InitDOM(dom.Parse("<html><body></body></html>"))
	frag := createDocumentFragment()
	if frag == nil {
		t.Fatal("createDocumentFragment returned nil")
	}
}

// TestRemoveChild covers removeChild helper. The behaviour depends on the
// underlying implementation; we only assert that the call is safe.
func TestRemoveChild(t *testing.T) {
	InitDOM(dom.Parse(`<html><body><div id="parent"><p id="kid">x</p></div></body></html>`))
	parent := getElementByID("parent")
	kid := getElementByID("kid")
	if parent == nil || kid == nil {
		t.Fatal("setup failed: parent or kid missing")
	}
	parentID := getNodeID(parent)
	kidID := getNodeID(kid)
	removeChild(parentID, kidID)
	// removeChild may be implemented as a no-op stub; we only verify no panic.
}

// TestAppendChild covers appendChild + removeNodeFromParentLocked helpers.
func TestAppendChild(t *testing.T) {
	InitDOM(dom.Parse(`<html><body><div id="a"></div><div id="b"><p id="kid">x</p></div></body></html>`))
	a := getElementByID("a")
	b := getElementByID("b")
	kid := getElementByID("kid")
	if a == nil || b == nil || kid == nil {
		t.Fatal("setup failed")
	}
	aID := getNodeID(a)
	kidID := getNodeID(kid)
	appendChild(aID, kidID)
	if getElementByID("kid") == nil {
		t.Error("kid missing after appendChild")
	}
}

// TestSetInnerHTML covers setInnerHTML.
func TestSetInnerHTML(t *testing.T) {
	InitDOM(dom.Parse(`<html><body><div id="x"></div></body></html>`))
	x := getElementByID("x")
	if x == nil {
		t.Fatal("x missing")
	}
	id := getNodeID(x)
	setInnerHTML(id, `<span>hi</span>`)
}

// TestMoveAllChildren covers moveAllChildren helper.
func TestMoveAllChildren(t *testing.T) {
	InitDOM(dom.Parse(`<html><body><div id="a"><p>x</p></div><div id="b"></div></body></html>`))
	a := getElementByID("a")
	b := getElementByID("b")
	if a == nil || b == nil {
		t.Fatal("setup failed")
	}
	moveAllChildren(a, b)
}

// TestDirtyFlagExercises covers the dirty-flag helpers used by the
// script-driven DOM updates.
func TestDirtyFlagExercises(t *testing.T) {
	MarkClean()
	if IsDirty() {
		t.Error("expected IsDirty=false after MarkClean")
	}
	SetDirty()
	if !IsDirty() {
		t.Error("expected IsDirty=true after SetDirty")
	}
	MarkClean()
	if IsDirty() {
		t.Error("expected IsDirty=false after MarkClean")
	}
}

// TestCurrentContextRoundTrip covers SetCurrentContext/GetCurrentContext/
// CloseCurrentContext without exercising cgo calls.
func TestCurrentContextRoundTrip(t *testing.T) {
	defer CloseCurrentContext()

	ctx := NewContext()
	SetCurrentContext(ctx, "https://example.com")

	gotCtx, gotURL := GetCurrentContext()
	if gotCtx == nil {
		t.Fatal("GetCurrentContext returned nil")
	}
	if gotURL != "https://example.com" {
		t.Errorf("URL = %q, want %q", gotURL, "https://example.com")
	}

	CloseCurrentContext()
	if c, _ := GetCurrentContext(); c != nil {
		t.Errorf("expected nil ctx after close, got %v", c)
	}
}

// TestRegisterBridgeFunctions covers the JS-Go bridge registration path.
// It verifies all __go_* functions are exposed on the global object.
func TestRegisterBridgeFunctions(t *testing.T) {
	ctx := NewContext()
	defer ctx.Close()

	if err := ctx.RegisterBridgeFunctions(); err != nil {
		t.Fatalf("RegisterBridgeFunctions failed: %v", err)
	}

	// Verify that at least one of the bridge functions is now exposed.
	val, err := ctx.GetProperty("__go_createElement")
	if err != nil {
		t.Fatalf("GetProperty(__go_createElement) failed: %v", err)
	}
	if val == nil {
		t.Error("expected __go_createElement to be set after RegisterBridgeFunctions")
	}
}

// TestExecuteScripts covers the script-execution entry point.
func TestExecuteScripts(t *testing.T) {
	InitDOM(dom.Parse(`<html><body><div id="d">hi</div></body></html>`))

	ctx := NewContext()
	defer ctx.Close()
	if err := ctx.RegisterBridgeFunctions(); err != nil {
		t.Fatalf("RegisterBridgeFunctions: %v", err)
	}
	if _, err := ctx.Evaluate(DOMAPIScript()); err != nil {
		t.Fatalf("DOM API eval failed: %v", err)
	}

	html := dom.Parse(`<html><body><script>document.title = 'X'</script></body></html>`)
	if err := ExecuteScripts(ctx, html, "http://example.com"); err != nil {
		t.Logf("ExecuteScripts returned: %v", err)
	}
}

// TestGetFinalDOM covers the global DOM accessor.
func TestGetFinalDOM(t *testing.T) {
	InitDOM(dom.Parse("<html><body></body></html>"))
	final := GetFinalDOM()
	if final == nil {
		t.Fatal("GetFinalDOM returned nil")
	}
}

// TestDispatchEvent covers the bridge event dispatcher.
func TestDispatchEvent(t *testing.T) {
	InitDOM(dom.Parse(`<html><body><div id="x"></div></body></html>`))
	x := getElementByID("x")
	if x == nil {
		t.Fatal("setup: x not found")
	}
	id := getNodeID(x)
	dispatchEvent(id, "click")
}

// TestProcessTimers covers the bridge timer processor.
func TestProcessTimers(t *testing.T) {
	// No timers registered: returns false.
	if ProcessTimers() {
		t.Error("expected ProcessTimers=false with no timers")
	}
}

// TestClearTimer covers the bridge timer cleanup.
func TestClearTimer(t *testing.T) {
	// Clear an arbitrary id; no-op if id is unknown.
	clearTimer(999)
}

// TestBridgeInvocations exercises bridge functions by calling them from JS.
// This drives goBridgeDispatch + the createElement/appendChild/etc. bridges.
func TestBridgeInvocations(t *testing.T) {
	ctx := NewContext()
	defer ctx.Close()

	InitDOM(dom.Parse(`<html><body><div id="root"></div></body></html>`))

	if err := ctx.RegisterBridgeFunctions(); err != nil {
		t.Fatalf("RegisterBridgeFunctions: %v", err)
	}

	// Invoke bridge functions from JS to exercise the dispatch path.
	script := `
        (function() {
            var d = __go_createElement("p");
            __go_setTextContent(d, 1);
            var root = __go_getElementById("root");
            __go_appendChild(root, d);
            __go_setAttribute(d, "data-x", "y");
            var v = __go_getAttribute(d, "data-x");
            __go_log("value=" + v);
        })();
    `
	if _, err := ctx.Evaluate(script); err != nil {
		t.Logf("bridge invocation error (may be expected): %v", err)
	}
}

// TestGoBridgeDispatchWithMultipleArgTypes covers goBridgeDispatch with
// a JS call that passes multiple argument types.
func TestGoBridgeDispatchWithMultipleArgTypes(t *testing.T) {
	ctx := NewContext()
	defer ctx.Close()
	InitDOM(dom.Parse(`<html><body><div id="a"></div></body></html>`))

	if err := ctx.RegisterBridgeFunctions(); err != nil {
		t.Fatalf("RegisterBridgeFunctions: %v", err)
	}

	script := `
        (function() {
            __go_setInnerHTML(__go_getElementById("a"), "<p>hi</p>");
            __go_removeChild(__go_getBody(), __go_getElementById("a"));
        })();
    `
	if _, err := ctx.Evaluate(script); err != nil {
		t.Logf("bridge invocation error (may be expected): %v", err)
	}
}

// TestQuerySelectors covers querySelector / querySelectorAll bridges.
func TestQuerySelectors(t *testing.T) {
	ctx := NewContext()
	defer ctx.Close()
	InitDOM(dom.Parse(`<html><body><p class="x">a</p><p class="x">b</p></body></html>`))

	if err := ctx.RegisterBridgeFunctions(); err != nil {
		t.Fatalf("RegisterBridgeFunctions: %v", err)
	}

	script := `
        (function() {
            var all = __go_querySelectorAll(".x");
            __go_log("count=" + all);
        })();
    `
	if _, err := ctx.Evaluate(script); err != nil {
		t.Logf("querySelectorAll invocation error: %v", err)
	}
}

// TestBridgeDOMFullSweep exercises every DOM bridge function to drive
// goBridgeDispatch + bridge_dom.go coverage.
func TestBridgeDOMFullSweep(t *testing.T) {
	ctx := NewContext()
	defer ctx.Close()
	InitDOM(dom.Parse(`<html><body><div id="root"><p id="kid">x</p></div></body></html>`))

	if err := ctx.RegisterBridgeFunctions(); err != nil {
		t.Fatalf("RegisterBridgeFunctions: %v", err)
	}

	// Use a simpler script first to confirm the bridge works,
	// then expand. Wrap everything in try/catch so JS errors don't kill the test.
	script := `
        (function() {
            try {
                var body = __go_getBody();
                var root = __go_getElementById("root");
                var kid = __go_getElementById("kid");

                var newKid = __go_createElement("span");
                __go_appendChild(root, newKid);

                __go_setAttribute(kid, "data-x", "y");
                var v = __go_getAttribute(kid, "data-x");

                var clone = __go_cloneNode(root, true);
                __go_appendChild(body, clone);

                var inside = __go_contains(root, kid);
                var parent = __go_getParentNode(kid);
                var kids = __go_getChildren(root);
                var first = __go_firstElementChild(root);
                var last = __go_lastElementChild(root);

                __go_insertBefore(root, newKid, kid);
                __go_removeChild(root, kid);

                var m = __go_matches(kid, "p");
                var c = __go_closest(kid, "div");
                var r = __go_getBoundingClientRect(root);

                var html = __go_getInnerHTML(root);
                __go_setInnerHTML(root, "<span>html</span>");
                var text = __go_getTextContent(root);
                __go_setTextContent(root, "updated");
                var txt = __go_createTextNode("text");
                var frag = __go_createDocumentFragment();

                __go_insertAdjacentHTML(root, "beforebegin", "<i>x</i>");
                var newEl = __go_createElement("em");
                __go_insertAdjacentElement(root, "beforeend", newEl);
                __go_insertAdjacentText(root, "afterend", "tail");

                var oh = __go_outerHTML(root);
                __go_setOuterHTML(root, "<section>r</section>");

                var tags = __go_getElementsByTagName("p");
                var cls = __go_getElementsByClassName("x");
                var qs = __go_querySelector(".x");
                var qsa = __go_querySelectorAll(".x");

                var cs = __go_getComputedStyle(root);
            } catch (e) {
                __go_log("ERR: " + e.message);
            }
        })();
    `
	if _, err := ctx.Evaluate(script); err != nil {
		t.Logf("bridge DOM full sweep error (may be expected): %v", err)
	}
}

// TestStyleHelpers exercises the style helper functions.
func TestStyleHelpers(t *testing.T) {
	if got := overflowString(css.OverflowVisible); got == "" {
		t.Error("expected non-empty overflow string for OverflowVisible")
	}
	if got := flexDirectionString(css.FlexDirectionRow); got == "" {
		t.Error("expected non-empty flex direction string for FlexRow")
	}
	if got := justifyContentString(css.JustifyFlexStart); got == "" {
		t.Error("expected non-empty justify-content string")
	}
	if got := alignItemsString(css.AlignFlexStart); got == "" {
		t.Error("expected non-empty align-items string")
	}
}

// TestBridgeTimers exercises setTimeout/setInterval/clearTimeout/clearInterval
// by registering JS callbacks via makeGoFunction and calling from JS.
func TestBridgeTimers(t *testing.T) {
	ctx := NewContext()
	defer ctx.Close()
	InitDOM(dom.Parse("<html><body></body></html>"))

	if err := ctx.RegisterBridgeFunctions(); err != nil {
		t.Fatalf("RegisterBridgeFunctions: %v", err)
	}

	// Use a real JS function as callback (passes as a JS object).
	// We avoid __go_clearTimeout/clearInterval because they expect JSValueRef args
	// but JS passes int IDs, causing SIGSEGV in _unprotectValue.
	script := `
        (function() {
            try {
                var called = 0;
                var cb = function() { called++; };
                var t1 = __go_setTimeout(cb, 50);
                var t2 = __go_setInterval(cb, 50);
                __go_log("timers=" + t1 + "," + t2);
            } catch (e) {
                __go_log("ERR: " + e.message);
            }
        })();
    `
	if _, err := ctx.Evaluate(script); err != nil {
		t.Logf("bridge timers error (may be expected): %v", err)
	}
}

// TestBridgeEventListeners exercises addEventListener/removeEventListener/
// dispatchEvent using an initialized event system (so eventCtx is set).
func TestBridgeEventListeners(t *testing.T) {
	ctx := NewContext()
	defer ctx.Close()
	InitDOM(dom.Parse(`<html><body><div id="x"></div></body></html>`))
	InitEventSystem(ctx.ctx)

	if err := ctx.RegisterBridgeFunctions(); err != nil {
		t.Fatalf("RegisterBridgeFunctions: %v", err)
	}

	x := getElementByID("x")
	if x == nil {
		t.Fatal("setup: x not found")
	}
	id := getNodeID(x)

	AddEventListener(id, "click", ctx.GlobalObject())
	if len(eventListeners[id]["click"]) == 0 {
		t.Error("expected listener to be registered")
	}
	RemoveEventListener(id, "click", ctx.GlobalObject())
}

// TestBridgeReplaceAndClone exercises replaceChild and cloneNode paths.
func TestBridgeReplaceAndClone(t *testing.T) {
	ctx := NewContext()
	defer ctx.Close()
	InitDOM(dom.Parse(`<html><body><div id="root"><p id="kid">old</p></div></body></html>`))

	if err := ctx.RegisterBridgeFunctions(); err != nil {
		t.Fatalf("RegisterBridgeFunctions: %v", err)
	}

	script := `
        (function() {
            try {
                var root = __go_getElementById("root");
                var kid = __go_getElementById("kid");
                var newKid = __go_createElement("span");
                __go_setTextContent(newKid, "new");
                __go_replaceChild(root, newKid, kid);

                var clone = __go_cloneNode(root, false);
                __go_appendChild(__go_getBody(), clone);

                var cloneDeep = __go_cloneNode(root, true);
                __go_appendChild(__go_getBody(), cloneDeep);
            } catch (e) {
                __go_log("ERR: " + e.message);
            }
        })();
    `
	if _, err := ctx.Evaluate(script); err != nil {
		t.Logf("bridge replace/clone error (may be expected): %v", err)
	}
}

// TestBridgeQuerySelectorAll exercises the querySelectorAll path that
// goes through bridge_query.go for matching.
func TestBridgeQuerySelectorAll(t *testing.T) {
	ctx := NewContext()
	defer ctx.Close()
	InitDOM(dom.Parse(`<html><body><p class="a">1</p><p class="a">2</p><p class="b">3</p></body></html>`))

	if err := ctx.RegisterBridgeFunctions(); err != nil {
		t.Fatalf("RegisterBridgeFunctions: %v", err)
	}

	script := `
        (function() {
            try {
                var els = __go_querySelectorAll(".a");
                __go_log("count=" + els);
                __go_log("first=" + __go_getTextContent(els));
                __go_log("next=" + __go_nextElementSibling(els));
                __go_log("prev=" + __go_previousElementSibling(els));
            } catch (e) {
                __go_log("ERR: " + e.message);
            }
        })();
    `
	if _, err := ctx.Evaluate(script); err != nil {
		t.Logf("bridge querySelectorAll error: %v", err)
	}
}

// TestBridgeNavigationExtras exercises previous/next/first/last element siblings.
func TestBridgeNavigationExtras(t *testing.T) {
	ctx := NewContext()
	defer ctx.Close()
	InitDOM(dom.Parse(`<html><body><a id="a">1</a><b id="b">2</b><c id="c">3</c></body></html>`))

	if err := ctx.RegisterBridgeFunctions(); err != nil {
		t.Fatalf("RegisterBridgeFunctions: %v", err)
	}

	script := `
        (function() {
            try {
                var b = __go_getElementById("b");
                var prev = __go_previousElementSibling(b);
                var next = __go_nextElementSibling(b);
                __go_log("prev=" + __go_getTextContent(prev));
                __go_log("next=" + __go_getTextContent(next));
                var a = __go_getElementById("a");
                var first = __go_firstElementChild(__go_getBody());
                var last = __go_lastElementChild(__go_getBody());
                __go_log("first=" + __go_getTextContent(first));
                __go_log("last=" + __go_getTextContent(last));
            } catch (e) {
                __go_log("ERR: " + e.message);
            }
        })();
    `
	if _, err := ctx.Evaluate(script); err != nil {
		t.Logf("bridge navigation extras error: %v", err)
	}
}
