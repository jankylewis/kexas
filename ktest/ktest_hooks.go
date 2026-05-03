package ktest

import (
	"fmt"
	"sync"

	"github.com/jankylewis/kexas"
)

// ========================================
// 🎯 KTEST GLOBAL HOOKS FOR ALPHA INIT PATTERN
// ========================================
//
// These functions provide global hook functionality for the AlphaInit
// zero-boilerplate pattern. They work alongside the Group system
// to provide TestNG/NUnit-like lifecycle management.
//
// Usage:
//   var _ = kexas.AlphaInit(
//       ktest.BeforeAll(func() {
//           // Global setup
//       }),
//       ktest.Group("Test Group", func() {
//           ktest.BeforeEach(func(page *kexas.Page) {
//               // Per-test setup
//           }),
//           ktest.Test("TestName", func(page *kexas.Page) {
//               // Test logic
//           }),
//       }),
//   )
// ========================================

// Global hooks storage (mutex-protected for parallel safety)
var (
	globalHooksMu    sync.Mutex
	globalBeforeAll  func()
	globalAfterAll   func()
	globalBeforeEach func(*kexas.Page)
	globalAfterEach  func(*kexas.Page)
)

// BeforeAll registers a global hook that runs once before all tests
func BeforeAll(hookFunc func()) interface{} {
	if hookFunc == nil {
		fmt.Println("❌ ktest.BeforeAll: hook function cannot be nil")
		return nil
	}

	globalHooksMu.Lock()
	globalBeforeAll = hookFunc
	globalHooksMu.Unlock()
	fmt.Println("✅ ktest.BeforeAll: Global before-all hook registered")
	return nil
}

// AfterAll registers a global hook that runs once after all tests
func AfterAll(hookFunc func()) interface{} {
	if hookFunc == nil {
		fmt.Println("❌ ktest.AfterAll: hook function cannot be nil")
		return nil
	}

	globalHooksMu.Lock()
	globalAfterAll = hookFunc
	globalHooksMu.Unlock()
	fmt.Println("✅ ktest.AfterAll: Global after-all hook registered")
	return nil
}

// BeforeEach registers a global hook that runs before each test
func BeforeEach(hookFunc func(*kexas.Page)) interface{} {
	if hookFunc == nil {
		fmt.Println("❌ ktest.BeforeEach: hook function cannot be nil")
		return nil
	}

	globalHooksMu.Lock()
	globalBeforeEach = hookFunc
	globalHooksMu.Unlock()
	fmt.Println("✅ ktest.BeforeEach: Global before-each hook registered")
	return nil
}

// AfterEach registers a global hook that runs after each test
func AfterEach(hookFunc func(*kexas.Page)) interface{} {
	if hookFunc == nil {
		fmt.Println("❌ ktest.AfterEach: hook function cannot be nil")
		return nil
	}

	globalHooksMu.Lock()
	globalAfterEach = hookFunc
	globalHooksMu.Unlock()
	fmt.Println("✅ ktest.AfterEach: Global after-each hook registered")
	return nil
}

// GetGlobalBeforeAll returns the registered global BeforeAll hook
func GetGlobalBeforeAll() func() {
	globalHooksMu.Lock()
	var hook func() = globalBeforeAll
	globalHooksMu.Unlock()
	return hook
}

// GetGlobalAfterAll returns the registered global AfterAll hook
func GetGlobalAfterAll() func() {
	globalHooksMu.Lock()
	var hook func() = globalAfterAll
	globalHooksMu.Unlock()
	return hook
}

// GetGlobalBeforeEach returns the registered global BeforeEach hook
func GetGlobalBeforeEach() func(*kexas.Page) {
	globalHooksMu.Lock()
	var hook func(*kexas.Page) = globalBeforeEach
	globalHooksMu.Unlock()
	return hook
}

// GetGlobalAfterEach returns the registered global AfterEach hook
func GetGlobalAfterEach() func(*kexas.Page) {
	globalHooksMu.Lock()
	var hook func(*kexas.Page) = globalAfterEach
	globalHooksMu.Unlock()
	return hook
}

// ExecuteGlobalBeforeAll executes the global BeforeAll hook if registered
func ExecuteGlobalBeforeAll() {
	var hook func() = GetGlobalBeforeAll()
	if hook != nil {
		fmt.Println("🚀 ktest: Executing global BeforeAll hook")
		hook()
	}
}

// ExecuteGlobalAfterAll executes the global AfterAll hook if registered
func ExecuteGlobalAfterAll() {
	var hook func() = GetGlobalAfterAll()
	if hook != nil {
		fmt.Println("🏁 ktest: Executing global AfterAll hook")
		hook()
	}
}

// ExecuteGlobalBeforeEach executes the global BeforeEach hook if registered
func ExecuteGlobalBeforeEach(page *kexas.Page) {
	var hook func(*kexas.Page) = GetGlobalBeforeEach()
	if hook != nil {
		fmt.Println("📖 ktest: Executing global BeforeEach hook")
		hook(page)
	}
}

// ExecuteGlobalAfterEach executes the global AfterEach hook if registered
func ExecuteGlobalAfterEach(page *kexas.Page) {
	var hook func(*kexas.Page) = GetGlobalAfterEach()
	if hook != nil {
		fmt.Println("🧹 ktest: Executing global AfterEach hook")
		hook(page)
	}
}

// ResetGlobalHooks clears all registered global hooks
// Useful for testing or when you want to re-register hooks
func ResetGlobalHooks() {
	globalHooksMu.Lock()
	globalBeforeAll = nil
	globalAfterAll = nil
	globalBeforeEach = nil
	globalAfterEach = nil
	globalHooksMu.Unlock()
	fmt.Println("🔄 ktest: Global hooks reset")
}

// ========================================
// 🎯 INTEGRATION WITH ALPHA INIT
// ========================================
//
// These global hooks integrate seamlessly with the AlphaInit pattern
// and the Group system to provide a complete testing lifecycle.
//
// Execution Order:
// 1. Global BeforeAll
// 2. Group BeforeEach (if any)
// 3. Global BeforeEach
// 4. Test Function
// 5. Global AfterEach
// 6. Group AfterEach (if any)
// 7. Global AfterAll
// ========================================
