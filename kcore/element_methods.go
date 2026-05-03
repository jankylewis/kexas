package kcore

import (
	"fmt"

	"github.com/jankylewis/kexas/errors"
	"github.com/jankylewis/kexas/internal/cdp"
)

// GetAttribute retrieves the value of the specified attribute from the element.
//
// Two implementation paths depending on which identifier the element carries:
//   - nodeID-backed (XPath path): CDP DOM.getAttributes (cheap; returns all attrs)
//   - objectID-backed (CSS path): Runtime.callFunctionOn with this.getAttribute(name)
//
// CSS-found elements only have an objectID — they used to silently fail this call
// (the older "nodeID <= 0 → ErrElementInvalidNodeID" guard rejected them). Now both
// paths work uniformly.
//
// If the attribute doesn't exist, returns an empty string and no error.
func (e *Element) GetAttribute(name string) (string, error) {
	if e == nil {
		return "", errors.ErrElementNil
	}
	if e.page == nil {
		return "", errors.ErrElementNoPage
	}
	if e.nodeID <= 0 && e.objectID == "" {
		return "", errors.ErrElementInvalidNodeID
	}
	if name == "" {
		return "", fmt.Errorf("attribute name cannot be empty")
	}

	if e.objectID != "" {
		return e.getAttributeViaObjectID(name)
	}
	return e.getAttributeViaNodeID(name)
}

// getAttributeViaNodeID is the original DOM.getAttributes path for nodeID-backed
// elements (XPath selector results).
func (e *Element) getAttributeViaNodeID(name string) (string, error) {
	var params map[string]interface{} = map[string]interface{}{
		"nodeId": e.nodeID,
	}

	var result map[string]interface{}
	var err error
	result, err = e.page.sendCommand(cdp.CmdDOMGetAttributes, params)
	if err != nil {
		return "", fmt.Errorf("failed to get attributes: %w", err)
	}

	var attributes []interface{}
	var ok bool
	attributes, ok = result["attributes"].([]interface{})
	if !ok {
		return "", fmt.Errorf("invalid attributes response")
	}

	for i := 0; i < len(attributes)-1; i += 2 {
		var attrName string
		attrName, ok = attributes[i].(string)
		if !ok {
			continue
		}
		if attrName == name {
			var value string
			value, ok = attributes[i+1].(string)
			if ok {
				return value, nil
			}
		}
	}
	return "", nil
}

// getAttributeViaObjectID handles CSS-selector-found elements which carry
// objectID but no nodeID. Calls Element.getAttribute via Runtime.callFunctionOn.
func (e *Element) getAttributeViaObjectID(name string) (string, error) {
	var fn string = `function(name) { return this.getAttribute(name); }`
	var result map[string]interface{}
	var err error
	result, err = e.page.sendCommand(cdp.CmdRuntimeCallFunctionOn, map[string]interface{}{
		"functionDeclaration": fn,
		"objectId":            e.objectID,
		"arguments":           []interface{}{map[string]interface{}{"value": name}},
		"returnByValue":       true,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get attribute via objectID: %w", err)
	}

	var resultObj map[string]interface{}
	var ok bool
	resultObj, ok = result["result"].(map[string]interface{})
	if !ok {
		return "", nil
	}
	var value interface{}
	value, ok = resultObj["value"]
	if !ok || value == nil {
		return "", nil
	}
	var s string
	s, ok = value.(string)
	if !ok {
		return "", nil
	}
	return s, nil
}

// GetText retrieves the text content of the element.
//
// This returns the visible text content of the element, excluding any
// text from hidden child elements.
func (e *Element) GetText() (string, error) {
	if e == nil {
		return "", errors.ErrElementNil
	}

	if e.page == nil {
		return "", errors.ErrElementNoPage
	}

	if e.nodeID <= 0 && e.objectID == "" {
		return "", errors.ErrElementInvalidNodeID
	}

	// Resolve objectId for callFunctionOn
	var objectID string
	var err error
	objectID, err = e.resolveObjectID()
	if err != nil {
		return "", errors.ResolveElementFailed(err)
	}

	// Use Runtime.callFunctionOn to get innerText (visible text only)
	var result map[string]interface{}
	result, err = e.page.sendCommand(cdp.CmdRuntimeCallFunctionOn, map[string]interface{}{
		"functionDeclaration": `function() { return this.innerText || this.textContent || ''; }`,
		"objectId":            objectID,
		"returnByValue":       true,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get text: %w", err)
	}

	var text string
	var ok bool
	if r, ok2 := result["result"].(map[string]interface{}); ok2 {
		text, ok = r["value"].(string)
	}
	if !ok {
		return "", fmt.Errorf("invalid text response")
	}

	return text, nil
}

// Validate checks if the element is still valid and attached to the DOM.
//
// Elements can become invalid if the DOM is reloaded or the element
// is removed from the page.
func (e *Element) Validate() error {
	if e == nil {
		return errors.ErrElementNil
	}

	if e.page == nil {
		return errors.ErrElementNoPage
	}

	if e.nodeID <= 0 {
		return errors.ErrElementInvalidNodeID
	}

	// Use CDP DOM.describeNode to check if the node still exists
	var params map[string]interface{} = map[string]interface{}{
		"nodeId": e.nodeID,
	}

	var result map[string]interface{}
	var err error
	result, err = e.page.sendCommand(cdp.CmdDOMDescribeNode, params)
	if err != nil {
		return fmt.Errorf("failed to describe node: %w", err)
	}

	var node map[string]interface{}
	var ok bool
	node, ok = result["node"].(map[string]interface{})
	if !ok {
		return errors.ErrElementInvalidNodeID
	}

	var nodeID int64
	nodeID, ok = node["nodeId"].(int64)
	if !ok || nodeID != int64(e.nodeID) {
		return errors.ErrElementInvalidNodeID
	}

	return nil
}
