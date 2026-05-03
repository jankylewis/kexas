// Package errors provides custom error types for Kexas.
//
// All Kexas-specific errors are defined here, including sentinel
// errors (ErrXxx variables) and structured error types (XxxError structs)
// that carry contextual information for debugging.
package errors

import (
	"fmt"
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
