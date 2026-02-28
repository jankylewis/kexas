
package tests

import (
	"errors"
	"fmt"
	"testing"

	"github.com/kexas-project/kexas/kassert"
)

// ========================================
// MatchesRegex tests
// ========================================

func TestThat_MatchesRegex_Success(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, "https://www.saucedemo.com/inventory.html").MatchesRegex(`^https://www\.saucedemo\.com/.*`)
	if mock.failed {
		t.Error("Expected regex assertion to pass")
	}
}

func TestThat_MatchesRegex_Failure(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, "http://example.com").MatchesRegex(`^https://www\.saucedemo\.com/.*`)
	if !mock.failed {
		t.Error("Expected regex assertion to fail")
	}
}

func TestThat_MatchesRegex_InvalidPattern(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, "hello").MatchesRegex(`[invalid`)
	if !mock.failed {
		t.Error("Expected invalid regex to cause failure")
	}
}

func TestThat_MatchesRegex_NotString(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, 42).MatchesRegex(`\d+`)
	if !mock.failed {
		t.Error("Expected non-string to cause failure")
	}
}

func TestThat_MatchesRegex_EmptyString(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, "").MatchesRegex(`^$`)
	if mock.failed {
		t.Error("Expected empty string matching ^$ to pass")
	}
}

// ========================================
// IsOneOf tests
// ========================================

func TestThat_IsOneOf_Success(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, "enabled").IsOneOf("enabled", "disabled")
	if mock.failed {
		t.Error("Expected IsOneOf to pass")
	}
}

func TestThat_IsOneOf_Failure(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, "unknown").IsOneOf("enabled", "disabled")
	if !mock.failed {
		t.Error("Expected IsOneOf to fail")
	}
}

func TestThat_IsOneOf_SingleValue(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, 42).IsOneOf(42)
	if mock.failed {
		t.Error("Expected IsOneOf with single value to pass")
	}
}

func TestThat_IsOneOf_IntValues(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, 3).IsOneOf(1, 2, 3, 4)
	if mock.failed {
		t.Error("Expected IsOneOf with int values to pass")
	}
}

func TestThat_IsOneOf_NoMatch(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, "x").IsOneOf("a", "b", "c")
	if !mock.failed {
		t.Error("Expected IsOneOf to fail when no match")
	}
}

// ========================================
// ContainsAll tests
// ========================================

func TestThat_ContainsAll_Success(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, "Sauce Labs Backpack||Test.allTheThings()").ContainsAll("Sauce Labs Backpack", "Test.allTheThings()")
	if mock.failed {
		t.Error("Expected ContainsAll to pass")
	}
}

func TestThat_ContainsAll_Failure_OneMissing(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, "Sauce Labs Backpack||Other Product").ContainsAll("Sauce Labs Backpack", "Missing Product")
	if !mock.failed {
		t.Error("Expected ContainsAll to fail when one substring missing")
	}
}

func TestThat_ContainsAll_Failure_AllMissing(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, "hello world").ContainsAll("xyz", "abc")
	if !mock.failed {
		t.Error("Expected ContainsAll to fail when all substrings missing")
	}
}

func TestThat_ContainsAll_NotString(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, 42).ContainsAll("4", "2")
	if !mock.failed {
		t.Error("Expected ContainsAll to fail for non-string")
	}
}

func TestThat_ContainsAll_SingleSubstring(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, "hello world").ContainsAll("hello")
	if mock.failed {
		t.Error("Expected ContainsAll with single substring to pass")
	}
}

// ========================================
// HasLengthGreaterThan tests
// ========================================

func TestThat_HasLengthGreaterThan_String_Success(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, "hello").HasLengthGreaterThan(3)
	if mock.failed {
		t.Error("Expected HasLengthGreaterThan to pass")
	}
}

func TestThat_HasLengthGreaterThan_String_Failure(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, "hi").HasLengthGreaterThan(5)
	if !mock.failed {
		t.Error("Expected HasLengthGreaterThan to fail")
	}
}

func TestThat_HasLengthGreaterThan_String_Equal(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, "abc").HasLengthGreaterThan(3)
	if !mock.failed {
		t.Error("Expected HasLengthGreaterThan to fail when equal (not strictly greater)")
	}
}

func TestThat_HasLengthGreaterThan_Slice(t *testing.T) {
	var mock *mockT = &mockT{}
	var items []int = []int{1, 2, 3}
	kassert.That(mock, items).HasLengthGreaterThan(0)
	if mock.failed {
		t.Error("Expected HasLengthGreaterThan for slice to pass")
	}
}

func TestThat_HasLengthGreaterThan_Nil(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, nil).HasLengthGreaterThan(0)
	if !mock.failed {
		t.Error("Expected HasLengthGreaterThan to fail for nil")
	}
}

func TestThat_HasLengthGreaterThan_NotCollection(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.That(mock, 42).HasLengthGreaterThan(0)
	if !mock.failed {
		t.Error("Expected HasLengthGreaterThan to fail for non-collection")
	}
}

// ========================================
// ThatError.Is() tests
// ========================================

var errSentinel error = errors.New("sentinel error")

func TestThatError_Is_Success(t *testing.T) {
	var mock *mockT = &mockT{}
	var wrappedErr error = fmt.Errorf("wrapped: %w", errSentinel)
	kassert.ThatError(mock, wrappedErr).Is(errSentinel)
	if mock.failed {
		t.Error("Expected Is() to pass for wrapped sentinel error")
	}
}

func TestThatError_Is_ExactMatch(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.ThatError(mock, errSentinel).Is(errSentinel)
	if mock.failed {
		t.Error("Expected Is() to pass for exact sentinel match")
	}
}

func TestThatError_Is_Failure(t *testing.T) {
	var mock *mockT = &mockT{}
	var otherErr error = errors.New("other error")
	kassert.ThatError(mock, otherErr).Is(errSentinel)
	if !mock.failed {
		t.Error("Expected Is() to fail for non-matching error")
	}
}

func TestThatError_Is_NilError(t *testing.T) {
	var mock *mockT = &mockT{}
	kassert.ThatError(mock, nil).Is(errSentinel)
	if !mock.failed {
		t.Error("Expected Is() to fail for nil error")
	}
}

func TestThatError_Is_DeepWrapped(t *testing.T) {
	var mock *mockT = &mockT{}
	var level1 error = fmt.Errorf("level1: %w", errSentinel)
	var level2 error = fmt.Errorf("level2: %w", level1)
	kassert.ThatError(mock, level2).Is(errSentinel)
	if mock.failed {
		t.Error("Expected Is() to pass for deeply wrapped sentinel error")
	}
}

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
