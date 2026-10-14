package ascii

import (
    "regexp"
    "sort"
    "strings"
    "testing"

    "github.com/starframe/abyss/internal/css"
    "github.com/starframe/abyss/internal/dom"
    "github.com/starframe/abyss/internal/layout"
    "github.com/starframe/abyss/internal/render"
)

// TestColors_RichPage verifies that Abyss's render pipeline produces diverse
// colors when the source HTML contains colorful CSS styles.
func TestColors_RichPage(t *testing.T) {
    html := `
<!DOCTYPE html>
<html>
<head>
<style>
body {
    background: #ffffff;
    color: #1f2937;
    margin: 0;
    padding: 20px;
}
h1 {
    color: #fbbf24;
    font-size: 48px;
    border-bottom: 3px solid #ef4444;
    background-color: #1e293b;
}
h2 {
    color: #34d399;
}
a {
    color: #60a5fa;
    text-decoration: underline;
}
.red-box {
    background-color: #ef4444;
    color: #ffffff;
    padding: 20px;
    margin: 10px;
}
.blue-box {
    background-color: #3b82f6;
    color: #fef3c7;
    padding: 15px;
}
.green-box {
    background-color: #10b981;
    color: #1f2937;
    padding: 10px;
}
.yellow-box {
    background-color: #f59e0b;
    color: #1f2937;
}
.purple {
    background-color: #9333ea;
    color: #fef3c7;
}
#special {
    background-color: #1e293b;
    border: 5px solid #06b6d4;
    padding: 20px;
    color: #fbbf24;
}
</style>
</head>
<body>
<h1>Colorful Test Page</h1>
<p>This is a regular paragraph with <a href="#">a blue link</a> and some text.</p>
<h2>Colored Boxes</h2>
<div class="red-box">Red Box</div>
<div class="blue-box">Blue Box</div>
<div class="green-box">Green Box</div>
<div class="yellow-box">Yellow Box</div>
<div class="purple">Purple Box</div>
<div id="special">
    <p>Special bordered box</p>
    <p>Visit <a href="https://example.com">example.com</a> for more.</p>
</div>
</body>
</html>`

    doc := dom.Parse(html)
    styles := css.Compute(doc)
    box := layout.Build(doc, styles, 320)
    fb := render.Render(box, 320)
    lines := Render(fb, box, 320)
    ansiLines := RenderANSI(fb, box, 320)

    if len(ansiLines) == 0 {
        t.Fatal("no ANSI lines generated")
    }
    if len(lines) == 0 {
        t.Fatal("no plain lines generated")
    }

    // Collect all unique true-color codes used in render output
    ansiRe := regexp.MustCompile(`\x1b\[(?:38|48);2;\d+;\d+;\d+`)
    colors := map[string]bool{}
    for _, line := range ansiLines {
        for _, m := range ansiRe.FindAllString(line, -1) {
            colors[m] = true
        }
    }

    if len(colors) < 5 {
        t.Logf("Found only %d unique true-color codes:", len(colors))
        sorted := sortKeys(colors)
        for _, c := range sorted {
            t.Logf("  %s\n", c)
        }
        t.Logf("Rendered ANSI (first 8 lines):\n%s", joinRender(ansiLines, 8))
        t.Fatalf("expected at least 5 unique true-color codes, got %d", len(colors))
    }

    t.Logf("✅ Found %d unique true-color codes:", len(colors))
    sorted := sortKeys(colors)
    for _, c := range sorted {
        t.Logf("  %s", c)
    }

    // Also verify specific colors are present (proves styles are applied).
    mustContain := []string{
        `38;2;251;191;36`,               // h1 color #fbbf24 (yellow)
        `48;2;30;41;59`,                 // h1 background #1e293b (dark blue)
        `38;2;251;191;36;48;2;30;41;59`, // heading text must sit on its own background
        `38;2;96;165;250`,               // a link color #60a5fa (light blue)
        `48;2;239;68;68`,                // red-box bg #ef4444
        `48;2;59;130;246`,               // blue-box bg #3b82f6
        `48;2;16;185;129`,               // green-box bg #10b981
        `48;2;245;158;11`,               // yellow-box bg #f59e0b
        `48;2;147;51;234`,               // purple bg #9333ea
    }
    rendered := strings.Join(ansiLines, "")
    for _, want := range mustContain {
        if !strings.Contains(rendered, want) {
            t.Errorf("missing color code %s — CSS color may not be flowing through pipeline", want)
        }
    }
}

func TestRender_ImagePlaceholder(t *testing.T) {
    doc := dom.Parse(`<div><img alt="Article preview" src="https://example.com/image.png"></div>`)
    styles := css.Compute(doc)
    box := layout.Build(doc, styles, 80)
    fb := render.Render(box, 80)

    rendered := strings.Join(Render(fb, box, 80), "\n")
    if !strings.Contains(rendered, "[image] Article preview") {
        t.Fatalf("image placeholder is missing from render: %q", rendered)
    }
}

func sortKeys(m map[string]bool) []string {
    keys := make([]string, 0, len(m))
    for k := range m {
        keys = append(keys, k)
    }
    sort.Strings(keys)
    return keys
}

func joinRender(lines []string, max int) string {
    if len(lines) < max {
        max = len(lines)
    }
    var b strings.Builder
    for i := 0; i < max; i++ {
        b.WriteString(lines[i])
        b.WriteString("\n")
    }
    return b.String()
}
