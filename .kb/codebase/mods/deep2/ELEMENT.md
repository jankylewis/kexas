# Element Module — Deep2 Level (For Complete Beginners)

No coding experience required. Everything explained with everyday analogies.

---

## What Is an "Element"?

Open any website. Everything you see — buttons, text fields, images, links, menus — is an **element**. An element is one building block of a web page.

In kexas, after you find an element on a page, you get a remote control for that specific element. You can:
- **Click it** — Like pressing a button with your finger.
- **Type into it** — Like using a keyboard to fill in a text field.
- **Hover over it** — Like moving your mouse over a menu to see a dropdown.
- **Read it** — Like reading the text on a label or the URL behind a link.

---

## Finding Elements — Like Finding a Book in a Library

Before you can interact with an element, you need to find it. This is like searching for a book in a library. You use a **selector** — a description of what you're looking for:

- `"#login-button"` — "Find the one called 'login-button'" (like searching by book title)
- `".menu-item"` — "Find all with the category 'menu-item'" (like searching by genre)
- `"button"` — "Find all buttons" (like asking for "all hardcover books")

> **FAQ: What if there are multiple matches?**
> `Find` returns the first match (like "give me the first Harry Potter book you find"). `FindAll` returns all matches (like "give me every Harry Potter book").

> **FAQ: What if the element hasn't appeared yet?**
> Websites load gradually. The login button might appear 1 second after the page starts loading. `WaitAndFind` keeps looking every 0.1 seconds until it finds the element (or gives up after a timeout). It's like waiting at a bus stop — you keep looking until the bus arrives.

---

## How Kexas Identifies Elements

When you find an element, Chrome gives kexas two types of "ID cards":

### nodeID — The Seat Number

Imagine a classroom. The teacher assigns seat numbers: seat 1, seat 2, seat 3... If a student moves to a different seat, their seat number changes.

A `nodeID` is like a seat number. Chrome assigns it when it inspects the page. But if the page changes (someone moves seats), the number might not match the same student anymore.

### objectID — The Student ID Card

A student ID card has a unique barcode. Even if the student changes seats, the barcode still identifies them.

An `objectID` is like a student ID. Chrome creates it when kexas asks "give me a permanent handle to this element." It works reliably as long as the student is still in the school (the page hasn't navigated away).

> **FAQ: Why does kexas need two IDs?**
> Different Chrome commands need different types of IDs. Some commands (like "what are this element's attributes?") use seat numbers (nodeID). Other commands (like "click this element") use student IDs (objectID). Kexas handles this automatically — you never need to worry about it.

> **FAQ: When do these IDs become invalid?**
> When you navigate to a new page. It's like the school year ending — all seat assignments and student IDs are reset. You need to find the element again on the new page.

---

## Clicking — What Really Happens

When you click a button on a real website, here's what happens (simplified):

1. Your finger touches the mouse button.
2. The computer detects the click.
3. Chrome figures out which element is under the mouse cursor.
4. Chrome tells that element: "You were clicked!"
5. The element reacts (e.g., a link opens a new page, a button submits a form).

When kexas clicks an element, it skips steps 1-3 and goes directly to step 4: "Hey element, you were clicked!" Chrome creates the same notifications (called **events**) that a real click would produce.

> **FAQ: What are "events"?**
> Notifications that something happened. When you click a button, Chrome creates a "click event" — a message saying "someone clicked this button." The website's code can listen for these events and react accordingly (e.g., show a popup, submit a form, navigate to another page).

> **FAQ: Does kexas's click look exactly like a real click?**
> Almost. The main difference: kexas's click doesn't have real mouse coordinates (because no actual mouse was moved). 99% of websites don't check coordinates, so this works fine. The rare exception: drawing apps or games that track where exactly you clicked.

### The Three Types of Click

| Method | What It Does | When To Use |
|--------|-------------|-------------|
| `Click()` | Clicks immediately | When you're sure the element is ready |
| `WaitAndClick()` | Waits until the element is visible, then clicks | Default choice — handles timing automatically |
| `WaitAndClickFor(30s)` | Waits up to 30 seconds, then clicks | When elements are slow to appear |

> **FAQ: Why three types? Why not just one?**
> Different situations need different approaches. Think of crossing a street:
> - `Click()` = Walk immediately (you already looked both ways)
> - `WaitAndClick()` = Wait for the walk signal, then cross (safe default)
> - `WaitAndClickFor(30s)` = Wait at a slow intersection (you know it takes a while)

---

## Typing — Like a Robot Typist

When you type "hello" on a real keyboard, each key press creates three events:

1. **Key goes down** — Your finger pushes the key. Chrome says: "The 'h' key is being pressed."
2. **Character appears** — The letter "h" appears in the text field. Chrome says: "The character 'h' was typed."
3. **Key comes back up** — Your finger releases the key. Chrome says: "The 'h' key was released."

Kexas simulates this exact sequence for each character. For "hello" (5 letters), that's 15 events (5 × 3).

> **FAQ: Why not just paste the whole word at once?**
> Many websites run code when you type. For example, a search box might show suggestions after each letter. If you paste "hello" all at once, the website might not update its suggestions because it never "saw" the individual key presses. By typing one letter at a time, kexas triggers all the website's typing-related code.

> **FAQ: Does it work with languages like Chinese, Japanese, Arabic?**
> Yes! Kexas handles all Unicode characters. Each character — whether it's English, Japanese, Arabic, or emoji — is sent as a separate key event.

### Before Typing: Focus and Clear

Before typing, kexas does two things:
1. **Focus** — "Click on the text field to select it." Like tapping on a text field on your phone before the keyboard appears.
2. **Clear** — "Delete whatever text is already in the field." Like pressing Select All + Delete before typing new text.

> **FAQ: What is "focus"?**
> The currently active element on a page. Only one element can have focus at a time. When you tap a text field, it gets focus (a cursor appears). When you tap elsewhere, it loses focus. Keyboard input goes to whichever element has focus.

---

## Hovering — Like Moving Your Mouse Over Something

When you move your mouse over a menu on a website, a dropdown might appear. That's a **hover** event.

Kexas simulates this by telling Chrome: "Pretend the mouse just moved over this element." Chrome then fires the same notifications that a real mouse hover would produce.

> **FAQ: Does hovering change visual styles (like buttons changing color)?**
> Partially. If the website uses JavaScript to change styles on hover (common for dropdowns and tooltips), it works. If the website uses pure CSS `:hover` styles, it might not work because Chrome's visual system tracks the actual mouse position, not simulated events.

---

## Scrolling — Like Scrolling on Your Phone

If an element is off the bottom of the screen (you'd need to scroll down to see it), `ScrollIntoView()` scrolls the page to bring that element into the visible area.

> **FAQ: Do I need to scroll before clicking?**
> No! Kexas's click works even on elements that are off-screen. But if you want to take a screenshot that includes the element, you should scroll first.

---

## Reading Element Data

### Getting Text

```
text := elem.GetText()  // Returns "Submit Order"
```

Like reading the label on a button. Returns whatever text is visible on the element.

### Getting Attributes

```
url := elem.GetAttribute("href")  // Returns "https://example.com/dashboard"
```

Elements have hidden properties called **attributes**. A link element `<a href="https://example.com">Click here</a>` has:
- Visible text: "Click here"
- Hidden attribute `href`: "https://example.com" (where the link goes)

> **FAQ: What's the difference between text and attributes?**
> Text is what you see on screen. Attributes are the hidden settings behind the scenes. Think of a physical button: the text is the label printed on it ("Power"), the attributes are the wiring behind it (connects to the power supply).

### Checking if an Element Is Valid

```
elem.Validate()
```

Checks if the element still exists on the page. Like asking "is this person still in the room?" Elements can disappear if the website updates its content (e.g., a loading spinner that disappears after data loads).

---

## What Can Go Wrong?

| Problem | Plain English | What To Do |
|---------|---------------|-----------|
| Element nil | You're trying to use an element that doesn't exist | Make sure `Find()` succeeded before using the element |
| Element no page | The element lost its connection to the page | The page was closed — find the element again on a new page |
| Text empty | You tried to type nothing | Check that your text variable isn't empty |
| Click failed | Chrome couldn't click the element | The element might have been removed from the page |
| Timeout | The element didn't appear in time | The website is slow — try a longer timeout, or check if the element actually exists |

---

## Summary

An Element is your remote control for one part of a web page:
- **Click** → Press a button, follow a link
- **Type** → Fill in a text field (one character at a time, like a real keyboard)
- **Hover** → Move the mouse over something (triggers dropdowns, tooltips)
- **Scroll** → Bring off-screen elements into view
- **Read** → Get the text or hidden attributes of an element
- **Validate** → Check if the element still exists

Use `WaitAnd*` methods for reliable test code — they wait for elements to be ready before acting.
