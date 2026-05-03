package errors

import (
	"fmt"
)

// Helper functions for common formatted errors.
// These wrap fmt.Errorf with consistent message templates so call sites stay terse.

// ElementNotFound creates a formatted "element not found" error
func ElementNotFound(selector string) error {
	return fmt.Errorf("element %s not found", selector)
}

// ElementNotVisibleWithin creates a formatted "element not visible within timeout" error
func ElementNotVisibleWithin(selector string, timeout interface{}) error {
	return fmt.Errorf("element %s not visible within %v", selector, timeout)
}

// ElementNotClickableWithin creates a formatted "element not clickable within timeout" error
func ElementNotClickableWithin(selector string, timeout interface{}) error {
	return fmt.Errorf("element %s not clickable within %v", selector, timeout)
}

// TimeoutInvalidFormat creates a formatted timeout validation error
func TimeoutInvalidFormat(duration interface{}) error {
	return fmt.Errorf("timeout must be at least 1 second, got %v", duration)
}

// URLDidNotContainWithin creates a formatted "URL did not contain X within timeout" error
func URLDidNotContainWithin(substr string, timeout interface{}) error {
	return fmt.Errorf("URL did not contain %q within %v", substr, timeout)
}

// AgentEnableFailed creates a formatted agent enablement failure error
func AgentEnableFailed(agentName string, cause error) error {
	return fmt.Errorf("failed to enable %s agent: %w", agentName, cause)
}

// AgentDisableFailed creates a formatted agent disablement failure error
func AgentDisableFailed(agentName string, cause error) error {
	return fmt.Errorf("failed to disable %s agent: %w", agentName, cause)
}

// AgentNotFound creates a formatted "agent not found" error
func AgentNotFound(agentName string) error {
	return fmt.Errorf("agent %s not found", agentName)
}

// AgentContextNotFound creates a formatted "agent context not found" error
func AgentContextNotFound(agentName string) error {
	return fmt.Errorf("agent context not found for %s", agentName)
}

// AgentNotReady creates a formatted "agent not ready" error
func AgentNotReady(agentName string, cause error) error {
	return fmt.Errorf("agent %s not ready: %w", agentName, cause)
}

// ResolveElementFailed creates a formatted "failed to resolve element" error
func ResolveElementFailed(cause error) error {
	return fmt.Errorf("failed to resolve element: %w", cause)
}

// ClickElementFailed creates a formatted "failed to click element" error
func ClickElementFailed(cause error) error {
	return fmt.Errorf("failed to click element: %w", cause)
}

// HoverElementFailed creates a formatted "failed to hover over element" error
func HoverElementFailed(cause error) error {
	return fmt.Errorf("failed to hover over element: %w", cause)
}

// ElementNotVisibleWrap creates a wrapping "element not visible" error
func ElementNotVisibleWrap(cause error) error {
	return fmt.Errorf("element not visible: %w", cause)
}

// ElementNotVisibleWithinWrap creates a wrapping "element not visible within timeout" error
func ElementNotVisibleWithinWrap(timeout interface{}, cause error) error {
	return fmt.Errorf("element not visible within %v: %w", timeout, cause)
}

// FocusElementFailed creates a formatted "failed to focus element" error
func FocusElementFailed(cause error) error {
	return fmt.Errorf("failed to focus element: %w", cause)
}
