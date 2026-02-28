# Local/Session Storage — Plain Language Guide

---

## What Is Web Storage?

Think of web storage as a **notepad** that a website can write notes on, and those notes stay in your browser even after you close the page.

There are two types of notepads:

| Type | Analogy |
|------|---------|
| **localStorage** | A notepad that stays forever — even if you close the browser and come back tomorrow, the notes are still there. All tabs from the same website can read the same notepad. |
| **sessionStorage** | A sticky note on a single tab. If you close that tab, the note is gone. Other tabs can't see it. |

## Why Does Kexas Need Storage?

Modern websites store a LOT of information in these notepads:
- **Login tokens** — "This user is logged in as admin"
- **User preferences** — "Dark mode is ON" or "Language is Spanish"
- **Feature flags** — "Show the new checkout flow"
- **Onboarding state** — "User has completed the tutorial"

When testing, you often want to **pre-fill these notes** so the website behaves a certain way without going through the full user flow.

### Example

Without storage management:
> Open website → Click "Settings" → Click "Dark Mode" → Go back → Verify dark mode is active

With storage management:
> Write "dark_mode = true" to the notepad → Open website → It's already in dark mode ✓

## What Can You Do?

| Action | What It Means |
|--------|---------------|
| **Set** | Write a note (key-value pair) to the notepad |
| **Get** | Read a specific note |
| **Remove** | Erase one specific note |
| **Clear** | Erase ALL notes on the notepad |
| **Get All** | Read every note on the notepad |
| **Length** | Count how many notes are on the notepad |
| **Has** | Check if a specific note exists |
| **Set Many** | Write multiple notes at once (faster than one at a time) |

## localStorage vs sessionStorage: When to Use Which?

| Scenario | Which to Use |
|----------|-------------|
| Auth tokens that persist across sessions | localStorage |
| Temporary form data | sessionStorage |
| User preferences | localStorage |
| Shopping cart in a single tab | sessionStorage |
| Testing multi-tab behavior | localStorage (shared across tabs) |
