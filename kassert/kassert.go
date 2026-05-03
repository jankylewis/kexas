// Package kassert provides fluent assertions for Kexas tests.
//
// Inspired by FluentAssertions (C#) and Playwright assertions,
// kassert provides a rich, readable assertion API with detailed
// error messages and chaining support.
//
// Basic usage:
//
//	kassert.That(t, actualValue).Equals(expectedValue)
//	kassert.That(t, text).Contains("substring")
//	kassert.That(t, err).IsNil()
//	kassert.That(t, list).HasLength(5)
//
// All assertions provide clear error messages on failure.
package kassert

import (
	"fmt"
	"reflect"
)

// TestingT is an interface that matches the subset of testing.T methods we need.
// This allows for easier testing of the assertion library itself.
type TestingT interface {
	Helper()
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
}

// Assertion provides fluent assertion methods.
type Assertion struct {
	t      TestingT
	actual interface{}
	name   string
}

// That creates a new assertion for the given value.
func That(t TestingT, actual interface{}) *Assertion {
	return &Assertion{
		t:      t,
		actual: actual,
		name:   "value",
	}
}

// Named sets a custom name for the value being asserted (for better error messages).
func (a *Assertion) Named(name string) *Assertion {
	a.name = name
	return a
}

// Equals asserts that the actual value equals the expected value.
func (a *Assertion) Equals(expected interface{}) *Assertion {
	if !reflect.DeepEqual(a.actual, expected) {
		a.t.Helper()
		a.t.Errorf("Expected %s to equal:\n  %v\nbut got:\n  %v", a.name, expected, a.actual)
		logFail(a.name, fmt.Sprintf("Equals(%v) — got %v", expected, a.actual))
		return a
	}
	logPass(a.name, fmt.Sprintf("Equals(%v)", expected))
	return a
}

// NotEquals asserts that the actual value does not equal the expected value.
func (a *Assertion) NotEquals(unexpected interface{}) *Assertion {
	if reflect.DeepEqual(a.actual, unexpected) {
		a.t.Helper()
		a.t.Errorf("Expected %s to not equal:\n  %v", a.name, unexpected)
		logFail(a.name, fmt.Sprintf("NotEquals(%v)", unexpected))
		return a
	}
	logPass(a.name, fmt.Sprintf("NotEquals(%v)", unexpected))
	return a
}

// IsNil asserts that the actual value is nil.
func (a *Assertion) IsNil() *Assertion {
	if a.actual != nil && !reflect.ValueOf(a.actual).IsNil() {
		a.t.Helper()
		a.t.Errorf("Expected %s to be nil, but got:\n  %v", a.name, a.actual)
		logFail(a.name, "IsNil")
		return a
	}
	logPass(a.name, "IsNil")
	return a
}

// IsNotNil asserts that the actual value is not nil.
func (a *Assertion) IsNotNil() *Assertion {
	if a.actual == nil {
		a.t.Helper()
		a.t.Errorf("Expected %s to not be nil", a.name)
		return a
	}

	var val reflect.Value = reflect.ValueOf(a.actual)
	if val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface ||
		val.Kind() == reflect.Slice || val.Kind() == reflect.Map ||
		val.Kind() == reflect.Chan || val.Kind() == reflect.Func {
		if val.IsNil() {
			a.t.Helper()
			a.t.Errorf("Expected %s to not be nil", a.name)
		}
	}
	return a
}

// IsTrue asserts that the actual value is true.
func (a *Assertion) IsTrue() *Assertion {
	var b bool
	var ok bool
	b, ok = a.actual.(bool)
	if !ok {
		a.t.Helper()
		a.t.Errorf("Expected %s to be a bool, but got type %T", a.name, a.actual)
		logFail(a.name, "IsTrue — not a bool")
		return a
	}
	if !b {
		a.t.Helper()
		a.t.Errorf("Expected %s to be true, but got false", a.name)
		logFail(a.name, "IsTrue")
		return a
	}
	logPass(a.name, "IsTrue")
	return a
}

// IsFalse asserts that the actual value is false.
func (a *Assertion) IsFalse() *Assertion {
	var b bool
	var ok bool
	b, ok = a.actual.(bool)
	if !ok {
		a.t.Helper()
		a.t.Errorf("Expected %s to be a bool, but got type %T", a.name, a.actual)
		logFail(a.name, "IsFalse — not a bool")
		return a
	}
	if b {
		a.t.Helper()
		a.t.Errorf("Expected %s to be false, but got true", a.name)
		logFail(a.name, "IsFalse")
		return a
	}
	logPass(a.name, "IsFalse")
	return a
}
