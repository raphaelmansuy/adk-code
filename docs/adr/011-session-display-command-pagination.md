# ADR-011: `/session` Command with Rich Terminal Display and Pagination

**Status:** Accepted (Revised)  
**Date:** 2025-11-16  
**Last Updated:** 2025-11-16 (Session Management Extensions)  
**Authors:** Raphaël MANSUY  
**Deciders:** adk-code Architecture Team  
**Scope:** REPL Command, Display System, Session Management  

---

## Executive Summary

This ADR specifies the implementation of a **new `/session` REPL command** that displays session event history with **rich terminal formatting, comprehensive event visualization, smart pagination, and subcommand support**. The command provides users with an intuitive, visually organized exploration interface for both current and past sessions. Additionally, comprehensive session management commands enable users to create, list, and delete sessions directly from the REPL without exiting.

**Core Capabilities:**
- ✅ **Multi-Mode Display**: View current session overview, specific sessions by ID, or full event details
- ✅ **Rich Terminal Formatting**: Styled headers, color-coded event types, emojis, formatted timestamps
- ✅ **Smart Pagination**: Automatic pagination for large content (>24 lines)
- ✅ **Event Breakdown**: User inputs, model responses, tool calls/results, compaction events
- ✅ **Token Accounting**: Per-event and cumulative token metrics
- ✅ **Message Preview & Expansion**: Truncated previews with ability to read full content by event ID
- ✅ **Session Management**: Create, list, delete sessions without leaving REPL
- ✅ **Safety Features**: Confirmation prompts, prevent deletion of current session

**REPL Command Hierarchy:**

**Session Display:**
- `/session` - Display current session overview with event timeline
- `/session <session-id>` - Display specific session from database
- `/session event <event-id>` - Display full event content (no truncation)
- `/session help` - Show command documentation

**Session Management:**
- `/list-sessions` - List all available sessions
- `/new-session <name>` - Create a new session
- `/delete-session <name>` - Delete a session (with confirmation)

**Integration Points:**
- Uses `google.golang.org/adk/session` interfaces (Session, Event, Events)
- Integrates with existing `display.Paginator` component (`internal/display/components/paginator.go`)
- Extends REPL command handlers (`internal/cli/commands/repl.go`)
- Aligns with session compaction metadata (`internal/session/compaction/`)
- Shares session manager with CLI commands (`internal/cli/commands/session.go`)

---

## Problem Statement

### Current State

The adk-code agent currently supports session management CLI commands:

| Command | Purpose | REPL Equivalent | Output |
|---------|---------|-----------------|--------|
| `./code-agent list-sessions` | List all sessions | `/list-sessions` | Tabular format (session name + event count) |
| `./code-agent new-session <name>` | Create session | `/new-session <name>` | Confirmation message |
| `./code-agent delete-session <name>` | Delete session | `/delete-session <name>` | Confirmation message |
| `./code-agent --session <name>` | Resume session | N/A (exit & restart) | Resumes, no preview of history |
| N/A | View session history | `/session` | Rich formatted display |

**Missing Capability:**
Users cannot **interactively manage sessions** from within the REPL. To manage sessions or view history, users must:
1. Exit the REPL (for management)
2. Manually query the database or write a script to inspect events

3. Parse raw JSON/event data mentally

### Impact

| Aspect | Issue |
|--------|-------|
| **Developer Experience** | No visibility into what happened in a previous session without leaving the agent |
| **Debugging** | Cannot trace conversation flow, tool calls, or compaction events interactively |
| **Session Analysis** | Token usage metrics exist but aren't connected to specific events |
| **Discoverability** | New users unaware of how many events are in a session or conversation structure |

### Success Criteria

✅ **Display detailed session event history** with clear hierarchical structure  
✅ **Automatic pagination** for sessions with >24 lines of content  
✅ **Color-coded event types** (user input, model response, tool call, tool result, compaction)  
✅ **Token accounting** per event (if available from LLM response metadata)  
✅ **Navigation UX** matching existing `/help`, `/tools`, `/models` commands  
✅ **Zero latency** - display from already-loaded session (no new DB queries for data already in memory)  
✅ **Support current session preview** without requiring session resumption  

---

## Decision

### Core Decision

**Implement a `/session` REPL command with subcommand support:**

1. **View Current Session (default)**
   - Command: `/session`
   - Shows current session overview with paginated event timeline
   - Event metadata: timestamp, author, type, content preview, token counts
   - Compaction events show compression statistics

2. **View Specific Session**
   - Command: `/session <session-id>`
   - Retrieves and displays any session from database by ID
   - Same formatted display as current session
   - Useful for reviewing past conversations

3. **View Full Event Content**
   - Command: `/session event <event-id>`
   - Displays complete event without truncation
   - Shows full message text, tool parameters, results
   - Useful for inspecting message details without leaving REPL
   - Includes token counts and metadata

4. **Command Help**
   - Command: `/session help`
   - Displays usage examples and subcommand documentation

### Design Rationale

**Singular `/session` vs Plural `/sessions`:**
- More intuitive (similar to `/help`, `/tools`)
- Primary use case is viewing *one* session at a time
- Subcommands clarify different modes

**Subcommand Pattern:**
- Follows established REPL patterns (`/mcp list`, `/mcp status`)
- Extensible for future features (filters, search, export)
- Clear command hierarchy and help system

**Three-Layer Access Pattern:**
1. Overview: Quick view of event structure (default display)
2. Session-level: Switch between sessions by ID
3. Event-level: Detailed inspection of specific events

---

## Session Management Extensions

### Additional REPL Commands

In addition to the core `/session` command, the following session management commands have been implemented directly in the REPL:

### List All Sessions

- Command: `/list-sessions`
- Shows all available sessions for the current user
- Displays: session name, event count, last update time
- Useful for discovering past conversations without exiting REPL

### Create New Session

- Command: `/new-session <session-name>`
- Creates a new empty session
- Provides CLI hint for resuming the session later
- Validates session name is provided

### Delete Session

- Command: `/delete-session <session-name>`
- Deletes a session from the database
- Includes safety features:
  - Confirms session exists before deletion
  - Prevents accidental deletion of current session
  - Requires user confirmation ("yes") before proceeding
  - Shows informative error messages

### Design Rationale for Extensions

**Why These Commands in REPL:**

- Users need session management without exiting REPL
- Improves workflow continuity (create/explore/switch sessions fluidly)
- Provides visual feedback with formatting/emojis (✨, 📋, 🗑️)

**Safety First Approach:**

- Cannot delete current session (prevents data loss)
- Explicit confirmation required for deletion (typo-resistant)
- Pre-deletion verification (session exists)

**Command Hierarchy:**

- CLI commands: Used at shell startup/bootstrap (`./code-agent new-session <name>`)
- REPL commands: Used during interactive session (`/new-session <name>`)
- Both have identical functionality, different UX contexts

---

## Implementation Details

### Architecture Overview

```
REPL Input: /session [subcommand] [args]
         ↓
    HandleBuiltinCommand() [repl.go]
         ↓
    handleSessionCommand() - dispatcher
         ├─→ handleSessionOverview() - show current session
         ├─→ handleSessionByID() - load specific session
         ├─→ handleEventDetail() - show full event content
         └─→ handleSessionHelp() - show usage
         ↓
    buildSessionDisplayLines() [repl_builders.go]
    buildEventDisplayLines() [repl_builders.go]
         ↓
    Display.Paginator.DisplayPaged()
    (internal/display/components/paginator.go)
         ↓
    Terminal Output
```

### 1. REPL Command Handler Integration

**File:** `adk-code/internal/cli/commands/repl.go`

Add case to `HandleBuiltinCommand()`:

```go
case "/session":
    handleSessionCommand(ctx, renderer, appConfig)
    return true
    
// For subcommands starting with /session
if strings.HasPrefix(input, "/session ") {
    handleSessionCommand(ctx, renderer, appConfig)
    return true
}
```

Main dispatcher function:

```go
// handleSessionCommand handles /session with subcommands
func handleSessionCommand(ctx context.Context, input string, renderer *display.Renderer, appConfig interface{}) {
    cfg, ok := appConfig.(*config.Config)
    if !ok {
        fmt.Println(renderer.Red("Error: Configuration not available"))
        return
    }

    parts := strings.Fields(input)
    
    // Default to overview if no subcommand
    if len(parts) == 1 {
        handleSessionOverview(ctx, renderer, cfg)
        return
    }

    subcommand := parts[1]
    
    switch subcommand {
    case "help":
        handleSessionHelp(renderer)
    case "event":
        if len(parts) < 3 {
            fmt.Println(renderer.Yellow("⚠ Usage: /session event <event-id>"))
            return
        }
        eventID := parts[2]
        handleEventDetail(ctx, renderer, cfg, eventID)
    default:
        // Treat as session ID
        sessionID := subcommand
        handleSessionByID(ctx, renderer, cfg, sessionID)
    }
}
```

### 2. Handler Functions

**`handleSessionOverview()`** - Display current session with timeline
**`handleSessionByID(sessionID)`** - Load and display any session by ID
**`handleEventDetail(eventID)`** - Show full event content without truncation
**`handleSessionHelp()`** - Display command usage

### 3. Display Builder Functions

**File:** `adk-code/internal/cli/commands/repl_builders.go`

- `buildSessionDisplayLines()` - Session overview with event timeline
- `buildEventDisplayLines()` - Detailed view of single event
- Helper functions for formatting and truncation
    lines = append(lines, "")

    // === EVENT SUMMARY ===
    events := sess.Events()
    totalEvents := events.Len()
    
    eventCounts := countEventsByType(events)
    lines = append(lines, renderer.Bold(fmt.Sprintf("📊 Events: %d total", totalEvents)))
    
    if eventCounts.userMessages > 0 {
        lines = append(lines, fmt.Sprintf("  %s User inputs:      %d",
            renderer.Dim("•"), eventCounts.userMessages))
    }
    if eventCounts.modelResponses > 0 {
        lines = append(lines, fmt.Sprintf("  %s Model responses:  %d",
            renderer.Dim("•"), eventCounts.modelResponses))
    }
    if eventCounts.toolCalls > 0 {
        lines = append(lines, fmt.Sprintf("  %s Tool calls:       %d",
            renderer.Dim("•"), eventCounts.toolCalls))
    }
    if eventCounts.toolResults > 0 {
        lines = append(lines, fmt.Sprintf("  %s Tool results:     %d",
            renderer.Dim("•"), eventCounts.toolResults))
    }
    if eventCounts.compactions > 0 {
        lines = append(lines, fmt.Sprintf("  %s Compactions:      %d",
            renderer.Green("★"), eventCounts.compactions))
    }
    lines = append(lines, "")

    // === CUMULATIVE TOKENS ===
    totalTokens := countTotalTokens(events)
    if totalTokens > 0 {
        lines = append(lines, renderer.Bold(fmt.Sprintf("🎯 Token Usage: %d total", totalTokens)))
        lines = append(lines, "")
    }

    // === EVENT TIMELINE ===
    if totalEvents == 0 {
        lines = append(lines, renderer.Yellow("⚠ No events in this session yet"))
        return lines
    }

    lines = append(lines, renderer.Bold("📜 Event Timeline:"))
    lines = append(lines, "")

    // Iterate through events
    for i := 0; i < events.Len(); i++ {
        event := events.At(i)
        if event == nil {
            continue
        }

        // Event number and timestamp
        ts := event.Timestamp.Format("15:04:05")
        eventNum := fmt.Sprintf("[%d/%d]", i+1, totalEvents)
        
        lines = append(lines, renderer.Dim(fmt.Sprintf("  %s %s", eventNum, ts)))

        // Determine event type and format accordingly
        if compactionMeta, err := compaction.GetCompactionMetadata(event); err == nil {
            // === COMPACTION EVENT ===
            lines = append(lines, renderer.Green(fmt.Sprintf("    ★ COMPACTION EVENT")))
            lines = append(lines, fmt.Sprintf("      Events compressed:    %d → summary",
                compactionMeta.EventCount))
            lines = append(lines, fmt.Sprintf("      Tokens saved:          %d → %d (%.1f%% compression)",
                compactionMeta.OriginalTokens,
                compactionMeta.CompactedTokens,
                compactionMeta.CompressionRatio*100,
            ))
            lines = append(lines, fmt.Sprintf("      Period:                %s to %s",
                compactionMeta.StartTimestamp.Format("15:04:05"),
                compactionMeta.EndTimestamp.Format("15:04:05"),
            ))
        } else if event.LLMResponse.Content != nil {
            // === REGULAR EVENT ===
            author := event.Author
            if author == "" {
                author = "system"
            }

            // Color-code by author
            var authorStr string
            switch author {
            case "user":
                authorStr = renderer.Blue(fmt.Sprintf("👤 USER"))
            case "model":
                authorStr = renderer.Green(fmt.Sprintf("🤖 MODEL"))
            case "system":
                authorStr = renderer.Yellow(fmt.Sprintf("⚙️  SYSTEM"))
            default:
                authorStr = renderer.Dim(fmt.Sprintf("❓ %s", author))
            }

            lines = append(lines, fmt.Sprintf("    %s (ID: %s)",
                authorStr,
                truncateID(event.ID, 8),
            ))

            // Display content preview
            if len(event.LLMResponse.Content.Parts) > 0 {
                for _, part := range event.LLMResponse.Content.Parts {
                    if part != nil && part.Text != "" {
                        preview := truncateText(part.Text, 60)
                        lines = append(lines, fmt.Sprintf("      %s", renderer.Dim(preview)))
                    }
                }
            }

            // Show token count if available
            if event.LLMResponse.UsageMetadata != nil {
                promptTokens := int(event.LLMResponse.UsageMetadata.PromptTokenCount)
                outputTokens := int(event.LLMResponse.UsageMetadata.CandidatesTokenCount)
                if promptTokens > 0 || outputTokens > 0 {
                    lines = append(lines, fmt.Sprintf("      %s",
                        renderer.Dim(fmt.Sprintf("Tokens: %d prompt + %d output",
                            promptTokens, outputTokens))))
                }
            }
        }

        // Add spacing between events
        if i < events.Len()-1 {
            lines = append(lines, "")
        }
    }

    // === FOOTER ===
    lines = append(lines, "")
    lines = append(lines, renderer.Dim("Press SPACE to continue, Q to quit"))
    lines = append(lines, "")

    return lines
}

// Helper functions

type eventTypeCounts struct {
    userMessages    int
    modelResponses  int
    toolCalls       int
    toolResults     int
    compactions     int
}

func countEventsByType(events session.Events) eventTypeCounts {
    counts := eventTypeCounts{}
    
    for iter := events.All(); iter != nil; {
        evt := iter
        // Determine type based on Author and Content
        // This is a simplified approach - actual implementation 
        // may need to inspect Actions field for tool calls/results
        if compaction.IsCompactionEvent(evt) {
            counts.compactions++
        } else if evt.Author == "user" {
            counts.userMessages++
        } else if evt.Author == "model" {
            counts.modelResponses++
        }
    }
    
    return counts
}

func countTotalTokens(events session.Events) int {
    total := 0
    for iter := events.All(); iter != nil; {
        evt := iter
        if evt.LLMResponse.UsageMetadata != nil {
            total += int(evt.LLMResponse.UsageMetadata.TotalTokenCount)
        }
    }
    return total
}

func truncateText(text string, maxLen int) string {
    if len(text) <= maxLen {
        return text
    }
    return text[:maxLen-3] + "..."
}

func truncateID(id string, maxLen int) string {
    if len(id) <= maxLen {
        return id
    }
    return id[:maxLen]
}

func formatTimeAgo(t time.Time) string {
    elapsed := time.Since(t)
    if elapsed < time.Minute {
        return "just now"
    } else if elapsed < time.Hour {
        minutes := int(elapsed.Minutes())
        return fmt.Sprintf("%dm ago", minutes)
    } else if elapsed < 24*time.Hour {
        hours := int(elapsed.Hours())
        return fmt.Sprintf("%dh ago", hours)
    }
    days := int(elapsed.Hours()) / 24
    return fmt.Sprintf("%dd ago", days)
}
```

### 3. Event Type Detection

Events in Google ADK session API have:
- `ID`: Unique event identifier
- `Author`: "user", "model", or "system"
- `Timestamp`: When event occurred
- `InvocationID`: Groups related events
- `LLMResponse.Content`: The actual content (text, tool calls, etc.)
- `LLMResponse.UsageMetadata`: Token counts

**Reference:** `research/adk-go/session/session.go` lines 95-120

### 4. Integration with Existing Display Components

The implementation **reuses existing infrastructure**:

| Component | Usage | Reference |
|-----------|-------|-----------|
| `display.Paginator` | Line-by-line pagination with navigation prompts | `internal/display/components/paginator.go` |
| `display.Renderer` | Color/styling functions (Bold, Cyan, Dim, etc.) | `internal/display/renderer/` |
| Command builder pattern | Follows `/help`, `/tools`, `/models` command structure | `internal/cli/commands/repl.go` |

### 5. UX Examples

#### Example 1: Small Session (Single Page)

```
════════════════════════════════════════════════════════════════
                    Session: user-session-001
════════════════════════════════════════════════════════════════

📋 Session Details:
  • App:      code_agent
  • User:     user1
  • ID:       user-session-001
  • Updated:  2025-11-16 14:32:15 (5m ago)

📊 Events: 12 total
  • User inputs:      3
  • Model responses:  3
  • Tool calls:       4
  • Tool results:     2

🎯 Token Usage: 2,847 total

📜 Event Timeline:

  [1/12] 14:27:42
    👤 USER (ID: evt_abc123)
      Create a README with project setup instructions
      Tokens: 145 prompt + 0 output

  [2/12] 14:27:58
    🤖 MODEL (ID: evt_def456)
      I'll create a comprehensive README for your project...
      Tokens: 524 prompt + 892 output

  [3/12] 14:28:01
    👤 USER (ID: evt_ghi789)
      Tool call: read_file
      Tokens: 0 prompt + 0 output

  ...

Press SPACE to continue, Q to quit
```

#### Example 2: Large Session with Compaction (Multi-Page)

```
[Page 1/3] Press SPACE to continue, Q to quit: 
```

After pressing Space:

```
  [47/87] 14:52:15
    ★ COMPACTION EVENT
      Events compressed:    46 → summary
      Tokens saved:          18,342 → 2,156 (88.3% compression)
      Period:                14:27:42 to 14:51:30

  [48/87] 14:52:18
    🤖 MODEL (ID: evt_xyz999)
      Based on our previous work, here's what I've learned...
      Tokens: 2,340 prompt + 1,200 output

[Page 2/3] Press SPACE to continue, Q to quit:
```

---

## Code References and Verification

### Google ADK Session API

**File:** `research/adk-go/session/session.go`

Key interfaces we depend on:

```go
type Session interface {
    ID() string                          // Get session ID
    AppName() string                     // Get app name
    UserID() string                      // Get user ID
    State() State                        // Access session state
    Events() Events                      // Access event sequence
    LastUpdateTime() time.Time          // Get last update
}

type Events interface {
    All() iter.Seq[*Event]             // Iterate all events
    Len() int                           // Total event count
    At(i int) *Event                   // Get event by index
}

type Event struct {
    model.LLMResponse                  // Contains Content and UsageMetadata
    ID            string                // Unique event ID
    Timestamp     time.Time             // Event timestamp
    InvocationID  string                // Groups related calls
}
```

✅ **Verified**: All interfaces are stable in research/adk-go  
✅ **Current**: Session manager in adk-code/internal/session/manager.go implements these  

### Existing Pagination Component

**File:** `adk-code/internal/display/components/paginator.go`

```go
type Paginator struct {
    renderer core.StyleRenderer
}

func (p *Paginator) DisplayPaged(lines []string) bool {
    // Handles pagination, terminal size detection, user input
}
```

✅ **Verified**: Used by `/help`, `/tools`, `/models`, `/providers` commands  
✅ **Tested**: `paginator_test.go` exists with coverage  

### Session Compaction Metadata

**File:** `adk-code/internal/session/compaction/compaction.go`

```go
type CompactionMetadata struct {
    EventCount           int
    OriginalTokens       int
    CompactedTokens      int
    CompressionRatio     float64
    StartTimestamp       time.Time
    EndTimestamp         time.Time
}

func IsCompactionEvent(evt *session.Event) bool { ... }
func GetCompactionMetadata(evt *session.Event) (*CompactionMetadata, error) { ... }
```

✅ **Verified**: Compaction event detection available  

### REPL Command Handler Pattern

**File:** `adk-code/internal/cli/commands/repl.go`

All REPL commands follow this pattern:

```go
case "/command":
    handleCommandCommand(renderer, ...)
    return true

func handleCommandCommand(renderer *display.Renderer, ...) {
    lines := buildCommandLines(renderer, ...)
    paginator := display.NewPaginator(renderer)
    paginator.DisplayPaged(lines)
}
```

✅ **Verified**: `/help`, `/tools`, `/models`, `/providers`, `/tokens` follow this pattern  
✅ **References**: `internal/cli/commands/repl.go` lines 24-63 + builder functions  

---

## Consequences

### Positive Outcomes

1. **Enhanced User Experience**
   - Users can explore session history without leaving REPL
   - Clear visibility into what happened in a session
   - Easier debugging and understanding of agent decisions

2. **Better Session Management**
   - Makes session compaction events visible and understandable
   - Token metrics connected to actual events
   - Helps users understand token usage patterns

3. **Reduced Complexity**
   - No new database queries (uses existing in-memory session)
   - Leverages existing pagination and display infrastructure
   - Minimal code changes - single command handler + builder function

4. **Architectural Alignment**
   - Uses Google ADK session interface correctly
   - Follows REPL command pattern consistently
   - Integrates with compaction system seamlessly

### Trade-offs

| Trade-off | Resolution |
|-----------|-----------|
| **Memory**: Large sessions with 100+ events = 100+ lines to format | Paginator handles this; respects terminal height |
| **Performance**: Iterating events O(n) | Session already in memory; negligible for typical 50-100 event sessions |
| **Scope**: Only shows current session (not arbitrary session by name) | Future enhancement; covers primary use case first |

### Migration/Deprecation

- No breaking changes to existing session commands
- `/sessions` is purely additive
- Existing `list-sessions`, `delete-session` CLI commands remain unchanged

---

## Alternatives Considered

### Alt 1: Add `--verbose` flag to `list-sessions`

**Rejected** because:
- Not interactive or discoverable from REPL
- Requires exit from agent context
- Doesn't fit REPL command pattern

### Alt 2: New CLI command `code-agent view-session <name>`

**Rejected** because:
- Doesn't solve primary use case (understanding *current* session)
- Users must leave REPL to investigate
- Doesn't integrate with REPL workflow

### Alt 3: Rich table display using `lipgloss` borders

**Rejected** because:
- Overkill for linear event stream
- Harder to parse for copy-paste
- Adds dependency on additional styling

### Alt 4: Export session as JSON/CSV file

**Rejected** because:
- Solves external analysis use case, not interactive exploration
- Doesn't answer "what happened in my session?" immediately
- Requires external tools to view

**Decision**: REPL command `/sessions` with paginated event timeline best meets the immediate need while supporting future enhancements.

---

## Implementation Roadmap

### Phase 1: Basic Implementation (Priority: High)

- [ ] Add `handleSessionsCommand()` to `repl.go`
- [ ] Create `buildSessionDisplayLines()` in `repl_builders.go`
- [ ] Implement event type detection and formatting
- [ ] Test pagination with 50+ event session
- [ ] Manual testing in REPL with /sessions command

### Phase 2: Enhanced Metrics (Priority: Medium)

- [ ] Add tool call/result detection and formatting
- [ ] Calculate and display invocation breakdown (show which turns made tool calls)
- [ ] Show token usage deltas between events
- [ ] Add optional `--format` parameter (json, csv future)

### Phase 3: Session Selection (Priority: Low)

- [ ] Extend to `/sessions <session-id>` to view any session
- [ ] Add session filtering/search capabilities
- [ ] Integrate with list-sessions for quick navigation

---

## Testing Strategy

### Unit Tests

```go
// Test helpers exist
func TestBuildSessionDisplayLines(t *testing.T)
func TestCountEventsByType(t *testing.T)
func TestTruncateText(t *testing.T)
func TestFormatTimeAgo(t *testing.T)
```

### Integration Tests

```bash
# Manual test with real session
echo "/sessions" | ./adk-code 2>&1 | head -50

# Verify pagination works for large sessions
# Verify formatting (colors, alignment)
# Verify event count accuracy
```

### Test Scenarios

1. **Empty session** → Should show "No events in this session yet"
2. **Single event** → Should display in one page
3. **50+ events** → Should paginate and show "Page X/Y"
4. **Session with compaction** → Should display compaction metadata correctly
5. **Mixed author types** → Should color-code user/model/system correctly

---

## Success Metrics

✅ **Functional:**
- `/session` command displays current session event history
- `/session <id>` retrieves and displays any session by ID
- `/session event <id>` shows full event content without truncation
- Pagination works for sessions with >24 events
- All event types properly formatted and visible

✅ **UX:**
- Users understand session flow at a glance
- Time values human-readable ("5m ago" not millisecond timestamps)
- Clear visual hierarchy (headers, event blocks, spacing)
- Event IDs visible for reference in `/session event` command
- Help text explains subcommand usage

✅ **Code Quality:**
- ~300-400 lines new code (handler dispatcher + builders)
- Zero new external dependencies
- Consistent with existing REPL command patterns
- All existing tests continue to pass

✅ **Performance:**
- Command executes in <100ms for typical sessions
- Session lookup by ID in <50ms (database query)
- No impact on REPL startup time

---

## Implementation Roadmap

### Phase 1: Core Implementation (Priority: High)

- [ ] Add `handleSessionCommand()` dispatcher to `repl.go`
- [ ] Implement `handleSessionOverview()` for current session
- [ ] Implement `handleSessionByID()` for loading sessions
- [ ] Implement `handleEventDetail()` for full event display
- [ ] Update `HandleBuiltinCommand()` to route `/session` commands
- [ ] Create `buildSessionDisplayLines()` in `repl_builders.go`
- [ ] Create `buildEventDisplayLines()` in `repl_builders.go`
- [ ] Update help message with `/session` documentation
- [ ] Test pagination with 50+ event session
- [ ] Test session lookup by ID from database

### Phase 2: Enhanced Formatting (Priority: Medium)

- [ ] Improve tool call/result formatting (show tool name, parameters)
- [ ] Display thinking output if present in events
- [ ] Better text wrapping (not just truncation at 60 chars)
- [ ] Show invocation IDs linking tool calls to results
- [ ] Syntax highlighting for code blocks

### Phase 3: Advanced Features (Priority: Low)

- [ ] Add filtering: `/session --filter model` or `--filter user`
- [ ] Add search: `/session search "keyword"`
- [ ] Add statistics: token usage trends, invocation breakdown
- [ ] Add export: `/session --format json` for external analysis
- [ ] Event range display: `/session events 5-15` for specific range

---

## Related Documents and References

**Architecture:**
- [adk-code ARCHITECTURE.md](../ARCHITECTURE.md) - System design
- [Session Compaction ADR-010](./010-native-sdk-session-compaction.md) - Session compaction design

**Code References:**
- [Paginator Component](../../adk-code/internal/display/components/paginator.go)
- [REPL Command Handlers](../../adk-code/internal/cli/commands/repl.go)
- [Session Manager](../../adk-code/internal/session/manager.go)
- [Google ADK Session API](../../research/adk-go/session/session.go)
- [Display Facades](../../adk-code/internal/display/facade.go)
- [Compaction Metadata](../../adk-code/internal/session/compaction/compaction.go)

**Similar Patterns in Reference Code:**
- [ADK Go Session Service](../../research/adk-go/session/service.go) - Session interface design
- [Existing REPL Commands](../../adk-code/internal/cli/commands/repl_builders.go) - Command builder pattern

---

## Appendix A: Display Mockup

```
════════════════════════════════════════════════════════════════
                    Session: session-abc123
════════════════════════════════════════════════════════════════

📋 Session Details:
  • App:      code_agent
  • User:     user1
  • ID:       session-abc123
  • Updated:  2025-11-16 15:45:20 (2m ago)

📊 Events: 24 total
  • User inputs:      6
  • Model responses:  6
  • Tool calls:       8
  • Tool results:     4

🎯 Token Usage: 5,234 total

📜 Event Timeline:

  [1/24] 15:30:12
    👤 USER (ID: evt_001)
      Create a Python script that reads CSV files
      Tokens: 156 prompt + 0 output

  [2/24] 15:30:28
    🤖 MODEL (ID: evt_002)
      I'll create a Python script that reads CSV files and...
      Tokens: 245 prompt + 1,842 output

  [3/24] 15:30:35
    👤 USER (ID: evt_003)
      Tool call: write_file (path: script.py)
      Tokens: 0 prompt + 0 output

  ...
  
  [23/24] 15:44:58
    ★ COMPACTION EVENT
      Events compressed:    20 → summary
      Tokens saved:         12,456 → 1,234 (90.1% compression)
      Period:               15:30:12 to 15:42:15

  [24/24] 15:45:20
    🤖 MODEL (ID: evt_024)
      Based on our earlier work, the script is now...
      Tokens: 1,340 prompt + 892 output

[Page 1/2] Press SPACE to continue, Q to quit:
```

---

**Document Version:** 1.0  
**Last Updated:** 2025-11-16  
**Status:** Ready for Review
