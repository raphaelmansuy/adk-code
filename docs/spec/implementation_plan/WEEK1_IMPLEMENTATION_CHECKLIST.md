# Phase 2 Week 1: ExecutionContext & Event Streaming - Implementation Checklist

**Sprint**: Week 1 (Nov 18-22, 2025)  
**Objective**: Implement Specs 0001-0003 (ExecutionContext expansion + Event-Based Execution)  
**Target Duration**: 12 hours (3 days)  
**Status**: Ready to Start  

---

## Sprint Overview

### Objectives
1. ✅ Expand ExecutionContext with session/memory/artifact integration
2. ✅ Create session/event infrastructure
3. ✅ Implement event-streaming execution (ExecuteStream)
4. ✅ Maintain 100% backward compatibility

### Success Criteria
- [ ] ExecutionContext compiles with all new fields
- [ ] ExecuteStream method implemented and working
- [ ] Event types defined and validated
- [ ] All new code has unit tests (80%+ coverage)
- [ ] Integration tests passing
- [ ] Backward compatibility verified
- [ ] make check passes 100%

### Deliverables
1. Expanded ExecutionContext (Spec 0001)
2. Session/Event infrastructure (foundation for Spec 0003)
3. Event-streaming executor (Spec 0003)
4. Comprehensive test suite
5. Integration documentation

---

## Task Breakdown

### Task 1.1: Create Session Infrastructure (4 hours)

**File**: `internal/session/session.go` (NEW)  
**Status**: Not Started  
**Effort**: 4 hours  

**Deliverables**:
- [ ] Session interface definition
  - ID() string
  - AppName() string  
  - UserID() string
  - State() State
  - Events() Events
  - LastUpdateTime() time.Time
  
- [ ] State interface definition
  - Get(key string) (any, error)
  - Set(key string, value any) error
  - Delete(key string) error
  - All() iter.Seq2[string, any]
  
- [ ] Event type definition
  - ID, Timestamp, InvocationID, Author, Type, Content
  - Data interface{}, ExecutionTime, Success, Error
  - EventType constants (start, progress, tool_call, complete, error, etc.)
  
- [ ] EventActions type (for advanced use)
  - TransferAgent, EndInvocation, etc.
  
- [ ] InMemoryState implementation
  - map-backed state with RWMutex
  - Get/Set/Delete operations

**Code Template**:
```go
package session

import (
    "time"
    "iter"
)

// Session represents a conversation session
type Session interface {
    ID() string
    AppName() string
    UserID() string
    State() State
    Events() Events
    LastUpdateTime() time.Time
}

// Event represents a discrete step in execution
type Event struct {
    ID            string
    Timestamp     time.Time
    InvocationID  string
    Author        string
    Type          string
    Content       string
    Data          interface{}
    ExecutionTime time.Duration
    Success       bool
    Error         string
}

// EventType constants
const (
    EventTypeStart      = "start"
    EventTypeProgress   = "progress"
    EventTypeToolCall   = "tool_call"
    EventTypeToolResult = "tool_result"
    EventTypeComplete   = "complete"
    EventTypeError      = "error"
    EventTypePartial    = "partial"
)

// State interface for key-value state
type State interface {
    Get(string) (any, error)
    Set(string, any) error
    Delete(string) error
    All() iter.Seq2[string, any]
}

// Implementation goes here...
```

**Tests** (`internal/session/session_test.go`):
- [ ] Test Event creation
- [ ] Test State Get/Set/Delete
- [ ] Test State iteration
- [ ] Test state thread safety

**Acceptance Criteria**:
- Session interface compiles
- Event type created with all fields
- State implementation works with Get/Set/Delete
- Thread-safe state operations
- No external dependencies

---

### Task 1.2: Expand ExecutionContext (3 hours)

**File**: `pkg/agents/execution.go` (MODIFY)  
**Status**: Not Started  
**Effort**: 3 hours  

**Current ExecutionContext**:
```go
type ExecutionContext struct {
    Agent           *Agent
    Params          map[string]interface{}
    Timeout         time.Duration
    WorkDir         string
    Env             map[string]string
    CaptureOutput   bool
    ReturnRawOutput bool
    Context         context.Context
}
```

**Changes Required**:
- [ ] Add Session field: `Session *session.Session`
- [ ] Add Memory field: `Memory memory.Memory`
- [ ] Add Artifacts field: `Artifacts artifact.Service`
- [ ] Add State field: `State session.State`
- [ ] Add User field: `User string`
- [ ] Add InvocationID field: `InvocationID string`
- [ ] Add FunctionCallID field: `FunctionCallID string`
- [ ] Add EventActions field: `EventActions *session.EventActions`

**New Structure**:
```go
type ExecutionContext struct {
    // Existing fields (unchanged)
    Agent           *Agent
    Params          map[string]interface{}
    Timeout         time.Duration
    WorkDir         string
    Env             map[string]string
    CaptureOutput   bool
    ReturnRawOutput bool
    Context         context.Context
    
    // New fields (Spec 0001)
    Session       *session.Session
    Memory        memory.Memory
    Artifacts     artifact.Service
    State         session.State
    User          string
    InvocationID  string
    FunctionCallID string
    EventActions  *session.EventActions
}
```

**Additional Functions**:
- [ ] ValidateExecutionContext() error
- [ ] NewExecutionContext() constructor (optional)
- [ ] WithSession() fluent builder (optional)
- [ ] WithTimeout() fluent builder (optional)

**Tests** (`pkg/agents/execution_test.go`):
- [ ] Test ExecutionContext creation
- [ ] Test field initialization
- [ ] Test ValidateExecutionContext with missing required fields
- [ ] Test backward compatibility (all new fields optional)

**Acceptance Criteria**:
- All new fields added
- Compiles without errors
- Existing Execute() method still works
- New fields are zero-valued (optional)
- Validation catches critical issues
- 100% backward compatible

---

### Task 1.3: Add ExecuteStream Method (5 hours)

**File**: `pkg/agents/execution.go` (ADD)  
**Status**: Not Started  
**Effort**: 5 hours  

**New Method Signature**:
```go
// ExecuteStream runs the agent and streams execution events.
// Returns an iterator that yields events as they occur.
// Blocks until execution completes or context is cancelled.
//
// Example:
//  ctx := ExecutionContext{...}
//  for event, err := range runner.ExecuteStream(ctx) {
//      if err != nil {
//          log.Printf("Error: %v", err)
//          return
//      }
//      log.Printf("Event: %v (%v)", event.Type, event.Content)
//  }
//
func (r *AgentRunner) ExecuteStream(ctx ExecutionContext) iter.Seq2[*session.Event, error] {
    return func(yield func(*session.Event, error) bool) {
        // Implementation here
    }
}
```

**Implementation Steps**:
1. [ ] Create invocation ID (use google/uuid)
2. [ ] Yield "start" event with agent info
3. [ ] Start agent process (existing code)
4. [ ] Yield "progress" events as output arrives
5. [ ] Handle process completion
6. [ ] Yield "complete" event with exit code
7. [ ] Handle errors and emit "error" events
8. [ ] Ensure context cancellation works

**Event Generation**:
- [ ] Start event: When execution begins
- [ ] Progress events: As output is captured
- [ ] Complete event: When process exits successfully
- [ ] Error event: If execution fails

**Code Structure**:
```go
func (r *AgentRunner) ExecuteStream(ctx ExecutionContext) iter.Seq2[*session.Event, error] {
    return func(yield func(*session.Event, error) bool) {
        invocationID := uuid.NewString()
        startTime := time.Now()
        
        // 1. Validate context
        if ctx.Agent == nil {
            if !yield(nil, fmt.Errorf("agent is nil")) {
                return
            }
        }
        
        // 2. Yield start event
        startEvent := &session.Event{
            ID:           uuid.NewString(),
            Timestamp:    startTime,
            InvocationID: invocationID,
            Author:       ctx.Agent.Name,
            Type:         session.EventTypeStart,
            Content:      fmt.Sprintf("Starting agent: %s", ctx.Agent.Name),
        }
        if !yield(startEvent, nil) {
            return
        }
        
        // 3. Execute process (your existing code here)
        result, err := r.Execute(ctx)
        
        // 4. Yield output as progress events
        if result.Output != "" {
            progressEvent := &session.Event{
                ID:           uuid.NewString(),
                Timestamp:    time.Now(),
                InvocationID: invocationID,
                Author:       ctx.Agent.Name,
                Type:         session.EventTypeProgress,
                Content:      result.Output,
                ExecutionTime: result.Duration,
            }
            if !yield(progressEvent, nil) {
                return
            }
        }
        
        // 5. Yield completion event
        completeEvent := &session.Event{
            ID:           uuid.NewString(),
            Timestamp:    time.Now(),
            InvocationID: invocationID,
            Author:       ctx.Agent.Name,
            Type:         session.EventTypeComplete,
            Content:      fmt.Sprintf("Agent completed with exit code: %d", result.ExitCode),
            Data:         result,
            ExecutionTime: time.Since(startTime),
            Success:       result.Success,
            Error:         result.Error,
        }
        if !yield(completeEvent, nil) {
            return
        }
    }
}
```

**Keep Existing Method**:
- [ ] Maintain Execute() method for backward compatibility
- [ ] Mark as Deprecated in comment
- [ ] Redirect to ExecuteStream internally (optional)

**Tests** (`pkg/agents/execution_test.go`):
- [ ] Test event sequence (start → progress → complete)
- [ ] Test error handling (error events)
- [ ] Test context cancellation
- [ ] Test with nil agent (error event)
- [ ] Test with timeout
- [ ] Verify event IDs are unique
- [ ] Verify timestamps are sequential

**Acceptance Criteria**:
- ExecuteStream method exists
- Events yielded in correct order
- Error handling works correctly
- Context cancellation respected
- Backward compatible (old Execute still works)
- All tests passing

---

### Task 1.4: Create Test Infrastructure (2 hours)

**Files**: 
- `internal/session/session_test.go` (NEW)
- `pkg/agents/execution_test.go` (UPDATE)  
**Status**: Not Started  
**Effort**: 2 hours  

**Test Files to Create/Update**:
- [ ] `internal/session/session_test.go`
  - Test Event creation
  - Test State operations
  - Test State thread safety
  - Test Events iteration
  
- [ ] `pkg/agents/execution_test.go`
  - Test ExecutionContext with new fields
  - Test ValidateExecutionContext
  - Test ExecuteStream event generation
  - Test backward compatibility with Execute
  - Test error handling in ExecuteStream
  - Test context cancellation

**Test Coverage Goals**:
- [ ] 80%+ code coverage for new code
- [ ] Critical paths tested
- [ ] Error conditions tested
- [ ] Concurrency tested (state operations)

**Mock/Fixture Setup**:
- [ ] Mock Session for testing
- [ ] Mock Memory interface
- [ ] Mock Artifacts interface
- [ ] Sample agent for testing
- [ ] Temporary directory for execution

**Acceptance Criteria**:
- All tests passing
- 80%+ coverage for new code
- No race conditions detected
- Test execution time < 5 seconds

---

### Task 1.5: Documentation & Integration (1 hour)

**Files**:
- Code comments (inline)
- `docs/PHASE2_GUIDE.md` (update with Week 1 progress)
- Update `draft_session.md` with implementation notes

**Documentation Tasks**:
- [ ] Godoc comments for Session interface
- [ ] Godoc comments for Event type
- [ ] Godoc comments for ExecuteStream method
- [ ] Update ExecutionContext field comments
- [ ] Add example usage in code comments
- [ ] Update Phase 2 guide with session/event model
- [ ] Update draft session with implementation notes

**Integration Tasks**:
- [ ] Ensure code compiles: `go build ./...`
- [ ] Run linter: `golangci-lint run`
- [ ] Run formatter: `gofmt -w pkg/agents internal/session`
- [ ] Run vet: `go vet ./...`
- [ ] Run tests: `go test ./...`
- [ ] Check coverage: `go test -cover ./...`

**Acceptance Criteria**:
- [ ] make check passes
- [ ] All Godoc comments present
- [ ] Documentation updated
- [ ] Integration tests ready for next sprint

---

## Daily Standup Template

### Day 1 (Mon): Session Infrastructure
- [ ] Tasks 1.1 complete
- [ ] Session types compiling
- [ ] Initial tests passing
- [ ] Code review scheduled

### Day 2 (Tue): ExecutionContext Expansion
- [ ] Task 1.2 complete
- [ ] ExecutionContext expanded
- [ ] New fields zero-valued
- [ ] Backward compat verified

### Day 3 (Wed): ExecuteStream Implementation
- [ ] Task 1.3 complete
- [ ] ExecuteStream working
- [ ] Events generating correctly
- [ ] Tests comprehensive

### Day 4 (Thu-Fri): Testing & Polish
- [ ] All tests passing
- [ ] Coverage >= 80%
- [ ] make check passing 100%
- [ ] Code review complete
- [ ] Ready for Spec 0004

---

## Code Quality Checklist

### Before Committing
- [ ] `go fmt ./...` - Format code
- [ ] `go vet ./...` - Static analysis
- [ ] `golangci-lint run` - Linter (if available)
- [ ] `go test -race ./...` - Race condition detection
- [ ] `go test -cover ./...` - Coverage check (>= 80%)

### Before Code Review
- [ ] Godoc comments complete
- [ ] No exported symbols without docs
- [ ] Error messages clear and helpful
- [ ] Examples in comments
- [ ] No debug prints left

### Before Merge
- [ ] All tests passing
- [ ] Coverage >= 80%
- [ ] Code review approved
- [ ] Integration tests passing
- [ ] make check passes

---

## Risk Mitigation

### Risk: Breaking Changes
- **Mitigation**: All new fields optional (zero-valued), Execute() method unchanged
- **Verification**: Run existing code paths, verify no errors

### Risk: Event Ordering Issues
- **Mitigation**: Yield events in strict order, unique IDs for each
- **Verification**: Test event sequence, verify timestamps sequential

### Risk: Race Conditions in State
- **Mitigation**: Use RWMutex for state protection
- **Verification**: `go test -race ./...` must pass

### Risk: Session Integration Breaks Execution
- **Mitigation**: Session/Memory/Artifacts are optional (nil ok)
- **Verification**: Test with nil session, test with populated session

---

## Success Metrics

### Code Metrics
- ✅ make check passing (fmt, vet, lint, test)
- ✅ 80%+ code coverage
- ✅ 0 race conditions
- ✅ 0 lint errors

### Functional Metrics
- ✅ ExecuteStream yields correct event sequence
- ✅ ExecutionContext accepts new fields
- ✅ Backward compatibility maintained
- ✅ All tests passing

### Quality Metrics
- ✅ Code review approved
- ✅ Documentation complete
- ✅ Examples working
- ✅ Integration tests ready

---

## Resources & References

### Code Examples
- See Spec 0001 for ExecutionContext design
- See Spec 0003 for Event-based execution pattern
- Google ADK patterns in research/adk-go/

### Tools
- `make check` - Run all quality gates
- `go test -v` - Verbose test output
- `go test -run TestName` - Run specific test
- `go test -cover` - Coverage summary

### Documentation
- [Spec 0001: ExecutionContext](./0001_execution_context_expansion.md)
- [Spec 0003: Event-Based Execution](./0003_event_based_execution_model.md)
- [ARCHITECTURE.md](../ARCHITECTURE.md)

---

## Notes

### Important Reminders
1. Keep ExecutionContext backward compatible (all new fields optional)
2. ExecuteStream should NOT replace Execute() (both exist)
3. Session/State are interfaces (can have multiple implementations)
4. Events are immutable once yielded
5. InvocationID tracks events to same execution

### Future Considerations
- Week 2: Agent-as-Tool wrapper uses ExecuteStream
- Week 3: Memory/Artifact integration (agent can write)
- Week 4: Session persistence (store events in DB)

---

## Completion Checklist

### By End of Sprint
- [ ] All 5 tasks completed
- [ ] All tests passing
- [ ] Coverage >= 80%
- [ ] Code review approved
- [ ] make check passing
- [ ] Documentation complete
- [ ] Ready for Spec 0004

### Sign-Off
- [ ] Tech Lead Review
- [ ] QA Review
- [ ] Documentation Review
- [ ] Ready to Merge

---

**Sprint Status**: READY TO START  
**Created**: November 15, 2025  
**Target Completion**: November 22, 2025  
**Effort**: 12 hours (3 days)  
**Next Sprint**: Week 2 - Spec 0004 (Agent-as-Tool)
