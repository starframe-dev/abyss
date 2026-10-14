package js

import (
    "os"
    "path/filepath"
    "runtime"
    "strings"
)

var domapiJSSource string

// domapiFragments lists the thematic DOM API fragments in load order. They are
// concatenated inside a single IIFE, preserving the shared closure scope.
var domapiFragments = []string{
    "style",
    "element",
    "global",
}

func init() {
    // Resolve the domapi directory relative to this source file
    _, filename, _, _ := runtime.Caller(0)
    dir := filepath.Join(filepath.Dir(filename), "domapi")
    if _, err := os.Stat(dir); err != nil {
        // Fallback: try relative to the working directory (test binaries)
        dir = "internal/js/domapi"
    }

    var parts []string
    for _, name := range domapiFragments {
        data, err := os.ReadFile(filepath.Join(dir, name+".js"))
        if err != nil {
            panic("cannot read domapi fragment " + name + ".js: " + err.Error())
        }
        parts = append(parts, string(data))
    }

    domapiJSSource = "(function() {\n" + strings.Join(parts, "\n") + "\n})()\n"
}
