# Spec 0004: Agent-as-Tool Integration

**Status**: Ready for Implementation  
**Priority**: P2  
**Effort**: 3 hours  
**Dependencies**: Spec 0001, 0003  
**File**: `pkg/agents/tool_adapter.go`  

## Summary

Enable agents to be invoked as tools by other agents, creating a composable agent hierarchy.

## Changes

Create `ToolAdapter` that wraps Agent as Tool:

```go
type ToolAdapter struct {
    agent *Agent
    runner *AgentRunner
}

func (t *ToolAdapter) Name() string
func (t *ToolAdapter) Description() string
func (t *ToolAdapter) Execute(ctx context.Context, input map[string]interface{}) (interface{}, error)
func (t *ToolAdapter) IsLongRunning() bool
```

## Implementation Steps

1. Define `ToolAdapter` struct wrapping Agent + AgentRunner
2. Implement Tool interface methods
3. Add ExecutionContext wrapper from Tool context
4. Add FunctionCallID tracking for tool invocations
5. Add tests for agent-as-tool execution
6. Add tests for nested agent hierarchies

## Integration Points

- Agent executing calls agent -> Tool context -> ExecutionContext
- FunctionCallID tracks tool invocation chain
- Memory and Artifacts shared up hierarchy
- Events propagated to root session

## Testing

- Unit tests for ToolAdapter
- Nested agent execution tests
- FunctionCallID tracking tests
- Tool registry discovery tests

## Success Criteria

- [ ] ToolAdapter wraps Agent as Tool
- [ ] Tool interface fully implemented
- [ ] ExecutionContext created from Tool context
- [ ] FunctionCallID tracked correctly
- [ ] Nested agents execute successfully
- [ ] All tests pass

---

**Version**: 1.0  
**Updated**: November 15, 2025
