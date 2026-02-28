package kassert

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// AssertionStats tracks pass/fail counts for assertions.
type AssertionStats struct {
	Passed int64
	Failed int64
}

// LogFunc is a function type for assertion logging.
// It receives the level ("PASS" or "FAIL"), the value name, and a description.
type LogFunc func(level string, name string, description string)

// package-level logger and stats
var (
	assertLogFunc LogFunc
	logMu         sync.RWMutex

	statsPassed atomic.Int64
	statsFailed atomic.Int64
)

// SetLogFunc sets the logging function for assertion pass/fail events.
// If set to nil, logging is disabled (default behavior).
//
// Thread-safe: can be called before test suite starts.
func SetLogFunc(fn LogFunc) {
	logMu.Lock()
	defer logMu.Unlock()
	assertLogFunc = fn
}

// ResetStats resets the assertion pass/fail counters to zero.
func ResetStats() {
	statsPassed.Store(0)
	statsFailed.Store(0)
}

// Stats returns the current assertion pass/fail counts.
func Stats() AssertionStats {
	return AssertionStats{
		Passed: statsPassed.Load(),
		Failed: statsFailed.Load(),
	}
}

// PrintStats prints a summary of assertion statistics.
func PrintStats() {
	var s AssertionStats = Stats()
	fmt.Printf("[kassert] Assertions: %d passed, %d failed\n", s.Passed, s.Failed)
}

// logPass logs a passing assertion and increments the pass counter.
func logPass(name string, assertion string) {
	statsPassed.Add(1)

	logMu.RLock()
	var fn LogFunc = assertLogFunc
	logMu.RUnlock()

	if fn != nil {
		fn("PASS", name, assertion)
	}
}

// logFail logs a failing assertion and increments the fail counter.
func logFail(name string, assertion string) {
	statsFailed.Add(1)

	logMu.RLock()
	var fn LogFunc = assertLogFunc
	logMu.RUnlock()

	if fn != nil {
		fn("FAIL", name, assertion)
	}
}
