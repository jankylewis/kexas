package kexas

import (
	"fmt"
	"time"

	"github.com/kexas-project/kexas/errors"
	"github.com/kexas-project/kexas/internal/cdp"
)

// Click performs an immediate mouse click on the element without waiting.
//
// This method clicks the element directly without checking if it's clickable.
// Use this when you're certain the element is ready or want maximum speed.
// For most cases, use WaitAndClick() instead.
func (e *Element) Click() error {
	if e == nil {
		return errors.ErrElementNil
	}

	if e.page == nil {
		return errors.ErrElementNoPage
	}

	if e.nodeID <= 0 && e.objectID == "" {
		return errors.ErrElementInvalidNodeID
	}

	// Resolve nodeId to objectId (CDP requires objectId for callFunctionOn)
	var objectID string
	var err error
	objectID, err = e.resolveObjectID()
	if err != nil {
		return fmt.Errorf("failed to resolve element: %w", err)
	}

	// Use CDP Runtime.callFunctionOn to click the element immediately
	var clickFunction string = `
		function() {
			this.click();
			return true;
		}
	`

	var params map[string]interface{} = map[string]interface{}{
		"functionDeclaration": clickFunction,
		"objectId":            objectID,
		"returnByValue":       true,
	}

	var result map[string]interface{}
	result, err = e.page.sendCommand(cdp.CmdRuntimeCallFunctionOn, params)
	if err != nil {
		return fmt.Errorf("failed to click element: %w", err)
	}

	// Check if the click was successful
	var success bool
	var ok bool
	success, ok = result["result"].(map[string]interface{})["value"].(bool)
	if !ok || !success {
		return errors.ErrClickOperationFailed
	}

	return nil
}

// WaitAndClick performs a mouse click on the element after waiting for it to be clickable.
//
// This method waits for the element to be clickable (visible and enabled)
// before performing the click action. It uses the element's default timeout.
// This is the recommended method for most use cases.
func (e *Element) WaitAndClick() error {
	if e == nil {
		return errors.ErrElementNil
	}

	if e.page == nil {
		return errors.ErrElementNoPage
	}

	if e.nodeID <= 0 && e.objectID == "" {
		return errors.ErrElementInvalidNodeID
	}

	// Wait for element to be clickable using default timeout
	var _, err error
	_, err = e.page.WaitForElementClickable(e.selector, e.timeout)
	if err != nil {
		return fmt.Errorf("element not clickable: %w", err)
	}

	// Resolve nodeId to objectId (CDP requires objectId for callFunctionOn)
	var objectID string
	objectID, err = e.resolveObjectID()
	if err != nil {
		return fmt.Errorf("failed to resolve element: %w", err)
	}

	// Use CDP Runtime.callFunctionOn to click the element
	var clickFunction string = `
		function() {
			this.click();
			return true;
		}
	`

	var params map[string]interface{} = map[string]interface{}{
		"functionDeclaration": clickFunction,
		"objectId":            objectID,
		"returnByValue":       true,
	}

	var result map[string]interface{}
	result, err = e.page.sendCommand(cdp.CmdRuntimeCallFunctionOn, params)
	if err != nil {
		return fmt.Errorf("failed to click element: %w", err)
	}

	// Check if the click was successful
	var success bool
	var ok bool
	success, ok = result["result"].(map[string]interface{})["value"].(bool)
	if !ok || !success {
		return errors.ErrClickOperationFailed
	}

	return nil
}

// WaitAndClickFor performs a mouse click on the element after waiting for it to be clickable with a custom timeout.
//
// This method waits for the element to be clickable using the specified timeout.
// Use this when you need more control over the wait time than the default timeout.
func (e *Element) WaitAndClickFor(timeout time.Duration) error {
	// Validate timeout
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

	// Wait for element to be clickable using custom timeout
	var _, err error
	_, err = e.page.WaitForElementClickable(e.selector, timeout)
	if err != nil {
		return fmt.Errorf("element not clickable within %v: %w", timeout, err)
	}

	// Resolve nodeId to objectId
	var objectID string
	objectID, err = e.resolveObjectID()
	if err != nil {
		return fmt.Errorf("failed to resolve element: %w", err)
	}

	// Use CDP Runtime.callFunctionOn to click the element
	var clickFunction string = `
		function() {
			this.click();
			return true;
		}
	`

	var params map[string]interface{} = map[string]interface{}{
		"functionDeclaration": clickFunction,
		"objectId":            objectID,
		"returnByValue":       true,
	}

	var result map[string]interface{}
	result, err = e.page.sendCommand(cdp.CmdRuntimeCallFunctionOn, params)
	if err != nil {
		return fmt.Errorf("failed to click element: %w", err)
	}

	// Check if the click was successful
	var success bool
	var ok bool
	success, ok = result["result"].(map[string]interface{})["value"].(bool)
	if !ok || !success {
		return errors.ErrClickOperationFailed
	}

	return nil
}

// Type performs immediate typing into the element without waiting.
//
// This method types the text directly into the element without checking if it's visible.
// Use this when you're certain the element is ready or want maximum speed.
// For most cases, use WaitAndType() instead.
//
// This method works with input elements, textarea elements, and other
// content-editable elements. It clears any existing content before typing.
func (e *Element) Type(text string) error {
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

	// Resolve nodeId to objectId
	var objectID string
	var err error
	objectID, err = e.resolveObjectID()
	if err != nil {
		return fmt.Errorf("failed to resolve element: %w", err)
	}

	// Focus and clear the element via JS
	var focusFunction string = `
		function() {
			this.focus();
			this.value = '';
			return true;
		}
	`
	_, err = e.page.sendCommand(cdp.CmdRuntimeCallFunctionOn, map[string]interface{}{
		"functionDeclaration": focusFunction,
		"objectId":            objectID,
		"returnByValue":       true,
	})
	if err != nil {
		return fmt.Errorf("failed to focus element: %w", err)
	}

	// Type each character using CDP Input.dispatchKeyEvent (Playwright pattern)
	// Playwright dispatches 3 events per character: keyDown → char → keyUp
	// keyDown/keyUp carry "key" only; "char" carries "text" to insert the character.
	// If keyDown includes "text", Chrome inserts the char twice (once on keyDown, once on char).
	for _, ch := range text {
		var charStr string = string(ch)

		// keyDown (no "text" — only identifies which key was pressed)
		_, err = e.page.sendCommand("Input.dispatchKeyEvent", map[string]interface{}{
			"type": "keyDown",
			"key":  charStr,
		})
		if err != nil {
			return fmt.Errorf("failed to dispatch keyDown: %w", err)
		}

		// char (carries "text" — this is what actually inserts the character)
		_, err = e.page.sendCommand("Input.dispatchKeyEvent", map[string]interface{}{
			"type":           "char",
			"text":           charStr,
			"unmodifiedText": charStr,
			"key":            charStr,
		})
		if err != nil {
			return fmt.Errorf("failed to dispatch char: %w", err)
		}

		// keyUp (no "text" — only signals key release)
		_, err = e.page.sendCommand("Input.dispatchKeyEvent", map[string]interface{}{
			"type": "keyUp",
			"key":  charStr,
		})
		if err != nil {
			return fmt.Errorf("failed to dispatch keyUp: %w", err)
		}
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
		return fmt.Errorf("element not visible: %w", err)
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
		return fmt.Errorf("element not visible within %v: %w", timeout, err)
	}

	return e.Type(text)
}
