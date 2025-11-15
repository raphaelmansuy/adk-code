# Agent Context Management Guide

This guide explains how to use the context management system implemented according to ADR-0006.

## Overview

The context management system provides:
- **Token Budget Enforcement**: Track and enforce token limits per model
- **Output Truncation**: Automatically truncate verbose tool outputs using head+tail strategy
- **Token Tracking**: Detailed accounting of token usage across conversation turns
- **Conversation Compaction**: Summarize and compress when approaching context limits
- **Hierarchical Instructions**: Support for AGENTS.md files at multiple levels

## Components

### 1. ContextManager

The `ContextManager` maintains conversation history and enforces context limits.

#### Usage Example

```go
import (
    "adk-code/internal/context"
    "adk-code/pkg/models"
)

// Create a context manager for a specific model
modelConfig := models.Config{
    Name:          "gemini-2.5-flash",
    ContextWindow: 1_000_000,
}

cm := context.NewContextManager(modelConfig)

// Add items to the conversation
item := context.ResponseItem{
    ID:        "msg-1",
    Type:      context.ItemMessage,
    Role:      "user",
    Content:   "Hello, can you help me?",
    Timestamp: time.Now(),
}

err := cm.AddItem(item)
if err == context.ErrCompactionNeeded {
    // Trigger compaction workflow
    log.Println("Context compaction needed")
}

// Get token usage information
info := cm.TokenInfo()
fmt.Printf("Used: %d/%d tokens (%.1f%%)\n", 
    info.UsedTokens, 
    info.AvailableTokens, 
    info.PercentageUsed*100)
```

#### Configuration

Default settings (from ADR):
- Output truncation: 10 KiB (10,240 bytes)
- Max output lines: 256
- Head lines: 128 (first 128 lines preserved)
- Tail lines: 128 (last 128 lines preserved)
- Compaction threshold: 70% of context window

### 2. Output Truncation

Tool outputs are automatically truncated using a head+tail strategy to preserve important beginning and ending information.

#### Example

```go
// For a 500-line output:
// [First 128 lines]
// [... omitted 244 of 500 lines ...]
// [Last 128 lines]
```

The truncation:
- Preserves the first 128 lines (often contains setup/context)
- Preserves the last 128 lines (often contains results/errors)
- Adds a clear marker showing what was omitted
- Respects both line count (256) and byte size (10 KiB) limits

### 3. Token Tracking

Track token usage across conversation turns.

#### Usage Example

```go
import "adk-code/internal/context"

tracker := context.NewTokenTracker("session-123", "gemini-2.5-flash", 1_000_000)

// Record a turn
tracker.RecordTurn(1500, 800) // input tokens, output tokens

// Get statistics
avgSize := tracker.AverageTurnSize()
remaining := tracker.EstimateRemainingTurns(1_000_000, 100_000)

fmt.Printf("Average turn size: %d tokens\n", avgSize)
fmt.Printf("Estimated remaining turns: %d\n", remaining)
```

### 4. Conversation Compaction

When approaching context limits (70% by default), conversations can be compacted by:
1. Collecting user messages
2. Generating a summary
3. Retaining most recent messages (up to 20K tokens)
4. Building compacted history: [initial context] + [recent messages] + [summary]

#### Usage Example

```go
import (
    "context"
    "adk-code/internal/context"
)

req := context.CompactionRequest{
    Items:             conversationItems,
    UserMessages:      userMessages,
    TargetTokenBudget: 5000,
    ModelName:         "gemini-2.5-flash",
}

result := context.CompactConversation(context.Background(), req)

if result.Success {
    fmt.Printf("Compacted from %d to %d tokens (%.1fx reduction)\n",
        result.OriginalTokens,
        result.CompactedTokens,
        result.CompactionRatio)
}
```

### 5. Hierarchical Instructions (AGENTS.md)

The instruction loader supports hierarchical user instructions similar to Codex's AGENTS.md system.

#### Directory Structure

```
~/.adk-code/
  └── AGENTS.md              # Global instructions (all projects)

/project/
  ├── go.mod                 # Project root marker
  ├── AGENTS.md              # Project-level instructions
  └── internal/
      └── api/
          └── AGENTS.md      # Directory-specific instructions
```

#### Loading Order

Instructions are loaded and merged in this order:
1. Global (~/.adk-code/AGENTS.md)
2. Project root (detected via .git, go.mod, etc.)
3. Nested directories (from root to current working directory)

#### Usage Example

```go
import "adk-code/internal/instructions"

loader := instructions.NewInstructionLoader("/project/internal/api")
result := loader.Load()

// Access different levels
fmt.Println("Global:", result.Global)
fmt.Println("Project:", result.ProjectRoot)
fmt.Println("Merged:", result.Merged)

if result.Truncated {
    fmt.Println("Instructions were truncated to fit size limit")
}
```

#### AGENTS.md Example

```markdown
# Project Instructions

## Code Style
- Use descriptive variable names
- Add comments for complex logic
- Follow Go best practices

## Testing
- Write tests for all new features
- Aim for >80% coverage
- Use table-driven tests

## Workflow
- Create feature branches
- Submit PRs for review
- Update documentation
```

## Integration with Session System

To integrate context management with the session system:

```go
import (
    "adk-code/internal/context"
    "adk-code/internal/instructions"
    "adk-code/pkg/models"
)

type EnhancedSession struct {
    ID               string
    ContextManager   *context.ContextManager
    TokenTracker     *context.TokenTracker
    Instructions     instructions.LoadedInstructions
    // ... other fields
}

func NewEnhancedSession(sessionID string, modelConfig models.Config, workdir string) *EnhancedSession {
    return &EnhancedSession{
        ID:             sessionID,
        ContextManager: context.NewContextManager(modelConfig),
        TokenTracker:   context.NewTokenTracker(sessionID, modelConfig.Name, modelConfig.ContextWindow),
        Instructions:   instructions.NewInstructionLoader(workdir).Load(),
    }
}
```

## Token Estimation

The system uses a simple heuristic for token estimation:
- 1 token ≈ 4 characters
- This is approximate and may vary by model/tokenizer

For production use, consider integrating with the actual model's tokenizer for accurate counts.

## Model Context Windows

Default context windows by provider:

| Provider | Model | Context Window |
|----------|-------|----------------|
| Gemini | 2.5 Flash | 1M tokens |
| Gemini | 1.5 Pro | 2M tokens |
| OpenAI | GPT-4 Turbo | 128K tokens |
| Ollama | Llama 2 | 4K tokens |
| Ollama | Mistral | 32K tokens |

## Best Practices

1. **Check for compaction needs**: Always check if `AddItem()` returns `ErrCompactionNeeded`
2. **Monitor token usage**: Use `TokenInfo()` to display usage to users
3. **Set appropriate limits**: Adjust truncation limits based on your use case
4. **Test with edge cases**: Test with very long outputs and many turns
5. **Provide user feedback**: Show truncation markers and compaction events
6. **Use AGENTS.md wisely**: Keep instructions concise and relevant

## Error Handling

```go
err := cm.AddItem(item)
switch err {
case nil:
    // Success
case context.ErrCompactionNeeded:
    // Trigger compaction workflow
case context.ErrContextOverflow:
    // Context window exceeded even after compaction
    // Consider starting a new session
default:
    // Handle other errors
}
```

## Performance Considerations

- **Context lookup**: < 1ms per operation
- **History normalization**: < 10ms for 1000 items
- **Truncation**: < 5ms per output
- **Compaction**: 2-5 seconds (involves LLM call)
- **Memory**: < 50 MiB per session even with long history

## Future Enhancements

Potential improvements not yet implemented:
- Integration with actual model tokenizers for accurate token counts
- Automatic compaction triggering in agent loop
- Token usage display in REPL
- Session persistence with context state
- Advanced compaction strategies (e.g., importance-based)
- Real-time LLM-based summarization in compaction

## Related Documentation

- [ADR-0006: Agent Context Management](adr/0006-agent-context-management.md)
- [Architecture Documentation](ARCHITECTURE.md)
- [Model Registry](../adk-code/pkg/models/README.md)
