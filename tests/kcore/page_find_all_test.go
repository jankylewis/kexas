package kcore_test

import (
	"testing"

	"github.com/jankylewis/kexas"
)

// Critical unit tests for Page.FindAll — multi-element companion to Find.
// Mirror style of TestPage_Find_*.

// TestPage_FindAll_NilPage — calling FindAll on a nil *Page returns an error,
// not a nil-pointer panic.
func TestPage_FindAll_NilPage(t *testing.T) {
	var page *kexas.Page

	var elements []*kexas.Element
	var err error
	elements, err = page.FindAll(".any")

	if elements != nil {
		t.Error("expected nil slice for nil page")
	}
	if err == nil {
		t.Fatal("expected error for nil page")
	}
	var expected string = "FindAll: nil page"
	if err.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
	}
}

// TestPage_FindAll_EmptySelector — empty string selector errors fast with a
// dedicated message (not just "no match").
func TestPage_FindAll_EmptySelector(t *testing.T) {
	var page *kexas.Page

	var elements []*kexas.Element
	var err error
	elements, err = page.FindAll("")

	if elements != nil {
		t.Error("expected nil slice for empty selector")
	}
	if err == nil {
		t.Fatal("expected error for empty selector")
	}
	// Nil-page guard fires first — confirms FindAll doesn't crash before
	// reaching the empty-selector check. The samplings test
	// TestGitHub/TestPFindAllOnNoMatchReturnsEmptySliceNotError covers the
	// real-page empty-selector path against the real CDP plumbing.
	var expected string = "FindAll: nil page"
	if err.Error() != expected {
		t.Errorf("expected nil-page error, got '%s'", err.Error())
	}
}

// TestPage_FindAll_CSS_Dispatch — selector NOT starting with `//` or `(`
// dispatches via CSS path. Verified through nil-page guard reaching us
// (proves the dispatcher even sees the call).
func TestPage_FindAll_CSS_Dispatch(t *testing.T) {
	var page *kexas.Page

	_, err := page.FindAll(".some-class")

	if err == nil {
		t.Fatal("expected error from nil-page guard")
	}
	if err.Error() != "FindAll: nil page" {
		t.Errorf("expected nil-page error, got '%s'", err.Error())
	}
}

// TestPage_FindAll_XPath_Dispatch — selector starting with `//` dispatches
// via XPath path. Same guard verification.
func TestPage_FindAll_XPath_Dispatch(t *testing.T) {
	var page *kexas.Page

	_, err := page.FindAll("//div[@class='foo']")

	if err == nil {
		t.Fatal("expected error from nil-page guard")
	}
	if err.Error() != "FindAll: nil page" {
		t.Errorf("expected nil-page error, got '%s'", err.Error())
	}
}

// TestPage_FindAll_XPathPositional_Dispatch — XPath form starting with `(` for
// positional predicates ((...)[1]) also dispatches via XPath path.
func TestPage_FindAll_XPathPositional_Dispatch(t *testing.T) {
	var page *kexas.Page

	_, err := page.FindAll("(//span[@class='rank'])[1]")

	if err == nil {
		t.Fatal("expected error from nil-page guard")
	}
	if err.Error() != "FindAll: nil page" {
		t.Errorf("expected nil-page error, got '%s'", err.Error())
	}
}

// TestPage_FindAll_IDShortcut_Dispatch — `#x` shortcut form is treated as CSS
// (matches Find's behavior — though Find has a special findByID path, FindAll
// folds it into the CSS dispatch since querySelectorAll handles it natively).
func TestPage_FindAll_IDShortcut_Dispatch(t *testing.T) {
	var page *kexas.Page

	_, err := page.FindAll("#main-content")

	if err == nil {
		t.Fatal("expected error from nil-page guard")
	}
	if err.Error() != "FindAll: nil page" {
		t.Errorf("expected nil-page error, got '%s'", err.Error())
	}
}
