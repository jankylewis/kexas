package kcore

import (
	"fmt"

	"github.com/jankylewis/kexas/errors"
	"github.com/jankylewis/kexas/internal/cdp"
)

// ScrollIntoView scrolls the page until this element is visible in the viewport.
//
// Uses Element.scrollIntoView({ block: 'center', inline: 'center' }) to center
// the element in the viewport. This mirrors Playwright's scrollIntoViewIfNeeded().
//
// This is an element-level scroll action — use Page.ScrollToTop/ScrollToBottom
// for page-level scrolling.
func (e *Element) ScrollIntoView() error {
	if e == nil {
		return errors.ErrElementNil
	}

	if e.page == nil {
		return errors.ErrElementNoPage
	}

	if e.nodeID <= 0 && e.objectID == "" {
		return errors.ErrElementInvalidNodeID
	}

	var objectID string
	var err error
	objectID, err = e.resolveObjectID()
	if err != nil {
		return fmt.Errorf("failed to resolve element for scroll: %w", err)
	}

	var scrollFunction string = `
		function() {
			this.scrollIntoView({ behavior: 'instant', block: 'center', inline: 'center' });
			return true;
		}
	`

	var params map[string]interface{} = map[string]interface{}{
		"functionDeclaration": scrollFunction,
		"objectId":            objectID,
		"returnByValue":       true,
	}

	var result map[string]interface{}
	result, err = e.page.sendCommand(cdp.CmdRuntimeCallFunctionOn, params)
	if err != nil {
		return fmt.Errorf("scroll into view failed: %w", err)
	}

	var success bool
	var ok bool
	success, ok = result["result"].(map[string]interface{})["value"].(bool)
	if !ok || !success {
		return errors.ErrScrollOperationFailed
	}

	return nil
}
