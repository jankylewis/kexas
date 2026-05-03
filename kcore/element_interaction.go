package kcore

import (
	"fmt"
	"time"

	"github.com/jankylewis/kexas/errors"
	"github.com/jankylewis/kexas/internal/cdp"
)

// Use this when you're certain the element is ready or want maximum speed.
// For most cases, use WaitAndClick() instead.
func (e *Element) Click() error {
	var err error = e.checkClickArgs()
	if err != nil {
		return err
	}
	return e.dispatchClickViaObjectID()
}

// WaitAndClick performs a mouse click on the element after waiting for it to be clickable.
//
// This method waits for the element to be clickable (visible and enabled)
// before performing the click action. It uses the element's default timeout.
// This is the recommended method for most use cases.
func (e *Element) WaitAndClick() error {
	var err error = e.checkClickArgs()
	if err != nil {
		return err
	}
	_, err = e.page.WaitForElementClickable(e.selector, e.timeout)
	if err != nil {
		return fmt.Errorf("element not clickable: %w", err)
	}
	return e.dispatchClickViaObjectID()
}

// WaitAndClickFor performs a mouse click on the element after waiting for it to be clickable with a custom timeout.
//
// This method waits for the element to be clickable using the specified timeout.
// Use this when you need more control over the wait time than the default timeout.
func (e *Element) WaitAndClickFor(timeout time.Duration) error {
	if timeout < 1*time.Second {
		return errors.TimeoutInvalidFormat(timeout)
	}
	var err error = e.checkClickArgs()
	if err != nil {
		return err
	}
	_, err = e.page.WaitForElementClickable(e.selector, timeout)
	if err != nil {
		return fmt.Errorf("element not clickable within %v: %w", timeout, err)
	}
	return e.dispatchClickViaObjectID()
}

// checkClickArgs validates the receiver is in a usable state for a click.
// Same guard set as Type/Fill — kept as a separate helper because Click and
// WaitAndClickFor both need it.
func (e *Element) checkClickArgs() error {
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

// dispatchClickViaObjectID resolves the element to an objectId and dispatches
// the actual click via Runtime.callFunctionOn. Shared by all click variants.
func (e *Element) dispatchClickViaObjectID() error {
	var objectID string
	var err error
	objectID, err = e.resolveObjectID()
	if err != nil {
		return errors.ResolveElementFailed(err)
	}
	var clickFunction string = `function() { this.click(); return true; }`
	var result map[string]interface{}
	result, err = e.page.sendCommand(cdp.CmdRuntimeCallFunctionOn, map[string]interface{}{
		"functionDeclaration": clickFunction,
		"objectId":            objectID,
		"returnByValue":       true,
	})
	if err != nil {
		return errors.ClickElementFailed(err)
	}
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
