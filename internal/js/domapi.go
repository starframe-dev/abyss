package js

// DOMAPIScript returns the JavaScript source code for the DOM API bridge.
// It must be evaluated after RegisterBridgeFunctions.
func DOMAPIScript() string {
    return domapiJSSource
}