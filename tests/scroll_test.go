
package tests

import (
	"testing"
	"time"

	"github.com/kexas-project/kexas"
	kexaserrors "github.com/kexas-project/kexas/errors"
)

// --- Page.ScrollToTop tests ---

func TestPage_ScrollToTop_NilPage(t *testing.T) {
	var page *kexas.Page = nil

	var err error = page.ScrollToTop()

	if err == nil {
		t.Fatal("expected error for nil page")
	}

	if err != kexaserrors.ErrPageNil {
		t.Errorf("expected ErrPageNil, got: %v", err)
	}
}

// --- Page.ScrollToBottom tests ---

func TestPage_ScrollToBottom_NilPage(t *testing.T) {
	var page *kexas.Page = nil

	var err error = page.ScrollToBottom()

	if err == nil {
		t.Fatal("expected error for nil page")
	}

	if err != kexaserrors.ErrPageNil {
		t.Errorf("expected ErrPageNil, got: %v", err)
	}
}

// --- Page.ScrollBy tests ---

func TestPage_ScrollBy_NilPage(t *testing.T) {
	var page *kexas.Page = nil

	var err error = page.ScrollBy(0, 500)

	if err == nil {
		t.Fatal("expected error for nil page")
	}

	if err != kexaserrors.ErrPageNil {
		t.Errorf("expected ErrPageNil, got: %v", err)
	}
}

func TestPage_ScrollBy_ZeroPixels_NilPage(t *testing.T) {
	// Nil page check fires before zero pixels check
	var page *kexas.Page = nil

	var err error = page.ScrollBy(0, 0)

	if err == nil {
		t.Fatal("expected error")
	}

	if err != kexaserrors.ErrPageNil {
		t.Errorf("expected ErrPageNil (fires before zero pixels check), got: %v", err)
	}
}

func TestPage_ScrollBy_NegativeY(t *testing.T) {
	// Negative y is valid (scroll up), but nil page should still error
	var page *kexas.Page = nil

	var err error = page.ScrollBy(0, -200)

	if err == nil {
		t.Fatal("expected error for nil page")
	}

	if err != kexaserrors.ErrPageNil {
		t.Errorf("expected ErrPageNil, got: %v", err)
	}
}

func TestPage_ScrollBy_PositiveX(t *testing.T) {
	// Positive x is valid (scroll right), but nil page should still error
	var page *kexas.Page = nil

	var err error = page.ScrollBy(300, 0)

	if err == nil {
		t.Fatal("expected error for nil page")
	}

	if err != kexaserrors.ErrPageNil {
		t.Errorf("expected ErrPageNil, got: %v", err)
	}
}

// --- Page.ScrollPosition tests ---

func TestPage_ScrollPosition_NilPage(t *testing.T) {
	var page *kexas.Page = nil

	var x int
	var y int
	var err error
	x, y, err = page.ScrollPosition()

	if err == nil {
		t.Fatal("expected error for nil page")
	}

	if err != kexaserrors.ErrPageNil {
		t.Errorf("expected ErrPageNil, got: %v", err)
	}

	if x != 0 || y != 0 {
		t.Errorf("expected (0, 0) for nil page, got (%d, %d)", x, y)
	}
}

// --- Element.ScrollIntoView tests ---

func TestElement_ScrollIntoView_NilElement(t *testing.T) {
	var element *kexas.Element = nil

	var err error = element.ScrollIntoView()

	if err == nil {
		t.Fatal("expected error for nil element")
	}

	if err != kexaserrors.ErrElementNil {
		t.Errorf("expected ErrElementNil, got: %v", err)
	}
}

func TestElement_ScrollIntoView_NoPage(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#target", 123, 5*time.Second)

	var err error = element.ScrollIntoView()

	if err == nil {
		t.Fatal("expected error for element with no page")
	}

	if err != kexaserrors.ErrElementNoPage {
		t.Errorf("expected ErrElementNoPage, got: %v", err)
	}
}

func TestElement_ScrollIntoView_InvalidNodeID(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#target", 0, 5*time.Second)

	var err error = element.ScrollIntoView()

	if err == nil {
		t.Fatal("expected error")
	}

	// nil page fires before nodeID check
	if err != kexaserrors.ErrElementNoPage {
		t.Errorf("expected ErrElementNoPage, got: %v", err)
	}
}

func TestElement_ScrollIntoView_WithObject_NoPage(t *testing.T) {
	var element *kexas.Element = kexas.NewElementWithObject(nil, "div.item", 100, "obj-123", 5*time.Second)

	var err error = element.ScrollIntoView()

	if err == nil {
		t.Fatal("expected error for element with no page")
	}

	if err != kexaserrors.ErrElementNoPage {
		t.Errorf("expected ErrElementNoPage, got: %v", err)
	}
}
