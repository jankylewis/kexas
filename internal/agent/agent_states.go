package agent

import (
	"time"
)

// AgentStates stores detailed state information for all agents
type AgentStates struct {
	DOM      *AgentState
	Input    *AgentState
	Runtime  *AgentState
	Network  *AgentState
	Page     *AgentState
	Security *AgentState
	Debugger *AgentState
	Profiler *AgentState
}

// GetState returns the state for a specific agent
func (as AgentStates) GetState(agentName string) *AgentState {
	switch agentName {
	case AgentDOM:
		return as.DOM
	case AgentInput:
		return as.Input
	case AgentRuntime:
		return as.Runtime
	case AgentNetwork:
		return as.Network
	case AgentPage:
		return as.Page
	case AgentSecurity:
		return as.Security
	case AgentDebugger:
		return as.Debugger
	case AgentProfiler:
		return as.Profiler
	default:
		return nil
	}
}

// SetState sets the state for a specific agent
func (as *AgentStates) SetState(agentName string, state *AgentState) {
	switch agentName {
	case AgentDOM:
		as.DOM = state
	case AgentInput:
		as.Input = state
	case AgentRuntime:
		as.Runtime = state
	case AgentNetwork:
		as.Network = state
	case AgentPage:
		as.Page = state
	case AgentSecurity:
		as.Security = state
	case AgentDebugger:
		as.Debugger = state
	case AgentProfiler:
		as.Profiler = state
	}
}

// List returns all agent states (including nil ones)
func (as AgentStates) List() []*AgentState {
	return []*AgentState{
		as.DOM, as.Input, as.Runtime, as.Network,
		as.Page, as.Security, as.Debugger, as.Profiler,
	}
}

// CountEnabled returns the number of enabled agents
func (as AgentStates) CountEnabled() int {
	var count int
	if as.DOM != nil && as.DOM.Enabled {
		count++
	}
	if as.Input != nil && as.Input.Enabled {
		count++
	}
	if as.Runtime != nil && as.Runtime.Enabled {
		count++
	}
	if as.Network != nil && as.Network.Enabled {
		count++
	}
	if as.Page != nil && as.Page.Enabled {
		count++
	}
	if as.Security != nil && as.Security.Enabled {
		count++
	}
	if as.Debugger != nil && as.Debugger.Enabled {
		count++
	}
	if as.Profiler != nil && as.Profiler.Enabled {
		count++
	}
	return count
}

// TotalMemoryUsage returns the total memory usage across all agents
func (as AgentStates) TotalMemoryUsage() int {
	var total int
	if as.DOM != nil {
		total += as.DOM.MemoryUsage
	}
	if as.Input != nil {
		total += as.Input.MemoryUsage
	}
	if as.Runtime != nil {
		total += as.Runtime.MemoryUsage
	}
	if as.Network != nil {
		total += as.Network.MemoryUsage
	}
	if as.Page != nil {
		total += as.Page.MemoryUsage
	}
	if as.Security != nil {
		total += as.Security.MemoryUsage
	}
	if as.Debugger != nil {
		total += as.Debugger.MemoryUsage
	}
	if as.Profiler != nil {
		total += as.Profiler.MemoryUsage
	}
	return total
}

// ResetAll resets all agent states to disabled.
func (as *AgentStates) ResetAll() {
	resetState(as.DOM)
	resetState(as.Input)
	resetState(as.Runtime)
	resetState(as.Network)
	resetState(as.Page)
	resetState(as.Security)
	resetState(as.Debugger)
	resetState(as.Profiler)
}

// resetState clears one AgentState back to its zero/disabled form. Nil-safe so
// callers don't have to nil-check each field.
func resetState(s *AgentState) {
	if s == nil {
		return
	}
	s.Enabled = false
	s.Context = nil
	s.LastUsed = time.Time{}
	s.Error = nil
}
