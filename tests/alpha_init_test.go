
package tests

import (
	"testing"

	"github.com/kexas-project/kexas"
)

// TestEnforceOneAlphaInit_FirstCallSucceeds verifies first call from a file works.
func TestEnforceOneAlphaInit_FirstCallSucceeds(t *testing.T) {
	kexas.ResetAlphaInitTracking()

	// Should not panic
	var didPanic bool
	func() {
		defer func() {
			if r := recover(); r != nil {
				didPanic = true
			}
		}()
		kexas.AlphaInit()
	}()

	if didPanic {
		t.Error("first AlphaInit call should not panic")
	}
}

// TestEnforceOneAlphaInit_SecondCallPanics verifies second call from same file panics.
func TestEnforceOneAlphaInit_SecondCallPanics(t *testing.T) {
	kexas.ResetAlphaInitTracking()

	// First call — should succeed
	kexas.AlphaInit()

	// Second call from the same file — should panic
	var didPanic bool
	var panicMsg string
	func() {
		defer func() {
			if r := recover(); r != nil {
				didPanic = true
				panicMsg = r.(error).Error()
			}
		}()
		kexas.AlphaInit()
	}()

	if !didPanic {
		t.Error("second AlphaInit call from same file should panic")
	}
	if panicMsg == "" {
		t.Error("panic message should not be empty")
	}
	if !containsStr(panicMsg, "only one AlphaInit allowed per file") {
		t.Errorf("unexpected panic message: %s", panicMsg)
	}
}

// TestEnforceOneAlphaInit_ResetAllowsAgain verifies reset clears tracking.
func TestEnforceOneAlphaInit_ResetAllowsAgain(t *testing.T) {
	kexas.ResetAlphaInitTracking()

	// First call
	kexas.AlphaInit()

	// Reset
	kexas.ResetAlphaInitTracking()

	// Should succeed after reset
	var didPanic bool
	func() {
		defer func() {
			if r := recover(); r != nil {
				didPanic = true
			}
		}()
		kexas.AlphaInit()
	}()

	if didPanic {
		t.Error("AlphaInit should succeed after ResetAlphaInitTracking")
	}
}

// TestEnforceOneAlphaInit_PanicMessageContainsFilename verifies filename in error.
func TestEnforceOneAlphaInit_PanicMessageContainsFilename(t *testing.T) {
	kexas.ResetAlphaInitTracking()

	kexas.AlphaInit()

	var panicMsg string
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicMsg = r.(error).Error()
			}
		}()
		kexas.AlphaInit()
	}()

	// The panic message should contain the test file name
	if !containsStr(panicMsg, ".go") {
		t.Errorf("panic message should reference a .go file, got: %s", panicMsg)
	}
}

func containsStr(s string, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && stringContains(s, sub))
}

func stringContains(s string, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
