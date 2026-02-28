# Browser Module — Deep2 Level (For Complete Beginners)

Imagine you've never written code before. This document explains the Browser module using everyday analogies.

---

## What Is This Project About?

Kexas is a tool that controls a web browser (Google Chrome) automatically — like a robot that can open Chrome, go to websites, click buttons, and fill in forms, all by itself. This is used for **testing websites**: instead of a human manually clicking through every page to check if it works, kexas does it automatically in seconds.

> **FAQ: Why would you want a robot to use a website?**
> Imagine you built an online store. Every time you change the code, you need to check: Can users still log in? Can they add items to a cart? Can they check out? Doing this manually takes hours. Kexas does it in seconds, every time you make a change.

---

## The Browser Module — What It Does

The Browser module is like a **remote control for Chrome**. When your program says "open Chrome," the Browser module:

1. **Starts Chrome** — Just like double-clicking the Chrome icon, but done by your program.
2. **Connects to Chrome** — Establishes a communication line so your program can send commands.
3. **Opens tabs** — Like clicking the "+" button to open a new tab.
4. **Shuts down** — Closes Chrome completely when done.

---

## Key Concepts — Explained With Analogies

### What Is a "Process"?

Think of your computer as a building with many offices. Each running program gets its own office (its own **process**). Your program (kexas) is in Office A. Chrome is in Office B. They can't directly reach into each other's offices — they communicate by sending messages through the building's mail system.

> **FAQ: Why can't they just share the same office?**
> Isolation. If Chrome crashes (the office catches fire), your program survives in its own office. If they shared an office, Chrome's crash would take your program down too.

> **FAQ: What is a "PID"?**
> Process ID — a unique number the operating system assigns to each running program. Like an office number. "PID 12345" means "the program in office #12345."

### What Is a "WebSocket"?

Imagine a phone call between two people. In a normal phone call (like HTTP), one person asks a question, the other answers, and then they hang up. If you have another question, you have to call again.

A WebSocket is like keeping the phone line open. Both people can talk whenever they want, as long as they want, without hanging up and redialing. This is how your program talks to Chrome — one continuous phone call.

> **FAQ: Why not just use regular web requests (HTTP)?**
> HTTP is like sending letters — you send a request, wait for a reply, and the connection ends. For controlling a browser, you need to send hundreds of commands per second and receive responses instantly. A permanent phone line (WebSocket) is much faster than sending hundreds of letters.

> **FAQ: What is "TCP"?**
> TCP (Transmission Control Protocol) is the delivery system that makes sure messages arrive correctly. Think of it as certified mail — you get confirmation that the letter was delivered, and if it gets lost, it's automatically resent. WebSocket runs on top of TCP.

> **FAQ: What are "SYN" and "ACK"?**
> Before the phone call starts, the two sides do a quick check:
> - **SYN** (synchronize): "Hey, I want to talk to you." (Your program → Chrome)
> - **SYN-ACK**: "Sure, I'm ready to talk too." (Chrome → Your program)
> - **ACK** (acknowledge): "Great, let's start." (Your program → Chrome)
> This is called the "3-way handshake" — like saying hello before a conversation.

### What Is "Full Duplex"?

Full duplex means both sides can talk at the same time. Like a phone call where both people can speak simultaneously. The opposite would be a walkie-talkie where only one person can talk at a time (half-duplex), or a TV broadcast where only the station sends and you just watch (simplex).

### What Is a "Temp Directory"?

A temporary folder on your computer where Chrome stores its stuff (saved passwords, browsing history, cookies) during the test. When the test is done, this folder is deleted.

> **FAQ: Why not use Chrome's normal data folder?**
> Because tests need a clean start. If Chrome remembered your login from a previous test, the "test login works" test might pass even if login is actually broken. A temp directory gives Chrome amnesia — every test starts fresh.

> **FAQ: What if the temp folder isn't deleted?**
> If your program crashes unexpectedly, the folder stays behind like trash after a party. Kexas has a cleanup function that deletes these leftover folders the next time it runs.

---

## How It Works — Step by Step

Think of it as ordering a taxi through an app:

### 1. "Call a taxi" → `Launch()`

You open the taxi app and request a ride. Behind the scenes:
- The app finds an available driver (finds Chrome on your computer).
- The driver starts driving to you (Chrome process starts).
- The driver shares their phone number (Chrome shares its WebSocket URL).
- You call the driver (your program connects via WebSocket).

Now you're connected and can give directions.

### 2. "Open a map" → `NewPage()` or `FirstPage()`

You tell the driver: "Open Google Maps on the car's screen." This creates a new **tab** in Chrome — like opening a new page in the browser.

- `NewPage()` = "Open a brand new, empty tab."
- `FirstPage()` = "Use the tab that's already open." (Chrome always starts with one blank tab.)

> **FAQ: What's a "tab"?**
> A tab is a single web page view in Chrome. You can have multiple tabs open at once — each shows a different website. In kexas, each tab is represented by a `Page` object.

### 3. "Give directions" → Sending Commands

Now you can tell the driver (Chrome): "Go to google.com" (navigate), "Click the search button" (click), "Type 'hello'" (type). Each command is sent as a message over the WebSocket "phone line."

### 4. "End the ride" → `Close()`

When done, you close the app:
- The phone call ends (WebSocket disconnects).
- The driver goes home (Chrome process is killed).
- The ride receipt is deleted (temp folder is removed).

---

## What If Things Go Wrong?

| Problem | What Happens | Analogy |
|---------|-------------|---------|
| Chrome not installed | `Launch()` fails with "Chrome not found" | Taxi app can't find any drivers |
| Chrome crashes | All Page operations fail | Driver's car broke down |
| Close() never called | Chrome keeps running, wasting memory | Taxi meter keeps running |
| Network port busy | Launch finds a different port | Parking spot taken, finds another |

---

## Multiple Browsers — For Parallel Testing

Imagine you need to test 6 different features of a website. Instead of one taxi driver testing them one by one (slow), you hire 6 drivers simultaneously. Each driver:
- Gets their own car (Chrome process).
- Gets their own map (Page).
- Tests one feature independently.
- Reports back when done.

This is **parallel testing**. It's faster because all 6 tests run at the same time.

> **FAQ: Don't the 6 Chrome instances use a lot of memory?**
> Yes — about 150-200 MB each, so ~1 GB for 6. Most modern computers have 8-16 GB of RAM, so this is manageable. On servers with limited memory, you can reduce the number of parallel workers.

---

## Summary

The Browser module is the starting point of everything in kexas. It:
1. Starts Chrome (a separate program on your computer).
2. Connects to it (via WebSocket — a persistent phone line).
3. Opens tabs (each tab is a Page you can control).
4. Cleans up when done (kills Chrome, deletes temp files).

Everything else — navigating to websites, clicking buttons, typing text, checking results — builds on top of this foundation.
