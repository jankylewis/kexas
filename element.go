package kexas

import (
	"fmt"
	"time"

	"github.com/kexas-project/kexas/internal/cdp"
)

// Element represents a DOM element on a web page.
//
// Element provides methods for interacting with DOM elements such as clicking,
// typing text, getting attributes, and checking visibility state.
// Following go-rod's pattern, Element stores the remote objectId for reliable
// interaction via Runtime.callFunctionOn.
type Element struct {
	page     *Page         // Reference to the page containing this element
	selector string        // The selector used to find this element
	nodeID   cdp.NodeID    // CDP node identifier for this element
	objectID string        // CDP remote object ID for Runtime.callFunctionOn
	timeout  time.Duration // Default timeout for operations
}

// NewElement creates a new Element instance.
//
// This is typically called by Page.Find() methods and should not be called
// directly by users.
func NewElement(page *Page, selector string, nodeID cdp.NodeID, timeout time.Duration) *Element {
	return &Element{
		page:     page,
		selector: selector,
		nodeID:   nodeID,
		timeout:  timeout,
	}
}

// NewElementWithObject creates a new Element with both nodeId and objectId.
// This is the preferred constructor as it avoids extra DOM.resolveNode calls.
func NewElementWithObject(page *Page, selector string, nodeID cdp.NodeID, objectID string, timeout time.Duration) *Element {
	return &Element{
		page:     page,
		selector: selector,
		nodeID:   nodeID,
		objectID: objectID,
		timeout:  timeout,
	}
}

// Selector returns the CSS selector used to find this element.
func (e *Element) Selector() string {
	return e.selector
}

// NodeID returns the CDP node ID of this element.
func (e *Element) NodeID() cdp.NodeID {
	return e.nodeID
}

// Timeout returns the default timeout for element operations.
func (e *Element) Timeout() time.Duration {
	return e.timeout
}

// String returns a string representation of the element.
func (e *Element) String() string {
	return fmt.Sprintf("Element{selector: %s, nodeID: %d}", e.selector, e.nodeID)
}

// IsVisible checks if the element is visible on the page.
//
// Uses JavaScript to check the element's visibility, computed style, and dimensions.
// Returns true if the element is visible, false otherwise.
func (e *Element) IsVisible() bool {
	if e == nil || e.page == nil || (e.nodeID <= 0 && e.objectID == "") {
		return false
	}

	// Get objectId (cached or resolved)
	var objectID string
	var err error
	objectID, err = e.resolveObjectID()
	if err != nil {
		return false
	}

	// Use Runtime.callFunctionOn to check visibility via JavaScript
	var visibilityCheck string = `
		function() {
			var style = window.getComputedStyle(this);
			if (style.display === 'none' || style.visibility === 'hidden' || style.opacity === '0') {
				return false;
			}
			var rect = this.getBoundingClientRect();
			return rect.width > 0 && rect.height > 0;
		}
	`

	var result map[string]interface{}
	result, err = e.page.sendCommand(cdp.CmdRuntimeCallFunctionOn, map[string]interface{}{
		"functionDeclaration": visibilityCheck,
		"objectId":            objectID,
		"returnByValue":       true,
	})
	if err != nil {
		return false
	}

	var resultObj map[string]interface{}
	var ok bool
	resultObj, ok = result["result"].(map[string]interface{})
	if !ok {
		return false
	}

	var visible bool
	visible, ok = resultObj["value"].(bool)
	return ok && visible
}

// resolveObjectID returns the element's objectId, resolving from nodeId if needed.
func (e *Element) resolveObjectID() (string, error) {
	// Use cached objectId if available (go-rod pattern)
	if e.objectID != "" {
		return e.objectID, nil
	}

	var result map[string]interface{}
	var err error
	result, err = e.page.sendCommand("DOM.resolveNode", map[string]interface{}{
		"nodeId": e.nodeID,
	})
	if err != nil {
		return "", fmt.Errorf("DOM.resolveNode failed: %w", err)
	}

	var obj map[string]interface{}
	var ok bool
	obj, ok = result["object"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid resolveNode response")
	}

	var objectID string
	objectID, ok = obj["objectId"].(string)
	if !ok {
		return "", fmt.Errorf("no objectId in resolveNode response")
	}

	// Cache for future use
	e.objectID = objectID
	return objectID, nil
}

// Equals checks if two elements refer to the same DOM node.
func (e *Element) Equals(other *Element) bool {
	if e == nil {
		return other == nil
	}
	if other == nil {
		return false
	}
	return e.nodeID == other.nodeID
}
