package agent

import (
	"context"
	"time"
)

// initDOMContext initializes DOM agent context
func (am *AgentManager) initDOMContext(ctx context.Context, agentName string) error {
	// Create DOM context structure with defaults
	var domContext *DOMContext = &DOMContext{
		Root: &DOMRoot{
			NodeID:   1,
			NodeType: 9,
			NodeName: "#document",
		},
	}

	// Try to get document root node, but don't fail if page hasn't loaded yet
	var result map[string]interface{}
	var err error
	result, err = am.cdp.SendToSession(ctx, am.sessionID, "DOM.getDocument", nil)
	if err == nil {
		// Extract actual node ID from result if available
		var rootData map[string]interface{}
		var ok bool
		rootData, ok = result["root"].(map[string]interface{})
		if ok {
			var nodeID float64
			nodeID, ok = rootData["nodeId"].(float64)
			if ok {
				domContext.Root.NodeID = int(nodeID)
			}
			var nodeType float64
			nodeType, ok = rootData["nodeType"].(float64)
			if ok {
				domContext.Root.NodeType = int(nodeType)
			}
			var nodeName string
			nodeName, ok = rootData["nodeName"].(string)
			if ok {
				domContext.Root.NodeName = nodeName
			}
		}
	}

	// Store the context using the new struct method
	am.agentContext.SetContext(agentName, domContext)

	if am.config.Debug {
		am.log.Debug("DOM agent context initialized", "nodeId", domContext.Root.NodeID)
	}

	return nil
}

// initInputContext initializes Input agent context
func (am *AgentManager) initInputContext(agentName string) error {
	// Input agent doesn't need special context initialization
	// Just mark that it's ready for input operations
	var inputContext *InputContext = &InputContext{
		Ready:     true,
		EnabledAt: time.Now(),
	}
	am.agentContext.SetContext(agentName, inputContext)

	if am.config.Debug {
		am.log.Debug("Input agent context initialized")
	}

	return nil
}

// initRuntimeContext initializes Runtime agent context
func (am *AgentManager) initRuntimeContext(agentName string) error {
	// Runtime agent might need execution context
	// For now, just mark it as ready
	var runtimeContext *RuntimeContext = &RuntimeContext{
		ExecutionContextID: 1, // Default execution context
	}
	am.agentContext.SetContext(agentName, runtimeContext)

	if am.config.Debug {
		am.log.Debug("Runtime agent context initialized")
	}

	return nil
}

// initNetworkContext initializes Network agent context
func (am *AgentManager) initNetworkContext(agentName string) error {
	// Network agent context
	var networkContext *NetworkContext = &NetworkContext{
		RequestID: "", // Will be set when actual network requests occur
	}
	am.agentContext.SetContext(agentName, networkContext)

	if am.config.Debug {
		am.log.Debug("Network agent context initialized")
	}

	return nil
}

// initPageContext initializes Page agent context
func (am *AgentManager) initPageContext(agentName string) error {
	// Page agent context
	var pageContext *PageContext = &PageContext{
		FrameID: "", // Will be set when page loads
	}
	am.agentContext.SetContext(agentName, pageContext)

	if am.config.Debug {
		am.log.Debug("Page agent context initialized")
	}

	return nil
}

// initUnknownContext initializes unknown agent context
func (am *AgentManager) initUnknownContext(agentName string) error {
	// Unknown agent, just mark as enabled without specific context
	// No specific context structure for unknown agents
	if am.config.Debug {
		am.log.Debug("Unknown agent context initialized", "agent", agentName)
	}

	return nil
}

// initAgentContext sets up agent-specific context after enablement
func (am *AgentManager) initAgentContext(agentName string) error {
	var ctx context.Context
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(context.Background(), am.config.EnableTimeout)
	defer cancel()

	switch agentName {
	case "DOM":
		return am.initDOMContext(ctx, agentName)
	case "Input":
		return am.initInputContext(agentName)
	case "Runtime":
		return am.initRuntimeContext(agentName)
	case "Network":
		return am.initNetworkContext(agentName)
	case "Page":
		return am.initPageContext(agentName)
	default:
		return am.initUnknownContext(agentName)
	}
}
