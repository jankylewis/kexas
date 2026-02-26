package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/kexas-project/kexas/errors"
)

// EnsureAgent ensures an agent is enabled and returns its context
// This is the core smart enablement logic like Playwright
func (am *AgentManager) EnsureAgent(agentName string) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	// Check if agent is already enabled
	if am.enabledAgents.IsEnabled(agentName) {
		// Update last used time
		var state *AgentState = am.agentState.GetState(agentName)
		if state != nil {
			state.LastUsed = time.Now()
		}
		return nil
	}

	// Enable the agent
	if am.config.Debug {
		am.log.Debug("enabling agent", "name", agentName)
	}

	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), am.config.EnableTimeout)
	defer cancel()

	// Send enable command to the page session (not browser target)
	var _, err = am.cdp.SendToSession(ctx, am.sessionID, agentName+".enable", nil)
	if err != nil {
		var state *AgentState = am.agentState.GetState(agentName)
		if state != nil {
			state.Error = errors.AgentEnableFailed(agentName, err)
		}
		return errors.AgentEnableFailed(agentName, err)
	}

	// Mark agent as enabled
	am.enabledAgents.Enable(agentName)

	// Update agent state
	var state *AgentState = am.agentState.GetState(agentName)
	if state != nil {
		state.Enabled = true
		state.LastUsed = time.Now()
		state.Error = nil
	}

	// Initialize agent-specific context
	var initErr error = am.initAgentContext(agentName)
	if initErr != nil {
		return errors.AgentEnableFailed(agentName, initErr)
	}

	if am.config.Debug {
		am.log.Info("agent enabled", "name", agentName)
	}

	return nil
}

// EnsureAgents enables multiple agents efficiently
// This is useful for operations that require multiple agents
func (am *AgentManager) EnsureAgents(agentNames ...string) error {
	var errs []error

	// Enable agents in parallel for better performance
	type result struct {
		agent string
		err   error
	}

	var results chan result = make(chan result, len(agentNames))

	for i := 0; i < len(agentNames); i++ {
		var agentName string = agentNames[i]
		go func(agentName string) {
			var err error = am.EnsureAgent(agentName)
			var res result = result{agent: agentName, err: err}
			results <- res
		}(agentName)
	}

	// Collect results
	for i := 0; i < len(agentNames); i++ {
		var res result = <-results
		if res.err != nil {
			errs = append(errs, errors.AgentEnableFailed(res.agent, res.err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("multiple agent enablement errors: %v", errs)
	}

	return nil
}

// GetDocumentNodeId returns the cached document node ID from DOM agent
// This is commonly needed for DOM operations
func (am *AgentManager) GetDocumentNodeId() (interface{}, error) {
	// Ensure DOM agent is enabled
	var err error = am.EnsureAgent("DOM")
	if err != nil {
		return nil, err
	}

	am.mutex.RLock()
	defer am.mutex.RUnlock()

	var domContext interface{} = am.agentContext.GetContext("DOM")
	if domContext == nil {
		return nil, errors.AgentContextNotFound("DOM")
	}

	// Extract document node ID from the cached DOM context
	var domCtx *DOMContext
	var ok bool
	domCtx, ok = domContext.(*DOMContext)
	if !ok || domCtx == nil {
		return nil, errors.AgentContextNotFound("DOM")
	}

	return domCtx.Root.NodeID, nil
}

// IsAgentReady checks if an agent is enabled and ready for use
func (am *AgentManager) IsAgentReady(agentName string) bool {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	// Check if agent is enabled
	if !am.enabledAgents.IsEnabled(agentName) {
		return false
	}

	// Check if agent has context
	var context interface{} = am.agentContext.GetContext(agentName)
	if context == nil {
		return false
	}

	// For DOM agent, any context means it's ready
	if agentName == "DOM" {
		return true
	}

	// For other agents, check if context indicates readiness
	switch agentName {
	case "Input":
		var inputCtx *InputContext
		var ok bool
		inputCtx, ok = context.(*InputContext)
		if ok {
			return inputCtx.Ready
		}
	case "Runtime":
		var runtimeCtx *RuntimeContext
		var ok bool
		runtimeCtx, ok = context.(*RuntimeContext)
		if ok {
			return runtimeCtx.ExecutionContextID > 0
		}
	}

	return true
}

// WaitForAgentReady waits for an agent to become ready
// Useful for scenarios where agent enablement might be delayed
func (am *AgentManager) WaitForAgentReady(agentName string, timeout time.Duration) error {
	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var ticker *time.Ticker = time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return errors.AgentNotReady(agentName, ctx.Err())
		case <-ticker.C:
			if am.IsAgentReady(agentName) {
				return nil
			}
		}
	}
}

// RefreshAgentContext refreshes the context for an agent
// Useful when the context might have become stale
func (am *AgentManager) RefreshAgentContext(agentName string) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	if !am.enabledAgents.IsEnabled(agentName) {
		return errors.ErrAgentNotEnabled
	}

	// Re-initialize the context
	return am.initAgentContext(agentName)
}
