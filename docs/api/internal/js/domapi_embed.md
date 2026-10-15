# Embedded DOM API

`internal/js/domapi_embed.go` packages the JavaScript fragments that implement the DOM API and assembles them into one executable script. The fragments are embedded at Go build time, so the resulting source does not need runtime filesystem access.

## Public API

This file declares no exported identifiers. Its variables and function are package-private implementation details of `internal/js`.

## Embedded files and load order

The `//go:embed domapi/*.js` directive embeds every JavaScript file matching that pattern into `domapiFiles`, an `embed.FS`. The assembler reads only the fragments named by `domapiFragments`, in this exact order:

1. `domapi/style.js`
2. `domapi/element.js`
3. `domapi/global.js`

The fragments are joined without reordering, with a newline separator. The comment on `domapiFragments` specifies that the fragments share an IIFE scope; order can therefore matter when later fragments use declarations or state established earlier.

## Initialization and execution

`domapiJSSource` is initialized at package initialization by calling `loadDOMAPI()`. That function reads each named file from the embedded filesystem and constructs a single JavaScript expression:

```js
(function() {
// fragment contents, joined with newlines
})()
```

A trailing newline is included. Wrapping the concatenated code in an immediately invoked function expression gives the fragments a shared local scope while containing their declarations from the surrounding JavaScript scope. The resulting string is intended to be consumed by other code in the package that evaluates or supplies the DOM API source.

## Errors and maintenance considerations

`loadDOMAPI` panics if any listed fragment cannot be read, with a message identifying the fragment and underlying read error. Since the files are embedded at build time, a missing or unreadable listed fragment represents a packaging or source-layout error rather than a runtime filesystem failure.

When adding a fragment, add its base name to `domapiFragments` at the required position. The embed pattern alone does not cause a file to be loaded. Keep the shared-scope assumptions and dependencies between fragments in mind when changing the order.
