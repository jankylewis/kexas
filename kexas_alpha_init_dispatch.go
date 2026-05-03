package kexas

import (
	"fmt"
	"reflect"
)

// isGroupFunction checks if the component is a ktest.Group function call.
// A Group has signature func(string, func()).
func isGroupFunction(component interface{}) bool {
	if component == nil {
		return false
	}

	var componentType reflect.Type = reflect.TypeOf(component)
	if componentType.Kind() != reflect.Func {
		return false
	}

	if componentType.NumIn() == 2 &&
		componentType.In(0).Kind() == reflect.String &&
		componentType.In(1).Kind() == reflect.Func {
		return true
	}

	return false
}

// isSuiteFunction checks if the component is a ktest.Suite function call.
// Suite shares Group's signature, so this delegates.
func isSuiteFunction(component interface{}) bool {
	return isGroupFunction(component)
}

// isHookFunction checks if the component is a hook function (BeforeAll, BeforeEach, etc.)
func isHookFunction(component interface{}) bool {
	if component == nil {
		return false
	}

	var componentType reflect.Type = reflect.TypeOf(component)

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

// isTestFunction checks if the component is a ktest.Test function call.
// A Test has signature func(string, func(*Page)).
func isTestFunction(component interface{}) bool {
	if component == nil {
		return false
	}

	var componentType reflect.Type = reflect.TypeOf(component)
	if componentType.Kind() != reflect.Func {
		return false
	}

	if componentType.NumIn() != 2 ||
		componentType.In(0).Kind() != reflect.String ||
		componentType.In(1).Kind() != reflect.Func {
		return false
	}

	var paramType reflect.Type = componentType.In(1)
	if paramType.NumIn() != 1 {
		return false
	}
	var pageType reflect.Type = paramType.In(0)
	return pageType.Kind() == reflect.Ptr && pageType.Elem().Name() == "Page"
}

// registerGroup registers a group component
func registerGroup(component interface{}) error {
	var componentValue reflect.Value = reflect.ValueOf(component)

	if componentValue.Type().NumIn() >= 2 {
		var name string = componentValue.Type().In(0).String()
		fmt.Printf("📁 Registering group: %s\n", name)

		if componentValue.Type().NumIn() == 2 {
			var groupFunc reflect.Value = componentValue
			if groupFunc.IsValid() && !groupFunc.IsNil() {
				fmt.Printf("✅ Group component acknowledged\n")
			}
		}
	}

	return nil
}

// registerSuite registers a suite component (alias for group)
func registerSuite(component interface{}) error {
	fmt.Printf("🏢 Registering suite component\n")
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
