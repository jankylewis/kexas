# 🏗️ CHROME AGENT ARCHITECTURE DEEP DIVE

## 🔍 **What's Behind an Agent? (DOM Agent Example)**

### **DOM Agent Internal Architecture**

```
┌─────────────────────────────────────────────────────────────┐
│                    DOM Agent Internals                        │
├─────────────────────────────────────────────────────────────┤
│  CDP Command Layer                                           │
│  ├── DOM.enable()           // Enable the agent             │
│  ├── DOM.disable()          // Disable the agent            │
│  ├── DOM.getDocument()      // Get root document node       │
│  ├── DOM.querySelector()    // Find single element          │
│  ├── DOM.querySelectorAll() // Find multiple elements       │
│  └── DOM.performSearch()    // XPath search                 │
├─────────────────────────────────────────────────────────────┤
│  Blink Engine Integration                                    │
│  ├── Document Object Model (DOM Tree)                       │
│  ├── Node Management (creation, deletion, modification)      │
│  ├── CSS Selector Engine                                     │
│  ├── XPath Engine                                            │
│  └── Event Handling                                         │
├─────────────────────────────────────────────────────────────┤
│  Memory Management                                           │
│  ├── Node References (weak/strong)                          │
│  ├── Node IDs (unique identifiers)                          │
│  ├── Garbage Collection                                      │
│  └── Memory Pool Management                                  │
├─────────────────────────────────────────────────────────────┤
│  Security & Sandboxing                                       │
│  ├── Same-Origin Policy                                     │
│  ├── Cross-Origin Restrictions                              │
│  ├── Permission Checks                                       │
│  └── Isolation from other processes                         │
└─────────────────────────────────────────────────────────────┘
```

### **DOM Agent Lifecycle**

```javascript
// 1. Agent Enablement
DOM.enable() → Creates DOM agent instance → Allocates memory → Registers with Blink

// 2. Document Access
DOM.getDocument() → Gets root node → Assigns nodeId (1) → Returns node info

// 3. Element Finding
DOM.querySelector({nodeId: 1, selector: "#button"}) 
→ CSS selector engine parses selector
→ Traverses DOM tree
→ Finds matching node
→ Assigns nodeId (2) to found element
→ Returns node info with nodeId

// 4. Element Interaction
DOM.getAttributes({nodeId: 2}) → Gets element attributes
DOM.setAttributeValue({nodeId: 2, name: "value", value: "test"}) → Sets attribute
```

### **Memory Management Details**

```javascript
// Node ID System
nodeId 1: Document root
nodeId 2: <html> element
nodeId 3: <body> element
nodeId 4: <div id="button"> element
nodeId 5: <span> element inside button

// Reference Counting
nodeId 4: refCount = 1 (held by automation framework)
nodeId 5: refCount = 1 (held by automation framework)

// Garbage Collection
// When refCount = 0 → Node can be garbage collected
// When page navigates → All nodeIds invalidated
```

### **Security Model**

```javascript
// Same-Origin Policy
DOM.querySelector({selector: "#button"}) // ✅ Same origin - allowed
DOM.querySelector({selector: "iframe#cross-origin button"}) // ❌ Cross-origin - blocked

// Permission Checks
DOM.setAttributeValue({nodeId: 4, name: "value", value: "sensitive"}) // ✅ Allowed
DOM.setAttributeValue({nodeId: 4, name: "type", value: "file"}) // ❌ Blocked (security)
```

---

## 🎭 **AGENT COMPARISON: PLAYWRIGHT VS KEXAS**

### **Playwright's Smart Enablement Strategy**

```javascript
// Playwright Internal Implementation (simplified)
class Page {
  constructor() {
    this.enabledAgents = new Set();
    this.cdpConnection = new CDPConnection();
  }

  async click(selector) {
    // Smart enablement logic
    if (!this.enabledAgents.has('DOM')) {
      await this.cdpConnection.send('DOM.enable');
      this.enabledAgents.add('DOM');
    }
    
    if (!this.enabledAgents.has('Input')) {
      await this.cdpConnection.send('Input.enable');
      this.enabledAgents.add('Input');
    }

    // Now perform the actual click
    const element = await this.findElement(selector);
    await this.performClick(element);
  }

  async findElement(selector) {
    // DOM agent is guaranteed to be enabled here
    return await this.cdpConnection.send('DOM.querySelector', {
      nodeId: this.documentNodeId,
      selector: selector
    });
  }
}
```

### **Agent Caching Strategy**

```javascript
// Playwright's agent state management
const agentState = {
  DOM: {
    enabled: false,
    documentNodeId: null,
    lastUsed: null,
    memoryUsage: 0
  },
  Input: {
    enabled: false,
    lastUsed: null,
    queuedEvents: []
  },
  Runtime: {
    enabled: false,
    executionContextId: null,
    isolatedWorlds: []
  }
};

// Smart enablement with caching
async function ensureAgentEnabled(agentName) {
  const agent = agentState[agentName];
  
  if (!agent.enabled) {
    await cdp.send(`${agentName}.enable`);
    agent.enabled = true;
    agent.lastUsed = Date.now();
    
    // Initialize agent-specific state
    if (agentName === 'DOM') {
      const {root} = await cdp.send('DOM.getDocument');
      agent.documentNodeId = root.nodeId;
    }
  }
  
  return agent;
}
```

---

## 🧠 **AGENT ENABLEMENT PATTERNS**

### **Pattern 1: Lazy Enablement (Playwright)**
```javascript
// Enable only when first needed
async function click(selector) {
  await ensureAgentEnabled('DOM');  // Enables if needed, skips if already enabled
  await ensureAgentEnabled('Input'); // Same logic
  
  // Perform operation
  const element = await findElement(selector);
  await performClick(element);
}
```

### **Pattern 2: Eager Enablement (Selenium)**
```javascript
// Enable all agents at startup
async function createPage() {
  await cdp.send('DOM.enable');
  await cdp.send('Input.enable');
  await cdp.send('Runtime.enable');
  await cdp.send('Network.enable');
  // ... all agents enabled immediately
}
```

### **Pattern 3: Hybrid Enablement (Current Kexas)**
```go
// Kexas uses a hybrid approach:
// - Stealth domains (Network, Page) are enabled eagerly in attachToPage()
// - Operational domains (DOM, Runtime) are enabled lazily via AgentManager.EnsureAgent()
// - Input domain is auto-enabled (no .enable() call needed)

func (p *Page) sendCommand(method string, params map[string]interface{}) {
    // ensureAgentsForCommand auto-enables the required domain
    p.ensureAgentsForCommand(method)  // e.g. "Runtime.evaluate" → EnsureAgent("Runtime")
    p.client.SendToSession(p.ctx, p.sessionID, method, params)
}
```

---

## 🔄 **AGENT LIFECYCLE MANAGEMENT**

### **Enablement Flow**
```
User Action → Framework Method → Agent Check → Enable if Needed → Execute Operation
     ↓              ↓                ↓              ↓                ↓
page.click() → click() → isDOMEnabled? → DOM.enable() → performClick()
                           ↓
                    isInputEnabled? → Input.enable()
```

### **Disablement Flow**
```
Page Close → Framework Cleanup → Agent Disable → Memory Cleanup
     ↓              ↓                ↓              ↓
page.close() → cleanup() → DOM.disable() → freeMemory()
                           ↓
                    Input.disable()
```

### **Memory Management**
```javascript
// Node reference tracking
const nodeReferences = new Map(); // nodeId → {refCount, lastAccessed}

// Automatic cleanup
setInterval(() => {
  const now = Date.now();
  for (const [nodeId, info] of nodeReferences) {
    if (info.refCount === 0 && (now - info.lastAccessed) > 30000) {
      // Release node after 30 seconds of inactivity
      cdp.send('DOM.discardNode', {nodeId});
      nodeReferences.delete(nodeId);
    }
  }
}, 10000);
```

---

## 🚀 **OPTIMIZATION STRATEGIES**

### **1. Smart Caching**
```javascript
// Cache frequently accessed elements
const elementCache = new Map(); // selector → {nodeId, timestamp}

async function getCachedElement(selector) {
  const cached = elementCache.get(selector);
  if (cached && (Date.now() - cached.timestamp) < 5000) {
    // Use cached element if less than 5 seconds old
    return cached.nodeId;
  }
  
  // Find fresh element
  const element = await findElement(selector);
  elementCache.set(selector, {nodeId: element.nodeId, timestamp: Date.now()});
  return element.nodeId;
}
```

### **2. Batch Operations**
```javascript
// Enable multiple agents in one call
async function enableAgents(agents) {
  const promises = agents.map(agent => 
    !agentState[agent].enabled ? cdp.send(`${agent}.enable`) : Promise.resolve()
  );
  await Promise.all(promises);
  agents.forEach(agent => agentState[agent].enabled = true);
}
```

### **3. Predictive Enablement**
```javascript
// Enable agents based on usage patterns
async function predictAndEnable(operation) {
  const requiredAgents = getRequiredAgents(operation);
  await enableAgents(requiredAgents);
}

const operationAgents = {
  'click': ['DOM', 'Input'],
  'type': ['DOM', 'Input'],
  'evaluate': ['Runtime'],
  'waitForRequest': ['Network'],
  'screenshot': ['Page']
};
```

---

## 🎯 **IMPLEMENTATION FOR KEXAS**

### **Smart Agent Manager**
```go
type AgentManager struct {
    enabledAgents map[string]bool
    cdp           *CDPConnection
    agentState    map[string]*AgentState
    mutex         sync.RWMutex
}

type AgentState struct {
    Enabled     bool      `json:"enabled"`
    LastUsed    time.Time `json:"last_used"`
    MemoryUsage int       `json:"memory_usage"`
    Context     interface{} `json:"context"`
}

func (am *AgentManager) EnsureAgent(agentName string) error {
    am.mutex.Lock()
    defer am.mutex.Unlock()
    
    state := am.agentState[agentName]
    if !state.Enabled {
        // Enable the agent
        err := am.cdp.Call(agentName + ".enable")
        if err != nil {
            return fmt.Errorf("failed to enable %s agent: %w", agentName, err)
        }
        
        state.Enabled = true
        state.LastUsed = time.Now()
        
        // Initialize agent-specific context
        switch agentName {
        case "DOM":
            // Get document root node
            result, err := am.cdp.Call("DOM.getDocument")
            if err == nil {
                state.Context = result
            }
        }
    }
    
    state.LastUsed = time.Now()
    return nil
}

func (am *AgentManager) FindElement(selector string) (*Element, error) {
    // Ensure DOM agent is enabled
    if err := am.EnsureAgent("DOM"); err != nil {
        return nil, err
    }
    
    // Now perform the search
    return am.cdp.Call("DOM.querySelector", map[string]interface{}{
        "nodeId": am.agentState["DOM"].Context.(map[string]interface{})["root"].(map[string]interface{})["nodeId"],
        "selector": selector,
    })
}
```

### **Usage in Page Methods**
```go
func (p *Page) Find(selector string) (*Element, error) {
    return p.agentManager.FindElement(selector)
}

func (p *Page) Click(selector string) error {
    // Ensure both DOM and Input agents are enabled
    if err := p.agentManager.EnsureAgent("DOM"); err != nil {
        return err
    }
    if err := p.agentManager.EnsureAgent("Input"); err != nil {
        return err
    }
    
    // Perform click operation
    element, err := p.Find(selector)
    if err != nil {
        return err
    }
    
    return p.performClick(element)
}
```

---

*This deep dive shows how agents work internally and how to implement smart enablement like Playwright.* 🚀
