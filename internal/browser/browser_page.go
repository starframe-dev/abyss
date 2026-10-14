package browser

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"net/url"
	"os"
	"strings"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/starframe-dev/abyss/internal/ascii"
	"github.com/starframe-dev/abyss/internal/asciirender"
	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/starframe-dev/abyss/internal/fetcher"
	"github.com/starframe-dev/abyss/internal/js"
	"github.com/starframe-dev/abyss/internal/layout"
	"github.com/starframe-dev/abyss/internal/render"
)

const (
	maxDataURLBytes       = 16 << 20
	maxImageCacheBytes    = 64 << 20
	maxExternalImages     = 64
	maxExternalImageBytes = 64 << 20
)

var (
	jsLifecycleMu   sync.Mutex
	jsRuntimeTabID  uint64
	jsRuntimeActive bool
	jsRuntimes      = make(map[uint64]*js.Runtime)
	imageCacheMu    sync.RWMutex
	imageCache      = make(map[string]image.Image)
	imageCacheOrder []string
	imageCacheBytes int64
)

func imageMemory(img image.Image) int64 {
	b := img.Bounds()
	if b.Dx() <= 0 || b.Dy() <= 0 || b.Dx() > asciirender.MaxImagePixels/b.Dy() {
		return 0
	}
	// RGBA64 decoders can require eight bytes per pixel; use that conservative
	// estimate so the cache budget cannot be exceeded by under-accounting.
	return int64(b.Dx()) * int64(b.Dy()) * 8
}

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

func decodeDataImage(src string) image.Image {
	const prefix = "data:"
	if !strings.HasPrefix(src, prefix) {
		return nil
	}
	comma := strings.IndexByte(src, ',')
	if comma < 0 || len(src) > maxDataURLBytes {
		return nil
	}
	meta, body := src[len(prefix):comma], src[comma+1:]
	if !strings.Contains(meta, "base64") || len(body) > base64.StdEncoding.EncodedLen(maxDataURLBytes) {
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
	size := imageMemory(img)
	if size <= 0 || size > maxImageCacheBytes {
		return
	}
	imageCacheMu.Lock()
	defer imageCacheMu.Unlock()
	if old, ok := imageCache[src]; ok {
		imageCacheBytes -= imageMemory(old)
		delete(imageCache, src)
		imageCacheOrder = removeImageCacheKey(imageCacheOrder, src)
	}
	for imageCacheBytes+size > maxImageCacheBytes && len(imageCacheOrder) > 0 {
		key := imageCacheOrder[0]
		imageCacheOrder[0] = ""
		imageCacheOrder = imageCacheOrder[1:]
		if old, ok := imageCache[key]; ok {
			imageCacheBytes -= imageMemory(old)
			delete(imageCache, key)
		}
	}
	imageCache[src] = img
	imageCacheOrder = append(imageCacheOrder, src)
	imageCacheBytes += size
}

func resetImageCache() {
	imageCacheMu.Lock()
	imageCache = make(map[string]image.Image)
	imageCacheOrder = nil
	imageCacheBytes = 0
	imageCacheMu.Unlock()
}

func removeImageCacheKey(keys []string, target string) []string {
	for i, key := range keys {
		if key == target {
			copy(keys[i:], keys[i+1:])
			keys[len(keys)-1] = ""
			return keys[:len(keys)-1]
		}
	}
	return keys
}

func imgOrNil(img image.Image, ok bool) image.Image {
	if !ok {
		return nil
	}
	return img
}

func init() { render.ImageProvider = cachedImage }

func fetchImages(doc *dom.Node, baseURL string) {
	fetchImagesContext(context.Background(), doc, baseURL)
}

func fetchImagesContext(ctx context.Context, doc *dom.Node, baseURL string) {
	imgs := doc.QueryAll("img")
	seen := make(map[string]struct{}, min(len(imgs), maxExternalImages))
	sources := make([]string, 0, min(len(imgs), maxExternalImages))
	imageCacheMu.RLock()
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
		if _, cached := imageCache[abs]; !cached {
			sources = append(sources, abs)
		}
		if len(seen) >= maxExternalImages {
			break
		}
	}
	imageCacheMu.RUnlock()
	if len(sources) == 0 {
		return
	}

	type imageJob struct {
		index int
		url   string
	}
	type imageResult struct {
		index     int
		url       string
		bodyBytes int64
		img       image.Image
	}
	workerCount := min(2, len(sources))
	jobs := make(chan imageJob)
	fetched := make(chan imageResult, workerCount)
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				body, _, err := fetcher.FetchBytesContext(ctx, job.url)
				item := imageResult{index: job.index, url: job.url, bodyBytes: int64(len(body))}
				if errors.Is(err, fetcher.ErrResponseTooLarge) {
					item.bodyBytes = fetcher.MaxImageBytes
				}
				var readErr *fetcher.ResponseReadError
				if errors.As(err, &readErr) {
					item.bodyBytes = readErr.BytesRead
				}
				if err == nil && len(body) > 0 {
					item.img, _ = asciirender.Decode(body)
				}
				fetched <- item
			}
		}()
	}
	launched, next, budgetUsed := 0, 0, int64(0)
	launch := func() {
		if launched >= len(sources) || budgetUsed >= maxExternalImageBytes || ctx.Err() != nil {
			return
		}
		jobs <- imageJob{index: launched, url: sources[launched]}
		launched++
	}
	for i := 0; i < workerCount && ctx.Err() == nil; i++ {
		launch()
	}
	pending := make(map[int]imageResult, workerCount)
	for next < launched {
		item := <-fetched
		pending[item.index] = item
		for {
			current, ok := pending[next]
			if !ok {
				break
			}
			delete(pending, next)
			if current.bodyBytes > 0 {
				if current.bodyBytes <= maxExternalImageBytes-budgetUsed {
					budgetUsed += current.bodyBytes
					if current.img != nil {
						storeImage(current.url, current.img)
					}
				} else {
					budgetUsed = maxExternalImageBytes
				}
			}
			next++
			launch()
		}
	}
	close(jobs)
	wg.Wait()
}

func closeJSCContext() {
	jsLifecycleMu.Lock()
	defer jsLifecycleMu.Unlock()
	if context, _ := js.GetCurrentContext(); context != nil {
		js.CloseCurrentContext()
	}
	jsRuntimeTabID = 0
	jsRuntimeActive = false
	for tabID, runtime := range jsRuntimes {
		delete(jsRuntimes, tabID)
		js.CloseRuntime(runtime)
	}
}

func suspendActiveJSCContextLocked() {
	context, _ := js.GetCurrentContext()
	if context == nil {
		return
	}
	jsRuntimes[jsRuntimeTabID] = js.CaptureRuntime()
	jsRuntimeTabID = 0
	jsRuntimeActive = false
}

func activateJSCContextForTabLocked(tabID uint64) bool {
	if jsRuntimeActive && jsRuntimeTabID == tabID {
		if context, _ := js.GetCurrentContext(); context != nil {
			return true
		}
	}
	suspendActiveJSCContextLocked()
	runtime, ok := jsRuntimes[tabID]
	if !ok || runtime == nil {
		return false
	}
	delete(jsRuntimes, tabID)
	js.ActivateRuntime(runtime)
	context, _ := js.GetCurrentContext()
	if context == nil {
		js.CloseCurrentContext()
		return false
	}
	jsRuntimeTabID = tabID
	jsRuntimeActive = true
	return true
}

func captureActiveJSCContextForTabLocked(tabID uint64) {
	if !jsRuntimeActive || jsRuntimeTabID != tabID {
		return
	}
	jsRuntimes[tabID] = js.CaptureRuntime()
	jsRuntimeTabID = 0
	jsRuntimeActive = false
}

func closeJSCContextForTab(tabID uint64) {
	jsLifecycleMu.Lock()
	defer jsLifecycleMu.Unlock()
	closeJSCContextForTabLocked(tabID)
}

func closeJSCContextForTabLocked(tabID uint64) {
	if jsRuntimeActive && jsRuntimeTabID == tabID {
		js.CloseCurrentContext()
		jsRuntimeTabID = 0
		jsRuntimeActive = false
	} else {
		suspendActiveJSCContextLocked()
	}
	if runtime, ok := jsRuntimes[tabID]; ok {
		delete(jsRuntimes, tabID)
		js.CloseRuntime(runtime)
	}
}

func processJSTimers(expectedTabID uint64, expectedURL string) (dirty bool) {
	jsLifecycleMu.Lock()
	defer jsLifecycleMu.Unlock()
	if !activateJSCContextForTabLocked(expectedTabID) {
		return false
	}
	if context, currentURL := js.GetCurrentContext(); context != nil && currentURL == expectedURL {
		js.ProcessTimers()
		dirty = js.IsDirty()
	}
	captureActiveJSCContextForTabLocked(expectedTabID)
	return dirty
}

func loadPage(pageURL string, viewportWidth int) tea.Cmd {
	return loadPageContext(context.Background(), pageURL, viewportWidth)
}

func loadTabPage(ctx context.Context, pageURL string, viewportWidth int, tabID, requestID uint64) tea.Cmd {
	load := loadPageForTab(ctx, pageURL, viewportWidth, tabID)
	return func() tea.Msg {
		msg := load()
		switch result := msg.(type) {
		case PageLoadedMsg:
			result.TabID, result.RequestID = tabID, requestID
			return result
		case PageErrorMsg:
			result.TabID, result.RequestID = tabID, requestID
			return result
		default:
			return msg
		}
	}
}

func loadPageContext(ctx context.Context, pageURL string, viewportWidth int) tea.Cmd {
	return loadPageForTab(ctx, pageURL, viewportWidth, 0)
}

func loadPageForTab(ctx context.Context, pageURL string, viewportWidth int, tabID uint64) tea.Cmd {
	return func() tea.Msg {
		html, err := fetcher.FetchContext(ctx, pageURL)
		if err != nil {
			return PageErrorMsg{Err: err}
		}
		if err := ctx.Err(); err != nil {
			return PageErrorMsg{Err: err}
		}
		doc := dom.Parse(html)
		if err := ctx.Err(); err != nil {
			return PageErrorMsg{Err: err}
		}
		jsLifecycleMu.Lock()
		if err := ctx.Err(); err != nil {
			jsLifecycleMu.Unlock()
			return PageErrorMsg{Err: err}
		}
		suspendActiveJSCContextLocked()
		closeJSCContextForTabLocked(tabID)
		js.CloseCurrentContext()
		context := js.NewContext()
		js.SetCurrentContext(context, pageURL)
		js.InitDOM(doc)
		jsRuntimeTabID = tabID
		jsRuntimeActive = true
		if err := context.RegisterBridgeFunctions(); err != nil {
			js.CloseCurrentContext()
			jsRuntimeTabID = 0
			jsRuntimeActive = false
			jsLifecycleMu.Unlock()
			return PageErrorMsg{Err: fmt.Errorf("js bridge init: %w", err)}
		}
		js.InitEventSystem(context.GetContextRef())
		if _, err := context.Evaluate(js.DOMAPIScript()); err != nil {
			js.CloseCurrentContext()
			jsRuntimeTabID = 0
			jsRuntimeActive = false
			jsLifecycleMu.Unlock()
			return PageErrorMsg{Err: fmt.Errorf("js dom api error: %w", err)}
		}
		safeURL := strings.ReplaceAll(pageURL, "\\", "\\\\")
		safeURL = strings.ReplaceAll(safeURL, "'", "\\'")
		safeURL = strings.ReplaceAll(safeURL, "\n", "\\n")
		_, _ = context.Evaluate(fmt.Sprintf("document._url='%s';", safeURL))
		_, _ = context.Evaluate(`window.getComputedStyle=function(el){if(!el||el.__go_id===undefined)return null;return __go_getComputedStyle(el);};`)
		_ = js.ExecuteScriptsContext(ctx, context, doc, pageURL)
		_, _ = context.Evaluate(`if(document.dispatchEvent){var ev={type:'DOMContentLoaded',defaultPrevented:false,_propagationStopped:false,preventDefault:function(){this.defaultPrevented=true;},stopPropagation:function(){this._propagationStopped=true;}};document.dispatchEvent(ev);}`)
		finalDOM := js.GetFinalDOM()
		jsRuntimes[tabID] = js.CaptureRuntime()
		jsRuntimeTabID = 0
		jsRuntimeActive = false
		jsLifecycleMu.Unlock()
		if err := ctx.Err(); err != nil {
			jsLifecycleMu.Lock()
			closeJSCContextForTabLocked(tabID)
			jsLifecycleMu.Unlock()
			return PageErrorMsg{Err: err}
		}
		externalCSS := fetchExternalCSSContext(ctx, finalDOM, pageURL)
		fetchImagesContext(ctx, finalDOM, pageURL)
		if err := ctx.Err(); err != nil {
			jsLifecycleMu.Lock()
			closeJSCContextForTabLocked(tabID)
			jsLifecycleMu.Unlock()
			return PageErrorMsg{Err: err}
		}
		styles := css.ComputeWithExternalAtWidth(finalDOM, externalCSS, float64(viewportWidth))
		box := layout.Build(finalDOM, styles, float64(viewportWidth))
		jsLifecycleMu.Lock()
		if activateJSCContextForTabLocked(tabID) {
			current, currentURL := js.GetCurrentContext()
			if current == context && currentURL == pageURL {
				js.SetComputedStyles(styles)
				js.SetLayoutBoxes(buildLayoutBoxMap(box))
			}
			captureActiveJSCContextForTabLocked(tabID)
		}
		jsLifecycleMu.Unlock()
		fb := render.Render(box, viewportWidth)
		lines, ansiLines := ascii.Render(fb, box, viewportWidth), ascii.RenderANSI(fb, box, viewportWidth)
		if path := os.Getenv("ABYSS_DUMP_ANSI"); path != "" {
			dump := strings.Join(ansiLines, "\n")
			_ = os.WriteFile(path, []byte(dump), 0o644)
			fmt.Fprintf(os.Stderr, "abyss: dumped %d bytes to %s\n", len(dump), path)
		}
		return PageLoadedMsg{Lines: lines, ANSI: ansiLines, URL: pageURL, ExternalCSS: externalCSS}
	}
}

func reRenderCmd(viewportWidth int, zoomFactor float64, tabID, requestID uint64, pageURL string, externalCSS []string) tea.Cmd {
	return func() tea.Msg {
		jsLifecycleMu.Lock()
		defer jsLifecycleMu.Unlock()
		if !activateJSCContextForTabLocked(tabID) {
			return nil
		}
		defer captureActiveJSCContextForTabLocked(tabID)
		context, currentURL := js.GetCurrentContext()
		if context == nil || currentURL != pageURL {
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
		styles := css.ComputeWithExternalAtWidth(finalDOM, externalCSS, float64(scaledWidth))
		js.SetComputedStyles(styles)
		box := layout.Build(finalDOM, styles, float64(scaledWidth))
		js.SetLayoutBoxes(buildLayoutBoxMap(box))
		fb := render.Render(box, scaledWidth)
		lines, ansiLines := ascii.Render(fb, box, scaledWidth), ascii.RenderANSI(fb, box, scaledWidth)
		return PageRenderedMsg{TabID: tabID, RequestID: requestID, Lines: lines, ANSI: ansiLines}
	}
}

const (
	maxExternalStylesheets     = 64
	maxExternalCSSBytes        = 4 << 20
	maxExternalMediaQueryBytes = 256
	stylesheetWorkers          = 4
)

func fetchExternalCSS(doc *dom.Node, baseURL string) []string {
	return fetchExternalCSSContext(context.Background(), doc, baseURL)
}

func fetchExternalCSSContext(ctx context.Context, doc *dom.Node, baseURL string) []string {
	type stylesheetRef struct {
		url   string
		media string
	}
	links := doc.FindLinkStylesheets()
	urls := make([]stylesheetRef, 0, min(len(links), maxExternalStylesheets))
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
		media := strings.TrimSpace(link.GetAttr("media"))
		if len(media) > maxExternalMediaQueryBytes {
			continue
		}
		key := cssURL + "\x00" + media
		if _, ok := seen[key]; ok {
			continue
		}
		if len(urls) >= maxExternalStylesheets {
			break
		}
		seen[key] = struct{}{}
		urls = append(urls, stylesheetRef{url: cssURL, media: media})
	}
	type cssJob struct {
		index int
		url   string
	}
	type cssResult struct {
		index int
		body  string
	}
	workerCount := min(stylesheetWorkers, len(urls))
	jobs := make(chan cssJob)
	fetched := make(chan cssResult, workerCount)
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				body, err := fetcher.FetchCSS(ctx, job.url)
				if err != nil {
					body = ""
				}
				fetched <- cssResult{index: job.index, body: body}
			}
		}()
	}
	launched := 0
	for launched < workerCount {
		jobs <- cssJob{index: launched, url: urls[launched].url}
		launched++
	}

	pending := make(map[int]string, workerCount)
	result := make([]string, 0, len(urls))
	budgetUsed, next := 0, 0
	for next < len(urls) {
		item := <-fetched
		pending[item.index] = item.body
		for {
			body, ok := pending[next]
			if !ok {
				break
			}
			delete(pending, next)
			media := urls[next].media
			if media != "" {
				body = "@media " + media + " {\n" + body + "\n}"
			}
			if body != "" && len(body) <= maxExternalCSSBytes-budgetUsed {
				result = append(result, body)
				budgetUsed += len(body)
			}
			next++
			if launched < len(urls) {
				jobs <- cssJob{index: launched, url: urls[launched].url}
				launched++
			}
		}
	}
	close(jobs)
	wg.Wait()
	return result
}

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

func ResolveURLForTest(base, rel string) string { return ResolveURL(base, rel) }
func FetchExternalCSSForTest(doc *dom.Node, baseURL string) []string {
	return fetchExternalCSS(doc, baseURL)
}
