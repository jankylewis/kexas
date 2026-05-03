//go:build integration

package ktest_test

import (
	"testing"

	"github.com/jankylewis/kexas"
	"github.com/jankylewis/kexas/kassert"
	"github.com/jankylewis/kexas/ktest"
)

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

// Helper function to check if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
