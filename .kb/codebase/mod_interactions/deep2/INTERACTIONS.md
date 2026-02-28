# Element Interactions — Deep2 Level (For Complete Beginners)

No coding experience required. Everything explained with everyday analogies.

---

## What Are "Interactions"?

When you use a website, you interact with it:
- **Click** a "Buy Now" button.
- **Type** your email into a login form.
- **Hover** over a menu to see dropdown options.
- **Scroll** down to see more content.

Kexas simulates these interactions automatically. Instead of a human moving a mouse and pressing keys, kexas tells Chrome: "Click this button" or "Type this text into that field."

---

## Clicking — Like Pressing a Button

### What Happens When You Click (In Real Life)

1. You move your mouse over a button.
2. You press the mouse button down.
3. You release the mouse button.
4. The website reacts (opens a new page, submits a form, etc.).

### What Kexas Does

Kexas skips steps 1-3 and goes directly to step 4. It tells Chrome: "Hey, pretend this button was just clicked." Chrome then does everything it would normally do — open a link, submit a form, toggle a checkbox.

> **FAQ: Is this the same as a real click?**
> Almost identical. The website's code receives the same "click" notification it would get from a real mouse click. The only difference: kexas's click doesn't include real mouse coordinates (because no real mouse was moved). This matters for <1% of websites.

> **FAQ: What if the button hasn't appeared yet?**
> Websites load gradually. The button might appear 1 second after the page starts loading. `WaitAndClick` keeps checking: "Is the button there? No... No... Yes! Click!" This is like waiting for an elevator — you keep pressing the button until the door opens.

### The Three Types of Click

Think of crossing a busy street:

| Method | Analogy | When To Use |
|--------|---------|-------------|
| `Click()` | Cross immediately (you already looked both ways) | You're 100% sure the button is ready |
| `WaitAndClick()` | Wait for the walk signal, then cross | Default — handles timing for you |
| `WaitAndClickFor(30s)` | Wait at a very slow intersection | Button takes a long time to appear |

---

## Typing — Like a Very Fast Typist

### What Happens When You Type (In Real Life)

When you press the "A" key on your keyboard:
1. Your finger pushes the key **down** → Computer says: "A key is being pressed"
2. The letter "a" **appears** in the text field → Computer says: "The character 'a' was typed"
3. Your finger **releases** the key → Computer says: "The A key was released"

Three separate events for one letter!

### What Kexas Does

Kexas simulates all three events for each character. For typing "hello":
- "h" → down, appear, release
- "e" → down, appear, release
- "l" → down, appear, release
- "l" → down, appear, release
- "o" → down, appear, release

That's 15 events for 5 characters.

> **FAQ: Why not just paste the whole word at once?**
> Many websites react to EACH key press. A search box might show suggestions after each letter:
> - You type "h" → suggestions for "h" appear
> - You type "he" → suggestions update for "he"
> - You type "hel" → suggestions update for "hel"
>
> If you pasted "hello" all at once, the website would only see the final result and might skip the intermediate suggestions. By typing one letter at a time, kexas triggers all the website's typing-related behavior — just like a real person typing.

> **FAQ: Does it work with non-English characters?**
> Yes! Japanese, Chinese, Arabic, emojis — all work. Each character is sent as a separate event, regardless of the language.

### Before Typing: Focus and Clear

Before typing, kexas does two things:

1. **Focus** — "Click on the text field to select it." Like tapping on a text field on your phone — the keyboard appears and you can start typing. Without focus, the typed characters go nowhere.

> **FAQ: What is "focus"?**
> The text field that's currently "selected" and ready to receive keyboard input. Only one field can have focus at a time. When you click a text field, it gets a blinking cursor — that's focus. If you click somewhere else, it loses focus.

2. **Clear** — "Delete whatever text is already there." Like pressing Select All + Delete before typing something new. This ensures you start with a clean field.

---

## Hovering — Like Moving Your Mouse Over Something

When you move your mouse over a menu item on a website, a dropdown might appear. This is called **hovering**.

Kexas tells Chrome: "Pretend the mouse just moved over this element." Chrome then runs whatever code the website has for "mouse is over this" — showing dropdowns, changing colors, displaying tooltips.

> **FAQ: Does the element visually change color (like buttons that light up on hover)?**
> It depends. If the color change is done by JavaScript code (common for menus and tooltips), yes. If it's done by pure CSS styling, it might not. For testing purposes, the important thing is that dropdown menus and tooltips appear correctly — and they do.

---

## Scrolling — Like Scrolling on Your Phone

If a web page is longer than your screen, some elements are "below the fold" — you'd need to scroll down to see them.

`ScrollIntoView()` tells Chrome: "Scroll the page so this element is visible and centered on screen." It's like turning the pages of a book to get to a specific paragraph.

> **FAQ: Do I need to scroll before clicking?**
> No! Kexas's click works on elements that are off-screen. But if you want to take a screenshot that SHOWS the element, scroll first.

---

## Reading Data — Like Reading a Label

### Getting Text

```
text := elem.GetText()
```

Returns the text you'd see on the element. For a button labeled "Submit Order", it returns `"Submit Order"`.

### Getting Hidden Information (Attributes)

HTML elements have hidden settings. A link like `<a href="https://google.com">Click here</a>` has:
- **Visible text**: "Click here" (what you see)
- **Hidden attribute `href`**: "https://google.com" (where the link goes)

```
url := elem.GetAttribute("href")  // Returns "https://google.com"
```

> **FAQ: What's the difference between "text" and "attributes"?**
> Think of a physical button in an elevator:
> - **Text** = The label printed on it ("Floor 5")
> - **Attributes** = The wiring behind it (connects to the 5th floor motor)
>
> You can see the text; attributes are hidden "under the hood."

### Checking If an Element Still Exists

```
elem.Validate()
```

Websites constantly update their content. A "Loading..." spinner might disappear after data loads. `Validate` checks: "Is this element still on the page?"

> **FAQ: Why would an element disappear?**
> - The page navigated to a new URL (everything changes).
> - JavaScript removed it (like hiding a popup after the user clicks "close").
> - A framework (React, Angular) re-rendered the page and replaced the old elements with new ones.

---

## The Waiting Game — Why Timing Matters

Websites don't appear instantly. When you navigate to amazon.com:
- 0.0s: Page starts loading
- 0.2s: Basic HTML structure appears
- 0.5s: Main content loads
- 1.0s: Images appear
- 1.5s: Interactive elements become clickable

If kexas tries to click a button at 0.3s, the button might not exist yet! The `WaitAnd*` methods solve this by checking repeatedly:

```
"Is the button there?" → No → wait 0.1s
"Is the button there?" → No → wait 0.1s
"Is the button there?" → Yes! → Click!
```

> **FAQ: How often does it check?**
> Every 100 milliseconds (0.1 seconds). This is a balance: checking too often wastes computing power, checking too rarely makes tests slow.

> **FAQ: What if the button never appears?**
> After a timeout (default: 10-30 seconds), kexas gives up and reports an error: "Element not found after 30 seconds." This prevents tests from hanging forever.

---

## What Can Go Wrong?

| Problem | What It Means | What To Do |
|---------|--------------|-----------|
| "Element nil" | You're trying to use an element that doesn't exist | Make sure `Find()` worked before using the element |
| "Text empty" | You tried to type nothing | Check your test data |
| "Click failed" | Chrome couldn't click the element | The element might have been removed |
| "Timeout" | Element didn't appear in time | Use a longer timeout, or check the website |

---

## Summary

Kexas simulates human interactions with websites:

| Action | What It Simulates | Key Detail |
|--------|-------------------|-----------|
| **Click** | Pressing a button/link | Fires the same events as a real click |
| **Type** | Typing on a keyboard | Types one character at a time (3 events per character) |
| **Hover** | Moving mouse over something | Triggers dropdowns and tooltips |
| **Scroll** | Scrolling the page | Centers the element on screen |
| **Read** | Looking at text/attributes | Gets visible text or hidden settings |

Always use `WaitAnd*` methods (like `WaitAndClick`) in test code — they handle timing automatically. Only use immediate methods (like `Click`) when you're absolutely sure the element is ready.
