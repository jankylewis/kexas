package kcore

import (
	"fmt"
	"strings"

	"github.com/jankylewis/kexas/errors"
)


// StorageType represents the type of web storage.
type StorageType string

// Storage type constants.
const (
	StorageTypeLocal   StorageType = "localStorage"
	StorageTypeSession StorageType = "sessionStorage"
)

// Storage provides an interface for interacting with localStorage or sessionStorage.
type Storage struct {
	page        *Page
	storageType StorageType
}

// LocalStorage returns a Storage instance bound to localStorage.
func (p *Page) LocalStorage() *Storage {
	return &Storage{
		page:        p,
		storageType: StorageTypeLocal,
	}
}

// SessionStorage returns a Storage instance bound to sessionStorage.
func (p *Page) SessionStorage() *Storage {
	return &Storage{
		page:        p,
		storageType: StorageTypeSession,
	}
}

// Set stores a key-value pair in the storage.
func (s *Storage) Set(key string, value string) error {
	if key == "" {
		return errors.ErrStorageKeyEmpty
	}

	var escapedKey string = escapeJSString(key)
	var escapedValue string = escapeJSString(value)
	var expression string = fmt.Sprintf("%s.setItem('%s', '%s')", s.storageType, escapedKey, escapedValue)

	var err error
	_, err = s.page.Evaluate(expression)
	if err != nil {
		return fmt.Errorf("storage set '%s' failed: %w", key, err)
	}

	return nil
}

// Get retrieves the value for a key from the storage.
// Returns empty string if the key does not exist.
func (s *Storage) Get(key string) (string, error) {
	if key == "" {
		return "", errors.ErrStorageKeyEmpty
	}

	var escapedKey string = escapeJSString(key)
	var expression string = fmt.Sprintf("%s.getItem('%s')", s.storageType, escapedKey)

	var result interface{}
	var err error
	result, err = s.page.Evaluate(expression)
	if err != nil {
		return "", fmt.Errorf("storage get '%s' failed: %w", key, err)
	}

	if result == nil {
		return "", nil
	}

	var value string
	var ok bool
	value, ok = result.(string)
	if !ok {
		return "", nil
	}

	return value, nil
}

// Remove deletes a key from the storage.
func (s *Storage) Remove(key string) error {
	if key == "" {
		return errors.ErrStorageKeyEmpty
	}

	var escapedKey string = escapeJSString(key)
	var expression string = fmt.Sprintf("%s.removeItem('%s')", s.storageType, escapedKey)

	var err error
	_, err = s.page.Evaluate(expression)
	if err != nil {
		return fmt.Errorf("storage remove '%s' failed: %w", key, err)
	}

	return nil
}

// Clear removes all items from the storage.
func (s *Storage) Clear() error {
	var expression string = fmt.Sprintf("%s.clear()", s.storageType)

	var err error
	_, err = s.page.Evaluate(expression)
	if err != nil {
		return fmt.Errorf("storage clear failed: %w", err)
	}

	return nil
}

// GetAll returns all key-value pairs in the storage as a map.

func escapeJSString(s string) string {
	var replacer *strings.Replacer = strings.NewReplacer(
		`\`, `\\`,
		`'`, `\'`,
		`"`, `\"`,
		"\n", `\n`,
		"\r", `\r`,
		"\t", `\t`,
	)
	return replacer.Replace(s)
}
