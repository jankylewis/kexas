package kcore

import (
	"context"
	"fmt"

	"github.com/jankylewis/kexas/internal/agent"
	"github.com/jankylewis/kexas/internal/cdp"
	"github.com/jankylewis/kexas/internal/logger"
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
