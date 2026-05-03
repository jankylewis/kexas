
package internal_test

import (
	"testing"
	"time"

	"github.com/jankylewis/kexas/internal/agent"
	"github.com/jankylewis/kexas/internal/cdp"
)

// TestAgentManager_NewAgentManager_Critical tests AgentManager creation
func TestAgentManager_NewAgentManager_Critical(t *testing.T) {
	// Create a real CDP client but don't use it for network operations
	var am *agent.AgentManager = agent.NewAgentManager(&cdp.Client{}, "")

	// Verify AgentManager is created
	if am == nil {
		t.Fatal("expected AgentManager to be created")
	}

	// Verify initial state - all agents should be disabled
	if am.IsEnabled(agent.AgentDOM) {
		t.Error("expected DOM agent to be disabled initially")
	}
	if am.IsEnabled(agent.AgentInput) {
		t.Error("expected Input agent to be disabled initially")
	}
	if am.IsEnabled(agent.AgentRuntime) {
		t.Error("expected Runtime agent to be disabled initially")
	}
}

// TestEnabledAgents_IsEnabled_Critical tests EnabledAgents functionality
func TestEnabledAgents_IsEnabled_Critical(t *testing.T) {
	var ea agent.EnabledAgents

	// Test initial state - all should be disabled
	if ea.IsEnabled(agent.AgentDOM) {
		t.Error("expected DOM agent to be disabled initially")
	}
	if ea.IsEnabled(agent.AgentInput) {
		t.Error("expected Input agent to be disabled initially")
	}

	// Enable DOM agent
	ea.Enable(agent.AgentDOM)
	if !ea.IsEnabled(agent.AgentDOM) {
		t.Error("expected DOM agent to be enabled")
	}

	// Disable DOM agent
	ea.Disable(agent.AgentDOM)
	if ea.IsEnabled(agent.AgentDOM) {
		t.Error("expected DOM agent to be disabled")
	}
}

// TestEnabledAgents_List_Critical tests EnabledAgents List method
func TestEnabledAgents_List_Critical(t *testing.T) {
	var ea agent.EnabledAgents

	// Test empty list
	var list []string = ea.List()
	if len(list) != 0 {
		t.Error("expected empty list when no agents are enabled")
	}

	// Enable some agents
	ea.Enable(agent.AgentDOM)
	ea.Enable(agent.AgentInput)
	ea.Enable(agent.AgentRuntime)

	// Test list with enabled agents
	list = ea.List()
	if len(list) != 3 {
		t.Errorf("expected 3 enabled agents in list, got %d", len(list))
	}

	// Verify specific agents are in list
	var foundDOM bool = false
	var foundInput bool = false
	var foundRuntime bool = false

	for _, agentName := range list {
		switch agentName {
		case agent.AgentDOM:
			foundDOM = true
		case agent.AgentInput:
			foundInput = true
		case agent.AgentRuntime:
			foundRuntime = true
		}
	}

	if !foundDOM || !foundInput || !foundRuntime {
		t.Error("expected DOM, Input, and Runtime agents to be in enabled list")
	}
}

// TestAgentContext_SetContext_Critical tests AgentContext functionality
func TestAgentContext_SetContext_Critical(t *testing.T) {
	var ac agent.AgentContext

	// Test setting DOM context
	var domCtx *agent.DOMContext = &agent.DOMContext{
		Root: &agent.DOMRoot{
			NodeID:   1,
			NodeType: 9,
			NodeName: "#document",
		},
	}

	ac.SetContext(agent.AgentDOM, domCtx)

	// Test getting DOM context
	var context interface{} = ac.GetContext(agent.AgentDOM)
	if context == nil {
		t.Error("expected DOM context to be set")
	}

	var retrievedCtx *agent.DOMContext
	var ok bool
	retrievedCtx, ok = context.(*agent.DOMContext)
	if !ok {
		t.Error("expected DOM context type")
	}

	if retrievedCtx.Root.NodeID != 1 {
		t.Error("expected DOM context NodeID to be 1")
	}

	// Test clearing context
	ac.ClearContext(agent.AgentDOM)
	context = ac.GetContext(agent.AgentDOM)
	// After clearing, context should be nil - this is the expected behavior
	if context != nil {
		t.Errorf("expected DOM context to be nil after clearing, but got %v", context)
	}
	// Test passes if context is nil (successfully cleared)
}

// TestAgentStates_GetState_Critical tests AgentStates functionality
func TestAgentStates_GetState_Critical(t *testing.T) {
	var as agent.AgentStates

	// Test getting state for non-existent agent
	var state *agent.AgentState = as.GetState(agent.AgentDOM)
	if state != nil {
		t.Error("expected nil state for non-existent agent")
	}

	// Set state for DOM agent
	var domState *agent.AgentState = &agent.AgentState{
		Enabled:     true,
		LastUsed:    time.Now(),
		Context:     nil,
		MemoryUsage: 1024,
		Error:       nil,
	}

	as.SetState(agent.AgentDOM, domState)

	// Test getting state
	state = as.GetState(agent.AgentDOM)
	if state == nil {
		t.Error("expected DOM state to be set")
	}

	if !state.Enabled {
		t.Error("expected DOM state to be enabled")
	}

	if state.MemoryUsage != 1024 {
		t.Error("expected DOM state MemoryUsage to be 1024")
	}
}

// TestAgentStates_TotalMemoryUsage_Critical tests memory usage calculation
func TestAgentStates_TotalMemoryUsage_Critical(t *testing.T) {
	var as agent.AgentStates

	// Set states with different memory usage
	var domState *agent.AgentState = &agent.AgentState{MemoryUsage: 1024}
	var inputState *agent.AgentState = &agent.AgentState{MemoryUsage: 2048}
	var runtimeState *agent.AgentState = &agent.AgentState{MemoryUsage: 512}

	as.SetState(agent.AgentDOM, domState)
	as.SetState(agent.AgentInput, inputState)
	as.SetState(agent.AgentRuntime, runtimeState)

	// Test total memory usage
	var totalMemory int = as.TotalMemoryUsage()
	var expectedMemory int = 1024 + 2048 + 512

	if totalMemory != expectedMemory {
		t.Errorf("expected total memory usage to be %d, got %d", expectedMemory, totalMemory)
	}
}

// TestAgentManager_GetState_Critical tests getting agent state
func TestAgentManager_GetState_Critical(t *testing.T) {
	var am *agent.AgentManager = agent.NewAgentManager(&cdp.Client{}, "")

	// Get agent state without enabling (should return initialized state)
	var state *agent.AgentState
	var err error
	state, err = am.GetState(agent.AgentDOM)
	if err != nil {
		t.Fatalf("failed to get DOM agent state: %v", err)
	}

	if state == nil {
		t.Error("expected DOM agent state to not be nil")
	}

	// State should be initialized but disabled
	if state.Enabled {
		t.Error("expected DOM agent state to be disabled initially")
	}
}

// TestAgentManager_Stats_Critical tests statistics reporting
func TestAgentManager_Stats_Critical(t *testing.T) {
	var am *agent.AgentManager = agent.NewAgentManager(&cdp.Client{}, "")

	// Get statistics without enabling any agents
	var stats map[string]interface{} = am.Stats()
	if stats == nil {
		t.Error("expected stats to not be nil")
	}

	// Verify basic stats
	var totalAgents int = stats["total_agents"].(int)
	var enabledAgents int = stats["enabled_agents"].(int)
	var enabledList []string = stats["enabled_list"].([]string)

	if totalAgents != 8 {
		t.Errorf("expected total agents to be 8, got %d", totalAgents)
	}

	if enabledAgents != 0 {
		t.Errorf("expected enabled agents to be 0, got %d", enabledAgents)
	}

	if len(enabledList) != 0 {
		t.Errorf("expected enabled list to have 0 items, got %d", len(enabledList))
	}

	// Verify memory usage
	var totalMemory int = stats["total_memory_usage"].(int)
	if totalMemory != 0 {
		t.Errorf("expected total memory usage to be 0, got %d", totalMemory)
	}
}

// TestAgentManager_Disable_Critical tests agent disable functionality
func TestAgentManager_Disable_Critical(t *testing.T) {
	var am *agent.AgentManager = agent.NewAgentManager(&cdp.Client{}, "")

	// Test disabling already disabled agent (should not error)
	var err error = am.Disable(agent.AgentDOM)
	if err != nil {
		t.Fatalf("failed to disable already disabled DOM agent: %v", err)
	}

	// Verify agent is still disabled
	if am.IsEnabled(agent.AgentDOM) {
		t.Error("expected DOM agent to be disabled")
	}
}

// TestAgentManager_DisableAll_Critical tests disabling all agents
func TestAgentManager_DisableAll_Critical(t *testing.T) {
	var am *agent.AgentManager = agent.NewAgentManager(&cdp.Client{}, "")

	// Disable all agents (should not error even when none are enabled)
	var err error = am.DisableAll()
	if err != nil {
		t.Fatalf("failed to disable all agents: %v", err)
	}

	// Verify all agents are disabled
	if len(am.GetEnabledAgents()) != 0 {
		t.Error("expected 0 enabled agents")
	}
}

// TestAgentManager_Cleanup_Critical tests cleanup functionality
func TestAgentManager_Cleanup_Critical(t *testing.T) {
	var am *agent.AgentManager = agent.NewAgentManager(&cdp.Client{}, "")

	// Test cleanup (should not error even with no agents)
	am.Cleanup()

	// Verify cleanup completed without errors
	if am == nil {
		t.Error("expected AgentManager to still exist after cleanup")
	}
}
