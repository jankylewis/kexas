# Multi-Tab Management — Plain Language Guide

---

## What Are Tabs?

When you browse the web, you can open multiple pages side by side using **tabs** at the top of your browser. Each tab shows a different webpage, but they all share the same browser window.

## Why Does Kexas Need Multi-Tab Support?

Many real websites open new tabs:
- **"Log in with Google"** — Opens a Google login popup in a new tab
- **"Open in new tab"** — Right-clicking a link or `target="_blank"` links
- **Payment pages** — Redirecting to Stripe or PayPal in a new tab
- **PDF downloads** — Some PDFs open in a new tab

If a testing tool can only see one tab, it's blind to what happens in the popup. The test would fail or hang.

## What Can You Do?

| Action | What It Means |
|--------|---------------|
| **List all tabs** | See every tab that's currently open |
| **Count tabs** | "How many tabs are open right now?" |
| **Get tab by number** | "Give me the 2nd tab" |
| **Find tab by URL** | "Give me the tab that's showing google.com" |
| **Wait for new tab** | "Click this link, then wait for a new tab to appear" |
| **Close other tabs** | "Close everything except this one tab" |
| **Bring tab to front** | "Switch to this tab" (like clicking on it) |
| **Check if tab is closed** | "Is this tab still open?" |

## Real-World Example: OAuth Login

1. You're on `myapp.com/login`
2. You click "Sign in with Google"
3. A **new tab** opens showing `accounts.google.com`
4. You type your email and password in the Google tab
5. Google redirects back, and the Google tab closes
6. You're now logged in on the original `myapp.com` tab

Without multi-tab support, step 3-5 would be impossible to automate.

## How Tabs and Cookies Interact

All tabs in the same browser **share cookies**. If you log in on tab 1, tab 2 is also logged in. This matches how a real browser works.

However, **sessionStorage is per-tab**. Each tab has its own private notepad that other tabs can't see. localStorage is shared across all tabs.
