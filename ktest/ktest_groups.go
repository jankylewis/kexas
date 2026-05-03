package ktest

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/jankylewis/kexas"
)

// ========================================
// KTEST GROUPS - TEST ORGANIZATION SYSTEM
// ========================================
//
// Group provides flexible test organization similar to TestNG test classes
// or Playwright test.describe() blocks. It supports unlimited nesting
// and group-level hooks for complete test organization.
//
// Usage:
//   ktest.Group("Amazon Authentication", func() {
//       ktest.Test("should login successfully", func(page *kexas.Page) {
//           // Test logic
//       }),
//   })
//
// This gives you the organization of TestNG with the flexibility of Playwright!
// ========================================

// TestPriority controls execution order when tests are sorted before dispatch.
// Lower numeric value = higher priority = runs first.
// Kexas uses alphabetical ordering: A runs before B, B before C, etc.
type TestPriority int

// prioritySet provides namespaced priority constants accessed via ktest.Priority.
// Alphabetical ordering: A (highest) through Z (lowest).
// Tests with Priority.A execute before Priority.B, and so on.
type prioritySet struct {
	A TestPriority
	B TestPriority
	C TestPriority
	D TestPriority
	E TestPriority
	F TestPriority
	G TestPriority
	H TestPriority
	I TestPriority
	J TestPriority
	K TestPriority
	L TestPriority
	M TestPriority
	N TestPriority
	O TestPriority
	P TestPriority
	Q TestPriority
	R TestPriority
	S TestPriority
	T TestPriority
	U TestPriority
	V TestPriority
	W TestPriority
	X TestPriority
	Y TestPriority
	Z TestPriority
}

// Priority provides alphabetically-ordered priority constants (A–Z).
// Usage: ktest.Test("Login", fn).WithPriority(ktest.Priority.A)
var Priority prioritySet = prioritySet{
	A: 0, B: 1, C: 2, D: 3, E: 4, F: 5, G: 6, H: 7, I: 8, J: 9,
	K: 10, L: 11, M: 12, N: 13, O: 14, P: 15, Q: 16, R: 17, S: 18, T: 19,
	U: 20, V: 21, W: 22, X: 23, Y: 24, Z: 25,
}

// Legacy aliases — deprecated, use ktest.Priority.A / .B / .C instead.
const (
	PriorityHigh   TestPriority = 0 // Deprecated: use Priority.A
	PriorityNormal TestPriority = 1 // Deprecated: use Priority.B
	PriorityLow    TestPriority = 2 // Deprecated: use Priority.C
)

// NamedTest represents a named test function
type NamedTest struct {
	Name     string
	Func     func(*kexas.Page, KTestT)
	Filename string
	Priority TestPriority
	Order    int // registration order for stable sort
}

// SortTestsByPriority sorts tests by priority (high first), then by registration order.
func SortTestsByPriority(tests []NamedTest) {
	sort.SliceStable(tests, func(i int, j int) bool {
		if tests[i].Priority != tests[j].Priority {
			return tests[i].Priority < tests[j].Priority
		}
		return tests[i].Order < tests[j].Order
	})
}

// Global test registration
var (
	registeredTests []NamedTest
	testMutex       sync.RWMutex
	globalTestOrder int
)

// registerNamedTest registers a test function globally and returns a pointer to it.
func registerNamedTest(name string, testFunc func(*kexas.Page, KTestT)) *NamedTest {
	testMutex.Lock()
	defer testMutex.Unlock()

	var baseFilename string = resolveCallerBaseFilename()
	globalTestOrder++
	registeredTests = append(registeredTests, NamedTest{
		Name:     name,
		Func:     testFunc,
		Filename: baseFilename,
		Priority: PriorityNormal,
		Order:    globalTestOrder,
	})
	var displayName string = fmt.Sprintf("<%s.%s>", baseFilename, extractShortName(name))
	fmt.Printf("✅ Registered test: %s\n", displayName)
	return &registeredTests[len(registeredTests)-1]
}

// resolveCallerBaseFilename walks the call stack and returns the base filename (no
// extension) of the first frame outside the ktest/kexas internals — i.e., the user's
// test file. Returns "test" as a fallback. Used by all test-registration entry points.
func resolveCallerBaseFilename() string {
	var callerFilename string
	for depth := 0; depth <= 10; depth++ {
		var _, file, _, ok = runtime.Caller(depth)
		if !ok {
			break
		}
		if file != "" && !strings.Contains(file, "/ktest/") && !strings.Contains(file, "/kexas/") {
			callerFilename = file
			break
		}
	}
	if callerFilename == "" {
		return "test"
	}
	var base string = filepath.Base(callerFilename)
	return strings.TrimSuffix(base, filepath.Ext(base))
}
