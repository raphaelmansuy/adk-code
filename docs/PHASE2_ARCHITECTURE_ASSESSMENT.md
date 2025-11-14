# Phase 2 Architecture Assessment: Agent System & ADK Go Framework Integration

**Assessment Date**: November 14, 2025  
**Status**: ✅ Complete  
**Recommendation**: Proceed with refined architecture (see below)

## Executive Summary

The initial Phase 2 execution.go design is **fundamentally sound** but requires ADK integration refinement. The key insight: **execution.go should remain a pure utility package with NO ADK dependencies**, and a separate **run_agent tool** should bridge it to the ADK framework.

This separation of concerns provides:
- ✅ Clean architecture (domain logic ≠ framework integration)
- ✅ Testability (no mocking ADK framework needed)
- ✅ Reusability (execution logic works outside ADK context)
- ✅ Future-proofing (Phase 3 and 4 can build on stable foundation)

---

## 1. Google ADK Go Framework Analysis

### 1.1 Core Architecture

**ADK Framework Components** (from research/adk-go):
```
runner.Runner
    ├── Agent interface (must implement Name, Description, Run, SubAgents)
    ├── LLMAgent (delegates to models via llmagent/)
    ├── Session management (session/)
    ├── Tool system (tool/functiontool)
    └── Event streaming (Session.Event)
```

**Key Concepts**:
- **Agent**: Implements `agent.Agent` interface with `Run()` method
- **Tool**: Registered via `functiontool.New()` with Input/Output structs
- **Event**: Streamed from Run() method as `iter.Seq2[*session.Event, error]`
- **LLMAgent**: Coordinates model calls + tool invocations + agent delegation

### 1.2 Current adk-code Integration

**What Works**:
- ✅ Main coding assistant uses ADK's LLMAgent
- ✅ Tools registered via functiontool pattern
- ✅ Agent discovery exists in pkg/agents (Phase 1)
- ✅ Tool wrappers in tools/agents/ (list_agents, discover_paths)

**What's Missing**:
- ❌ No execution capability for discovered agents
- ❌ No tool to invoke agents
- ❌ No dependency resolution system
- ❌ No version constraint support

### 1.3 Tool Pattern (Currently Used)

All existing tools follow this pattern:

```go
// tools/agents/agents_tool.go
func NewListAgentsTool() (tool.Tool, error) {
    handler := func(ctx tool.Context, input ListAgentsInput) ListAgentsOutput {
        // Domain logic using pkg/agents
        discoverer := agents.NewDiscovererWithConfig(...)
        agents, err := discoverer.DiscoverAll()
        
        // Format for output
        return ListAgentsOutput{...}
    }
    
    return functiontool.New(functiontool.Config{
        Name:        "list_agents",
        Description: "...",
        InputType:   ListAgentsInput{},
        Handler:     handler,
    })
}
```

**Key Pattern Elements**:
1. Input struct with json tags and jsonschema comments
2. Output struct with json tags
3. Handler function receives tool.Context + Input
4. Handler returns Output directly
5. functiontool.New() wraps it for ADK

---

## 2. Phase 2 Execution System Assessment

### 2.1 Initial execution.go Design ✅

The execution.go structure is correct:

```go
// Domain logic - pure Go, no ADK dependencies
ExecutionContext       // Input parameters
ExecutionResult        // Output results
ExecutionRequirements  // System requirements
AgentRunner           // Main execution engine
  └── Execute()       // Core method
```

**Strengths**:
- ✅ Clean, focused responsibility
- ✅ Standard Go patterns (exec.Command, context.Context)
- ✅ No external dependencies
- ✅ Testable in isolation
- ✅ Reusable by multiple contexts

**What Was Missing**:
- ❌ ADK tool wrapper to expose to LLM
- ❌ Integration with functiontool pattern

### 2.2 Critical Insight: Two-Layer Architecture

```
Layer 1: Domain Logic (pkg/agents/)
├── execution.go       → ExecutionContext, ExecutionResult, AgentRunner
├── dependencies.go    → DependencyGraph, ResolveDependencies
├── version.go         → Version, Constraint parsing
└── *_test.go         → Unit tests (NO ADK mocks needed)

Layer 2: Framework Integration (tools/agents/)
├── run_agent.go       → ADK Tool wrapper for execution
├── resolve_deps.go    → ADK Tool wrapper for dependencies (future)
├── agents_tool.go     → Existing list_agents tool
└── *_test.go         → Functional tests (mock tool.Context)
```

**Why This Matters**:
- Domain logic can be tested WITHOUT ADK framework
- Tools can be updated without affecting execution logic
- Agents can be executed from CLI, HTTP, or other contexts
- Clear separation enables independent evolution

---

## 3. Phase 2 Implementation Strategy (Revised)

### 3.1 Week 1: Execution System (Domain + Tool)

#### Task 1.1: Core Execution (pkg/agents/execution.go)

**Status**: ✅ Ready to complete (fix file corruption and finalize)

**Components**:
```go
// Input context
ExecutionContext {
    Agent *Agent
    Params map[string]interface{}
    Timeout time.Duration
    WorkDir string
    Env map[string]string
    CaptureOutput bool
    Context context.Context
}

// Output result
ExecutionResult {
    Output string
    Error string
    ExitCode int
    Duration time.Duration
    Success bool
    Stderr string
    StartTime time.Time
    EndTime time.Time
}

// System requirements
ExecutionRequirements {
    SupportedOS []string
    MinGoVersion string
    MinMemoryMB int
    TimeoutSeconds int
    RequiredEnv []string
    Features []string
}

// Main execution engine
AgentRunner {
    Execute(ctx ExecutionContext) (*ExecutionResult, error)
    ValidateRequirements(req *ExecutionRequirements) error
    GetAgentByName(name string) (*Agent, error)
    ExecuteAndStream(ctx ExecutionContext) <-chan *ExecutionResult
}
```

**Key Methods**:
- `Execute()` - Run agent and capture output
- `ValidateRequirements()` - Check system compatibility
- `GetAgentByName()` - Retrieve agent by name
- `ExecuteAndStream()` - Stream results to channel

**Testing** (pkg/agents/execution_test.go):
- ✅ Context validation
- ✅ Result handling
- ✅ Timeout behavior
- ✅ Parameter passing
- ✅ Error handling
- ~15 test cases

#### Task 1.2: ADK Tool Wrapper (tools/agents/run_agent.go)

**NEW FILE**: Create run_agent.go following functiontool pattern

**Components**:
```go
// Tool input
RunAgentInput {
    AgentName string       // Which agent to run
    Params map[string]interface{} // Agent parameters
    Timeout int            // Seconds
    CaptureOutput bool     // Capture stdout/stderr
    Detailed bool          // Include detailed output
}

// Tool output
RunAgentOutput {
    Output string
    Error string
    ExitCode int
    Duration int64    // milliseconds
    Success bool
    Agent string       // Agent name
    StartTime string   // Timestamp
    EndTime string     // Timestamp
}

// Tool creator
NewRunAgentTool() (tool.Tool, error) {
    handler := func(ctx tool.Context, input RunAgentInput) RunAgentOutput {
        // Use pkg/agents execution
        discoverer := agents.NewDiscovererWithConfig(".")
        runner := agents.NewAgentRunner(discoverer)
        
        agent, err := runner.GetAgentByName(input.AgentName)
        if err != nil {
            return error output
        }
        
        result, err := runner.Execute(ExecutionContext{
            Agent: agent,
            Params: input.Params,
            Timeout: time.Duration(input.Timeout) * time.Second,
            CaptureOutput: input.CaptureOutput,
        })
        
        return RunAgentOutput{...}
    }
    
    return functiontool.New(functiontool.Config{
        Name: "run_agent",
        Description: "Execute an agent with parameters",
        InputType: RunAgentInput{},
        Handler: handler,
    })
}
```

**Key Features**:
- ✅ Finds agent by name using discoverer
- ✅ Validates agent requirements
- ✅ Passes parameters
- ✅ Captures output
- ✅ Returns formatted results for LLM

**Testing** (tools/agents/run_agent_test.go):
- ✅ Tool input validation
- ✅ Agent discovery
- ✅ Parameter passing
- ✅ Output formatting
- ✅ Error handling
- ~8 test cases

#### Task 1.3: Tool Registration

**Update**: tools/agents/agents_tool.go RegisterAgentTools()

```go
func RegisterAgentTools(reg *common.ToolRegistry) error {
    // Existing tools
    if err := reg.Register(NewListAgentsTool()); err != nil {
        return err
    }
    
    if err := reg.Register(NewDiscoverPathsTool()); err != nil {
        return err
    }
    
    // NEW: Agent execution
    if err := reg.Register(NewRunAgentTool()); err != nil {
        return err
    }
    
    return nil
}
```

**Or register in init()**:
```go
func init() {
    // Let existing pattern handle discovery
    _ = NewListAgentsTool
    _ = NewDiscoverPathsTool
    _ = NewRunAgentTool  // NEW
}
```

### 3.2 Week 2-3: Dependency System

#### Task 2.1: Dependency Resolution (pkg/agents/dependencies.go)

```go
DependencyGraph {
    Agents map[string]*Agent
    Edges map[string][]string
    
    AddAgent(agent *Agent) error
    ResolveDependencies(agentName string) ([]*Agent, error)
    DetectCycles() []string
    GetTransitiveDeps(agentName string) ([]string, error)
}
```

**Features**:
- ✅ Topological sorting
- ✅ Cycle detection
- ✅ Transitive dependency resolution
- ✅ Conflict detection

#### Task 2.2: Version System (pkg/agents/version.go)

```go
Version {
    Major, Minor, Patch int
    Prerelease string
    
    String() string
}

Constraint {
    Type ConstraintType  // ^, ~, >=, <=, ==, etc.
    Version *Version
    
    Matches(v *Version) bool
}
```

**Supported Constraints**:
- `^1.0.0` - Compatible versions
- `~1.0.0` - Patch versions
- `>=1.0.0`, `<=1.0.0`, `==1.0.0`
- `1.0.0 - 2.0.0` - Range

### 3.3 Week 3: Enhanced Metadata

**Extend** pkg/agents/agents.go:
- Add ExecutionRequirements struct to Agent
- Parse from YAML frontmatter
- Validation in Execute()

### 3.4 Week 4: Integration & Tools

**Optional Tools** (based on need):
- `resolve_deps` - Resolve agent dependencies
- `validate_agent` - Check agent requirements

**Documentation**: AGENT_EXECUTION.md with examples

---

## 4. ADK Integration Points

### 4.1 How run_agent Tool Integrates

```
LLM Agent (ADK)
    ↓ calls tool
run_agent Tool (functiontool)
    ↓ uses
AgentRunner (pkg/agents)
    ↓ executes
Discovered Agent (binary)
```

**Data Flow**:
1. LLM decides to run agent → calls run_agent tool
2. Tool receives RunAgentInput with agent name + params
3. Tool uses AgentRunner to find and execute agent
4. Tool returns RunAgentOutput with results
5. LLM receives results and continues

### 4.2 Future Integration (Phase 3: SubAgents)

Agents could become ADK SubAgents:
```go
// Future capability (not Phase 2)
agent1 := llmagent.New(llmagent.Config{
    Name: "task-processor",
    Tools: [...],
    Instruction: "Process tasks",
})

agent2 := agent.New(agent.Config{
    Name: "discovered-agent",
    Run: func(ctx agent.InvocationContext) iter.Seq2[*session.Event, error] {
        // Wrap discovered agent execution
        runner := agents.NewAgentRunner(discoverer)
        result, _ := runner.Execute(...)
        // Convert to Event and yield
    },
})

mainAgent := llmagent.New(llmagent.Config{
    SubAgents: []agent.Agent{agent1, agent2},
})
```

This is **Phase 3** - requires agent state management.

### 4.3 Marketplace Integration (Phase 4)

Version system + dependency resolution enable:
- Plugin registry queries
- Dependency compatibility checking
- Agent installation with version constraints
- Automatic dependency resolution

---

## 5. Design Principles

### 5.1 Separation of Concerns ✅

| Layer | Responsibility | Dependencies |
|-------|-----------------|--------------|
| pkg/agents/execution | Execute agents | None (pure Go) |
| pkg/agents/dependencies | Resolve deps | None (pure Go) |
| pkg/agents/version | Version constraints | None (pure Go) |
| tools/agents/run_agent | ADK tool wrapper | ADK tool framework |
| tools/agents/agents_tool | Agent discovery tool | ADK tool framework |

**Benefit**: Each layer can be tested, evolved, and replaced independently.

### 5.2 Testability ✅

```
Unit Tests (No Mocks)
├── pkg/agents/execution_test.go (15 tests)
├── pkg/agents/dependencies_test.go (12 tests)
└── pkg/agents/version_test.go (10 tests)

Functional Tests (Mock tool.Context)
├── tools/agents/run_agent_test.go (8 tests)
└── tools/agents/resolve_deps_test.go (8 tests)

Integration Tests (Real agents)
└── pkg/agents/execution_integration_test.go (6 tests)
```

**No ADK mocking needed for core logic** ← Critical advantage

### 5.3 Scalability ✅

- Execution: O(1) per agent
- Dependencies: O(n) with cycle detection
- Version matching: O(log n) per constraint
- Tool call: <100ms setup + execution time

---

## 6. Risk Assessment

| Risk | Mitigation | Priority |
|------|-----------|----------|
| Execution failures | Comprehensive error handling, timeout handling | HIGH |
| Circular dependencies | Cycle detection with clear error messages | HIGH |
| Version mismatches | Extensive version constraint tests | MEDIUM |
| Tool integration issues | Follow existing functiontool pattern exactly | MEDIUM |
| Performance bottlenecks | Benchmark dependency resolution | LOW |

---

## 7. Comparison: Pure Execution vs. Full Agent Lifecycle

### Why NOT Make Agent an ADK Agent (Phase 2)

**❌ Would Require**:
- agent.Agent interface implementation
- Session state management
- Event stream handling
- Context.InvocationContext integration
- Model integration if interactive
- Callback handling (before/after)
- Subagent lifecycle management

**❌ Complexity**: 2-3x more code
**❌ Testing**: Requires ADK mocks/framework setup
**❌ Scope**: Pushes Phase 2 into Phase 3

### ✅ Why Tool Wrapper Is Better (Phase 2)

**✅ Simpler**: Just wraps utility (AgentRunner)
**✅ Testable**: No ADK mocks needed
**✅ Follows Pattern**: Same as other tools
**✅ Reusable**: Works from CLI, HTTP, subagents
**✅ Staged**: Sets up Phase 3 agent delegation

---

## 8. Implementation Roadmap

### Phase 2 Week 1 (Execution)
```
✅ execution.go (180 LOC)
✅ execution_test.go (180 LOC)
✅ run_agent.go (220 LOC) - NEW
✅ run_agent_test.go (120 LOC) - NEW
= 700 LOC, 23 tests
```

### Phase 2 Week 2-3 (Dependencies)
```
✅ dependencies.go (200 LOC)
✅ dependencies_test.go (120 LOC)
✅ version.go (150 LOC)
✅ version_test.go (110 LOC)
= 580 LOC, 22 tests
```

### Phase 2 Week 3-4 (Metadata + Tools)
```
✅ Enhanced Agent struct with requirements
✅ YAML parsing for execution metadata
✅ resolve_deps.go (optional tool)
✅ Integration tests (150 LOC)
✅ Documentation (400 LOC)
= 550+ LOC, 10+ tests
```

**Total Phase 2**: ~1,430 LOC, 55+ tests, 85%+ coverage

---

## 9. Key Recommendations

### ✅ DO:
1. **Keep execution.go as pure utility** (no ADK imports)
2. **Create run_agent tool** to bridge to ADK
3. **Follow functiontool pattern** exactly
4. **Test pkg/agents without ADK** framework
5. **Version constraint system early** (needed for dependencies)
6. **Document ADK integration points** for Phase 3

### ❌ DON'T:
1. **Don't make Agent an ADK agent yet** (Phase 3)
2. **Don't mix framework with domain logic**
3. **Don't require Session/Runner for execution**
4. **Don't add ADK dependencies to pkg/agents**
5. **Don't skip testing utilities independently**

---

## 10. Conclusion

**Phase 2 is well-positioned with two-layer architecture:**

1. **Pure Go utilities** (pkg/agents/) - Fast, testable, reusable
2. **ADK tool wrappers** (tools/agents/) - Integrated with LLM

This provides:
- ✅ Clean separation of concerns
- ✅ Excellent testability
- ✅ Strong foundation for Phase 3 (agent delegation)
- ✅ Building blocks for Phase 4 (marketplace)

**Recommendation**: Proceed with refined implementation as described above.

---

## Appendix: File Structure Reference

```
adk-code/
├── pkg/agents/
│   ├── agents.go                    (EXISTING, Phase 1)
│   ├── agents_test.go               (EXISTING, Phase 1)
│   ├── config.go                    (EXISTING, Phase 1)
│   ├── config_test.go               (EXISTING, Phase 1)
│   ├── integration_test.go           (EXISTING, Phase 1)
│   │
│   ├── execution.go                 (NEW, Phase 2 Week 1)
│   ├── execution_test.go            (NEW, Phase 2 Week 1)
│   │
│   ├── dependencies.go              (NEW, Phase 2 Week 2)
│   ├── dependencies_test.go         (NEW, Phase 2 Week 2)
│   │
│   ├── version.go                   (NEW, Phase 2 Week 2)
│   ├── version_test.go              (NEW, Phase 2 Week 2)
│   │
│   └── execution_integration_test.go (NEW, Phase 2 Week 4)
│
└── tools/agents/
    ├── agents_tool.go               (EXISTING, Phase 1)
    ├── agents_tool_test.go          (EXISTING, Phase 1)
    ├── discover_paths.go            (EXISTING, Phase 1)
    ├── discover_paths_test.go       (EXISTING, Phase 1)
    │
    ├── run_agent.go                 (NEW, Phase 2 Week 1)
    ├── run_agent_test.go            (NEW, Phase 2 Week 1)
    │
    └── resolve_deps.go              (NEW, Phase 2 Week 2, optional)
```

---

**Assessment Complete** ✅  
**Status**: Ready for refined implementation  
**Next Steps**: Complete execution.go + create run_agent.go
