
package kcore_test

import (
	"testing"
	"time"

	"github.com/jankylewis/kexas"
)

func TestElement_GetText_Enhanced(t *testing.T) {
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

	if err.Error() != "element has no associated page" {
		t.Errorf("expected error message 'element has no associated page', got '%s'", err.Error())
	}
}

func TestElement_GetText_NilElement_Enhanced(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test-element", 0, 0)

	var text string
	var err error
	text, err = element.GetText()

	if text != "" {
		t.Error("expected empty text for element with no page")
	}

	if err == nil {
		t.Error("expected error for element with no associated page")
	}

	if err.Error() != "element has no associated page" {
		t.Errorf("expected error message 'element has no associated page', got '%s'", err.Error())
	}
}

func TestElement_GetText_ZeroNodeID_Enhanced(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test-element", 0, 10*time.Second)

	var text string
	var err error
	text, err = element.GetText()

	if text != "" {
		t.Error("expected empty text for element with no page")
	}

	if err == nil {
		t.Error("expected error for element with no associated page")
	}

	if err.Error() != "element has no associated page" {
		t.Errorf("expected error message 'element has no associated page', got '%s'", err.Error())
	}
}

func TestElement_IsVisible_Enhanced(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#visible-element", 12345, 10*time.Second)

	// With new implementation, IsVisible should return false for element with no page
	var visible bool = element.IsVisible()

	if visible {
		t.Error("expected IsVisible to return false for element with no page")
	}
}

func TestElement_IsVisible_NilElement_Enhanced(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#visible-element", 0, 0)

	var visible bool = element.IsVisible()

	if visible {
		t.Error("expected IsVisible to return false for nil element")
	}
}

func TestElement_IsVisible_ZeroNodeID_Enhanced(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#visible-element", 0, 10*time.Second)

	var visible bool = element.IsVisible()

	if visible {
		t.Error("expected IsVisible to return false for element with invalid node ID")
	}
}

func TestElement_IsVisible_NegativeNodeID_Enhanced(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#visible-element", -1, 10*time.Second)

	var visible bool = element.IsVisible()

	if visible {
		t.Error("expected IsVisible to return false for element with negative node ID")
	}
}

func TestElement_GetAttribute_Enhanced(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test-element", 12345, 10*time.Second)
	var attributeName string = "href"

	// For now, GetAttribute returns error (placeholder implementation)
	var value string
	var err error
	value, err = element.GetAttribute(attributeName)

	if value != "" {
		t.Error("expected empty value for placeholder implementation")
	}

	if err == nil {
		t.Error("expected error for unimplemented GetAttribute method")
	}

	if err.Error() != "element has no associated page" {
		t.Errorf("expected error message 'element has no associated page', got '%s'", err.Error())
	}
}

func TestElement_GetAttribute_EmptyName_Enhanced(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test-element", 12345, 10*time.Second)
	var attributeName string = ""

	var value string
	var err error
	value, err = element.GetAttribute(attributeName)

	if value != "" {
		t.Error("expected empty value for placeholder implementation")
	}

	if err == nil {
		t.Error("expected error for unimplemented GetAttribute method")
	}

	if err.Error() != "element has no associated page" {
		t.Errorf("expected error message 'element has no associated page', got '%s'", err.Error())
	}
}

func TestElement_GetAttribute_SpecialCharacters_Enhanced(t *testing.T) {
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

	if err.Error() != "element has no associated page" {
		t.Errorf("expected error message 'element has no associated page', got '%s'", err.Error())
	}
}

func TestElement_GetAttribute_NilElement_Enhanced(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test-element", 0, 0)
	var attributeName string = "class"

	var value string
	var err error
	value, err = element.GetAttribute(attributeName)

	if value != "" {
		t.Error("expected empty value for nil element")
	}

	if err == nil {
		t.Error("expected error for nil element")
	}

	if err.Error() != "element has no associated page" {
		t.Errorf("expected error message 'element has no associated page', got '%s'", err.Error())
	}
}

func TestElement_Validate_Enhanced(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test-element", 12345, 10*time.Second)

	// With new implementation, Validate should return error for element with no page
	var err error
	err = element.Validate()

	if err == nil {
		t.Error("expected error for element with no associated page")
	}

	if err.Error() != "element has no associated page" {
		t.Errorf("expected error message 'element has no associated page', got '%s'", err.Error())
	}
}

func TestElement_Validate_NilElement_Enhanced(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test-element", 0, 0)

	var err error
	err = element.Validate()

	if err == nil {
		t.Error("expected error for nil element")
	}

	if err.Error() != "element has no associated page" {
		t.Errorf("expected error message 'element has no associated page', got '%s'", err.Error())
	}
}

func TestElement_Validate_ZeroNodeID_Enhanced(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test-element", 0, 10*time.Second)

	var err error
	err = element.Validate()

	if err == nil {
		t.Error("expected error for element with invalid node ID")
	}

	if err.Error() != "element has no associated page" {
		t.Errorf("expected error message 'element has no associated page', got '%s'", err.Error())
	}
}
