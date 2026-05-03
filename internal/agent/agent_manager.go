package agent

import (
	"sync"
	"time"

	"github.com/jankylewis/kexas/internal/cdp"
	"github.com/jankylewis/kexas/internal/logger"
)

// AgentState represents the runtime state of a CDP agent
type AgentState struct {
	Enabled     bool        `json:"enabled"`
	LastUsed    time.Time   `json:"last_used"`
	Context     interface{} `json:"context"`
	MemoryUsage int         `json:"memory_usage"`
	Error       error       `json:"error"`
}

// Config holds configuration for AgentManager
type Config struct {
	AutoEnable     bool
	EnableTimeout  time.Duration
	ContextTimeout time.Duration
	Debug          bool
}

// DefaultConfig returns the default configuration for AgentManager
func DefaultConfig() *Config {
	return &Config{
		AutoEnable:     true,
		EnableTimeout:  5 * time.Second,
		ContextTimeout: 30 * time.Minute,
		Debug:          false,
	}
}

// AgentManager manages CDP agent enablement and state caching
type AgentManager struct {
	enabledAgents EnabledAgents
	agentContext  AgentContext
	agentState    AgentStates
	cdp           *cdp.Client
	sessionID     string
	mutex         sync.RWMutex
	config        *Config
	log           *logger.Logger
}

// NewAgentManager creates a new AgentManager instance
func NewAgentManager(cdpConn *cdp.Client, sessionID string) *AgentManager {
	return NewAgentManagerWithConfig(cdpConn, sessionID, DefaultConfig())
}

// NewAgentManagerWithConfig creates a new AgentManager with custom configuration
func NewAgentManagerWithConfig(cdpConn *cdp.Client, sessionID string, config *Config) *AgentManager {
	var am *AgentManager = &AgentManager{
		enabledAgents: EnabledAgents{}, // All agents start disabled
		agentContext:  AgentContext{},  // All contexts start empty
		agentState:    AgentStates{},   // All states start empty
		cdp:           cdpConn,
		sessionID:     sessionID,
		config:        config,
		log:           logger.New("agent"),
	}

	am.initAgentStates()

	return am
}

// initAgentStates sets up initial state for common CDP agents
func (am *AgentManager) initAgentStates() {
	var commonAgents []string = []string{
		AgentDOM, AgentInput, AgentRuntime, AgentNetwork,
		AgentPage, AgentSecurity, AgentDebugger, AgentProfiler,
	}

	for i := 0; i < len(commonAgents); i++ {
		var agent string = commonAgents[i]
		var agentState *AgentState = &AgentState{
			Enabled:     false,
			LastUsed:    time.Time{},
			Context:     nil,
			MemoryUsage: 0,
			Error:       nil,
		}
		am.agentState.SetState(agent, agentState)
	}
}
