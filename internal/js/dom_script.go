package js

/*
#include <JavaScriptCore/JavaScriptCore.h>
*/
import "C"
import (
    "fmt"
    "io"
    "net/http"
    "net/url"
    "strings"
    "time"

    "github.com/starframe/abyss/internal/dom"
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
    propName := C.JSStringCreateWithUTF8CString(C.CString("__go_id"))
    defer C.JSStringRelease(propName)
    C.JSObjectSetProperty(ctx.ctx, jsObj, propName, ctx.MakeNumber(float64(id)), C.kJSPropertyAttributeNone, nil)

    // Set nodeType
    propName2 := C.JSStringCreateWithUTF8CString(C.CString("nodeType"))
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
        propName3 := C.JSStringCreateWithUTF8CString(C.CString("tagName"))
        defer C.JSStringRelease(propName3)
        C.JSObjectSetProperty(ctx.ctx, jsObj, propName3, ctx.MakeString(strings.ToUpper(n.TagName)), C.kJSPropertyAttributeNone, nil)
    }

    // Set textContent for text nodes
    if n.Type == dom.TextNode {
        propName4 := C.JSStringCreateWithUTF8CString(C.CString("textContent"))
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
    scripts := doc.QueryAll("script")
    for _, script := range scripts {
        src := script.GetAttr("src")
        if src != "" {
            // External script - fetch and evaluate
            code, err := fetchScript(baseURL, src)
            if err != nil {
                // Silently skip - most pages have JS errors
                continue
            }
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
        _, err := ctx.Evaluate(code)
        if err != nil {
            // Silently skip - most pages have JS errors
        }
    }
    return nil
}

// fetchScript запрашивает и возвращает содержимое внешнего JavaScript-скрипта.
func fetchScript(baseURL, src string) (string, error) {
    base, err := url.Parse(baseURL)
    if err != nil {
        return "", fmt.Errorf("parse base URL %s: %w", baseURL, err)
    }
    ref, err := url.Parse(src)
    if err != nil {
        return "", fmt.Errorf("parse script src %s: %w", src, err)
    }
    resolved := base.ResolveReference(ref)

    // Use a short timeout for script fetching
    client := &http.Client{Timeout: 3 * time.Second}
    req, reqErr := http.NewRequest("GET", resolved.String(), nil)
    if reqErr != nil {
        return "", reqErr
    }
    req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15")
    resp, respErr := client.Do(req)
    if respErr != nil {
        return "", respErr
    }
    defer resp.Body.Close()
    body, readErr := io.ReadAll(resp.Body)
    if readErr != nil {
        return "", readErr
    }
    return string(body), nil
}
