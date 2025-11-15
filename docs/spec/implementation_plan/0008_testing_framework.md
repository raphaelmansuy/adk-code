# Spec 0008: Testing Framework

**Status**: Ready for Implementation  
**Priority**: P1  
**Effort**: 2 hours  
**Files**: `pkg/testutil/`, `*_test.go`  

## Summary

Establish comprehensive testing infrastructure for all Phase 2 components.

## Mock Implementations

Create mocks for testing:

```go
// Mock Session
type MockSession struct {
    id       string
    userID   string
    state    *MockState
    events   []Event
}

// Mock Memory
type MockMemory struct {
    data map[string]SearchResult
}

// Mock Artifact Service
type MockArtifactService struct {
    artifacts map[string]*Artifact
}

// Mock Tool
type MockTool struct {
    name string
    fn   func(context.Context, map[string]interface{}) (interface{}, error)
}
```

## Test Fixtures

Provide reusable test data:

- Sample agents
- Sample sessions
- Sample events
- Sample memories
- Sample artifacts
- Sample tools

## Coverage Goals

- Unit test coverage: 80%+
- Integration test coverage: 60%+
- No untested public API functions

## Testing Patterns

1. **Unit tests** - test individual functions in isolation
2. **Integration tests** - test component interactions
3. **Mock tests** - use mocks for external dependencies
4. **Table-driven tests** - use for multiple scenarios

## Success Criteria

- [ ] Mock implementations provided
- [ ] Test fixtures available
- [ ] 80%+ code coverage achieved
- [ ] All tests green
- [ ] No flaky tests

---

**Version**: 1.0  
**Updated**: November 15, 2025
