# Page Module — Deep2 Level (For Complete Beginners)

No coding experience required. Everything explained with everyday analogies.

---

## What Is a "Page"?

When you open Chrome and type "google.com" in the address bar, you see Google's homepage. That's a **page** — one website view in one browser tab.

In kexas, a `Page` is a remote control for one Chrome tab. You can tell it:
- "Go to this website" (navigate)
- "Find the login button" (find element)
- "Take a photo of what's on screen" (screenshot)
- "Run this script" (evaluate JavaScript)

---

## How Does a Page Connect to Chrome?

Remember from the Browser doc: your program and Chrome are like two people on a phone call (WebSocket). A Page is like a specific **topic** in that conversation.

Imagine Chrome has 3 tabs open. Your program says: "I want to talk about Tab 2." Chrome says: "OK, here's a conversation ID for Tab 2: `session-xyz`." Now whenever your program sends a command with `session-xyz`, Chrome knows it's about Tab 2.

> **FAQ: What is a "session ID"?**
> A unique code that identifies which tab you're talking about. Like a ticket number at a deli counter — you say "order #42" and they know which sandwich is yours.

> **FAQ: What is a "target ID"?**
> Chrome's internal name for a tab. Think of it as the tab's permanent room number. The session ID is your visitor badge for that room — you might get a new badge, but the room number stays the same.

---

## What Can You Do With a Page?

### Navigate — "Go to this website"

```
page.Navigate("https://amazon.com")
```

This tells Chrome: "Load amazon.com in this tab." Chrome then:
1. **Looks up the address** — Translates "amazon.com" into a number-based address (like translating a street name into GPS coordinates). This is called **DNS lookup**.

> **FAQ: What is DNS?**
> Domain Name System — the internet's phone book. You know "amazon.com" (the name), but computers need "52.94.236.248" (the number). DNS translates between them.

2. **Connects to the server** — Establishes a connection to Amazon's computer somewhere in the world.

3. **Downloads the page** — Gets the HTML (the page's blueprint), CSS (the visual design), JavaScript (the interactive behavior), and images.

4. **Builds the page** — Chrome reads the HTML and constructs a tree of elements (the DOM). It applies styles from CSS. It runs JavaScript.

5. **Signals "done"** — Chrome sets `document.readyState` to `"complete"`, and kexas stops waiting.

> **FAQ: What is "readyState"?**
> A status indicator, like a traffic light:
> - 🔴 `"loading"` — Still downloading and reading the HTML.
> - 🟡 `"interactive"` — HTML is done, but images/scripts are still loading.
> - 🟢 `"complete"` — Everything is done. Safe to interact.

> **FAQ: How does kexas know when the page is done loading?**
> It asks Chrome every 100 milliseconds: "What's the readyState?" When Chrome says `"complete"`, kexas proceeds. This is like repeatedly asking "Are we there yet?" every few seconds on a road trip.

### Find — "Find this button/input/link"

```
page.Find("#login-button")
```

This says: "Find the HTML element with ID 'login-button'." Chrome searches through the page's structure and returns a reference to that element.

> **FAQ: What is "#login-button"?**
> A **CSS selector** — a search pattern for finding HTML elements. The `#` means "find by ID." Common patterns:
> - `"#login"` → Find the element with `id="login"`
> - `".menu"` → Find elements with `class="menu"`
> - `"button"` → Find all `<button>` elements
> - `"div > p"` → Find `<p>` elements directly inside a `<div>`

> **FAQ: What if the element doesn't exist yet?**
> Use `WaitAndFind` instead of `Find`. It keeps checking every 100ms until the element appears (or gives up after a timeout). Websites often load elements gradually — the button might appear 500ms after the page starts loading.

### Screenshot — "Take a photo"

```
page.Screenshot()
```

Chrome captures exactly what's visible in the browser window as an image (PNG format). This is useful for:
- Debugging failed tests ("what did the page look like when the test failed?")
- Visual comparisons ("does the page look the same as yesterday?")

> **FAQ: What is PNG?**
> A picture format (like JPEG), but without losing quality. Text in screenshots stays sharp. The tradeoff: PNG files are larger than JPEG.

### Evaluate — "Run a script"

```
page.Evaluate("document.title")
```

This runs a small piece of JavaScript code in the page and returns the result. `document.title` returns the text in the browser's tab bar (e.g., "Amazon.com: Online Shopping").

> **FAQ: What is JavaScript?**
> The programming language that makes websites interactive. When you click a "Show More" button and new content appears without reloading the page — that's JavaScript. Chrome has a built-in JavaScript engine (called V8) that runs this code.

### SetContent — "Replace the whole page"

```
page.SetContent("<html><body><h1>Hello</h1></body></html>")
```

Instead of loading a website from the internet, this directly sets the page's content. Like writing on a whiteboard instead of projecting a PowerPoint. Useful for testing without needing a real website.

---

## How Commands Travel

When you call `page.Navigate("google.com")`, here's the journey:

```
Your Program          WebSocket (phone line)         Chrome
    │                        │                         │
    │── "Go to google.com" ─►│── delivers message ───►│
    │                        │                         │── loads google.com
    │                        │                         │── builds the page
    │◄── "Done loading" ─────│◄── sends response ─────│
    │                        │                         │
```

Every command follows this pattern: your program sends a message, Chrome does the work, Chrome sends back the result. The WebSocket stays open the whole time — like a phone call that never hangs up.

> **FAQ: How fast is this?**
> Each message takes about 0.1–5 milliseconds to go back and forth. That's 0.001–0.005 seconds. You wouldn't notice it, but for a computer doing thousands of operations, it adds up.

---

## The "Agent Manager" — Chrome's Feature Switches

Chrome has many features (called "domains"): DOM inspection, JavaScript execution, screenshot capture, network monitoring, etc. But they're all turned OFF by default — you have to explicitly turn them ON before using them.

The Page has an invisible helper called the **Agent Manager** that handles this automatically. When you call `page.Find(...)`, the Agent Manager notices it needs the DOM feature, turns it on silently, and then proceeds with the find.

> **FAQ: Why are features off by default?**
> Performance. Each enabled feature makes Chrome do extra work (tracking changes, sending notifications). If you're only taking screenshots, you don't want Chrome wasting energy tracking every DOM change. The Agent Manager ensures only the features you actually use are turned on.

Think of it like a hotel room: the lights in each room are off until someone walks in. The Agent Manager is the automatic light sensor.

---

## What Happens When Things Go Wrong?

| Problem | What kexas tells you | What it means |
|---------|---------------------|---------------|
| "page load timeout after 30s" | The website took too long to load | Slow internet, or the site is down |
| "element not found: #login" | The element doesn't exist | Wrong selector, or element hasn't loaded yet |
| "failed to navigate" | Chrome couldn't go to the URL | Invalid URL, or no internet connection |

> **FAQ: What should I do when a test fails?**
> 1. Read the error message — it usually tells you exactly what went wrong.
> 2. Check if the website is actually working (open it in a real browser).
> 3. Use `WaitAndFind` instead of `Find` if elements load slowly.
> 4. Take a screenshot before the failing step to see what the page looked like.

---

## Summary

A Page is your remote control for one browser tab. You can:
- **Navigate** → Go to websites
- **Find** → Locate buttons, inputs, links
- **Screenshot** → Capture what's visible
- **Evaluate** → Run JavaScript
- **SetContent** → Inject HTML directly

Commands travel over a WebSocket (persistent phone line) to Chrome and back. The Agent Manager silently handles Chrome's feature switches so you don't have to.
