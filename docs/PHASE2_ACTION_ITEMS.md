# Phase 2 Implementation Action Items
**Status**: READY FOR EXECUTION  
**Priority**: HIGH - Blocks Phase 2 Completion  
**Target Date**: December 1, 2025  
**Owner**: adk-code Team

---

## Overview

This document breaks down the comprehensive audit report into executable action items organized by priority, dependency, and effort.

## P1: Critical Blockers (Must Complete Phase 2)

### P1.1: Event-Based Execution Model Refactor

**Epic**: Convert ExecutionContext to support event streaming  
**Status**: NOT STARTED  
**Priority**: CRITICAL  
**Effort**: 3-4 days  
**Blocker**: Yes - blocks all agent execution

#### Tasks

##### P1.1.1: Update ExecutionContext Structure
**File**: `pkg/agents/execution.go`  
**Effort**: 2 hours  

```go
// CHANGE: Add session-related fields to ExecutionContext
type ExecutionContext struct {
    // Existing fields (keep)
    Agent             *Agent
    Params            map[string]interface{}
    Timeout           time.Duration
    WorkDir           string
    Env               map[string]string
    CaptureOutput     bool
    ReturnRawOutput   bool
    Context           context.Context
    
    // NEW FIELDS (add these)
    Session           *session.Session         // From internal/session
    Memory            memory.Memory            // From pkg/memory (create)
    Artifacts         artifact.Service         // From pkg/artifact (create)
    State             session.State            // From internal/session
    User              string                   // User ID
    InvocationID      string                   // Invocation tracking
    FunctionCallID    string                   // For tool context
    EventActions      *session.EventActions    // From internal/session
}
```

**Acceptance Criteria**:
- [ ] ExecutionContext includes all new fields
- [ ] Old fields remain unchanged
- [ ] Compile with no errors
- [ ] Backward compat: old code still works with zero values

##### P1.1.2: Create Memory Interface
**File**: `pkg/memory/memory.go` (NEW)  
**Effort**: 4 hours  

```go
package memory

import "context"

type Memory interface {
    // Save stores a memory entry
    Save(ctx context.Context, content string, metadata map[string]interface{}) error
    
    // Search performs semantic search
    Search(ctx context.Context, query string, limit int) ([]SearchResult, error)
    
    // Get retrieves specific memory by ID
    Get(ctx context.Context, id string) (string, error)
    
    // Delete removes memory entry
    Delete(ctx context.Context, id string) error
}

type SearchResult struct {
    ID       string
    Content  string
    Score    float32
    Metadata map[string]interface{}
}

// DefaultMemory returns a no-op implementation for now
func DefaultMemory() Memory {
    return &noopMemory{}
}

type noopMemory struct{}

func (n *noopMemory) Save(ctx context.Context, content string, metadata map[string]interface{}) error {
    return nil
}

func (n *noopMemory) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
    return []SearchResult{}, nil
}

func (n *noopMemory) Get(ctx context.Context, id string) (string, error) {
    return "", nil
}

func (n *noopMemory) Delete(ctx context.Context, id string) error {
    return nil
}
```

**Acceptance Criteria**:
- [ ] Memory interface defined
- [ ] No-op implementation available
- [ ] Compile with no errors
- [ ] Ready for Phase 3 implementation

##### P1.1.3: Create Artifact Interface
**File**: `pkg/artifact/artifact.go` (NEW)  
**Effort**: 4 hours  

```go
package artifact

import "context"

type Service interface {
    // Save stores an artifact
    Save(ctx context.Context, artifact *Artifact) error
    
    // Load retrieves an artifact by ID
    Load(ctx context.Context, id string) (*Artifact, error)
    
    // List returns all artifacts
    List(ctx context.Context) ([]*Artifact, error)
    
    // Delete removes an artifact
    Delete(ctx context.Context, id string) error
}

type Artifact struct {
    ID       string
    Name     string
    Type     string // "file", "document", "result"
    Content  []byte
    Metadata map[string]interface{}
}

// DefaultService returns a no-op implementation for now
func DefaultService() Service {
    return &noopService{}
}

type noopService struct{}

func (n *noopService) Save(ctx context.Context, artifact *Artifact) error {
    return nil
}

func (n *noopService) Load(ctx context.Context, id string) (*Artifact, error) {
    return nil, nil
}

func (n *noopService) List(ctx context.Context) ([]*Artifact, error) {
    return []*Artifact{}, nil
}

func (n *noopService) Delete(ctx context.Context, id string) error {
    return nil
}
```

**Acceptance Criteria**:
- [ ] Artifact interface defined
- [ ] No-op implementation available
- [ ] Compile with no errors
- [ ] Ready for Phase 3 implementation

##### P1.1.4: Update AgentRunner.Execute() Signature
**File**: `pkg/agents/execution.go`  
**Effort**: 8 hours  

**Current**:
```go
func (ar *AgentRunner) Execute(ctx ExecutionContext) (*ExecutionResult, error)
```

**New**:
```go
// Change return type to event iterator
func (ar *AgentRunner) Execute(ctx ExecutionContext) iter.Seq2[*session.Event, error] {
    return func(yield func(*session.Event, error) bool) {
        // Initialize execution
        startTime := time.Now()
        
        // Yield start event
        if !yield(&session.Event{
            Type: "agent.start",
            Timestamp: time.Now(),
            Data: map[string]interface{}{
                "agent": ctx.Agent.Name,
                "invocation_id": ctx.InvocationID,
            },
        }, nil) {
            return
        }
        
        // Execute agent (process-based for now, event-based in future)
        output, err := executeAgentProcess(ctx)
        
        if err != nil {
            // Yield error event
            yield(&session.Event{
                Type: "agent.error",
                Timestamp: time.Now(),
                Data: map[string]interface{}{
                    "error": err.Error(),
                },
            }, nil)
            return
        }
        
        // Yield result event
        if !yield(&session.Event{
            Type: "agent.result",
            Timestamp: time.Now(),
            Data: map[string]interface{}{
                "output": output,
                "duration": time.Since(startTime),
            },
        }, nil) {
            return
        }
        
        // Yield completion event
        yield(&session.Event{
            Type: "agent.complete",
            Timestamp: time.Now(),
        }, nil)
    }
}
```

**Acceptance Criteria**:
- [ ] Execute() returns iter.Seq2[*session.Event, error]
- [ ] Events are yielded in order: start, result/error, complete
- [ ] Session events are persisted
- [ ] Errors propagated as error events
- [ ] Tests pass with event assertions

##### P1.1.5: Update AgentRunner Type Definition
**File**: `pkg/agents/execution.go`  
**Effort**: 1 hour  

```go
// Update to include session context
type AgentRunner struct {
    discoverer *Discoverer
    session    *session.Session  // NEW
    config     *config.Config     // NEW
}

// Update constructor
func NewAgentRunner(discoverer *Discoverer, sess *session.Session, cfg *config.Config) *AgentRunner {
    return &AgentRunner{
        discoverer: discoverer,
        session:    sess,
        config:     cfg,
    }
}
```

**Acceptance Criteria**:
- [ ] AgentRunner includes session reference
- [ ] Constructor updated
- [ ] All callers updated
- [ ] No compilation errors

#### Test Cases for P1.1

**File**: `pkg/agents/execution_test.go`

```go
func TestExecute_YieldsStartEvent(t *testing.T) {
    runner := NewAgentRunner(...)
    ctx := ExecutionContext{...}
    
    events := make([]*session.Event, 0)
    for event, err := range runner.Execute(ctx) {
        require.NoError(t, err)
        events = append(events, event)
    }
    
    require.Greater(t, len(events), 0)
    require.Equal(t, "agent.start", events[0].Type)
}

func TestExecute_YieldsCompleteEvent(t *testing.T) {
    runner := NewAgentRunner(...)
    ctx := ExecutionContext{...}
    
    events := make([]*session.Event, 0)
    for event, err := range runner.Execute(ctx) {
        require.NoError(t, err)
        events = append(events, event)
    }
    
    require.Equal(t, "agent.complete", events[len(events)-1].Type)
}

func TestExecute_YieldsErrorEvent_OnFailure(t *testing.T) {
    runner := NewAgentRunner(...)
    ctx := ExecutionContext{Agent: invalidAgent}
    
    hasError := false
    for event, err := range runner.Execute(ctx) {
        if event.Type == "agent.error" {
            hasError = true
        }
    }
    
    require.True(t, hasError)
}
```

---

### P1.2: Agent-as-Tool Integration

**Epic**: Make agents callable as tools by other agents  
**Status**: NOT STARTED  
**Priority**: CRITICAL  
**Effort**: 2-3 days  
**Blocker**: Yes - blocks subagent delegation

#### Tasks

##### P1.2.1: Create agenttool Package
**File**: `tools/agents/agenttool/agenttool.go` (NEW)  
**Effort**: 3 hours  

```go
package agenttool

import (
    "context"
    "strings"
    
    "google.golang.org/adk/tool"
    "adk-code/pkg/agents"
)

// NewAgentTool wraps an agent as a callable tool
func NewAgentTool(agent *agents.Agent) (tool.Tool, error) {
    return &agentTool{
        agent: agent,
    }, nil
}

type agentTool struct {
    agent *agents.Agent
}

func (at *agentTool) Name() string {
    // Convert agent name to tool name: "code-reviewer" -> "agent_code_reviewer"
    return "agent_" + strings.ReplaceAll(at.agent.Name, "-", "_")
}

func (at *agentTool) Description() string {
    return at.agent.Description
}

func (at *agentTool) IsLongRunning() bool {
    // Agents are long-running by nature
    return true
}
```

**Acceptance Criteria**:
- [ ] agentTool implements tool.Tool interface
- [ ] Name conversion correct (kebab-case to underscore)
- [ ] Package compiles
- [ ] Ready for tool registration

##### P1.2.2: Create Agent Tool Factory
**File**: `tools/agents/agenttool/factory.go` (NEW)  
**Effort**: 3 hours  

```go
package agenttool

import (
    "context"
    
    "google.golang.org/adk/tool"
    "google.golang.org/adk/tool/functiontool"
    "adk-code/pkg/agents"
)

// InputSchema for agent tool invocation
type InputSchema struct {
    Request string                 `json:"request"`
    Params  map[string]interface{} `json:"params,omitempty"`
}

// OutputSchema for agent tool results
type OutputSchema struct {
    Output   string `json:"output"`
    Error    string `json:"error,omitempty"`
    ExitCode int    `json:"exit_code"`
}

// NewAgentToolWithSchema creates a tool from agent with proper schema
func NewAgentToolWithSchema(agent *agents.Agent, runner *agents.AgentRunner) (tool.Tool, error) {
    toolName := "agent_" + strings.ReplaceAll(agent.Name, "-", "_")
    
    return functiontool.New(functiontool.Config{
        Name: toolName,
        Description: agent.Description,
        InputType: &InputSchema{},
        Fn: func(ctx tool.Context, input *InputSchema) (*OutputSchema, error) {
            // Create execution context
            execCtx := agents.ExecutionContext{
                Agent:  agent,
                Params: input.Params,
                // Fill in other fields from ctx
            }
            
            // Execute agent and collect events
            output := ""
            var lastErr string
            exitCode := 0
            
            for event, err := range runner.Execute(execCtx) {
                if err != nil {
                    lastErr = err.Error()
                    break
                }
                
                if event.Type == "agent.result" {
                    if val, ok := event.Data["output"]; ok {
                        output = val.(string)
                    }
                }
                if event.Type == "agent.error" {
                    if val, ok := event.Data["error"]; ok {
                        lastErr = val.(string)
                    }
                }
            }
            
            return &OutputSchema{
                Output:   output,
                Error:    lastErr,
                ExitCode: exitCode,
            }, nil
        },
    })
}
```

**Acceptance Criteria**:
- [ ] Factory creates tools with proper schema
- [ ] InputSchema captures agent request
- [ ] OutputSchema returns results
- [ ] Event processing correct
- [ ] Error handling works

##### P1.2.3: Update Tool Registry to Include Agent Tools
**File**: `pkg/models/registry.go`  
**Effort**: 2 hours  

```go
// Add this method to Registry or ToolSet
func (r *Registry) DiscoverAgentTools(ctx context.Context) ([]tool.Tool, error) {
    agentTools := make([]tool.Tool, 0)
    
    // Discover agents
    discoverer := agents.NewDiscoverer(".")
    result, err := discoverer.DiscoverAll()
    if err != nil {
        return nil, err
    }
    
    // Convert subagents to tools
    for _, agent := range result.Agents {
        if agent.Type == agents.TypeSubagent {
            // Create agent tool
            agentTool, err := agenttool.NewAgentToolWithSchema(agent, r.runner)
            if err != nil {
                // Log error but continue
                continue
            }
            agentTools = append(agentTools, agentTool)
        }
    }
    
    return agentTools, nil
}

// Update GetTools to include agent tools
func (r *Registry) GetTools(ctx context.Context) ([]tool.Tool, error) {
    // Get built-in tools
    tools, err := r.getBuiltInTools(ctx)
    if err != nil {
        return nil, err
    }
    
    // Add agent tools
    agentTools, err := r.DiscoverAgentTools(ctx)
    if err == nil {
        tools = append(tools, agentTools...)
    }
    
    return tools, nil
}
```

**Acceptance Criteria**:
- [ ] DiscoverAgentTools() discovers subagents
- [ ] Converts agents to tools
- [ ] GetTools() includes agent tools
- [ ] Proper error handling

#### Test Cases for P1.2

**File**: `tools/agents/agenttool/agenttool_test.go`

```go
func TestNewAgentTool_NameFormat(t *testing.T) {
    agent := &agents.Agent{
        Name: "code-reviewer",
        Description: "Reviews code",
    }
    
    tool, err := agenttool.NewAgentTool(agent)
    require.NoError(t, err)
    require.Equal(t, "agent_code_reviewer", tool.Name())
}

func TestNewAgentTool_IsLongRunning(t *testing.T) {
    agent := &agents.Agent{Name: "test"}
    tool, err := agenttool.NewAgentTool(agent)
    require.NoError(t, err)
    require.True(t, tool.IsLongRunning())
}

func TestNewAgentToolWithSchema_Execution(t *testing.T) {
    agent := &agents.Agent{
        Name: "test-agent",
        Description: "Test",
    }
    runner := NewMockAgentRunner()
    
    tool, err := agenttool.NewAgentToolWithSchema(agent, runner)
    require.NoError(t, err)
    require.NotNil(t, tool)
}
```

---

### P1.3: Update CLI Tools for Event-Based Execution

**Epic**: Update /run-agent and related commands  
**Status**: PARTIAL  
**Priority**: HIGH  
**Effort**: 2 days  
**Blocker**: Yes - user-facing feature

#### Tasks

##### P1.3.1: Update /run-agent Command
**File**: `tools/agents/run_agent.go`  
**Effort**: 4 hours  

```go
// Current implementation needs update for event streaming
// Changes:
// 1. Accept event iterator from Execute()
// 2. Stream events to display
// 3. Format output properly

func (c *RunAgentCommand) Execute(ctx context.Context, args ...string) error {
    // ... parse args, find agent ...
    
    execCtx := agents.ExecutionContext{
        Agent:   agent,
        Params:  parseParams(args),
        Timeout: 5 * time.Minute,
        Context: ctx,
    }
    
    // Collect all events
    events := make([]*session.Event, 0)
    
    // Execute and stream events
    for event, err := range c.runner.Execute(execCtx) {
        if err != nil {
            // Display error event
            fmt.Fprintf(c.display, "Error: %v\n", err)
            return err
        }
        
        // Collect event
        events = append(events, event)
        
        // Display event based on type
        switch event.Type {
        case "agent.start":
            fmt.Fprintf(c.display, "▶️  Running agent: %s\n", agent.Name)
        case "agent.result":
            if output, ok := event.Data["output"]; ok {
                fmt.Fprintf(c.display, "✅ Result:\n%s\n", output)
            }
        case "agent.error":
            if errMsg, ok := event.Data["error"]; ok {
                fmt.Fprintf(c.display, "❌ Error: %s\n", errMsg)
            }
        case "agent.complete":
            fmt.Fprintf(c.display, "✓ Complete\n")
        }
    }
    
    return nil
}
```

**Acceptance Criteria**:
- [ ] Processes event iterator
- [ ] Displays events appropriately
- [ ] Error handling works
- [ ] Backward compat with old commands

##### P1.3.2: Update Validation/Lint Commands
**File**: `tools/agents/validate_agent.go`, `tools/agents/lint_agent.go`  
**Effort**: 2 hours  

No changes needed - these don't use Execute(). Just verify they still work.

**Acceptance Criteria**:
- [ ] Commands still work
- [ ] Tests pass
- [ ] No regressions

##### P1.3.3: Create New /agent-invoke Command
**File**: `tools/agents/invoke_agent.go` (NEW)  
**Effort**: 3 hours  

This is for calling agents as tools from the REPL:

```go
package agents

// /agent-invoke code-reviewer "Review this code"
// Invokes an agent as a tool with the given request

func NewInvokeAgentTool(runner *AgentRunner) *FunctionTool {
    return functiontool.New(functiontool.Config{
        Name: "invoke_agent",
        Description: "Invoke an agent as a tool",
        InputType: &InvokeAgentInput{},
        Fn: func(ctx tool.Context, input *InvokeAgentInput) (*InvokeAgentOutput, error) {
            // Find agent by name
            discoverer := NewDiscoverer(".")
            result, _ := discoverer.DiscoverAll()
            
            var targetAgent *Agent
            for _, agent := range result.Agents {
                if agent.Name == input.AgentName {
                    targetAgent = agent
                    break
                }
            }
            
            if targetAgent == nil {
                return &InvokeAgentOutput{
                    Error: "Agent not found: " + input.AgentName,
                }, nil
            }
            
            // Execute agent
            execCtx := ExecutionContext{
                Agent:  targetAgent,
                Params: input.Params,
            }
            
            // Collect results from events
            var output string
            for event, err := range runner.Execute(execCtx) {
                if err != nil {
                    return &InvokeAgentOutput{
                        Error: err.Error(),
                    }, nil
                }
                if event.Type == "agent.result" {
                    if val, ok := event.Data["output"]; ok {
                        output = val.(string)
                    }
                }
            }
            
            return &InvokeAgentOutput{
                Output: output,
            }, nil
        },
    })
}

type InvokeAgentInput struct {
    AgentName string                 `json:"agent_name"`
    Request   string                 `json:"request"`
    Params    map[string]interface{} `json:"params,omitempty"`
}

type InvokeAgentOutput struct {
    Output string `json:"output"`
    Error  string `json:"error,omitempty"`
}
```

**Acceptance Criteria**:
- [ ] Tool created and registered
- [ ] Finds agent by name
- [ ] Executes with parameters
- [ ] Returns results properly

#### Test Cases for P1.3

**File**: `tools/agents/run_agent_test.go`

```go
func TestRunAgent_StreamsEvents(t *testing.T) {
    cmd := NewRunAgentCommand(mockRunner, mockDisplay)
    err := cmd.Execute(context.Background(), "test-agent")
    
    require.NoError(t, err)
    // Verify display was called with events
}

func TestRunAgent_HandlesError(t *testing.T) {
    cmd := NewRunAgentCommand(mockRunner, mockDisplay)
    err := cmd.Execute(context.Background(), "invalid-agent")
    
    require.Error(t, err)
    // Verify error was displayed
}
```

---

## P2: Important But Not Blocking

### P2.1: Session State Interface
**File**: `internal/session/state.go`  
**Effort**: 2 hours  
**Status**: NOT STARTED

Implement State interface from Google ADK for key-value store.

### P2.2: Update internal/session Package
**Files**: `internal/session/*.go`  
**Effort**: 4 hours  
**Status**: PARTIAL

Make sure Session implements necessary interfaces:
- [ ] Session.State() returns State interface
- [ ] Events properly persisted
- [ ] Event iteration works

### P2.3: REPL Event Display
**File**: `internal/repl/repl.go`  
**Effort**: 3 hours  
**Status**: BLOCKED (depends on P1.3)

Update REPL to display events in real-time during agent execution.

---

## P3: Nice-to-Have

### P3.1: Before/After Callbacks
**File**: `pkg/agents/execution.go`  
**Effort**: 2 hours

Add BeforeAgentCallbacks and AfterAgentCallbacks like Google ADK.

### P3.2: Agent Transfer
**File**: `pkg/agents/execution.go`  
**Effort**: 3 hours

Implement ability to transfer between agents mid-execution.

### P3.3: Error Callbacks
**File**: `pkg/agents/execution.go`  
**Effort**: 2 hours

Add error callback support for custom error handling.

---

## Implementation Sequence

### Week 1: Core Refactoring

1. **Day 1**: P1.1.1-1.1.3 (ExecutionContext, Memory, Artifact)
2. **Day 2**: P1.1.4-1.1.5 (Execute() refactoring, AgentRunner)
3. **Day 3**: P1.1 tests, verification
4. **Day 4**: P1.2.1-1.2.2 (agenttool package)
5. **Day 5**: P1.2.3 (Tool registry), verification

### Week 2: Tool Integration & CLI

1. **Day 1-2**: P1.2 tests
2. **Day 3-4**: P1.3.1-1.3.3 (CLI updates)
3. **Day 5**: P1.3 tests, verification

### Week 3: Integration & Polish

1. **Day 1-2**: P2.1-2.3 (Session/REPL integration)
2. **Day 3-4**: Documentation updates
3. **Day 5**: Final testing, release prep

---

## Acceptance Criteria Checklist

### Code Quality
- [ ] All code reviewed by team
- [ ] 85%+ test coverage on new code
- [ ] Zero compilation warnings
- [ ] All tests passing
- [ ] No performance regressions

### Functionality
- [ ] All P1 items complete
- [ ] Event streaming works
- [ ] Agent tools discoverable
- [ ] CLI commands updated
- [ ] REPL displays events

### Documentation
- [ ] COMPREHENSIVE_AUDIT_REPORT.md updated
- [ ] API docs for new functions
- [ ] Migration guide for breaking changes
- [ ] Examples for new features

### Testing
- [ ] Unit tests: 85%+ coverage
- [ ] Integration tests: End-to-end workflows
- [ ] Manual testing: All commands verified
- [ ] Performance: No regressions

---

## Risk Mitigation

### Risk: Breaking Changes
**Mitigation**: Feature flag new APIs, deprecate old ones gracefully

### Risk: Integration Issues
**Mitigation**: Integration tests written before implementation

### Risk: Performance Degradation
**Mitigation**: Baseline benchmarks before changes, verify after

### Risk: Event Streaming Bugs
**Mitigation**: Extensive testing of iterator pattern, edge cases

---

## Success Metrics

### Pre-Implementation
- [ ] Audit report reviewed and approved
- [ ] All P1 tasks broken down
- [ ] Resource allocation confirmed
- [ ] Timeline agreed upon

### Post-Implementation
- [ ] All tests passing
- [ ] Code coverage 85%+
- [ ] Performance baseline maintained
- [ ] No critical bugs in production
- [ ] User documentation complete
- [ ] Team training completed

---

## Communication Plan

### Weekly Standups
- Tuesday 10:00 AM: Progress review
- Friday 4:00 PM: Week recap

### Documentation
- Daily: Update status in this document
- Weekly: Summary in logs/
- Bi-weekly: Report to stakeholders

### Code Review
- All PRs reviewed before merge
- At least 1 approval required
- Tests must pass before merge

---

## Next Steps

1. **Review This Document** - Team alignment on tasks
2. **Create Feature Branch** - `feat/phase2-execution-refactor`
3. **Start Week 1** - Begin P1.1.1 implementation
4. **Daily Sync** - Progress updates
5. **Iterative Delivery** - Weekly PRs with working code

---

**Owner**: adk-code Team  
**Last Updated**: November 15, 2025  
**Status**: READY FOR IMPLEMENTATION  
**Next Review**: After Week 1 completion
