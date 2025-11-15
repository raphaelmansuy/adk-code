# Spec 0003: Event-Based Execution Model

**Status**: Ready for Implementation  
**Priority**: P1  
**Effort**: 3 hours  
**Dependencies**: Spec 0001  
**File**: `pkg/agents/execution.go`, `internal/session/events.go`  

## Summary

Implement real-time event streaming during agent execution using Go 1.22+ iterators (iter.Seq2).

## Changes

1. **Event struct** in `internal/session/events.go`:

```go
type Event struct {
    ID           string
    Timestamp    time.Time
    InvocationID string
    Type         string  // "started", "completed", "error", "output"
    Content      string
    Metadata     map[string]interface{}
}
```

2. **Session Events** - add event persistence to Session

```go
type Session interface {
    // ... existing methods ...
    Events() []Event  // Get event history
    AddEvent(Event) error  // Add event to history
}
```

3. **Execution returns Event iterator**

```go
func (r *AgentRunner) Execute(ctx ExecutionContext) iter.Seq2[*session.Event, error]
```

## Implementation Steps

1. Define Event struct with ID, Timestamp, InvocationID, Type, Content
2. Add Events() method to Session interface
3. Implement event recording in Execute()
4. Return iter.Seq2 instead of ExecutionResult
5. Emit events during execution: started, output chunks, completed, errors
6. Add tests for event ordering and completeness

## Backward Compatibility

⚠️ **Breaking change** - Execute() returns iter.Seq2 instead of ExecutionResult

Mitigation:

- Create adapter function: `ExecuteWithResult()` for backward compatibility
- Deprecate old Execute signature
- Migrate callers gradually

## Testing

- Unit tests for event emission order
- Integration tests with real sessions
- Tests for error event handling
- Iterator protocol compliance tests

## Success Criteria

- [ ] Event struct defined
- [ ] Session supports event history
- [ ] Execute() returns iter.Seq2[*Event, error]
- [ ] Events emitted in correct order
- [ ] Error handling tested
- [ ] Backward compatibility adapter works
- [ ] All tests pass

---

**Version**: 1.0  
**Updated**: November 15, 2025
