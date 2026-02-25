package ktest

import (
	"fmt"

	"github.com/kexas-project/kexas"
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

// Global hooks storage
var (
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

	globalBeforeAll = hookFunc
	fmt.Println("✅ ktest.BeforeAll: Global before-all hook registered")
	return nil
}

// AfterAll registers a global hook that runs once after all tests
func AfterAll(hookFunc func()) interface{} {
	if hookFunc == nil {
		fmt.Println("❌ ktest.AfterAll: hook function cannot be nil")
		return nil
	}

	globalAfterAll = hookFunc
	fmt.Println("✅ ktest.AfterAll: Global after-all hook registered")
	return nil
}

// BeforeEach registers a global hook that runs before each test
func BeforeEach(hookFunc func(*kexas.Page)) interface{} {
	if hookFunc == nil {
		fmt.Println("❌ ktest.BeforeEach: hook function cannot be nil")
		return nil
	}

	globalBeforeEach = hookFunc
	fmt.Println("✅ ktest.BeforeEach: Global before-each hook registered")
	return nil
}

// AfterEach registers a global hook that runs after each test
func AfterEach(hookFunc func(*kexas.Page)) interface{} {
	if hookFunc == nil {
		fmt.Println("❌ ktest.AfterEach: hook function cannot be nil")
		return nil
	}

	globalAfterEach = hookFunc
	fmt.Println("✅ ktest.AfterEach: Global after-each hook registered")
	return nil
}

// GetGlobalBeforeAll returns the registered global BeforeAll hook
func GetGlobalBeforeAll() func() {
	return globalBeforeAll
}

// GetGlobalAfterAll returns the registered global AfterAll hook
func GetGlobalAfterAll() func() {
	return globalAfterAll
}

// GetGlobalBeforeEach returns the registered global BeforeEach hook
func GetGlobalBeforeEach() func(*kexas.Page) {
	return globalBeforeEach
}

// GetGlobalAfterEach returns the registered global AfterEach hook
func GetGlobalAfterEach() func(*kexas.Page) {
	return globalAfterEach
}

// ExecuteGlobalBeforeAll executes the global BeforeAll hook if registered
func ExecuteGlobalBeforeAll() {
	if globalBeforeAll != nil {
		fmt.Println("🚀 ktest: Executing global BeforeAll hook")
		globalBeforeAll()
	}
}

// ExecuteGlobalAfterAll executes the global AfterAll hook if registered
func ExecuteGlobalAfterAll() {
	if globalAfterAll != nil {
		fmt.Println("🏁 ktest: Executing global AfterAll hook")
		globalAfterAll()
	}
}

// ExecuteGlobalBeforeEach executes the global BeforeEach hook if registered
func ExecuteGlobalBeforeEach(page *kexas.Page) {
	if globalBeforeEach != nil {
		fmt.Println("📖 ktest: Executing global BeforeEach hook")
		globalBeforeEach(page)
	}
}

// ExecuteGlobalAfterEach executes the global AfterEach hook if registered
func ExecuteGlobalAfterEach(page *kexas.Page) {
	if globalAfterEach != nil {
		fmt.Println("🧹 ktest: Executing global AfterEach hook")
		globalAfterEach(page)
	}
}

// ResetGlobalHooks clears all registered global hooks
// Useful for testing or when you want to re-register hooks
func ResetGlobalHooks() {
	globalBeforeAll = nil
	globalAfterAll = nil
	globalBeforeEach = nil
	globalAfterEach = nil
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
