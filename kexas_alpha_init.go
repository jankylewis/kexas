package kexas

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"sync"
)

// ========================================
// 🎯 KEXAS ALPHA INIT - ZERO BOILERPLATE TRIGGER
// ========================================
//
// AlphaInit provides the ultimate zero-boilerplate testing experience.
// It serves as a framework-level trigger that handles all the magic
// of test registration without requiring func init() or func main().
//
// Usage:
//   kexas.AlphaInit(
//       ktest.Group("Amazon Tests", func() {
//           ktest.Test("should load homepage", func(page *kexas.Page) {
//               page.Navigate("https://amazon.com")
//           }),
//       }),
//   )
//
// This gives you the TestNG/NUnit annotation experience in Go!
// ========================================

// alphaInitFiles tracks which source files have already called AlphaInit.
// Only one AlphaInit call is allowed per .go file.
var alphaInitFiles sync.Map

// AlphaInit is the zero-boilerplate trigger function that registers
// all test components (groups, hooks, tests) for later execution.
// It can be called at package level without any boilerplate.
// Only one AlphaInit call is allowed per source file.
func AlphaInit(components ...interface{}) interface{} {
	var callerFile string = alphaInitCallerFile()
	if err := enforceOneAlphaInitPerFile(callerFile); err != nil {
		panic(err)
	}

	fmt.Println("🔥 Kexas AlphaInit - Zero Boilerplate Trigger!")
	fmt.Printf("📦 Registering %d components from %s...\n", len(components), callerFile)

	for i := 0; i < len(components); i++ {
		var component interface{} = components[i]
		if err := registerComponent(component); err != nil {
			fmt.Printf("❌ Error registering component %d: %v\n", i+1, err)
		}
	}

	fmt.Printf("✅ AlphaInit completed - %d components processed\n", len(components))
	return nil
}

// alphaInitCallerFile walks the call stack to find the user's source file.
// Caller(0) = alphaInitCallerFile, Caller(1) = AlphaInit, Caller(2) = user file.
func alphaInitCallerFile() string {
	var _, file, _, ok = runtime.Caller(2)
	if !ok {
		return "unknown"
	}
	return filepath.Base(file)
}

// enforceOneAlphaInitPerFile returns an error if the file already called AlphaInit.
func enforceOneAlphaInitPerFile(filename string) error {
	if filename == "unknown" {
		return nil
	}
	var _, loaded = alphaInitFiles.LoadOrStore(filename, true)
	if loaded {
		return fmt.Errorf("kexas: only one AlphaInit allowed per file, but '%s' called it twice", filename)
	}
	return nil
}

// ResetAlphaInitTracking clears the per-file tracking. Used only in tests.
func ResetAlphaInitTracking() {
	alphaInitFiles.Range(func(key interface{}, value interface{}) bool {
		alphaInitFiles.Delete(key)
		return true
	})
}

// registerComponent processes a single component and dispatches it to the
// appropriate ktest registration path based on its reflected signature.
// Nil components (e.g., hooks that have already self-registered) are skipped.
func registerComponent(component interface{}) error {
	if component == nil {
		return nil
	}

	var componentValue reflect.Value = reflect.ValueOf(component)

	switch {
	case isGroupFunction(component):
		return registerGroup(component)

	case isSuiteFunction(component):
		return registerSuite(component)

	case isHookFunction(component):
		return registerHook(component)

	case isTestFunction(component):
		return registerTest(component)

	default:
		return fmt.Errorf("unsupported component type: %T (value: %v)", component, componentValue)
	}
}
