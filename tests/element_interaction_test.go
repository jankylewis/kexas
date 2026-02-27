package tests

import (
	"testing"
	"time"

	"github.com/kexas-project/kexas"
	kexaserrors "github.com/kexas-project/kexas/errors"
	"github.com/kexas-project/kexas/internal/cdp"
)

// --- Hover tests (previously zero coverage) ---

func TestElement_Hover_NoPage(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#menu-item", 123, 10*time.Second)

	var err error = element.Hover()

	if err == nil {
		t.Fatal("expected error for element with no page")
	}

	if err != kexaserrors.ErrElementNoPage {
		t.Errorf("expected ErrElementNoPage, got: %v", err)
	}
}

func TestElement_WaitAndHover_NoPage(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#dropdown", 456, 5*time.Second)

	var err error = element.WaitAndHover()

	if err == nil {
		t.Fatal("expected error for element with no page")
	}

	if err != kexaserrors.ErrElementNoPage {
		t.Errorf("expected ErrElementNoPage, got: %v", err)
	}
}

func TestElement_WaitAndHoverFor_InvalidTimeout(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#dropdown", 456, 5*time.Second)

	var err error = element.WaitAndHoverFor(500 * time.Millisecond)

	if err == nil {
		t.Fatal("expected error for timeout < 1s")
	}

	var expected string = "timeout must be at least 1 second, got 500ms"
	if err.Error() != expected {
		t.Errorf("expected '%s', got '%s'", expected, err.Error())
	}
}

func TestElement_WaitAndHoverFor_ZeroTimeout(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#dropdown", 456, 5*time.Second)

	var err error = element.WaitAndHoverFor(0)

	if err == nil {
		t.Fatal("expected error for zero timeout")
	}

	var expected string = "timeout must be at least 1 second, got 0s"
	if err.Error() != expected {
		t.Errorf("expected '%s', got '%s'", expected, err.Error())
	}
}

func TestElement_WaitAndHoverFor_NoPage(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#dropdown", 456, 5*time.Second)

	var err error = element.WaitAndHoverFor(3 * time.Second)

	if err == nil {
		t.Fatal("expected error for element with no page")
	}

	if err != kexaserrors.ErrElementNoPage {
		t.Errorf("expected ErrElementNoPage, got: %v", err)
	}
}

// --- WaitAndClick tests (previously zero coverage) ---

func TestElement_WaitAndClick_NoPage(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#submit-btn", 789, 5*time.Second)

	var err error = element.WaitAndClick()

	if err == nil {
		t.Fatal("expected error for element with no page")
	}

	if err != kexaserrors.ErrElementNoPage {
		t.Errorf("expected ErrElementNoPage, got: %v", err)
	}
}

// --- WaitAndType tests (previously zero coverage) ---

func TestElement_WaitAndType_NoPage(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#username", 100, 5*time.Second)

	var err error = element.WaitAndType("test_user")

	if err == nil {
		t.Fatal("expected error for element with no page")
	}

	if err != kexaserrors.ErrElementNoPage {
		t.Errorf("expected ErrElementNoPage, got: %v", err)
	}
}

// --- Type edge cases ---

func TestElement_Type_EmptyText_NilPage(t *testing.T) {
	// With nil page, page check fires before text empty check
	var element *kexas.Element = kexas.NewElement(nil, "#input", 100, 5*time.Second)

	var err error = element.Type("")

	if err == nil {
		t.Fatal("expected error")
	}

	if err != kexaserrors.ErrElementNoPage {
		t.Errorf("expected ErrElementNoPage (fires before ErrTextEmpty), got: %v", err)
	}
}

func TestElement_WaitAndTypeFor_EmptyText_NoPage(t *testing.T) {
	// Empty text with nil page: page check fires before text check
	var element *kexas.Element = kexas.NewElement(nil, "#input", 100, 5*time.Second)

	var err error = element.WaitAndTypeFor("", 3*time.Second)

	if err == nil {
		t.Fatal("expected error")
	}

	// page nil check fires before text empty check
	if err != kexaserrors.ErrElementNoPage {
		t.Errorf("expected ErrElementNoPage, got: %v", err)
	}
}

// --- Invalid nodeID / objectID guard tests ---

func TestElement_Click_InvalidNodeID_NoObjectID(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#btn", 0, 5*time.Second)

	var err error = element.Click()

	if err == nil {
		t.Fatal("expected error")
	}

	// nil page check fires before nodeID check
	if err != kexaserrors.ErrElementNoPage {
		t.Errorf("expected ErrElementNoPage, got: %v", err)
	}
}

func TestElement_Hover_InvalidNodeID_NoObjectID(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#item", 0, 5*time.Second)

	var err error = element.Hover()

	if err == nil {
		t.Fatal("expected error")
	}

	if err != kexaserrors.ErrElementNoPage {
		t.Errorf("expected ErrElementNoPage, got: %v", err)
	}
}

func TestElement_Type_InvalidNodeID_NoObjectID(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#field", 0, 5*time.Second)

	var err error = element.Type("hello")

	if err == nil {
		t.Fatal("expected error")
	}

	if err != kexaserrors.ErrElementNoPage {
		t.Errorf("expected ErrElementNoPage, got: %v", err)
	}
}

// --- NewElementWithObject tests ---

func TestElement_NewElementWithObject(t *testing.T) {
	var page *kexas.Page = nil
	var selector string = "div.container"
	var nodeID cdp.NodeID = 999
	var objectID string = "objectId-12345"
	var timeout time.Duration = 8 * time.Second

	var element *kexas.Element = kexas.NewElementWithObject(page, selector, nodeID, objectID, timeout)

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
}

func TestElement_NewElementWithObject_Click_NoPage(t *testing.T) {
	var element *kexas.Element = kexas.NewElementWithObject(nil, "button.submit", 100, "obj-abc", 5*time.Second)

	var err error = element.Click()

	if err == nil {
		t.Fatal("expected error for element with no page")
	}

	if err != kexaserrors.ErrElementNoPage {
		t.Errorf("expected ErrElementNoPage, got: %v", err)
	}
}

func TestElement_NewElementWithObject_Type_NoPage(t *testing.T) {
	var element *kexas.Element = kexas.NewElementWithObject(nil, "input.email", 200, "obj-xyz", 5*time.Second)

	var err error = element.Type("test@example.com")

	if err == nil {
		t.Fatal("expected error for element with no page")
	}

	if err != kexaserrors.ErrElementNoPage {
		t.Errorf("expected ErrElementNoPage, got: %v", err)
	}
}

func TestElement_NewElementWithObject_Hover_NoPage(t *testing.T) {
	var element *kexas.Element = kexas.NewElementWithObject(nil, "nav.menu", 300, "obj-nav", 5*time.Second)

	var err error = element.Hover()

	if err == nil {
		t.Fatal("expected error for element with no page")
	}

	if err != kexaserrors.ErrElementNoPage {
		t.Errorf("expected ErrElementNoPage, got: %v", err)
	}
}

// --- WaitAndClickFor timeout validation ---

func TestElement_WaitAndClickFor_NegativeTimeout(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#btn", 123, 5*time.Second)

	var err error = element.WaitAndClickFor(-1 * time.Second)

	if err == nil {
		t.Fatal("expected error for negative timeout")
	}
}

// --- WaitAndTypeFor timeout validation ---

func TestElement_WaitAndTypeFor_NegativeTimeout(t *testing.T) {
	var element *kexas.Element = kexas.NewElement(nil, "#input", 123, 5*time.Second)

	var err error = element.WaitAndTypeFor("text", -1*time.Second)

	if err == nil {
		t.Fatal("expected error for negative timeout")
	}
}
