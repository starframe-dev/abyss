package js

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/dom"
)

func TestPublicDOMAPI(t *testing.T) {
	InitDOM(dom.Parse(`<html><body><div id="root"><p class="x">one</p><p>two</p></div></body></html>`))
	defer resetDOMState()
	body := GetBody()
	if body == nil { t.Fatal("GetBody") }
	root := GetElementByID("root")
	if root == nil { t.Fatal("GetElementByID") }
	rid := GetNodeID(root)
	if n, ok := GetNodeByID(rid); !ok || n != root { t.Fatal("GetNodeByID") }
	el := CreateElement("article")
	text := CreateTextNode("hello")
	frag := CreateDocumentFragment()
	if el == nil || text == nil || frag == nil { t.Fatal("creation") }
	AppendChild(rid, GetNodeID(el))
	AppendChild(GetNodeID(el), GetNodeID(text))
	if !Contains(rid, GetNodeID(text)) { t.Fatal("Contains") }
	InsertBefore(rid, GetNodeID(CreateElement("b")), GetNodeID(el))
	ReplaceChild(rid, GetNodeID(CreateElement("i")), GetNodeID(el))
	SetAttribute(rid, "data-k", "v")
	if GetAttribute(rid, "data-k") != "v" { t.Fatal("attribute") }
	SetTextContent(rid, "content")
	if GetTextContent(rid) != "content" { t.Fatal("text content") }
	SetInnerHTML(rid, `<span>html</span><strong>tail</strong>`)
	if GetInnerHTML(rid) == "" { t.Fatal("inner html") }
	if FirstElementChild(rid) == nil || LastElementChild(rid) == nil { t.Fatal("element children") }
	if QuerySelector(rid, "span") == nil || len(QuerySelectorAll(rid, "span")) != 1 { t.Fatal("query selector") }
	span := QuerySelector(rid, "span")
	if !Matches(GetNodeID(span), "span") || Closest(GetNodeID(span), "div") != root { t.Fatal("matches/closest") }
	clone := CloneNode(rid, true)
	if clone == nil { t.Fatal("clone") }
	RemoveChild(rid, GetNodeID(LastElementChild(rid)))
	if _, ok := GetNodeByID(-1); ok { t.Fatal("invalid lookup") }
	if CreateElement("") == nil || frag == nil { t.Fatal("creation helpers") }
}
