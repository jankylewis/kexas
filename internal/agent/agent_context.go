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

// GetContext returns the context for a specific agent. nilOrTyped folds the
// nil-check + typed-return into a single helper so the dispatch is tiny.
func (ac AgentContext) GetContext(agentName string) interface{} {
	switch agentName {
	case AgentDOM:
		return nilOrTyped(ac.DOM)
	case AgentInput:
		return nilOrTyped(ac.Input)
	case AgentRuntime:
		return nilOrTyped(ac.Runtime)
	case AgentNetwork:
		return nilOrTyped(ac.Network)
	case AgentPage:
		return nilOrTyped(ac.Page)
	case AgentSecurity:
		return nilOrTyped(ac.Security)
	case AgentDebugger:
		return nilOrTyped(ac.Debugger)
	case AgentProfiler:
		return nilOrTyped(ac.Profiler)
	default:
		return nil
	}
}

// nilOrTyped returns nil when v's underlying pointer is nil; otherwise the
// typed pointer wrapped in interface{}. Avoids the typed-nil-as-non-nil
// interface gotcha that returning v directly would cause.
func nilOrTyped[T any](v *T) interface{} {
	if v == nil {
		return nil
	}
	return v
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
