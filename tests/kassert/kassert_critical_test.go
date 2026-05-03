package kassert_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jankylewis/kexas/kassert"
)

// renderingMockT is a richer mock than the package-shared mockT — it actually
// formats Errorf args via Sprintf so message-content assertions can inspect
// the rendered output.
type renderingMockT struct {
	failed  bool
	message string
}

func (m *renderingMockT) Helper()                  {}
func (m *renderingMockT) Error(args ...any)        { m.failed = true; m.message = fmt.Sprint(args...) }
func (m *renderingMockT) Errorf(f string, a ...any) { m.failed = true; m.message = fmt.Sprintf(f, a...) }
func (m *renderingMockT) Fatal(args ...any)        { m.failed = true; m.message = fmt.Sprint(args...) }
func (m *renderingMockT) Fatalf(f string, a ...any) { m.failed = true; m.message = fmt.Sprintf(f, a...) }
func (m *renderingMockT) Log(args ...any)          {}
func (m *renderingMockT) Logf(f string, a ...any)  {}

// 15 critical-after-existing unit tests for kassert.
// Existing: assertion methods, error matchers, MatchesRegex/IsOneOf/etc.
// This batch: chain return-self, multi-failure capture, edge cases on
// numeric comparators, assertion stats lifecycle, named-error formatting.

func TestThat_ReturnsAssertionForChaining(t *testing.T) {
	var m *mockT = &mockT{}
	var a *kassert.Assertion = kassert.That(m, "value")
	if a == nil {
		t.Fatal("That should return non-nil Assertion")
	}
	if a != a.Equals("value") {
		t.Error("Equals should return self for chaining")
	}
}

func TestThatError_ReturnsAssertionForChaining(t *testing.T) {
	var m *mockT = &mockT{}
	var a *kassert.ErrorAssertion = kassert.ThatError(m, nil)
	if a == nil {
		t.Fatal("ThatError should return non-nil ErrorAssertion")
	}
	if a != a.IsNil() {
		t.Error("IsNil should return self for chaining")
	}
}

func TestAssertion_Named_AddsContextToFailureMessage(t *testing.T) {
	var m *renderingMockT = &renderingMockT{}
	kassert.That(m, false).Named("my-flag").IsTrue()
	if !m.failed {
		t.Fatal("expected failure")
	}
	if !contains(m.message, "my-flag") {
		t.Errorf("expected 'my-flag' in rendered message; got %q", m.message)
	}
}

func TestErrorAssertion_Named_AddsContextToFailureMessage(t *testing.T) {
	var m *renderingMockT = &renderingMockT{}
	kassert.ThatError(m, errors.New("oops")).Named("login-call").IsNil()
	if !m.failed {
		t.Fatal("expected failure")
	}
	if !contains(m.message, "login-call") {
		t.Errorf("expected 'login-call' in rendered message; got %q", m.message)
	}
}

func TestAssertion_IsGreaterThan_OnEqualValue_Fails(t *testing.T) {
	var m *mockT = &mockT{}
	kassert.That(m, 5).IsGreaterThan(5)
	if !m.failed {
		t.Error("5 > 5 should fail")
	}
}

func TestAssertion_IsLessThan_OnEqualValue_Fails(t *testing.T) {
	var m *mockT = &mockT{}
	kassert.That(m, 5).IsLessThan(5)
	if !m.failed {
		t.Error("5 < 5 should fail")
	}
}

func TestAssertion_IsGreaterThanOrEqual_OnEqualValue_Passes(t *testing.T) {
	var m *mockT = &mockT{}
	kassert.That(m, 5).IsGreaterThanOrEqual(5)
	if m.failed {
		t.Error("5 >= 5 should pass")
	}
}

func TestAssertion_IsLessThanOrEqual_OnEqualValue_Passes(t *testing.T) {
	var m *mockT = &mockT{}
	kassert.That(m, 5).IsLessThanOrEqual(5)
	if m.failed {
		t.Error("5 <= 5 should pass")
	}
}

func TestAssertion_HasType_Match(t *testing.T) {
	var m *mockT = &mockT{}
	kassert.That(m, "string-value").HasType("")
	if m.failed {
		t.Errorf("expected type match; %q", m.message)
	}
}

func TestAssertion_HasType_Mismatch(t *testing.T) {
	var m *mockT = &mockT{}
	kassert.That(m, 42).HasType("")
	if !m.failed {
		t.Error("int vs string should fail HasType")
	}
}

func TestAssertion_StartsWith_NotString_Fails(t *testing.T) {
	var m *mockT = &mockT{}
	kassert.That(m, 42).StartsWith("4")
	if !m.failed {
		t.Error("StartsWith on int should fail")
	}
}

func TestAssertion_EndsWith_NotString_Fails(t *testing.T) {
	var m *mockT = &mockT{}
	kassert.That(m, []byte("hello")).EndsWith("lo")
	if !m.failed {
		t.Error("EndsWith on []byte should fail")
	}
}

func TestAssertionStats_RecordsPassFail(t *testing.T) {
	kassert.ResetStats()
	var m *mockT = &mockT{}
	kassert.That(m, true).IsTrue()  // pass
	kassert.That(m, false).IsTrue() // fail
	var stats kassert.AssertionStats = kassert.Stats()
	if stats.Passed < 1 {
		t.Errorf("expected ≥1 pass, got %d", stats.Passed)
	}
	if stats.Failed < 1 {
		t.Errorf("expected ≥1 fail, got %d", stats.Failed)
	}
}

func TestAssertionStats_ResetStats_ClearsCounts(t *testing.T) {
	var m *mockT = &mockT{}
	kassert.That(m, true).IsTrue()
	kassert.ResetStats()
	var stats kassert.AssertionStats = kassert.Stats()
	if stats.Passed != 0 || stats.Failed != 0 {
		t.Errorf("ResetStats should zero counts; got %+v", stats)
	}
}

func TestErrorAssertion_HasMessage_PartialMismatch_Fails(t *testing.T) {
	var m *mockT = &mockT{}
	kassert.ThatError(m, errors.New("login failed: 403")).HasMessage("login failed: 401")
	if !m.failed {
		t.Error("HasMessage with mismatched message should fail")
	}
}

// helper used in this file. Must be lowercase to avoid colliding with Go test.
func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
