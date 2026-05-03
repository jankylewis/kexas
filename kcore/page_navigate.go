package kcore

import (
	"fmt"
	"time"

	"github.com/jankylewis/kexas/internal/agent"
	"github.com/jankylewis/kexas/internal/cdp"
	"github.com/jankylewis/kexas/kwait"
)

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
