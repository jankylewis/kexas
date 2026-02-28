package kassert

import (
	"fmt"
	"strings"
)

// ErrorAssertion provides assertions for errors.
type ErrorAssertion struct {
	t   TestingT
	err error
}

// ThatError creates a new error assertion.
func ThatError(t TestingT, err error) *ErrorAssertion {
	return &ErrorAssertion{
		t:   t,
		err: err,
	}
}

// IsNil asserts that the error is nil.
func (e *ErrorAssertion) IsNil() *ErrorAssertion {
	if e.err != nil {
		e.t.Helper()
		e.t.Errorf("Expected no error, but got:\n  %v", e.err)
		logFail("error", fmt.Sprintf("IsNil — got %v", e.err))
	} else {
		logPass("error", "IsNil")
	}
	return e
}

// IsNotNil asserts that the error is not nil.
func (e *ErrorAssertion) IsNotNil() *ErrorAssertion {
	if e.err == nil {
		e.t.Helper()
		e.t.Error("Expected an error, but got nil")
		logFail("error", "IsNotNil — got nil")
	} else {
		logPass("error", "IsNotNil")
	}
	return e
}

// HasMessage asserts that the error message contains the expected text.
func (e *ErrorAssertion) HasMessage(expected string) *ErrorAssertion {
	if e.err == nil {
		e.t.Helper()
		e.t.Error("Expected an error with message, but got nil")
		return e
	}

	var msg string = e.err.Error()
	if !strings.Contains(msg, expected) {
		e.t.Helper()
		e.t.Errorf("Expected error message to contain:\n  %q\nbut got:\n  %q", expected, msg)
		logFail("error", fmt.Sprintf("HasMessage(%q)", expected))
	} else {
		logPass("error", fmt.Sprintf("HasMessage(%q)", expected))
	}
	return e
}

// Fail explicitly fails the test with a message.
func Fail(t TestingT, message string, args ...interface{}) {
	t.Helper()
	logFail("explicit", "Fail")
	if len(args) > 0 {
		t.Errorf(message, args...)
	} else {
		t.Error(message)
	}
}
