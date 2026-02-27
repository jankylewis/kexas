package kexas

import (
	"fmt"
	"reflect"
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

// AlphaInit is the zero-boilerplate trigger function that registers
// all test components (groups, hooks, tests) for later execution.
// It can be called at package level without any boilerplate.
func AlphaInit(components ...interface{}) interface{} {
	fmt.Println("🔥 Kexas AlphaInit - Zero Boilerplate Trigger!")
	fmt.Printf("📦 Registering %d components...\n", len(components))

	// Process each component passed to AlphaInit
	for i := 0; i < len(components); i++ {
		var component interface{} = components[i]
		if err := registerComponent(component); err != nil {
			fmt.Printf("❌ Error registering component %d: %v\n", i+1, err)
		}
	}

	fmt.Printf("✅ AlphaInit completed - %d components processed\n", len(components))
	return nil
}

// registerComponent processes a single component and registers it
// with the appropriate ktest system
func registerComponent(component interface{}) error {
	// Skip nil components (hooks return nil after registration)
	if component == nil {
		return nil
	}

	// Use reflection to determine component type
	var componentValue reflect.Value = reflect.ValueOf(component)

	// Handle different component types
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

// isGroupFunction checks if the component is a ktest.Group function call
func isGroupFunction(component interface{}) bool {
	// For now, we'll check if it's a function that takes a string and func()
	if component == nil {
		return false
	}

	var componentType reflect.Type = reflect.TypeOf(component)
	if componentType.Kind() != reflect.Func {
		return false
	}

	// Check if it matches ktest.Group signature: func(string, func())
	if componentType.NumIn() == 2 &&
		componentType.In(0).Kind() == reflect.String &&
		componentType.In(1).Kind() == reflect.Func {
		return true
	}

	return false
}

// isSuiteFunction checks if the component is a ktest.Suite function call
func isSuiteFunction(component interface{}) bool {
	// Suite has the same signature as Group
	return isGroupFunction(component)
}

// isHookFunction checks if the component is a hook function (BeforeAll, BeforeEach, etc.)
func isHookFunction(component interface{}) bool {
	if component == nil {
		return false
	}

	var componentType reflect.Type = reflect.TypeOf(component)

	// Check for hook function signatures
	switch componentType {
	case reflect.TypeOf(func(func()) {}):
		// BeforeAll or AfterAll hook
		return true
	case reflect.TypeOf(func(func(*Page)) {}):
		// BeforeEach or AfterEach hook
		return true
	default:
		return false
	}
}

// isTestFunction checks if the component is a test function
func isTestFunction(component interface{}) bool {
	// For now, we'll check if it's a function that takes a string and func(*Page)
	if component == nil {
		return false
	}

	var componentType reflect.Type = reflect.TypeOf(component)
	if componentType.Kind() != reflect.Func {
		return false
	}

	// Check if it matches ktest.Test signature: func(string, func(*Page))
	if componentType.NumIn() == 2 &&
		componentType.In(0).Kind() == reflect.String &&
		componentType.In(1).Kind() == reflect.Func {
		// Check if second parameter is func(*Page)
		var paramType reflect.Type = componentType.In(1)
		if paramType.Kind() == reflect.Func && paramType.NumIn() == 1 {
			// Check if parameter is *Page
			var pageType reflect.Type = paramType.In(0)
			if pageType.Kind() == reflect.Ptr && pageType.Elem().Name() == "Page" {
				return true
			}
		}
	}

	return false
}

// registerGroup registers a group component
func registerGroup(component interface{}) error {
	// Execute the group function to register the group
	var componentValue reflect.Value = reflect.ValueOf(component)

	if componentValue.Type().NumIn() >= 2 {
		var name string = componentValue.Type().In(0).String()
		fmt.Printf("📁 Registering group: %s\n", name)

		// Call the group function
		if componentValue.Type().NumIn() == 2 {
			var groupFunc reflect.Value = componentValue
			if groupFunc.IsValid() && !groupFunc.IsNil() {
				// For now, we'll just acknowledge the group
				fmt.Printf("✅ Group component acknowledged\n")
			}
		}
	}

	return nil
}

// registerSuite registers a suite component (alias for group)
func registerSuite(component interface{}) error {
	fmt.Printf("🏢 Registering suite component\n")
	// Suite is just an alias for group
	return registerGroup(component)
}

// registerHook registers a hook component
func registerHook(component interface{}) error {
	fmt.Printf("🔗 Registering hook component: %T\n", component)
	// Implementation will be added when hook system is ready
	return nil
}

// registerTest registers a test component
func registerTest(component interface{}) error {
	fmt.Printf("🧪 Registering test component: %T\n", component)
	// Implementation will be added when ktest.Test is ready
	return nil
}
