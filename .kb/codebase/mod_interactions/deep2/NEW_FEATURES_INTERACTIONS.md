# New Feature Module Interactions — Plain Language Guide

---

## How the New Features Work Together

Think of kexas like a team of specialists. Each specialist handles one job, and they communicate through a shared walkie-talkie (the WebSocket connection to Chrome).

### The Team

| Specialist | Job |
|-----------|-----|
| **Cookie Manager** | Handles the browser's "remember me" tickets |
| **Storage Manager** | Handles the browser's notepad (localStorage/sessionStorage) |
| **Tab Manager** | Keeps track of which tabs are open and helps switch between them |
| **Video Recorder** | Takes pictures of the screen and stitches them into a video |
| **KAPI** | The team leader who coordinates everyone and gives simple instructions |

### How They Work Together

**Cookies + Tabs:** All tabs share the same cookies. If you log in on Tab 1, Tab 2 is also logged in. This is just how browsers work in real life too.

**Storage + Tabs:** localStorage is shared across tabs (like a shared whiteboard). sessionStorage is per-tab (like a personal notebook). If Tab 1 writes to localStorage, Tab 2 can read it. But Tab 1's sessionStorage is invisible to Tab 2.

**Recorder + Tabs:** The recorder captures what one tab looks like. If you switch tabs, the recording follows the active tab.

**KAPI + Everything:** KAPI is the "easy mode" that lets you use all the other specialists with simple one-line commands instead of detailed step-by-step instructions.

### A Real-World Scenario

Imagine testing an e-commerce website:

1. **KAPI** opens the website
2. **Cookie Manager** sets the login cookie (skip the login page)
3. **Storage Manager** sets "preferred_currency=EUR" in localStorage
4. User clicks "View Product" which opens a **new tab** (Tab Manager tracks it)
5. **Video Recorder** captures everything for the bug report
6. The test checks the product price shows in EUR
7. **Tab Manager** closes the product tab, returns to the main page
8. **KAPI** checks for errors — all good!

### The Communication Chain

When you tell KAPI to set a cookie:

```
You → KAPI → Cookie Manager → Page → Chrome
```

Each person passes the message to the next, adds a little context, and passes back the result. If something goes wrong, each person adds their own note about what happened, so the error message tells the full story.
