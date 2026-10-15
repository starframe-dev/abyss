# DOM API Script

Package `internal/js` exposes `DOMAPIScript`, a Go accessor for the JavaScript source that implements the DOM API bridge.

## Public API

### `DOMAPIScript() string`

Returns the JavaScript source code held in the package-level `domapiJSSource` value. The returned string is intended to be evaluated by the caller in the JavaScript runtime.

The DOM API script must be evaluated **after** `RegisterBridgeFunctions` has registered the bridge functions it relies on. This ordering is part of the API contract; evaluating the script first may leave its bridge dependencies unavailable.

## Implementation details

`DOMAPIScript` is a thin accessor: it returns `domapiJSSource` directly and performs no evaluation, registration, validation, or error handling itself. Callers are responsible for evaluating the returned source and for ensuring bridge registration has already occurred.
