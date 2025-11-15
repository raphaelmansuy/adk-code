# Spec 0001: ExecutionContext Expansion

**Status**: Ready for Implementation  
**Priority**: P1  
**Effort**: 2 hours  
**File**: `pkg/agents/execution.go`  

## Summary

Add 8 optional fields to `ExecutionContext` struct for Google ADK Go alignment.

## Changes

Add these fields to the ExecutionContext struct:

```go
// Session context
Session      *session.Session
Memory       memory.Memory
Artifacts    artifact.Service
State        session.State

// Execution tracking
User         string
InvocationID string
FunctionCallID string
EventActions *session.EventActions
```

## Steps

1. **Add imports** to `pkg/agents/execution.go`
2. **Add the 8 fields** to ExecutionContext struct
3. **Add helper constructors**:
   - `NewExecutionContextWithSession()` - recommended
   - `NewExecutionContextSimple()` - backward compatible
4. **Add validation**:
   - `ValidateExecutionContext()` - require Agent, Context
5. **Add tests** for new fields and backward compatibility

## Backward Compatibility

✅ **Fully backward compatible** - all new fields are optional with zero values

## Testing

- Unit tests for constructors
- Validation tests
- Backward compatibility tests
- Integration tests with real Session

## Success Criteria

- [ ] 8 new fields added
- [ ] 2 constructors implemented
- [ ] Validation function works
- [ ] All tests pass (100% new code coverage)
- [ ] Existing code still compiles
- [ ] No performance regression

## Dependencies

- `internal/session` - Session, State interfaces
- `pkg/memory` - Memory interface (Spec 0002)
- `pkg/artifact` - Artifact Service interface (Spec 0002)

## Blocks

- Spec 0003: Event-Based Execution
- Spec 0004: Agent-as-Tool

---

**Version**: 1.0  
**Updated**: November 15, 2025
