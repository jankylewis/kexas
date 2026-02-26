package tests

import (
	"testing"
	"time"

	"github.com/kexas-project/kexas"
	"github.com/kexas-project/kexas/internal/cdp"
)

// TestElement_NewElement_Critical tests the core Element creation functionality
func TestElement_NewElement_Critical(t *testing.T) {
	// Test with valid parameters
	var page *kexas.Page = nil // Will be nil for testing
	var selector string = "#test-button"
	var nodeID cdp.NodeID = 12345
	var timeout time.Duration = 10 * time.Second

	var element *kexas.Element
	element = kexas.NewElement(page, selector, nodeID, timeout)

	// Verify element properties
	if element == nil {
		t.Fatal("expected element to be created")
	}

	if element.Selector() != selector {
		t.Errorf("expected selector %s, got %s", selector, element.Selector())
	}

	if element.NodeID() != nodeID {
		t.Errorf("expected nodeID %d, got %d", nodeID, element.NodeID())
	}

	if element.Timeout() != timeout {
		t.Errorf("expected timeout %v, got %v", timeout, element.Timeout())
	}

	// Test string representation
	var expected string = "Element{selector: #test-button, nodeID: 12345}"
	if element.String() != expected {
		t.Errorf("expected string representation %s, got %s", expected, element.String())
	}
}

// TestElement_Equals_Critical tests element comparison functionality
func TestElement_Equals_Critical(t *testing.T) {
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

	// Self comparison should return true
	if !element1.Equals(element1) {
		t.Error("element should equal itself")
	}
}

// TestElement_ZeroValues_Critical tests element with zero values
func TestElement_ZeroValues_Critical(t *testing.T) {
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

	if element.String() != "Element{selector: , nodeID: 0}" {
		t.Error("expected string representation with zero values")
	}
}

// TestElement_NegativeNodeID_Critical tests element with negative nodeID
func TestElement_NegativeNodeID_Critical(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test", -1, 10*time.Second)

	if element.NodeID() != -1 {
		t.Error("expected negative nodeID to be preserved")
	}

	if element.String() != "Element{selector: #test, nodeID: -1}" {
		t.Error("expected string representation with negative nodeID")
	}
}

// TestElement_GetText_Critical tests the GetText method implementation
func TestElement_GetText_Critical(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test-element", 12345, 10*time.Second)

	// With new implementation, GetText should return "element has no associated page" error
	var text string
	var err error
	text, err = element.GetText()

	if text != "" {
		t.Error("expected empty text for element with no page")
	}

	if err == nil {
		t.Error("expected error for element with no associated page")
	}

	var expected string = "element has no associated page"
	if err.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
	}
}

// TestElement_IsVisible_Critical tests visibility checking functionality
func TestElement_IsVisible_Critical(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#visible-element", 12345, 10*time.Second)

	// With new implementation, IsVisible should return false for element with no page
	var visible bool = element.IsVisible()

	if visible {
		t.Error("expected IsVisible to return false for element with no page")
	}
}

// TestElement_GetAttribute_Critical tests attribute retrieval functionality
func TestElement_GetAttribute_Critical(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test-element", 12345, 10*time.Second)
	var attributeName string = "href"

	var value string
	var err error
	value, err = element.GetAttribute(attributeName)

	if value != "" {
		t.Error("expected empty value for element with no page")
	}

	if err == nil {
		t.Error("expected error for element with no associated page")
	}

	var expected string = "element has no associated page"
	if err.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
	}
}

// TestElement_GetAttribute_EmptyName_Critical tests with empty attribute name
func TestElement_GetAttribute_EmptyName_Critical(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test-element", 12345, 10*time.Second)
	var attributeName string = ""

	var value string
	var err error
	value, err = element.GetAttribute(attributeName)

	if value != "" {
		t.Error("expected empty value for empty attribute name")
	}

	if err == nil {
		t.Error("expected error for empty attribute name")
	}

	var expected string = "element has no associated page"
	if err.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
	}
}

// TestElement_GetAttribute_SpecialCharacters_Critical tests with special characters in attribute name
func TestElement_GetAttribute_SpecialCharacters_Critical(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test-element", 12345, 10*time.Second)
	var attributeName string = "data-action='submit-form'"

	var value string
	var err error
	value, err = element.GetAttribute(attributeName)

	if value != "" {
		t.Error("expected empty value for element with no page")
	}

	if err == nil {
		t.Error("expected error for element with no associated page")
	}

	var expected string = "element has no associated page"
	if err.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
	}
}

// TestElement_Validate_Critical tests element validation functionality
func TestElement_Validate_Critical(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test-element", 12345, 10*time.Second)

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

// TestElement_Click_Critical tests click functionality
func TestElement_Click_Critical(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test-button", 12345, 10*time.Second)

	var err error
	err = element.Click()

	if err == nil {
		t.Error("expected error for element with no associated page")
	}

	var expected string = "element has no associated page"
	if err.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
	}
}

// TestElement_Type_Critical tests typing functionality
func TestElement_Type_Critical(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test-input", 12345, 10*time.Second)
	var testText string = "Hello, World!"

	var err error
	err = element.Type(testText)

	if err == nil {
		t.Error("expected error for element with no associated page")
	}

	var expected string = "element has no associated page"
	if err.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
	}
}

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
