package js

import (
	"embed"
	"strings"
)

//go:embed domapi/*.js
var domapiFiles embed.FS

var domapiJSSource = loadDOMAPI()

// domapiFragments lists the API fragments in load order; they share an IIFE scope.
var domapiFragments = []string{"style", "element", "global"}

func loadDOMAPI() string {
	parts := make([]string, 0, len(domapiFragments))
	for _, name := range domapiFragments {
		data, err := domapiFiles.ReadFile("domapi/" + name + ".js")
		if err != nil {
			panic("cannot load embedded domapi fragment " + name + ".js: " + err.Error())
		}
		parts = append(parts, string(data))
	}
	return "(function() {\n" + strings.Join(parts, "\n") + "\n})()\n"
}
