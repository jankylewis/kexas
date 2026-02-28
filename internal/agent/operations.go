package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/kexas-project/kexas/errors"
)

// IsEnabled checks if an agent is currently enabled
func (am *AgentManager) IsEnabled(agentName string) bool {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	return am.enabledAgents.IsEnabled(agentName)
}

// GetContext returns the cached context for an agent
func (am *AgentManager) GetContext(agentName string) (interface{}, error) {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	var context interface{} = am.agentContext.GetContext(agentName)
	if context == nil {
		return nil, errors.AgentContextNotFound(agentName)
	}

	return context, nil
}

// GetState returns the detailed state for an agent
func (am *AgentManager) GetState(agentName string) (*AgentState, error) {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	var state *AgentState = am.agentState.GetState(agentName)
	if state == nil {
		return nil, errors.AgentNotFound(agentName)
	}

	return state, nil
}

// GetEnabledAgents returns a list of all currently enabled agents
func (am *AgentManager) GetEnabledAgents() []string {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	return am.enabledAgents.List()
}

// ResetAgent marks an agent as disabled and clears its context without sending
// a CDP disable command. This is used after navigation to force re-enablement
// with fresh document state on the next EnsureAgent call.
func (am *AgentManager) ResetAgent(agentName string) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	am.enabledAgents.Disable(agentName)
	am.agentContext.ClearContext(agentName)

	var state *AgentState = am.agentState.GetState(agentName)
	if state != nil {
		state.Enabled = false
		state.Context = nil
	}
}

// Disable disables an agent and clears its context
func (am *AgentManager) Disable(agentName string) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	if !am.enabledAgents.IsEnabled(agentName) {
		return nil // Already disabled
	}

	// Call CDP to disable agent
	var _, err = am.cdp.SendToSession(context.Background(), am.sessionID, agentName+".disable", nil)
	if err != nil {
		return errors.AgentDisableFailed(agentName, err)
	}

	// Update enabled status
	am.enabledAgents.Disable(agentName)

	// Clear cached context
	am.agentContext.ClearContext(agentName)

	// Update agent state
	var state *AgentState = am.agentState.GetState(agentName)
	if state != nil {
		state.Enabled = false
		state.Context = nil
		state.LastUsed = time.Time{}
	}

	return nil
}

// DisableAll disables all enabled agents
func (am *AgentManager) DisableAll() error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	var errs []error
	var enabledList []string = am.enabledAgents.List()

	for i := 0; i < len(enabledList); i++ {
		var agent string = enabledList[i]
		var _, err = am.cdp.SendToSession(context.Background(), am.sessionID, agent+".disable", nil)
		if err != nil {
			errs = append(errs, errors.AgentDisableFailed(agent, err))
		}
	}

	// Clear all cached state
	am.enabledAgents = EnabledAgents{}
	am.agentContext = AgentContext{}

	// Reset agent states
	am.agentState.ResetAll()

	if len(errs) > 0 {
		return fmt.Errorf("multiple agent disable errors: %v", errs)
	}

	return nil
}

// Cleanup removes stale agent contexts
func (am *AgentManager) Cleanup() {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	var now time.Time = time.Now()
	var allStates []*AgentState = am.agentState.List()

	for i := 0; i < len(allStates); i++ {
		var state *AgentState = allStates[i]
		// Clean up contexts that haven't been used recently
		if state != nil && state.Context != nil && am.config.ContextTimeout > 0 {
			if now.Sub(state.LastUsed) > am.config.ContextTimeout {
				if am.config.Debug {
					am.log.Debug("Cleaning up stale context for agent", "agent", "unknown")
				}
				state.Context = nil
			}
		}
	}
}

// Stats returns statistics about the agent manager
func (am *AgentManager) Stats() map[string]interface{} {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	var enabledCount int = len(am.enabledAgents.List())
	var contextCount int = 0

	// Count non-nil contexts
	if am.agentContext.DOM != nil {
		contextCount++
	}
	if am.agentContext.Input != nil {
		contextCount++
	}
	if am.agentContext.Runtime != nil {
		contextCount++
	}
	if am.agentContext.Network != nil {
		contextCount++
	}
	if am.agentContext.Page != nil {
		contextCount++
	}
	if am.agentContext.Security != nil {
		contextCount++
	}
	if am.agentContext.Debugger != nil {
		contextCount++
	}
	if am.agentContext.Profiler != nil {
		contextCount++
	}

	var totalAgents int = len(am.agentState.List())
	var totalMemory int = am.agentState.TotalMemoryUsage()

	var stats map[string]interface{} = map[string]interface{}{
		"total_agents":       totalAgents,
		"enabled_agents":     enabledCount,
		"cached_contexts":    contextCount,
		"enabled_list":       am.GetEnabledAgents(),
		"total_memory_usage": totalMemory,
	}

	return stats
}
