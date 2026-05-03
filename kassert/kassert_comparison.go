package kassert

import (
	"fmt"
	"reflect"
)

// IsGreaterThan asserts that a numeric value is greater than the expected value.
func (a *Assertion) IsGreaterThan(expected interface{}) *Assertion {
	var actualFloat, expectedFloat float64
	var ok bool
	actualFloat, expectedFloat, ok = a.coerceNumericPair(expected)
	if !ok {
		return a
	}
	if actualFloat <= expectedFloat {
		a.t.Helper()
		a.t.Errorf("Expected %s to be greater than %v, but got %v", a.name, expected, a.actual)
		logFail(a.name, fmt.Sprintf("IsGreaterThan(%v)", expected))
		return a
	}
	logPass(a.name, fmt.Sprintf("IsGreaterThan(%v)", expected))
	return a
}

// IsLessThan asserts that a numeric value is less than the expected value.
func (a *Assertion) IsLessThan(expected interface{}) *Assertion {
	var actualFloat, expectedFloat float64
	var ok bool
	actualFloat, expectedFloat, ok = a.coerceNumericPair(expected)
	if !ok {
		return a
	}
	if actualFloat >= expectedFloat {
		a.t.Helper()
		a.t.Errorf("Expected %s to be less than %v, but got %v", a.name, expected, a.actual)
		logFail(a.name, fmt.Sprintf("IsLessThan(%v)", expected))
		return a
	}
	logPass(a.name, fmt.Sprintf("IsLessThan(%v)", expected))
	return a
}

// IsGreaterThanOrEqual asserts that a numeric value is greater than or equal to the expected value.
func (a *Assertion) IsGreaterThanOrEqual(expected interface{}) *Assertion {
	var actualFloat, expectedFloat float64
	var ok bool
	actualFloat, expectedFloat, ok = a.coerceNumericPair(expected)
	if !ok {
		return a
	}
	if actualFloat < expectedFloat {
		a.t.Helper()
		a.t.Errorf("Expected %s to be >= %v, but got %v", a.name, expected, a.actual)
		logFail(a.name, fmt.Sprintf("IsGreaterThanOrEqual(%v)", expected))
		return a
	}
	logPass(a.name, fmt.Sprintf("IsGreaterThanOrEqual(%v)", expected))
	return a
}

// IsLessThanOrEqual asserts that a numeric value is less than or equal to the expected value.
func (a *Assertion) IsLessThanOrEqual(expected interface{}) *Assertion {
	var actualFloat, expectedFloat float64
	var ok bool
	actualFloat, expectedFloat, ok = a.coerceNumericPair(expected)
	if !ok {
		return a
	}
	if actualFloat > expectedFloat {
		a.t.Helper()
		a.t.Errorf("Expected %s to be <= %v, but got %v", a.name, expected, a.actual)
		logFail(a.name, fmt.Sprintf("IsLessThanOrEqual(%v)", expected))
		return a
	}
	logPass(a.name, fmt.Sprintf("IsLessThanOrEqual(%v)", expected))
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
		return a
	}
	logPass(a.name, fmt.Sprintf("HasType(%v)", expectedType))
	return a
}

// coerceNumericPair attempts to convert both a.actual and expected to float64.
// On failure it reports the type error and returns ok=false; the calling assertion
// short-circuits, matching the original per-assertion early-return behavior.
func (a *Assertion) coerceNumericPair(expected interface{}) (float64, float64, bool) {
	var actualFloat float64
	var ok bool
	actualFloat, ok = toFloat64(a.actual)
	if !ok {
		a.t.Helper()
		a.t.Errorf("Expected %s to be numeric, but got type %T", a.name, a.actual)
		return 0, 0, false
	}
	var expectedFloat float64
	expectedFloat, ok = toFloat64(expected)
	if !ok {
		a.t.Helper()
		a.t.Errorf("Expected value to be numeric, but got type %T", expected)
		return 0, 0, false
	}
	return actualFloat, expectedFloat, true
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
