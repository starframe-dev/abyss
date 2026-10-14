package js

/*
#include <JavaScriptCore/JavaScriptCore.h>
#include <stdlib.h>
*/
import "C"
import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"
	"unsafe"

	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/starframe-dev/abyss/internal/fetcher"
)

// nodeToJS создаёт JS-объект-обертку для узла DOM.
func nodeToJS(ctx *Context, n *dom.Node) C.JSValueRef {
	id := getNodeID(n)
	if id < 0 {
		return ctx.MakeNull()
	}

	// Create a JS object
	jsObj := C.JSObjectMake(ctx.ctx, nil, nil)

	// Set __go_id
	goID := C.CString("__go_id")
	defer C.free(unsafe.Pointer(goID))
	propName := C.JSStringCreateWithUTF8CString(goID)
	defer C.JSStringRelease(propName)
	C.JSObjectSetProperty(ctx.ctx, jsObj, propName, ctx.MakeNumber(float64(id)), C.kJSPropertyAttributeNone, nil)

	// Set nodeType
	nodeTypeName := C.CString("nodeType")
	defer C.free(unsafe.Pointer(nodeTypeName))
	propName2 := C.JSStringCreateWithUTF8CString(nodeTypeName)
	defer C.JSStringRelease(propName2)
	switch n.Type {
	case dom.ElementNode:
		C.JSObjectSetProperty(ctx.ctx, jsObj, propName2, ctx.MakeNumber(1), C.kJSPropertyAttributeNone, nil)
	case dom.TextNode:
		C.JSObjectSetProperty(ctx.ctx, jsObj, propName2, ctx.MakeNumber(3), C.kJSPropertyAttributeNone, nil)
	case dom.CommentNode:
		C.JSObjectSetProperty(ctx.ctx, jsObj, propName2, ctx.MakeNumber(8), C.kJSPropertyAttributeNone, nil)
	case dom.DocumentNode:
		C.JSObjectSetProperty(ctx.ctx, jsObj, propName2, ctx.MakeNumber(9), C.kJSPropertyAttributeNone, nil)
	case dom.DocumentFragmentNode:
		C.JSObjectSetProperty(ctx.ctx, jsObj, propName2, ctx.MakeNumber(11), C.kJSPropertyAttributeNone, nil)
	}

	// Set tagName for elements
	if n.Type == dom.ElementNode {
		tagName := C.CString("tagName")
		defer C.free(unsafe.Pointer(tagName))
		propName3 := C.JSStringCreateWithUTF8CString(tagName)
		defer C.JSStringRelease(propName3)
		C.JSObjectSetProperty(ctx.ctx, jsObj, propName3, ctx.MakeString(strings.ToUpper(n.TagName)), C.kJSPropertyAttributeNone, nil)
	}

	// Set textContent for text nodes
	if n.Type == dom.TextNode {
		textContent := C.CString("textContent")
		defer C.free(unsafe.Pointer(textContent))
		propName4 := C.JSStringCreateWithUTF8CString(textContent)
		defer C.JSStringRelease(propName4)
		C.JSObjectSetProperty(ctx.ctx, jsObj, propName4, ctx.MakeString(n.Text), C.kJSPropertyAttributeNone, nil)
	}

	return jsObjToVal(jsObj)
}

// GetFinalDOM возвращает финальное DOM-дерево после выполнения JavaScript-скриптов.
func GetFinalDOM() *dom.Node {
	return rootDoc
}

// ExecuteScripts выполняет все скрипты из тегов <script> в документе.
func ExecuteScripts(ctx *Context, doc *dom.Node, baseURL string) error {
	return ExecuteScriptsContext(context.Background(), ctx, doc, baseURL)
}

const (
	maxScriptElements  = 128
	maxPageScriptBytes = 8 << 20
)

// ExecuteScriptsContext fetches external scripts with caller cancellation and
// limits the total script work accepted from one page.
func ExecuteScriptsContext(execCtx context.Context, ctx *Context, doc *dom.Node, baseURL string) error {
	execCtx, cancel := context.WithTimeout(execCtx, 30*time.Second)
	defer cancel()
	scripts := doc.QueryAll("script")
	if len(scripts) > maxScriptElements {
		scripts = scripts[:maxScriptElements]
	}
	remainingBytes := maxPageScriptBytes
	for _, script := range scripts {
		if err := execCtx.Err(); err != nil {
			return err
		}
		if remainingBytes <= 0 {
			break
		}
		src := script.GetAttr("src")
		if src != "" {
			// External script - fetch and evaluate
			code, err := fetchScriptContext(execCtx, baseURL, src)
			if err != nil {
				// Silently skip - most pages have JS errors
				continue
			}
			if len(code) > remainingBytes {
				break
			}
			remainingBytes -= len(code)
			_, err = ctx.Evaluate(code)
			if err != nil {
				// Silently skip - most pages have JS errors
			}
			continue
		}
		// Inline script
		code := script.TextContent()
		if strings.TrimSpace(code) == "" {
			continue
		}
		if len(code) > remainingBytes {
			break
		}
		remainingBytes -= len(code)
		_, err := ctx.Evaluate(code)
		if err != nil {
			// Silently skip - most pages have JS errors
		}
	}
	return nil
}

// fetchScript запрашивает и возвращает содержимое внешнего JavaScript-скрипта.
func fetchScript(baseURL, src string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return fetchScriptContext(ctx, baseURL, src)
}

func fetchScriptContext(parent context.Context, baseURL, src string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("parse base URL %s: %w", baseURL, err)
	}
	ref, err := url.Parse(src)
	if err != nil {
		return "", fmt.Errorf("parse script src %s: %w", src, err)
	}
	resolved := base.ResolveReference(ref)

	ctx, cancel := context.WithTimeout(parent, 3*time.Second)
	defer cancel()
	return fetcher.FetchJS(ctx, resolved.String())
}
