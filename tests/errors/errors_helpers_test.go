package errors_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	kerrors "github.com/jankylewis/kexas/errors"
)

// Medium-risk gap-fill tests for the formatted-error helpers in
// errors/errors_helpers.go. These wrap fmt.Errorf with templates; the tests
// verify the format strings include their args (so error messages remain
// useful for debugging).

func TestElementNotFound_IncludesSelector(t *testing.T) {
	var err error = kerrors.ElementNotFound("#login-button")
	if !strings.Contains(err.Error(), "#login-button") {
		t.Errorf("expected selector in message, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' phrase, got %q", err.Error())
	}
}

func TestElementNotVisibleWithin_IncludesSelectorAndTimeout(t *testing.T) {
	var err error = kerrors.ElementNotVisibleWithin(".spinner", 7*time.Second)
	var msg string = err.Error()
	if !strings.Contains(msg, ".spinner") {
		t.Errorf("expected selector in message, got %q", msg)
	}
	if !strings.Contains(msg, "7s") {
		t.Errorf("expected timeout in message, got %q", msg)
	}
}

func TestElementNotClickableWithin_IncludesSelectorAndTimeout(t *testing.T) {
	var err error = kerrors.ElementNotClickableWithin("button[type=submit]", 3*time.Second)
	var msg string = err.Error()
	if !strings.Contains(msg, "button[type=submit]") {
		t.Errorf("expected selector in message, got %q", msg)
	}
	if !strings.Contains(msg, "3s") {
		t.Errorf("expected timeout in message, got %q", msg)
	}
}

func TestTimeoutInvalidFormat_IncludesDuration(t *testing.T) {
	var err error = kerrors.TimeoutInvalidFormat(0)
	var msg string = err.Error()
	if !strings.Contains(msg, "timeout") {
		t.Errorf("expected 'timeout' in message, got %q", msg)
	}
	if !strings.Contains(msg, "0") {
		t.Errorf("expected duration value in message, got %q", msg)
	}
}

func TestURLDidNotContainWithin_IncludesSubstrAndTimeout(t *testing.T) {
	var err error = kerrors.URLDidNotContainWithin("/dashboard", 10*time.Second)
	var msg string = err.Error()
	if !strings.Contains(msg, "/dashboard") {
		t.Errorf("expected substring in message, got %q", msg)
	}
	if !strings.Contains(msg, "10s") {
		t.Errorf("expected timeout in message, got %q", msg)
	}
}

func TestAgentEnableFailed_WrapsCause(t *testing.T) {
	var cause error = errors.New("websocket closed")
	var err error = kerrors.AgentEnableFailed("DOM", cause)
	var msg string = err.Error()
	if !strings.Contains(msg, "DOM") {
		t.Errorf("expected agent name in message, got %q", msg)
	}
	if !errors.Is(err, cause) {
		t.Error("AgentEnableFailed should wrap the cause via %w")
	}
}

func TestResolveElementFailed_WrapsCause(t *testing.T) {
	var cause error = errors.New("nodeID 0 not resolvable")
	var err error = kerrors.ResolveElementFailed(cause)
	if !errors.Is(err, cause) {
		t.Error("ResolveElementFailed should wrap the cause via %w")
	}
}

func TestClickElementFailed_WrapsCause(t *testing.T) {
	var cause error = errors.New("element not in viewport")
	var err error = kerrors.ClickElementFailed(cause)
	if !errors.Is(err, cause) {
		t.Error("ClickElementFailed should wrap the cause via %w")
	}
}

func TestHoverElementFailed_WrapsCause(t *testing.T) {
	var cause error = errors.New("invalid mouse position")
	var err error = kerrors.HoverElementFailed(cause)
	if !errors.Is(err, cause) {
		t.Error("HoverElementFailed should wrap the cause via %w")
	}
}

func TestFocusElementFailed_WrapsCause(t *testing.T) {
	var cause error = errors.New("not a focusable element")
	var err error = kerrors.FocusElementFailed(cause)
	if !errors.Is(err, cause) {
		t.Error("FocusElementFailed should wrap the cause via %w")
	}
}
