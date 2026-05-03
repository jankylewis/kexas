package kcore

import (
	"fmt"
	"time"

	"github.com/jankylewis/kexas/errors"
	"github.com/jankylewis/kexas/internal/cdp"
)

func (p *Page) findByXPath(xpath string) (*Element, error) {
	p.log.Debug("finding element by XPath", "xpath", xpath)

	var searchID string
	var resultCount int
	var err error
	searchID, resultCount, err = p.xpathPerformSearch(xpath)
	if err != nil {
		return nil, err
	}
	if resultCount == 0 {
		return nil, errors.ElementNotFound(xpath)
	}
	if resultCount > 1 {
		return nil, fmt.Errorf(
			"xpath %s matched %d elements, expected exactly 1",
			xpath, resultCount,
		)
	}

	var nodeID cdp.NodeID
	nodeID, err = p.xpathFetchFirstNodeID(searchID, xpath)
	if err != nil {
		return nil, err
	}

	var element *Element = NewElement(p, xpath, nodeID, 10*time.Second)
	p.log.Debug("element found by XPath", "xpath", xpath, "nodeID", nodeID)
	return element, nil
}

// xpathPerformSearch issues DOM.performSearch and returns the searchID and match count.
func (p *Page) xpathPerformSearch(xpath string) (string, int, error) {
	var params map[string]interface{} = map[string]interface{}{
		"query": xpath,
	}
	var result map[string]interface{}
	var err error
	result, err = p.sendCommand(cdp.CmdDOMPerformSearch, params)
	if err != nil {
		return "", 0, fmt.Errorf("XPath search failed: %w", err)
	}

	var searchID string
	var ok bool
	searchID, ok = result["searchId"].(string)
	if !ok {
		return "", 0, fmt.Errorf("invalid search ID response")
	}

	var resultCount float64
	resultCount, ok = result["resultCount"].(float64)
	if !ok {
		return searchID, 0, nil
	}
	return searchID, int(resultCount), nil
}

// xpathFetchFirstNodeID retrieves the first node ID from a previously-issued performSearch.
func (p *Page) xpathFetchFirstNodeID(searchID, xpath string) (cdp.NodeID, error) {
	var results map[string]interface{}
	var err error
	results, err = p.sendCommand(cdp.CmdDOMGetSearchResults, map[string]interface{}{
		"searchId":  searchID,
		"fromIndex": 0,
		"toIndex":   1,
	})
	if err != nil {
		return 0, fmt.Errorf("get search results failed: %w", err)
	}

	var nodeIDs []interface{}
	var ok bool
	nodeIDs, ok = results["nodeIds"].([]interface{})
	if !ok || len(nodeIDs) == 0 {
		return 0, errors.ElementNotFound(xpath)
	}

	var nodeIDFloat float64
	nodeIDFloat, ok = nodeIDs[0].(float64)
	if !ok {
		return 0, fmt.Errorf("invalid node ID type")
	}
	return cdp.NodeID(int64(nodeIDFloat)), nil
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
