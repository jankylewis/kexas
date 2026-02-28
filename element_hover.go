package kexas

import (
	"fmt"
	"time"

	"github.com/kexas-project/kexas/errors"
	"github.com/kexas-project/kexas/internal/cdp"
)

// Hover performs an immediate mouse hover over the element without waiting.
//
// This method hovers over the element directly without checking if it's visible.
// Use this when you're certain the element is ready or want maximum speed.
// For most cases, use WaitAndHover() instead.
func (e *Element) Hover() error {
	if e == nil {
		return errors.ErrElementNil
	}

	if e.page == nil {
		return errors.ErrElementNoPage
	}

	if e.nodeID <= 0 && e.objectID == "" {
		return errors.ErrElementInvalidNodeID
	}

	// Resolve nodeId to objectId
	var objectID string
	var err error
	objectID, err = e.resolveObjectID()
	if err != nil {
		return errors.ResolveElementFailed(err)
	}

	return e.dispatchHoverEvent(objectID)
}

// WaitAndHover performs a mouse hover over the element after waiting for it to be visible.
//
// This method waits for the element to be visible before performing the hover action.
// It uses the element's default timeout. This is the recommended method for most use cases.
func (e *Element) WaitAndHover() error {
	if e == nil {
		return errors.ErrElementNil
	}

	if e.page == nil {
		return errors.ErrElementNoPage
	}

	if e.nodeID <= 0 && e.objectID == "" {
		return errors.ErrElementInvalidNodeID
	}

	// Use CDP to wait for element to be visible using default timeout
	var _, err error
	_, err = e.page.WaitForElementVisible(e.selector, e.timeout)
	if err != nil {
		return errors.ElementNotVisibleWrap(err)
	}

	// Resolve nodeId to objectId
	var objectID string
	objectID, err = e.resolveObjectID()
	if err != nil {
		return errors.ResolveElementFailed(err)
	}

	return e.dispatchHoverEvent(objectID)
}

// WaitAndHoverFor performs a mouse hover over the element after waiting for it to be visible with a custom timeout.
//
// This method waits for the element to be visible using the specified timeout.
// Use this when you need more control over the wait time than the default timeout.
func (e *Element) WaitAndHoverFor(timeout time.Duration) error {
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

	// Wait for element to be visible using custom timeout
	var _, err error
	_, err = e.page.WaitForElementVisible(e.selector, timeout)
	if err != nil {
		return errors.ElementNotVisibleWithinWrap(timeout, err)
	}

	// Resolve nodeId to objectId
	var objectID string
	objectID, err = e.resolveObjectID()
	if err != nil {
		return errors.ResolveElementFailed(err)
	}

	return e.dispatchHoverEvent(objectID)
}

// dispatchHoverEvent dispatches a mouseover event on the element via CDP.
// Shared by Hover(), WaitAndHover(), and WaitAndHoverFor().
func (e *Element) dispatchHoverEvent(objectID string) error {
	var hoverFunction string = `
		function() {
			var event = new MouseEvent('mouseover', {
				'view': window,
				'bubbles': true,
				'cancelable': true
			});
			this.dispatchEvent(event);
			return true;
		}
	`

	var params map[string]interface{} = map[string]interface{}{
		"functionDeclaration": hoverFunction,
		"objectId":            objectID,
		"returnByValue":       true,
	}

	var result map[string]interface{}
	var err error
	result, err = e.page.sendCommand(cdp.CmdRuntimeCallFunctionOn, params)
	if err != nil {
		return fmt.Errorf("failed to hover over element: %w", err)
	}

	// Check if the hover was successful
	var success bool
	var ok bool
	success, ok = result["result"].(map[string]interface{})["value"].(bool)
	if !ok || !success {
		return errors.ErrHoverOperationFailed
	}

	return nil
}
