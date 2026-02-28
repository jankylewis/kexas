# KTest Module — Deep2 Level (For Complete Beginners)

No coding experience required. Everything explained with everyday analogies.

---

## What Is KTest?

Imagine you own a restaurant. Every morning, you need to check:
- Does the door lock work?
- Does the cash register open?
- Does the oven heat up?
- Does the fridge keep food cold?

You could check each thing manually, every single morning. Or you could hire a helper who runs through the checklist automatically and reports: "Everything works!" or "The oven is broken!"

**KTest is that helper, but for websites.** It automatically runs tests on your website and tells you what's working and what's broken.

---

## How Tests Are Organized

### Tests — The Individual Checks

Each test checks one specific thing:

```
Test: "Login should work with valid password"
  1. Go to the login page
  2. Type username
  3. Type password
  4. Click the login button
  5. Check: Am I on the dashboard page? → YES = PASS, NO = FAIL
```

### Groups — Like Folders for Tests

Tests can be organized into groups, like files in folders:

```
📁 Authentication
  ├── 📁 Login
  │   ├── ✅ should work with valid password
  │   └── ❌ should fail with wrong password
  ├── 📁 Logout
  │   └── ✅ should redirect to home page
📁 Shopping Cart
  ├── ✅ should add items
  └── ✅ should calculate total correctly
```

> **FAQ: Why organize tests into groups?**
> The same reason you organize files into folders — to find things easily. When you have 500 tests, "Authentication > Login > should work with valid password" is much clearer than "test_247."

### Hooks — Setup and Cleanup

**Before each test**: "Open a fresh browser tab and go to the login page."
**After each test**: "Close the tab and clear cookies."

These are called **hooks** — code that runs automatically at specific points:

| Hook | When It Runs | Analogy |
|------|-------------|---------|
| BeforeAll | Once before ALL tests start | "Open the restaurant" |
| BeforeEach | Before EACH individual test | "Clean the table for the next customer" |
| AfterEach | After EACH individual test | "Clear the dishes" |
| AfterAll | Once after ALL tests finish | "Close the restaurant" |

> **FAQ: Why do we need hooks?**
> To avoid repeating the same setup code in every test. Instead of writing "go to the login page" at the start of every test, you write it once in BeforeEach, and it runs automatically.

---

## How Tests Are Registered — AlphaInit

In kexas, you declare your tests at the top of your file:

```go
var _ = kexas.AlphaInit(
    ktest.Test("Login works", func(page, t) {
        // test code here
    }),
)
```

This is like filling out a registration form. When the program starts, it collects all registrations and knows which tests to run.

> **FAQ: What does "AlphaInit" mean?**
> "Alpha" means "first" (like Alpha, Beta, Gamma in the Greek alphabet). "Init" means "initialize." AlphaInit runs at the very beginning of the program, before anything else. It's the first thing that happens.

> **FAQ: Why the weird `var _ = ...` syntax?**
> Think of it as a hack. Go (the programming language) doesn't have a built-in way to say "run this code at startup." The `var _ = ...` trick forces Go to execute the code when the program starts, even though the result is thrown away (the `_`).

---

## Running Tests — Sequential vs Parallel

### Sequential — One at a Time

```
Test 1: Login .............. 3 seconds
Test 2: Logout ............. 2 seconds
Test 3: Add to cart ........ 4 seconds
Test 4: Checkout ........... 5 seconds
Test 5: Search ............. 3 seconds
Test 6: Profile ............ 2 seconds
─────────────────────────────────────
Total:                       19 seconds
```

Like one cashier serving six customers in a line.

### Parallel — Multiple at Once

```
Worker 1: Test 1 (3s) → Test 4 (5s)          = 8 seconds
Worker 2: Test 2 (2s) → Test 5 (3s) → idle   = 5 seconds
Worker 3: Test 3 (4s) → Test 6 (2s)          = 6 seconds
─────────────────────────────────────────────────────────
Total (wall clock):                             8 seconds  ← 2.4x faster!
```

Like three cashiers serving six customers simultaneously.

> **FAQ: What is a "worker"?**
> A helper that runs tests. In kexas, each worker is a **goroutine** — a lightweight thread of execution. Think of it as an employee. 6 workers = 6 employees working simultaneously.

> **FAQ: What is a "goroutine"?**
> A way for a Go program to do multiple things at once. It's like having multiple hands — you can cook, stir, and chop simultaneously. Each goroutine runs independently but shares the same computer resources.

### Process Isolation — Each Worker Gets Its Own Chrome

This is critical. Each worker opens its **own separate Chrome browser**. Why?

Imagine testing a website with two workers:
- Worker 1 is testing login.
- Worker 2 is testing logout.

If they shared the same browser, Worker 1 logs in, and then Worker 2 logs out — now Worker 1's test fails because it was kicked out! By giving each worker their own browser, they can't interfere with each other.

> **FAQ: Doesn't running 6 Chrome browsers use a lot of memory?**
> Yes — about 1 GB total. But modern computers have 8-16 GB. The time savings (running in 8 seconds instead of 19) are worth the memory cost.

> **FAQ: What is a "channel" (how workers get their tasks)?**
> Imagine a to-do list posted on a bulletin board. Workers walk up, grab the next task, and go do it. A channel is like that bulletin board — tasks go in one end, and workers pull tasks from the other end. The first idle worker gets the next task.

---

## What Happens When a Test Fails?

### Panic Recovery — Catching Crashes

Sometimes a test crashes unexpectedly (called a **panic** in Go). Like a customer tripping and falling in the restaurant.

Without protection, one crashing test would bring down ALL workers — like one falling customer causing the entire restaurant to evacuate.

KTest has "crash guards" that catch panics: "This test crashed, mark it as FAILED, and move on to the next test." Other workers continue unaffected.

> **FAQ: What causes a "panic"?**
> Common causes:
> - Trying to interact with something that doesn't exist (like clicking a button that was removed)
> - A programming mistake in the test code
> - Chrome crashing unexpectedly

### Screenshots on Failure

When a test fails, kexas can automatically take a screenshot of the browser at the moment of failure. This is like a security camera — you can look at the photo later to understand what went wrong.

> **FAQ: Where are screenshots saved?**
> In the `screenshotDir` specified in your config file (default: `./screenshots/`). Each screenshot is named after the failed test.

---

## Assertions — Checking If Things Are Correct

After performing actions (navigate, click, type), you need to CHECK if the result is correct. This is called an **assertion**.

```
"Assert that the page title equals 'Dashboard'"
```

If the title IS "Dashboard" → ✅ PASS
If the title is something else → ❌ FAIL

KTest uses **kassert** — a library for writing readable assertions:

```go
kassert.That(t, title).Equals("Dashboard")
kassert.That(t, items).HasLength(5)
kassert.That(t, price).IsGreaterThan(0)
kassert.That(t, error).IsNil()
```

These read almost like English:
- "Assert that [title] equals 'Dashboard'"
- "Assert that [items] has length 5"
- "Assert that [price] is greater than 0"
- "Assert that [error] is nil" (nil = nothing/no error)

> **FAQ: What's the difference between "Error" and "Fatal"?**
> - **Error**: "This check failed, but keep running other checks." Like a teacher marking one answer wrong but continuing to grade the rest.
> - **Fatal**: "This check failed, and there's no point continuing." Like a teacher seeing the student wrote their name wrong — the whole exam is invalid.

---

## HTML Reports

After all tests finish, ktest generates a beautiful HTML report — a web page you can open in any browser:

```
📊 Test Report - February 27, 2026
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ 45 passed  ❌ 2 failed  ⏱️ 12.3 seconds

🟢 Auth > Login > valid credentials ........ 1.2s  Worker 1
🟢 Auth > Login > remember me .............. 0.8s  Worker 2
🔴 Auth > Login > expired token ............ 3.1s  Worker 1  📸
🟢 Shopping > Add to cart .................. 2.1s  Worker 3
🔴 Shopping > Checkout > payment ........... 4.5s  Worker 2  📸
...
```

📸 = screenshot available for failed tests.

The report is saved in two files:
- `report.html` — Always overwritten (latest results).
- `report-20260227-143022.html` — Timestamped copy (historical record).

> **FAQ: Why save two copies?**
> So you can compare today's results with yesterday's. If Monday had 0 failures and Tuesday has 2, you know something changed. The timestamped copy is your historical record.

---

## The Complete Flow — From Start to Finish

```
1. 📝 REGISTRATION: Tests are registered (AlphaInit)
       "Here are the 47 tests I want to run"

2. ⚙️ CONFIGURATION: Settings are loaded
       "Run 6 workers, headless mode, take screenshots on failure"

3. 🔍 FILTERING: Optional test filtering
       "Only run tests with 'Login' in the name"

4. 📊 SORTING: Tests ordered by priority
       "Run high-priority tests first"

5. 🏁 BEFORE ALL: Global setup
       "Seed the test database"

6. 🏃 EXECUTION: Workers run tests
       Worker 1: [Test A] [Test D] [Test G]
       Worker 2: [Test B] [Test E] [Test H]
       Worker 3: [Test C] [Test F]

7. 🏁 AFTER ALL: Global cleanup
       "Clear the test database"

8. 📋 REPORT: HTML report generated
       "45 passed, 2 failed, total time 12.3s"

9. 🚪 EXIT: Program ends
       Exit code 0 (all passed) or 1 (some failed)
```

> **FAQ: What is an "exit code"?**
> A number the program returns when it finishes. 0 means "everything is fine." 1 (or any non-zero number) means "something went wrong." Automated build systems (like GitHub Actions) check this number to decide if the build passed or failed.

---

## Summary

KTest is kexas's built-in test runner:
- **Tests** check if website features work correctly.
- **Groups** organize tests into categories.
- **Hooks** run setup/cleanup code automatically.
- **Parallel execution** runs tests simultaneously for speed.
- **Crash guards** prevent one failing test from crashing everything.
- **kassert** provides readable assertion checks.
- **HTML reports** show results beautifully.
