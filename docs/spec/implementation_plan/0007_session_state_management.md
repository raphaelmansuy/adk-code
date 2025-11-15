# Spec 0007: Session & State Management

**Status**: Ready for Implementation  
**Priority**: P1  
**Effort**: 2 hours  
**Dependencies**: Spec 0001  
**File**: `internal/session/session.go`  

## Summary

Implement Session and State interfaces for stateful agent execution with scoped state.

## Changes

Session interface:

```go
type Session interface {
    ID() string
    UserID() string
    State() State
    Events() []Event
    AddEvent(Event) error
}

type State interface {
    Get(key string) (interface{}, error)
    Set(key string, value interface{}) error
    All() map[string]interface{}
}
```

Scoped state paths:

- `app:/key` - application-level state
- `user:/key` - user-level state  
- `temp:/key` - temporary session state

## Implementation Steps

1. Define Session interface with ID, UserID, State, Events
2. Define State interface with Get, Set, All
3. Implement in-memory Session
4. Implement scoped State with path parsing
5. Add event persistence to Session
6. Add tests for state scoping and isolation

## Testing

- State CRUD operations
- State scoping tests
- Event recording tests
- Session isolation tests

## Success Criteria

- [ ] Session interface defined
- [ ] State interface with scoping implemented
- [ ] Event recording works
- [ ] State isolation verified
- [ ] All tests pass

---

**Version**: 1.0  
**Updated**: November 15, 2025
