package kcore_test

import (
	"strings"
	"testing"
	"time"

	"github.com/jankylewis/kexas"
)

// 15 tests for the new advanced waiters: WaitForElementHidden,
// WaitForElementDetached, WaitForFunction, WaitForTitleContains,
// WaitForElementText. Nil-page guards + timeout-validation are the focus
// here (real DOM behavior is exercised in samplings).

// --- WaitForElementHidden ---

func TestWaitForElementHidden_NilPage_Errors(t *testing.T) {
	var p *kexas.Page
	err := p.WaitForElementHidden(".x", 200*time.Millisecond)
	if err == nil {
		t.Fatal("expected error from nil page")
	}
	if !strings.Contains(err.Error(), "nil page") {
		t.Errorf("expected nil-page error; got %q", err.Error())
	}
}

func TestWaitForElementHidden_BelowTimeoutFloor_Errors(t *testing.T) {
	var p *kexas.Page
	err := p.WaitForElementHidden(".x", 50*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout-validation error")
	}
	if !strings.Contains(err.Error(), "100ms") {
		t.Errorf("expected 100ms-floor message; got %q", err.Error())
	}
}

func TestWaitForElementHidden_ZeroTimeout_Errors(t *testing.T) {
	var p *kexas.Page
	err := p.WaitForElementHidden(".x", 0)
	if err == nil {
		t.Fatal("expected zero-timeout error")
	}
}

// --- WaitForElementDetached ---

func TestWaitForElementDetached_NilPage_Errors(t *testing.T) {
	var p *kexas.Page
	err := p.WaitForElementDetached(".x", 200*time.Millisecond)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "nil page") {
		t.Errorf("expected nil-page error; got %q", err.Error())
	}
}

func TestWaitForElementDetached_BelowTimeoutFloor_Errors(t *testing.T) {
	var p *kexas.Page
	err := p.WaitForElementDetached(".x", 50*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestWaitForElementDetached_NegativeTimeout_Errors(t *testing.T) {
	var p *kexas.Page
	err := p.WaitForElementDetached(".x", -1*time.Second)
	if err == nil {
		t.Fatal("expected negative-timeout error")
	}
}

// --- WaitForFunction ---

func TestWaitForFunction_NilPage_Errors(t *testing.T) {
	var p *kexas.Page
	err := p.WaitForFunction("true", 200*time.Millisecond)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "nil page") {
		t.Errorf("expected nil-page error; got %q", err.Error())
	}
}

func TestWaitForFunction_BelowTimeoutFloor_Errors(t *testing.T) {
	var p *kexas.Page
	err := p.WaitForFunction("true", 50*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestWaitForFunction_AcceptsArbitraryExpression_OnNilPage(t *testing.T) {
	// The expression is evaluated against the page; on nil page we never get
	// to evaluate, so this just tests the signature accepts strings cleanly.
	var p *kexas.Page
	err := p.WaitForFunction("window.myAppReady === true", 200*time.Millisecond)
	if err == nil {
		t.Error("expected nil-page error regardless of expression")
	}
}

// --- WaitForTitleContains ---

func TestWaitForTitleContains_NilPage_Errors(t *testing.T) {
	var p *kexas.Page
	err := p.WaitForTitleContains("kexas", 200*time.Millisecond)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWaitForTitleContains_BelowTimeoutFloor_Errors(t *testing.T) {
	var p *kexas.Page
	err := p.WaitForTitleContains("anything", 50*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestWaitForTitleContains_EmptySubstr_StillErrorsOnNilPage(t *testing.T) {
	// Empty substr is technically always-true (every string contains ""), but
	// nil-page guard fires first.
	var p *kexas.Page
	err := p.WaitForTitleContains("", 200*time.Millisecond)
	if err == nil {
		t.Error("expected nil-page error even with empty substr")
	}
}

// --- WaitForElementText ---

func TestWaitForElementText_NilPage_Errors(t *testing.T) {
	var p *kexas.Page
	err := p.WaitForElementText(".x", "marker", 200*time.Millisecond)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "nil page") {
		t.Errorf("expected nil-page error; got %q", err.Error())
	}
}

func TestWaitForElementText_BelowTimeoutFloor_Errors(t *testing.T) {
	var p *kexas.Page
	err := p.WaitForElementText(".x", "marker", 50*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestWaitForElementText_EmptySubstr_AcceptsAndErrorsOnNilPage(t *testing.T) {
	var p *kexas.Page
	err := p.WaitForElementText(".x", "", 200*time.Millisecond)
	if err == nil {
		t.Error("expected nil-page error")
	}
}
