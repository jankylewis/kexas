package kcore

import (
	"fmt"
	"time"

	"github.com/jankylewis/kexas/internal/agent"
	"github.com/jankylewis/kexas/internal/logger"
)


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
//
// Retries up to 5 times with 100/200/400/800/1600 ms backoff because the
// initial blank-page target appears asynchronously after Chrome launch — when
// many test packages spawn Chrome simultaneously (parallel `go test ./...`),
// Target.getTargets sometimes races ahead of the page target's registration
// and returns only the browser target. A single retry pass closes the race.
func (b *Browser) FirstPage() (*Page, error) {
	const maxAttempts int = 5
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		var page *Page
		var err error
		page, err = b.firstPageAttempt()
		if err == nil {
			return page, nil
		}
		lastErr = err
		if attempt < maxAttempts-1 {
			var backoff time.Duration = time.Duration(1<<uint(attempt)) * 100 * time.Millisecond
			b.log.Debug("FirstPage no page targets yet, retrying", "attempt", attempt+1, "backoff", backoff.String())
			time.Sleep(backoff)
		}
	}
	return nil, fmt.Errorf("kexas: no page targets found after %d attempts: %w", maxAttempts, lastErr)
}

// firstPageAttempt is a single non-retrying iteration of FirstPage.
// Returns (page, nil) on success; (nil, error) on transient or hard failures.
func (b *Browser) firstPageAttempt() (*Page, error) {
	b.log.Debug("getting first page")

	var params map[string]interface{} = map[string]interface{}{}
	var result map[string]interface{}
	var err error
	result, err = b.client.Send(b.ctx, "Target.getTargets", params)
	if err != nil {
		return nil, fmt.Errorf("kexas: failed to get targets: %w", err)
	}

	var targetInfos []interface{}
	var ok bool
	targetInfos, ok = result["targetInfos"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("kexas: invalid targetInfos in response")
	}

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

	return nil, fmt.Errorf("no page targets in this poll")
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
