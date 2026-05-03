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
	"github.com/jankylewis/kexas"
)

// Suite is the base type for test suites.
// Embed this in your test suite structs to get lifecycle hooks and fixtures.
type Suite struct {
	Browser *kexas.Browser
	Page    *kexas.Page
	t       KTestT
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

// AutoRun is defined in ktest_autorun.go
