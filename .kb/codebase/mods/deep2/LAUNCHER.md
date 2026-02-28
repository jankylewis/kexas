# Launcher Module — Deep2 Level (For Complete Beginners)

No coding experience required. Everything explained with everyday analogies.

---

## What Is the Launcher?

The Launcher is the part of kexas that **starts Google Chrome** on your computer. But not the Chrome you normally use — it starts a fresh, temporary, isolated copy of Chrome specifically for testing.

Think of it like renting a hotel room instead of using your house. The hotel room is clean, has no personal items, and is thrown away (cleaned) after you leave. Your house (normal Chrome) keeps all your bookmarks, passwords, and history — you don't want tests messing with that.

---

## What Does the Launcher Do?

When you ask kexas to start testing, the Launcher:

1. **Finds Chrome** on your computer (like finding a car in a parking lot).
2. **Creates a temporary folder** for Chrome's data (like checking into a hotel room).
3. **Picks an available port** for communication (like choosing a radio channel).
4. **Starts Chrome** as a separate program (like starting the car engine).
5. **Gets Chrome's address** so kexas can talk to it (like getting the car's phone number).

When testing is done:

6. **Kills Chrome** (turns off the engine).
7. **Deletes the temporary folder** (checks out of the hotel).

---

## Key Concepts

### What Is a "Port"?

Imagine your computer is an apartment building. The building has one street address (your IP address, like `127.0.0.1`). But inside the building, there are 65,535 apartments (ports). Each running program can "live" in one or more apartments.

When Chrome starts, it moves into apartment #9222 (or whatever port kexas assigns). Kexas knows to send messages to apartment #9222 to reach Chrome.

> **FAQ: What is `127.0.0.1`?**
> It's a special address that means "this computer." Like saying "my house." Messages sent to `127.0.0.1` never leave your computer — they loop back to yourself. Also called "localhost."

> **FAQ: What if the port is already taken?**
> The Launcher asks the operating system: "Give me any available apartment." The OS finds an empty one (e.g., #52431). No conflicts.

### What Is a "Temp Directory"?

A folder on your computer that's meant to be short-lived. Like a disposable coffee cup — use it once, throw it away.

The Launcher creates a folder like `/tmp/kexas-chrome-1709012345-abc/` where Chrome stores its cookies, cache, and browsing data. When testing ends, this folder is deleted.

> **FAQ: Why a new folder each time?**
> **Clean state.** If tests shared Chrome's data, Test 1's login cookies might affect Test 2. A fresh folder means Chrome starts with zero history, zero cookies, zero cache — every time.

> **FAQ: What if the folder isn't deleted?**
> If your program crashes, the folder stays behind (like leaving trash after a picnic). Kexas cleans up leftover folders the next time it runs — but only once, before any tests start.

### What Is "Headless Mode"?

Running Chrome without showing a window on your screen. Chrome does everything normally (loads pages, runs JavaScript, renders layouts) but invisibly.

> **FAQ: Why invisible?**
> 1. **Speed** — Drawing pixels on screen takes time. Headless mode skips the display, making Chrome faster.
> 2. **Servers** — Test servers in the cloud don't have monitors. Headless mode lets Chrome run on these monitor-less machines.
> 3. **Less distraction** — You don't want Chrome windows popping up every time tests run.

> **FAQ: Can I see what Chrome is doing?**
> Yes! Set headless to `false` in the config, and Chrome opens a visible window. This is helpful for debugging — you can watch the robot click buttons in real time.

### What Are "Chrome Flags"?

When Chrome starts, you can give it special instructions called "flags" or "arguments." Like telling a car: "Start in eco mode, turn off the radio, use parking mode."

| Flag | What It Means in Plain English |
|------|-------------------------------|
| `--headless=new` | "Don't show a window" |
| `--no-first-run` | "Don't show the 'Welcome to Chrome' screen" |
| `--disable-extensions` | "Don't load any browser extensions" |
| `--remote-debugging-port=9222` | "Listen for remote control commands on port 9222" |
| `--user-data-dir=/tmp/...` | "Use this temporary folder for your data" |
| `--window-size=1280,720` | "Pretend the window is 1280 pixels wide and 720 pixels tall" |

> **FAQ: Why disable extensions?**
> Extensions (like ad blockers) can change how websites look and behave. A test expecting an ad to be visible would fail if an ad blocker is installed. Disabling extensions ensures predictable, consistent behavior.

---

## Starting Chrome — The "fork + exec" Process

When kexas starts Chrome, here's what happens at the computer level:

1. **Fork** — The operating system creates a copy of your program. Now there are two identical programs running. Think of it as cloning yourself.

2. **Exec** — The clone replaces itself with Chrome's code. The clone is no longer a copy of your program — it's now Chrome. Think of the clone putting on a Chrome costume.

Now you have two separate programs running: your original program (kexas) and Chrome.

> **FAQ: Why this two-step process?**
> It's how ALL programs are started on Mac and Linux. There's no "just start a program" command — you always clone yourself first, then the clone becomes the new program. It sounds weird, but it's been this way since the 1970s.

> **FAQ: What is a "PID"?**
> Process ID — a unique number the operating system gives to each running program. Like an employee badge number. "PID 12345" means "the program with badge #12345." Kexas remembers Chrome's PID so it can later tell the OS: "Kill program #12345."

---

## Getting Chrome's Address

After Chrome starts, it prints a message: "I'm ready for remote control at ws://127.0.0.1:9222/devtools/browser/abc-123"

Kexas reads this message from Chrome's output and uses this address to connect. It's like Chrome leaving a sticky note on the fridge: "Call me at this number."

> **FAQ: What is `ws://`?**
> It stands for "WebSocket" — the communication protocol. Just like `https://` means "secure web page," `ws://` means "WebSocket connection." It's the phone line between kexas and Chrome.

> **FAQ: What is `abc-123` at the end?**
> A unique identifier for this specific Chrome instance. If you started 6 Chrome instances, each would have a different code. Like serial numbers on products.

---

## The "sync.Once" Story — A Bug Fix

This is a story about a real bug that was fixed:

**The Problem**: When running 6 tests in parallel, kexas starts 6 Chrome instances simultaneously. Each one tries to clean up leftover temp folders from previous runs. But Worker 1's Chrome has ALREADY created its temp folder, and Worker 2's cleanup DELETES it! Chrome 1 crashes because its folder disappeared.

**The Fix**: Clean up leftover folders ONCE, BEFORE any Chrome instances start. A mechanism called `sync.Once` ensures the cleanup code runs exactly one time, no matter how many workers try to run it.

> **FAQ: What is `sync.Once`?**
> A promise that something will happen exactly once. Like a "break glass in case of emergency" alarm — the first person to break the glass triggers the alarm. Everyone else who tries to break the glass finds it's already broken and the alarm is already ringing.

---

## Shutting Down — `Close()`

When testing is done:

1. **SIGKILL Chrome** — The operating system forcibly terminates Chrome. Like pulling the power cord.

> **FAQ: What is SIGKILL?**
> A command from the operating system to a program: "Stop. Now. Non-negotiable." The program cannot refuse, negotiate, or do any cleanup. It's the most forceful way to stop a program. There's a gentler option called SIGTERM ("please stop when convenient"), but Chrome sometimes ignores it, so kexas uses SIGKILL to guarantee Chrome actually stops.

2. **Delete temp folder** — Removes all of Chrome's temporary data.

3. **Release the port** — When Chrome dies, the operating system reclaims the port for other programs to use.

---

## Parallel Safety — How 6 Chromes Coexist

When running 6 tests simultaneously:

| Concern | How It's Solved |
|---------|----------------|
| 6 Chromes sharing data | Each gets its own temp folder (hotel rooms) |
| 6 Chromes on the same port | Each gets a unique port (radio channels) |
| Cleanup deleting active folders | Cleanup runs once before any Chrome starts |
| One Chrome crashing others | Each is a separate process — they can't affect each other |

---

## Summary

The Launcher is the "Chrome factory":
1. Finds Chrome on your computer.
2. Creates an isolated temporary folder (clean state).
3. Picks a unique communication port.
4. Starts Chrome as a separate program (invisible by default).
5. Reads Chrome's address for remote control.
6. When done: kills Chrome, deletes the folder, frees the port.

It handles all the messy operating system details so the rest of kexas can focus on testing.
