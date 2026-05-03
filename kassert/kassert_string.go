package kassert

import (
	"fmt"
	"reflect"
	"strings"
)

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
		return a
	}
	logPass(a.name, fmt.Sprintf("Contains(%q)", substring))
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
		return a
	}
	logPass(a.name, fmt.Sprintf("NotContains(%q)", substring))
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
		return a
	}
	logPass(a.name, fmt.Sprintf("StartsWith(%q)", prefix))
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
		return a
	}
	logPass(a.name, fmt.Sprintf("EndsWith(%q)", suffix))
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
			return a
		}
		logPass(a.name, "IsEmpty")
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
			return a
		}
		logPass(a.name, "IsNotEmpty")
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
			return a
		}
		logPass(a.name, fmt.Sprintf("HasLength(%d)", expected))
	default:
		a.t.Helper()
		a.t.Errorf("Expected %s to be a string, slice, map, or array, but got type %T", a.name, a.actual)
	}
	return a
}
