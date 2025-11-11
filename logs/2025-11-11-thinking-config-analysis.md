# Analysis: ThinkingConfig Support for Google ADK Go Coding Agent

**Date**: November 11, 2025  
**Reference**: https://google.github.io/adk-docs/agents/llm-agents/#planner  
**Status**: Feature Available, Implementation Optional

---

## Executive Summary

**ThinkingConfig is ALREADY AVAILABLE** in the ADK Go framework but **NOT YET CONFIGURED** in our coding agent. The feature exists in the `google.golang.org/genai` package (v1.34.0+) and can be enabled by adding configuration to our existing `GenerateContentConfig` setup.

---

## 1. Feature Overview

### What is ThinkingConfig?

ThinkingConfig enables models (specifically Gemini 2.5 Flash and later) to expose their internal reasoning process, allowing:

1. **Transparent reasoning**: See the model's "thoughts" before generating a response
2. **Controlled thinking depth**: Limit token budget for reasoning to balance quality vs. cost
3. **Better debugging**: Understand why the model made certain decisions

### Two Approaches (from Python ADK)

#### A. BuiltInPlanner
- Uses **native model thinking** (Gemini's built-in capability)
- **Configuration**: `ThinkingConfig` with `include_thoughts` and `thinking_budget`
- **Best for**: Gemini 2.5+ models with native thinking support

#### B. PlanReActPlanner
- **Structured output** with explicit sections:
  - `/*PLANNING*/` - Model's plan
  - `/*ACTION*/` - Tool calls
  - `/*REASONING*/` - Why actions were taken
  - `/*FINAL_ANSWER*/` - Final response
- **Best for**: Models without native thinking OR when you need structured reasoning format

---

## 2. Current State of ADK Go

### ✅ What's Available

From `google.golang.org/genai` package (version 1.34.0):

```go
// Type definition from genai package
type ThinkingConfig struct {
    // Optional. Indicates whether to include thoughts in the response.
    IncludeThoughts bool `json:"includeThoughts,omitempty"`
    
    // Optional. Indicates the thinking budget in tokens.
    ThinkingBudget *int32 `json:"thinkingBudget,omitempty"`
}
```

**Usage in GenerateContentConfig**:
```go
type GenerateContentConfig struct {
    // ... other fields ...
    
    // Optional. The thinking features configuration.
    ThinkingConfig *ThinkingConfig `json:"thinkingConfig,omitempty"`
    
    // ... other fields ...
}
```

### 📍 Where It's Used

From ADK Go test file (`agent/llmagent/llmagent_test.go`):

```go
a, err := llmagent.New(llmagent.Config{
    Name:        "calculator",
    Description: "calculating agent",
    Model:       model,
    Instruction: "Think deep. Always double check the answer before making the conclusion.",
    GenerateContentConfig: &genai.GenerateContentConfig{
        ThinkingConfig: &genai.ThinkingConfig{
            IncludeThoughts: true, // Enables thought visibility
        },
    },
})
```

### ❌ What's NOT Available in ADK Go

Unlike Python ADK, the Go implementation does **NOT** have:
- `BuiltInPlanner` type/wrapper
- `PlanReActPlanner` type/wrapper
- High-level planner abstractions

**However**: The underlying capability is available via `ThinkingConfig` in `GenerateContentConfig`.

---

## 3. Current Code Agent Configuration

### What We Have Now

From `code_agent/agent/coding_agent.go` (lines 177-186):

```go
codingAgent, err := llmagent.New(llmagent.Config{
    Name:        "coding_agent",
    Model:       cfg.Model,
    Description: "An expert coding assistant that can read, write, and modify code, execute commands, and solve programming tasks.",
    Instruction: instruction,
    Tools:       registeredTools,
    GenerateContentConfig: &genai.GenerateContentConfig{
        Temperature: genai.Ptr(float32(0.7)),
        // ⚠️ ThinkingConfig NOT CONFIGURED
    },
})
```

### What We're Missing

```go
// NOT currently in our code:
ThinkingConfig: &genai.ThinkingConfig{
    IncludeThoughts: true,
    ThinkingBudget:  genai.Ptr(int32(1024)),
},
```

---

## 4. Should We Add This Feature?

### ✅ Reasons TO Add ThinkingConfig

1. **Better Debugging**: See why the agent chose certain tools or actions
2. **Improved Transparency**: Users can understand the agent's reasoning
3. **Enhanced Learning**: Helps users learn from the agent's thought process
4. **Already Supported**: The infrastructure exists, low implementation cost
5. **Models Support It**: Gemini 2.5 Flash has native thinking capability

### ⚠️ Reasons to Consider Carefully

1. **Token Cost**: Thinking tokens count toward your API quota
2. **Increased Latency**: More tokens = slightly slower responses
3. **UI Complexity**: Need to display/handle thoughts in output
4. **Not All Models Support It**: Only newer Gemini models have native thinking

### 💡 Recommended Approach: **Make It Optional**

Best practice: Add as an **optional configuration parameter** with sensible defaults.

---

## 5. Implementation Recommendation

### Proposed Changes

#### A. Update Config Struct

```go
// In code_agent/agent/coding_agent.go
type Config struct {
    Model                model.LLM
    WorkingDirectory     string
    EnableMultiWorkspace bool
    
    // NEW: Optional thinking configuration
    EnableThinking   bool   // Default: false
    ThinkingBudget   int32  // Default: 1024 tokens (only if EnableThinking=true)
}
```

#### B. Update NewCodingAgent Function

```go
// In NewCodingAgent function
generateConfig := &genai.GenerateContentConfig{
    Temperature: genai.Ptr(float32(0.7)),
}

// Add thinking config if enabled
if cfg.EnableThinking {
    generateConfig.ThinkingConfig = &genai.ThinkingConfig{
        IncludeThoughts: true,
        ThinkingBudget:  genai.Ptr(cfg.ThinkingBudget),
    }
}

codingAgent, err := llmagent.New(llmagent.Config{
    Name:                  "coding_agent",
    Model:                 cfg.Model,
    Description:           "An expert coding assistant...",
    Instruction:           instruction,
    Tools:                 registeredTools,
    GenerateContentConfig: generateConfig,
})
```

#### C. Update CLI Flags

```go
// In code_agent/cli.go or main.go
var (
    enableThinking  = flag.Bool("enable-thinking", false, "Enable model thinking/reasoning output")
    thinkingBudget  = flag.Int("thinking-budget", 1024, "Token budget for thinking (only if --enable-thinking is set)")
)
```

#### D. Update Display Layer

```go
// In code_agent/display/ - handle thoughts in event processing
// Detect and format thinking content from model responses
if event.Content.Parts[0].Text != "" {
    // Check if content contains thoughts
    if isThinkingContent(event.Content.Parts[0].Text) {
        renderer.RenderThinking(event.Content.Parts[0].Text)
    }
}
```

---

## 6. Testing Strategy

### Test Cases to Implement

1. **Without Thinking (Default)**
   ```bash
   ./code-agent
   # Should work as before, no thoughts displayed
   ```

2. **With Thinking Enabled**
   ```bash
   ./code-agent --enable-thinking
   # Should show model's reasoning process
   ```

3. **Custom Thinking Budget**
   ```bash
   ./code-agent --enable-thinking --thinking-budget 512
   # Limits thinking to 512 tokens
   ```

4. **Model Compatibility**
   - Test with Gemini 2.5 Flash (supports native thinking)
   - Test with older models (should gracefully ignore or error)

### Example Expected Output (With Thinking)

```
User: Fix the error in main.go

🧠 Agent is thinking...
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Thought Process:
1. Need to read main.go to understand the error
2. Likely a compilation error based on context
3. Should use read_file tool first
4. Then analyze and propose fix
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔧 Reading main.go
✓ Tool completed: read_file

🧠 Reasoning: Found syntax error on line 45 - missing closing brace

🔧 Applying fix...
✓ Tool completed: replace_in_file

✓ Fixed error in main.go - added missing brace on line 45
```

---

## 7. Comparison: Python ADK vs Go ADK

| Feature | Python ADK | Go ADK | Notes |
|---------|------------|--------|-------|
| ThinkingConfig | ✅ Yes | ✅ Yes | Core config available |
| BuiltInPlanner | ✅ Yes | ❌ No | Python has wrapper class |
| PlanReActPlanner | ✅ Yes | ❌ No | Python has wrapper class |
| Native Thinking | ✅ Yes | ✅ Yes | Via GenerateContentConfig |
| Token Budget | ✅ Yes | ✅ Yes | Both support setting limits |

**Conclusion**: Go ADK has the **core capability** but lacks the **convenience wrappers** that Python provides. We can still achieve the same result with direct configuration.

---

## 8. Token Cost Analysis

### Thinking Budget Impact

| Budget | Use Case | Typical Cost |
|--------|----------|--------------|
| 0 (disabled) | Standard operation | Baseline |
| 256 tokens | Quick reasoning | +$0.0001 per request* |
| 512 tokens | Moderate thinking | +$0.0002 per request* |
| 1024 tokens | Deep reasoning | +$0.0004 per request* |
| 2048 tokens | Complex problems | +$0.0008 per request* |

*Approximate, based on Gemini 2.5 Flash pricing ($0.000375/1K input tokens)

### Cost vs. Value Tradeoff

**When Thinking Tokens Are Worth It**:
- Complex debugging scenarios
- Multi-step code refactoring
- Learning/educational use cases
- Production issue diagnosis

**When to Skip Thinking**:
- Simple file operations
- Repetitive tasks
- Batch processing
- Cost-sensitive applications

---

## 9. Alternative: PlanReActPlanner-Style Output (Manual Implementation)

Since Go ADK doesn't have `PlanReActPlanner`, we could **manually structure the system prompt** to achieve similar output:

### Enhanced System Prompt (Optional)

```go
instruction := BuildEnhancedPromptWithContext(registry, promptCtx) + `

## Reasoning Structure

When solving complex problems, structure your response as follows:

1. **Planning Phase**: Outline your approach
   - What tools will you use?
   - What's the sequence of operations?
   - What are potential issues?

2. **Execution Phase**: Take actions
   - Use tools as planned
   - Adapt if unexpected results occur
   
3. **Reasoning Phase**: Explain decisions
   - Why did you choose these tools?
   - What did you learn from results?
   
4. **Final Answer**: Provide solution
   - Summarize what was done
   - Confirm success or explain issues
`
```

**Pros**: Works with any model, no API changes needed  
**Cons**: Less structured than native thinking, depends on prompt following

---

## 10. Recommendations

### Immediate Action (Low-Hanging Fruit)

✅ **YES - Add ThinkingConfig Support**

**Why**:
1. Infrastructure already exists (genai.ThinkingConfig)
2. Low implementation cost (~50 lines of code)
3. High value for debugging and transparency
4. Models like Gemini 2.5 Flash already support it
5. Can be optional/disabled by default

**Implementation Priority**: **MEDIUM**
- Not critical for core functionality
- But valuable for user experience and debugging

### Suggested Implementation Order

1. **Phase 1**: Add config flags and basic support (1-2 hours)
   - Add `EnableThinking` and `ThinkingBudget` to Config
   - Wire up to GenerateContentConfig
   - Add CLI flags

2. **Phase 2**: Enhance display layer (2-3 hours)
   - Detect thinking/reasoning in output
   - Format thoughts nicely in terminal
   - Add clear visual separation

3. **Phase 3**: Documentation and examples (1 hour)
   - Update README with thinking config usage
   - Add example use cases
   - Document token cost implications

**Total Effort**: ~4-6 hours

### Future Enhancements (Optional)

1. **Dynamic Thinking Budget**: Adjust based on task complexity
2. **Thinking Analytics**: Track how often thinking improves outcomes
3. **Structured Output Parsing**: Extract structured reasoning from thoughts
4. **UI Toggle**: Allow users to show/hide thoughts on demand

---

## 11. Conclusion

### Current State
- ✅ ThinkingConfig **IS AVAILABLE** in ADK Go
- ❌ We are **NOT USING IT** in our coding agent
- ⚠️ Go ADK lacks high-level planner wrappers (unlike Python)

### Recommendation
**IMPLEMENT with the following approach**:

```go
// Default: Disabled (backward compatible)
Config{
    EnableThinking: false,  // User must opt-in
    ThinkingBudget: 1024,   // Reasonable default
}

// When enabled via CLI:
./code-agent --enable-thinking --thinking-budget 1024
```

### Benefits
1. **Better debugging** for complex tasks
2. **Educational value** for users learning from the agent
3. **Transparency** in agent decision-making
4. **Optional** - doesn't impact users who don't need it
5. **Low cost** - small token overhead when enabled

### Next Steps
1. Update `Config` struct in `coding_agent.go`
2. Add CLI flags for thinking configuration
3. Wire up ThinkingConfig in GenerateContentConfig
4. Enhance display layer to show thoughts
5. Add documentation and examples
6. Test with Gemini 2.5 Flash

---

## References

- **ADK Docs**: https://google.github.io/adk-docs/agents/llm-agents/#planner
- **Gemini Thinking**: https://ai.google.dev/gemini-api/docs/thinking
- **genai Package**: https://pkg.go.dev/google.golang.org/genai
- **ADK Go Tests**: `google/adk-go/agent/llmagent/llmagent_test.go`

---

**Status**: Analysis complete. Ready for implementation if approved.
