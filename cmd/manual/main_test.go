package main

import (
    "bytes"
    "io"
    "os"
    "strings"
    "testing"

    "github.com/starframe/abyss/internal/dom"
)

func TestPrintDOMTree_HTMLBody(t *testing.T) {
    doc := dom.Parse(`<html><body><h1>hi</h1><p>world</p></body></html>`)

    old := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w
    defer func() { os.Stdout = old }()

    printDOMTree(doc)
    printHTMLBody(doc.Children[0])

    _ = w.Close()
    var buf bytes.Buffer
    _, _ = io.Copy(&buf, r)
    out := buf.String()

    if !strings.Contains(out, "<body>") {
        t.Errorf("expected <body> in output, got: %s", out)
    }
    if !strings.Contains(out, "<h1>") {
        t.Errorf("expected <h1> in output, got: %s", out)
    }
}

func TestTestPageContainsHTMLStructure(t *testing.T) {
    if !strings.Contains(testPage, "<html>") {
        t.Error("testPage missing <html>")
    }
    if !strings.Contains(testPage, "display: flex") {
        t.Error("testPage missing flex CSS")
    }
    if !strings.Contains(testPage, "display: grid") {
        t.Error("testPage missing grid CSS")
    }
    if !strings.Contains(testPage, "class=\"red\"") {
        t.Error("testPage missing red class")
    }
    if !strings.Contains(testPage, "<script>") {
        t.Error("testPage missing inline JS")
    }
}
