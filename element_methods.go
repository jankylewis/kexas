package kexas

import (
	"fmt"

	"github.com/kexas-project/kexas/errors"
	"github.com/kexas-project/kexas/internal/cdp"
)

// GetAttribute retrieves the value of the specified attribute from the element.
//
// If the attribute doesn't exist, returns an empty string and no error.
func (e *Element) GetAttribute(name string) (string, error) {
	if e == nil {
		return "", errors.ErrElementNil
	}

	if e.page == nil {
		return "", errors.ErrElementNoPage
	}

	if e.nodeID <= 0 {
		return "", errors.ErrElementInvalidNodeID
	}

	if name == "" {
		return "", fmt.Errorf("attribute name cannot be empty")
	}

	// Use CDP DOM.getAttributes to get all attributes
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

	// Attributes are returned as alternating name/value pairs
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

	// Attribute not found
	return "", nil
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
		return "", fmt.Errorf("failed to resolve element: %w", err)
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
