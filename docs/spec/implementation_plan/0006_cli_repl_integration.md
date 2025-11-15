# Spec 0006: CLI/REPL Integration

**Status**: Ready for Implementation  
**Priority**: P2  
**Effort**: 4 hours  
**Dependencies**: Spec 0001-0005  
**File**: `internal/repl/repl.go`  

## Summary

Integrate ExecutionContext and session management into the CLI REPL for interactive agent development.

## Changes

1. **REPL Session Management**:
   - Create session on startup
   - Maintain session state across commands
   - Display session info in REPL prompt

2. **Context Injection**:
   - Inject ExecutionContext into all agent runs
   - Pass session, memory, artifacts to agents
   - Track invocation chain

3. **Event Display**:
   - Stream events in real-time to terminal
   - Show progress indicators
   - Display errors with context

4. **REPL Commands**:
   - `/session` - show current session
   - `/history` - show event history
   - `/state` - show session state
   - `/memory` - search memory
   - `/artifacts` - list artifacts

## Implementation Steps

1. Modify REPL to create Session on startup
2. Update `RunAgent()` to create ExecutionContext with session
3. Add event streaming to terminal display
4. Implement REPL commands for session/history/state
5. Add tests for REPL commands

## Testing

- REPL startup tests
- Session persistence tests
- Event streaming tests
- Command execution tests

## Success Criteria

- [ ] REPL creates session on startup
- [ ] ExecutionContext injected with session
- [ ] Events stream to terminal
- [ ] Session commands work
- [ ] All tests pass

---

**Version**: 1.0  
**Updated**: November 15, 2025
