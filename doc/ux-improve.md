# Event Display UX/UI Improvements - Brainstorm Report

**Date**: November 11, 2025  
**Focus**: Practical, pragmatic improvements to event visibility and user understanding  
**Philosophy**: Don't over-engineer. Leverage existing systems. Maximum impact with minimal code.

---

## Executive Summary

The code agent's event display currently works well but could be significantly improved with **visual hierarchy**, **event streaming feedback**, and **progress context**. Most improvements can be implemented in the existing `display/` package without architectural changes.

Current pain points:

- Tool execution feels "silent" - user doesn't know what's happening
- Token metrics display is hidden in spinner (only visible briefly)
- Thinking/analysis events lack visual distinction from tool operations
- Long operations have no intermediate feedback on progress
- Event sequence is hard to follow when multiple tools run

---

## Priority 1: Event Timeline View (High Impact, Low Effort)

When an agent runs multiple operations, the user can't see the overall timeline or sequence. They see individual outputs but lose context of what happened overall.

**Solution**: Display a compact summary sidebar/bar at the bottom of the screen showing the sequence of events in real-time.

```text
Tool execution:  [read_file] → [grep_search] → [write_file] → [execute_command]
Execution:       ✓            ✓              ✓               ⧖

Status: Agent executing 4 tasks (current: write_file)
Elapsed: 2.34s | Estimated: 3.5s (tool takes ~0.7s avg)
```

**Why It Works**:

- **Minimal rendering cost**: Single line update, no layout recalculation
- **Provides context**: User sees what's been done and what's next
- **No intrusive**: Stays at bottom, doesn't interrupt tool output
- **Extensible**: Can add more data without breaking layout

**Implementation Approach**:

1. Add a new `EventTimeline` struct to track event sequence
2. Update `printEventEnhanced()` to append events to timeline
3. Create a `RenderTimeline()` method (single line with ANSI color)
4. Display at bottom before prompt (or in a persistent status line if TTY supports it)

**Estimated effort**: 2-3 hours

---

## Priority 2: Visual Event Types (High Impact, Very Low Effort)

All events look similar. User can't quickly distinguish between thinking, tool execution, results, errors, and success.

**Solution**: Use consistent visual prefixes for different event types:

```text
🧠 Thinking: Analyzing the code structure...
🔧 Executing: read_file(config.json)
📊 Result: Found 3 configuration sections
✓ Success: File written successfully
⚠️  Warning: Using deprecated API
❌ Error: File not found
📍 Progress: Processing item 3 of 10
```

**Why It Works**:

- **Immediate visual scanning**: Users know event type at a glance
- **Emoji universality**: Works across terminals, localizes naturally
- **Minimal code change**: Just update spinner and renderer
- **Already partially done**: Code uses similar patterns (✓, ✗ in existing code)

**Implementation Approach**:

1. Create `EventType` enum: `Thinking`, `Executing`, `Result`, `Success`, `Warning`, `Error`, `Progress`
2. Update `RenderToolExecution()`, `RenderToolResult()` to include type indicator
3. Add `renderer.EventTypeIcon(eventType)` helper
4. Update spinner messages to use indicators

**Estimated effort**: 1-2 hours

---

## Priority 3: Token Metrics Dashboard (Medium Impact, Low Effort)

Token metrics are recorded but only displayed in the spinner (briefly visible). Users interested in API cost/performance can't easily track cumulative usage.

**Solution**: Show token usage in the footer/status area, updated in real-time:

```text
Tokens: 2,341 prompt | 892 cached | 1,205 response | 427 thinking | 189 tool-use | Total: 5,054
Cost estimate: $0.015 (at Gemini 2.5 Flash rates)
```

**Why It Works**:

- **Transparency**: Users know what they're being charged for
- **Minimal overhead**: Single line, updates are fast
- **Useful for debugging**: Helps understand model behavior
- **Already collected**: Just needs display layer

**Implementation Approach**:

1. Update `SessionTokens` to expose current metrics (not just summary)
2. Modify spinner's `UpdateWithMetrics()` to also display in footer
3. Create `FormatTokenMetrics()` in tracking package
4. Add to status line render (after event timeline)

**Estimated effort**: 1-2 hours

---

## Priority 4: Multi-Tool Progress Indication (Medium Impact, Medium Effort)

When an agent runs 5+ operations in sequence, users see individual spinners and outputs but don't understand overall progress.

**Solution**: Show completion percentage and operation count:

```text
Progress: [████████░░] 80% | 4 of 5 operations complete | ETA: 2.1s
```

**Why It Works**:

- **Context awareness**: Users understand how close to completion
- **Realistic expectations**: ETA helps them know if something is stuck
- **Non-intrusive**: Single line, uses spinner infrastructure
- **Low overhead**: Just needs average timing per tool type

**Implementation Approach**:

1. Add `OperationCounter` to track: total expected, completed, failed
2. Modify `printEventEnhanced()` to increment counters on success
3. Update spinner to show progress bar and ETA
4. Base ETA on historical average for each tool type (stored in `SessionTokens`)

**Estimated effort**: 2-3 hours

---

## Priority 5: Thinking/Analysis Mode Visual (Medium Impact, Low Effort)

When agent is "thinking," the spinner message says "Agent is thinking" but there's no visual distinction from other operations. Users might wonder if something is hanging.

**Solution**: Use different spinner animation and styling when agent is in thinking mode:

```text
🧠 Agent is thinking deeply... (pause, analyze code structure)
⠋ Agent is thinking deeply...
⠙ Agent is thinking deeply...
[slightly slower animation, different emoji/color]
```

versus:

```text
🔧 Reading config.json
⠋ Reading config.json
⠙ Reading config.json
[faster animation, different color]
```

**Why It Works**:

- **Clear state indication**: Thinking versus doing
- **Visually distinct**: Different colors/animations prevent confusion
- **Reassuring**: Users know system is working (not hung)
- **Already exists**: Just needs configuration

**Implementation Approach**:

1. Detect "thinking" text patterns in `printEventEnhanced()`
2. Add `SpinnerMode` enum: `Tool`, `Thinking`, `Analysis`
3. Create mode-specific styles with different animations and colors
4. Switch spinner mode based on content detection

**Estimated effort**: 1-2 hours

---

## Priority 6: Tool Output Structuring (Medium Impact, Medium Effort)

Tool results dump raw JSON or text. User can't quickly understand what happened.

**Solution**: Parse and summarize key tool result types:

```text
✓ Read file: config.json (342 bytes)
  └─ Content preview: [first 3 lines or summary]

✓ Search complete: 12 matches found in 4 files
  └─ files: config.json (3), main.go (5), test.go (4)

✓ Command executed: npm test
  └─ Exit code: 0
  └─ Timestamp: 2.34s elapsed
```

**Why It Works**:

- **Faster comprehension**: Structured summary versus raw output
- **Signal-to-noise**: Hides verbose output details
- **Consistent format**: Users develop mental model
- **Partial implementation exists**: `ToolResultParser` already exists

**Implementation Approach**:

1. Enhance existing `ToolResultParser` with more tool types
2. Add summary formatting for each tool result type
3. Create collapsible sections (expandable for full output)
4. Store detailed output but show summary by default

**Estimated effort**: 2-4 hours

---

## Priority 7: Error Highlighting & Recovery (Medium Impact, Low Effort)

Errors blend into normal output. User might miss them. No clear "what to do next" guidance.

**Solution**: Make errors unmissable and actionable:

```text
❌ ERROR: File not found
   Path: /Users/project/missing.txt
   
   💡 Suggestions:
   • Check file path spelling
   • Run '/tools list' to see available operations
   • Try: list_directory /Users/project
```

**Why It Works**:

- **Visual prominence**: Red + emoji + spacing
- **Actionable**: Suggests next steps
- **Context-aware**: Different errors show relevant suggestions
- **Already partially done**: Error rendering exists, just needs enhancement

**Implementation Approach**:

1. Create `ErrorSuggestions` mapping in renderer
2. Enhance `RenderError()` to include suggestions
3. Add error type detection
4. Format with spacing and colors for prominence

**Estimated effort**: 1-2 hours

---

## Priority 8: Command Execution Streaming (Medium Impact, High Effort)

When running commands (especially long-running ones), users see the spinner but not the actual command output. They don't know if command is working.

**Solution**: Stream command output in real-time:

```text
🔧 Running: npm test
┌─────────────────────────────────────────────────┐
│ PASS  src/utils.test.js                         │
│ PASS  src/api.test.js                           │
│ Test Suites: 2 passed, 2 total                  │
│ Tests:       15 passed, 15 total                │
└─────────────────────────────────────────────────┘
✓ Command completed (exit code: 0)
```

**Why It Works**:

- **User sees progress**: Confirms command is working
- **Builds confidence**: Especially for long-running operations
- **Error diagnosis**: Output shows actual errors, not hidden
- **More engaging**: Interactive feel

**Implementation Approach**:

1. Modify `ExecuteCommandTool` to stream output
2. Update `printEventEnhanced()` to capture streamed output
3. Add output buffering with max-length limits (prevent huge outputs)
4. Format with subtle border to distinguish from other output

**Estimated effort**: 3-5 hours (requires tool modification)

---

## Priority 9: Session Overview Header (Low Impact, Very Low Effort)

When resuming a session, users don't see what was done in previous interactions.

**Solution**: Display session context at start:

```text
📖 Resumed session: "feature-development" 
   Events: 23 | Tools: read_file(5), write_file(3), execute_command(8)
   Tokens used: 12,432 | Time: 5.2 min
```

**Why It Works**:

- **Contextual awareness**: Know what you're resuming
- **Progress visibility**: See cumulative work
- **Already available**: Data already in persistence layer

**Implementation Approach**:

1. Extract session summary from `SessionManager`
2. Create `FormatSessionHeader()` display function
3. Call during session resume

**Estimated effort**: 0.5-1 hour

---

## Priority 10: Collapsible Tool Output (Low Impact, High Effort)

Large tool outputs (file contents, command output) take up screen real estate.

**Solution**: Show previews, allow expansion:

```text
▶ Read file: src/main.go (342 lines)
  [Show first 3 lines and "... 336 more lines"]
  
Expand with: [SPACE] or [Enter]

▼ Read file: src/main.go (342 lines)  [EXPANDED]
  package main
  
  import (
    ...full file content...
  )
```

**Why It Works**:

- **Clean display**: Don't overwhelm user
- **On-demand detail**: Expand when needed
- **Natural scrolling**: User can scroll to content

**Implementation Approach**:

1. Add collapsible state to tool result renderer
2. Implement keyboard handler for expand/collapse
3. Use pager infrastructure (already exists)

**Estimated effort**: 3-4 hours

---

## Priority 11: Status Indicators in Prompt (Very Low Impact, Very Low Effort)

No indication of system state in the prompt. User doesn't know if session is active, if there are unsaved events, etc.

**Solution**: Modify prompt based on context:

```text
❯ normal input

[after tool execution]
✓ ❯ ready for input

[multiple tools ran]
⚡ ❯ enhanced, ready

[long operation]
⏱️  ❯ still processing...
```

**Why It Works**:

- **Minimal change**: Just emoji in prompt
- **Status at a glance**: Users know system state
- **Psychological benefit**: Feels responsive

**Implementation Approach**:

1. Track system state (last operation, timing, etc.)
2. Modify prompt generation to include indicator
3. Update in readline config

**Estimated effort**: 0.5-1 hour

---

## Recommended Implementation Order

### Phase 1: Quick Wins (4-6 hours)

1. Visual event types (emojis) - **1-2 hrs**
2. Thinking mode distinct display - **1-2 hrs**
3. Status indicators in prompt - **0.5-1 hr**
4. Session overview header - **0.5-1 hr**

**Result**: Significantly improved visual clarity with minimal code.

### Phase 2: Context & Progress (6-8 hours)

1. Event timeline view - **2-3 hrs**
2. Token metrics dashboard - **1-2 hrs**
3. Multi-tool progress indication - **2-3 hrs**
4. Error highlighting & recovery - **1-2 hrs**

**Result**: Users understand overall progress and can track costs.

### Phase 3: Enhanced Experience (6-10 hours)

1. Tool output structuring (enhance existing parser) - **2-4 hrs**
2. Command execution streaming - **3-5 hrs**
3. Collapsible tool output - **3-4 hrs**

**Result**: Professional, engaging CLI experience.

---

## Implementation Guidelines

### Keep It Simple

- Use existing color/style infrastructure
- Build on current spinner and renderer
- Don't create new major UI paradigms

### Performance First

- All updates should complete in <100ms
- Don't add blocking operations
- Use async rendering where possible

### Respect Terminal

- Check `IsTTY()` before fancy rendering
- Gracefully degrade to plain text
- Respect user's terminal settings

### Backward Compatible

- New features should be optional
- Existing workflows unchanged
- No breaking changes to APIs

---

## Quick Reference: Files to Modify

| Priority | Feature | Files to Modify | Complexity |
|----------|---------|-----------------|------------|
| 1 | Event types icons | `display/renderer.go`, `events.go` | Low |
| 2 | Thinking mode styling | `display/spinner.go`, `events.go` | Low |
| 3 | Prompt indicators | `main.go`, `display/renderer.go` | Very Low |
| 4 | Session header | `display/banner.go` | Very Low |
| 5 | Event timeline | `display/renderer.go`, `events.go` | Medium |
| 6 | Token dashboard | `tracking/`, `display/renderer.go` | Low |
| 7 | Progress indication | `display/spinner.go`, `events.go` | Medium |
| 8 | Error enhancements | `display/renderer.go` | Low |
| 9 | Result structuring | `display/tool_result_parser.go` | Medium |
| 10 | Command streaming | `tools/terminal_tools.go`, `events.go` | High |
| 11 | Collapsible output | `display/pager.go`, `display/renderer.go` | High |

---

## Expected User Impact

### Before Improvements

```text
❯ analyze my project
⠼ Agent is thinking...

Agent is reading config.json

Agent is searching for imports

Agent is executing npm test

...test output...

Task completed
❯
```

User experience: Unclear what's happening, feels slow, no progress visibility.

### After Phase 1 Improvements

```text
❯ analyze my project
🧠 Agent is thinking deeply...

🔧 Reading config.json
✓ Result: Found 42 configurations

🔍 Searching for imports
✓ Result: 23 files with imports

⚡ Executing npm test
✓ Command completed (exit code: 0)

Timeline: [read_file] → [grep_search] → [execute_command]
Tokens: 2,341 prompt | 892 cached | Total: 5,054
✓ Task completed
❯
```

User experience: Clear progression, understands what's happening, sees results and costs.

---

## Technical Notes

### Token Metrics Already Available

The `tracking/SessionTokens` already collects comprehensive metrics. Just needs a proper display layer.

### Spinner Infrastructure Mature

The spinner system is well-designed and can easily support different modes, animations, and styles. Extend rather than replace.

### Renderer Facade Pattern

The `display.Renderer` already abstracts styling. Use existing methods and add new ones for consistency.

### Event Stream Already Handled

The `printEventEnhanced()` function processes all events. Extend it rather than create new handlers.

---

## Conclusion

The event display system has strong foundations. Most improvements are **display-layer changes**, not architectural rewrites. Focusing on Phases 1-2 would provide **dramatic UX improvements** with minimal code complexity.

The key insight: **Users don't need more data, they need better visualization of existing data.**

Recommended next step: **Start with Phase 1** (visual event types + thinking mode styling). Should take 4-6 hours and provide immediate, visible improvements.
