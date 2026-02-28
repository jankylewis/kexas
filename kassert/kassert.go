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
	"strings"
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
	} else {
		logPass(a.name, fmt.Sprintf("Equals(%v)", expected))
	}
	return a
}

// NotEquals asserts that the actual value does not equal the expected value.
func (a *Assertion) NotEquals(unexpected interface{}) *Assertion {
	if reflect.DeepEqual(a.actual, unexpected) {
		a.t.Helper()
		a.t.Errorf("Expected %s to not equal:\n  %v", a.name, unexpected)
		logFail(a.name, fmt.Sprintf("NotEquals(%v)", unexpected))
	} else {
		logPass(a.name, fmt.Sprintf("NotEquals(%v)", unexpected))
	}
	return a
}

// IsNil asserts that the actual value is nil.
func (a *Assertion) IsNil() *Assertion {
	if a.actual != nil && !reflect.ValueOf(a.actual).IsNil() {
		a.t.Helper()
		a.t.Errorf("Expected %s to be nil, but got:\n  %v", a.name, a.actual)
		logFail(a.name, "IsNil")
	} else {
		logPass(a.name, "IsNil")
	}
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
	} else {
		logPass(a.name, "IsTrue")
	}
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
	} else {
		logPass(a.name, "IsFalse")
	}
	return a
}

// Contains asserts that a string contains a substring.
func (a *Assertion) Contains(substring string) *Assertion {
	var str string
	var ok bool
	str, ok = a.actual.(string)
	if !ok {
		a.t.Helper()
		a.t.Errorf("Expected %s to be a string, but got type %T", a.name, a.actual)
		logFail(a.name, "Contains — not a string")
		return a
	}
	if !strings.Contains(str, substring) {
		a.t.Helper()
		a.t.Errorf("Expected %s to contain:\n  %q\nbut got:\n  %q", a.name, substring, str)
		logFail(a.name, fmt.Sprintf("Contains(%q)", substring))
	} else {
		logPass(a.name, fmt.Sprintf("Contains(%q)", substring))
	}
	return a
}

// NotContains asserts that a string does not contain a substring.
func (a *Assertion) NotContains(substring string) *Assertion {
	var str string
	var ok bool
	str, ok = a.actual.(string)
	if !ok {
		a.t.Helper()
		a.t.Errorf("Expected %s to be a string, but got type %T", a.name, a.actual)
		logFail(a.name, "NotContains — not a string")
		return a
	}
	if strings.Contains(str, substring) {
		a.t.Helper()
		a.t.Errorf("Expected %s to not contain:\n  %q\nbut it does:\n  %q", a.name, substring, str)
		logFail(a.name, fmt.Sprintf("NotContains(%q)", substring))
	} else {
		logPass(a.name, fmt.Sprintf("NotContains(%q)", substring))
	}
	return a
}

// StartsWith asserts that a string starts with a prefix.
func (a *Assertion) StartsWith(prefix string) *Assertion {
	var str string
	var ok bool
	str, ok = a.actual.(string)
	if !ok {
		a.t.Helper()
		a.t.Errorf("Expected %s to be a string, but got type %T", a.name, a.actual)
		return a
	}
	if !strings.HasPrefix(str, prefix) {
		a.t.Helper()
		a.t.Errorf("Expected %s to start with:\n  %q\nbut got:\n  %q", a.name, prefix, str)
		logFail(a.name, fmt.Sprintf("StartsWith(%q)", prefix))
	} else {
		logPass(a.name, fmt.Sprintf("StartsWith(%q)", prefix))
	}
	return a
}

// EndsWith asserts that a string ends with a suffix.
func (a *Assertion) EndsWith(suffix string) *Assertion {
	var str string
	var ok bool
	str, ok = a.actual.(string)
	if !ok {
		a.t.Helper()
		a.t.Errorf("Expected %s to be a string, but got type %T", a.name, a.actual)
		return a
	}
	if !strings.HasSuffix(str, suffix) {
		a.t.Helper()
		a.t.Errorf("Expected %s to end with:\n  %q\nbut got:\n  %q", a.name, suffix, str)
		logFail(a.name, fmt.Sprintf("EndsWith(%q)", suffix))
	} else {
		logPass(a.name, fmt.Sprintf("EndsWith(%q)", suffix))
	}
	return a
}

// IsEmpty asserts that a string, slice, map, or array is empty.
func (a *Assertion) IsEmpty() *Assertion {
	if a.actual == nil {
		return a
	}

	var val reflect.Value = reflect.ValueOf(a.actual)
	var kind reflect.Kind = val.Kind()

	switch kind {
	case reflect.String, reflect.Slice, reflect.Map, reflect.Array:
		if val.Len() != 0 {
			a.t.Helper()
			a.t.Errorf("Expected %s to be empty, but has length %d", a.name, val.Len())
			logFail(a.name, "IsEmpty")
		} else {
			logPass(a.name, "IsEmpty")
		}
	default:
		a.t.Helper()
		a.t.Errorf("Expected %s to be a string, slice, map, or array, but got type %T", a.name, a.actual)
	}
	return a
}

// IsNotEmpty asserts that a string, slice, map, or array is not empty.
func (a *Assertion) IsNotEmpty() *Assertion {
	if a.actual == nil {
		a.t.Helper()
		a.t.Errorf("Expected %s to not be empty, but it is nil", a.name)
		return a
	}

	var val reflect.Value = reflect.ValueOf(a.actual)
	var kind reflect.Kind = val.Kind()

	switch kind {
	case reflect.String, reflect.Slice, reflect.Map, reflect.Array:
		if val.Len() == 0 {
			a.t.Helper()
			a.t.Errorf("Expected %s to not be empty", a.name)
			logFail(a.name, "IsNotEmpty")
		} else {
			logPass(a.name, "IsNotEmpty")
		}
	default:
		a.t.Helper()
		a.t.Errorf("Expected %s to be a string, slice, map, or array, but got type %T", a.name, a.actual)
	}
	return a
}

// HasLength asserts that a string, slice, map, or array has the expected length.
func (a *Assertion) HasLength(expected int) *Assertion {
	if a.actual == nil {
		a.t.Helper()
		a.t.Errorf("Expected %s to have length %d, but it is nil", a.name, expected)
		return a
	}

	var val reflect.Value = reflect.ValueOf(a.actual)
	var kind reflect.Kind = val.Kind()

	switch kind {
	case reflect.String, reflect.Slice, reflect.Map, reflect.Array:
		var actualLen int = val.Len()
		if actualLen != expected {
			a.t.Helper()
			a.t.Errorf("Expected %s to have length %d, but got %d", a.name, expected, actualLen)
			logFail(a.name, fmt.Sprintf("HasLength(%d) — got %d", expected, actualLen))
		} else {
			logPass(a.name, fmt.Sprintf("HasLength(%d)", expected))
		}
	default:
		a.t.Helper()
		a.t.Errorf("Expected %s to be a string, slice, map, or array, but got type %T", a.name, a.actual)
	}
	return a
}

// IsGreaterThan asserts that a numeric value is greater than the expected value.
func (a *Assertion) IsGreaterThan(expected interface{}) *Assertion {
	var actualFloat float64
	var expectedFloat float64
	var ok bool

	actualFloat, ok = toFloat64(a.actual)
	if !ok {
		a.t.Helper()
		a.t.Errorf("Expected %s to be numeric, but got type %T", a.name, a.actual)
		return a
	}

	expectedFloat, ok = toFloat64(expected)
	if !ok {
		a.t.Helper()
		a.t.Errorf("Expected value to be numeric, but got type %T", expected)
		return a
	}

	if actualFloat <= expectedFloat {
		a.t.Helper()
		a.t.Errorf("Expected %s to be greater than %v, but got %v", a.name, expected, a.actual)
		logFail(a.name, fmt.Sprintf("IsGreaterThan(%v)", expected))
	} else {
		logPass(a.name, fmt.Sprintf("IsGreaterThan(%v)", expected))
	}
	return a
}

// IsLessThan asserts that a numeric value is less than the expected value.
func (a *Assertion) IsLessThan(expected interface{}) *Assertion {
	var actualFloat float64
	var expectedFloat float64
	var ok bool

	actualFloat, ok = toFloat64(a.actual)
	if !ok {
		a.t.Helper()
		a.t.Errorf("Expected %s to be numeric, but got type %T", a.name, a.actual)
		return a
	}

	expectedFloat, ok = toFloat64(expected)
	if !ok {
		a.t.Helper()
		a.t.Errorf("Expected value to be numeric, but got type %T", expected)
		return a
	}

	if actualFloat >= expectedFloat {
		a.t.Helper()
		a.t.Errorf("Expected %s to be less than %v, but got %v", a.name, expected, a.actual)
		logFail(a.name, fmt.Sprintf("IsLessThan(%v)", expected))
	} else {
		logPass(a.name, fmt.Sprintf("IsLessThan(%v)", expected))
	}
	return a
}

// IsGreaterThanOrEqual asserts that a numeric value is greater than or equal to the expected value.
func (a *Assertion) IsGreaterThanOrEqual(expected interface{}) *Assertion {
	var actualFloat float64
	var expectedFloat float64
	var ok bool

	actualFloat, ok = toFloat64(a.actual)
	if !ok {
		a.t.Helper()
		a.t.Errorf("Expected %s to be numeric, but got type %T", a.name, a.actual)
		return a
	}

	expectedFloat, ok = toFloat64(expected)
	if !ok {
		a.t.Helper()
		a.t.Errorf("Expected value to be numeric, but got type %T", expected)
		return a
	}

	if actualFloat < expectedFloat {
		a.t.Helper()
		a.t.Errorf("Expected %s to be >= %v, but got %v", a.name, expected, a.actual)
		logFail(a.name, fmt.Sprintf("IsGreaterThanOrEqual(%v)", expected))
	} else {
		logPass(a.name, fmt.Sprintf("IsGreaterThanOrEqual(%v)", expected))
	}
	return a
}

// IsLessThanOrEqual asserts that a numeric value is less than or equal to the expected value.
func (a *Assertion) IsLessThanOrEqual(expected interface{}) *Assertion {
	var actualFloat float64
	var expectedFloat float64
	var ok bool

	actualFloat, ok = toFloat64(a.actual)
	if !ok {
		a.t.Helper()
		a.t.Errorf("Expected %s to be numeric, but got type %T", a.name, a.actual)
		return a
	}

	expectedFloat, ok = toFloat64(expected)
	if !ok {
		a.t.Helper()
		a.t.Errorf("Expected value to be numeric, but got type %T", expected)
		return a
	}

	if actualFloat > expectedFloat {
		a.t.Helper()
		a.t.Errorf("Expected %s to be <= %v, but got %v", a.name, expected, a.actual)
		logFail(a.name, fmt.Sprintf("IsLessThanOrEqual(%v)", expected))
	} else {
		logPass(a.name, fmt.Sprintf("IsLessThanOrEqual(%v)", expected))
	}
	return a
}

// HasType asserts that the actual value has the expected type.
func (a *Assertion) HasType(expected interface{}) *Assertion {
	var actualType reflect.Type = reflect.TypeOf(a.actual)
	var expectedType reflect.Type = reflect.TypeOf(expected)

	if actualType != expectedType {
		a.t.Helper()
		a.t.Errorf("Expected %s to have type %v, but got %v", a.name, expectedType, actualType)
		logFail(a.name, fmt.Sprintf("HasType(%v)", expectedType))
	} else {
		logPass(a.name, fmt.Sprintf("HasType(%v)", expectedType))
	}
	return a
}

// toFloat64 converts various numeric types to float64.
func toFloat64(value interface{}) (float64, bool) {
	var val reflect.Value = reflect.ValueOf(value)
	var kind reflect.Kind = val.Kind()

	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(val.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(val.Uint()), true
	case reflect.Float32, reflect.Float64:
		return val.Float(), true
	default:
		return 0, false
	}
}
