package errors_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	kerrors "github.com/jankylewis/kexas/errors"
)

// 15 critical-after-existing unit tests for the errors package.
// Existing tests cover: sentinels, structured types, basic helpers, Is/As wraps.
// This batch adds: less-covered formatted helpers, edge cases (zero/negative
// timeouts, empty selectors), structured-type edge cases, and contract-by-
// example tests for cross-helper consistency.

func TestAgentDisableFailed_IncludesAgentName(t *testing.T) {
	var err error = kerrors.AgentDisableFailed("Network", errors.New("dropped"))
	if !strings.Contains(err.Error(), "Network") {
		t.Errorf("missing agent name; got %q", err.Error())
	}
}

func TestAgentDisableFailed_WrapsCause(t *testing.T) {
	var cause error = errors.New("ws drop")
	if !errors.Is(kerrors.AgentDisableFailed("DOM", cause), cause) {
		t.Error("AgentDisableFailed must wrap cause via %w")
	}
}

func TestAgentNotFound_IncludesAgentName(t *testing.T) {
	var err error = kerrors.AgentNotFound("Fetch")
	if !strings.Contains(err.Error(), "Fetch") {
		t.Errorf("missing agent name; got %q", err.Error())
	}
}

func TestAgentContextNotFound_IncludesAgentName(t *testing.T) {
	var err error = kerrors.AgentContextNotFound("Profiler")
	if !strings.Contains(err.Error(), "Profiler") {
		t.Errorf("missing agent name; got %q", err.Error())
	}
}

func TestAgentNotReady_WrapsCauseAndNamesAgent(t *testing.T) {
	var cause error = errors.New("init pending")
	var err error = kerrors.AgentNotReady("Page", cause)
	if !errors.Is(err, cause) {
		t.Error("must wrap cause")
	}
	if !strings.Contains(err.Error(), "Page") {
		t.Errorf("missing agent name; got %q", err.Error())
	}
}

func TestElementNotVisibleWrap_WrapsCause(t *testing.T) {
	var cause error = errors.New("display:none")
	if !errors.Is(kerrors.ElementNotVisibleWrap(cause), cause) {
		t.Error("must wrap cause")
	}
}

func TestElementNotVisibleWithinWrap_IncludesTimeoutAndWrapsCause(t *testing.T) {
	var cause error = errors.New("polled out")
	var err error = kerrors.ElementNotVisibleWithinWrap(7*time.Second, cause)
	if !errors.Is(err, cause) {
		t.Error("must wrap cause")
	}
	if !strings.Contains(err.Error(), "7s") {
		t.Errorf("missing timeout; got %q", err.Error())
	}
}

func TestElementNotFound_HandlesEmptySelector(t *testing.T) {
	var err error = kerrors.ElementNotFound("")
	if err == nil {
		t.Fatal("nil err for empty selector")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' phrase; got %q", err.Error())
	}
}

func TestElementNotVisibleWithin_HandlesZeroTimeout(t *testing.T) {
	var err error = kerrors.ElementNotVisibleWithin("#x", time.Duration(0))
	if !strings.Contains(err.Error(), "0s") {
		t.Errorf("expected zero-duration formatting; got %q", err.Error())
	}
}

func TestNewElementError_ExposesAllFields(t *testing.T) {
	var cause error = errors.New("ko")
	var ee *kerrors.ElementError = kerrors.NewElementError("#login", "click", cause)
	if ee == nil {
		t.Fatal("nil ElementError")
	}
	if ee.Selector != "#login" || ee.Operation != "click" {
		t.Errorf("fields not preserved: selector=%q op=%q", ee.Selector, ee.Operation)
	}
	if !errors.Is(ee, cause) {
		t.Error("must wrap cause via Unwrap")
	}
}

func TestNewTimeoutError_ExposesAllFields(t *testing.T) {
	var cause error = errors.New("deadline")
	var te *kerrors.TimeoutError = kerrors.NewTimeoutError("page-load", 5*time.Second, cause)
	if te.Operation != "page-load" {
		t.Errorf("operation: got %q", te.Operation)
	}
	if te.Timeout != 5*time.Second {
		t.Errorf("timeout: got %v", te.Timeout)
	}
	if !errors.Is(te, cause) {
		t.Error("must wrap cause")
	}
}

func TestNewSelectorError_ExposesAllFields(t *testing.T) {
	var cause error = errors.New("bad")
	var se *kerrors.SelectorError = kerrors.NewSelectorError("//p[bad", "xpath", cause)
	if se.Selector != "//p[bad" || se.Type != "xpath" {
		t.Errorf("fields not preserved: %+v", *se)
	}
	if !errors.Is(se, cause) {
		t.Error("must wrap cause")
	}
}

func TestNewValidationError_ExposesAllFields(t *testing.T) {
	var ve *kerrors.ValidationError = kerrors.NewValidationError("#x", "must be button", nil)
	if ve.Reason != "must be button" {
		t.Errorf("reason: got %q", ve.Reason)
	}
	// nil cause should be allowed
	if ve.Error() == "" {
		t.Error("Error() should produce a message even with nil cause")
	}
}

func TestSentinelErrors_DifferentInstancesNotEqual(t *testing.T) {
	// Two distinct sentinels, even with similar messages, must differ.
	if errors.Is(kerrors.ErrElementNotFound, kerrors.ErrElementNoPage) {
		t.Error("distinct sentinels collapsed via errors.Is")
	}
}

func TestStructuredErrors_ChainPreservedAcrossWrap(t *testing.T) {
	// Error chain via two layers of wrap should still resolve to root cause.
	var root error = errors.New("network drop")
	var mid error = kerrors.AgentEnableFailed("Network", root)
	var top error = kerrors.AgentNotReady("Network", mid)
	if !errors.Is(top, root) {
		t.Error("multi-layer wrap chain must preserve root cause via errors.Is")
	}
}
