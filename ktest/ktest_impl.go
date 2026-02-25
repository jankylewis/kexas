package ktest

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/kexas-project/kexas"
)

// discoverTestFunctions finds all Test* functions in the current package
func discoverTestFunctions() ([]reflect.Value, error) {
	// This is a simplified implementation using reflection
	// In a full implementation, we'd use go/parser to parse files

	// For now, we'll use a registration approach
	// Users would register their test functions
	return registeredAutoTests, nil
}

// findGlobalHook finds global hook functions (BeforeAll, AfterAll, etc.)
func findGlobalHook(hookName string) reflect.Value {
	// This would find global functions by name
	// For now, return zero value (hooks are optional)
	return reflect.Value{}
}

// registeredAutoTests holds test functions registered for auto-discovery
var registeredAutoTests []reflect.Value

// RegisterTest allows users to register test functions for auto-discovery
// Usage: ktest.RegisterTest(TestAmazonHomepage)
func RegisterTest(testFunc interface{}) {
	var val reflect.Value = reflect.ValueOf(testFunc)
	if val.Kind() == reflect.Func {
		registeredAutoTests = append(registeredAutoTests, val)
	}
}

// runTestsSequential runs tests one by one.
func runTestsSequential(t KTestT, suiteValue reflect.Value, baseSuite *Suite, methods []reflect.Method, config *Config) {
	for _, method := range methods {
		runSingleTest(t, suiteValue, baseSuite, method, config)
	}
}

// runTestsParallel runs tests in parallel.
func runTestsParallel(t KTestT, suiteValue reflect.Value, baseSuite *Suite, methods []reflect.Method, config *Config) {
	var wg sync.WaitGroup
	for _, method := range methods {
		wg.Add(1)
		var m reflect.Method = method
		go func() {
			defer wg.Done()
			runSingleTest(t, suiteValue, baseSuite, m, config)
		}()
	}
	wg.Wait()
}

// runSingleTest runs a single test method with retries.
func runSingleTest(t KTestT, suiteValue reflect.Value, baseSuite *Suite, method reflect.Method, config *Config) {
	var testName string = method.Name

	t.Run(testName, func(t KTestT) {
		var attempts int = config.Retries + 1
		var lastErr error

		for attempt := 0; attempt < attempts; attempt++ {
			if attempt > 0 {
				t.Logf("ktest: retry %d/%d for %s", attempt, config.Retries, testName)
			}

			var success bool = runTestAttempt(t, suiteValue, baseSuite, method, config)
			if success {
				return
			}

			lastErr = fmt.Errorf("test failed")
		}

		if lastErr != nil {
			t.Errorf("ktest: %s failed after %d attempts", testName, attempts)
		}
	})
}

// runTestAttempt runs a single attempt of a test.
func runTestAttempt(t KTestT, suiteValue reflect.Value, baseSuite *Suite, method reflect.Method, config *Config) bool {
	baseSuite.SetT(t)

	// Call BeforeEach hook
	callHook(suiteValue, "BeforeEach")

	// Track if test failed
	var failed bool = false

	// Recover from panics
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ktest: test panicked: %v", r)
			failed = true
		}

		// Take screenshot on failure
		if failed && config.ScreenshotOnFail {
			takeScreenshot(t, baseSuite.Page, method.Name, config.ScreenshotDir)
		}

		// Call AfterEach hook
		callHook(suiteValue, "AfterEach")
	}()

	// Run test with timeout
	var done chan bool = make(chan bool, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("ktest: test panicked: %v", r)
				failed = true
			}
			done <- true
		}()

		t.Logf("ktest: running %s", method.Name)
		var testMethod reflect.Value = suiteValue.MethodByName(method.Name)
		testMethod.Call(nil)
	}()

	select {
	case <-done:
		failed = t.Failed()
		return !failed
	case <-time.After(config.Timeout):
		t.Errorf("ktest: test timeout after %v", config.Timeout)
		failed = true
		return false
	}
}

// callHook calls a lifecycle hook method if it exists.
func callHook(suiteValue reflect.Value, hookName string) {
	var hook reflect.Value = suiteValue.MethodByName(hookName)
	if hook.IsValid() {
		hook.Call(nil)
	}
}

// findTestMethods returns all methods that start with "Test".
// If testFilter is non-empty, only returns methods matching the filter.
func findTestMethods(suiteType reflect.Type, testFilter string) []reflect.Method {
	var methods []reflect.Method = []reflect.Method{}

	for i := 0; i < suiteType.NumMethod(); i++ {
		var method reflect.Method = suiteType.Method(i)
		if strings.HasPrefix(method.Name, "Test") && method.Name != "TestMain" {
			// Apply filter if specified
			if testFilter == "" || strings.Contains(method.Name, testFilter) {
				methods = append(methods, method)
			}
		}
	}

	return methods
}

// takeScreenshot captures and saves a screenshot on test failure.
func takeScreenshot(t KTestT, page *kexas.Page, testName string, dir string) {
	t.Logf("ktest: capturing screenshot for failed test: %s", testName)

	var data []byte
	var err error
	data, err = page.Screenshot()
	if err != nil {
		t.Logf("ktest: failed to capture screenshot: %v", err)
		return
	}

	// Decode base64
	var imgData []byte
	imgData, err = base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		t.Logf("ktest: failed to decode screenshot: %v", err)
		return
	}

	// Save to file
	var timestamp string = time.Now().Format("20060102-150405")
	var filename string = fmt.Sprintf("%s-%s.png", testName, timestamp)
	var filepath string = filepath.Join(dir, filename)

	err = os.WriteFile(filepath, imgData, 0644)
	if err != nil {
		t.Logf("ktest: failed to save screenshot: %v", err)
		return
	}

	t.Logf("ktest: screenshot saved to %s", filepath)
}
