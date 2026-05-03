package kcore

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jankylewis/kexas/internal/cdp"
)

// FindAll returns every element matching the selector. Companion to Find,
// which strict-matches a single element.
//
// Selector type is auto-detected the same way as Find: leading `//` or `(`
// → XPath, anything else → CSS. ID-shortcut form (`#x`) is treated as CSS.
//
// Behavior:
//   - Returns the matched elements in document order.
//   - Returns an empty slice (not an error) when nothing matches.
//   - Does NOT auto-wait for elements to appear (unlike Find's 7s retry).
//     For "wait for at least N to be present", combine with WaitForElement
//     on a parent or use the underlying Page.Evaluate yourself.
//
// Each returned *Element is bound to a Runtime objectId (no nodeID), so
// methods that depend on nodeID (rare) need the objectID-dispatch path.
func (p *Page) FindAll(selector string) ([]*Element, error) {
	if p == nil || p.log == nil {
		return nil, fmt.Errorf("FindAll: nil page")
	}
	if selector == "" {
		return nil, fmt.Errorf("FindAll: empty selector")
	}

	p.log.Debug("FindAll", "selector", selector)

	if strings.HasPrefix(selector, "//") || strings.HasPrefix(selector, "(") {
		return p.findAllByXPath(selector)
	}
	return p.findAllByCSS(selector)
}

// findAllByCSS resolves selector via document.querySelectorAll, then
// enumerates each element's remote objectId.
func (p *Page) findAllByCSS(selector string) ([]*Element, error) {
	var jsExpr string = fmt.Sprintf(`Array.from(document.querySelectorAll(%q))`, selector)
	return p.findAllFromArrayExpression(selector, jsExpr)
}

// findAllByXPath uses document.evaluate's snapshot iterator to collect every
// matching node into a JS array, then enumerates objectIds. XPath constant 7
// is XPathResult.ORDERED_NODE_SNAPSHOT_TYPE — gives us indexable, document-
// ordered results.
func (p *Page) findAllByXPath(xpath string) ([]*Element, error) {
	var jsExpr string = fmt.Sprintf(`(function() {
		const snap = document.evaluate(%q, document, null, 7, null);
		const arr = [];
		for (let i = 0; i < snap.snapshotLength; i++) arr.push(snap.snapshotItem(i));
		return arr;
	})()`, xpath)
	return p.findAllFromArrayExpression(xpath, jsExpr)
}

// findAllFromArrayExpression evaluates jsExpr (which must return a JS Array of
// element nodes), then uses Runtime.getProperties to walk the array and pull
// each element's objectId. Wraps each into a kexas *Element.
func (p *Page) findAllFromArrayExpression(selector, jsExpr string) ([]*Element, error) {
	var arrayObjectID string
	var err error
	arrayObjectID, err = p.evaluateToObjectID(jsExpr)
	if err != nil {
		return nil, fmt.Errorf("FindAll evaluate failed for %s: %w", selector, err)
	}
	if arrayObjectID == "" {
		// Empty array serialises to [] without an objectId. Return empty slice.
		return []*Element{}, nil
	}

	var props map[string]interface{}
	props, err = p.sendCommand(cdp.CmdRuntimeGetProperties, map[string]interface{}{
		"objectId":      arrayObjectID,
		"ownProperties": true,
	})
	if err != nil {
		return nil, fmt.Errorf("FindAll getProperties failed for %s: %w", selector, err)
	}

	return wrapArrayPropertiesAsElements(p, selector, props), nil
}

// evaluateToObjectID runs jsExpr via Runtime.evaluate and returns the result's
// remote objectId. Returns ("", nil) when the result is null/undefined or a
// by-value primitive (e.g., empty array serialised inline). Returns an error
// only on CDP failure or JS exception.
func (p *Page) evaluateToObjectID(jsExpr string) (string, error) {
	var result map[string]interface{}
	var err error
	result, err = p.sendCommand(cdp.CmdRuntimeEvaluate, map[string]interface{}{
		"expression":    jsExpr,
		"returnByValue": false,
	})
	if err != nil {
		return "", err
	}
	if exDesc, hasEx := result["exceptionDetails"]; hasEx {
		return "", fmt.Errorf("JS exception: %v", exDesc)
	}

	var resultObj map[string]interface{}
	var ok bool
	resultObj, ok = result["result"].(map[string]interface{})
	if !ok {
		return "", nil
	}

	var objectID string
	objectID, _ = resultObj["objectId"].(string)
	return objectID, nil
}

// indexedObjectID pairs an element's array index with its CDP remote handle.
// Used to preserve original array order through the unordered property walk.
type indexedObjectID struct {
	idx      int
	objectID string
}

// wrapArrayPropertiesAsElements walks the property descriptors returned by
// Runtime.getProperties on a JS array, returning elements in original order.
func wrapArrayPropertiesAsElements(p *Page, selector string, props map[string]interface{}) []*Element {
	var rawProps []interface{}
	var ok bool
	rawProps, ok = props["result"].([]interface{})
	if !ok {
		return []*Element{}
	}
	var collected []indexedObjectID = collectIndexedObjectIDs(rawProps)
	insertionSortByIndex(collected)
	return objectIDsToElements(p, selector, collected)
}

// collectIndexedObjectIDs filters property descriptors to numeric-indexed
// entries (0, 1, 2…) and extracts each element's objectId. Skips "length",
// "__proto__", and other named noise.
func collectIndexedObjectIDs(rawProps []interface{}) []indexedObjectID {
	var collected []indexedObjectID = make([]indexedObjectID, 0, len(rawProps))
	for _, raw := range rawProps {
		prop, propOk := raw.(map[string]interface{})
		if !propOk {
			continue
		}
		var name string
		name, _ = prop["name"].(string)
		var idx int
		var convErr error
		idx, convErr = strconv.Atoi(name)
		if convErr != nil {
			continue
		}
		valueObj, vOk := prop["value"].(map[string]interface{})
		if !vOk {
			continue
		}
		objectID, oOk := valueObj["objectId"].(string)
		if !oOk {
			continue
		}
		collected = append(collected, indexedObjectID{idx: idx, objectID: objectID})
	}
	return collected
}

// insertionSortByIndex sorts in place. Small N (typically <100), and stdlib
// `sort` would force a goroutine-allocating closure for a trivial comparator.
func insertionSortByIndex(collected []indexedObjectID) {
	for i := 1; i < len(collected); i++ {
		for j := i; j > 0 && collected[j-1].idx > collected[j].idx; j-- {
			collected[j-1], collected[j] = collected[j], collected[j-1]
		}
	}
}

// objectIDsToElements wraps each collected objectID into a *Element bound to
// the page. The same selector is used for every element's debug label.
func objectIDsToElements(p *Page, selector string, collected []indexedObjectID) []*Element {
	var elements []*Element = make([]*Element, 0, len(collected))
	for _, c := range collected {
		elements = append(elements, NewElementWithObject(p, selector, 0, c.objectID, 10*time.Second))
	}
	return elements
}
