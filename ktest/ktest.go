// Package ktest provides a robust test automation framework for Kexas.
//
// Inspired by Playwright Test and TestNG, ktest provides:
// - Suite-based test organization
// - Lifecycle hooks (BeforeAll, BeforeEach, AfterEach, AfterAll)
// - Browser and page fixtures
// - Parallel execution support
// - Test retries on failure
// - Screenshot on failure
// - Detailed reporting
// - Complete testing framework (no testing package needed)
//
// For assertions, use the kassert package.
//
// Basic usage (NEW API - no testing package needed):
//
//	type GoogleSuite struct {
//	    ktest.Suite
//	}
//
//	func (s *GoogleSuite) TestGoogleHomepage() {
//	    s.Page.Navigate("https://google.com")
//	    var title string
//	    var err error
//	    title, err = s.Page.Title()
//	    kassert.ThatError(s.T(), err).IsNil()
//	    kassert.That(s.T(), title).Contains("Google")
//	}
//
//	func main() {
//	    ktest.Main(new(GoogleSuite))
//	}
//
// Legacy API (still supported for backward compatibility):
//
//	func TestGoogle(t *testing.T) {
//	    ktest.Run(t, new(GoogleSuite))
//	}
package ktest

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/kexas-project/kexas"
	"github.com/kexas-project/kexas/launcher"
)

// Suite is the base type for test suites.
// Embed this in your test suite structs to get lifecycle hooks and fixtures.
type Suite struct {
	Browser *kexas.Browser
	Page    *kexas.Page
	t       KTestT // Changed from *testing.T to KTestT
	config  *Config
}

// KTestT interface matches the subset of testing.T methods we need.
// This allows us to create our own T implementation for the Main() API.
type KTestT interface {
	Helper()
	Log(args ...interface{})
	Logf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Failed() bool
	Name() string
	Run(name string, f func(KTestT)) bool
}

// Config holds test execution configuration.
type Config struct {
	Headless          bool
	Timeout           time.Duration
	Retries           int
	Parallel          bool
	ParallelSet       int
	ScreenshotOnFail  bool
	ScreenshotDir     string
	VideoDir          string
	ReportDir         string
	SlowMo            time.Duration
	BaseURL           string
	BrowserExecutable string
	ViewportWidth     int
	ViewportHeight    int
}

// DefaultConfig returns sensible default configuration.
func DefaultConfig() *Config {
	return &Config{
		Headless:         true,
		Timeout:          20 * time.Second, // Changed from 30s to 20s
		Retries:          0,
		Parallel:         false,
		ParallelSet:      1,
		ScreenshotOnFail: true,
		ScreenshotDir:    "./test-results/screenshots",
		VideoDir:         "./test-results/videos",
		ReportDir:        "./test-results",
		SlowMo:           0,
		BaseURL:          "",
		ViewportWidth:    1280,
		ViewportHeight:   720,
	}
}

// configJSON represents the JSON structure of kexas.config.json.
type configJSON struct {
	Headless          *bool   `json:"headless"`
	Timeout           *int    `json:"timeout"`
	Retries           *int    `json:"retries"`
	Parallel          *bool   `json:"parallel"`
	ParallelSet       *int    `json:"parallelSet"`
	ScreenshotOnFail  *bool   `json:"screenshotOnFail"`
	ScreenshotDir     *string `json:"screenshotDir"`
	VideoDir          *string `json:"videoDir"`
	ReportDir         *string `json:"reportDir"`
	SlowMo            *int    `json:"slowMo"`
	BaseURL           *string `json:"baseURL"`
	BrowserExecutable *string `json:"browserExecutable"`
	ViewportWidth     *int    `json:"viewportWidth"`
	ViewportHeight    *int    `json:"viewportHeight"`
}

// loadConfig loads configuration from kexas.config.json if it exists.
// Falls back to DefaultConfig() if file not found or invalid.
func loadConfig() *Config {
	return LoadConfigFromFile("./kexas.config.json")
}

// LoadConfigFromFile loads configuration from a specific file path.
// Primarily used by loadConfig and unit tests.
func LoadConfigFromFile(path string) *Config {
	var config *Config = DefaultConfig()

	// Try to read config file
	var data []byte
	var err error
	data, err = os.ReadFile(path)
	if err != nil {
		// File not found or can't read - use defaults
		return config
	}

	// Parse JSON
	var jsonConfig configJSON
	err = json.Unmarshal(data, &jsonConfig)
	if err != nil {
		// Invalid JSON - use defaults
		return config
	}

	// Apply config values (only if specified in JSON)
	if jsonConfig.Headless != nil {
		config.Headless = *jsonConfig.Headless
	}
	if jsonConfig.Timeout != nil {
		config.Timeout = time.Duration(*jsonConfig.Timeout) * time.Millisecond
	}
	if jsonConfig.Retries != nil {
		config.Retries = *jsonConfig.Retries
	}
	if jsonConfig.Parallel != nil {
		config.Parallel = *jsonConfig.Parallel
	}
	if jsonConfig.ParallelSet != nil && *jsonConfig.ParallelSet >= 1 {
		config.ParallelSet = *jsonConfig.ParallelSet
		if config.ParallelSet > 1 {
			config.Parallel = true
		}
	}
	if jsonConfig.ScreenshotOnFail != nil {
		config.ScreenshotOnFail = *jsonConfig.ScreenshotOnFail
	}
	if jsonConfig.ScreenshotDir != nil {
		config.ScreenshotDir = *jsonConfig.ScreenshotDir
	}
	if jsonConfig.VideoDir != nil {
		config.VideoDir = *jsonConfig.VideoDir
	}
	if jsonConfig.ReportDir != nil {
		config.ReportDir = *jsonConfig.ReportDir
	}
	if jsonConfig.SlowMo != nil {
		config.SlowMo = time.Duration(*jsonConfig.SlowMo) * time.Millisecond
	}
	if jsonConfig.BaseURL != nil {
		config.BaseURL = *jsonConfig.BaseURL
	}
	if jsonConfig.BrowserExecutable != nil {
		config.BrowserExecutable = *jsonConfig.BrowserExecutable
	}
	if jsonConfig.ViewportWidth != nil && *jsonConfig.ViewportWidth > 0 {
		config.ViewportWidth = *jsonConfig.ViewportWidth
	}
	if jsonConfig.ViewportHeight != nil && *jsonConfig.ViewportHeight > 0 {
		config.ViewportHeight = *jsonConfig.ViewportHeight
	}

	return config
}

// T returns the underlying KTestT for assertions.
func (s *Suite) T() KTestT {
	return s.t
}

// SetT sets the test instance (used internally).
func (s *Suite) SetT(t KTestT) {
	s.t = t
}

// Config returns the test configuration.
func (s *Suite) Config() *Config {
	return s.config
}

// BeforeAll runs once before all tests in the suite.
// Override this method in your suite to add setup logic.
func (s *Suite) BeforeAll() {
	// Default: do nothing
}

// AfterAll runs once after all tests in the suite.
// Override this method in your suite to add teardown logic.
func (s *Suite) AfterAll() {
	// Default: do nothing
}

// BeforeEach runs before each test method.
// Override this method in your suite to add per-test setup.
func (s *Suite) BeforeEach() {
	// Default: do nothing
}

// AfterEach runs after each test method.
// Override this method in your suite to add per-test teardown.
func (s *Suite) AfterEach() {
	// Default: do nothing
}

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
	// Create our own T implementation
	var t *ktestT = newKTestT()

	// Check for test filter environment variable
	var testFilter string = os.Getenv("KEXAS_TEST_RUN")
	if testFilter != "" {
		t.Logf("ktest: running only tests matching: %s", testFilter)
	}

	RunWithConfigInternal(t, suite, config, testFilter)

	// Exit with proper code
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

	// Validate suite structure
	var baseSuite *Suite = validateAndSetupSuite(t, suiteValue)
	if baseSuite == nil {
		return // Validation failed, error already reported
	}
	baseSuite.SetT(t)
	baseSuite.config = config

	// Setup test environment
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

	// Run test lifecycle
	runTestLifecycle(t, suiteValue, suiteType, baseSuite, config, testFilter)
}

// validateAndSetupSuite validates the suite structure and returns the base suite
func validateAndSetupSuite(t KTestT, suiteValue reflect.Value) *Suite {
	if suiteValue.Kind() != reflect.Ptr {
		t.Fatal("ktest: suite must be a pointer")
		return nil
	}

	// Get the embedded Suite field
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
	// Create screenshot directory
	if config.ScreenshotOnFail {
		os.MkdirAll(config.ScreenshotDir, 0755)
	}

	// Launch browser
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

	// Get the first page (default tab) instead of creating a new one
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
	// Call BeforeAll hook
	t.Log("ktest: running BeforeAll")
	callHook(suiteValue, "BeforeAll")

	// Find and run tests
	var testMethods []reflect.Method = findTestMethods(suiteType, testFilter)
	t.Logf("ktest: found %d test methods", len(testMethods))

	if len(testMethods) == 0 {
		t.Log("ktest: no test methods found (methods must start with 'Test')")
		return
	}

	// Run tests (parallel or sequential)
	if config.Parallel {
		runTestsParallel(t, suiteValue, baseSuite, testMethods, config)
	} else {
		runTestsSequential(t, suiteValue, baseSuite, testMethods, config)
	}

	// Call AfterAll hook
	t.Log("ktest: running AfterAll")
	callHook(suiteValue, "AfterAll")
}

// AutoRun is defined in ktest_autorun.go
