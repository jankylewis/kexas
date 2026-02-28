// Package errors provides custom error types for Kexas.
//
// All Kexas-specific errors are defined here, including sentinel
// errors (ErrXxx variables) and structured error types (XxxError structs)
// that carry contextual information for debugging.
package errors

import (
	"fmt"
	"time"
)

// Sentinel errors for common Kexas operations
var (
	// ErrElementNotFound indicates an element could not be found
	ErrElementNotFound = fmt.Errorf("element not found")

	// ErrElementNotVisible indicates an element exists but is not visible
	ErrElementNotVisible = fmt.Errorf("element not visible")

	// ErrElementNotClickable indicates an element is not clickable
	ErrElementNotClickable = fmt.Errorf("element not clickable")

	// ErrTimeout indicates an operation timed out
	ErrTimeout = fmt.Errorf("operation timed out")

	// ErrInvalidSelector indicates a selector is invalid
	ErrInvalidSelector = fmt.Errorf("invalid selector")

	// ErrInvalidTimeout indicates a timeout value is invalid
	ErrInvalidTimeout = fmt.Errorf("invalid timeout")

	// ErrElementNotAttached indicates an element is no longer attached to the DOM
	ErrElementNotAttached = fmt.Errorf("element not attached to DOM")

	// ErrBrowserNotConnected indicates browser connection issues
	ErrBrowserNotConnected = fmt.Errorf("browser not connected")

	// Common element state errors (used frequently across codebase)
	ErrElementNoPage        = fmt.Errorf("element has no associated page")
	ErrElementNil           = fmt.Errorf("element is nil")
	ErrElementInvalidNodeID = fmt.Errorf("element has invalid node ID")

	// Common page state errors
	ErrPageNil            = fmt.Errorf("page is nil")
	ErrPageNoBrowser      = fmt.Errorf("page has no associated browser")
	ErrPageInvalidSession = fmt.Errorf("page has invalid session")

	// Common selector errors
	ErrSelectorEmpty       = fmt.Errorf("selector cannot be empty")
	ErrSelectorUnsupported = fmt.Errorf("selector type is not supported")

	// Common attribute and text errors
	ErrAttributeEmpty    = fmt.Errorf("attribute name cannot be empty")
	ErrAttributeNotFound = fmt.Errorf("attribute not found")
	ErrTextEmpty         = fmt.Errorf("text cannot be empty")

	// Common CDP errors
	ErrCDPCommandFailed   = fmt.Errorf("CDP command failed")
	ErrCDPResponseInvalid = fmt.Errorf("CDP response is invalid")
	ErrCDPConnectionLost  = fmt.Errorf("CDP connection lost")

	// Common operation errors (frequently used across codebase)
	ErrFailedToClickElement    = fmt.Errorf("failed to click element")
	ErrClickOperationFailed    = fmt.Errorf("click operation failed")
	ErrFailedToTypeIntoElement = fmt.Errorf("failed to type into element")
	ErrTypeOperationFailed     = fmt.Errorf("type operation failed")
	ErrFailedToHoverElement    = fmt.Errorf("failed to hover element")
	ErrHoverOperationFailed    = fmt.Errorf("hover operation failed")
	ErrScrollOperationFailed   = fmt.Errorf("scroll operation failed")
	ErrScrollInvalidPixels     = fmt.Errorf("scroll pixels must not be zero")

	// Common timeout validation error
	ErrTimeoutInvalid = fmt.Errorf("timeout must be at least 1 second")

	// Agent management errors
	ErrAgentNotEnabled      = fmt.Errorf("agent not enabled")
	ErrAgentEnableFailed    = fmt.Errorf("failed to enable agent")
	ErrAgentDisableFailed   = fmt.Errorf("failed to disable agent")
	ErrAgentNotFound        = fmt.Errorf("agent not found")
	ErrAgentContextNotFound = fmt.Errorf("agent context not found")
	ErrAgentNotReady        = fmt.Errorf("agent not ready")

	// Cookie management errors
	ErrCookieNameEmpty    = fmt.Errorf("cookie name cannot be empty")
	ErrCookieValueEmpty   = fmt.Errorf("cookie value cannot be empty")
	ErrCookieSetFailed    = fmt.Errorf("failed to set cookie")
	ErrCookieGetFailed    = fmt.Errorf("failed to get cookies")
	ErrCookieDeleteFailed = fmt.Errorf("failed to delete cookie")
	ErrCookieNotFound     = fmt.Errorf("cookie not found")

	// Storage errors
	ErrStorageKeyEmpty     = fmt.Errorf("storage key cannot be empty")
	ErrStorageSetFailed    = fmt.Errorf("failed to set storage item")
	ErrStorageGetFailed    = fmt.Errorf("failed to get storage item")
	ErrStorageRemoveFailed = fmt.Errorf("failed to remove storage item")
	ErrStorageClearFailed  = fmt.Errorf("failed to clear storage")

	// Multi-tab errors
	ErrPageAlreadyClosed     = fmt.Errorf("page is already closed")
	ErrPageIndexOutOfRange   = fmt.Errorf("page index out of range")
	ErrNoPageMatchingURL     = fmt.Errorf("no page matching URL pattern")
	ErrWaitForNewPageTimeout = fmt.Errorf("timed out waiting for new page")

	// Recorder errors
	ErrRecorderNotStarted     = fmt.Errorf("recorder not started")
	ErrRecorderAlreadyStarted = fmt.Errorf("recorder already started")
	ErrFfmpegNotFound         = fmt.Errorf("ffmpeg not found")
)

// ElementError represents an error related to element operations
type ElementError struct {
	Selector  string
	Operation string
	Cause     error
}

func (e *ElementError) Error() string {
	return fmt.Sprintf("element operation '%s' failed for selector '%s': %v",
		e.Operation, e.Selector, e.Cause)
}

func (e *ElementError) Unwrap() error {
	return e.Cause
}

// TimeoutError represents a timeout error with context
type TimeoutError struct {
	Operation string
	Timeout   time.Duration
	Cause     error
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("operation '%s' timed out after %v: %v",
		e.Operation, e.Timeout, e.Cause)
}

func (e *TimeoutError) Unwrap() error {
	return e.Cause
}

// SelectorError represents an error related to selector operations
type SelectorError struct {
	Selector string
	Type     string // "ID", "CSS", "XPath"
	Cause    error
}

func (e *SelectorError) Error() string {
	return fmt.Sprintf("selector error for %s selector '%s': %v",
		e.Type, e.Selector, e.Cause)
}

func (e *SelectorError) Unwrap() error {
	return e.Cause
}

// ValidationError represents an element validation error
type ValidationError struct {
	Selector string
	Reason   string
	Cause    error
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("element validation failed for selector '%s': %s",
		e.Selector, e.Reason)
}

func (e *ValidationError) Unwrap() error {
	return e.Cause
}

// NewElementError creates a new ElementError
func NewElementError(selector, operation string, cause error) *ElementError {
	return &ElementError{
		Selector:  selector,
		Operation: operation,
		Cause:     cause,
	}
}

// NewTimeoutError creates a new TimeoutError
func NewTimeoutError(operation string, timeout time.Duration, cause error) *TimeoutError {
	return &TimeoutError{
		Operation: operation,
		Timeout:   timeout,
		Cause:     cause,
	}
}

// NewSelectorError creates a new SelectorError
func NewSelectorError(selector, selectorType string, cause error) *SelectorError {
	return &SelectorError{
		Selector: selector,
		Type:     selectorType,
		Cause:    cause,
	}
}

// NewValidationError creates a new ValidationError
func NewValidationError(selector, reason string, cause error) *ValidationError {
	return &ValidationError{
		Selector: selector,
		Reason:   reason,
		Cause:    cause,
	}
}

// Helper functions for common formatted errors

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
