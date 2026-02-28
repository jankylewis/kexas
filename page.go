package kexas

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/kexas-project/kexas/internal/agent"
	"github.com/kexas-project/kexas/internal/cdp"
	"github.com/kexas-project/kexas/internal/logger"
	"github.com/kexas-project/kexas/kwait"
)

// Page represents a browser page (tab).
type Page struct {
	browser      *Browser
	targetID     string
	sessionID    string
	log          *logger.Logger
	ctx          context.Context
	agentManager *agent.AgentManager
	closed       bool
}

// Navigate navigates the page to the given URL.
// Optional waitUntil parameter controls when navigation is considered complete.
// If not specified, defaults to WaitUntilLoad.
//
// Like Playwright's goto(), this sends the navigate command and waits for the
// initial page response. Element availability is handled by Find()'s auto-retry.
func (p *Page) Navigate(url string, waitUntil ...kwait.WaitUntil) error {
	p.log.Debug("navigating", "url", url)

	// CDP Page.navigate only accepts url, referrer, transitionType, frameId
	var params map[string]interface{} = map[string]interface{}{
		"url": url,
	}

	var err error
	_, err = p.sendCommand(cdp.CmdPageNavigate, params)
	if err != nil {
		p.log.Error("navigation failed", "url", url, "error", err)
		return fmt.Errorf("navigation failed: %w", err)
	}

	p.log.Info("navigate command sent", "url", url)

	// Wait for readyState to transition through loading → complete
	// This ensures the server has responded and the initial HTML is parsed.
	var navTimeout time.Duration = 30 * time.Second
	var start time.Time = time.Now()
	var pollInterval time.Duration = 100 * time.Millisecond
	var sawLoading bool = false

	for time.Since(start) < navTimeout {
		var result map[string]interface{}
		result, err = p.browser.client.SendToSession(p.ctx, p.sessionID, cdp.CmdRuntimeEvaluate, map[string]interface{}{
			"expression":    "document.readyState",
			"returnByValue": true,
		})
		if err != nil {
			// During navigation, the execution context is destroyed; this means navigation started
			sawLoading = true
			time.Sleep(pollInterval)
			continue
		}

		var resultObj map[string]interface{}
		var ok bool
		resultObj, ok = result["result"].(map[string]interface{})
		if ok {
			var readyState string
			readyState, ok = resultObj["value"].(string)
			if ok {
				if readyState != "complete" {
					sawLoading = true
				}
				if sawLoading && readyState == "complete" {
					p.log.Info("page loaded", "url", url, "elapsed", time.Since(start))
					// Reset DOM agent for the new document
					p.agentManager.ResetAgent(agent.AgentDOM)
					return nil
				}
			}
		}

		time.Sleep(pollInterval)
	}

	return fmt.Errorf("navigation timeout after %v", navTimeout)
}

// WaitForLoadState waits for the page to reach a specific load state.
func (p *Page) WaitForLoadState(waitUntil kwait.WaitUntil, timeout time.Duration) error {
	p.log.Debug("waiting for load state", "waitUntil", waitUntil, "timeout", timeout)

	var start time.Time = time.Now()
	var pollInterval time.Duration = 100 * time.Millisecond

	// Determine the target readyState based on waitUntil
	var targetState string
	switch waitUntil {
	case kwait.WaitUntilDOMContentLoaded:
		targetState = "interactive"
	default:
		targetState = "complete"
	}

	for time.Since(start) < timeout {
		var result map[string]interface{}
		var err error
		result, err = p.sendCommand(cdp.CmdRuntimeEvaluate, map[string]interface{}{
			"expression":    "document.readyState",
			"returnByValue": true,
		})
		if err != nil {
			// During navigation, Runtime.evaluate may fail temporarily; retry
			time.Sleep(pollInterval)
			continue
		}

		var resultObj map[string]interface{}
		var ok bool
		resultObj, ok = result["result"].(map[string]interface{})
		if ok {
			var readyState string
			readyState, ok = resultObj["value"].(string)
			if ok && (readyState == targetState || readyState == "complete") {
				p.log.Debug("load state reached", "state", readyState, "elapsed", time.Since(start))
				return nil
			}
		}

		time.Sleep(pollInterval)
	}

	return fmt.Errorf("wait for load state timeout after %v", timeout)
}

// WaitForNavigationCompleted waits for navigation to complete within the timeout.
func (p *Page) WaitForNavigationCompleted(timeout time.Duration) error {
	return p.WaitForLoadState(kwait.WaitUntilLoad, timeout)
}

// sendCommand sends a command to the page's CDP session.
// Automatically ensures required agents are enabled before sending.
func (p *Page) sendCommand(method string, params map[string]interface{}) (map[string]interface{}, error) {
	// Ensure required agents are enabled before sending the command
	var err error = p.ensureAgentsForCommand(method)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure agents for command %s: %w", method, err)
	}

	return p.browser.client.SendToSession(p.ctx, p.sessionID, method, params)
}

// ensureAgentsForCommand ensures the required agents are enabled for a given CDP command
func (p *Page) ensureAgentsForCommand(method string) error {
	switch method {
	case cdp.CmdDOMPerformSearch, cdp.CmdDOMGetSearchResults, cdp.CmdDOMQuerySelector,
		cdp.CmdDOMGetComputedStyle, cdp.CmdDOMGetBoxModel, cdp.CmdDOMGetAttributes,
		cdp.CmdDOMGetOuterHTML, cdp.CmdDOMDescribeNode, "DOM.requestNode", "DOM.getDocument", cdp.CmdDOMResolveNode:
		return p.agentManager.EnsureAgent(agent.AgentDOM)
	case cdp.CmdRuntimeEvaluate, cdp.CmdRuntimeCallFunctionOn:
		return p.agentManager.EnsureAgent(agent.AgentRuntime)
	case cdp.CmdPageNavigate, cdp.CmdPageCaptureScreenshot,
		cdp.CmdPageStartScreencast, cdp.CmdPageStopScreencast,
		cdp.CmdPageScreencastFrameAck, cdp.CmdPageBringToFront:
		return p.agentManager.EnsureAgent(agent.AgentPage)
	case cdp.CmdNetworkGetCookies, cdp.CmdNetworkSetCookie,
		cdp.CmdNetworkDeleteCookies, cdp.CmdNetworkClearBrowserCookies:
		return p.agentManager.EnsureAgent(agent.AgentNetwork)
	case cdp.CmdInputDispatchKeyEvent, "Input.dispatchMouseEvent", "Input.dispatchTouchEvent":
		// Input domain is auto-enabled, no agent needed
		return nil
	default:
		// For unknown commands, don't require any specific agent
		return nil
	}
}

// WaitForLoad waits for the page to load within the timeout.
func (p *Page) WaitForLoad(timeout time.Duration) error {
	p.log.Debug("waiting for page load", "timeout", timeout)

	var start time.Time = time.Now()
	var pollInterval time.Duration = 100 * time.Millisecond

	for time.Since(start) < timeout {
		var result map[string]interface{}
		var err error
		result, err = p.sendCommand(cdp.CmdRuntimeEvaluate, map[string]interface{}{
			"expression": "document.readyState",
		})
		if err != nil {
			return fmt.Errorf("failed to check ready state: %w", err)
		}

		var readyState string
		var ok bool
		readyState, ok = result["result"].(map[string]interface{})["value"].(string)
		if ok && readyState == "complete" {
			p.log.Debug("page loaded", "elapsed", time.Since(start))
			return nil
		}

		time.Sleep(pollInterval)
	}

	return fmt.Errorf("page load timeout after %v", timeout)
}

// Close closes the page and cleans up resources.
func (p *Page) Close() error {
	if p.closed {
		return nil
	}

	p.log.Debug("closing page")

	var err error
	_, err = p.sendCommand(cdp.CmdPageClose, nil)
	if err != nil {
		p.log.Error("failed to close page", "error", err)
		return fmt.Errorf("failed to close page: %w", err)
	}

	p.closed = true

	// Remove from browser's tracked pages
	if p.browser != nil {
		p.browser.removePage(p)
	}

	p.log.Info("page closed")
	return nil
}

// IsClosed returns whether this page has been closed.
func (p *Page) IsClosed() bool {
	return p.closed
}

// BringToFront activates this tab (brings it to the foreground).
func (p *Page) BringToFront() error {
	p.log.Debug("bringing page to front")

	var err error
	_, err = p.sendCommand(cdp.CmdPageBringToFront, nil)
	if err != nil {
		return fmt.Errorf("bring to front failed: %w", err)
	}

	return nil
}

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
