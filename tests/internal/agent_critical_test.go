package internal_test

import (
	"strings"
	"sync"
	"testing"

	"github.com/jankylewis/kexas/internal/agent"
	"github.com/jankylewis/kexas/internal/cdp"
)

// 15 critical-after-existing unit tests for internal/agent + internal/cdp.
// Existing: cdp_test (Send), logger_test, agent_test (basic Manager).
// This batch: AgentStates / EnabledAgents / AgentContext edge cases,
// CDP command-registry queries, and AgentManager guard paths.

func TestEnabledAgents_EnableThenDisable_ReturnsFalse(t *testing.T) {
	var ea agent.EnabledAgents
	ea.Enable("DOM")
	if !ea.IsEnabled("DOM") {
		t.Error("Enable should make IsEnabled true")
	}
	ea.Disable("DOM")
	if ea.IsEnabled("DOM") {
		t.Error("Disable should make IsEnabled false")
	}
}

func TestEnabledAgents_DisableUnknown_NoOp(t *testing.T) {
	var ea agent.EnabledAgents
	ea.Disable("NeverWasEnabled") // should not panic
	if ea.IsEnabled("NeverWasEnabled") {
		t.Error("disabling an unknown agent should leave it disabled")
	}
}

func TestEnabledAgents_List_ReturnsAllEnabled(t *testing.T) {
	var ea agent.EnabledAgents
	ea.Enable(agent.AgentDOM)
	ea.Enable(agent.AgentInput)
	ea.Enable(agent.AgentRuntime)

	var got []string = ea.List()
	if len(got) != 3 {
		t.Errorf("expected 3 agents, got %d: %v", len(got), got)
	}
}

func TestEnabledAgents_ListEmpty_ReturnsEmpty(t *testing.T) {
	var ea agent.EnabledAgents
	var got []string = ea.List()
	if len(got) != 0 {
		t.Errorf("expected 0 agents, got %v", got)
	}
}

func TestAgentStates_GetState_UnknownAgent_ReturnsNil(t *testing.T) {
	var as agent.AgentStates
	if as.GetState("NotARealAgent") != nil {
		t.Error("GetState on unknown agent should return nil")
	}
}

func TestAgentStates_SetState_RoundTripsThroughGetState(t *testing.T) {
	var as agent.AgentStates
	var s *agent.AgentState = &agent.AgentState{Enabled: true}
	as.SetState(agent.AgentDOM, s)

	var got *agent.AgentState = as.GetState(agent.AgentDOM)
	if got == nil || !got.Enabled {
		t.Errorf("SetState/GetState round-trip failed: %+v", got)
	}
}

func TestAgentStates_ResetAll_ClearsEnabledFlags(t *testing.T) {
	var as agent.AgentStates
	as.SetState(agent.AgentDOM, &agent.AgentState{Enabled: true})
	as.SetState(agent.AgentInput, &agent.AgentState{Enabled: true})
	as.ResetAll()
	if s := as.GetState(agent.AgentDOM); s != nil && s.Enabled {
		t.Error("ResetAll should disable DOM")
	}
	if s := as.GetState(agent.AgentInput); s != nil && s.Enabled {
		t.Error("ResetAll should disable Input")
	}
}

func TestAgentStates_CountEnabled_OnlyTrueOnes(t *testing.T) {
	var as agent.AgentStates
	as.SetState(agent.AgentDOM, &agent.AgentState{Enabled: true})
	as.SetState(agent.AgentInput, &agent.AgentState{Enabled: false})
	as.SetState(agent.AgentRuntime, &agent.AgentState{Enabled: true})
	if c := as.CountEnabled(); c != 2 {
		t.Errorf("expected 2 enabled, got %d", c)
	}
}

func TestAgentContext_GetContext_UnknownAgent_ReturnsNil(t *testing.T) {
	var ac agent.AgentContext
	if got := ac.GetContext("NotARealAgent"); got != nil {
		t.Errorf("expected nil for unknown agent, got %v", got)
	}
}

func TestAgentContext_GetContext_NilFieldReturnsNil(t *testing.T) {
	// Even for a known agent, if the field is nil, GetContext returns nil.
	var ac agent.AgentContext
	if got := ac.GetContext(agent.AgentDOM); got != nil {
		t.Errorf("expected nil for unset DOM context, got %v", got)
	}
}

func TestAgentContext_ClearContext_NilSafe(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("ClearContext panicked: %v", r)
		}
	}()
	var ac agent.AgentContext
	ac.ClearContext(agent.AgentDOM) // unset; should not panic
	ac.ClearContext("NotARealAgent")
}

func TestCDP_GetCommand_KnownMethod_ReturnsTrue(t *testing.T) {
	var cmd cdp.Command
	var ok bool
	cmd, ok = cdp.GetCommand("Page.navigate")
	if !ok {
		t.Fatal("expected Page.navigate to be a known command")
	}
	if cmd.Method != "Page.navigate" {
		t.Errorf("returned cmd has wrong method: %v", cmd.Method)
	}
}

func TestCDP_GetCommand_UnknownMethod_ReturnsFalse(t *testing.T) {
	var _, ok = cdp.GetCommand("Bogus.notReal")
	if ok {
		t.Error("expected unknown method to return false")
	}
}

func TestCDP_AllCategories_NonEmpty(t *testing.T) {
	var cats []string = cdp.AllCategories()
	if len(cats) == 0 {
		t.Error("AllCategories should not be empty")
	}
	// Sanity: at least one category contains "Page" or similar.
	var found bool
	for _, c := range cats {
		if strings.Contains(c, "Page") || strings.Contains(c, "DOM") || strings.Contains(c, "Runtime") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected canonical categories, got %v", cats)
	}
}

func TestEnabledAgents_ConcurrentEnable_NoRace(t *testing.T) {
	// Stress-test concurrent Enable. EnabledAgents may need the mutex; test
	// asserts no panic and final state has all 8 standard agents enabled.
	const N int = 8
	var ea agent.EnabledAgents
	var wg sync.WaitGroup
	var agents []string = []string{
		agent.AgentDOM, agent.AgentInput, agent.AgentRuntime, agent.AgentNetwork,
		agent.AgentPage, agent.AgentSecurity, agent.AgentDebugger, agent.AgentProfiler,
	}
	for _, a := range agents {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			ea.Enable(name)
		}(a)
	}
	wg.Wait()

	var enabledCount int
	for _, a := range agents {
		if ea.IsEnabled(a) {
			enabledCount++
		}
	}
	if enabledCount != N {
		t.Errorf("expected all %d enabled after concurrent Enable, got %d", N, enabledCount)
	}
}
