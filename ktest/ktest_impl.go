package ktest

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/jankylewis/kexas"
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

// runTestsSequential runs tests one by one and collects per-test results
// for downstream HTML report generation.
func runTestsSequential(t KTestT, suiteValue reflect.Value, baseSuite *Suite, methods []reflect.Method, config *Config) []testResult {
	var results []testResult
	var suiteName string = suiteFilename(suiteValue)
	for _, method := range methods {
		results = append(results, runSingleTest(t, suiteValue, baseSuite, method, config, suiteName))
	}
	return results
}

// runTestsParallel runs tests in parallel and collects per-test results.
func runTestsParallel(t KTestT, suiteValue reflect.Value, baseSuite *Suite, methods []reflect.Method, config *Config) []testResult {
	var results []testResult
	var resultsMu sync.Mutex
	var wg sync.WaitGroup
	var suiteName string = suiteFilename(suiteValue)

	for _, method := range methods {
		wg.Add(1)
		var m reflect.Method = method
		go func() {
			defer wg.Done()
			var r testResult = runSingleTest(t, suiteValue, baseSuite, m, config, suiteName)
			resultsMu.Lock()
			results = append(results, r)
			resultsMu.Unlock()
		}()
	}
	wg.Wait()
	return results
}

// runSingleTest runs a single test method with retries and returns a testResult
// (used by the HTML report generator). Outcome flows back via closure variables
// because t.Run blocks until the closure completes.
//
// Always-on artifacts captured per test (regardless of pass/fail):
//   - One end-of-test screenshot at config.ScreenshotDir/<TestName>.png
//   - One per-test video recording at config.VideoDir/<TestName>.mp4
//     (falls back to <TestName>.mp4_frames/ when ffmpeg is missing)
func runSingleTest(t KTestT, suiteValue reflect.Value, baseSuite *Suite, method reflect.Method, config *Config, suiteName string) testResult {
	var testName string = method.Name
	var passed bool = true
	var elapsed time.Duration
	var errMsg string
	var screenshotPath string
	var videoPath string

	t.Run(testName, func(t KTestT) {
		var recorder *kexas.Recorder = startTestRecording(t, baseSuite.Page, config)
		var start time.Time = time.Now()
		passed, elapsed, errMsg = runWithRetries(t, suiteValue, baseSuite, method, config, testName, start)
		screenshotPath = captureEndOfTestScreenshot(t, baseSuite.Page, testName, config)
		videoPath = stopAndSaveTestVideo(t, recorder, testName, config)
	})

	return testResult{
		name:           testName,
		passed:         passed,
		filename:       suiteName,
		elapsed:        elapsed,
		workerID:       0,
		errorMsg:       errMsg,
		screenshotPath: screenshotPath,
		videoPath:      videoPath,
	}
}

// startTestRecording begins per-test video capture. Returns nil on disabled
// (config.VideoDir == "") or on failure (logged, not propagated).
func startTestRecording(t KTestT, page *kexas.Page, config *Config) *kexas.Recorder {
	if config.VideoDir == "" {
		return nil
	}
	var recorder *kexas.Recorder
	var err error
	recorder, err = page.StartRecording()
	if err != nil {
		t.Logf("ktest: video record start failed: %v", err)
		return nil
	}
	return recorder
}

// stopAndSaveTestVideo finalizes the recorder into config.VideoDir/<TestName>.mp4.
// Returns the saved path (empty on disabled / failure).
func stopAndSaveTestVideo(t KTestT, recorder *kexas.Recorder, testName string, config *Config) string {
	if recorder == nil {
		return ""
	}
	var stopErr error = recorder.Stop()
	if stopErr != nil {
		t.Logf("ktest: video record stop failed: %v", stopErr)
	}
	var mkdirErr error = os.MkdirAll(config.VideoDir, 0755)
	if mkdirErr != nil {
		t.Logf("ktest: failed to create video dir %s: %v", config.VideoDir, mkdirErr)
		return ""
	}
	var outputPath string = filepath.Join(config.VideoDir, testName+".mp4")
	var savedPath string
	var saveErr error
	savedPath, saveErr = recorder.SaveVideo(outputPath)
	if saveErr != nil {
		t.Logf("ktest: video save failed: %v", saveErr)
		return ""
	}
	return savedPath
}

// captureEndOfTestScreenshot takes one screenshot per test (always-on).
// Returns the saved path (empty when ScreenshotOnFail is false / on failure).
func captureEndOfTestScreenshot(t KTestT, page *kexas.Page, testName string, config *Config) string {
	if !config.ScreenshotOnFail {
		return ""
	}
	var data []byte
	var err error
	data, err = page.Screenshot()
	if err != nil {
		t.Logf("ktest: end-of-test screenshot failed: %v", err)
		return ""
	}
	err = os.MkdirAll(config.ScreenshotDir, 0755)
	if err != nil {
		t.Logf("ktest: failed to create screenshot dir: %v", err)
		return ""
	}
	var outputPath string = filepath.Join(config.ScreenshotDir, testName+".png")
	err = os.WriteFile(outputPath, data, 0644)
	if err != nil {
		t.Logf("ktest: failed to save screenshot: %v", err)
		return ""
	}
	return outputPath
}

// suiteFilename returns the suite's Go type name (e.g., "LoginSuite") for use
// as the report's per-suite grouping key. Falls back to "suite" if reflection
// can't extract a useful name.
func suiteFilename(suiteValue reflect.Value) string {
	var t reflect.Type = suiteValue.Type()
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	var name string = t.Name()
	if name == "" {
		return "suite"
	}
	return name
}

// runTestAttempt runs a single attempt of a test.
// runWithRetries runs the test up to config.Retries+1 times. Returns
// (passed, elapsed, errMsg). Logs each retry. Records final failure on `t`.
func runWithRetries(t KTestT, suiteValue reflect.Value, baseSuite *Suite, method reflect.Method, config *Config, testName string, start time.Time) (bool, time.Duration, string) {
	var attempts int = config.Retries + 1
	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			t.Logf("ktest: retry %d/%d for %s", attempt, config.Retries, testName)
		}
		var success bool = runTestAttempt(t, suiteValue, baseSuite, method, config)
		if success {
			return true, time.Since(start), ""
		}
	}
	t.Errorf("ktest: %s failed after %d attempts", testName, attempts)
	return false, time.Since(start), fmt.Sprintf("failed after %d attempts", attempts)
}

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

	// Save to file - data is already PNG bytes from page.Screenshot()
	var timestamp string = time.Now().Format("20060102-150405")
	var filename string = fmt.Sprintf("%s-%s.png", testName, timestamp)
	var filepath string = filepath.Join(dir, filename)

	err = os.WriteFile(filepath, data, 0644)
	if err != nil {
		t.Logf("ktest: failed to save screenshot: %v", err)
		return
	}

	t.Logf("ktest: screenshot saved to %s", filepath)
}
