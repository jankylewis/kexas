package kexas

import (
	"context"
	"fmt"
	"time"

	"github.com/kexas-project/kexas/internal/logger"
	"github.com/kexas-project/kexas/kwait"
)

// Page represents a browser page (tab).
type Page struct {
	browser   *Browser
	targetID  string
	sessionID string
	log       *logger.Logger
	ctx       context.Context
}

// Navigate navigates the page to the given URL.
// Optional waitUntil parameter controls when navigation is considered complete.
// If not specified, defaults to WaitUntilLoad.
func (p *Page) Navigate(url string, waitUntil ...kwait.WaitUntil) error {
	p.log.Debug("navigating", "url", url)

	var params map[string]interface{} = map[string]interface{}{
		"url": url,
	}

	var result map[string]interface{}
	var err error
	result, err = p.sendCommand("Page.navigate", params)
	if err != nil {
		p.log.Error("navigation failed", "url", url, "err", err)
		return fmt.Errorf("kexas: navigation failed: %w", err)
	}

	var frameID string
	var ok bool
	frameID, ok = result["frameId"].(string)
	if !ok {
		return fmt.Errorf("kexas: invalid frameId in navigation response")
	}

	p.log.Info("navigated successfully", "url", url, "frameId", frameID)

	// Determine wait strategy (default to load)
	var strategy kwait.WaitUntil = kwait.WaitUntilLoad
	if len(waitUntil) > 0 {
		strategy = waitUntil[0]
	}

	// Wait for navigation to complete
	err = p.WaitForLoadState(strategy, 5*time.Second)
	if err != nil {
		p.log.Warn("navigation wait timeout", "url", url, "waitUntil", strategy, "err", err)
		// Don't fail navigation, just log the warning
	}

	return nil
}

// WaitForLoadState waits for the page to reach a specific load state.
func (p *Page) WaitForLoadState(waitUntil kwait.WaitUntil, timeout time.Duration) error {
	p.log.Debug("waiting for load state", "waitUntil", waitUntil, "timeout", timeout)

	var err error = kwait.ForPageLoad(p.ctx, p.Title, waitUntil, timeout)
	if err != nil {
		return fmt.Errorf("kexas: %w", err)
	}

	p.log.Debug("load state reached", "waitUntil", waitUntil)
	return nil
}

// WaitForNavigationCompleted waits for the page to finish loading after navigation.
// Deprecated: Use WaitForLoadState instead.
func (p *Page) WaitForNavigationCompleted(timeout time.Duration) error {
	return p.WaitForLoadState(kwait.WaitUntilLoad, timeout)
}

// sendCommand sends a CDP command to this page's session.
func (p *Page) sendCommand(method string, params map[string]interface{}) (map[string]interface{}, error) {
	return p.browser.client.SendToSession(p.ctx, p.sessionID, method, params)
}

// WaitForLoad waits for the page load event to fire.
func (p *Page) WaitForLoad(timeout time.Duration) error {
	p.log.Debug("waiting for page load", "timeout", timeout)

	var loadChan chan bool = make(chan bool, 1)

	// Register event handler for Page.loadEventFired
	p.browser.client.On("Page.loadEventFired", func(params map[string]interface{}) {
		select {
		case loadChan <- true:
		default:
		}
	})

	// Wait for load event or timeout
	var timeoutCtx context.Context
	var cancel context.CancelFunc
	timeoutCtx, cancel = context.WithTimeout(p.ctx, timeout)
	defer cancel()

	select {
	case <-loadChan:
		p.log.Debug("page load event fired")
		return nil
	case <-timeoutCtx.Done():
		p.log.Warn("page load timeout", "timeout", timeout)
		return fmt.Errorf("kexas: page load timeout after %v", timeout)
	}
}

// Close closes the page (tab).
func (p *Page) Close() error {
	p.log.Debug("closing page", "targetId", p.targetID)

	var params map[string]interface{} = map[string]interface{}{
		"targetId": p.targetID,
	}

	var err error
	_, err = p.browser.client.Send(p.ctx, "Target.closeTarget", params)
	if err != nil {
		p.log.Error("failed to close page", "err", err)
		return fmt.Errorf("kexas: failed to close page: %w", err)
	}

	p.log.Info("page closed")
	return nil
}

// URL returns the current page URL.
func (p *Page) URL() (string, error) {
	var params map[string]interface{} = map[string]interface{}{
		"targetId": p.targetID,
	}

	var result map[string]interface{}
	var err error
	result, err = p.browser.client.Send(p.ctx, "Target.getTargetInfo", params)
	if err != nil {
		return "", fmt.Errorf("kexas: failed to get target info: %w", err)
	}

	var targetInfo map[string]interface{}
	var ok bool
	targetInfo, ok = result["targetInfo"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("kexas: invalid targetInfo in response")
	}

	var url string
	url, ok = targetInfo["url"].(string)
	if !ok {
		return "", fmt.Errorf("kexas: invalid url in targetInfo")
	}

	return url, nil
}

// Screenshot captures a screenshot of the page and returns the image data as base64-encoded PNG.
func (p *Page) Screenshot() ([]byte, error) {
	p.log.Debug("capturing screenshot")

	var params map[string]interface{} = map[string]interface{}{
		"format":  "png",
		"quality": 100,
	}

	var result map[string]interface{}
	var err error
	result, err = p.sendCommand("Page.captureScreenshot", params)
	if err != nil {
		p.log.Error("screenshot failed", "err", err)
		return nil, fmt.Errorf("kexas: screenshot failed: %w", err)
	}

	var dataStr string
	var ok bool
	dataStr, ok = result["data"].(string)
	if !ok {
		return nil, fmt.Errorf("kexas: invalid data in screenshot response")
	}

	p.log.Info("screenshot captured", "size", len(dataStr))
	return []byte(dataStr), nil
}

// Title returns the page title.
func (p *Page) Title() (string, error) {
	p.log.Debug("getting page title")

	var params map[string]interface{} = map[string]interface{}{
		"targetId": p.targetID,
	}

	var result map[string]interface{}
	var err error
	result, err = p.browser.client.Send(p.ctx, "Target.getTargetInfo", params)
	if err != nil {
		return "", fmt.Errorf("kexas: failed to get target info: %w", err)
	}

	var targetInfo map[string]interface{}
	var ok bool
	targetInfo, ok = result["targetInfo"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("kexas: invalid targetInfo in response")
	}

	var title string
	title, ok = targetInfo["title"].(string)
	if !ok {
		return "", fmt.Errorf("kexas: invalid title in targetInfo")
	}

	return title, nil
}
