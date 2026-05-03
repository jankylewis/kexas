package agent

import (
	"time"
)

// AgentContext stores agent-specific runtime context data
type AgentContext struct {
	DOM      *DOMContext
	Input    *InputContext
	Runtime  *RuntimeContext
	Network  *NetworkContext
	Page     *PageContext
	Security *SecurityContext
	Debugger *DebuggerContext
	Profiler *ProfilerContext
}

// DOMContext stores DOM agent runtime data
type DOMContext struct {
	Root *DOMRoot `json:"root"`
}

type DOMRoot struct {
	NodeID   int    `json:"nodeId"`
	NodeType int    `json:"nodeType"`
	NodeName string `json:"nodeName"`
}

// InputContext stores Input agent runtime data
type InputContext struct {
	Ready     bool      `json:"ready"`
	EnabledAt time.Time `json:"enabled_at"`
}

// RuntimeContext stores Runtime agent runtime data
type RuntimeContext struct {
	ExecutionContextID int `json:"executionContextId"`
}

// NetworkContext stores Network agent runtime data
type NetworkContext struct {
	RequestID string `json:"requestId"`
}

// PageContext stores Page agent runtime data
type PageContext struct {
	FrameID string `json:"frameId"`
}

// SecurityContext stores Security agent runtime data
type SecurityContext struct {
	CertificateID string `json:"certificateId"`
}

// DebuggerContext stores Debugger agent runtime data
type DebuggerContext struct {
	BreakpointID string `json:"breakpointId"`
}

// ProfilerContext stores Profiler agent runtime data
type ProfilerContext struct {
	ProfileID string `json:"profileId"`
}

// GetContext returns the context for a specific agent
func (ac AgentContext) GetContext(agentName string) interface{} {
	switch agentName {
	case AgentDOM:
		if ac.DOM == nil {
			return nil
		}
		return ac.DOM
	case AgentInput:
		if ac.Input == nil {
			return nil
		}
		return ac.Input
	case AgentRuntime:
		if ac.Runtime == nil {
			return nil
		}
		return ac.Runtime
	case AgentNetwork:
		if ac.Network == nil {
			return nil
		}
		return ac.Network
	case AgentPage:
		if ac.Page == nil {
			return nil
		}
		return ac.Page
	case AgentSecurity:
		if ac.Security == nil {
			return nil
		}
		return ac.Security
	case AgentDebugger:
		if ac.Debugger == nil {
			return nil
		}
		return ac.Debugger
	case AgentProfiler:
		if ac.Profiler == nil {
			return nil
		}
		return ac.Profiler
	default:
		return nil
	}
}

// SetContext sets the context for a specific agent. Silently drops a context whose
// dynamic type does not match the agent's expected context struct.
func (ac *AgentContext) SetContext(agentName string, context interface{}) {
	switch agentName {
	case AgentDOM:
		assignIfTypeMatches(&ac.DOM, context)
	case AgentInput:
		assignIfTypeMatches(&ac.Input, context)
	case AgentRuntime:
		assignIfTypeMatches(&ac.Runtime, context)
	case AgentNetwork:
		assignIfTypeMatches(&ac.Network, context)
	case AgentPage:
		assignIfTypeMatches(&ac.Page, context)
	case AgentSecurity:
		assignIfTypeMatches(&ac.Security, context)
	case AgentDebugger:
		assignIfTypeMatches(&ac.Debugger, context)
	case AgentProfiler:
		assignIfTypeMatches(&ac.Profiler, context)
	}
}

// assignIfTypeMatches sets *dst = src if src's dynamic type is T. No-op on mismatch.
// Used by SetContext to collapse 8 nearly-identical type-assert-and-assign blocks.
func assignIfTypeMatches[T any](dst *T, src interface{}) {
	if v, ok := src.(T); ok {
		*dst = v
	}
}

// ClearContext removes the context for a specific agent
func (ac *AgentContext) ClearContext(agentName string) {
	switch agentName {
	case AgentDOM:
		ac.DOM = nil
	case AgentInput:
		ac.Input = nil
	case AgentRuntime:
		ac.Runtime = nil
	case AgentNetwork:
		ac.Network = nil
	case AgentPage:
		ac.Page = nil
	case AgentSecurity:
		ac.Security = nil
	case AgentDebugger:
		ac.Debugger = nil
	case AgentProfiler:
		ac.Profiler = nil
	}
}
