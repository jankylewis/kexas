# Cookie Management — Plain Language Guide

---

## What Are Cookies?

When you visit a website and log in, the website gives your browser a small "ticket" called a **cookie**. Every time you visit that website again, your browser shows the ticket, and the website says "Oh, I remember you — you're logged in."

Without cookies, you'd have to log in every single time you click a link.

## Why Does Kexas Need to Manage Cookies?

Imagine you're testing a banking website. Every test starts by:
1. Opening the login page
2. Typing in the username
3. Typing in the password
4. Clicking "Log In"
5. Waiting for the dashboard to load

That takes 3-5 seconds. If you have 100 tests, that's 5-8 minutes just logging in.

**With cookie management**, you can skip all of that. You give the browser the "logged in" ticket directly, and jump straight to testing the dashboard. Your 100 tests now save 5-8 minutes.

## What Can You Do?

| Action | What It Means |
|--------|---------------|
| **Set a cookie** | Give the browser a ticket (e.g., "this user is logged in") |
| **Get cookies** | See all the tickets the browser currently has |
| **Delete a cookie** | Remove a specific ticket |
| **Clear all cookies** | Remove ALL tickets (like logging out of everything) |
| **Check if cookie exists** | Ask "does the browser have this ticket?" |
| **Get cookie value** | Ask "what does this specific ticket say?" |

## Security Features

| Feature | What It Means in Plain Language |
|---------|-------------------------------|
| **HttpOnly** | "Only the server can read this ticket — no one else on the page can peek at it" |
| **Secure** | "Only send this ticket over encrypted connections" |
| **SameSite** | "Don't show this ticket to other websites" |

These protect users from having their login tickets stolen by malicious code.
