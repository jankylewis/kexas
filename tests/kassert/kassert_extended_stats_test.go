
package kassert_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jankylewis/kexas/kassert"
)

// ========================================
// AssertionStats tests
// ========================================

func TestAssertionStats_Counting(t *testing.T) {
	kassert.ResetStats()

	var mock *mockT = &mockT{}

	// 2 passing assertions
	kassert.That(mock, 42).Equals(42)
	kassert.That(mock, "hello").Contains("hello")

	// 1 failing assertion
	kassert.That(mock, 42).Equals(99)

	var stats kassert.AssertionStats = kassert.Stats()
	if stats.Passed != 2 {
		t.Errorf("Expected 2 passed, got %d", stats.Passed)
	}
	if stats.Failed != 1 {
		t.Errorf("Expected 1 failed, got %d", stats.Failed)
	}
}

func TestAssertionStats_Reset(t *testing.T) {
	kassert.ResetStats()

	var mock *mockT = &mockT{}
	kassert.That(mock, 42).Equals(42)

	kassert.ResetStats()

	var stats kassert.AssertionStats = kassert.Stats()
	if stats.Passed != 0 {
		t.Errorf("Expected 0 passed after reset, got %d", stats.Passed)
	}
	if stats.Failed != 0 {
		t.Errorf("Expected 0 failed after reset, got %d", stats.Failed)
	}
}

func TestAssertionStats_ErrorAssertions(t *testing.T) {
	kassert.ResetStats()

	var mock *mockT = &mockT{}
	var err error = errors.New("test error")

	// 1 pass
	kassert.ThatError(mock, err).IsNotNil()
	// 1 pass
	kassert.ThatError(mock, nil).IsNil()
	// 1 fail
	kassert.ThatError(mock, nil).IsNotNil()

	var stats kassert.AssertionStats = kassert.Stats()
	if stats.Passed != 2 {
		t.Errorf("Expected 2 passed, got %d", stats.Passed)
	}
	if stats.Failed != 1 {
		t.Errorf("Expected 1 failed, got %d", stats.Failed)
	}
}

// ========================================
// SetLogFunc tests
// ========================================

func TestSetLogFunc_Called(t *testing.T) {
	kassert.ResetStats()

	var logCalls []string
	kassert.SetLogFunc(func(level string, name string, description string) {
		logCalls = append(logCalls, fmt.Sprintf("%s:%s:%s", level, name, description))
	})
	defer kassert.SetLogFunc(nil)

	var mock *mockT = &mockT{}
	kassert.That(mock, 42).Named("answer").Equals(42)

	if len(logCalls) != 1 {
		t.Errorf("Expected 1 log call, got %d", len(logCalls))
		return
	}

	if logCalls[0] != "PASS:answer:Equals(42)" {
		t.Errorf("Expected PASS log, got: %s", logCalls[0])
	}
}

func TestSetLogFunc_Nil_NoLogs(t *testing.T) {
	kassert.SetLogFunc(nil)

	// Should not panic
	var mock *mockT = &mockT{}
	kassert.That(mock, 42).Equals(42)
	kassert.That(mock, 42).Equals(99)
}

// ========================================
// Chaining with new assertions
// ========================================

func TestThat_Chaining_Extended(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, "https://www.saucedemo.com/inventory.html").
		Named("page URL").
		Contains("saucedemo").
		StartsWith("https://").
		MatchesRegex(`^https://.*\.html$`).
		ContainsAll("saucedemo", "inventory")
	if mock.failed {
		t.Error("Expected all chained extended assertions to pass")
	}
}
