# Phase 2 Implementation Specification Index

**Current Status**: 50% Complete (5 of 10 specs detailed)  
**Last Updated**: November 15, 2025  
**Phase**: Phase 2 - Foundation & Integration  

---

## Quick Navigation

### 📋 Completed Specifications (5)
1. **[0001 - ExecutionContext Expansion](./implementation_plan/0001_execution_context_expansion.md)**
   - Extend ExecutionContext with session, memory, artifacts integration
   - Effort: 4 hours | Foundation for all subsequent specs

2. **[0002 - Memory & Artifact Interfaces](./implementation_plan/0002_memory_artifact_interfaces.md)**
   - Define Memory and Artifact service interfaces
   - Effort: 4 hours | Minimal implementation for Phase 2

3. **[0003 - Event-Based Execution Model](./implementation_plan/0003_event_based_execution_model.md)**
   - Transform Execute() to iter.Seq2[*session.Event, error] pattern
   - Effort: 6 hours | Critical blocker for event streaming

4. **[0004 - Agent-as-Tool Integration](./implementation_plan/0004_agent_as_tool_integration.md)**
   - Expose agents as callable tools through registry
   - Effort: 8 hours | Enables agent composition

5. **[0005 - Tool Registry Enhancement](./implementation_plan/0005_tool_registry_enhancement.md)**
   - Add dynamic discovery, filtering, and listing to registry
   - Effort: 6 hours | Supports REPL integration

### 📅 In Development (5)
6. **0006 - CLI & REPL Integration** (PENDING)
   - Real-time event display in REPL
   - Agent invocation and listing commands
   - Effort: 8 hours

7. **0007 - Session State Management** (PENDING)
   - Session persistence and state scoping
   - Event history storage
   - Effort: 6 hours

8. **0008 - Testing Framework** (PENDING)
   - Mock implementations and test fixtures
   - Integration and chaos testing
   - Effort: 6 hours

9. **0009 - Migration & Rollout** (PENDING)
   - Phased rollout timeline
   - Deprecation strategy
   - Effort: 4 hours

10. **0010 - Appendix & Reference** (PENDING)
    - API reference documentation
    - Example workflows and troubleshooting
    - Effort: 2 hours

---

## Key Concepts Overview

### The Five Foundational Changes (Specs 0001-0005)

#### 1. Extended Execution Context (Spec 0001)
**Problem**: ExecutionContext lacks session integration  
**Solution**: Add Session, User, InvocationID, Memory, Artifacts, State fields  
**Files**: 2 created, 1 modified | 4 hrs

```go
// Before (current)
type ExecutionContext struct {
    Agent       *Agent
    Params      map[string]interface{}
    Timeout     time.Duration
    // ... no session access ...
}

// After (Spec 0001)
type ExecutionContext struct {
    Agent         *Agent
    Params        map[string]interface{}
    Session       *session.Session         // NEW
    User          string                   // NEW
    InvocationID  string                   // NEW
    Memory        memory.Memory            // NEW
    Artifacts     artifact.Service         // NEW
    State         interface{}              // NEW
    // ... rest unchanged ...
}
```

**Impact**: Unlocks session integration, memory access, artifact storage

#### 2. Memory & Artifact Interfaces (Spec 0002)
**Problem**: No standard interfaces for memory and artifacts  
**Solution**: Define minimal interfaces with no-op Phase 2 implementations  
**Files**: 4 created | 4 hrs

```go
// Memory interface for semantic search and storage
type Memory interface {
    Save(ctx context.Context, content string, metadata map[string]interface{}) error
    Search(ctx context.Context, query string, limit int) ([]SearchResult, error)
    Get(ctx context.Context, id string) (string, error)
    Delete(ctx context.Context, id string) error
}

// Artifact interface for file/document storage
type Service interface {
    Save(ctx context.Context, artifact *Artifact) error
    Load(ctx context.Context, id string) (*Artifact, error)
    List(ctx context.Context) ([]*Artifact, error)
    Delete(ctx context.Context, id string) error
}
```

**Impact**: Agents can store and search memories; save execution artifacts

#### 3. Event-Based Execution (Spec 0003)
**Problem**: Execute() synchronous, can't stream progress  
**Solution**: Return iter.Seq2[*session.Event, error] instead of ExecutionResult  
**Files**: 5 created, 2 modified | 6 hrs

```go
// Before (current)
func (r *AgentRunner) Execute(ctx ExecutionContext) (*ExecutionResult, error) {
    // Synchronous: returns single result after completion
}

// After (Spec 0003)
func (r *AgentRunner) Execute(ctx ExecutionContext) iter.Seq2[*session.Event, error] {
    // Streams events: start, progress*, tool_call*, tool_result*, complete/error
    // Matches Google ADK Go's Runner.Run() pattern exactly
}

// Usage:
for event, err := range runner.Execute(ctx) {
    if err != nil {
        handleError(err)
        continue
    }
    handleEvent(event)
}
```

**Impact**: Real-time progress display; event-driven architecture; session persistence

#### 4. Agent-as-Tool (Spec 0004)
**Problem**: Agents and tools are separate systems; can't compose agents  
**Solution**: Wrap agents to implement tool.Tool interface  
**Files**: 5 created, 1 modified | 8 hrs

```go
// AgentTool wraps an Agent to be callable as a tool
type AgentTool struct {
    agent       *Agent
    runner      *AgentRunner
    description string
}

// Implements tool.Tool interface
func (t *AgentTool) Name() string { return t.agent.Name }
func (t *AgentTool) Description() string { return t.description }
func (t *AgentTool) IsLongRunning() bool { /* ... */ }
func (t *AgentTool) Execute(ctx context.Context, input *AgentInvocationInput) (*AgentInvocationOutput, error) {
    // Invoke agent with event streaming
}

// Usage: Agents can call other agents
registry.Register(ToolMetadata{Tool: agentTool, Category: CategoryAgents})
```

**Impact**: Agent composition; nested agent invocation; agents as building blocks

#### 5. Tool Registry Enhancement (Spec 0005)
**Problem**: Registry is static; can't discover agent tools dynamically  
**Solution**: Add discoverers and predicates for dynamic filtering  
**Files**: 4 created, 1 modified | 6 hrs

```go
// Register a discoverer for dynamic tool loading
registry.RegisterDiscoverer(func(ctx context.Context) ([]ToolMetadata, error) {
    // Discover agent tools from files
    agents, _ := discoverer.DiscoverAll()
    var tools []ToolMetadata
    for _, agent := range agents {
        tool, _ := NewAgentTool(agent, runner)
        tools = append(tools, ToolMetadata{Tool: tool, Category: CategoryAgents})
    }
    return tools, nil
})

// Apply filters to tool set
registry.RegisterFilter(AllowToolsPredicate([]string{"file-read", "echo"}))

// Discover and list all tools
tools, _ := registry.ListTools(ctx)
```

**Impact**: Dynamic agent discovery; context-based tool filtering; REPL integration

---

## Architecture Overview

### Event Flow
```
User Input
    ↓
AgentRunner.Execute(ctx)
    ↓
┌─ Event: start
│  Author: agent_name
├─ Event: progress
│  Content: output
├─ Event: tool_call (if applicable)
│  Data: tool_invocation
├─ Event: tool_result
│  Content: result
└─ Event: complete/error
   Success: true/false

Events automatically persisted in session.
```

### Agent Composition
```
LLM Agent
    ↓
Calls Tool: "analyze-code"
    ↓
AgentTool (agent_analyze_code)
    ↓
AgentRunner.Execute()
    ↓
Internal Agent Process
    ↓
Events streamed back
    ↓
Results integrated into parent agent
```

### Session Integration
```
Session {
    ID: "session-123"
    UserID: "user-456"
    AppName: "adk-code"
    State: map[string]interface{}
    Events: []*Event
}

Each agent invocation:
1. Gets session context
2. Can read/write state (app:, user:, temp: prefixes)
3. Yields events
4. Events auto-persisted
5. Memory & artifacts accessible
```

---

## Dependencies & Implementation Order

### Critical Path (Must Follow Order)
```
0001 (ExecutionContext)
    ↓
0002 (Memory/Artifacts)  [Can do in parallel with 0001]
    ↓
0003 (Event-Based Execution)
    ↓
0004 (Agent-as-Tool)
    ↓
0005 (Tool Registry)
    ↓
0006 (CLI/REPL)
    ↓
0007 (Session State)
```

### Timeline
- **Week 1**: Specs 0001-0005 (28 hours) → **Foundation Complete**
- **Week 2**: Specs 0006-0007 (14 hours) → **Integration Complete**
- **Week 3**: Spec 0008 (6 hours) → **Quality Complete**
- **Week 4**: Specs 0009-0010 (6 hours) → **Delivery Complete**

---

## Quality Standards

### Each Specification Includes
✅ Objective and problem statement  
✅ Design with code examples  
✅ 4-5 implementation steps with exact file paths  
✅ Comprehensive test code (unit + integration)  
✅ Backward compatibility analysis  
✅ Risk assessment with mitigations  
✅ Success criteria checklist  
✅ Google ADK Go comparison  

### Target Metrics
- **Code Coverage**: 80%+ across all changes
- **Test Count**: 10+ unit tests + 5+ integration tests per spec
- **Breaking Changes**: Zero
- **Backward Compat**: All old code works via ExecuteSync(), other wrappers
- **Performance**: Event streaming ~100ms/event

---

## Reference Documents

### Supporting Documentation
- **[COMPREHENSIVE_AUDIT_REPORT.md](../COMPREHENSIVE_AUDIT_REPORT.md)** - Full technical audit of Google ADK Go vs adk-code (8,000 words)
- **[PHASE2_ACTION_ITEMS.md](../PHASE2_ACTION_ITEMS.md)** - Breakdown of 20+ actionable tasks (6,000 words)
- **[VISUAL_ALIGNMENT_GUIDE.md](../VISUAL_ALIGNMENT_GUIDE.md)** - Diagrams and comparisons (3,500 words)
- **[IMPLEMENTATION_SPECS_STATUS.md](./IMPLEMENTATION_SPECS_STATUS.md)** - Detailed status and roadmap

### Implementation Specifications
All in `docs/spec/implementation_plan/` directory:
- 0001_execution_context_expansion.md (8 pages, 4 hrs)
- 0002_memory_artifact_interfaces.md (7 pages, 4 hrs)
- 0003_event_based_execution_model.md (10 pages, 6 hrs)
- 0004_agent_as_tool_integration.md (11 pages, 8 hrs)
- 0005_tool_registry_enhancement.md (10 pages, 6 hrs)
- 0006-0010 (PENDING - planned for next phase)

---

## Alignment with Google ADK Go

### Event Streaming Pattern
**Google ADK**: `Runner.Run() iter.Seq2[*session.Event, error]`  
**adk-code**: `AgentRunner.Execute() iter.Seq2[*session.Event, error]`  
**Alignment**: 100% ✅

### Tool Interface
**Google ADK**: `Tool interface { Name(), Description(), IsLongRunning() }`  
**adk-code**: `AgentTool struct implementing same interface`  
**Alignment**: 100% ✅

### Memory & Artifacts
**Google ADK**: Full implementations with embedding search  
**adk-code**: Minimal interfaces, no-op Phase 2, real Phase 3  
**Alignment**: 90% (same patterns, different scope)

### Session Management
**Google ADK**: Mutable session with state and event persistence  
**adk-code**: Session service with state scoping (app/user/temp)  
**Alignment**: 85% (different architecture, same capabilities)

### Agent Composition
**Google ADK**: Agents have SubAgents()  
**adk-code**: Agents become tools, compose via tool registry  
**Alignment**: 80% (different mechanisms, same capability)

---

## Success Criteria (Macro Level)

### Phase 2 Completion Checklist
- [ ] All 5 foundation specs (0001-0005) implemented
- [ ] 80%+ code coverage across changes
- [ ] All backward compatibility tests pass
- [ ] ExecuteSync() wraps new Execute() correctly
- [ ] Agent tools discoverable and invokable
- [ ] Event ordering invariants validated
- [ ] Memory and artifact access verified
- [ ] Session persistence works
- [ ] REPL integration complete (Spec 0006)
- [ ] Documentation 100% complete

### Definition of Done (Per Spec)
✅ Code implemented per specification  
✅ All test code written and passing  
✅ Backward compatibility verified  
✅ Integration tests pass  
✅ Documentation updated  
✅ Code review approved  

---

## Getting Started as a Developer

### Step 1: Understand the Audit
Read these to understand why changes are needed:
- `docs/COMPREHENSIVE_AUDIT_REPORT.md` (20 min read)
- `docs/PHASE2_ACTION_ITEMS.md` (15 min scan)

### Step 2: Understand This Phase
Read this document (you're doing it!)
- Focus on the "Five Foundational Changes" section
- Review Architecture Overview

### Step 3: Pick Your Spec
Start with **Spec 0001** or **Spec 0002** (can be done in parallel)
- Read the entire spec first
- Understand objectives and design
- Review test code
- Examine Google ADK Go comparisons

### Step 4: Implement Following Steps
Each spec has 4-5 implementation steps
- Create/modify files in exact order specified
- Implement tests alongside code
- Run tests after each step
- Verify no breaking changes

### Step 5: Cross-Reference
While implementing:
- Check Google ADK Go (research/adk-go/)
- Verify patterns match
- Review actual code from both systems
- Validate with test code examples

### Step 6: Move to Next Spec
After implementation + review:
- Mark as "Ready for Merge"
- Move to next spec in dependency order
- Don't skip specs (dependencies exist for reason)

---

## FAQ

**Q: Why iterate over executing directly?**  
A: Go 1.22+ iterators are more efficient, allow consumer control, support real-time streaming

**Q: Why no-op implementations for Memory/Artifacts?**  
A: Unblocks Phase 2 development; real implementations (embeddings, databases) done in Phase 3

**Q: Why agent-as-tool wrapper instead of native agents?**  
A: Keeps adk-code's strategic file-based discovery; unifies with existing tool system

**Q: How does backward compatibility work?**  
A: ExecuteSync() wraps new Execute() and collects all events; old code calls ExecuteSync()

**Q: What if I find a bug in completed specs?**  
A: Specs are frozen after review. Document as Phase 2.X issue for later patching.

**Q: Can I skip a spec?**  
A: No. Each spec depends on previous ones. Must implement in order.

**Q: How long will full implementation take?**  
A: ~54 hours total. Week 1: foundation, Week 2: integration, Week 3: testing, Week 4: delivery

---

## Quick Ref: File Changes by Spec

### Spec 0001 (ExecutionContext)
- Create: `pkg/agents/execution_context.go`
- Create: `pkg/agents/execution_context_test.go`
- Modify: `pkg/agents/execution.go` (add imports)

### Spec 0002 (Memory/Artifacts)
- Create: `pkg/memory/memory.go`
- Create: `pkg/memory/memory_test.go`
- Create: `pkg/artifact/artifact.go`
- Create: `pkg/artifact/artifact_test.go`
- Modify: `pkg/agents/execution.go` (add Memory, Artifacts fields)

### Spec 0003 (Event-Based Execution)
- Create: `internal/session/event.go`
- Create: `internal/session/event_test.go`
- Create: `pkg/agents/execution_event_test.go`
- Modify: `pkg/agents/execution.go` (new Execute(), keep ExecuteSync())
- Modify: `pkg/agents/types.go` (ExecutionContext updates)

### Spec 0004 (Agent-as-Tool)
- Create: `pkg/agents/agent_tool.go`
- Create: `pkg/agents/agent_tool_test.go`
- Create: `pkg/agents/agent_tool_types.go`
- Create: `pkg/agents/agent_tool_registry.go`
- Create: `pkg/agents/agent_tool_integration_test.go`
- Modify: `tools/base/registry.go` (add CategoryAgents)

### Spec 0005 (Tool Registry)
- Modify: `tools/base/registry.go` (add Discover, Filter, ListTools)
- Create: `tools/base/predicates.go`
- Create: `tools/base/listing.go`
- Create: `tools/base/registry_enhance_test.go`
- Create: `pkg/agents/agent_tool_discoverer.go`
- Create: `pkg/agents/agent_tool_discovery_test.go`

---

**Status**: Ready for Phase 2 Implementation  
**Next**: Begin Spec 0001 Implementation  
**Questions**: See FAQ or reference COMPREHENSIVE_AUDIT_REPORT.md
