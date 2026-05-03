package kcore

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jankylewis/kexas/errors"
)

func (s *Storage) GetAll() (map[string]string, error) {
	var expression string = fmt.Sprintf(`(function() {
		var result = {};
		for (var i = 0; i < %s.length; i++) {
			var key = %s.key(i);
			result[key] = %s.getItem(key);
		}
		return JSON.stringify(result);
	})()`, s.storageType, s.storageType, s.storageType)

	var result interface{}
	var err error
	result, err = s.page.Evaluate(expression)
	if err != nil {
		return nil, fmt.Errorf("storage getAll failed: %w", err)
	}

	if result == nil {
		return map[string]string{}, nil
	}

	var jsonStr string
	var ok bool
	jsonStr, ok = result.(string)
	if !ok {
		return map[string]string{}, nil
	}

	var items map[string]string = make(map[string]string)
	err = json.Unmarshal([]byte(jsonStr), &items)
	if err != nil {
		return nil, fmt.Errorf("storage getAll parse failed: %w", err)
	}

	return items, nil
}

// Length returns the number of items in the storage.
func (s *Storage) Length() (int, error) {
	var expression string = fmt.Sprintf("%s.length", s.storageType)

	var result interface{}
	var err error
	result, err = s.page.Evaluate(expression)
	if err != nil {
		return 0, fmt.Errorf("storage length failed: %w", err)
	}

	// CDP returns numbers as float64
	var length float64
	var ok bool
	length, ok = result.(float64)
	if !ok {
		return 0, nil
	}

	return int(length), nil
}

// Has checks if a key exists in the storage.
func (s *Storage) Has(key string) (bool, error) {
	if key == "" {
		return false, errors.ErrStorageKeyEmpty
	}

	var escapedKey string = escapeJSString(key)
	var expression string = fmt.Sprintf("%s.getItem('%s') !== null", s.storageType, escapedKey)

	var result interface{}
	var err error
	result, err = s.page.Evaluate(expression)
	if err != nil {
		return false, fmt.Errorf("storage has '%s' failed: %w", key, err)
	}

	var exists bool
	var ok bool
	exists, ok = result.(bool)
	if !ok {
		return false, nil
	}

	return exists, nil
}

// SetMany stores multiple key-value pairs in the storage in a single evaluation.
func (s *Storage) SetMany(items map[string]string) error {
	if len(items) == 0 {
		return nil
	}

	var builder strings.Builder
	builder.WriteString("(function() { ")
	for key, value := range items {
		if key == "" {
			return errors.ErrStorageKeyEmpty
		}
		var escapedKey string = escapeJSString(key)
		var escapedValue string = escapeJSString(value)
		builder.WriteString(fmt.Sprintf("%s.setItem('%s', '%s'); ", s.storageType, escapedKey, escapedValue))
	}
	builder.WriteString("return true; })()")

	var err error
	_, err = s.page.Evaluate(builder.String())
	if err != nil {
		return fmt.Errorf("storage setMany failed: %w", err)
	}

	return nil
}

// escapeJSString escapes a string for safe inclusion in a JavaScript string literal.
