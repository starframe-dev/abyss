package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/starframe-dev/abyss/internal/ascii"
	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/starframe-dev/abyss/internal/js"
	"github.com/starframe-dev/abyss/internal/layout"
	"github.com/starframe-dev/abyss/internal/render"
)

// testPage is a self-contained HTML with inline JS, flex/grid layout and colored blocks.
const testPage = `<!DOCTYPE html>
<html>
<head>
<style>
  body { margin: 0; padding: 10px; background: #1a1a2e; color: #eee; }
  .flex-row { display: flex; gap: 10px; margin-bottom: 10px; }
  .flex-col { display: flex; flex-direction: column; gap: 5px; width: 200px; }
  .grid-2 { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; margin-bottom: 10px; }
  .red { background: #e74c3c; width: 80px; height: 80px; border-radius: 8px; }
  .green { background: #2ecc71; width: 80px; height: 80px; border-radius: 8px; }
  .blue { background: #3498db; width: 80px; height: 80px; }
  .purple { background: #9b59b6; height: 60px; border-radius: 12px; }
  .yellow { background: #f1c40f; height: 60px; }
  .card { display: flex; flex-direction: column; align-items: center; 
          background: #16213e; border: 2px solid #0f3460; border-radius: 10px; 
          padding: 15px; }
  h1 { color: #e94560; }
  .js-added { background: #e94560; color: white; padding: 5px; border-radius: 5px; }
</style>
<script>
  (function() {
    // Test that __go_log exists
    if (typeof __go_log === 'function') {
      __go_log('__go_log is defined and working');
    }

    // Test basic DOM manipulation
    var el = document.createElement('div');
    el.textContent = 'Created by JS!';
    document.body.appendChild(el);

    var el2 = document.createElement('div');
    el2.textContent = 'Second JS element';
    document.body.appendChild(el2);

    // Test className
    var el3 = document.createElement('div');
    el3.className = 'js-added';
    el3.textContent = 'className test';
    document.body.appendChild(el3);

    // Test classList
    var el4 = document.createElement('div');
    el4.classList.add('red');
    el4.textContent = 'classList test';
    document.body.appendChild(el4);

    // Test style
    var el5 = document.createElement('div');
    el5.style.backgroundColor = '#2ecc71';
    el5.style.color = 'white';
    el5.style.padding = '5px';
    el5.style.borderRadius = '5px';
    el5.textContent = 'style.color test';
    document.body.appendChild(el5);

    // Test getAttribute / setAttribute
    var el6 = document.createElement('div');
    el6.setAttribute('data-test', 'hello');
    var val = el6.getAttribute('data-test');
    el6.textContent = 'getAttribute: ' + val;
    document.body.appendChild(el6);

    // Test parentNode / children
    var parent = document.createElement('div');
    parent.style.border = '1px solid #3498db';
    parent.style.padding = '5px';
    parent.style.marginTop = '10px';
    var child1 = document.createElement('span');
    child1.textContent = 'child1 ';
    var child2 = document.createElement('span');
    child2.textContent = 'child2';
    parent.appendChild(child1);
    parent.appendChild(child2);
    document.body.appendChild(parent);
    console.log('Parent children count: ' + parent.children.length);
    try { console.log('First child text: ' + parent.firstChild.textContent); } catch(e) { __go_log('ERROR: ' + e.message); }
    try { console.log('Last child text: ' + parent.lastChild.textContent); } catch(e) { __go_log('ERROR: ' + e.message); }

    // Test innerHTML getter
    var el7 = document.createElement('div');
    el7.innerHTML = '<span style="color:#f1c40f">innerHTML set</span>';
    document.body.appendChild(el7);
    try { console.log('innerHTML get: ' + el7.innerHTML); } catch(e) { __go_log('ERROR: ' + e.message); }

    // Test document.write
    document.write('<div class="blue" style="height:30px;margin-top:5px">document.write</div>');

    // Test getElementById
    var target = document.createElement('div');
    target.id = 'target-div';
    target.textContent = 'getElementById target';
    document.body.appendChild(target);
    var found = document.getElementById('target-div');
    if (found) {
      try { console.log('getElementById found: ' + found.textContent); } catch(e) { __go_log('ERROR: ' + e.message); }
    }

    // Test getElementsByTagName
    var spans = document.getElementsByTagName('span');
    try { console.log('getElementsByTagName(span) count: ' + spans.length); } catch(e) { __go_log('ERROR: ' + e.message); }

    // Test removeChild
    var removable = document.createElement('div');
    removable.textContent = 'will be removed';
    document.body.appendChild(removable);
    document.body.removeChild(removable);
    try { console.log('removeChild done'); } catch(e) { __go_log('ERROR: ' + e.message); }
  })();
</script>
</head>
<body>
  <h1>Hello from Abyss!</h1>
  <p>This page has flex, grid, colored blocks, and inline JS.</p>

  <div class="flex-row">
    <div class="red"></div>
    <div class="green"></div>
    <div class="blue"></div>
  </div>

  <div class="grid-2">
    <div class="purple">Grid A</div>
    <div class="yellow">Grid B</div>
  </div>

  <div class="flex-row">
    <div class="card">
      <div class="red" style="width:40px;height:40px"></div>
      <span>Card 1</span>
    </div>
    <div class="card">
      <div class="green" style="width:40px;height:40px"></div>
      <span>Card 2</span>
    </div>
  </div>
</body>
</html>`

func main() {
	filePath := filepath.Join(os.TempDir(), "abyss_test_page.html")
	if err := os.WriteFile(filePath, []byte(testPage), 0644); err != nil {
		panic(err)
	}
	defer os.Remove(filePath)

	fmt.Println("=== Rendering test page with inline JS ===")
	fmt.Printf("Page saved to: %s\n\n", filePath)

	// Parse HTML
	doc := dom.Parse(testPage)
	fmt.Printf("DOM parsed: %d children under html\n", len(doc.Children))

	// Execute JavaScript
	jsc := js.NewContext()
	defer jsc.Close()

	js.InitDOM(doc)
	js.LogWriter = os.Stderr

	if err := jsc.RegisterBridgeFunctions(); err != nil {
		fmt.Printf("Bridge init error: %v\n", err)
		os.Exit(1)
	}

	if _, err := jsc.Evaluate(js.DOMAPIScript()); err != nil {
		fmt.Printf("DOM API script error: %v\n", err)
		os.Exit(1)
	}

	if err := js.ExecuteScripts(jsc, doc, ""); err != nil {
		fmt.Printf("Script execution: %v\n", err)
	}

	finalDOM := js.GetFinalDOM()

	if finalDOM != nil {
		fmt.Printf("Final DOM root children: %d\n", len(finalDOM.Children))
		printDOMTree(finalDOM)
	}

	// Render pipeline
	fmt.Println("\n=== Rendering with CSS/Layout/Render ===")
	styles := css.Compute(finalDOM)
	box := layout.Build(finalDOM, styles, 800)
	fb := render.Render(box, 800)
	lines := ascii.Render(fb, box, 800)

	fmt.Printf("Framebuffer: %dx%d\n", fb.Width, fb.Height)
	fmt.Printf("ANSI lines: %d\n\n", len(lines))

	fmt.Println("=== OUTPUT ===")
	for _, line := range lines {
		fmt.Println(line)
	}
	fmt.Println("=== END ===")

	// Let user observe
	time.Sleep(2 * time.Second)
}

func printDOMTree(root *dom.Node) {
	for _, c := range root.Children {
		if c.Type == dom.ElementNode {
			fmt.Printf("  <%s> (%d children)\n", c.TagName, len(c.Children))
			if c.TagName == "html" {
				printHTMLBody(c)
			}
		}
	}
}

func printHTMLBody(htmlEl *dom.Node) {
	for _, body := range htmlEl.Children {
		if body.Type != dom.ElementNode || body.TagName != "body" {
			continue
		}
		fmt.Printf("    <body> (%d children):\n", len(body.Children))
		for _, bc := range body.Children {
			if bc.Type == dom.ElementNode {
				fmt.Printf("      <%s> class=%s text='%s'\n", bc.TagName, bc.Attributes["class"], bc.TextContent())
			} else if bc.Type == dom.TextNode {
				text := strings.TrimSpace(bc.Text)
				if len(text) > 40 {
					text = text[:40] + "..."
				}
				fmt.Printf("      text: '%s'\n", text)
			}
		}
	}
}
