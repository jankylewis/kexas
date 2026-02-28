package ktest

import (
	"fmt"
	"time"
)

// ========================================
// KTEST STEPS — Playwright-Style Test Steps
// ========================================
//
// Steps provide structured logging within tests for better debugging
// and traceability. Each step records its title, duration, and status.
// Steps appear in the HTML report under each test for easy diagnosis.
//
// Usage:
//
//	ktest.Test("Checkout", func(page *kexas.Page, t ktest.KTestT) {
//	    ktest.Step(t, "Navigate to cart", func() {
//	        page.Navigate("https://example.com/cart")
//	    })
//	    ktest.Step(t, "Click checkout button", func() {
//	        page.Click("#checkout")
//	    })
//	})
//
// Steps are inspired by Playwright's test.step() API.
// ========================================

// Step executes a named step within a test, recording its duration and status.
// If the step function panics, the step is marked as failed and the panic
// is re-raised so the test framework can handle it.
//
// Usage:
//
//	ktest.Step(t, "Fill login form", func() {
//	    page.Fill("#username", "user")
//	    page.Fill("#password", "pass")
//	})
func Step(t KTestT, title string, stepFunc func()) {
	var kt *ktestT
	var ok bool
	kt, ok = t.(*ktestT)
	if !ok {
		// Fallback: just run the function with logging
		t.Logf("▸ Step: %s", title)
		var start time.Time = time.Now()
		stepFunc()
		t.Logf("  ✓ %s (%s)", title, time.Since(start).String())
		return
	}

	var start time.Time = time.Now()
	var step testStep = testStep{
		Title:  title,
		Status: "passed",
	}

	var startMsg string = fmt.Sprintf("  ▸ %s", title)
	fmt.Println(startMsg)
	kt.mu.Lock()
	kt.logs = append(kt.logs, startMsg)
	kt.mu.Unlock()

	// Run step with panic recovery
	var stepPanicked bool = false
	var panicVal interface{}
	func() {
		defer func() {
			if r := recover(); r != nil {
				stepPanicked = true
				panicVal = r
			}
		}()
		stepFunc()
	}()

	step.Duration = time.Since(start)

	if stepPanicked {
		step.Status = "failed"
		step.Error = fmt.Sprintf("%v", panicVal)
		var failMsg string = fmt.Sprintf("  ✗ %s (%s) — %s", title, step.Duration.String(), step.Error)
		fmt.Println(failMsg)
		kt.mu.Lock()
		kt.logs = append(kt.logs, failMsg)
		kt.mu.Unlock()
	} else if kt.Failed() {
		step.Status = "failed"
		if len(kt.errors) > 0 {
			step.Error = kt.errors[len(kt.errors)-1]
		}
		var failMsg string = fmt.Sprintf("  ✗ %s (%s)", title, step.Duration.String())
		fmt.Println(failMsg)
		kt.mu.Lock()
		kt.logs = append(kt.logs, failMsg)
		kt.mu.Unlock()
	} else {
		var passMsg string = fmt.Sprintf("  ✓ %s (%s)", title, step.Duration.String())
		fmt.Println(passMsg)
		kt.mu.Lock()
		kt.logs = append(kt.logs, passMsg)
		kt.mu.Unlock()
	}

	kt.mu.Lock()
	kt.steps = append(kt.steps, step)
	kt.mu.Unlock()

	// Re-raise panic so the test framework handles it
	if stepPanicked {
		panic(panicVal)
	}
}
