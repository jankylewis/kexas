
package critical_test

import (
	"testing"
	"time"

	"github.com/jankylewis/kexas"
)

// TestElement_ClickWithTimeout_Critical tests click with custom timeout
func TestElement_ClickWithTimeout_Critical(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test-button", 12345, 10*time.Second)
	var timeout time.Duration = 5 * time.Second

	var err error
	err = element.WaitAndClickFor(timeout)

	if err == nil {
		t.Error("expected error for element with no associated page")
	}

	var expected string = "element has no associated page"
	if err.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
	}
}

// TestElement_ClickWithTimeout_InvalidTimeout_Critical tests timeout validation
func TestElement_ClickWithTimeout_InvalidTimeout_Critical(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test-button", 12345, 10*time.Second)
	var timeout time.Duration = 500 * time.Millisecond // Less than 1 second

	var err error
	err = element.WaitAndClickFor(timeout)

	if err == nil {
		t.Error("expected error for timeout less than 1 second")
	}

	var expected string = "timeout must be at least 1 second, got 500ms"
	if err.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
	}
}

// TestElement_ClickWithTimeout_ZeroTimeout_Critical tests zero timeout
func TestElement_ClickWithTimeout_ZeroTimeout_Critical(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test-button", 12345, 10*time.Second)
	var timeout time.Duration = 0

	var err error
	err = element.WaitAndClickFor(timeout)

	if err == nil {
		t.Error("expected error for zero timeout")
	}

	var expected string = "timeout must be at least 1 second, got 0s"
	if err.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
	}
}

// TestElement_TypeWithTimeout_Critical tests typing with custom timeout
func TestElement_TypeWithTimeout_Critical(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test-input", 12345, 10*time.Second)
	var testText string = "Hello, World!"
	var timeout time.Duration = 5 * time.Second

	var err error
	err = element.WaitAndTypeFor(testText, timeout)

	if err == nil {
		t.Error("expected error for element with no associated page")
	}

	var expected string = "element has no associated page"
	if err.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
	}
}

// TestElement_TypeWithTimeout_InvalidTimeout_Critical tests timeout validation for typing
func TestElement_TypeWithTimeout_InvalidTimeout_Critical(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test-input", 12345, 10*time.Second)
	var timeout time.Duration = 999 * time.Millisecond // Less than 1 second

	var err error
	err = element.WaitAndTypeFor("test", timeout)

	if err == nil {
		t.Error("expected error for timeout less than 1 second")
	}

	var expected string = "timeout must be at least 1 second, got 999ms"
	if err.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
	}
}

// TestElement_TypeWithTimeout_EmptyText_Critical tests typing with empty text
func TestElement_TypeWithTimeout_EmptyText_Critical(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test-input", 12345, 10*time.Second)
	var testText string = ""
	var timeout time.Duration = 5 * time.Second

	var err error
	err = element.WaitAndTypeFor(testText, timeout)

	if err == nil {
		t.Error("expected error for element with no associated page")
	}

	var expected string = "element has no associated page"
	if err.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
	}
}

// TestElement_TypeWithTimeout_NilElement_Critical tests typing with nil element
func TestElement_TypeWithTimeout_NilElement_Critical(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test", 123, 10*time.Second)
	var timeout time.Duration = 2 * time.Second

	var err error
	err = element.WaitAndTypeFor("test", timeout)

	if err == nil {
		t.Error("expected error for element with no associated page")
	}

	var expected string = "element has no associated page"
	if err.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
	}
}
