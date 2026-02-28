package kexas

import (
	"fmt"
	"strings"
	"time"

	"github.com/kexas-project/kexas/errors"
	"github.com/kexas-project/kexas/internal/cdp"
)

// Find finds an element on the page using the given selector.
//
// This method supports ID selectors, CSS selectors, and XPath selectors.
// It automatically detects the selector type and uses the appropriate lookup method.
// Auto-retries for up to 10 seconds to handle SPA rendering delays.
// Returns the element when found, or error if not found after timeout.
func (p *Page) Find(selector string) (*Element, error) {
	// Handle nil page gracefully for testing
	if p == nil || p.log == nil {
		// For testing, just return error immediately
		return nil, fmt.Errorf("element %s not found", selector)
	}

	p.log.Debug("finding element", "selector", selector)

	var timeout time.Duration = 7 * time.Second
	var pollInterval time.Duration = 200 * time.Millisecond
	var start time.Time = time.Now()
	var lastErr error

	for time.Since(start) < timeout {
		var elem *Element
		var err error

		// Detect selector type and use appropriate method
		if strings.HasPrefix(selector, "#") {
			var id string = selector[1:]
			elem, err = p.findByID(id)
		} else if strings.HasPrefix(selector, "//") || strings.HasPrefix(selector, "(") {
			elem, err = p.findByXPath(selector)
		} else {
			elem, err = p.findByCSS(selector)
		}

		if err == nil && elem != nil {
			return elem, nil
		}

		lastErr = err
		time.Sleep(pollInterval)
	}

	// Log page title for context when element not found
	var titleResult map[string]interface{}
	titleResult, _ = p.sendCommand(cdp.CmdRuntimeEvaluate, map[string]interface{}{
		"expression":    "document.title",
		"returnByValue": true,
	})
	var pageTitle string
	if tr, ok2 := titleResult["result"].(map[string]interface{}); ok2 {
		pageTitle, _ = tr["value"].(string)
	}
	p.log.Info("element not found after timeout", "selector", selector, "timeout", timeout, "elapsed", time.Since(start), "pageTitle", pageTitle)
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("element %s not found after %v", selector, timeout)
}

// findByID finds an element by its ID attribute.
// Uses Runtime.evaluate + DOM.requestNode for reliability across navigations.
func (p *Page) findByID(id string) (*Element, error) {
	p.log.Debug("finding element by ID", "id", id)

	// Use document.getElementById via Runtime.evaluate (more reliable than DOM.performSearch)
	var jsExpr string = fmt.Sprintf(`document.getElementById(%q)`, id)
	var result map[string]interface{}
	var err error
	result, err = p.sendCommand(cdp.CmdRuntimeEvaluate, map[string]interface{}{
		"expression":    jsExpr,
		"returnByValue": false,
	})
	if err != nil {
		return nil, fmt.Errorf("ID search failed: %w", err)
	}

	var resultObj map[string]interface{}
	var ok bool
	resultObj, ok = result["result"].(map[string]interface{})
	if !ok {
		return nil, errors.ElementNotFound("#" + id)
	}

	var subtype string
	subtype, _ = resultObj["subtype"].(string)
	if subtype == "null" {
		return nil, errors.ElementNotFound("#" + id)
	}

	var objectID string
	objectID, ok = resultObj["objectId"].(string)
	if !ok {
		return nil, errors.ElementNotFound("#" + id)
	}

	// Create element with objectId (go-rod pattern: objectId is sufficient for all interactions)
	var element *Element = NewElementWithObject(p, "#"+id, 0, objectID, 10*time.Second)

	p.log.Debug("element found by ID", "id", id, "objectId", objectID)
	return element, nil
}

// findByCSS finds an element using CSS selector.
// Returns error if the selector matches zero or more than one element.
func (p *Page) findByCSS(selector string) (*Element, error) {
	p.log.Debug("finding element by CSS", "selector", selector)

	// First, check match count using querySelectorAll
	var countExpr string = fmt.Sprintf(
		`document.querySelectorAll(%q).length`, selector,
	)
	var countResult map[string]interface{}
	var err error
	countResult, err = p.sendCommand(cdp.CmdRuntimeEvaluate, map[string]interface{}{
		"expression":    countExpr,
		"returnByValue": true,
	})
	if err != nil {
		return nil, fmt.Errorf("CSS count check failed: %w", err)
	}

	var matchCount int = extractMatchCount(countResult)
	if matchCount == 0 {
		return nil, errors.ElementNotFound(selector)
	}
	if matchCount > 1 {
		return nil, fmt.Errorf(
			"selector %s matched %d elements, expected exactly 1",
			selector, matchCount,
		)
	}

	// Exactly 1 match — retrieve it with querySelector
	var jsExpr string = fmt.Sprintf(`document.querySelector(%q)`, selector)
	var result map[string]interface{}
	result, err = p.sendCommand(cdp.CmdRuntimeEvaluate, map[string]interface{}{
		"expression":    jsExpr,
		"returnByValue": false,
	})
	if err != nil {
		return nil, fmt.Errorf("CSS query failed: %w", err)
	}

	// Check if element was found
	var resultObj map[string]interface{}
	var ok bool
	resultObj, ok = result["result"].(map[string]interface{})
	if !ok {
		return nil, errors.ElementNotFound(selector)
	}

	// Check the type/subtype of the result
	var resultType string
	resultType, _ = resultObj["type"].(string)
	var subtype string
	subtype, _ = resultObj["subtype"].(string)

	// Check for exception (e.g. execution context destroyed)
	if exDesc, hasEx := result["exceptionDetails"]; hasEx {
		p.log.Debug("findByCSS: JS exception", "selector", selector, "exception", exDesc)
		return nil, errors.ElementNotFound(selector)
	}

	if subtype == "null" || resultType == "undefined" {
		return nil, errors.ElementNotFound(selector)
	}

	// Get the remote object ID
	var objectID string
	objectID, ok = resultObj["objectId"].(string)
	if !ok {
		p.log.Debug("findByCSS: no objectId in result", "type", resultType, "subtype", subtype, "result", resultObj)
		return nil, errors.ElementNotFound(selector)
	}

	// Create element with objectId (go-rod pattern: objectId is sufficient for all interactions)
	var element *Element = NewElementWithObject(p, selector, 0, objectID, 10*time.Second)

	p.log.Debug("element found by CSS", "selector", selector, "objectId", objectID)
	return element, nil
}

// findByXPath finds an element using XPath selector.
func (p *Page) findByXPath(xpath string) (*Element, error) {
	p.log.Debug("finding element by XPath", "xpath", xpath)

	var params map[string]interface{} = map[string]interface{}{
		"query": xpath,
	}

	var result map[string]interface{}
	var err error
	result, err = p.sendCommand(cdp.CmdDOMPerformSearch, params)
	if err != nil {
		return nil, fmt.Errorf("XPath search failed: %w", err)
	}

	var searchID string
	var ok bool
	searchID, ok = result["searchId"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid search ID response")
	}

	// Check resultCount before fetching results
	var resultCount float64
	resultCount, ok = result["resultCount"].(float64)
	if !ok || resultCount == 0 {
		return nil, errors.ElementNotFound(xpath)
	}

	// Validate match count — error if more than 1 match
	if resultCount > 1 {
		return nil, fmt.Errorf(
			"xpath %s matched %d elements, expected exactly 1",
			xpath, int(resultCount),
		)
	}

	// Get search results
	var results map[string]interface{}
	results, err = p.sendCommand(cdp.CmdDOMGetSearchResults, map[string]interface{}{
		"searchId":  searchID,
		"fromIndex": 0,
		"toIndex":   1,
	})
	if err != nil {
		return nil, fmt.Errorf("get search results failed: %w", err)
	}

	var nodeIDs []interface{}
	nodeIDs, ok = results["nodeIds"].([]interface{})
	if !ok || len(nodeIDs) == 0 {
		return nil, errors.ElementNotFound(xpath)
	}

	// Convert nodeID to cdp.NodeID (CDP returns nodeIds as numbers)
	var nodeIDFloat float64
	nodeIDFloat, ok = nodeIDs[0].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid node ID type")
	}

	var nodeID cdp.NodeID = cdp.NodeID(int64(nodeIDFloat))

	// Create element with default timeout
	var element *Element = NewElement(p, xpath, nodeID, 10*time.Second)

	p.log.Debug("element found by XPath", "xpath", xpath, "nodeID", nodeID)
	return element, nil
}

// extractMatchCount extracts an integer count from a CDP Runtime.evaluate result.
// Returns 0 if the result cannot be parsed.
func extractMatchCount(result map[string]interface{}) int {
	var resultObj map[string]interface{}
	var ok bool
	resultObj, ok = result["result"].(map[string]interface{})
	if !ok {
		return 0
	}

	var value float64
	value, ok = resultObj["value"].(float64)
	if !ok {
		return 0
	}

	return int(value)
}

// FindByXPath finds an element using XPath selector.
//
// This method searches for elements using XPath expressions.
// Use this for complex selections that CSS selectors cannot handle.
func (p *Page) FindByXPath(xpath string) (*Element, error) {
	return p.findByXPath(xpath)
}
