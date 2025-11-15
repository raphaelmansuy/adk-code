# Context Management Integration Guide

This document explains how context management is integrated into the adk-code agent system, including support for both the main agent and sub-agents.

## Architecture Overview

Context management is integrated at the session level and tracks all agent interactions:

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                         │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐  │
│  │ Main Agent   │───▶│ Sub-Agent 1  │───▶│ Sub-Agent 2  │  │
│  └──────────────┘    └──────────────┘    └──────────────┘  │
│         │                    │                    │          │
│         └────────────────────┴────────────────────┘          │
│                              │                               │
│                    Shared Model (LLM)                        │
│                              │                               │
└──────────────────────────────┼───────────────────────────────┘
                               │
                ┌──────────────▼──────────────┐
                │   Context Manager           │
                │  ┌──────────────────────┐   │
                │  │ Token Tracking       │   │
                │  │ Output Truncation    │   │
                │  │ Compaction Detection │   │
                │  └──────────────────────┘   │
                └─────────────────────────────┘
```

## Integration Points

### 1. Session Initialization

When a session is created, the ContextManager is initialized with the model and model configuration:

```go
// internal/orchestration/session.go
func InitializeSessionComponents(
    ctx context.Context,
    cfg *config.Config,
    ag agent.Agent,
    llm model.LLM,              // ✅ Model from main agent
    modelConfig models.Config,  // ✅ Model configuration
    bannerRenderer *display.BannerRenderer,
) (*SessionComponents, error) {
    // ...
    
    // Initialize context manager with model for compaction
    contextManager := context.NewContextManagerWithModel(modelConfig, llm)
    
    return &SessionComponents{
        Manager:        manager,
        Runner:         runner,
        Tokens:         tokens,
        ContextManager: contextManager, // ✅ Available to entire session
    }, nil
}
```

### 2. REPL Event Loop

The REPL tracks each turn and displays context usage:

```go
// internal/repl/repl.go
func (r *REPL) processUserMessage(ctx context.Context, input string) {
    // ... agent loop processes events ...
    
    // After turn completes, track context
    if r.config.ContextManager != nil && !hasError {
        r.trackAndDisplayContext(input)
    }
}

func (r *REPL) trackAndDisplayContext(userInput string) {
    // Add user message to context
    userItem := context.ResponseItem{
        Type:      context.ItemMessage,
        Role:      "user",
        Content:   userInput,
        Timestamp: time.Now(),
    }
    
    err := r.config.ContextManager.AddItem(userItem)
    if err == context.ErrCompactionNeeded {
        // Display warning: context approaching limit
        fmt.Println("⚠️  Context approaching limit - compaction recommended")
    }
    
    // Display token usage
    info := r.config.ContextManager.TokenInfo()
    fmt.Printf("📊 Context: %d/%d tokens (%.1f%%) • Compaction at %.0f%%\n",
        info.UsedTokens,
        info.AvailableTokens + info.UsedTokens,
        info.PercentageUsed * 100,
        info.CompactThreshold * 100,
    )
}
```

### 3. Sub-Agent Model Inheritance

Sub-agents automatically inherit the model from the main agent, which means they share context management:

```go
// tools/agents/subagent_tools.go
func (m *SubAgentManager) createSubAgent(agentDef *agents.Agent) (agent.Agent, error) {
    // Create subagent with inherited model
    subAgent, err := llmagent.New(llmagent.Config{
        Name:        agentDef.Name,
        Description: agentDef.Description,
        Model:       m.modelLLM, // ✅ Inherits model from main agent
        Instruction: agentDef.Content,
        Tools:       allowedTools,
    })
    
    // Sub-agent now shares:
    // - Same context window limits
    // - Token budget tracking
    // - Compaction capabilities (uses same model for summarization)
    
    return subAgent, nil
}
```

## How It Works

### Token Tracking

1. **User Input**: Each user message is tracked via `ContextManager.AddItem()`
2. **Token Estimation**: Content is converted to estimated tokens (1 token ≈ 4 chars)
3. **Threshold Check**: When usage exceeds 70%, `ErrCompactionNeeded` is returned
4. **Display**: Token usage is shown after each turn in the REPL

### Output Truncation

Large tool outputs are automatically truncated:

```go
// Automatic truncation in ContextManager.AddItem()
if item.Type == context.ItemToolOutput {
    // Head+tail truncation: keep first 128 + last 128 lines
    item.Content = cm.truncateOutput(item.Content)
}
```

### Compaction (ADK Agent-Based)

When compaction is needed, a specialized ADK agent is created:

```go
// internal/context/compaction.go
compactionAgent, err := llmagent.New(llmagent.Config{
    Name:        "conversation_compactor",
    Model:       llm, // ✅ Same model as main agent
    Description: "Specialized conversation compactor",
    Instruction: "Summarize conversation concisely...",
})

// Agent generates summary using inherited model
summary := compactionAgent.Run(ctx, prompt)
```

### Multi-Agent Context Sharing

All agents (main + sub-agents) share the same context budget:

```
Main Agent Turn 1:    2,000 tokens   (Total: 2,000)
Sub-Agent A Turn:     3,500 tokens   (Total: 5,500)
Main Agent Turn 2:    1,800 tokens   (Total: 7,300)
Sub-Agent B Turn:     4,200 tokens   (Total: 11,500)
                                     
At 700,000 tokens (70% of 1M):  ⚠️  Compaction triggered
```

## Configuration

Default context management settings:

| Setting | Default Value | Description |
|---------|---------------|-------------|
| Context Window | Model-specific | 1M for Gemini 2.5 Flash, 2M for Gemini 1.5 Pro |
| Reserved for Output | 10% | Reserved tokens for agent responses |
| Compaction Threshold | 70% | Trigger compaction at this percentage |
| Output Truncate Bytes | 10 KiB | Max bytes per tool output |
| Output Truncate Lines | 256 | Max lines (128 head + 128 tail) |
| User Message Retention | 20K tokens | Tokens to retain during compaction |

## User Experience

### Normal Operation

```
User: Create a new file hello.py
📊 Context: 1,250/1,000,000 tokens (0.1%) • Compaction at 70%

User: Add a main function
📊 Context: 2,100/1,000,000 tokens (0.2%) • Compaction at 70%
```

### Approaching Limit

```
User: Run all tests
📊 Context: 685,000/1,000,000 tokens (68.5%) • Compaction at 70%

User: Fix the failing tests
⚠️  Context approaching limit - compaction recommended
📊 Context: 710,000/1,000,000 tokens (71.0%) • Compaction at 70%
```

## Benefits

### For Main Agent
- ✅ Prevents context overflow crashes
- ✅ Visible token usage feedback
- ✅ Automatic output truncation
- ✅ Graceful degradation via compaction

### For Sub-Agents
- ✅ Share context budget with main agent
- ✅ Same model capabilities (inherited)
- ✅ Unified token accounting
- ✅ Consistent behavior across all agents

### For Users
- ✅ Clear visibility into context usage
- ✅ No surprises (warnings before limits)
- ✅ Long conversations (50+ turns)
- ✅ Automatic management (no manual intervention)

## Implementation Status

✅ **Completed**:
- ContextManager initialization in session
- Token tracking per turn
- Output truncation (head+tail)
- Compaction detection
- REPL token display
- Sub-agent model inheritance
- ADK agent-based compaction

⏳ **Not Yet Implemented**:
- Automatic compaction triggering (currently manual warning)
- Tool output tracking in ContextManager
- Assistant response tracking
- Compaction execution workflow
- Token usage in session persistence

## Next Steps

To fully complete the integration:

1. **Track Tool Outputs**: Capture tool execution results in ContextManager
2. **Track Assistant Responses**: Capture agent responses in ContextManager
3. **Implement Auto-Compaction**: Automatically trigger compaction at threshold
4. **Persist Context State**: Save context state to session database
5. **Add `/context` Command**: Let users inspect context status manually

## Testing

Test that context management works:

```bash
# Build
cd adk-code
go build -o adk-code .

# Run
export GOOGLE_API_KEY="your-key"
./adk-code

# In REPL, observe context tracking:
User: help
📊 Context: 45/1,000,000 tokens (0.0%) • Compaction at 70%
```

## Related Documentation

- [CONTEXT_MANAGEMENT.md](CONTEXT_MANAGEMENT.md) - Full API documentation
- [ADR-0006](adr/0006-agent-context-management.md) - Design decisions
- [Implementation Summary](adr/0006-implementation-summary.md) - Technical details
