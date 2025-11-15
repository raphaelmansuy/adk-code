# ADR-0006: Multi-Layer Context Management for Long-Running Agents

**Status**: Active  
**Date**: 2025-11-15  
**Revision**: 2  
**Owner**: adk-code Team  

---

## Decision

Implement **three integrated context management layers** to support 50+ turn agent workflows:

1. **Output Truncation** – Head+tail strategy (first 128 lines + last 128 lines, max 10 KiB) immediately after tool execution
2. **Token Tracking** – Real-time accounting per turn, metrics in REPL, enforce 95% hard limit
3. **Conversation Compaction** – Automatic LLM-powered summarization at 70% context window threshold

**Why**: Current implementation fails after ~20 turns due to unbounded context growth. Proven approach (Codex, Claude) enables 50+ turn workflows and matches Google ADK patterns.

---

## The Problem

| Issue | Impact | This ADR's Solution |
|-------|--------|-----|
| Context grows unbounded | Agent crashes at limit (~20 turns with 1M context Gemini) | Automatic compaction at 70% |
| No token visibility | Users don't know why conversations fail | Real-time metrics: `📊 45K/1M (4.5%) ~450 turns left` |
| Verbose outputs waste tokens | 10 KiB shell output = ~2.5% of model context | Head+tail truncation (128 lines start + end) |
| Unvalidated conversation state | Tool outputs can become orphaned after compaction | History normalization ensures call/output pairs |

**Business Impact**: Adk-code must support customer workflows that require 50+ turns of interaction. Without this, agent reliability is limited to single-session short tasks.

---

## Architecture

### Three-Layer System

```
Layer 1: OUTPUT TRUNCATION
├─ When: Immediately after each tool execution
├─ What: Head+tail (keep first 128 lines + last 128 lines)
├─ Size: Max 10 KiB per output
└─ Marker: "[... omitted X of Y lines ...]"

     ↓

Layer 2: TOKEN TRACKING
├─ When: After each LLM API call
├─ What: Record input + output tokens per turn
├─ Display: REPL metrics showing tokens used, turns remaining
└─ Limit: Hard stop at 95% of context window

     ↓

Layer 3: CONVERSATION COMPACTION
├─ When: Automatically at 70% context window
├─ What: LLM summarizes user intent + recent messages
├─ Result: [initial context] + [summary] + [last 10 messages]
└─ Savings: 50,000 token conversation → 5,000 tokens (~10x)
```

### Implementation Map

| Component | File | Size | Complexity | Days |
|-----------|------|------|-----------|------|
| ContextManager (core) | `internal/context/manager.go` | 300 L | Medium | 1.0 |
| Output truncation | `internal/context/truncate.go` | 80 L | Low | 0.5 |
| Token tracking | `internal/context/token_tracker.go` | 120 L | Low | 0.5 |
| Compaction logic | `internal/context/compaction.go` | 150 L | Medium | 1.0 |
| Instruction hierarchy | `internal/instructions/loader.go` | 150 L | Low | 0.5 |
| Unit tests | `internal/context/*_test.go` + `internal/instructions/*_test.go` | 600 L | High | 2.0 |
| Integration + REPL | Session/Agent/Display updates | 100 L | Medium | 1.5 |
| **Total** | - | **~1500 L** | - | **8-10 days** |

---

## Implementation

### 1. ContextManager (`internal/context/manager.go`)

**Core API**:
```go
package context

type ContextManager struct {
    items       []Item
    usedTokens  int64
    config      Config
    truncateLog []string
}

type Item struct {
    ID        string
    Type      ItemType // message|tool_call|tool_output|summary
    Content   string
    Tokens    int
    Original  int       // Pre-truncation size
    Truncated bool
}

// NewContextManager(cfg Config) *ContextManager
// AddItem(item Item) error                 // Returns ErrCompactionNeeded at 70%
// GetHistory() []Item                      // Conversation ready for model
// TokenInfo() TokenInfo                    // Current usage metrics
// TruncateLog() []string                   // Audit trail of all truncations
```

**Key methods**:
- `AddItem()` – Truncates tool outputs, estimates tokens, checks 70% threshold
- `TokenInfo()` – Returns: used tokens, available tokens, percent used, estimated turns remaining
- `truncateOutput()` – Implements head+tail: first 128 lines + last 128 lines, max 10 KiB

**Critical implementation detail**: Token tracking happens in `AddItem()`, NOT after API call. This enables accurate "turns remaining" estimates.

### 2. Output Truncation (`internal/context/truncate.go`)

**Algorithm**:
```
Input:  Full tool output (e.g., 10,000 lines, 500 KiB)
        ├─ Take first 128 lines
        ├─ Take last 128 lines
        └─ Omit middle with marker: "[... omitted 9,744 of 10,000 lines ...]"
Output: ~1 KiB (preserves critical start + end info)
```

**Pseudocode**:
```
if len(content) <= 10 KiB AND len(lines) <= 256:
    return content
else:
    head = first 128 lines
    tail = last 128 lines
    middle = omitted count
    return head + "\n[... omitted {middle} of {total} lines ...]\n" + tail
```

**Test case**: `TestTruncateOutput_PreservesStartAndEnd`
- Input: 1000 lines of output
- Expected: <260 lines returned, contains first line, contains last line, has elision marker

### 3. Token Tracking (`internal/context/token_tracker.go`)

**Purpose**: Maintain per-turn metrics for REPL display and compaction decisions.

**API**:
```go
type TokenTracker struct {
    turns []TurnUsage
}

type TurnUsage struct {
    Number       int
    InputTokens  int
    OutputTokens int
    Total        int
    Timestamp    time.Time
}

// Record(inputTokens, outputTokens int)
// AverageTurnSize() int
// Total() int
// Report() string // "Turn 5 | Input: 1200, Output: 850 | Avg/turn: 1050 | Total: 5250"
```

**Test case**: `TestTokenTracker_AccuracyWithin10Percent`
- Record 5 turns with known token counts
- Verify AverageTurnSize() calculates correctly

### 4. Conversation Compaction (`internal/context/compaction.go`)

**Trigger**: ContextManager.AddItem() returns `ErrCompactionNeeded` when >70% context used.

**Compaction flow**:
```
1. Collect all user messages from conversation
2. Call LLM: "Summarize user's intent in 2-3 sentences"
3. Keep: [initial system context] + [LLM summary] + [last 10 messages]
4. Result: 50K token conversation → ~5K tokens
5. Agent resumes with compacted history
```

**API**:
```go
func Compact(
    ctx context.Context,
    items []Item,
    modelCallFn func(context.Context, string) (string, error),
) (CompactionResult, error)

type CompactionResult struct {
    OriginalItems    int
    CompactedItems   int
    TokensSaved      int
    CompressionRatio float64
    Summary          string
}
```

**Test case**: `TestCompact_ReducesTokensBy10x`
- Input: 50,000 tokens of conversation history
- Expected: Output <5,000 tokens (compression ratio ~10x)
- Verify: Summary is 2-3 sentences

### 5. Instruction Hierarchy (`internal/instructions/loader.go`)

**Purpose**: Load user-provided instructions (AGENTS.md) from 3 levels, apply to each agent operation.

**Loading order** (first found wins):
1. `~/.adk-code/AGENTS.md` (global)
2. `$PROJECT_ROOT/AGENTS.md` (project-level)
3. `$PWD/AGENTS.md` (working directory)

**Merge**: Concatenate in order, enforce 32 KiB size limit.

**Example use case**: Customer can create `~/.adk-code/AGENTS.md`:
```
You are helping with TypeScript/React code.
Always use hooks, never class components.
Prefer functional patterns over OOP.
When suggesting packages, verify they're actively maintained.
```

**API**:
```go
type Loader struct {
    globalPath string
    projRoot   string
    workDir    string
}

func NewLoader(workDir string) *Loader
func (l *Loader) Load() (LoadResult, error)

type LoadResult struct {
    Merged     string
    NumSources int
    Bytes      int
    Truncated  bool
}
```

**Test case**: `TestLoad_MergesThreeLevels`
- Create temp files at 3 levels
- Verify merged result contains content from each level
- Verify size limit enforced (32 KiB)

---

## Testing Strategy

### Unit Tests (5 test files, ~600 lines total)

**context/manager_test.go**:
```go
TestContextManager_AddItem_TruncatesOutput // Output > 10 KiB → truncated
TestContextManager_AddItem_EstimatesTokens // Token count accurate ±10%
TestContextManager_TokenInfo_CalculatesUsage // Percent used calculated correctly
TestContextManager_AddItem_DetectsCompactionThreshold // At 70% returns ErrCompactionNeeded
```

**context/truncate_test.go**:
```go
TestTruncate_PreservesFirstAndLastLines // First 128 + last 128 preserved
TestTruncate_AddsElisionMarker // "[... omitted X of Y lines ...]" present
TestTruncate_RespontsToByteLimit // Result ≤ 10 KiB
TestTruncate_IdentityWhenUnderLimit // Small outputs returned unchanged
```

**context/token_tracker_test.go**:
```go
TestTokenTracker_RecordsAccurately // Per-turn metrics stored correctly
TestTokenTracker_AverageTurnSize_CalculatesCorrectly // Mean computed correctly
TestTokenTracker_Report_FormatsOutput // String output human-readable
```

**context/compaction_test.go**:
```go
TestCompact_ReducesConversation // 50K → ~5K tokens
TestCompact_PreservesUserIntent // Summary 2-3 sentences, accurate
TestCompact_KeepsRecentMessages // Last 10 messages retained in full
TestCompact_ErrorHandling_WhenNoMessages // Graceful error if no content to compact
```

**instructions/loader_test.go**:
```go
TestLoader_LoadsGlobal // ~/.adk-code/AGENTS.md loaded if exists
TestLoader_LoadsProjectRoot // $ROOT/AGENTS.md loaded if exists
TestLoader_LoadsWorkingDir // $PWD/AGENTS.md loaded if exists
TestLoader_MergesInOrder // Global + Project + Working-dir order
TestLoader_EnforcesSizeLimit // Total ≤ 32 KiB
TestLoader_FindsProjectRoot // Detects .git/.hg/go.mod/package.json
```

### Integration Tests (session_integration_test.go)

```go
TestContextManagement_FullWorkflow // 50 turns, auto-compaction at 70%, metrics display
TestTruncation_PreservesToolOutput // Tool outputs properly truncated, content accessible
TestCompaction_TransparentToAgent // Agent resumes seamlessly after compaction
TestInstructions_AppliedToAllTurns // Custom AGENTS.md instructions reflected in responses
```

---

## Success Criteria (Measurable)

| Criterion | Metric | Verification Method |
|-----------|--------|-----|
| Truncation accuracy | First line preserved + last line preserved + elision marker | Parse result, assert contains first + last + marker |
| Token accuracy | Estimated tokens within ±10% of actual LLM response | Compare estimateTokens() output vs. UsageMetadata |
| Compaction triggers | ErrCompactionNeeded returned when ≥70% context used | Verify error at exact threshold |
| Compaction effectiveness | 50K token conversation reduces to <5K | Measure token count before/after, assert 10x compression |
| Instruction loading | All 3 levels loaded, total ≤32 KiB | Write test files, verify merged result + size |
| History validity | No orphaned tool outputs after compaction | Validate call/output pairs in normalized history |
| REPL metrics | Token info displayed accurately (used/available/percent) | Parse REPL output, verify calculations |
| Long conversations | 50+ turns complete without crash | Integration test: run 50-turn workflow |
| No silent data loss | All truncations logged, visible in audit trail | Check TruncateLog() has entries for each truncation |

---

## Integration Points

### 1. Session Creation (`internal/session/manager.go`)
```go
// In CreateSession()
session.ContextManager = context.NewContextManager(Config{
    ModelName:        selectedModel.Name,
    ContextWindow:    selectedModel.ContextWindow,
    ReservedPercent:  0.1, // 10% reserved for output
    OutputMaxBytes:   10 * 1024,
    OutputMaxLines:   256,
    HeadLines:        128,
    TailLines:        128,
    CompactThreshold: 0.70,
})

session.TokenTracker = context.NewTokenTracker()

session.InstructionLoader = instructions.NewLoader(workDir)
session.Instructions = session.InstructionLoader.Load()
```

### 2. Agent Loop (`pkg/agents/agent.go`)
```go
// After tool execution, before returning to user
err := session.ContextManager.AddItem(context.Item{
    Type:    context.ItemToolOutput,
    Content: toolResult,
    Tokens:  estimateTokens(toolResult),
})

if err == context.ErrCompactionNeeded {
    // Trigger async compaction
    go func() {
        compactResult, err := context.Compact(ctx, 
            session.ContextManager.GetHistory(),
            func(ctx context.Context, prompt string) (string, error) {
                // Call LLM with compaction prompt
                return llmClient.Generate(ctx, prompt)
            },
        )
        if err == nil {
            session.ContextManager.Replace(compactResult.Items)
            display.ShowCompaction(compactResult)
        }
    }()
}
```

### 3. REPL Display (`internal/display/metrics.go`)
```go
// After each turn completes
tokenInfo := session.ContextManager.TokenInfo()
tracker := session.TokenTracker

fmt.Printf("\n📊 Context Usage:\n")
fmt.Printf("   Tokens: %d / %d (%.1f%% used)\n",
    tokenInfo.UsedTokens,
    tokenInfo.AvailableTokens,
    tokenInfo.PercentageUsed * 100,
)
fmt.Printf("   Latest turn: %s\n", tracker.Report())
fmt.Printf("   Compaction threshold: %.0f%%\n", 
    tokenInfo.CompactThreshold * 100,
)
if tokenInfo.TurnsRemaining > 0 {
    fmt.Printf("   Est. %d turns remaining\n", tokenInfo.TurnsRemaining)
}
```

### 4. Model Registry (`pkg/models/registry.go`)
```go
// Add context window to each model
gemini25Flash := &ModelInfo{
    Name:              "gemini-2.5-flash",
    Provider:          "google",
    ContextWindow:     1_000_000,  // 1M tokens
    MaxOutputTokens:   8_000,
    CostPer1MInput:    0.075,
    CostPer1MOutput:   0.30,
}
```

---

## Files to Create

```
internal/context/
├── manager.go              (300 L, ~1.0 days)
├── manager_test.go         (150 L, ~0.5 days)
├── truncate.go             (80 L, ~0.5 days)
├── truncate_test.go        (100 L, ~0.5 days)
├── token_tracker.go        (120 L, ~0.5 days)
├── token_tracker_test.go   (80 L, ~0.3 days)
├── compaction.go           (150 L, ~1.0 days)
└── compaction_test.go      (120 L, ~0.5 days)

internal/instructions/
├── loader.go               (150 L, ~0.5 days)
└── loader_test.go          (100 L, ~0.5 days)
```

## Files to Modify

```
internal/session/manager.go
├── Add: ContextManager field to Session struct
└── Add: Initialize ContextManager in CreateSession()

pkg/agents/agent.go
├── Add: Call ContextManager.AddItem() after tool execution
└── Add: Handle ErrCompactionNeeded, trigger compaction

internal/display/
├── Create: metrics.go (show token usage in REPL)
└── Add: Compaction notification display

pkg/models/registry.go
├── Add: ContextWindow int field to ModelInfo
└── Update: All model definitions with context window
```

---

## Timeline & Milestones

**Week 1**:
- Day 1-2: ContextManager + truncate.go + tests → production ready
- Day 2.5-3: token_tracker.go + compaction.go + tests → ready for integration
- Day 3.5-4: Session/Agent integration + REPL metrics → demo-ready

**Week 2**:
- Day 1: instructions/loader.go + tests → complete
- Day 1.5-2: Full integration testing, edge cases
- Day 2-3: Documentation, examples, edge cases

**Total**: 8-10 days from first line of code to production ready.

---

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Compaction latency (2-5s) | User perceives hang | Show spinner, run async, explain to user |
| Summary loses critical context | Agent performs worse after compaction | Retain all recent messages (last 10) in full, just summarize old history |
| Token counting off by >10% | Compaction triggers too early/late | Test against actual model UsageMetadata, calibrate estimateTokens() |
| History normalization breaks user intent | After compaction, conversation flow confused | Always maintain call/output pairs, keep messages in order |

---

## Alternatives Rejected

**Option 1: Simple FIFO Removal** – Just discard oldest messages
- ❌ Irreversible loss of context
- ❌ No user visibility
- ❌ Unpredictable agent behavior

**Option 2: Per-Tool Truncation Only** – Never do compaction
- ❌ Only handles outputs, not conversation growth
- ❌ Limited to ~10-20 turn workflows (not 50+)
- ❌ Ignores proven Codex approach

**Option 3: Manual User-Triggered Compaction** – User decides when
- ❌ Requires user awareness of context limits
- ❌ Creates friction, fails silently if user forgets
- ❌ Codex/Claude do automatic, superior approach

**→ Selected Option: Three-Layer Auto Management** (chosen above) – Proven, scalable, transparent

---

## Code Locations (Reference)

**Proven implementations to reference**:
- Codex truncation strategy: `research/codex/codex-rs/core/src/context_manager/truncate.rs`
- Codex compaction: `research/codex/codex-rs/core/src/compact.rs`
- Google ADK session system: `research/adk-go/session/session.go`
- Current adk-code session: `internal/session/manager.go`
- Current adk-code models: `pkg/models/registry.go`

---

## Checklist: Implementation

**Phase 1: Core (Day 1-2)**
- [ ] Create `internal/context/manager.go` with AddItem(), GetHistory(), TokenInfo()
- [ ] Create `internal/context/truncate.go` with head+tail algorithm
- [ ] Write 200+ test assertions covering truncation, token tracking
- [ ] `make test` passes (0 failures)
- [ ] `make lint` passes (0 warnings in context package)

**Phase 2: Compaction (Day 2.5-3)**
- [ ] Create `internal/context/compaction.go`
- [ ] Integrate ContextManager.AddItem() check into agent loop
- [ ] Handle ErrCompactionNeeded gracefully
- [ ] Write compaction tests (compression ratio, intent preservation)

**Phase 3: Instructions (Day 3.5-4)**
- [ ] Create `internal/instructions/loader.go`
- [ ] Implement 3-level loading (global/project/working-dir)
- [ ] Test with nested directory structures
- [ ] Validate 32 KiB limit enforced

**Phase 4: Integration (Day 4.5-5)**
- [ ] Update `internal/session/manager.go` to initialize ContextManager
- [ ] Update `pkg/agents/agent.go` to call AddItem()
- [ ] Create `internal/display/metrics.go` for REPL output
- [ ] Update `pkg/models/registry.go` with context windows

**Phase 5: Testing & Validation (Day 5-7)**
- [ ] Integration test: 50-turn workflow completes successfully
- [ ] Truncation test: Output >10 KiB verified preserved (start + end)
- [ ] Token accuracy test: ±10% vs. actual model response
- [ ] Instruction test: AGENTS.md loaded and applied
- [ ] Compaction test: 50K → <5K compression verified
- [ ] `make check` passes (fmt, vet, lint, test)

**Phase 6: Documentation (Day 7-8)**
- [ ] Update `ARCHITECTURE.md` – context management section
- [ ] Write `AGENTS.md` user guide (example file)
- [ ] Add REPL `/tokens` and `/compact` commands to help
- [ ] Create troubleshooting guide (when compaction triggers, what it means)

**Pre-Commit**:
- [ ] All tests pass (`make test`)
- [ ] All lints pass (`make check`)
- [ ] 50-turn integration test passes
- [ ] Token accuracy within ±10%
- [ ] Code review approval

---

## Status: Ready for Development

All decisions made. Implementation structure clear. Test strategy defined. Integration points mapped. References provided. Ready to assign to development team.
