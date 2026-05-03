package kcore

import (
	"encoding/base64"
	"fmt"

	"github.com/jankylewis/kexas/internal/cdp"
)

// URL returns the current URL of the page.
func (p *Page) URL() (string, error) {
	var result map[string]interface{}
	var err error
	result, err = p.sendCommand(cdp.CmdRuntimeEvaluate, map[string]interface{}{
		"expression": "window.location.href",
	})
	if err != nil {
		return "", fmt.Errorf("failed to get URL: %w", err)
	}

	var url string
	var ok bool
	url, ok = result["result"].(map[string]interface{})["value"].(string)
	if !ok {
		return "", fmt.Errorf("invalid URL response")
	}

	return url, nil
}

// Screenshot captures a screenshot of the current page.
func (p *Page) Screenshot() ([]byte, error) {
	p.log.Debug("capturing screenshot")

	var params map[string]interface{} = map[string]interface{}{
		"format": "png",
	}

	var result map[string]interface{}
	var err error
	result, err = p.sendCommand(cdp.CmdPageCaptureScreenshot, params)
	if err != nil {
		p.log.Error("screenshot failed", "error", err)
		return nil, fmt.Errorf("screenshot failed: %w", err)
	}

	var data string
	var ok bool
	data, ok = result["data"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid screenshot response")
	}

	// Decode base64
	var screenshot []byte
	screenshot, err = base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode screenshot: %w", err)
	}

	p.log.Info("screenshot captured", "size", len(screenshot))
	return screenshot, nil
}

// Title returns the title of the current page.
func (p *Page) Title() (string, error) {
	var result map[string]interface{}
	var err error
	result, err = p.sendCommand(cdp.CmdRuntimeEvaluate, map[string]interface{}{
		"expression": "document.title",
	})
	if err != nil {
		return "", fmt.Errorf("failed to get title: %w", err)
	}

	var title string
	var ok bool
	title, ok = result["result"].(map[string]interface{})["value"].(string)
	if !ok {
		return "", fmt.Errorf("invalid title response")
	}

	return title, nil
}

// SetContent sets the HTML content of the page using CDP Page.setDocumentContent.
// This is useful for testing without requiring navigation.
func (p *Page) SetContent(html string) error {
	// Get the frame ID from the page's target
	var frameResult map[string]interface{}
	var err error
	frameResult, err = p.sendCommand(cdp.CmdPageGetFrameTree, nil)
	if err != nil {
		return fmt.Errorf("failed to get frame tree: %w", err)
	}

	var frameTree map[string]interface{}
	var ok bool
	frameTree, ok = frameResult["frameTree"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid frame tree response")
	}

	var frame map[string]interface{}
	frame, ok = frameTree["frame"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid frame response")
	}

	var frameID string
	frameID, ok = frame["id"].(string)
	if !ok {
		return fmt.Errorf("invalid frame ID")
	}

	_, err = p.sendCommand(cdp.CmdPageSetDocumentContent, map[string]interface{}{
		"frameId": frameID,
		"html":    html,
	})
	if err != nil {
		return fmt.Errorf("set document content failed: %w", err)
	}

	return nil
}

// Evaluate executes a JavaScript expression and returns the result as an interface{}.
func (p *Page) Evaluate(expression string) (interface{}, error) {
	var result map[string]interface{}
	var err error
	result, err = p.sendCommand(cdp.CmdRuntimeEvaluate, map[string]interface{}{
		"expression":    expression,
		"returnByValue": true,
	})
	if err != nil {
		return nil, fmt.Errorf("evaluate failed: %w", err)
	}

	var resultObj map[string]interface{}
	var ok bool
	resultObj, ok = result["result"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid evaluate response")
	}

	return resultObj["value"], nil
}
