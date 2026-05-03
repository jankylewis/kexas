package kcore

import (
	"fmt"
	"strings"
	"time"

	"github.com/jankylewis/kexas/errors"
	"github.com/jankylewis/kexas/internal/cdp"
)

// This method supports ID selectors, CSS selectors, and XPath selectors.
// It automatically detects the selector type and uses the appropriate lookup method.
// Auto-retries for up to 10 seconds to handle SPA rendering delays.
// Returns the element when found, or error if not found after timeout.
func (p *Page) Find(selector string) (*Element, error) {
	if p == nil || p.log == nil {
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
		elem, err = p.dispatchFind(selector)
		if err == nil && elem != nil {
			return elem, nil
		}
		lastErr = err
		time.Sleep(pollInterval)
	}
	p.logFindTimeout(selector, timeout, time.Since(start))
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("element %s not found after %v", selector, timeout)
}

// dispatchFind routes to the appropriate finder based on selector form.
// Same dispatch as FindAll: `#x` → ID-shortcut → ID, `//`/`(` → XPath, else CSS.
func (p *Page) dispatchFind(selector string) (*Element, error) {
	if strings.HasPrefix(selector, "#") {
		return p.findByID(selector[1:])
	}
	if strings.HasPrefix(selector, "//") || strings.HasPrefix(selector, "(") {
		return p.findByXPath(selector)
	}
	return p.findByCSS(selector)
}

// logFindTimeout records a timeout-with-pageTitle observation. Title is best-
// effort; failures are silent because the find error itself is more important.
func (p *Page) logFindTimeout(selector string, timeout time.Duration, elapsed time.Duration) {
	var titleResult map[string]interface{}
	titleResult, _ = p.sendCommand(cdp.CmdRuntimeEvaluate, map[string]interface{}{
		"expression":    "document.title",
		"returnByValue": true,
	})
	var pageTitle string
	if tr, ok2 := titleResult["result"].(map[string]interface{}); ok2 {
		pageTitle, _ = tr["value"].(string)
	}
	p.log.Info("element not found after timeout", "selector", selector, "timeout", timeout, "elapsed", elapsed, "pageTitle", pageTitle)
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

	var matchCount int
	var err error
	matchCount, err = p.cssMatchCount(selector)
	if err != nil {
		return nil, err
	}
	if matchCount == 0 {
		return nil, errors.ElementNotFound(selector)
	}
	if matchCount > 1 {
		return nil, fmt.Errorf(
			"selector %s matched %d elements, expected exactly 1",
			selector, matchCount,
		)
	}

	var objectID string
	objectID, err = p.cssQuerySingleObjectID(selector)
	if err != nil {
		return nil, err
	}

	var element *Element = NewElementWithObject(p, selector, 0, objectID, 10*time.Second)
	p.log.Debug("element found by CSS", "selector", selector, "objectId", objectID)
	return element, nil
}

// cssMatchCount returns the number of elements matching the CSS selector.
func (p *Page) cssMatchCount(selector string) (int, error) {
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
		return 0, fmt.Errorf("CSS count check failed: %w", err)
	}
	return extractMatchCount(countResult), nil
}

// cssQuerySingleObjectID resolves the selector via querySelector and returns the
// remote objectId. Returns ElementNotFound for null/exception/undefined results.
func (p *Page) cssQuerySingleObjectID(selector string) (string, error) {
	var jsExpr string = fmt.Sprintf(`document.querySelector(%q)`, selector)
	var result map[string]interface{}
	var err error
	result, err = p.sendCommand(cdp.CmdRuntimeEvaluate, map[string]interface{}{
		"expression":    jsExpr,
		"returnByValue": false,
	})
	if err != nil {
		return "", fmt.Errorf("CSS query failed: %w", err)
	}

	if exDesc, hasEx := result["exceptionDetails"]; hasEx {
		p.log.Debug("findByCSS: JS exception", "selector", selector, "exception", exDesc)
		return "", errors.ElementNotFound(selector)
	}

	var resultObj map[string]interface{}
	var ok bool
	resultObj, ok = result["result"].(map[string]interface{})
	if !ok {
		return "", errors.ElementNotFound(selector)
	}

	var resultType string
	resultType, _ = resultObj["type"].(string)
	var subtype string
	subtype, _ = resultObj["subtype"].(string)
	if subtype == "null" || resultType == "undefined" {
		return "", errors.ElementNotFound(selector)
	}

	var objectID string
	objectID, ok = resultObj["objectId"].(string)
	if !ok {
		p.log.Debug("findByCSS: no objectId in result", "type", resultType, "subtype", subtype, "result", resultObj)
		return "", errors.ElementNotFound(selector)
	}
	return objectID, nil
}

// findByXPath finds an element using XPath selector.
