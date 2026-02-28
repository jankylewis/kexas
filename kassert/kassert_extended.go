package kassert

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
)

// MatchesRegex asserts that a string matches the given regex pattern.
//
// Useful for URL validation, dynamic content, or partial string matching.
func (a *Assertion) MatchesRegex(pattern string) *Assertion {
	var str string
	var ok bool
	str, ok = a.actual.(string)
	if !ok {
		a.t.Helper()
		a.t.Errorf("Expected %s to be a string, but got type %T", a.name, a.actual)
		return a
	}

	var re *regexp.Regexp
	var err error
	re, err = regexp.Compile(pattern)
	if err != nil {
		a.t.Helper()
		a.t.Errorf("Invalid regex pattern %q: %v", pattern, err)
		return a
	}

	if !re.MatchString(str) {
		a.t.Helper()
		a.t.Errorf("Expected %s to match regex:\n  %q\nbut got:\n  %q", a.name, pattern, str)
	}
	logPass(a.name, fmt.Sprintf("MatchesRegex(%q)", pattern))
	return a
}

// IsOneOf asserts that the actual value is one of the allowed values.
//
// Useful for status checks (e.g., "enabled" or "disabled").
func (a *Assertion) IsOneOf(allowed ...interface{}) *Assertion {
	for _, val := range allowed {
		if reflect.DeepEqual(a.actual, val) {
			logPass(a.name, fmt.Sprintf("IsOneOf(%v)", allowed))
			return a
		}
	}

	a.t.Helper()
	a.t.Errorf("Expected %s to be one of %v, but got:\n  %v", a.name, allowed, a.actual)
	logFail(a.name, fmt.Sprintf("IsOneOf — expected one of %v, got %v", allowed, a.actual))
	return a
}

// ContainsAll asserts that a string contains ALL of the given substrings.
//
// Useful for verifying page content with multiple expected texts.
func (a *Assertion) ContainsAll(substrings ...string) *Assertion {
	var str string
	var ok bool
	str, ok = a.actual.(string)
	if !ok {
		a.t.Helper()
		a.t.Errorf("Expected %s to be a string, but got type %T", a.name, a.actual)
		return a
	}

	var missing []string
	for _, sub := range substrings {
		if !containsString(str, sub) {
			missing = append(missing, sub)
		}
	}

	if len(missing) > 0 {
		a.t.Helper()
		a.t.Errorf("Expected %s to contain all of %q, but missing: %q\ngot:\n  %q",
			a.name, substrings, missing, str)
		logFail(a.name, fmt.Sprintf("ContainsAll — missing %q", missing))
	} else {
		logPass(a.name, "ContainsAll")
	}
	return a
}

// HasLengthGreaterThan asserts that a string, slice, map, or array has length > n.
//
// Common pattern: "cart has at least 1 item".
func (a *Assertion) HasLengthGreaterThan(n int) *Assertion {
	if a.actual == nil {
		a.t.Helper()
		a.t.Errorf("Expected %s to have length > %d, but it is nil", a.name, n)
		return a
	}

	var val reflect.Value = reflect.ValueOf(a.actual)
	var kind reflect.Kind = val.Kind()

	switch kind {
	case reflect.String, reflect.Slice, reflect.Map, reflect.Array:
		var actualLen int = val.Len()
		if actualLen <= n {
			a.t.Helper()
			a.t.Errorf("Expected %s to have length > %d, but got %d", a.name, n, actualLen)
			logFail(a.name, fmt.Sprintf("HasLengthGreaterThan(%d) — got %d", n, actualLen))
		} else {
			logPass(a.name, fmt.Sprintf("HasLengthGreaterThan(%d)", n))
		}
	default:
		a.t.Helper()
		a.t.Errorf("Expected %s to be a string, slice, map, or array, but got type %T", a.name, a.actual)
	}
	return a
}

// Is asserts that the error matches the target using errors.Is().
//
// This checks the entire error chain, supporting wrapped errors.
func (e *ErrorAssertion) Is(target error) *ErrorAssertion {
	if e.err == nil {
		e.t.Helper()
		e.t.Errorf("Expected an error matching %v, but got nil", target)
		logFail("error", fmt.Sprintf("Is(%v) — got nil", target))
		return e
	}

	if !errors.Is(e.err, target) {
		e.t.Helper()
		e.t.Errorf("Expected error to match:\n  %v\nbut got:\n  %v", target, e.err)
		logFail("error", fmt.Sprintf("Is(%v) — got %v", target, e.err))
	} else {
		logPass("error", fmt.Sprintf("Is(%v)", target))
	}
	return e
}

// containsString is a helper to avoid importing strings in this file.
func containsString(s string, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
