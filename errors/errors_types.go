package errors

import (
	"fmt"
	"time"
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
