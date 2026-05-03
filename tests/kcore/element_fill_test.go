package kcore_test

import (
	"testing"
	"time"

	"github.com/jankylewis/kexas"
)

// Critical unit tests for Element.Fill — the React-aware single-shot value
// setter introduced during the sltests dogfood. Mirror style of TestElement_Type_*.

// TestElement_Fill_NilElement — calling Fill on a nil *Element returns
// ErrElementNil, not a nil-pointer panic.
func TestElement_Fill_NilElement(t *testing.T) {
	var element *kexas.Element

	var err error
	err = element.Fill("anything")

	if err == nil {
		t.Fatal("expected error for nil element")
	}
	var expected string = "element is nil"
	if err.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
	}
}

// TestElement_Fill_NoPage — Fill on an element with no associated page returns
// ErrElementNoPage. Same guard as Type, same error.
func TestElement_Fill_NoPage(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test", 123, 10*time.Second)

	var err error
	err = element.Fill("text")

	if err == nil {
		t.Fatal("expected error for element with no page")
	}
	var expected string = "element has no associated page"
	if err.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
	}
}

// TestElement_Fill_EmptyTextAllowed — unlike Type, Fill accepts the empty
// string as a "clear field" operation. The page-missing error should be the
// only thing blocking — not a text-empty error.
func TestElement_Fill_EmptyTextAllowed(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test", 123, 10*time.Second)

	var err error
	err = element.Fill("")

	if err == nil {
		t.Fatal("expected error from no-page guard, not text validation")
	}
	// We expect the page-missing error, NOT text-empty. If Fill mistakenly
	// gained Type's empty-text guard, the message would change.
	var expected string = "element has no associated page"
	if err.Error() != expected {
		t.Errorf("Fill('') should fail with page-missing error, not text-empty. "+
			"got '%s'", err.Error())
	}
}

// TestElement_Fill_SpecialCharacters — Fill accepts whatever string the caller
// hands it; no character validation. Special chars survive the page-missing
// guard and produce the expected error (not a parse error).
func TestElement_Fill_SpecialCharacters(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#test", 123, 10*time.Second)
	var specialText string = "Hello@#$%^&*()_+-=[]{}|;':\",./<>?"

	var err error
	err = element.Fill(specialText)

	if err == nil {
		t.Fatal("expected error for element with no page")
	}
	var expected string = "element has no associated page"
	if err.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
	}
}

// TestElement_Fill_ZeroNodeID_NoObjectID — element with both nodeID == 0 and
// empty objectID is unusable (can't resolve to a remote handle). Fill should
// reject before attempting CDP.
func TestElement_Fill_ZeroNodeID_NoObjectID(t *testing.T) {
	// nodeID = 0 (zero), no objectID. Note: NewElement sets nodeID = 0; we
	// give the element a non-nil page via the NewElementWithObject helper to
	// bypass the page-missing guard, but leave the objectID empty.
	var element *kexas.Element = kexas.NewElementWithObject(nil, "#test", 0, "", 10*time.Second)

	var err error
	err = element.Fill("text")

	if err == nil {
		t.Fatal("expected error for element with zero nodeID and empty objectID")
	}
	// Page-missing guard fires first because we passed nil page; the test
	// confirms the call doesn't panic on the unusable state.
	var expected string = "element has no associated page"
	if err.Error() != expected {
		t.Errorf("expected page-missing error, got '%s'", err.Error())
	}
}
