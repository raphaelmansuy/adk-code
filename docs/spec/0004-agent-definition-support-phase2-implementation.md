# Phase 2 Implementation Plan - Agent Execution & Versioning

**Status**: 📋 Planning  
**Planned Start**: November 15, 2025  
**Planned Duration**: 4-5 weeks  
**Target Completion**: December 19, 2025  
**Feature Branch**: `feat/agent-definition-support-phase2`  
**Base**: `feat/agent-definition-support-phase1`

## Overview

Phase 2 extends Phase 1's discovery system into a fully functional agent invocation platform. This phase adds agent execution, dependency resolution, version constraint support, and enhanced metadata handling.

## Phase 1 Recap

**Delivered**:
- ✅ Multi-path agent discovery (project, user, plugin)
- ✅ Configuration system with YAML + environment variables
- ✅ Metadata support (version, author, tags, dependencies)
- ✅ CLI tools (list_agents, discover_paths)
- ✅ 63 tests, 88.3% coverage
- ✅ Backward compatible, production-ready

**Quality Metrics**:
- 1,200+ LOC of production code
- 5 new feature commits
- Clean git history
- Zero breaking changes

## Phase 2 Scope Definition

### In Scope ✅

1. **Agent Execution System**
   - Agent invocation capability
   - Parameter passing and validation
   - Output capture and formatting
   - Error handling and reporting
   - Execution context management

2. **Dependency Resolution**
   - Dependency graph building
   - Topological sorting
   - Circular dependency detection
   - Transitive dependency handling
   - Conflict detection

3. **Version Constraint System**
   - Semantic version parsing
   - Version constraint syntax (^, ~, >=, <=, etc.)
   - Version matching and resolution
   - Version conflict detection

4. **Enhanced Metadata**
   - Execution requirements (OS, Go version, etc.)
   - Agent capabilities and features
   - Performance metadata
   - Risk classifications

5. **Agent Marketplace Foundation**
   - Agent metadata standardization
   - Agent discovery enhancement
   - Plugin registry format
   - Agent package specification

6. **Testing & Documentation**
   - Execution system tests
   - Dependency resolution tests
   - Integration tests
   - Phase 2 documentation

### Out of Scope ❌

- Remote agent execution (Phase 3)
- Agent auto-update/upgrade (Phase 3)
- Claude Code integration (Phase 3)
- Agent marketplace (Phase 4)
- Web UI (Phase 4)

## Detailed Implementation Plan

### Task 1: Agent Execution System (Week 1-2)

**Goal**: 400-500 LOC  
**Priority**: CRITICAL

#### 1.1 Execution Context
- [ ] Create `pkg/agents/execution.go` (~150 LOC)
  - `ExecutionContext` struct with agent parameters
  - `ExecutionResult` struct with output/error
  - Timeout and resource limits
  - Environment variable handling
  - Working directory management

**Interfaces**:
```go
type ExecutionContext struct {
    Agent      *Agent
    Params     map[string]interface{}
    Timeout    time.Duration
    WorkDir    string
    Env        map[string]string
}

type ExecutionResult struct {
    Output     string
    Error      string
    ExitCode   int
    Duration   time.Duration
    Success    bool
}

type Executor interface {
    Execute(ctx ExecutionContext) (*ExecutionResult, error)
}
```

#### 1.2 Agent Runner
- [ ] Extend `pkg/agents/agents.go` (~150 LOC)
  - `AgentRunner` struct with execution capability
  - `Run()` method for execution
  - Parameter validation
  - Output formatting
  - Error handling

**Key Methods**:
```go
func (a *Agent) Validate() error
func (a *Agent) GetExecutionRequirements() *ExecutionReqs
func (d *Discoverer) GetAgent(name string) (*Agent, error)
func (r *AgentRunner) Execute(ctx ExecutionContext) (*ExecutionResult, error)
```

#### 1.3 CLI Execution Tool
- [ ] Create `tools/agents/run_agent.go` (~200 LOC)
  - New `run_agent` tool
  - Parameter passing interface
  - Output formatting
  - Error reporting

**Tool Input**:
```go
type RunAgentInput struct {
    AgentName  string                 // Agent to run
    Params     map[string]interface{} // Parameters
    Timeout    int                    // Seconds
    CaptureOut bool                   // Capture output
}
```

#### 1.4 Tests
- [ ] Create `pkg/agents/execution_test.go` (~150 LOC)
  - Execution context tests
  - Result validation tests
  - Timeout handling tests
  - Parameter passing tests
  - Error handling tests

**Metrics**:
- Code: 500 LOC
- Tests: 15+ test cases
- Coverage: >85%

---

### Task 2: Dependency Resolution (Week 2-3)

**Goal**: 350-400 LOC  
**Priority**: HIGH

#### 2.1 Dependency Resolver
- [ ] Create `pkg/agents/dependencies.go` (~200 LOC)
  - `DependencyGraph` struct
  - `ResolveDependencies()` function
  - Topological sorting
  - Circular dependency detection
  - Transitive dependency handling

**Key Functions**:
```go
type DependencyGraph struct {
    Agents map[string]*Agent
    Edges  map[string][]string // agent -> dependencies
}

func (dg *DependencyGraph) AddAgent(agent *Agent) error
func (dg *DependencyGraph) ResolveDependencies(agentName string) ([]*Agent, error)
func (dg *DependencyGraph) DetectCycles() []string
func (dg *DependencyGraph) GetTransitiveDeps(agentName string) ([]string, error)
```

#### 2.2 Version Constraint System
- [ ] Create `pkg/agents/version.go` (~150 LOC)
  - Semantic version parsing
  - Version constraint matching
  - Version range handling
  - Conflict detection

**Constraints Supported**:
- `1.0.0` - Exact version
- `^1.0.0` - Compatible versions (^1.0.0 matches >=1.0.0, <2.0.0)
- `~1.0.0` - Patch versions (~1.0.0 matches >=1.0.0, <1.1.0)
- `>=1.0.0` - Greater than or equal
- `<=1.0.0` - Less than or equal
- `1.0.0 - 2.0.0` - Range

**API**:
```go
type Version struct {
    Major, Minor, Patch int
    Prerelease          string
}

func ParseVersion(s string) (*Version, error)
func ParseConstraint(s string) (*Constraint, error)
func (c *Constraint) Matches(v *Version) bool
```

#### 2.3 Tests
- [ ] Create tests for dependencies and versions (~150 LOC)
  - Dependency graph tests
  - Cycle detection tests
  - Version parsing tests
  - Constraint matching tests
  - Conflict detection tests

**Metrics**:
- Code: 350 LOC
- Tests: 12+ test cases
- Coverage: >85%

---

### Task 3: Enhanced Metadata (Week 3)

**Goal**: 150-200 LOC  
**Priority**: MEDIUM

#### 3.1 Execution Requirements
- [ ] Extend `pkg/agents/agents.go` (~100 LOC)
  - `ExecutionRequirements` struct
  - OS/architecture constraints
  - Resource requirements
  - Plugin dependencies
  - Feature flags

**Schema**:
```yaml
---
name: my-agent
version: 1.0.0
execution:
  os: [linux, darwin]
  go_version: ">=1.24"
  memory_mb: 512
  timeout_seconds: 300
  features: [file-io, network]
---
```

#### 3.2 Capability Tracking
- [ ] Add capability fields to Agent struct (~50 LOC)
  - Supported input types
  - Output formats
  - Integration points
  - Performance characteristics

**Metrics**:
- Code: 150 LOC
- Tests: 6+ test cases

---

### Task 4: Integration & Documentation (Week 4)

**Goal**: 250-300 LOC + Docs  
**Priority**: CRITICAL

#### 4.1 Integration Tests
- [ ] Create `pkg/agents/execution_integration_test.go` (~150 LOC)
  - Full execution workflows
  - Dependency resolution scenarios
  - Version constraint scenarios
  - Error handling workflows

#### 4.2 Documentation
- [ ] Create `docs/AGENT_EXECUTION.md` (~300 LOC)
  - Agent execution guide
  - Dependency resolution docs
  - Version constraint reference
  - Execution requirements spec
  - Examples and use cases

#### 4.3 Examples
- [ ] Create example agents
  - Simple executable agent
  - Agent with dependencies
  - Agent with version constraints
  - Agent with custom execution requirements

#### 4.4 Finalization
- [ ] Phase 2 completion report
- [ ] Performance benchmarks
- [ ] Security review
- [ ] Code quality verification

**Metrics**:
- Integration code: 150 LOC
- Documentation: 300+ LOC
- 5+ example agents

---

## Architecture Enhancements

### Execution Flow

```
Agent Discovery (Phase 1)
        ↓
Agent Selection & Validation
        ↓
Dependency Resolution
        ↓
Version Constraint Checking
        ↓
Execution Context Setup
        ↓
Agent Execution (Run)
        ↓
Output Capture & Formatting
        ↓
Result Return
```

### Data Structures

```go
// Agent Execution
Agent → ExecutionContext → Executor → ExecutionResult

// Dependency Resolution
DependencyGraph → ResolveDependencies() → []*Agent

// Version Management
Version → Constraint → VersionResolver
```

### File Structure

```
pkg/agents/
├── agents.go           (updated: +50 LOC)
├── agents_test.go      (updated: +30 LOC)
├── execution.go        (NEW: 150 LOC)
├── execution_test.go   (NEW: 150 LOC)
├── dependencies.go     (NEW: 200 LOC)
├── dependencies_test.go (NEW: 100 LOC)
├── version.go          (NEW: 150 LOC)
├── version_test.go     (NEW: 100 LOC)
└── execution_integration_test.go (NEW: 150 LOC)

tools/agents/
├── run_agent.go        (NEW: 200 LOC)
├── run_agent_test.go   (NEW: 100 LOC)
└── agents_tool.go      (updated: +20 LOC)
```

## Success Criteria

### Functional Requirements
- [ ] Agents can be executed with parameters
- [ ] Dependencies are resolved correctly
- [ ] Version constraints are enforced
- [ ] Circular dependencies are detected
- [ ] Execution timeouts work
- [ ] Output is captured correctly

### Quality Requirements
- [ ] Code coverage >85%
- [ ] All tests passing
- [ ] Clean git history
- [ ] Documentation complete
- [ ] Backward compatible
- [ ] Performance <100ms execution setup

### Testing Requirements
- [ ] 20+ new test cases
- [ ] 5+ integration tests
- [ ] 3+ example agents
- [ ] Stress tests for large dependency graphs

## Timeline & Milestones

| Week | Focus | Deliverable | LOC |
|------|-------|-------------|-----|
| 1-2 | Execution System | run_agent tool | 500 |
| 2-3 | Dependencies | Dependency resolver | 350 |
| 3 | Metadata | Execution requirements | 150 |
| 4 | Integration | Tests + docs | 400 |
| **Total** | | | **1,400+** |

## Risk Mitigation

| Risk | Mitigation | Priority |
|------|-----------|----------|
| Circular dependencies | Cycle detection tests | HIGH |
| Performance bottlenecks | Benchmark & optimize | HIGH |
| Version mismatch issues | Comprehensive version tests | MEDIUM |
| Execution failures | Error handling tests | MEDIUM |
| Documentation gaps | Include examples | LOW |

## Dependencies on Phase 1

**Required Phase 1 Deliverables**:
- ✅ Agent discovery system
- ✅ Configuration system
- ✅ Metadata support
- ✅ CLI framework
- ✅ Test infrastructure

**Phase 1 APIs to Leverage**:
- `Discoverer.DiscoverAll()` - Get all agents
- `Agent` struct with metadata
- `LoadConfig()` for execution config
- Tool registration framework

## Phase 3 Preparation

**Phase 2 Outputs for Phase 3**:
- Stable execution API
- Version constraint infrastructure
- Dependency resolution system
- Test suite for validation
- Documentation framework

## Success Metrics

**Code Quality**:
- Coverage: >85% (target 90%)
- Tests: 20+ new test cases
- Commits: 6-8 logical commits
- Documentation: Complete with examples

**Performance**:
- Agent execution setup: <100ms
- Dependency resolution: <50ms for 100-agent graph
- Version constraint check: <1ms per constraint

**User Experience**:
- Clear error messages
- Intuitive parameter interface
- Helpful execution logging
- Complete documentation

## Conclusion

Phase 2 will transform the agent discovery system into a fully functional agent invocation platform. With execution capabilities, dependency resolution, and version management, the system will be ready for Phase 3's Claude Code integration and Phase 4's marketplace features.

**Target**: 1,400+ LOC, 20+ tests, complete documentation, production-ready execution system.
