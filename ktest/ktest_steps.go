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
		runStepFallback(t, title, stepFunc)
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

	var stepPanicked bool
	var panicVal interface{}
	stepPanicked, panicVal = runWithRecovery(stepFunc)
	step.Duration = time.Since(start)

	classifyAndLogStep(kt, title, &step, stepPanicked, panicVal)

	kt.mu.Lock()
	kt.steps = append(kt.steps, step)
	kt.mu.Unlock()

	// Re-raise panic so the test framework handles it
	if stepPanicked {
		panic(panicVal)
	}
}

// runStepFallback runs a step against a non-ktestT KTestT (the path used by raw
// *testing.T or other adapters). No structured step record is produced.
func runStepFallback(t KTestT, title string, stepFunc func()) {
	t.Logf("▸ Step: %s", title)
	var start time.Time = time.Now()
	stepFunc()
	t.Logf("  ✓ %s (%s)", title, time.Since(start).String())
}

// runWithRecovery executes fn and reports whether it panicked, returning the
// recovered value if so. The caller decides whether to re-raise.
func runWithRecovery(fn func()) (panicked bool, value interface{}) {
	defer func() {
		if r := recover(); r != nil {
			panicked = true
			value = r
		}
	}()
	fn()
	return
}

// classifyAndLogStep updates step.Status / step.Error based on whether the step
// panicked, was failed by the test, or passed cleanly — and emits the matching
// log line to both stdout and the ktestT log buffer.
func classifyAndLogStep(kt *ktestT, title string, step *testStep, stepPanicked bool, panicVal interface{}) {
	var msg string
	if stepPanicked {
		step.Status = "failed"
		step.Error = fmt.Sprintf("%v", panicVal)
		msg = fmt.Sprintf("  ✗ %s (%s) — %s", title, step.Duration.String(), step.Error)
	} else if kt.Failed() {
		step.Status = "failed"
		if len(kt.errors) > 0 {
			step.Error = kt.errors[len(kt.errors)-1]
		}
		msg = fmt.Sprintf("  ✗ %s (%s)", title, step.Duration.String())
	} else {
		msg = fmt.Sprintf("  ✓ %s (%s)", title, step.Duration.String())
	}
	fmt.Println(msg)
	kt.mu.Lock()
	kt.logs = append(kt.logs, msg)
	kt.mu.Unlock()
}
