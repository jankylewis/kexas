package tests

import (
	"testing"
	"time"

	"github.com/kexas-project/kexas"
	"github.com/kexas-project/kexas/internal/cdp"
)

func TestElement_NewElement(t *testing.T) {
	// Create a mock page (nil for now, will be implemented later)
	var page *kexas.Page = nil
	var selector string = "#test-button"
	var nodeID cdp.NodeID = 12345
	var timeout time.Duration = 10 * time.Second

	// Create element
	var element *kexas.Element = kexas.NewElement(page, selector, nodeID, timeout)

	// Verify element properties
	if element.Selector() != selector {
		t.Errorf("expected selector %s, got %s", selector, element.Selector())
	}

	if element.NodeID() != nodeID {
		t.Errorf("expected nodeID %d, got %d", nodeID, element.NodeID())
	}

	if element.Timeout() != timeout {
		t.Errorf("expected timeout %v, got %v", timeout, element.Timeout())
	}
}

func TestElement_String(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test", 123, 10*time.Second)
	var expected string = "Element{selector: #test, nodeID: 123}"

	if element.String() != expected {
		t.Errorf("expected %s, got %s", expected, element.String())
	}
}

func TestElement_Equals(t *testing.T) {
	var element1 *kexas.Element = kexas.NewElement(nil, "#test", 123, 10*time.Second)
	var element2 *kexas.Element = kexas.NewElement(nil, "#other", 123, 5*time.Second)
	var element3 *kexas.Element = kexas.NewElement(nil, "#test", 456, 10*time.Second)

	// Same nodeID should be equal
	if !element1.Equals(element2) {
		t.Error("elements with same nodeID should be equal")
	}

	// Different nodeID should not be equal
	if element1.Equals(element3) {
		t.Error("elements with different nodeID should not be equal")
	}

	// Comparison with nil should return false
	if element1.Equals(nil) {
		t.Error("element should not equal nil")
	}
}

func TestElement_IsVisible(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test", 123, 10*time.Second)

	// With new implementation, IsVisible returns false for element with no page
	if element.IsVisible() {
		t.Error("IsVisible should return false for element with no page")
	}
}

func TestElement_GetAttribute(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test", 123, 10*time.Second)

	// For now, GetAttribute returns error (placeholder implementation)
	var value string
	var err error
	value, err = element.GetAttribute("href")

	if value != "" {
		t.Errorf("expected empty value, got %s", value)
	}

	if err == nil {
		t.Error("expected error for unimplemented GetAttribute")
	}
}

func TestElement_GetText(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test", 123, 10*time.Second)

	// For now, GetText returns error (placeholder implementation)
	var text string
	var err error
	text, err = element.GetText()

	if text != "" {
		t.Errorf("expected empty text, got %s", text)
	}

	if err == nil {
		t.Error("expected error for unimplemented GetText")
	}
}

func TestElement_Validate(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test", 123, 10*time.Second)

	// With new implementation, Validate should return error for element with no page
	var err error
	err = element.Validate()

	if err == nil {
		t.Error("expected error for element with no associated page")
	}

	var expected string = "element has no associated page"
	if err.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestElement_ZeroValues(t *testing.T) {
	// Test element with zero values
	var element *kexas.Element = kexas.NewElement(nil, "", 0, 0)

	if element.Selector() != "" {
		t.Error("expected empty selector")
	}

	if element.NodeID() != 0 {
		t.Error("expected zero nodeID")
	}

	if element.Timeout() != 0 {
		t.Error("expected zero timeout")
	}
}

func TestElement_NegativeNodeID(t *testing.T) {
	// Test element with negative nodeID (should still work)
	var element *kexas.Element = kexas.NewElement(nil, "#test", -1, 10*time.Second)

	if element.NodeID() != -1 {
		t.Error("expected negative nodeID to be preserved")
	}

	if element.String() != "Element{selector: #test, nodeID: -1}" {
		t.Error("expected string representation with negative nodeID")
	}
}

func TestElement_Click(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test-button", 123, 10*time.Second)

	// For now, Click returns error (placeholder implementation)
	var err error
	err = element.Click()

	if err == nil {
		t.Error("expected error for unimplemented Click method")
	}

	var expected string = "element has no associated page"
	if err.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestElement_Click_NilElement(t *testing.T) {
	// Test click on element with nil page (should still attempt)
	var element *kexas.Element = kexas.NewElement(nil, "#test", 123, 10*time.Second)

	var err error
	err = element.Click()

	if err == nil {
		t.Error("expected error for unimplemented Click method")
	}

	// Should still return the not implemented error, not a nil pointer error
	var expected string = "element has no associated page"
	if err.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestElement_Click_ZeroTimeout(t *testing.T) {
	// Test click with zero timeout (should still work)
	var element *kexas.Element = kexas.NewElement(nil, "#test", 123, 0)

	var err error
	err = element.Click()

	if err == nil {
		t.Error("expected error for unimplemented Click method")
	}
}

func TestElement_Type(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test-input", 123, 10*time.Second)
	var testText string = "Hello, World!"

	// For now, Type returns error (placeholder implementation)
	var err error
	err = element.Type(testText)

	if err == nil {
		t.Error("expected error for unimplemented Type method")
	}

	var expected string = "element has no associated page"
	if err.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestElement_Type_EmptyText(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test-input", 123, 10*time.Second)

	// Test typing empty string
	var err error
	err = element.Type("")

	if err == nil {
		t.Error("expected error for unimplemented Type method")
	}
}

func TestElement_Type_SpecialCharacters(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test-input", 123, 10*time.Second)
	var specialText string = "Hello@#$%^&*()_+-=[]{}|;':\",./<>?"

	var err error
	err = element.Type(specialText)

	if err == nil {
		t.Error("expected error for unimplemented Type method")
	}

	// Should still return the not implemented error, not a parsing error
	var expected string = "element has no associated page"
	if err.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestElement_Type_NilElement(t *testing.T) {
	// Test type on element with nil page (should still attempt)
	var element *kexas.Element = kexas.NewElement(nil, "#test", 123, 10*time.Second)

	var err error
	err = element.Type("test")

	if err == nil {
		t.Error("expected error for unimplemented Type method")
	}

	// Should still return the not implemented error, not a nil pointer error
	var expected string = "element has no associated page"
	if err.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
	}
}

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
