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

// SetContext sets the context for a specific agent
func (ac *AgentContext) SetContext(agentName string, context interface{}) {
	switch agentName {
	case AgentDOM:
		var domCtx *DOMContext
		var ok bool
		domCtx, ok = context.(*DOMContext)
		if ok {
			ac.DOM = domCtx
		}
	case AgentInput:
		var inputCtx *InputContext
		var ok bool
		inputCtx, ok = context.(*InputContext)
		if ok {
			ac.Input = inputCtx
		}
	case AgentRuntime:
		var runtimeCtx *RuntimeContext
		var ok bool
		runtimeCtx, ok = context.(*RuntimeContext)
		if ok {
			ac.Runtime = runtimeCtx
		}
	case AgentNetwork:
		var networkCtx *NetworkContext
		var ok bool
		networkCtx, ok = context.(*NetworkContext)
		if ok {
			ac.Network = networkCtx
		}
	case AgentPage:
		var pageCtx *PageContext
		var ok bool
		pageCtx, ok = context.(*PageContext)
		if ok {
			ac.Page = pageCtx
		}
	case AgentSecurity:
		var securityCtx *SecurityContext
		var ok bool
		securityCtx, ok = context.(*SecurityContext)
		if ok {
			ac.Security = securityCtx
		}
	case AgentDebugger:
		var debuggerCtx *DebuggerContext
		var ok bool
		debuggerCtx, ok = context.(*DebuggerContext)
		if ok {
			ac.Debugger = debuggerCtx
		}
	case AgentProfiler:
		var profilerCtx *ProfilerContext
		var ok bool
		profilerCtx, ok = context.(*ProfilerContext)
		if ok {
			ac.Profiler = profilerCtx
		}
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
