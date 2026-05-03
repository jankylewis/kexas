package ktest

import (
	"fmt"
	"os"
	"reflect"

	"github.com/jankylewis/kexas"
	"github.com/jankylewis/kexas/klauncher"
)

// AutoRun automatically discovers and runs Test* functions in the current package
// This provides the Playwright-style experience with zero boilerplate
func AutoRun() {
	fmt.Println("🧪 Kexas Auto-Discovery Test Runner")
	fmt.Println("=====================================")

	// First, try to run registered tests from AlphaInit/Group system
	var registeredTests []NamedTest = GetRegisteredTests()
	if len(registeredTests) > 0 {
		fmt.Printf("🔍 Found %d registered tests from AlphaInit:\n", len(registeredTests))
		for _, test := range registeredTests {
			fmt.Printf("  - %s\n", formatTestDisplayName(test))
		}
		runRegisteredTests(registeredTests)
		return
	}

	// Fall back to traditional Test* function discovery
	runDiscoveredTests()
}

// runDiscoveredTests finds and runs Test* functions via reflection.
func runDiscoveredTests() {
	var tests []reflect.Value
	var err error
	tests, err = discoverTestFunctions()
	if err != nil {
		fmt.Printf("❌ Failed to discover tests: %v\n", err)
		os.Exit(1)
	}

	if len(tests) == 0 {
		fmt.Println("ℹ️  No Test* functions found. Create functions starting with 'Test'.")
		return
	}

	fmt.Printf("🔍 Discovered %d test functions:\n", len(tests))
	for _, test := range tests {
		fmt.Printf("  - %s\n", test.Type().Name())
	}

	var t *ktestT = newKTestT()
	t.Log("ktest: launching browser")
	var browser *kexas.Browser
	browser, err = kexas.Launch(klauncher.DefaultOptions())
	if err != nil {
		t.Fatalf("ktest: failed to launch browser: %v", err)
	}
	defer browser.Close()

	executeDiscoveredHook(t, "BeforeAll", nil)
	runEachDiscoveredTest(t, tests, browser)
	executeDiscoveredHook(t, "AfterAll", nil)

	if t.failed {
		os.Exit(1)
	}
	fmt.Println("🎉 All tests completed!")
}

// runEachDiscoveredTest runs each discovered test with page isolation.
func runEachDiscoveredTest(t *ktestT, tests []reflect.Value, browser *kexas.Browser) {
	for _, testFunc := range tests {
		var testName string = testFunc.Type().Name()
		t.Run(testName, func(t KTestT) {
			t.Logf("ktest: running %s", testName)

			var page *kexas.Page
			var err error
			page, err = browser.NewPage()
			if err != nil {
				t.Errorf("ktest: failed to create page: %v", err)
				return
			}
			defer page.Close()

			executeDiscoveredHook(t, "BeforeEach", page)
			defer executeDiscoveredTestCleanup(t, testName, page)

			testFunc.Call([]reflect.Value{reflect.ValueOf(page)})
			t.Logf("ktest: %s completed", testName)
		})
	}
}

// executeDiscoveredHook runs a global hook if it exists.
func executeDiscoveredHook(t KTestT, hookName string, page *kexas.Page) {
	var hookFunc reflect.Value = findGlobalHook(hookName)
	if !hookFunc.IsValid() {
		return
	}
	t.Logf("ktest: running global %s", hookName)
	if page != nil {
		hookFunc.Call([]reflect.Value{reflect.ValueOf(page)})
	} else {
		hookFunc.Call(nil)
	}
}

// executeDiscoveredTestCleanup handles panic recovery and AfterEach for discovered tests.
func executeDiscoveredTestCleanup(t KTestT, testName string, page *kexas.Page) {
	if r := recover(); r != nil {
		t.Errorf("ktest: test %s panicked: %v", testName, r)
	}
	executeDiscoveredHook(t, "AfterEach", page)
}
