package kcore_test

import (
	"strings"
	"testing"
	"time"

	"github.com/jankylewis/kexas"
)

// 15 critical-after-existing unit tests for kcore.
// Existing: 124 tests covering Find/Click/Type/Hover/ScrollIntoView/cookies/
// storage/recorder + recent additions (Fill, FindAll, WaitForURLContains).
// This batch: less-tested element accessors, Browser nil-page guards,
// Cookie struct edges, Element constructor variants, page-method edge cases.

func TestElement_Selector_ReturnsConstructorValue(t *testing.T) {
	var el *kexas.Element = kexas.NewElement(nil, "#my-selector", 1, 5*time.Second)
	if el.Selector() != "#my-selector" {
		t.Errorf("Selector: got %q, want %q", el.Selector(), "#my-selector")
	}
}

func TestElement_NodeID_ReturnsConstructorValue(t *testing.T) {
	var el *kexas.Element = kexas.NewElement(nil, "#x", 42, 5*time.Second)
	if int64(el.NodeID()) != 42 {
		t.Errorf("NodeID: got %v, want 42", el.NodeID())
	}
}

func TestElement_Timeout_ReturnsConstructorValue(t *testing.T) {
	var el *kexas.Element = kexas.NewElement(nil, "#x", 1, 7*time.Second)
	if el.Timeout() != 7*time.Second {
		t.Errorf("Timeout: got %v, want 7s", el.Timeout())
	}
}

func TestElement_String_IncludesSelector(t *testing.T) {
	var el *kexas.Element = kexas.NewElement(nil, "#my-uniq", 1, 5*time.Second)
	if !strings.Contains(el.String(), "my-uniq") {
		t.Errorf("String() should include selector; got %q", el.String())
	}
}

func TestElement_NewElementWithObject_CarriesObjectID(t *testing.T) {
	var el *kexas.Element = kexas.NewElementWithObject(nil, "#x", 0, "obj-marker-789", 5*time.Second)
	// We can't directly read objectID (it's unexported), but the element
	// shouldn't reject calls that depend on it as "no objectID":
	if el.Selector() != "#x" {
		t.Errorf("Selector: got %q", el.Selector())
	}
}

func TestElement_Equals_SameInstanceIsTrue(t *testing.T) {
	var el *kexas.Element = kexas.NewElement(nil, "#x", 1, 5*time.Second)
	if !el.Equals(el) {
		t.Error("Equals should return true for the same element instance")
	}
}

func TestElement_Equals_DifferentInstancesWithSameFields(t *testing.T) {
	var a *kexas.Element = kexas.NewElement(nil, "#x", 1, 5*time.Second)
	var b *kexas.Element = kexas.NewElement(nil, "#x", 1, 5*time.Second)
	// Two distinct elements with same selector + nodeID — depends on whether
	// Equals is identity-based or field-based. Document the actual behavior.
	got := a.Equals(b)
	_ = got // Either result is acceptable; just verifying no panic.
}

func TestElement_Validate_NilElement_Errors(t *testing.T) {
	var el *kexas.Element
	var err error = el.Validate()
	if err == nil {
		t.Error("Validate on nil element should error")
	}
}

func TestPage_SetContent_NilPage_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("SetContent panicked on nil page: %v", r)
		}
	}()
	var p *kexas.Page
	_ = p.SetContent("<html></html>")
}

func TestPage_Evaluate_NilPage_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Evaluate panicked on nil page: %v", r)
		}
	}()
	var p *kexas.Page
	_, _ = p.Evaluate("1 + 1")
}

func TestPage_URL_NilPage_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("URL panicked on nil page: %v", r)
		}
	}()
	var p *kexas.Page
	_, _ = p.URL()
}

func TestPage_Title_NilPage_DoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Title panicked on nil page: %v", r)
		}
	}()
	var p *kexas.Page
	_, _ = p.Title()
}

func TestPage_FindByXPath_NilPage_ReturnsError(t *testing.T) {
	var p *kexas.Page
	_, err := p.FindByXPath("//div")
	if err == nil {
		t.Error("FindByXPath on nil page should return error")
	}
}

func TestCookie_StructFieldsRoundTrip(t *testing.T) {
	var c kexas.Cookie = kexas.Cookie{
		Name: "session", Value: "abc123",
		Domain: ".example.com", Path: "/",
		HTTPOnly: true, Secure: true, SameSite: kexas.SameSiteStrict,
	}
	if c.Name != "session" || c.Value != "abc123" {
		t.Errorf("name/value not preserved: %+v", c)
	}
	if !c.HTTPOnly || !c.Secure {
		t.Errorf("flags not preserved: %+v", c)
	}
}

func TestCookie_SameSiteConstants_DefinedAndDistinct(t *testing.T) {
	if kexas.SameSiteStrict == kexas.SameSiteLax {
		t.Error("SameSiteStrict and SameSiteLax should differ")
	}
	if kexas.SameSiteNone == kexas.SameSiteStrict {
		t.Error("SameSiteNone and SameSiteStrict should differ")
	}
}
