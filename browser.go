// Package kexas provides high-performance browser automation for Chromium.
package kexas

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/kexas-project/kexas/errors"
	"github.com/kexas-project/kexas/internal/agent"
	"github.com/kexas-project/kexas/internal/cdp"
	"github.com/kexas-project/kexas/internal/logger"
	"github.com/kexas-project/kexas/launcher"
)

// Browser represents a browser instance with CDP connection.
type Browser struct {
	process *launcher.Browser
	client  *cdp.Client
	log     *logger.Logger
	ctx     context.Context
	pages   []*Page
	pagesMu sync.Mutex
}

// Launch starts a new browser instance and connects via CDP.
func Launch(opts *LaunchOptions) (*Browser, error) {
	var ctx context.Context = context.Background()
	return LaunchWithContext(ctx, opts)
}

// LaunchWithContext starts a new browser with a custom context.
func LaunchWithContext(ctx context.Context, opts *LaunchOptions) (*Browser, error) {
	var log *logger.Logger = logger.New("browser")

	// Launch browser process
	log.Debug("launching browser process")
	var process *launcher.Browser
	var err error
	process, err = launcher.Launch(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("kexas: failed to launch browser: %w", err)
	}

	// Connect to browser via CDP
	var wsURL string = process.WebSocketURL()
	log.Debug("connecting to CDP", "url", wsURL)

	var client *cdp.Client
	client, err = cdp.Connect(ctx, wsURL)
	if err != nil {
		process.Close()
		return nil, fmt.Errorf("kexas: failed to connect to CDP: %w", err)
	}

	log.Info("browser launched and connected")

	var browser *Browser = &Browser{
		process: process,
		client:  client,
		log:     log,
		ctx:     ctx,
	}

	return browser, nil
}

// NewPage creates a new browser page (tab).
func (b *Browser) NewPage() (*Page, error) {
	b.log.Debug("creating new page")

	// Create new target (tab)
	var params map[string]interface{} = map[string]interface{}{
		"url": "about:blank",
	}

	var result map[string]interface{}
	var err error
	result, err = b.client.Send(b.ctx, "Target.createTarget", params)
	if err != nil {
		b.log.Error("failed to create target", "err", err)
		return nil, fmt.Errorf("kexas: failed to create page: %w", err)
	}

	var targetID string
	var ok bool
	targetID, ok = result["targetId"].(string)
	if !ok {
		return nil, fmt.Errorf("kexas: invalid targetId in response")
	}

	b.log.Debug("page created", "targetId", targetID)

	var page *Page
	var attachErr error
	page, attachErr = b.attachToPage(targetID)
	if attachErr != nil {
		return nil, attachErr
	}

	b.pagesMu.Lock()
	b.pages = append(b.pages, page)
	b.pagesMu.Unlock()

	return page, nil
}

// FirstPage gets the first existing page (the default tab).
func (b *Browser) FirstPage() (*Page, error) {
	b.log.Debug("getting first page")

	// Get all targets
	var params map[string]interface{} = map[string]interface{}{}
	var result map[string]interface{}
	var err error
	result, err = b.client.Send(b.ctx, "Target.getTargets", params)
	if err != nil {
		b.log.Error("failed to get targets", "err", err)
		return nil, fmt.Errorf("kexas: failed to get targets: %w", err)
	}

	var targetInfos []interface{}
	var ok bool
	targetInfos, ok = result["targetInfos"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("kexas: invalid targetInfos in response")
	}

	// Find the first page target
	for _, targetInfoRaw := range targetInfos {
		var targetInfo map[string]interface{}
		targetInfo, ok = targetInfoRaw.(map[string]interface{})
		if !ok {
			continue
		}

		var targetType string
		targetType, ok = targetInfo["type"].(string)
		if !ok || targetType != "page" {
			continue
		}

		var targetID string
		targetID, ok = targetInfo["targetId"].(string)
		if !ok {
			continue
		}

		b.log.Debug("found first page", "targetId", targetID)

		var page *Page
		var attachErr error
		page, attachErr = b.attachToPage(targetID)
		if attachErr != nil {
			return nil, attachErr
		}

		b.pagesMu.Lock()
		b.pages = append(b.pages, page)
		b.pagesMu.Unlock()

		return page, nil
	}

	return nil, fmt.Errorf("kexas: no page targets found")
}

// attachToPage attaches to a page target and returns a Page object.
func (b *Browser) attachToPage(targetID string) (*Page, error) {
	// Attach to the target to get a session
	var params map[string]interface{} = map[string]interface{}{
		"targetId": targetID,
		"flatten":  true,
	}

	var result map[string]interface{}
	var err error
	result, err = b.client.Send(b.ctx, "Target.attachToTarget", params)
	if err != nil {
		b.log.Error("failed to attach to target", "err", err)
		return nil, fmt.Errorf("kexas: failed to attach to page: %w", err)
	}

	var sessionID string
	var ok bool
	sessionID, ok = result["sessionId"].(string)
	if !ok {
		return nil, fmt.Errorf("kexas: invalid sessionId in response")
	}

	b.log.Info("attached to page", "sessionId", sessionID, "targetId", targetID)

	// Stealth: enable required domains first
	_, _ = b.client.SendToSession(b.ctx, sessionID, "Network.enable", nil)
	_, _ = b.client.SendToSession(b.ctx, sessionID, "Page.enable", nil)

	// Stealth: override user-agent to remove "HeadlessChrome" (go-rod pattern)
	var userAgent string = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36"
	_, _ = b.client.SendToSession(b.ctx, sessionID, "Network.setUserAgentOverride", map[string]interface{}{
		"userAgent":      userAgent,
		"acceptLanguage": "en-US,en;q=0.9",
		"platform":       "macOS",
	})

	// Stealth: comprehensive automation fingerprint removal
	// Based on puppeteer-extra-plugin-stealth techniques
	_, _ = b.client.SendToSession(b.ctx, sessionID, "Page.addScriptToEvaluateOnNewDocument", map[string]interface{}{
		"source": `
			// 1. Remove navigator.webdriver
			Object.defineProperty(navigator, 'webdriver', {get: () => undefined});

			// 2. Fake plugins array (Chrome normally has PDF plugins)
			Object.defineProperty(navigator, 'plugins', {
				get: () => {
					var arr = [
						{name: 'Chrome PDF Plugin', filename: 'internal-pdf-viewer', description: 'Portable Document Format'},
						{name: 'Chrome PDF Viewer', filename: 'mhjfbmdgcfjbbpaeojofohoefgiehjai', description: ''},
						{name: 'Native Client', filename: 'internal-nacl-plugin', description: ''},
					];
					arr.item = function(i) { return this[i]; };
					arr.namedItem = function(name) { return this.find(function(p) { return p.name === name; }); };
					arr.refresh = function() {};
					return arr;
				}
			});

			// 3. Fake languages
			Object.defineProperty(navigator, 'languages', {get: () => ['en-US', 'en']});

			// 4. Fake chrome runtime (must exist and look real)
			window.chrome = {runtime: {}, loadTimes: function() {}, csi: function() {}};

			// 5. Fake permissions API to not reveal automation
			var origQuery = window.navigator.permissions.query;
			window.navigator.permissions.query = function(parameters) {
				if (parameters.name === 'notifications') {
					return Promise.resolve({state: Notification.permission});
				}
				return origQuery(parameters);
			};

			// 6. Prevent iframe detection of automation
			Object.defineProperty(HTMLIFrameElement.prototype, 'contentWindow', {
				get: function() {
					return new Proxy(window, {
						get: function(target, prop) {
							if (prop === 'chrome') return window.chrome;
							return Reflect.get(target, prop);
						}
					});
				}
			});

			// 7. Fix Notification.permission for headless
			if (typeof Notification !== 'undefined' && Notification.permission === 'denied') {
				Object.defineProperty(Notification, 'permission', {get: () => 'default'});
			}
		`,
	})

	var page *Page = &Page{
		browser:      b,
		targetID:     targetID,
		sessionID:    sessionID,
		log:          logger.New("page"),
		ctx:          b.ctx,
		agentManager: agent.NewAgentManager(b.client, sessionID),
	}

	return page, nil
}

// Pages returns all currently tracked pages.
func (b *Browser) Pages() []*Page {
	b.pagesMu.Lock()
	defer b.pagesMu.Unlock()

	var result []*Page = make([]*Page, len(b.pages))
	copy(result, b.pages)
	return result
}

// PageCount returns the number of currently tracked pages.
func (b *Browser) PageCount() int {
	b.pagesMu.Lock()
	defer b.pagesMu.Unlock()
	return len(b.pages)
}

// PageByIndex returns a page by its index in the tracked pages list.
func (b *Browser) PageByIndex(index int) (*Page, error) {
	b.pagesMu.Lock()
	defer b.pagesMu.Unlock()

	if index < 0 || index >= len(b.pages) {
		return nil, fmt.Errorf("page index %d: %w", index, errors.ErrPageIndexOutOfRange)
	}

	return b.pages[index], nil
}

// PageByURL finds a page whose URL contains the given pattern (substring match).
func (b *Browser) PageByURL(urlPattern string) (*Page, error) {
	b.pagesMu.Lock()
	var pageCopy []*Page = make([]*Page, len(b.pages))
	copy(pageCopy, b.pages)
	b.pagesMu.Unlock()

	for i := 0; i < len(pageCopy); i++ {
		var pageURL string
		var err error
		pageURL, err = pageCopy[i].URL()
		if err != nil {
			continue
		}
		if strings.Contains(pageURL, urlPattern) {
			return pageCopy[i], nil
		}
	}

	return nil, fmt.Errorf("url pattern '%s': %w", urlPattern, errors.ErrNoPageMatchingURL)
}

// WaitForNewPage executes the given action and waits for a new page (tab) to appear.
// The action typically triggers a new tab (e.g., clicking a link with target="_blank").
// Returns the new page, or an error if no new page appears within the timeout.
func (b *Browser) WaitForNewPage(action func(), timeout time.Duration) (*Page, error) {
	b.log.Debug("waiting for new page", "timeout", timeout)

	// Record current page count
	b.pagesMu.Lock()
	var beforeCount int = len(b.pages)
	b.pagesMu.Unlock()

	// Execute the action that should open a new tab
	action()

	// Poll for new targets
	var start time.Time = time.Now()
	var pollInterval time.Duration = 100 * time.Millisecond

	for time.Since(start) < timeout {
		time.Sleep(pollInterval)

		// Check for new targets via CDP
		var result map[string]interface{}
		var err error
		result, err = b.client.Send(b.ctx, cdp.CmdTargetGetTargets, map[string]interface{}{})
		if err != nil {
			continue
		}

		var targetInfos []interface{}
		var ok bool
		targetInfos, ok = result["targetInfos"].([]interface{})
		if !ok {
			continue
		}

		// Count page targets
		var pageTargets []string
		for _, targetInfoRaw := range targetInfos {
			var targetInfo map[string]interface{}
			targetInfo, ok = targetInfoRaw.(map[string]interface{})
			if !ok {
				continue
			}
			var targetType string
			targetType, ok = targetInfo["type"].(string)
			if !ok || targetType != "page" {
				continue
			}
			var tid string
			tid, _ = targetInfo["targetId"].(string)
			pageTargets = append(pageTargets, tid)
		}

		if len(pageTargets) > beforeCount {
			// Find the new target that we haven't attached to
			b.pagesMu.Lock()
			var knownTargets map[string]bool = make(map[string]bool)
			for _, p := range b.pages {
				knownTargets[p.targetID] = true
			}
			b.pagesMu.Unlock()

			for _, tid := range pageTargets {
				if !knownTargets[tid] {
					var page *Page
					page, err = b.attachToPage(tid)
					if err != nil {
						return nil, fmt.Errorf("failed to attach to new page: %w", err)
					}

					b.pagesMu.Lock()
					b.pages = append(b.pages, page)
					b.pagesMu.Unlock()

					b.log.Info("new page detected", "targetId", tid)
					return page, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("after %v: %w", timeout, errors.ErrWaitForNewPageTimeout)
}

// CloseAllPagesExcept closes all tracked pages except the specified one.
func (b *Browser) CloseAllPagesExcept(keep *Page) error {
	b.pagesMu.Lock()
	var pagesToClose []*Page = make([]*Page, 0)
	for _, p := range b.pages {
		if p != keep {
			pagesToClose = append(pagesToClose, p)
		}
	}
	b.pagesMu.Unlock()

	for _, p := range pagesToClose {
		var err error = p.Close()
		if err != nil {
			b.log.Error("failed to close page", "targetId", p.targetID, "error", err)
		}
	}

	// Update the pages list to only keep the one page
	b.pagesMu.Lock()
	if keep != nil {
		b.pages = []*Page{keep}
	} else {
		b.pages = nil
	}
	b.pagesMu.Unlock()

	return nil
}

// removePage removes a page from the tracked pages list.
func (b *Browser) removePage(page *Page) {
	b.pagesMu.Lock()
	defer b.pagesMu.Unlock()

	for i := 0; i < len(b.pages); i++ {
		if b.pages[i] == page {
			b.pages = append(b.pages[:i], b.pages[i+1:]...)
			return
		}
	}
}

// Close closes the browser and cleans up resources.
func (b *Browser) Close() error {
	b.log.Debug("closing browser")

	var err error = b.client.Close()
	if err != nil {
		b.log.Error("failed to close CDP client", "err", err)
	}

	err = b.process.Close()
	if err != nil {
		b.log.Error("failed to close browser process", "err", err)
		return err
	}

	b.log.Info("browser closed")
	return nil
}
