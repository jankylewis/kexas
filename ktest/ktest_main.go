package ktest

import (
	"os"
	"reflect"
	"testing"

	"github.com/jankylewis/kexas"
	"github.com/jankylewis/kexas/launcher"
)

// Main runs all test methods in the given suite without requiring the testing package.
// This is the recommended API for new tests - no "testing" import needed!
// Config is automatically loaded from kexas.config.json.
//
// To run specific tests, set KEXAS_TEST_RUN environment variable:
//
//	KEXAS_TEST_RUN=TestAmazonHomepage go run ./amz
func Main(suite interface{}) {
	var config *Config = loadConfig()
	MainWithConfig(suite, config)
}

// MainWithConfig executes all test methods with custom configuration.
// This is the complete testing framework API - no testing package required.
//
// To run specific tests, set KEXAS_TEST_RUN environment variable:
//
//	KEXAS_TEST_RUN=TestAmazonHomepage go run ./amz
func MainWithConfig(suite interface{}, config *Config) {
	var t *ktestT = newKTestT()

	var testFilter string = os.Getenv("KEXAS_TEST_RUN")
	if testFilter != "" {
		t.Logf("ktest: running only tests matching: %s", testFilter)
	}

	RunWithConfigInternal(t, suite, config, testFilter)

	if t.failed {
		os.Exit(1)
	}
}

// Run executes all test methods in the given suite with config from kexas.config.json.
// If kexas.config.json is not found, uses default configuration.
// Legacy API - requires testing package. Use Main() for new code.
func Run(t *testing.T, suite interface{}) {
	var config *Config = loadConfig()
	RunWithConfig(t, suite, config)
}

// RunWithConfig executes all test methods with custom configuration.
// Legacy API - requires testing package. Use MainWithConfig() for new code.
func RunWithConfig(t *testing.T, suite interface{}, config *Config) {
	var tWrapper *testingTWrapper = &testingTWrapper{t: t}
	RunWithConfigInternal(tWrapper, suite, config, "")
}

// RunWithConfigInternal is the shared implementation for both APIs.
func RunWithConfigInternal(t KTestT, suite interface{}, config *Config, testFilter string) {
	var suiteValue reflect.Value = reflect.ValueOf(suite)
	var suiteType reflect.Type = suiteValue.Type()

	var baseSuite *Suite = validateAndSetupSuite(t, suiteValue)
	if baseSuite == nil {
		return // Validation failed, error already reported
	}
	baseSuite.SetT(t)
	baseSuite.config = config

	var browser *kexas.Browser
	var page *kexas.Page
	browser, page = setupTestEnvironment(t, config)
	if browser == nil || page == nil {
		return // Setup failed, error already reported
	}
	defer browser.Close()
	defer page.Close()

	baseSuite.Browser = browser
	baseSuite.Page = page

	runTestLifecycle(t, suiteValue, suiteType, baseSuite, config, testFilter)
}

// validateAndSetupSuite validates the suite structure and returns the base suite
func validateAndSetupSuite(t KTestT, suiteValue reflect.Value) *Suite {
	if suiteValue.Kind() != reflect.Ptr {
		t.Fatal("ktest: suite must be a pointer")
		return nil
	}

	var suiteField reflect.Value = suiteValue.Elem().FieldByName("Suite")
	if !suiteField.IsValid() {
		t.Fatal("ktest: suite must embed ktest.Suite")
		return nil
	}

	var baseSuite *Suite = suiteField.Addr().Interface().(*Suite)
	return baseSuite
}

// setupTestEnvironment creates screenshot directory and launches browser
func setupTestEnvironment(t KTestT, config *Config) (*kexas.Browser, *kexas.Page) {
	if config.ScreenshotOnFail {
		os.MkdirAll(config.ScreenshotDir, 0755)
	}

	t.Log("ktest: launching browser")
	var opts *launcher.Options = launcher.DefaultOptions()
	opts.Headless = config.Headless
	if config.BrowserExecutable != "" {
		opts.ExecutablePath = config.BrowserExecutable
	}

	var browser *kexas.Browser
	var err error
	browser, err = kexas.Launch(opts)
	if err != nil {
		t.Fatalf("ktest: failed to launch browser: %v", err)
		return nil, nil
	}

	var page *kexas.Page
	page, err = browser.FirstPage()
	if err != nil {
		browser.Close()
		t.Fatalf("ktest: failed to get first page: %v", err)
		return nil, nil
	}

	return browser, page
}

// runTestLifecycle runs the complete test lifecycle with hooks
func runTestLifecycle(t KTestT, suiteValue reflect.Value, suiteType reflect.Type, baseSuite *Suite, config *Config, testFilter string) {
	t.Log("ktest: running BeforeAll")
	callHook(suiteValue, "BeforeAll")

	var testMethods []reflect.Method = findTestMethods(suiteType, testFilter)
	t.Logf("ktest: found %d test methods", len(testMethods))

	if len(testMethods) == 0 {
		t.Log("ktest: no test methods found (methods must start with 'Test')")
		return
	}

	var results []testResult
	if config.Parallel {
		results = runTestsParallel(t, suiteValue, baseSuite, testMethods, config)
	} else {
		results = runTestsSequential(t, suiteValue, baseSuite, testMethods, config)
	}

	t.Log("ktest: running AfterAll")
	callHook(suiteValue, "AfterAll")

	// Generate the HTML report after AfterAll so the report reflects the final
	// state of the suite (e.g., teardown failures count too if AfterAll uses t).
	if len(results) > 0 {
		generateHTMLReport(results, config)
	}
}
