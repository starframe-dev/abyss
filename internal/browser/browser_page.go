package browser

import (
    "encoding/base64"
    "fmt"
    "image"
    "io"
    "net/http"
    "net/url"
    "os"
    "strings"
    "sync"
    "time"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/starframe/abyss/internal/ascii"
    "github.com/starframe/abyss/internal/asciirender"
    "github.com/starframe/abyss/internal/css"
    "github.com/starframe/abyss/internal/dom"
    "github.com/starframe/abyss/internal/fetcher"
    "github.com/starframe/abyss/internal/js"
    "github.com/starframe/abyss/internal/layout"
    "github.com/starframe/abyss/internal/render"
)

// imageCache stores decoded raster images keyed by absolute URL so a single
// page load fetches each image only once.
var (
    imageCacheMu sync.RWMutex
    imageCache   = make(map[string]image.Image)
)

func cachedImage(src string) image.Image {
    imageCacheMu.RLock()
    img, ok := imageCache[src]
    imageCacheMu.RUnlock()
    if ok {
        return img
    }
    if img := decodeDataImage(src); img != nil {
        storeImage(src, img)
        return img
    }
    return nil
}

// decodeDataImage parses a data: URL such as
// "data:image/png;base64,iVBORw0..." and decodes its body into an image.
// Returns nil if the URL is not a data: URL or decoding fails.
func decodeDataImage(src string) image.Image {
    const prefix = "data:"
    if !strings.HasPrefix(src, prefix) {
        return nil
    }
    comma := strings.IndexByte(src, ',')
    if comma < 0 {
        return nil
    }
    meta := src[len(prefix):comma]
    body := src[comma+1:]
    if !strings.Contains(meta, "base64") {
        return nil
    }
    raw, err := base64.StdEncoding.DecodeString(body)
    if err != nil {
        return nil
    }
    img, err := asciirender.Decode(raw)
    if err != nil {
        return nil
    }
    return img
}

func storeImage(src string, img image.Image) {
    if img == nil {
        return
    }
    imageCacheMu.Lock()
    imageCache[src] = img
    imageCacheMu.Unlock()
}

func imgOrNil(img image.Image, ok bool) image.Image {
    if !ok {
        return nil
    }
    return img
}

func init() {
    render.ImageProvider = cachedImage
}

// fetchImages walks the document looking for <img src> elements, fetches
// each URL in parallel, decodes the body as an image, and stores the
// decoded image in the package cache.
func fetchImages(doc *dom.Node, baseURL string) {
    imgs := doc.QueryAll("img")
    if len(imgs) == 0 {
        return
    }

    seen := make(map[string]struct{}, len(imgs))
    var sources []string
    for _, node := range imgs {
        src := strings.TrimSpace(node.GetAttr("src"))
        if src == "" || strings.HasPrefix(src, "data:") {
            continue
        }
        abs := ResolveURL(baseURL, src)
        if abs == "" {
            continue
        }
        if _, dup := seen[abs]; dup {
            continue
        }
        seen[abs] = struct{}{}
        sources = append(sources, abs)
    }
    if len(sources) == 0 {
        return
    }

    semaphore := make(chan struct{}, 4)
    var wg sync.WaitGroup
    for _, src := range sources {
        wg.Add(1)
        go func(resource string) {
            defer wg.Done()
            semaphore <- struct{}{}
            defer func() { <-semaphore }()

            imageCacheMu.RLock()
            _, cached := imageCache[resource]
            imageCacheMu.RUnlock()
            if cached {
                return
            }

            body, _, err := fetcher.FetchBytes(resource)
            if err != nil || len(body) == 0 {
                return
            }
            img, err := asciirender.Decode(body)
            if err != nil {
                return
            }
            storeImage(resource, img)
        }(src)
    }
    wg.Wait()
}

func loadPage(pageURL string, viewportWidth int) tea.Cmd {
    return func() tea.Msg {
        html, err := fetcher.Fetch(pageURL)
        if err != nil {
            return PageErrorMsg{Err: err}
        }
        doc := dom.Parse(html)

        js.CloseCurrentContext()

        context := js.NewContext()
        js.SetCurrentContext(context, pageURL)
        js.InitDOM(doc)

        if err := context.RegisterBridgeFunctions(); err != nil {
            return PageErrorMsg{Err: fmt.Errorf("js bridge init: %w", err)}
        }

        js.InitEventSystem(context.GetContextRef())

        if _, err := context.Evaluate(js.DOMAPIScript()); err != nil {
            return PageErrorMsg{Err: fmt.Errorf("js dom api error: %w", err)}
        }

        safeURL := strings.ReplaceAll(pageURL, "\\", "\\\\")
        safeURL = strings.ReplaceAll(safeURL, "'", "\\'")
        safeURL = strings.ReplaceAll(safeURL, "\n", "\\n")
        _, _ = context.Evaluate(fmt.Sprintf("document._url='%s';", safeURL))
        _, _ = context.Evaluate(`window.getComputedStyle=function(el){if(!el||el.__go_id===undefined)return null;return __go_getComputedStyle(el);};`)

        _ = js.ExecuteScripts(context, doc, pageURL)
        _, _ = context.Evaluate(`if(document.dispatchEvent){var ev={type:'DOMContentLoaded',defaultPrevented:false,_propagationStopped:false,preventDefault:function(){this.defaultPrevented=true;},stopPropagation:function(){this._propagationStopped=true;}};document.dispatchEvent(ev);}`)

        finalDOM := js.GetFinalDOM()
        externalCSS := fetchExternalCSS(finalDOM, pageURL)
        fetchImages(finalDOM, pageURL)
        styles := css.ComputeWithExternal(finalDOM, externalCSS)
        js.SetComputedStyles(styles)
        box := layout.Build(finalDOM, styles, float64(viewportWidth))
        js.SetLayoutBoxes(buildLayoutBoxMap(box))
        framebuffer := render.Render(box, viewportWidth)
        lines := ascii.Render(framebuffer, box, viewportWidth)
        ansiLines := ascii.RenderANSI(framebuffer, box, viewportWidth)
        if path := os.Getenv("ABYSS_DUMP_ANSI"); path != "" {
            dump := strings.Join(ansiLines, "\n")
            _ = os.WriteFile(path, []byte(dump), 0o644)
            fmt.Fprintf(os.Stderr, "abyss: dumped %d bytes to %s\n", len(dump), path)
        }
        return PageLoadedMsg{Lines: lines, ANSI: ansiLines, URL: pageURL}
    }
}

func reRenderCmd(viewportWidth int, zoomFactor float64) tea.Cmd {
    return func() tea.Msg {
        context, _ := js.GetCurrentContext()
        if context == nil {
            return nil
        }
        js.MarkClean()

        finalDOM := js.GetFinalDOM()
        if finalDOM == nil {
            return nil
        }

        scaledWidth := int(float64(viewportWidth) * zoomFactor)
        if scaledWidth < 10 {
            scaledWidth = 10
        }

        styles := css.Compute(finalDOM)
        js.SetComputedStyles(styles)
        box := layout.Build(finalDOM, styles, float64(scaledWidth))
        js.SetLayoutBoxes(buildLayoutBoxMap(box))
        framebuffer := render.Render(box, scaledWidth)
        lines := ascii.Render(framebuffer, box, scaledWidth)
        ansiLines := ascii.RenderANSI(framebuffer, box, scaledWidth)
        return PageLoadedMsg{Lines: lines, ANSI: ansiLines}
    }
}

// fetchExternalCSS finds <link rel="stylesheet"> tags and fetches their CSS content.
const (
    maxExternalStylesheets = 64
    stylesheetWorkers      = 8
)

func fetchExternalCSS(doc *dom.Node, baseURL string) []string {
    links := doc.FindLinkStylesheets()
    if len(links) == 0 {
        return nil
    }

    urls := make([]string, 0, len(links))
    seen := make(map[string]struct{}, len(links))
    for _, link := range links {
        href := link.GetAttr("href")
        if href == "" {
            continue
        }
        cssURL := ResolveURL(baseURL, href)
        if cssURL == "" {
            continue
        }
        if _, ok := seen[cssURL]; ok {
            continue
        }
        if len(urls) >= maxExternalStylesheets {
            break
        }
        seen[cssURL] = struct{}{}
        urls = append(urls, cssURL)
    }

    if len(urls) == 0 {
        return nil
    }

    results := make([]string, len(urls))
    client := &http.Client{Timeout: 2 * time.Second}
    semaphore := make(chan struct{}, stylesheetWorkers)
    var wg sync.WaitGroup
    for i, cssURL := range urls {
        wg.Add(1)
        go func(index int, resourceURL string) {
            defer wg.Done()
            semaphore <- struct{}{}
            defer func() { <-semaphore }()

            req, err := http.NewRequest(http.MethodGet, resourceURL, nil)
            if err != nil {
                return
            }
            req.Header.Set("User-Agent", "Mozilla/5.0")
            resp, err := client.Do(req)
            if err != nil {
                return
            }
            defer func() { _ = resp.Body.Close() }()
            cssBytes, err := io.ReadAll(resp.Body)
            if err == nil {
                results[index] = string(cssBytes)
            }
        }(i, cssURL)
    }
    wg.Wait()

    result := make([]string, 0, len(results))
    for _, cssText := range results {
        if cssText != "" {
            result = append(result, cssText)
        }
    }
    return result
}

// ResolveURL returns the absolute URL by combining the base URL with a relative reference.
func ResolveURL(base, rel string) string {
    baseURL, err := url.Parse(base)
    if err != nil {
        return rel
    }
    relURL, err := url.Parse(rel)
    if err != nil {
        return ""
    }
    return baseURL.ResolveReference(relURL).String()
}

// ResolveURLForTest exposes ResolveURL for tests and external callers.
func ResolveURLForTest(base, rel string) string {
    return ResolveURL(base, rel)
}

// FetchExternalCSSForTest exposes fetchExternalCSS for tests and external callers.
func FetchExternalCSSForTest(doc *dom.Node, baseURL string) []string {
    return fetchExternalCSS(doc, baseURL)
}
