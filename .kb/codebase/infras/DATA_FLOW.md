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

## Summary

Key flow patterns:
1. **Commands**: User → Page → CDP → WebSocket → Browser → Response
2. **Events**: Browser → WebSocket → CDP → Handlers (goroutines)
3. **Cleanup**: Close CDP → Kill process → Remove temp files
4. **Waiting**: Poll condition → Check → Sleep → Repeat until success/timeout

All flows use:
- Context for cancellation
- Channels for synchronization
- Goroutines for concurrency
- Atomic operations for state

