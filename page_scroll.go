package kexas

import (
	"fmt"

	"github.com/kexas-project/kexas/errors"
	"github.com/kexas-project/kexas/internal/cdp"
)

// ScrollToTop scrolls the page to the very top (0, 0).
//
// Equivalent to window.scrollTo(0, 0) in JavaScript.
// This is a page-level scroll action.
func (p *Page) ScrollToTop() error {
	if p == nil {
		return errors.ErrPageNil
	}

	p.log.Info("scrolling to top")

	var err error
	_, err = p.sendCommand(cdp.CmdRuntimeEvaluate, map[string]interface{}{
		"expression":    "window.scrollTo(0, 0)",
		"returnByValue": true,
	})
	if err != nil {
		p.log.Error("scroll to top failed", "error", err)
		return fmt.Errorf("scroll to top failed: %w", err)
	}

	return nil
}

// ScrollToBottom scrolls the page to the very bottom.
//
// Equivalent to window.scrollTo(0, document.body.scrollHeight) in JavaScript.
// This is a page-level scroll action.
func (p *Page) ScrollToBottom() error {
	if p == nil {
		return errors.ErrPageNil
	}

	p.log.Info("scrolling to bottom")

	var err error
	_, err = p.sendCommand(cdp.CmdRuntimeEvaluate, map[string]interface{}{
		"expression":    "window.scrollTo(0, document.body.scrollHeight)",
		"returnByValue": true,
	})
	if err != nil {
		p.log.Error("scroll to bottom failed", "error", err)
		return fmt.Errorf("scroll to bottom failed: %w", err)
	}

	return nil
}

// ScrollBy scrolls the page by a relative amount of pixels.
//
// Positive x scrolls right, negative x scrolls left.
// Positive y scrolls down, negative y scrolls up.
// Both x and y must not be zero simultaneously.
//
// Equivalent to window.scrollBy(x, y) in JavaScript.
func (p *Page) ScrollBy(x int, y int) error {
	if p == nil {
		return errors.ErrPageNil
	}

	if x == 0 && y == 0 {
		return errors.ErrScrollInvalidPixels
	}

	p.log.Info("scrolling by pixels", "x", x, "y", y)

	var expression string = fmt.Sprintf("window.scrollBy(%d, %d)", x, y)

	var err error
	_, err = p.sendCommand(cdp.CmdRuntimeEvaluate, map[string]interface{}{
		"expression":    expression,
		"returnByValue": true,
	})
	if err != nil {
		p.log.Error("scroll by failed", "error", err, "x", x, "y", y)
		return fmt.Errorf("scroll by (%d, %d) failed: %w", x, y, err)
	}

	return nil
}

// ScrollPosition returns the current scroll position of the page as (x, y).
//
// Equivalent to [window.scrollX, window.scrollY] in JavaScript.
func (p *Page) ScrollPosition() (int, int, error) {
	if p == nil {
		return 0, 0, errors.ErrPageNil
	}

	var result map[string]interface{}
	var err error
	result, err = p.sendCommand(cdp.CmdRuntimeEvaluate, map[string]interface{}{
		"expression":    "JSON.stringify({x: Math.round(window.scrollX), y: Math.round(window.scrollY)})",
		"returnByValue": true,
	})
	if err != nil {
		return 0, 0, fmt.Errorf("get scroll position failed: %w", err)
	}

	var resultObj map[string]interface{}
	var ok bool
	resultObj, ok = result["result"].(map[string]interface{})
	if !ok {
		return 0, 0, fmt.Errorf("invalid scroll position response")
	}

	var jsonStr string
	jsonStr, ok = resultObj["value"].(string)
	if !ok {
		return 0, 0, fmt.Errorf("invalid scroll position value")
	}

	var scrollX int
	var scrollY int
	var n int
	n, err = fmt.Sscanf(jsonStr, `{"x":%d,"y":%d}`, &scrollX, &scrollY)
	if err != nil || n != 2 {
		return 0, 0, fmt.Errorf("failed to parse scroll position: %s", jsonStr)
	}

	return scrollX, scrollY, nil
}
