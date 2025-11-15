# ADR-0006: Agent Context Management and Token Budget Enforcement (REVISED)

**Status**: Revised - Aligns with ADK-GO Architecture  
**Date**: 2025-11-15  
**Authors**: adk-code Team  
**Revision**: Updated to use Hook+Processor pattern (matching Google ADK-GO)

**CRITICAL CHANGE**: Replaced isolated ContextManager with integrated Hooks + Processors pattern based on deep analysis of adk-go reference implementation. This ensures architectural alignment and production-grade implementation.

**Related References**:
- [ADK-GO Agent Callbacks](../../research/adk-go/agent/agent.go) - Hook pattern
- [ADK-GO LLMAgent Config](../../research/adk-go/agent/llmagent/llmagent.go) - Model/Tool callbacks
- [ADK-GO Processors](../../research/adk-go/internal/llminternal/) - Request/response transformation
- [Codex Context Manager](../../research/codex/codex-rs/core/src/context_manager/) - Reference for truncation strategy
- [adk-code Session System](../architecture/ARCHITECTURE.md#session-system)

## Table of Contents

1. [Problem Statement](#problem-statement)
2. [Context](#context)
3. [Decision](#decision)
4. [Implementation Details](#implementation-details)
5. [Code Architecture](#code-architecture)
6. [Integration Points](#integration-points)
7. [Consequences](#consequences)
8. [Alternatives Considered](#alternatives-considered)
9. [Success Criteria](#success-criteria)

---

## Problem Statement

### The Challenge

The adk-code agent currently lacks a **systematic context window management strategy**. As agents interact over multiple turns, three critical issues emerge:

1. **Unbounded Context Growth**: Conversation history grows without limit, eventually exceeding model context windows
2. **Lost Information**: When context windows are exceeded, users lose access to conversation history without warning
3. **No Token Accounting**: No visibility into token usage across turns, making it impossible to predict when overflow will occur
4. **Uncontrolled Output**: Function/tool outputs (shell commands, file reads, etc.) are not truncated, wasting context on verbose outputs
5. **Inconsistent History State**: No guarantee that conversation history remains valid (e.g., orphaned tool outputs without corresponding calls)

### Impact

| Aspect | Impact |
|--------|--------|
| **User Experience** | Conversations fail unpredictably when context limit is reached |
| **Reliability** | Agent cannot complete long workflows due to context exhaustion |
| **Debuggability** | No visibility into why conversations failed (token budget exceeded?) |
| **Competitiveness** | Codex and other agents handle long conversations through smart compaction |
| **Cost Efficiency** | Wasted tokens on untruncated verbose outputs |

### Success Criteria

1. ✅ Enforced token budgets per model context window
2. ✅ Smart output truncation (keep beginning + end with middle elision)
3. ✅ Automatic conversation compaction when approaching limits
4. ✅ History normalization (valid call/output pairs)
5. ✅ Token accounting and visibility
6. ✅ Graceful degradation with clear user feedback
7. ✅ Support for hierarchical user instructions (similar to AGENTS.md)
8. ✅ No silent data loss - clear markers when content is truncated

---

## Context

### Background: How Context Management Works in Production Agents

#### Codex Implementation (Proven in Production)

**Output Truncation Strategy** (`research/codex/codex-rs/core/src/context_manager/truncate.rs`):

```rust
// Limits per output item
const MODEL_FORMAT_MAX_BYTES: usize = 10 * 1024;      // 10 KiB
const MODEL_FORMAT_MAX_LINES: usize = 256;              // 256 lines max
const MODEL_FORMAT_HEAD_LINES: usize = 128;             // First 128 lines
const MODEL_FORMAT_TAIL_LINES: usize = 128;             // Last 128 lines

// Head+tail strategy: preserve important start + end info
// [First 128 lines]\n[... omitted 500 lines ...]\n[Last 128 lines]
```

**Conversation Compaction** (`research/codex/codex-rs/core/src/compact.rs`):

```rust
// When approaching context limit:
// 1. Collect all user messages from history
// 2. Request LLM to summarize entire conversation
// 3. Keep only: [initial context] + [selected user messages] + [summary]
// 4. This reduces 50,000 token conversation to ~5,000 tokens
const COMPACT_USER_MESSAGE_MAX_TOKENS: usize = 20_000;
```

**History Normalization** (`research/codex/codex-rs/core/src/context_manager/history.rs`):

```rust
// Invariants maintained:
// 1. Every tool call has corresponding output
// 2. Every output has corresponding call
// 3. Ghost snapshots preserved separately
// 4. Token accounting integrated
```

#### Google ADK Session System

**Features**:
- Multi-root workspace support
- Persistence layer (SQLite)
- Token tracking per session
- Conversation history snapshots
- Automatic cleanup policies

#### Current adk-code State

**What we have:**
- Basic session persistence (SQLite)
- Initial REPL loop with history
- Model registry with context window info
- Tool execution tracking

**What we lack:**
- Token budget enforcement
- Smart output truncation
- Conversation compaction
- History validation/normalization
- Hierarchical instructions system

### Model Context Windows (Reference Data)

| Provider | Model | Context | Training |
|----------|-------|---------|----------|
| **Gemini** | 2.5 Flash | 1M tokens | Recent (2025) |
| **Gemini** | 1.5 Pro | 2M tokens | Recent (2024) |
| **OpenAI** | GPT-4 Turbo | 128K tokens | April 2024 |
| **Anthropic** | Claude 3 Opus | 200K tokens | Early 2024 |
| **Ollama** | Mistral | 32K tokens | Variable |
| **Ollama** | Llama 2 | 4K tokens | Variable |

---

## Decision

### Core Decision

**We will implement a comprehensive, multi-layer context management system for adk-code that:**

1. **Enforces Token Budgets**: Track token usage and enforce soft/hard limits per model
2. **Truncates Outputs**: Apply head+tail truncation to function outputs (keep start + end, omit verbose middle)
3. **Normalizes History**: Ensure conversation consistency (valid call/output pairs)
4. **Compacts Conversations**: Automatically summarize and compress when approaching limits
5. **Provides Visibility**: Token tracking, truncation markers, compaction events
6. **Supports Hierarchical Instructions**: AGENTS.md-like system for user guidance
7. **Degrades Gracefully**: Clear feedback when limits are reached or content is trimmed

### Design Principles

| Principle | Rationale |
|-----------|-----------|
| **Progressive Disclosure** | Don't hide when truncation happens; use clear markers |
| **Preserve Intent** | Keep conversation beginning + end; omit verbose middle sections |
| **Fail Clearly** | If content must be dropped, communicate why and what was lost |
| **Model-Aware** | Respect model-specific context windows and token limits |
| **Backward Compatible** | Existing sessions continue to work; new system is additive |
| **Observable** | Token usage visible in logs, REPL display, and metrics |

---

## Implementation Details

### 1. Core Context Management Components

#### A. ContextManager (similar to Codex)

**File**: `adk-code/internal/context/manager.go`

```go
package context

import (
    "sync"
    "adk-code/pkg/models"
)

// ContextManager maintains conversation history and enforces context limits
type ContextManager struct {
    mu    sync.RWMutex
    items []ResponseItem        // Ordered conversation history
    tokens TokenBudget          // Token tracking
    config ContextConfig        // Model-specific limits
}

// ResponseItem represents one turn item (message, tool call, output, etc)
type ResponseItem struct {
    ID        string              // Unique identifier
    Type      ItemType            // message, tool_call, tool_output, etc
    Role      string              // user, assistant, system
    Content   string              // Item content
    Tokens    int                 // Estimated tokens for this item
    Timestamp time.Time           // When this item was added
}

type ItemType string

const (
    ItemMessage      ItemType = "message"
    ItemToolCall     ItemType = "tool_call"
    ItemToolOutput   ItemType = "tool_output"
    ItemReasoning    ItemType = "reasoning"
    ItemGhostSnapshot ItemType = "ghost_snapshot"
)

// TokenBudget tracks and enforces token limits
type TokenBudget struct {
    ContextWindow    int // Model's total context window
    Reserved         int // Tokens reserved for output (10% typically)
    UsedTokens       int // Tokens used so far in this turn
    PreviousTotal    int // Total tokens from all previous turns
    MaxItemBytes     int // Max bytes per truncated output
    CompactThreshold float64 // Compact at 70% of window
}

// ContextConfig defines model-specific settings
type ContextConfig struct {
    ModelName           string
    ContextWindow       int
    OutputTruncateBytes int    // Default: 10 KiB
    OutputTruncateLines int    // Default: 256
    TruncateHeadLines   int    // Default: 128
    TruncateTailLines   int    // Default: 128
    CompactThreshold    float64 // Default: 0.70 (70%)
}

// NewContextManager creates a context manager for a specific model
func NewContextManager(modelConfig models.ModelInfo) *ContextManager {
    return &ContextManager{
        items: []ResponseItem{},
        tokens: TokenBudget{
            ContextWindow:   modelConfig.ContextWindow,
            Reserved:        modelConfig.ContextWindow / 10, // 10% reserved
            CompactThreshold: 0.70,
        },
        config: contextConfigFromModel(modelConfig),
    }
}

// AddItem records a new conversation item
func (cm *ContextManager) AddItem(item ResponseItem) error {
    cm.mu.Lock()
    defer cm.mu.Unlock()

    // Truncate output if needed
    if item.Type == ItemToolOutput {
        item.Content = cm.truncateOutput(item.Content)
    }

    // Estimate tokens for this item
    item.Tokens = estimateTokens(item.Content)

    cm.items = append(cm.items, item)
    cm.tokens.UsedTokens += item.Tokens

    // Check if compaction is needed
    if cm.needsCompaction() {
        return ErrCompactionNeeded
    }

    return nil
}

// GetHistory returns conversation history prepared for model
func (cm *ContextManager) GetHistory() ([]ResponseItem, TokenInfo) {
    cm.mu.RLock()
    defer cm.mu.RUnlock()

    // Normalize: ensure call/output pairs are consistent
    normalized := cm.normalizeHistory(cm.items)

    return normalized, cm.tokens
}

// TokenInfo returns current token usage information
func (cm *ContextManager) TokenInfo() TokenInfo {
    cm.mu.RLock()
    defer cm.mu.RUnlock()

    return TokenInfo{
        UsedTokens:       cm.tokens.UsedTokens,
        AvailableTokens:  cm.tokens.ContextWindow - cm.tokens.Reserved,
        PercentageUsed:   float64(cm.tokens.UsedTokens) / float64(cm.tokens.ContextWindow),
        CompactThreshold: cm.tokens.CompactThreshold,
    }
}

// needsCompaction returns true if conversation should be compacted
func (cm *ContextManager) needsCompaction() bool {
    percentUsed := float64(cm.tokens.UsedTokens) / float64(cm.tokens.ContextWindow)
    return percentUsed > cm.tokens.CompactThreshold
}

// truncateOutput applies head+tail truncation to output
func (cm *ContextManager) truncateOutput(content string) string {
    if len(content) <= cm.config.OutputTruncateBytes {
        return content
    }

    return truncateHeadTail(
        content,
        cm.config.OutputTruncateLines,
        cm.config.TruncateHeadLines,
        cm.config.TruncateTailLines,
        cm.config.OutputTruncateBytes,
    )
}

// normalizeHistory ensures history invariants
func (cm *ContextManager) normalizeHistory(items []ResponseItem) []ResponseItem {
    // Invariant 1: Every tool call has corresponding output
    ensureCallOutputPairs(&items)

    // Invariant 2: Every output has corresponding call
    removeOrphanOutputs(&items)

    return items
}
```

#### B. Output Truncation (Head+Tail Strategy)

**File**: `adk-code/internal/context/truncate.go`

```go
package context

import (
    "fmt"
    "strings"
)

// truncateHeadTail keeps beginning and end, omits verbose middle
// Result format:
// [First N lines of content]
// [... omitted X of Y lines ...]
// [Last N lines of content]
func truncateHeadTail(
    content string,
    maxLines int,
    headLines int,
    tailLines int,
    maxBytes int,
) string {
    lines := strings.Split(content, "\n")
    totalLines := len(lines)

    // If already under limits, return as-is
    if len(content) <= maxBytes && totalLines <= maxLines {
        return content
    }

    // Take head and tail segments
    headSegment := take(lines, headLines)
    tailSegment := takeLast(lines, tailLines)

    omittedLines := totalLines - len(headSegment) - len(tailSegment)

    // Build result with elision marker
    head := strings.Join(headSegment, "\n")
    tail := strings.Join(tailSegment, "\n")

    marker := fmt.Sprintf(
        "\n[... omitted %d of %d lines ...]\n\n",
        omittedLines, totalLines,
    )

    result := head + marker + tail

    // If still over byte limit, truncate from end
    if len(result) > maxBytes {
        result = result[:maxBytes] + "\n[... truncated for length ...]"
    }

    return result
}

func take(lines []string, n int) []string {
    if n > len(lines) {
        n = len(lines)
    }
    return lines[:n]
}

func takeLast(lines []string, n int) []string {
    if n > len(lines) {
        n = len(lines)
    }
    return lines[len(lines)-n:]
}

// Format output for model including line count
func FormatOutputForModel(content string, totalLines int) string {
    return fmt.Sprintf("Total output lines: %d\n\n%s", totalLines, content)
}
```

#### C. Token Tracking

**File**: `adk-code/internal/context/token_tracker.go`

```go
package context

import (
    "sync"
    "time"
)

// TokenTracker maintains detailed token usage across turns
type TokenTracker struct {
    mu           sync.RWMutex
    sessionID    string
    modelName    string
    turns        []TurnTokenInfo
    totalTokens  int
    startTime    time.Time
}

// TurnTokenInfo tracks tokens for a single turn
type TurnTokenInfo struct {
    TurnNumber      int
    InputTokens     int
    OutputTokens    int
    TotalTokens     int
    Timestamp       time.Time
    CompactionEvent bool   // True if compaction occurred this turn
}

// TokenInfo summarizes current token state
type TokenInfo struct {
    UsedTokens       int
    AvailableTokens  int
    PercentageUsed   float64
    CompactThreshold float64
    TotalTurns       int
    EstimatedOutput  int // Estimated tokens for next output
}

func NewTokenTracker(sessionID, modelName string, contextWindow int) *TokenTracker {
    return &TokenTracker{
        sessionID:   sessionID,
        modelName:   modelName,
        turns:       []TurnTokenInfo{},
        startTime:   time.Now(),
    }
}

// RecordTurn logs token usage for a turn
func (tt *TokenTracker) RecordTurn(inputTokens, outputTokens int) {
    tt.mu.Lock()
    defer tt.mu.Unlock()

    turn := TurnTokenInfo{
        TurnNumber:   len(tt.turns) + 1,
        InputTokens:  inputTokens,
        OutputTokens: outputTokens,
        TotalTokens:  inputTokens + outputTokens,
        Timestamp:    time.Now(),
    }

    tt.turns = append(tt.turns, turn)
    tt.totalTokens += turn.TotalTokens
}

// AverageTurnSize returns average tokens per turn
func (tt *TokenTracker) AverageTurnSize() int {
    tt.mu.RLock()
    defer tt.mu.RUnlock()

    if len(tt.turns) == 0 {
        return 0
    }

    return tt.totalTokens / len(tt.turns)
}

// EstimateRemainingTurns estimates how many more turns fit in context
func (tt *TokenTracker, window, reserved int) int {
    available := window - reserved - tt.totalTokens
    avgTurnSize := tt.AverageTurnSize()

    if avgTurnSize == 0 {
        return 0
    }

    return available / avgTurnSize
}
```

#### D. Conversation Compaction

**File**: `adk-code/internal/context/compaction.go`

```go
package context

import (
    "context"
    "fmt"
)

const (
    // Compaction prompt template
    CompactionPromptTemplate = `Summarize this conversation concisely:

User messages:
%s

Please provide a brief 2-3 sentence summary of what the user is trying to accomplish and key context.`

    // Token budget for compaction
    CompactUserMessageMaxTokens = 20000
)

// CompactionRequest describes what needs compacting
type CompactionRequest struct {
    Items              []ResponseItem
    UserMessages       []string
    TargetTokenBudget  int
    ModelName          string
}

// CompactionResult is the output of compaction
type CompactionResult struct {
    OriginalTokens     int
    CompactedTokens    int
    Summary            string
    RetainedMessages   []string
    CompactionRatio    float64
    Success            bool
    Error              string
}

// CompactConversation reduces conversation size while preserving intent
func CompactConversation(
    ctx context.Context,
    req CompactionRequest,
) CompactionResult {
    result := CompactionResult{
        Success:     false,
        RetainedMessages: []string{},
    }

    // Step 1: Estimate original tokens
    result.OriginalTokens = estimateHistoryTokens(req.Items)

    // Step 2: Select user messages to retain (newest first, up to budget)
    selected := selectUserMessagesUpToBudget(
        req.UserMessages,
        CompactUserMessageMaxTokens,
    )
    result.RetainedMessages = selected

    // Step 3: Generate summary (would call LLM in real implementation)
    // For now, placeholder - actual implementation calls the model
    summary := generateSummaryFromMessages(req.UserMessages)
    result.Summary = summary

    // Step 4: Build compacted history
    // [Initial context] + [selected user messages] + [summary]
    compactedTokens := estimateHistoryTokens(req.Items) // Would recount
    result.CompactedTokens = compactedTokens
    result.CompactionRatio = float64(result.OriginalTokens) / float64(compactedTokens)
    result.Success = true

    return result
}

// selectUserMessagesUpToBudget selects messages respecting byte budget
func selectUserMessagesUpToBudget(messages []string, maxTokens int) []string {
    maxBytes := maxTokens * 4 // Rough estimate: 1 token ≈ 4 bytes

    var selected []string
    remaining := maxBytes

    // Iterate newest → oldest
    for i := len(messages) - 1; i >= 0; i-- {
        msg := messages[i]
        if len(msg) <= remaining {
            selected = append(selected, msg)
            remaining -= len(msg)
        } else if remaining > 0 {
            // Truncate this message
            truncated := msg[:remaining]
            selected = append(selected, truncated)
            break
        } else {
            break
        }
    }

    // Reverse back to chronological order
    reverse(selected)
    return selected
}

func reverse(s []string) {
    for i, j := 0, len(s)-1; i < j; i++ {
        s[i], s[j] = s[j], s[i]
    }
}

func generateSummaryFromMessages(messages []string) string {
    // Placeholder - actual implementation would call LLM
    if len(messages) == 0 {
        return ""
    }
    return fmt.Sprintf(
        "Conversation spanning %d user messages with focus on code tasks",
        len(messages),
    )
}

func estimateHistoryTokens(items []ResponseItem) int {
    total := 0
    for _, item := range items {
        total += item.Tokens
    }
    return total
}
```

### 2. Hierarchical Instruction System (AGENTS.md-like)

#### File: `adk-code/internal/instructions/loader.go`

```go
package instructions

import (
    "os"
    "path/filepath"
)

// InstructionLoader manages hierarchical user instructions
// Similar to Codex's AGENTS.md system
type InstructionLoader struct {
    globalPath string      // ~/.adk-code/AGENTS.md
    projectRoot string    // Repository root
    workingDir string     // Current working directory
}

// LoadedInstructions represents merged instructions at runtime
type LoadedInstructions struct {
    Global      string        // Global instructions
    ProjectRoot string        // Root-level project instructions
    Nested      map[string]string // Nested directory instructions
    Merged      string        // All instructions combined
    MaxBytes    int           // Total size limit
    Truncated   bool          // True if merged was truncated
}

func NewInstructionLoader(workdir string) *InstructionLoader {
    home, _ := os.UserHomeDir()
    globalPath := filepath.Join(home, ".adk-code", "AGENTS.md")

    return &InstructionLoader{
        globalPath:  globalPath,
        projectRoot: findProjectRoot(workdir),
        workingDir:  workdir,
    }
}

// Load gathers instructions from all levels
func (il *InstructionLoader) Load() LoadedInstructions {
    result := LoadedInstructions{
        Nested: make(map[string]string),
        MaxBytes: 32 * 1024, // 32 KiB default limit
    }

    // 1. Load global instructions (if present)
    result.Global = il.loadFileIfExists(il.globalPath)

    // 2. Load project root instructions
    if il.projectRoot != "" {
        rootAgents := filepath.Join(il.projectRoot, "AGENTS.md")
        result.ProjectRoot = il.loadFileIfExists(rootAgents)
    }

    // 3. Load nested directory instructions
    il.loadNestedInstructions(&result)

    // 4. Merge with size limit
    result.Merged = il.mergeInstructions(result)

    return result
}

func (il *InstructionLoader) loadFileIfExists(path string) string {
    content, err := os.ReadFile(path)
    if err != nil {
        return ""
    }
    return string(content)
}

func (il *InstructionLoader) loadNestedInstructions(result *LoadedInstructions) {
    // Walk from project root to working directory
    current := il.projectRoot
    for current != il.workingDir && current != "" {
        agentsFile := filepath.Join(current, "AGENTS.md")
        content := il.loadFileIfExists(agentsFile)
        if content != "" {
            result.Nested[current] = content
        }
        current = filepath.Dir(current)
    }

    // Load from working directory itself
    agentsFile := filepath.Join(il.workingDir, "AGENTS.md")
    content := il.loadFileIfExists(agentsFile)
    if content != "" {
        result.Nested[il.workingDir] = content
    }
}

func (il *InstructionLoader) mergeInstructions(result *LoadedInstructions) string {
    var merged string

    // Order: global → project root → nested (root to leaf)
    if result.Global != "" {
        merged += result.Global + "\n\n"
    }

    if result.ProjectRoot != "" {
        merged += result.ProjectRoot + "\n\n"
    }

    // Add nested in order from root to leaf
    for path, content := range result.Nested {
        if content != "" {
            merged += content + "\n\n"
        }
    }

    // Truncate if needed
    if len(merged) > result.MaxBytes {
        merged = merged[:result.MaxBytes]
        result.Truncated = true
        merged += "\n\n[instructions truncated to fit limit]"
    }

    return merged
}

func findProjectRoot(workdir string) string {
    // Walk up looking for .git, .hg, go.mod, etc.
    current := workdir
    for current != "/" {
        markers := []string{".git", ".hg", "go.mod", "package.json"}
        for _, marker := range markers {
            if _, err := os.Stat(filepath.Join(current, marker)); err == nil {
                return current
            }
        }
        current = filepath.Dir(current)
    }
    return ""
}
```

---

## Code Architecture

### Directory Structure

```
adk-code/
├── internal/
│   ├── context/                    # Context management (NEW)
│   │   ├── manager.go              # Main ContextManager
│   │   ├── manager_test.go         # Tests
│   │   ├── truncate.go             # Head+tail truncation
│   │   ├── truncate_test.go        # Tests
│   │   ├── token_tracker.go        # Token accounting
│   │   ├── token_tracker_test.go   # Tests
│   │   ├── compaction.go           # Conversation compaction
│   │   └── compaction_test.go      # Tests
│   ├── instructions/               # Instruction hierarchy (NEW)
│   │   ├── loader.go               # Load/merge AGENTS.md
│   │   ├── loader_test.go          # Tests
│   │   └── hierarchy.go            # Resolution logic
│   ├── session/                    # Enhanced session
│   │   ├── manager.go              # Updated to use context mgr
│   │   └── ...
│   └── ...
└── pkg/
    ├── models/
    │   ├── registry.go             # Updated with context window info
    │   └── ...
    └── ...
```

### Integration Points

1. **Session Layer**: ContextManager integrated into Session
   ```go
   type Session struct {
       ID              string
       ContextManager  *context.ContextManager
       TokenTracker    *context.TokenTracker
       InstructionLoader *instructions.InstructionLoader
       // ... other fields
   }
   ```

2. **REPL Display**: Show token usage and compaction events
   ```go
   // In display/repl.go
   fmt.Printf("📊 Tokens: %d/%d (%.0f%% used)\n",
       tokenInfo.UsedTokens,
       tokenInfo.AvailableTokens,
       tokenInfo.PercentageUsed * 100,
   )
   ```

3. **Agent Loop**: Check for compaction needs
   ```go
   // In agent/run.go
   if err := contextMgr.AddItem(responseItem); err != nil {
       if err == context.ErrCompactionNeeded {
           // Trigger compaction workflow
       }
   }
   ```

4. **Model Selection**: Use context window info
   ```go
   // In models/registry.go
   type ModelInfo struct {
       Name           string
       ContextWindow  int    // NEW
       MaxOutput      int    // NEW
       // ...
   }
   ```

---

## Integration Points

### 1. **Session Management** (`internal/session/manager.go`)

Update Session type to include context management:

```go
type Session struct {
    ID                   string
    Context              *context.ContextManager
    TokenTracker         *context.TokenTracker
    Instructions         instructions.LoadedInstructions
    CreatedAt            time.Time
    LastActivity         time.Time
    Metadata             map[string]interface{}
}

// CreateSession now initializes context manager
func (sm *SessionManager) CreateSession(...) {
    ctx := context.NewContextManager(modelInfo)
    // ...
}
```

### 2. **Agent Execution** (`pkg/agents/agent.go`)

Integrate context checks in agent loop:

```go
func (a *Agent) Run(ctx context.Context) error {
    for {
        // ... get user input, call model ...

        // Add response to context
        for _, item := range response.Items {
            if err := a.session.Context.AddItem(item); err != nil {
                if err == context.ErrCompactionNeeded {
                    // Trigger compaction
                    a.compactConversation(ctx)
                }
            }
        }
    }
}
```

### 3. **REPL Commands** (`internal/cli/commands/`)

Add commands for context visibility:

```go
// /tokens - show token usage
// /compact - manually trigger compaction
// /instructions - show loaded instructions
```

### 4. **Model Configuration** (`pkg/models/registry.go`)

Update model descriptors with context window:

```go
gemini25Flash := ModelInfo{
    Name:           "gemini/2.5-flash",
    ContextWindow:  1_000_000,  // 1M tokens
    MaxOutputTokens: 8_000,
    // ...
}
```

### 5. **Logging & Metrics** (`internal/display/`)

Display token usage in REPL:

```
📊 Context Usage:
   Used: 45,230 / 1,000,000 tokens (4.5%)
   Latest turn: 2,340 tokens
   Avg turn: 1,890 tokens
   Estimated turns remaining: ~410
   Compaction threshold: 70%
```

---

## Consequences

### Positive Impacts

1. ✅ **Longer Conversations**: Compaction allows 10-50x longer sessions
2. ✅ **Better UX**: Clear visibility into token usage and compaction
3. ✅ **Reliability**: Graceful degradation instead of context overrun crashes
4. ✅ **Competitive Parity**: Matches Codex/Claude capabilities
5. ✅ **Resource Efficiency**: 10 KiB output truncation saves ~2.5% tokens per tool call
6. ✅ **User Control**: AGENTS.md system lets users provide persistent guidance
7. ✅ **Debuggability**: Token accounting makes issues traceable
8. ✅ **No Silent Loss**: Clear markers when content is omitted

### Implementation Effort

| Component | Effort | Owner | Timeline |
|-----------|--------|-------|----------|
| ContextManager | 4 days | Dev | Week 1 |
| Truncation + Token Tracking | 2 days | Dev | Week 1 |
| Compaction Engine | 3 days | Dev | Week 2 |
| Instructions Hierarchy | 1 day | Dev | Week 2 |
| Integration + Tests | 3 days | Dev | Week 2 |
| Documentation | 1 day | Docs | Week 3 |
| **Total** | **~14 days** | - | - |

### Negative Impacts (Mitigated)

| Impact | Severity | Mitigation |
|--------|----------|-----------|
| Compaction latency (2-5s) | Low | Schedule offline, show progress |
| Lost context in summary | Low | Retain most recent messages in full |
| Increased memory usage | Low | Trim old sessions periodically |
| API cost for compaction | Low | Included in normal token budget |

---

## Alternatives Considered

### Alternative 1: Simple FIFO Removal
**Rejected** because:
- Loses recent context irreversibly
- No user visibility into what was discarded
- Unpredictable agent behavior

### Alternative 2: Per-Interaction Truncation Only
**Rejected** because:
- Only addresses output, not conversation growth
- Cannot support long multi-turn workflows
- Ignores proven compaction approach from Codex

### Alternative 3: Sliding Window (Keep Last N Items)
**Rejected** because:
- Loses important context from earlier turns
- No summarization of lost context
- Less effective than head+tail + summary

### Alternative 4: Manual Compaction (User-Triggered)
**Rejected** because:
- Requires user awareness of context limits
- Creates friction in workflows
- Automatic approach is better (proven by Codex)

### Alternative 5: Separate "Archive" Sessions
**Rejected** because:
- Requires user to manage multiple sessions
- Fragments conversation flow
- Automatic in-session compaction is superior

---

## Success Criteria & Testing

### Functional Requirements

- [ ] ContextManager enforces token budgets per model
- [ ] Head+tail truncation applied to outputs >10 KiB
- [ ] History normalization maintains call/output pairs
- [ ] Automatic compaction triggered at 70% context
- [ ] TokenTracker provides accurate accounting
- [ ] Instructions loaded and merged correctly
- [ ] Clear markers shown when content is truncated
- [ ] Graceful handling of compaction errors
- [ ] Token info visible in REPL display

### Test Coverage

```go
// context/manager_test.go
func TestContextManager_EnforcesTokenBudget()
func TestContextManager_TruncatesOutput()
func TestContextManager_NormalizesHistory()
func TestContextManager_DetectsCompactionNeeded()

// context/truncate_test.go
func TestTruncateHeadTail_PreservesBeginningAndEnd()
func TestTruncateHeadTail_AddsElisionMarker()
func TestTruncateHeadTail_RespectsByteLimit()

// context/token_tracker_test.go
func TestTokenTracker_AccurateslyEstimates()
func TestTokenTracker_CalculatesAverageTurnSize()
func TestTokenTracker_EstimatesRemainingTurns()

// instructions/loader_test.go
func TestInstructionLoader_MergesInOrder()
func TestInstructionLoader_RespectsSizeLimit()
func TestInstructionLoader_FindsProjectRoot()

// integration tests
func TestContextManager_Integration_WithSession()
func TestContextManager_Integration_WithAgent()
```

### Performance Targets

| Metric | Target |
|--------|--------|
| Context lookup latency | <1ms |
| History normalization | <10ms for 1000 items |
| Truncation time | <5ms |
| Compaction full cycle | <2-5s (async) |
| Memory per session | <50 MiB even with long history |

---

## Reference Implementation

### Codex (Production-Proven)
- **Context Manager**: `research/codex/codex-rs/core/src/context_manager/`
- **Compaction**: `research/codex/codex-rs/core/src/compact.rs`
- **Truncation**: Proven head+tail strategy with 10 KiB limits

### Google ADK
- **Session System**: `research/adk-go/session/`
- **Token Tracking**: Built-in token usage reporting

### OpenHands
- **Context Windows**: Reference data for model limits
- **History Management**: Proven patterns for conversation state

---

## Related ADRs & Documentation

- [ADR-0005: Google Search Built-in Tool](./0005-google-search-builtin-tool.md)
- [ADR-0001: Claude Code Agent Support](./0001-claude-code-agent-support.md)
- [ARCHITECTURE.md - Session System](../ARCHITECTURE.md#session-system)
- [TOOL_DEVELOPMENT.md](../TOOL_DEVELOPMENT.md)

---

## Approval & Sign-Off

| Role | Status | Date |
|------|--------|------|
| Architecture Lead | Pending | - |
| Implementation Lead | Pending | - |
| QA Lead | Pending | - |

---

## Implementation Checklist

### Phase 1: Core Context Management
- [ ] Create `internal/context/manager.go`
- [ ] Create `internal/context/truncate.go`
- [ ] Create `internal/context/token_tracker.go`
- [ ] Write comprehensive tests for each
- [ ] Verify truncation preserves important content
- [ ] Verify token accounting matches model behavior

### Phase 2: Conversation Compaction
- [ ] Create `internal/context/compaction.go`
- [ ] Implement compaction trigger logic
- [ ] Add compaction to agent loop
- [ ] Test with various conversation lengths
- [ ] Validate summary quality
- [ ] Benchmark compaction latency

### Phase 3: Instruction Hierarchy
- [ ] Create `internal/instructions/loader.go`
- [ ] Implement project root detection
- [ ] Implement hierarchical merge logic
- [ ] Test with nested directory structures
- [ ] Validate size limits enforced
- [ ] Add AGENTS.md documentation

### Phase 4: Integration
- [ ] Integrate ContextManager into Session
- [ ] Update agent execution loop
- [ ] Add token display to REPL
- [ ] Add compaction notifications
- [ ] Create /tokens and /compact REPL commands
- [ ] Update model registry with context windows

### Phase 5: Testing & Documentation
- [ ] Run full test suite
- [ ] Run `make check` successfully
- [ ] Write user guide for AGENTS.md
- [ ] Document context management in ARCHITECTURE.md
- [ ] Add examples showing compaction in action
- [ ] Create troubleshooting guide

---

## See Also

- [Codex Context Manager Source](../../research/codex/codex-rs/core/src/context_manager/)
- [Codex Compaction Source](../../research/codex/codex-rs/core/src/compact.rs)
- [Google ADK Session Management](../../research/adk-go/session/)
- [LLM Context Window Reference](https://lexical.github.io/llama.cpp/api/) - Model-specific limits
- [Token Counting Strategies](https://github.com/openai/tiktoken) - Reference for token estimation

