# Data Flow

> How data flows through Kexas components

**Last Updated:** February 24, 2026

---

## Browser Launch Flow

```
User Code
    │
    ├─> kexas.Launch(opts)
    │       │
    │       ├─> launcher.Launch(ctx, opts)
    │       │       │
    │       │       ├─> findChromium()
    │       │       │       │
    │       │       │       ├─> Check cache (~/.kexas/chrome/)
    │       │       │       ├─> Try download (Chrome for Testing)
    │       │       │       └─> Fall back to system browser
    │       │       │
    │       │       ├─> Create temp profile (/tmp/kexas-chrome-{timestamp}/)
    │       │       │
    │       │       ├─> buildArgs(opts)
    │       │       │       │
    │       │       │       └─> Return: [--remote-debugging-port=9222, --headless, ...]
    │       │       │
    │       │       ├─> exec.Command(chromium, args...)
    │       │       │       │
    │       │       │       └─> Start browser process
    │       │       │
    │       │       └─> extractDebuggerURL(stdout)
    │       │               │
    │       │               └─> Parse: ws://127.0.0.1:9222/devtools/browser/...
    │       │
    │       ├─> cdp.Connect(ctx, wsURL)
    │       │       │
    │       │       ├─> websocket.Dial(wsURL)
    │       │       │       │
    │       │       │       └─> Establish WebSocket connection
    │       │       │
    │       │       └─> Start readLoop() goroutine
    │       │               │
    │       │               └─> Continuously read CDP messages
    │       │
    │       └─> Return Browser{process, client}
    │
    └─> browser.NewPage()
            │
            ├─> client.Send("Target.createTarget", {url: "about:blank"})
            │       │
            │       └─> Return: {targetId: "ABC123"}
            │
            ├─> client.Send("Target.attachToTarget", {targetId, flatten: true})
            │       │
            │       └─> Return: {sessionId: "XYZ789"}
            │
            └─> Return Page{targetID, sessionID}
```

---

## Page Navigation Flow

```
User Code
    │
    ├─> page.Navigate(url, kwait.WaitUntilLoad)
    │       │
    │       ├─> page.sendCommand("Page.navigate", {url})
    │       │       │
    │       │       └─> client.SendToSession(sessionID, method, params)
    │       │               │
    │       │               ├─> Generate ID: nextID.Add(1) → 42
    │       │               │
    │       │               ├─> Create Request{
    │       │               │       id: 42,
    │       │               │       method: "Page.navigate",
    │       │               │       params: {url: "https://example.com"},
    │       │               │       sessionId: "XYZ789"
    │       │               │   }
    │       │               │
    │       │               ├─> Marshal to JSON
    │       │               │
    │       │               ├─> Create response channel
    │       │               │       pending[42] = chan *Response
    │       │               │
    │       │               ├─> websocket.Write(data)
    │       │               │       │
    │       │               │       └─> Send to browser
    │       │               │
    │       │               ├─> Wait for response
    │       │               │       │
    │       │               │       └─> <-respChan (blocks until response)
    │       │               │
    │       │               └─> Return result: {frameId: "..."}
    │       │
    │       └─> page.WaitForLoadState(WaitUntilLoad, timeout)
    │               │
    │               └─> kwait.ForPageLoad(ctx, getTitleFn, waitUntil, timeout)
    │                       │
    │                       ├─> Create ticker (100ms interval)
    │                       │
    │                       └─> Poll condition until met or timeout
    │                               │
    │                               ├─> Call getTitleFn() → page.Title()
    │                               ├─> Check: title != "" && title != "New Tab"
    │                               ├─> If true: return success
    │                               └─> If timeout: return error
    │
    └─> Success
```

---

## CDP Message Flow

### Outgoing Command

```
User Code
    │
    ├─> page.Navigate(url)
    │
    ├─> page.sendCommand("Page.navigate", params)
    │
    ├─> client.SendToSession(sessionID, method, params)
    │       │
    │       ├─> Generate ID: 42
    │       │
    │       ├─> Create Request
    │       │       {
    │       │         "id": 42,
    │       │         "method": "Page.navigate",
    │       │         "params": {"url": "https://example.com"},
    │       │         "sessionId": "XYZ789"
    │       │       }
    │       │
    │       ├─> Register pending response
    │       │       pending[42] = chan *Response
    │       │
    │       ├─> Marshal to JSON
    │       │
    │       └─> websocket.Write(jsonData)
    │
    ▼
WebSocket Connection
    │
    ▼
Browser Process (Chromium)
    │
    ├─> Execute command
    │
    └─> Send response
            {
              "id": 42,
              "result": {"frameId": "..."}
            }
```

### Incoming Response

```
Browser Process (Chromium)
    │
    └─> Send response via WebSocket
            │
            ▼
WebSocket Connection
    │
    ▼
CDP Client (readLoop goroutine)
    │
    ├─> conn.Read() → Read message
    │
    ├─> json.Unmarshal(data, &response)
    │       {
    │         "id": 42,
    │         "result": {"frameId": "..."}
    │       }
    │
    ├─> handleResponse(&response)
    │       │
    │       ├─> Find pending channel: pending[42]
    │       │
    │       └─> Send to channel: respChan <- response
    │
    └─> Loop back to read next message
            │
            ▼
SendToSession (waiting)
    │
    ├─> Receive from channel: <-respChan
    │
    └─> Return result to caller
```

### Incoming Event

```
Browser Process (Chromium)
    │
    └─> Send event via WebSocket
            {
              "method": "Page.loadEventFired",
              "params": {"timestamp": 123456}
            }
            │
            ▼
WebSocket Connection
    │
    ▼
CDP Client (readLoop goroutine)
    │
    ├─> conn.Read() → Read message
    │
    ├─> json.Unmarshal(data, &event)
    │       {
    │         "method": "Page.loadEventFired",
    │         "params": {"timestamp": 123456}
    │       }
    │
    ├─> handleEvent(&event)
    │       │
    │       ├─> Find handlers: handlers["Page.loadEventFired"]
    │       │
    │       └─> For each handler:
    │               go handler(event.Params)  // Concurrent execution
    │
    └─> Loop back to read next message
            │
            ▼
Event Handlers (goroutines)
    │
    ├─> Handler 1: Process event
    ├─> Handler 2: Process event
    └─> Handler N: Process event
```

---

## Screenshot Flow

```
User Code
    │
    ├─> page.Screenshot()
    │       │
    │       ├─> sendCommand("Page.captureScreenshot", {format: "png", quality: 100})
    │       │       │
    │       │       └─> client.SendToSession(sessionID, method, params)
    │       │               │
    │       │               ├─> Send CDP command
    │       │               │
    │       │               └─> Wait for response (may be large, up to 100MB)
    │       │
    │       ├─> Receive response: {data: "base64-encoded-png..."}
    │       │
    │       └─> Return []byte(dataStr)
    │
    └─> Write to file or process
```

**Note**: WebSocket read limit is set to 100MB to handle large screenshots.

---

## Browser Cleanup Flow

```
User Code
    │
    ├─> browser.Close()
    │       │
    │       ├─> client.Close()
    │       │       │
    │       │       ├─> Set closed flag: closed.Store(true)
    │       │       │
    │       │       ├─> Cancel context: cancel()
    │       │       │       │
    │       │       │       └─> Stops readLoop goroutine
    │       │       │
    │       │       └─> conn.Close(websocket.StatusNormalClosure)
    │       │
    │       └─> process.Close()
    │               │
    │               ├─> cmd.Process.Kill()
    │               │       │
    │               │       └─> Terminate browser process
    │               │
    │               ├─> cmd.Wait()
    │               │       │
    │               │       └─> Wait for process to exit
    │               │
    │               └─> os.RemoveAll(userDataDir)
    │                       │
    │                       └─> Delete temp profile: /tmp/kexas-chrome-{timestamp}/
    │
    └─> All resources cleaned up
```

---

## Wait Condition Flow

```
User Code
    │
    ├─> page.WaitForLoadState(kwait.WaitUntilLoad, 30*time.Second)
    │       │
    │       └─> kwait.ForPageLoad(ctx, getTitleFn, waitUntil, timeout)
    │               │
    │               ├─> Create options: {timeout: 30s, interval: 100ms}
    │               │
    │               └─> kwait.For(ctx, condition, opts)
    │                       │
    │                       ├─> Create deadline: now + 30s
    │                       │
    │                       ├─> Create ticker: 100ms
    │                       │
    │                       └─> Loop:
    │                               │
    │                               ├─> Check condition()
    │                               │       │
    │                               │       ├─> Call getTitleFn() → page.Title()
    │                               │       │       │
    │                               │       │       └─> CDP: Target.getTargetInfo
    │                               │       │
    │                               │       └─> Return: title != "" && title != "New Tab"
    │                               │
    │                               ├─> If true: return success
    │                               │
    │                               ├─> If past deadline: return timeout error
    │                               │
    │                               └─> Wait for next tick (100ms)
    │                                       │
    │                                       └─> Loop back
    │
    └─> Success or timeout
```

---

## Event Subscription Flow

```
User Code
    │
    ├─> client.On("Page.loadEventFired", handler)
    │       │
    │       ├─> Lock handlers map
    │       │
    │       ├─> Append handler to list
    │       │       handlers["Page.loadEventFired"] = append(handlers, handler)
    │       │
    │       └─> Unlock handlers map
    │
    └─> Handler registered
            │
            ▼
Browser sends event
            │
            ▼
readLoop receives event
            │
            ├─> handleEvent(&event)
            │       │
            │       ├─> Find handlers: handlers["Page.loadEventFired"]
            │       │
            │       └─> For each handler:
            │               go handler(event.Params)
            │
            ▼
Handler executes (in goroutine)
    │
    └─> Process event data
```

---

## Element Finding Flow (Updated February 27, 2026)

```
User Code
    │
    ├─> page.Find("input[type='submit']")
    │       │
    │       ├─> Detect selector type: CSS (no # or // prefix)
    │       │
    │       ├─> Start retry loop (timeout=10s, poll=200ms)
    │       │       │
    │       │       └─> findByCSS("input[type='submit']")
    │       │               │
    │       │               ├─> ensureAgentsForCommand("Runtime.evaluate")
    │       │               │       │
    │       │               │       └─> AgentManager.EnsureAgent("Runtime")
    │       │               │               │
    │       │               │               ├─> Check: enabledAgents.Runtime == true?
    │       │               │               ├─> If false: SendToSession("Runtime.enable")
    │       │               │               └─> Mark enabled, update LastUsed
    │       │               │
    │       │               ├─> sendCommand("Runtime.evaluate", {
    │       │               │       expression: "document.querySelector(\"input[type='submit']\")",
    │       │               │       returnByValue: false
    │       │               │   })
    │       │               │       │
    │       │               │       └─> CDP → Browser → Response:
    │       │               │           {result: {type: "object", objectId: "injected:1:42"}}
    │       │               │
    │       │               ├─> Check result.subtype != "null"
    │       │               ├─> Check result.type != "undefined"
    │       │               ├─> Check no exceptionDetails
    │       │               ├─> Extract objectId: "injected:1:42"
    │       │               │
    │       │               └─> NewElementWithObject(page, selector, 0, "injected:1:42", 10s)
    │       │
    │       └─> Return Element{objectId: "injected:1:42", nodeId: 0}
    │
    └─> Success (or timeout after 10s with page title in log)
```

### Element Finding by ID Flow

```
page.Find("#ap_email_login")
    │
    ├─> Detect: starts with "#" → findByID("ap_email_login")
    │       │
    │       ├─> Runtime.evaluate({expression: "document.getElementById(\"ap_email_login\")"})
    │       ├─> Extract objectId from result
    │       └─> NewElementWithObject(page, "#ap_email_login", 0, objectId, 10s)
    │
    └─> Return Element{objectId}
```

### Element Finding by XPath Flow

```
page.Find("//button[@id='submit']")
    │
    ├─> Detect: starts with "//" → findByXPath
    │       │
    │       ├─> DOM.performSearch({query: "//button[@id='submit']"})
    │       │       └─> {searchId: "search-1", resultCount: 1}
    │       │
    │       ├─> DOM.getSearchResults({searchId, fromIndex: 0, toIndex: 1})
    │       │       └─> {nodeIds: [42]}  ← float64 in JSON!
    │       │
    │       ├─> Cast: float64(42) → int64(42) → cdp.NodeID(42)
    │       └─> NewElement(page, xpath, 42, 10s)
    │
    └─> Return Element{nodeId: 42, objectId: ""}  ← will resolve lazily
```

---

## Element Click Flow (Updated February 27, 2026)

```
User Code
    │
    ├─> element.WaitAndClick()
    │       │
    │       ├─> Guard: e.nodeID > 0 || e.objectID != ""
    │       │
    │       ├─> page.WaitForElementClickable(selector, timeout)
    │       │       │
    │       │       └─> Retry loop: Find element → check IsVisible → check not disabled
    │       │
    │       ├─> resolveObjectID()
    │       │       │
    │       │       ├─> e.objectID != ""? → return cached (fast path)
    │       │       └─> DOM.resolveNode({nodeId}) → extract objectId → cache
    │       │
    │       ├─> sendCommand("Runtime.callFunctionOn", {
    │       │       functionDeclaration: "function() { this.click(); return true; }",
    │       │       objectId: "injected:1:42",
    │       │       returnByValue: true
    │       │   })
    │       │       │
    │       │       └─> CDP → Browser → executes click() → Response: {result: {value: true}}
    │       │
    │       └─> Check result.value == true
    │
    └─> Success
```

---

## Element Type Flow (Updated February 27, 2026)

```
User Code
    │
    ├─> element.Type("jankylewis.se@gmail.com")
    │       │
    │       ├─> Guard: e.nodeID > 0 || e.objectID != ""
    │       ├─> Guard: text != ""
    │       │
    │       ├─> resolveObjectID() → "injected:1:42"
    │       │
    │       ├─> Phase 1: Focus and clear
    │       │       │
    │       │       └─> Runtime.callFunctionOn({
    │       │               functionDeclaration: "function() { this.focus(); this.value = ''; return true; }",
    │       │               objectId: "injected:1:42"
    │       │           })
    │       │
    │       ├─> Phase 2: Type each character via keyboard events
    │       │       │
    │       │       ├─> For 'j':
    │       │       │       ├─> Input.dispatchKeyEvent({type: "keyDown", text: "j", key: "j"})
    │       │       │       └─> Input.dispatchKeyEvent({type: "keyUp",   text: "j", key: "j"})
    │       │       │
    │       │       ├─> For 'a':
    │       │       │       ├─> Input.dispatchKeyEvent({type: "keyDown", text: "a", key: "a"})
    │       │       │       └─> Input.dispatchKeyEvent({type: "keyUp",   text: "a", key: "a"})
    │       │       │
    │       │       └─> ... repeat for each character
    │       │
    │       └─> Return nil (success)
    │
    └─> Success (SPA framework receives each keystroke as real keyboard input)
```

---

## Stealth Setup Flow (Updated February 27, 2026)

```
Browser.NewPage() / Browser.FirstPage()
    │
    ├─> Target.createTarget / Target.getTargets
    │
    ├─> Target.attachToTarget({targetId, flatten: true})
    │       └─> {sessionId: "XYZ789"}
    │
    ├─> Stealth Setup (before any navigation)
    │       │
    │       ├─> Network.enable  ← MUST be first
    │       ├─> Page.enable     ← MUST be before addScript
    │       │
    │       ├─> Network.setUserAgentOverride({
    │       │       userAgent: "Mozilla/5.0 ... Chrome/146.0.0.0 Safari/537.36",
    │       │       acceptLanguage: "en-US,en;q=0.9",
    │       │       platform: "macOS"
    │       │   })
    │       │
    │       └─> Page.addScriptToEvaluateOnNewDocument({
    │               source: "Object.defineProperty(navigator, 'webdriver', ...); ..."
    │           })
    │
    └─> Return Page{sessionId, agentManager}
```

---

## Summary

Key flow patterns:
1. **Commands**: User → Page → ensureAgentsForCommand → CDP → WebSocket → Browser → Response
2. **Events**: Browser → WebSocket → CDP → Handlers (goroutines)
3. **Finding**: Find → Retry loop → Runtime.evaluate → objectId → Element
4. **Clicking**: Element → resolveObjectID → Runtime.callFunctionOn(this.click())
5. **Typing**: Element → resolveObjectID → Focus via JS → Input.dispatchKeyEvent per char
6. **Stealth**: attachToPage → Network.enable → setUserAgentOverride → addScriptToEvaluateOnNewDocument
7. **Cleanup**: Close CDP → Kill process → 2s port wait → Remove temp files
8. **Waiting**: Poll condition → Check → Sleep → Repeat until success/timeout

All flows use:
- Context for cancellation
- Channels for synchronization
- Goroutines for concurrency
- Atomic operations for state
- objectId-first pattern for element interaction

