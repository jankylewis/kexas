package kcore

import (
	"fmt"
	"time"

	"github.com/jankylewis/kexas/errors"
	"github.com/jankylewis/kexas/internal/cdp"
)

func (e *Element) Type(text string) error {
	var err error = e.checkTypeArgs(text)
	if err != nil {
		return err
	}

	var objectID string
	objectID, err = e.resolveObjectID()
	if err != nil {
		return errors.ResolveElementFailed(err)
	}

	err = e.focusAndClear(objectID)
	if err != nil {
		return err
	}

	for _, ch := range text {
		err = e.dispatchCharacter(string(ch))
		if err != nil {
			return err
		}
	}
	return nil
}

// Fill sets the input value in one programmatic shot — Playwright-style.
//
// Unlike Type which dispatches per-character native key events (slow + flaky
// on React/Vue/Angular controlled inputs after navigation), Fill uses the
// React-aware HTMLInputElement.value setter and dispatches a single bubbling
// 'input' event so framework onChange handlers fire correctly. Faster (one CDP
// roundtrip vs. N×3) and rock-solid against controlled inputs.
//
// Use Fill for forms (login, signup, profile, search-without-autocomplete).
// Use Type when you need real keyboard events (autocomplete, keyboard shortcuts,
// IME composition, onkeydown handlers).
func (e *Element) Fill(text string) error {
	var err error = e.checkFillArgs(text)
	if err != nil {
		return err
	}

	var objectID string
	objectID, err = e.resolveObjectID()
	if err != nil {
		return errors.ResolveElementFailed(err)
	}

	// React's _valueTracker only fires onChange when the value is set via the
	// native prototype setter. Direct `this.value = X` is silently swallowed.
	var fillFunction string = `
		function(value) {
			const proto = window.HTMLInputElement.prototype;
			const desc = Object.getOwnPropertyDescriptor(proto, 'value')
				|| Object.getOwnPropertyDescriptor(window.HTMLTextAreaElement.prototype, 'value');
			if (desc && desc.set) {
				desc.set.call(this, value);
			} else {
				this.value = value;
			}
			this.dispatchEvent(new Event('input', { bubbles: true }));
			this.dispatchEvent(new Event('change', { bubbles: true }));
			return true;
		}
	`
	_, err = e.page.sendCommand(cdp.CmdRuntimeCallFunctionOn, map[string]interface{}{
		"functionDeclaration": fillFunction,
		"objectId":            objectID,
		"arguments":           []interface{}{map[string]interface{}{"value": text}},
		"returnByValue":       true,
	})
	if err != nil {
		return fmt.Errorf("fill: callFunctionOn failed: %w", err)
	}
	return nil
}

// checkFillArgs is the Fill counterpart of checkTypeArgs. Allows empty string
// because Fill("") is a meaningful "clear the field" operation, unlike Type("").
func (e *Element) checkFillArgs(text string) error {
	_ = text // empty string is allowed for Fill
	if e == nil {
		return errors.ErrElementNil
	}
	if e.page == nil {
		return errors.ErrElementNoPage
	}
	if e.nodeID <= 0 && e.objectID == "" {
		return errors.ErrElementInvalidNodeID
	}
	return nil
}

// checkTypeArgs validates the receiver and text argument before typing.
func (e *Element) checkTypeArgs(text string) error {
	if e == nil {
		return errors.ErrElementNil
	}
	if e.page == nil {
		return errors.ErrElementNoPage
	}
	if e.nodeID <= 0 && e.objectID == "" {
		return errors.ErrElementInvalidNodeID
	}
	if text == "" {
		return errors.ErrTextEmpty
	}
	return nil
}

// focusAndClear focuses the element and clears its value via Runtime.callFunctionOn.
func (e *Element) focusAndClear(objectID string) error {
	var focusFunction string = `
		function() {
			this.focus();
			this.value = '';
			return true;
		}
	`
	var err error
	_, err = e.page.sendCommand(cdp.CmdRuntimeCallFunctionOn, map[string]interface{}{
		"functionDeclaration": focusFunction,
		"objectId":            objectID,
		"returnByValue":       true,
	})
	if err != nil {
		return errors.FocusElementFailed(err)
	}
	return nil
}

// dispatchCharacter sends a single character via the Playwright 3-event pattern:
// keyDown (key only) → char (text inserts the glyph) → keyUp (key only).
// Including "text" on keyDown would cause Chrome to insert the character twice.
func (e *Element) dispatchCharacter(charStr string) error {
	var err error
	_, err = e.page.sendCommand(cdp.CmdInputDispatchKeyEvent, map[string]interface{}{
		"type": "keyDown",
		"key":  charStr,
	})
	if err != nil {
		return fmt.Errorf("failed to dispatch keyDown: %w", err)
	}

	_, err = e.page.sendCommand(cdp.CmdInputDispatchKeyEvent, map[string]interface{}{
		"type":           "char",
		"text":           charStr,
		"unmodifiedText": charStr,
		"key":            charStr,
	})
	if err != nil {
		return fmt.Errorf("failed to dispatch char: %w", err)
	}

	_, err = e.page.sendCommand(cdp.CmdInputDispatchKeyEvent, map[string]interface{}{
		"type": "keyUp",
		"key":  charStr,
	})
	if err != nil {
		return fmt.Errorf("failed to dispatch keyUp: %w", err)
	}
	return nil
}

// WaitAndType performs typing into the element after waiting for it to be visible.
//
// This method waits for the element to be visible before performing the typing action.
// It uses the element's default timeout. This is the recommended method for most use cases.
//
// This method works with input elements, textarea elements, and other
// content-editable elements. It clears any existing content before typing.
func (e *Element) WaitAndType(text string) error {
	if e == nil {
		return errors.ErrElementNil
	}

	if e.page == nil {
		return errors.ErrElementNoPage
	}

	if e.nodeID <= 0 && e.objectID == "" {
		return errors.ErrElementInvalidNodeID
	}

	if text == "" {
		return errors.ErrTextEmpty
	}

	var _, err error
	_, err = e.page.WaitForElementVisible(e.selector, e.timeout)
	if err != nil {
		return errors.ElementNotVisibleWrap(err)
	}

	return e.Type(text)
}

// WaitAndTypeFor performs typing into the element after waiting for it to be visible with a custom timeout.
//
// This method waits for the element to be visible using the specified timeout.
// Use this when you need more control over the wait time than the default timeout.
//
// This method works with input elements, textarea elements, and other
// content-editable elements. It clears any existing content before typing.
func (e *Element) WaitAndTypeFor(text string, timeout time.Duration) error {
	if timeout < 1*time.Second {
		return errors.TimeoutInvalidFormat(timeout)
	}

	if e == nil {
		return errors.ErrElementNil
	}

	if e.page == nil {
		return errors.ErrElementNoPage
	}

	if e.nodeID <= 0 && e.objectID == "" {
		return errors.ErrElementInvalidNodeID
	}

	if text == "" {
		return errors.ErrTextEmpty
	}

	var _, err error
	_, err = e.page.WaitForElementVisible(e.selector, timeout)
	if err != nil {
		return errors.ElementNotVisibleWithinWrap(timeout, err)
	}

	return e.Type(text)
}
