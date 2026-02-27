# 🌐 Chrome DevTools Protocol (CDP) Knowledge Base

## 🔍 **What is Chrome DevTools Protocol?**

Chrome DevTools Protocol (CDP) is the **API that allows external tools to control Chrome browser**. It's the same protocol that Chrome DevTools uses internally when you inspect elements, debug JavaScript, or analyze network requests.

### **How it Works:**
```
Your Code → CDP Commands → Chrome Browser → CDP Events → Your Code
```

---

## 🏗️ **CDP Agents Overview**

Think of CDP agents like **specialized departments** in Chrome, each handling different capabilities:

```
Chrome DevTools Protocol
├── Page Agent     (navigation, screenshots, loading)
├── DOM Agent      (finding elements, getting attributes)
├── Runtime Agent  (executing JavaScript, evaluation)
├── Network Agent  (intercepting requests, response analysis)
├── Input Agent    (mouse/keyboard simulation)
├── Security Agent (certificate handling, security)
├── Debugger Agent (breakpoints, step execution)
└── Profiler Agent (performance analysis, memory)
```

---

## 📋 **Agent Capabilities**

### **🌐 Page Agent**
- **Navigation**: `Page.navigate`, `Page.reload`
- **Loading**: `Page.loadEventFired`, `Page.lifecycleEvent`
- **Screenshots**: `Page.captureScreenshot`, `Page.printToPDF`
- **Dialogs**: `Page.handleJavaScriptDialog`
- **History**: `Page.getNavigationHistory`

### **🏗️ DOM Agent**
- **Element Finding**: `DOM.querySelector`, `DOM.querySelectorAll`
- **XPath**: `DOM.performSearch`, `DOM.getSearchResults`
- **Element Info**: `DOM.getAttributes`, `DOM.getBoxModel`
- **DOM Tree**: `DOM.getDocument`, `DOM.describeNode`
- **Modifications**: `DOM.setAttributeValue`, `DOM.removeNode`

### **⚡ Runtime Agent**
- **JavaScript**: `Runtime.evaluate`, `Runtime.callFunctionOn`
- **Objects**: `Runtime.getProperties`, `Runtime.releaseObject`
- **Execution**: `Runtime.compileScript`, `Runtime.runScript`
- **Console**: `Runtime.consoleAPICalled`, `Runtime.exceptionThrown`

### **🌐 Network Agent**
- **Request Interception**: `Network.requestWillBeSent`, `Network.responseReceived`
- **Headers**: `Network.getResponseBody`, `Network.getAllResponseBody`
- **Caching**: `Network.setCacheDisabled`, `Network.clearBrowserCache`
- **Mocking**: `Network.emulateNetworkConditions`

### **🖱️ Input Agent**
- **Mouse**: `Input.dispatchMouseEvent`, `Input.synthesizePinchGesture`
- **Keyboard**: `Input.dispatchKeyEvent`, `Input.insertText`
- **Touch**: `Input.dispatchTouchEvent`, `Input.synthesizeTapGesture`
- **Auto-enabled**: Input domain does NOT require `.enable()` — it is always available

---

## 🔧 **Agent Enablement**

### **Default Enabled Agents:**
When you connect to Chrome CDP, **only Page agent is enabled by default**.
**Input agent is auto-enabled** (no `.enable()` call needed).

### **Manual Enablement Required:**
```javascript
// Must explicitly enable other agents:
Runtime.enable()  // For JavaScript execution
DOM.enable()      // For element finding
Network.enable()  // For request interception
// Input does NOT need .enable() — it's auto-enabled
```

### **Why Manual Enablement?**
- **Performance**: Only load what you need
- **Security**: Some capabilities require explicit permission
- **Resource efficiency**: Reduces overhead

---

## 🎯 **Common CDP Commands**

### **Navigation:**
```javascript
Page.navigate({url: "https://example.com"})
Page.reload()
Page.goBack()
Page.goForward()
```

### **Element Finding:**
```javascript
DOM.enable()
DOM.querySelector({nodeId: documentId, selector: "#button"})
DOM.querySelectorAll({nodeId: documentId, selector: ".items"})
```

### **JavaScript Execution:**
```javascript
Runtime.enable()
Runtime.evaluate({expression: "document.title"})
Runtime.callFunctionOn({functionDeclaration: "function() { return window.location; }"})
```

### **User Interaction:**
```javascript
// Input domain is auto-enabled — no Input.enable() needed
Input.dispatchMouseEvent({type: "mousePressed", x: 100, y: 200})
Input.dispatchKeyEvent({type: "keyDown", text: "h", key: "h", unmodifiedText: "h"})
Input.dispatchKeyEvent({type: "keyUp",   text: "h", key: "h", unmodifiedText: "h"})
```

### **Typing with Input.dispatchKeyEvent (SPA-Compatible):**
```javascript
// For each character, send keyDown + keyUp pair:
for (const ch of "hello") {
    Input.dispatchKeyEvent({type: "keyDown", text: ch, unmodifiedText: ch, key: ch})
    Input.dispatchKeyEvent({type: "keyUp",   text: ch, unmodifiedText: ch, key: ch})
}

// Why per-character instead of Input.insertText:
// - SPA frameworks (React, Angular) listen to keydown/keyup events
// - Input.insertText bypasses those event listeners
// - Per-character keyDown+keyUp triggers the full event chain:
//   keydown → keypress → input → keyup
// - This is how go-rod and Playwright handle typing
```

### **Why NOT JavaScript this.value = text:**
```javascript
// ❌ WRONG for SPA forms:
Runtime.callFunctionOn({functionDeclaration: "function(t) { this.value = t; }", ...})
// Sets the DOM property but does NOT trigger React/Angular state updates
// Form submission will use the old/empty state

// ✅ CORRECT for SPA forms:
// Use Input.dispatchKeyEvent per character (see above)
```

---

## 🚨 **Common Pitfalls**

### **1. Forgetting to Enable Agents**
```javascript
// ❌ WRONG - Will fail
DOM.querySelector({selector: "#button"})

// ✅ CORRECT - Enable first
DOM.enable()
DOM.querySelector({selector: "#button"})
```

### **2. Wrong Command Names**
```javascript
// ❌ WRONG - Command doesn't exist
Page.waitForLoadState()

// ✅ CORRECT - Use proper commands
Page.loadEventFired  // Listen for event
```

### **3. Timing Issues**
```javascript
// ❌ WRONG - Race condition
Page.navigate({url: "..."})
DOM.querySelector({selector: "#button"})  // Page might not be loaded

// ✅ CORRECT - Wait for events
Page.navigate({url: "..."})
Page.loadEventFired  // Wait for this event
DOM.querySelector({selector: "#button"})
```

### **4. nodeId vs objectId Confusion**
```javascript
// nodeId: integer, from DOM domain, goes stale after navigation
// objectId: string, from Runtime domain, preferred for interactions

// ❌ WRONG - nodeId becomes stale after page navigation
var nodeId = DOM.querySelector({selector: "#button"}).nodeId;
Page.navigate({url: "..."});
DOM.getAttributes({nodeId: nodeId}); // FAILS: "Could not find node with given id"

// ✅ CORRECT - Use Runtime.evaluate to get objectId
var result = Runtime.evaluate({expression: 'document.querySelector("#button")', returnByValue: false});
var objectId = result.result.objectId;
Runtime.callFunctionOn({objectId: objectId, functionDeclaration: "function() { this.click(); }"});
```

### **5. JSON Number Type Assertions (Go-specific)**
```go
// CDP returns numbers as float64 in JSON, not int64 or string
// ❌ WRONG:
nodeID, ok := nodeIDs[0].(int64)   // panics
nodeID, ok := nodeIDs[0].(string)  // panics

// ✅ CORRECT:
nodeIDFloat, ok := nodeIDs[0].(float64)
nodeID := cdp.NodeID(int64(nodeIDFloat))
```

### **6. Network.enable Before setUserAgentOverride**
```javascript
// ❌ WRONG - Override silently ignored
Network.setUserAgentOverride({userAgent: "..."})

// ✅ CORRECT - Enable domain first
Network.enable()
Network.setUserAgentOverride({userAgent: "..."})
```

---

## 📚 **Where This Knowledge Comes From**

### **Primary Sources:**
1. **Chrome DevTools Protocol Documentation** - Official Google docs
2. **Chrome Source Code** - Open-source Chromium project
3. **Playwright/Puppeteer Source** - Popular automation frameworks
4. **DevTools Internals** - Chrome's own DevTools implementation

### **Practical Experience:**
- **Browser automation** development
- **Web scraping** tools
- **Testing frameworks** like Playwright, Puppeteer
- **Chrome extension** development

### **Learning Resources:**
- [Chrome DevTools Protocol Viewer](https://chromedevtools.github.io/devtools-protocol/)
- [Playwright Documentation](https://playwright.dev/)
- [Puppeteer API Reference](https://pptr.dev/)

---

## 🎯 **Framework Implementation Guidelines**

### **Best Practices:**
1. **Enable agents on demand** based on functionality needed
2. **Handle agent enablement errors** gracefully
3. **Use correct CDP commands** (don't invent new ones)
4. **Wait for proper events** instead of arbitrary timeouts
5. **Clean up properly** by disabling agents when done

### **For Kexas Framework:**
```go
// Should automatically enable based on usage:
page.Find()     // → Auto-enable DOM agent
page.Evaluate() // → Auto-enable Runtime agent
page.Click()    // → Auto-enable Input agent
```

---

*This knowledge base serves as a reference for understanding Chrome DevTools Protocol and implementing robust browser automation frameworks.* 🚀
