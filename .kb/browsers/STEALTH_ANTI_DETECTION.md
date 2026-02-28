# Stealth and Anti-Detection

> How Kexas avoids bot detection when automating browsers, including user-agent override, navigator patching, and CDP domain pre-enablement.

**Last Updated:** February 27, 2026

---

## Overview

Modern websites (Amazon, Google, etc.) use bot detection techniques to block automated browsers. Kexas implements stealth measures in `browser.go`'s `attachToPage()` method, applied before any page navigation occurs.

**Stealth is applied per-page (per-session), not per-browser.** Each new page/tab gets its own stealth setup.

---

## Detection Vectors and Countermeasures

### 1. User-Agent String

**Detection:** Headless Chrome includes `HeadlessChrome` in the user-agent string.

**Fix:** Override via `Network.setUserAgentOverride`:

```go
// CDP: Network.setUserAgentOverride
{
    "userAgent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36",
    "acceptLanguage": "en-US,en;q=0.9",
    "platform":       "macOS",
}
```

**Requirement:** `Network.enable` MUST be called before `Network.setUserAgentOverride`, otherwise the override is silently ignored.

### 2. navigator.webdriver Property

**Detection:** Automated browsers set `navigator.webdriver = true`. Bot detectors check this property.

**Fix:** Patch via `Page.addScriptToEvaluateOnNewDocument`:

```javascript
Object.defineProperty(navigator, 'webdriver', {get: () => undefined});
```

**Why `Page.addScriptToEvaluateOnNewDocument`:** This injects the script before any page scripts execute, ensuring the patch is in place before bot detection code runs.

### 3. navigator.plugins

**Detection:** Headless Chrome has an empty `navigator.plugins` array. Real browsers always have plugins.

**Fix:** Fake a realistic plugins array matching real Chrome (PDF Plugin, PDF Viewer, Native Client):

```javascript
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
```

### 4. navigator.languages

**Detection:** Headless Chrome may have empty or missing `navigator.languages`.

**Fix:**

```javascript
Object.defineProperty(navigator, 'languages', {get: () => ['en-US', 'en']});
```

### 5. window.chrome Object

**Detection:** Real Chrome browsers have a `window.chrome` object with `runtime`, `loadTimes`, and `csi` properties. Headless Chrome does not.

**Fix:**

```javascript
window.chrome = {runtime: {}, loadTimes: function() {}, csi: function() {}};
```

### 6. Permissions API

**Detection:** Automated browsers may return inconsistent `Notification.permission` via the Permissions API.

**Fix:**

```javascript
var origQuery = window.navigator.permissions.query;
window.navigator.permissions.query = function(parameters) {
    if (parameters.name === 'notifications') {
        return Promise.resolve({state: Notification.permission});
    }
    return origQuery(parameters);
};
```

### 7. iframe contentWindow Detection

**Detection:** Bot detectors use iframes to check if `window.chrome` exists in child frames.

**Fix:**

```javascript
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
```

### 8. Chrome Launch Flags (Critical Lesson)

**NEVER use `--disable-blink-features=AutomationControlled` as a Chrome flag.**

This flag causes Chrome to show a yellow "unsupported command-line flag" banner, which:
1. Visually alerts that the browser is automated
2. Is detectable by websites (banner DOM element exists)
3. Is redundant — `navigator.webdriver` is already patched via CDP JS injection

Use `--disable-infobars` instead to suppress all info bars.

---

## Implementation in browser.go

All stealth measures are applied in `attachToPage()` immediately after obtaining the `sessionId`:

```go
func (b *Browser) attachToPage(targetID string) (*Page, error) {
    // ... attach to target, get sessionId ...

    // Step 1: Enable required CDP domains FIRST
    _, _ = b.client.SendToSession(b.ctx, sessionID, "Network.enable", nil)
    _, _ = b.client.SendToSession(b.ctx, sessionID, "Page.enable", nil)

    // Step 2: Override user-agent (requires Network.enable)
    _, _ = b.client.SendToSession(b.ctx, sessionID, "Network.setUserAgentOverride", map[string]interface{}{
        "userAgent":      userAgent,
        "acceptLanguage": "en-US,en;q=0.9",
        "platform":       "macOS",
    })

    // Step 3: Patch navigator properties (requires Page.enable)
    _, _ = b.client.SendToSession(b.ctx, sessionID, "Page.addScriptToEvaluateOnNewDocument", map[string]interface{}{
        "source": `
            Object.defineProperty(navigator, 'webdriver', {get: () => undefined});
            Object.defineProperty(navigator, 'plugins', {get: () => [1, 2, 3, 4, 5]});
            Object.defineProperty(navigator, 'languages', {get: () => ['en-US', 'en']});
            window.chrome = {runtime: {}};
        `,
    })

    // Step 4: Create Page object (stealth already applied)
    var page *Page = &Page{ ... }
    return page, nil
}
```

---

## Order of Operations (Critical)

The order matters. Getting it wrong causes silent failures:

```
1. Target.attachToTarget → get sessionId
2. Network.enable        → enables Network domain (MUST be before setUserAgentOverride)
3. Page.enable           → enables Page domain (MUST be before addScriptToEvaluateOnNewDocument)
4. Network.setUserAgentOverride → override user-agent
5. Page.addScriptToEvaluateOnNewDocument → inject navigator patches
6. Page.navigate         → navigate to target URL (stealth already active)
```

**If Network.enable is not called first:** `Network.setUserAgentOverride` silently does nothing. The browser sends the default `HeadlessChrome` user-agent. Bot detection catches this.

**If Page.enable is not called first:** `Page.addScriptToEvaluateOnNewDocument` may not work reliably. The navigator patches may not be injected before page scripts execute.

---

## Relationship with AgentManager

The stealth domain enablement (`Network.enable`, `Page.enable`) in `attachToPage` is separate from the `AgentManager`'s lazy enablement. This is intentional:

- **Stealth enablement:** Done eagerly in `attachToPage`, before any navigation
- **AgentManager enablement:** Done lazily on first use of commands like `DOM.querySelector`

The AgentManager will detect that Network and Page are already enabled when it checks (via `EnsureAgent`), so there's no conflict.

---

## Debugging Bot Detection

When Amazon (or similar sites) detect automation, they redirect to:

```
https://www.amazon.com/ap/cvf/request?arb=...
```

Page title: `"Authentication required"`
Body text: `"Conditions of Use  Privacy Notice  Help"`
All inputs: hidden fields only (no visible form elements)

**Diagnostic approach:**
1. Check page title and URL in the `Find()` timeout log
2. If title is "Authentication required" → bot detection triggered
3. Verify stealth measures are applied in correct order
4. Check if `Network.enable` is called before `Network.setUserAgentOverride`

---

## Limitations

These stealth measures handle basic bot detection but do not address:

- **CAPTCHA challenges** — Amazon may still show CAPTCHAs after successful login
- **Behavioral analysis** — Mouse movement patterns, typing speed, scroll behavior
- **TLS fingerprinting** — Browser's TLS handshake characteristics
- **Canvas/WebGL fingerprinting** — Headless Chrome has detectable rendering differences
- **IP reputation** — Datacenter IPs or rapid request rates

For production use, additional measures like proxy rotation and human-like behavioral delays may be needed.

---

## File Reference

- **`browser.go`** — `attachToPage()` applies all stealth measures
- **`internal/agent/`** — `AgentManager` handles lazy domain enablement (separate from stealth)

