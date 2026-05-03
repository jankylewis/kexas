package kassert

import (
	"fmt"
	"strings"
)

// ErrorAssertion provides assertions for errors.
type ErrorAssertion struct {
	t    TestingT
	err  error
	name string
}

// ThatError creates a new error assertion.
func ThatError(t TestingT, err error) *ErrorAssertion {
	return &ErrorAssertion{
		t:    t,
		err:  err,
		name: "error",
	}
}

// Named sets a custom name for the error being asserted (for better error messages
// + structured-log labeling). Mirrors Assertion.Named so error-flavored assertions
// can be tagged the same way as value-flavored ones.
func (e *ErrorAssertion) Named(name string) *ErrorAssertion {
	e.name = name
	return e
}

// IsNil asserts that the error is nil.
func (e *ErrorAssertion) IsNil() *ErrorAssertion {
	if e.err != nil {
		e.t.Helper()
		e.t.Errorf("Expected %s to be nil, but got:\n  %v", e.name, e.err)
		logFail(e.name, fmt.Sprintf("IsNil — got %v", e.err))
		return e
	}
	logPass(e.name, "IsNil")
	return e
}

// IsNotNil asserts that the error is not nil.
func (e *ErrorAssertion) IsNotNil() *ErrorAssertion {
	if e.err == nil {
		e.t.Helper()
		e.t.Errorf("Expected %s to be non-nil, but got nil", e.name)
		logFail(e.name, "IsNotNil — got nil")
		return e
	}
	logPass(e.name, "IsNotNil")
	return e
}

// HasMessage asserts that the error message contains the expected text.
func (e *ErrorAssertion) HasMessage(expected string) *ErrorAssertion {
	if e.err == nil {
		e.t.Helper()
		e.t.Errorf("Expected %s to have message, but got nil", e.name)
		return e
	}

	var msg string = e.err.Error()
	if !strings.Contains(msg, expected) {
		e.t.Helper()
		e.t.Errorf("Expected %s message to contain:\n  %q\nbut got:\n  %q", e.name, expected, msg)
		logFail(e.name, fmt.Sprintf("HasMessage(%q)", expected))
		return e
	}
	logPass(e.name, fmt.Sprintf("HasMessage(%q)", expected))
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
