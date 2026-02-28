
package tests

import (
	"errors"
	"testing"
	"time"

	kexaserrors "github.com/kexas-project/kexas/errors"
)

func TestSentinelErrors(t *testing.T) {
	// Test that sentinel errors are properly defined
	if kexaserrors.ErrElementNotFound == nil {
		t.Error("ErrElementNotFound should not be nil")
	}

	if kexaserrors.ErrElementNotVisible == nil {
		t.Error("ErrElementNotVisible should not be nil")
	}

	if kexaserrors.ErrElementNotClickable == nil {
		t.Error("ErrElementNotClickable should not be nil")
	}

	if kexaserrors.ErrTimeout == nil {
		t.Error("ErrTimeout should not be nil")
	}

	if kexaserrors.ErrInvalidSelector == nil {
		t.Error("ErrInvalidSelector should not be nil")
	}

	if kexaserrors.ErrInvalidTimeout == nil {
		t.Error("ErrInvalidTimeout should not be nil")
	}

	if kexaserrors.ErrElementNotAttached == nil {
		t.Error("ErrElementNotAttached should not be nil")
	}

	if kexaserrors.ErrBrowserNotConnected == nil {
		t.Error("ErrBrowserNotConnected should not be nil")
	}
}

func TestElementError(t *testing.T) {
	var selector string = "#test-button"
	var operation string = "click"
	var cause error = errors.New("underlying error")

	var elementErr *kexaserrors.ElementError
	elementErr = kexaserrors.NewElementError(selector, operation, cause)

	// Test Error() method
	var expected string = "element operation 'click' failed for selector '#test-button': underlying error"
	if elementErr.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, elementErr.Error())
	}

	// Test Unwrap() method
	if errors.Unwrap(elementErr) != cause {
		t.Error("Unwrap should return the cause error")
	}

	// Test fields
	if elementErr.Selector != selector {
		t.Errorf("expected selector %s, got %s", selector, elementErr.Selector)
	}

	if elementErr.Operation != operation {
		t.Errorf("expected operation %s, got %s", operation, elementErr.Operation)
	}

	if elementErr.Cause != cause {
		t.Errorf("expected cause %v, got %v", cause, elementErr.Cause)
	}
}

func TestTimeoutError(t *testing.T) {
	var operation string = "WaitForElementVisible"
	var timeout time.Duration = 5 * time.Second
	var cause error = errors.New("timeout exceeded")

	var timeoutErr *kexaserrors.TimeoutError
	timeoutErr = kexaserrors.NewTimeoutError(operation, timeout, cause)

	// Test Error() method
	var expected string = "operation 'WaitForElementVisible' timed out after 5s: timeout exceeded"
	if timeoutErr.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, timeoutErr.Error())
	}

	// Test Unwrap() method
	if errors.Unwrap(timeoutErr) != cause {
		t.Error("Unwrap should return the cause error")
	}

	// Test fields
	if timeoutErr.Operation != operation {
		t.Errorf("expected operation %s, got %s", operation, timeoutErr.Operation)
	}

	if timeoutErr.Timeout != timeout {
		t.Errorf("expected timeout %v, got %v", timeout, timeoutErr.Timeout)
	}

	if timeoutErr.Cause != cause {
		t.Errorf("expected cause %v, got %v", cause, timeoutErr.Cause)
	}
}

func TestSelectorError(t *testing.T) {
	var selector string = "//button[@type='submit']"
	var selectorType string = "XPath"
	var cause error = errors.New("invalid XPath expression")

	var selectorErr *kexaserrors.SelectorError
	selectorErr = kexaserrors.NewSelectorError(selector, selectorType, cause)

	// Test Error() method
	var expected string = "selector error for XPath selector '//button[@type='submit']': invalid XPath expression"
	if selectorErr.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, selectorErr.Error())
	}

	// Test Unwrap() method
	if errors.Unwrap(selectorErr) != cause {
		t.Error("Unwrap should return the cause error")
	}

	// Test fields
	if selectorErr.Selector != selector {
		t.Errorf("expected selector %s, got %s", selector, selectorErr.Selector)
	}

	if selectorErr.Type != selectorType {
		t.Errorf("expected selector type %s, got %s", selectorType, selectorErr.Type)
	}

	if selectorErr.Cause != cause {
		t.Errorf("expected cause %v, got %v", cause, selectorErr.Cause)
	}
}

func TestValidationError(t *testing.T) {
	var selector string = "#test-element"
	var reason string = "element no longer attached to DOM"
	var cause error = errors.New("node ID not found")

	var validationErr *kexaserrors.ValidationError
	validationErr = kexaserrors.NewValidationError(selector, reason, cause)

	// Test Error() method
	var expected string = "element validation failed for selector '#test-element': element no longer attached to DOM"
	if validationErr.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, validationErr.Error())
	}

	// Test Unwrap() method
	if errors.Unwrap(validationErr) != cause {
		t.Error("Unwrap should return the cause error")
	}

	// Test fields
	if validationErr.Selector != selector {
		t.Errorf("expected selector %s, got %s", selector, validationErr.Selector)
	}

	if validationErr.Reason != reason {
		t.Errorf("expected reason %s, got %s", reason, validationErr.Reason)
	}

	if validationErr.Cause != cause {
		t.Errorf("expected cause %v, got %v", cause, validationErr.Cause)
	}
}

func TestErrorIs(t *testing.T) {
	// Test error.Is with sentinel errors
	var elementErr *kexaserrors.ElementError
	elementErr = kexaserrors.NewElementError("#test", "click", kexaserrors.ErrElementNotFound)

	if !errors.Is(elementErr, kexaserrors.ErrElementNotFound) {
		t.Error("elementErr should wrap ErrElementNotFound")
	}

	// Test error.Is with custom errors
	var timeoutErr *kexaserrors.TimeoutError
	timeoutErr = kexaserrors.NewTimeoutError("wait", 5*time.Second, kexaserrors.ErrTimeout)

	if !errors.Is(timeoutErr, kexaserrors.ErrTimeout) {
		t.Error("timeoutErr should wrap ErrTimeout")
	}
}

func TestErrorAs(t *testing.T) {
	// Test error.As with custom error types
	var err error
	err = kexaserrors.NewElementError("#test", "click", errors.New("test"))

	var elementErr *kexaserrors.ElementError
	if !errors.As(err, &elementErr) {
		t.Error("error should be convertible to ElementError")
	}

	if elementErr.Operation != "click" {
		t.Error("elementErr should have correct operation")
	}

	// Test with non-matching type
	var timeoutErr *kexaserrors.TimeoutError
	if errors.As(err, &timeoutErr) {
		t.Error("ElementError should not be convertible to TimeoutError")
	}
}

func TestErrorNilCause(t *testing.T) {
	// Test errors with nil cause
	var elementErr *kexaserrors.ElementError
	elementErr = kexaserrors.NewElementError("#test", "click", nil)

	if elementErr.Cause != nil {
		t.Error("cause should be nil")
	}

	if elementErr.Error() != "element operation 'click' failed for selector '#test': <nil>" {
		t.Errorf("unexpected error message: %s", elementErr.Error())
	}
}

func TestErrorEmptyFields(t *testing.T) {
	// Test errors with empty fields
	var elementErr *kexaserrors.ElementError
	elementErr = kexaserrors.NewElementError("", "", nil)

	if elementErr.Selector != "" {
		t.Error("selector should be empty")
	}

	if elementErr.Operation != "" {
		t.Error("operation should be empty")
	}
}
