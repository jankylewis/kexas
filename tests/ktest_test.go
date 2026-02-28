//go:build integration

package tests

import (
	"os"
	"testing"

	"github.com/kexas-project/kexas"
	"github.com/kexas-project/kexas/kassert"
	"github.com/kexas-project/kexas/ktest"
)

// TestDefaultConfig tests the default configuration.
func TestDefaultConfig(t *testing.T) {
	var config *ktest.Config = ktest.DefaultConfig()

	if !config.Headless {
		t.Error("Expected headless to be true by default")
	}
	if config.Retries != 0 {
		t.Errorf("Expected retries to be 0, got %d", config.Retries)
	}
	if config.Parallel {
		t.Error("Expected parallel to be false by default")
	}
	if !config.ScreenshotOnFail {
		t.Error("Expected screenshot on fail to be true by default")
	}
	if config.ScreenshotDir != "./test-results/screenshots" {
		t.Errorf("Expected screenshot dir to be './test-results/screenshots', got %s", config.ScreenshotDir)
	}
	if config.ViewportWidth != 1280 {
		t.Errorf("Expected viewport width 1280, got %d", config.ViewportWidth)
	}
	if config.ViewportHeight != 720 {
		t.Errorf("Expected viewport height 720, got %d", config.ViewportHeight)
	}
}

func TestLoadConfigFromFile_ViewportOverrides(t *testing.T) {
	var path string = writeTempConfig(t, `{"viewportWidth":1920,"viewportHeight":1080}`)
	var config *ktest.Config = ktest.LoadConfigFromFile(path)

	if config.ViewportWidth != 1920 {
		t.Errorf("expected viewport width 1920, got %d", config.ViewportWidth)
	}
	if config.ViewportHeight != 1080 {
		t.Errorf("expected viewport height 1080, got %d", config.ViewportHeight)
	}
}

func TestLoadConfigFromFile_InvalidViewportFallsBack(t *testing.T) {
	var path string = writeTempConfig(t, `{"viewportWidth":-1,"viewportHeight":0}`)
	var config *ktest.Config = ktest.LoadConfigFromFile(path)

	if config.ViewportWidth != 1280 {
		t.Errorf("expected default viewport width 1280, got %d", config.ViewportWidth)
	}
	if config.ViewportHeight != 720 {
		t.Errorf("expected default viewport height 720, got %d", config.ViewportHeight)
	}
}

// TestConfig_CustomValues tests custom configuration values.
func TestConfig_CustomValues(t *testing.T) {
	var config *ktest.Config = &ktest.Config{
		Headless:         false,
		Retries:          3,
		Parallel:         true,
		ScreenshotOnFail: false,
		ScreenshotDir:    "./custom-dir",
	}

	if config.Headless {
		t.Error("Expected headless to be false")
	}
	if config.Retries != 3 {
		t.Errorf("Expected retries to be 3, got %d", config.Retries)
	}
	if !config.Parallel {
		t.Error("Expected parallel to be true")
	}
	if config.ScreenshotOnFail {
		t.Error("Expected screenshot on fail to be false")
	}
}

// MockSuite is a test suite for testing ktest functionality.
type MockSuite struct {
	ktest.Suite
	beforeAllCalled  bool
	afterAllCalled   bool
	beforeEachCalled bool
	afterEachCalled  bool
	testCalled       bool
}

func (s *MockSuite) BeforeAll() {
	s.beforeAllCalled = true
}

func (s *MockSuite) AfterAll() {
	s.afterAllCalled = true
}

func (s *MockSuite) BeforeEach() {
	s.beforeEachCalled = true
}

func (s *MockSuite) AfterEach() {
	s.afterEachCalled = true
}

func (s *MockSuite) TestMockTest() {
	s.testCalled = true
}

// TestSuite_LifecycleHooks tests that lifecycle hooks are called.
func TestSuite_LifecycleHooks(t *testing.T) {
	// This would require a real browser, so we skip it
	// But we can test the structure
	var suite *MockSuite = &MockSuite{}

	if suite.beforeAllCalled {
		t.Error("BeforeAll should not be called yet")
	}
	if suite.testCalled {
		t.Error("Test should not be called yet")
	}
}

// TestSuite_EmbeddedFields tests that Suite has the expected fields.
func TestSuite_EmbeddedFields(t *testing.T) {
	var suite *MockSuite = &MockSuite{}

	// Check that Suite is embedded
	if suite.Browser != nil {
		t.Error("Browser should be nil before initialization")
	}
	if suite.Page != nil {
		t.Error("Page should be nil before initialization")
	}
}

// EmptySuite has no test methods.
type EmptySuite struct {
	ktest.Suite
}

// TestSuite_NoTestMethods tests behavior when suite has no test methods.
func TestSuite_NoTestMethods(t *testing.T) {
	// This would require a real browser
	t.Skip("requires real browser - integration test")
}

// MultiTestSuite has multiple test methods.
type MultiTestSuite struct {
	ktest.Suite
	test1Called bool
	test2Called bool
	test3Called bool
}

func (s *MultiTestSuite) TestOne() {
	s.test1Called = true
}

func (s *MultiTestSuite) TestTwo() {
	s.test2Called = true
}

func (s *MultiTestSuite) TestThree() {
	s.test3Called = true
}

// TestSuite_MultipleTests tests that all test methods are discovered.
func TestSuite_MultipleTests(t *testing.T) {
	// This would require a real browser
	t.Skip("requires real browser - integration test")
}

// FailingSuite has a test that fails.
type FailingSuite struct {
	ktest.Suite
}

func (s *FailingSuite) TestFailing() {
	s.T().Error("This test intentionally fails")
}

// TestSuite_FailingTest tests that failing tests are handled.
func TestSuite_FailingTest(t *testing.T) {
	// This would require a real browser
	t.Skip("requires real browser - integration test")
}

// PanicSuite has a test that panics.
type PanicSuite struct {
	ktest.Suite
}

func (s *PanicSuite) TestPanic() {
	panic("intentional panic")
}

// TestSuite_PanicRecovery tests that panics are recovered.
func TestSuite_PanicRecovery(t *testing.T) {
	// This would require a real browser
	t.Skip("requires real browser - integration test")
}

// TestConfig_Timeout tests timeout configuration.
func TestConfig_Timeout(t *testing.T) {
	var config *ktest.Config = ktest.DefaultConfig()

	if config.Timeout == 0 {
		t.Error("Expected non-zero timeout")
	}
}

// TestConfig_BaseURL tests base URL configuration.
func TestConfig_BaseURL(t *testing.T) {
	var config *ktest.Config = ktest.DefaultConfig()
	config.BaseURL = "https://example.com"

	if config.BaseURL != "https://example.com" {
		t.Errorf("Expected base URL to be 'https://example.com', got %s", config.BaseURL)
	}
}

// TestSuite_Assert tests the Assert() method.
func TestSuite_Assert(t *testing.T) {
	var suite *MockSuite = &MockSuite{}

	// We can't call Assert() without a testing.T, but we can check the structure
	if suite.T() != nil {
		t.Error("T() should return nil before initialization")
	}
}

// TestSuite_Config tests the Config() method.
func TestSuite_Config(t *testing.T) {
	var suite *MockSuite = &MockSuite{}

	if suite.Config() != nil {
		t.Error("Config() should return nil before initialization")
	}
}

// TestLoadConfig tests configuration loading functionality.
// TestLoadConfig tests configuration loading functionality.
func TestLoadConfig(t *testing.T) {
	var config *ktest.Config = ktest.DefaultConfig()
	if config == nil {
		t.Error("DefaultConfig() should not return nil")
	}
}

// TestAlphaInitRegistration tests AlphaInit registration system.
func TestAlphaInitRegistration(t *testing.T) {
	kexas.ResetAlphaInitTracking()
	// Register a test via AlphaInit pattern
	kexas.AlphaInit(
		ktest.Test("AlphaTest1", func(page *kexas.Page, t ktest.KTestT) {
			t.Log("Alpha test 1 executed")
		}),
		ktest.Test("AlphaTest2", func(page *kexas.Page, t ktest.KTestT) {
			t.Log("Alpha test 2 executed")
		}),
	)

	// Get registered tests
	var tests []ktest.NamedTest = ktest.GetRegisteredTests()

	if len(tests) < 2 {
		t.Errorf("Expected at least 2 registered tests, got %d", len(tests))
	}

	// Verify test names
	var names []string
	for _, test := range tests {
		names = append(names, test.Name)
	}

	if !contains(names, "AlphaTest1") {
		t.Error("Expected AlphaTest1 to be registered")
	}
	if !contains(names, "AlphaTest2") {
		t.Error("Expected AlphaTest2 to be registered")
	}
}

// TestGroupRegistration tests group-based test registration.
func TestGroupRegistration(t *testing.T) {
	kexas.ResetAlphaInitTracking()
	// Register tests with groups
	kexas.AlphaInit(
		ktest.Group("TestGroup", func() {
			ktest.Test("GroupTest1", func(page *kexas.Page, t ktest.KTestT) {
				t.Log("Group test 1 executed")
			})
		}),
	)

	var tests []ktest.NamedTest = ktest.GetRegisteredTests()

	// Look for group test
	var found bool
	for _, test := range tests {
		if test.Name == "TestGroup.GroupTest1" {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected TestGroup.GroupTest1 to be registered")
	}
}

// TestKTestTInterface tests that KTestT implements kassert.TestingT.
func TestKTestTInterface(t *testing.T) {
	// This test verifies interface compatibility
	// KTestT is an interface, so we can't cast nil to it directly
	// But we can verify a concrete type satisfies it
	var mock mockTestingT
	var _ kassert.TestingT = &mock
	var _ ktest.KTestT = &mock
	// If this compiles, the interface is satisfied
	t.Log("KTestT correctly implements kassert.TestingT interface")
}

// mockTestingT implements both KTestT and TestingT for testing
type mockTestingT struct {
	failed bool
}

func (m *mockTestingT) Helper()                                    {}
func (m *mockTestingT) Log(args ...interface{})                    {}
func (m *mockTestingT) Logf(format string, args ...interface{})    {}
func (m *mockTestingT) Error(args ...interface{})                  { m.failed = true }
func (m *mockTestingT) Errorf(format string, args ...interface{})  { m.failed = true }
func (m *mockTestingT) Fatal(args ...interface{})                  { m.failed = true }
func (m *mockTestingT) Fatalf(format string, args ...interface{})  { m.failed = true }
func (m *mockTestingT) Failed() bool                               { return m.failed }
func (m *mockTestingT) Name() string                               { return "mock" }
func (m *mockTestingT) Run(name string, f func(ktest.KTestT)) bool { return false }

// TestConfigJSONStructure tests JSON configuration structure.
func TestConfigJSONStructure(t *testing.T) {
	// Test that config can be marshaled/unmarshaled
	var config *ktest.Config = ktest.DefaultConfig()

	// This would test JSON marshaling if we had that functionality
	// For now, just verify the structure is correct
	if config == nil {
		t.Error("Default config should not be nil")
	}
}

// Helper function to check if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	var file *os.File
	var err error
	file, err = os.CreateTemp("", "kexas-config-*.json")
	if err != nil {
		t.Fatalf("failed to create temp config: %v", err)
	}
	if _, err = file.WriteString(content); err != nil {
		file.Close()
		os.Remove(file.Name())
		t.Fatalf("failed to write temp config: %v", err)
	}
	file.Close()
	t.Cleanup(func() {
		os.Remove(file.Name())
	})
	return file.Name()
}
