
package kcore_test

import (
	"testing"
	"time"

	"github.com/jankylewis/kexas"
)

func TestElement_ClickWithTimeout(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test-button", 123, 10*time.Second)
	var timeout time.Duration = 5 * time.Second

	// For now, ClickWithTimeout returns error (placeholder implementation)
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

func TestElement_ClickWithTimeout_InvalidTimeout(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test-button", 123, 10*time.Second)
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

func TestElement_ClickWithTimeout_ZeroTimeout(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test-button", 123, 10*time.Second)
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

func TestElement_TypeWithTimeout(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test-input", 123, 10*time.Second)
	var timeout time.Duration = 3 * time.Second
	var testText string = "Hello, World!"

	// With new implementation, TypeWithTimeout returns error for element with no page
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

func TestElement_TypeWithTimeout_InvalidTimeout(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test-input", 123, 10*time.Second)
	var timeout time.Duration = 999 * time.Millisecond // Less than 1 second
	var text string = "test"

	var err error
	err = element.WaitAndTypeFor(text, timeout)

	if err == nil {
		t.Error("expected error for timeout less than 1 second")
	}

	var expected string = "timeout must be at least 1 second, got 999ms"
	if err.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestElement_TypeWithTimeout_EmptyText(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test-input", 123, 10*time.Second)
	var timeout time.Duration = 2 * time.Second

	// Test typing empty string with custom timeout
	var err error
	err = element.WaitAndTypeFor("", timeout)

	if err == nil {
		t.Error("expected error for element with no associated page")
	}

	var expected string = "element has no associated page"
	if err.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
	}
}
