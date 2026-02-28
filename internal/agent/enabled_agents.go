package agent

// Chrome DevTools Protocol agent names
const (
	AgentDOM      = "DOM"
	AgentInput    = "Input"
	AgentRuntime  = "Runtime"
	AgentNetwork  = "Network"
	AgentPage     = "Page"
	AgentSecurity = "Security"
	AgentDebugger = "Debugger"
	AgentProfiler = "Profiler"
)

// EnabledAgents tracks which Chrome DevTools Protocol agents are enabled
type EnabledAgents struct {
	DOM      bool
	Input    bool
	Runtime  bool
	Network  bool
	Page     bool
	Security bool
	Debugger bool
	Profiler bool
}

// IsEnabled checks if a specific agent is enabled
func (ea EnabledAgents) IsEnabled(agentName string) bool {
	switch agentName {
	case AgentDOM:
		return ea.DOM
	case AgentInput:
		return ea.Input
	case AgentRuntime:
		return ea.Runtime
	case AgentNetwork:
		return ea.Network
	case AgentPage:
		return ea.Page
	case AgentSecurity:
		return ea.Security
	case AgentDebugger:
		return ea.Debugger
	case AgentProfiler:
		return ea.Profiler
	default:
		return false
	}
}

// Enable sets an agent to enabled state
func (ea *EnabledAgents) Enable(agentName string) {
	switch agentName {
	case AgentDOM:
		ea.DOM = true
	case AgentInput:
		ea.Input = true
	case AgentRuntime:
		ea.Runtime = true
	case AgentNetwork:
		ea.Network = true
	case AgentPage:
		ea.Page = true
	case AgentSecurity:
		ea.Security = true
	case AgentDebugger:
		ea.Debugger = true
	case AgentProfiler:
		ea.Profiler = true
	}
}

// Disable sets an agent to disabled state
func (ea *EnabledAgents) Disable(agentName string) {
	switch agentName {
	case AgentDOM:
		ea.DOM = false
	case AgentInput:
		ea.Input = false
	case AgentRuntime:
		ea.Runtime = false
	case AgentNetwork:
		ea.Network = false
	case AgentPage:
		ea.Page = false
	case AgentSecurity:
		ea.Security = false
	case AgentDebugger:
		ea.Debugger = false
	case AgentProfiler:
		ea.Profiler = false
	}
}

// List returns all enabled agent names
func (ea EnabledAgents) List() []string {
	var enabled []string
	if ea.DOM {
		enabled = append(enabled, AgentDOM)
	}
	if ea.Input {
		enabled = append(enabled, AgentInput)
	}
	if ea.Runtime {
		enabled = append(enabled, AgentRuntime)
	}
	if ea.Network {
		enabled = append(enabled, AgentNetwork)
	}
	if ea.Page {
		enabled = append(enabled, AgentPage)
	}
	if ea.Security {
		enabled = append(enabled, AgentSecurity)
	}
	if ea.Debugger {
		enabled = append(enabled, AgentDebugger)
	}
	if ea.Profiler {
		enabled = append(enabled, AgentProfiler)
	}
	return enabled
}
