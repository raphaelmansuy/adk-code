# LLM Provider Abstraction: Multi-Provider Support

**⚠️ PROPOSED FEATURE**: Provider abstraction does not currently exist in code_agent. The agent is tightly coupled to Gemini 2.5 Flash (hardcoded in main.go, line ~70). This document describes a proposed abstraction layer based on DeepCode patterns.

## Introduction

**Current Limitation**: code_agent works only with Gemini via Google's genai client (hardcoded model selection). This creates lock-in: can't use Claude, GPT-4, local models, or cheaper alternatives.

**DeepCode Pattern**: Abstraction layer supporting any provider (Claude, OpenAI, local models, custom servers).

**Benefit**: Switch models without code changes, use cheapest capable model, fallback to alternatives on failure.

---

## Architecture: Provider Interface

### Core Abstraction

```go
type LLMProvider interface {
    // Core inference
    Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)
    
    // Streaming (for real-time output)
    StreamComplete(ctx context.Context, req *CompletionRequest) (<-chan string, error)
    
    // Model info
    GetCapabilities() ProviderCapabilities
    GetCost() CostInfo
    GetRateLimit() RateLimit
    IsAvailable() bool
}

type CompletionRequest struct {
    Model           string
    SystemPrompt    string
    Messages        []Message
    Temperature     float64
    MaxTokens       int
    TopP            float64
}

type CompletionResponse struct {
    Content       string
    TokensUsed    TokenUsage
    StopReason    string
    Model         string
    Provider      string
}

type ProviderCapabilities struct {
    MaxContextLength int
    SupportsStreaming bool
    SupportsImages   bool
    SupportsTools    bool
    SupportsFunctionCalling bool
    MaxRequestsPerMinute int
}

type CostInfo struct {
    PricePerMillionInputTokens  float64
    PricePerMillionOutputTokens float64
    CostPerRequest              float64
}
```

### Provider Implementations

**Gemini Provider** (existing):
```go
type GeminiProvider struct {
    client *genai.Client
    model  string
}

func (g *GeminiProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
    // Translate abstract request to Gemini API
    // Call g.client.GenerativeModel(g.model).GenerateContent()
    // Translate response back
}
```

**Claude Provider** (new):
```go
type ClaudeProvider struct {
    client *anthropic.Client
    model  string
}

func (c *ClaudeProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
    // Translate to Claude API
    // Call c.client.Messages.New()
}
```

**OpenAI Provider** (new):
```go
type OpenAIProvider struct {
    client *openai.Client
    model  string
}

func (o *OpenAIProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
    // Translate to OpenAI API
    // Call o.client.CreateChatCompletion()
}
```

**Local/Custom Provider** (new):
```go
type LocalProvider struct {
    endpoint string // e.g., http://localhost:8000/v1
    client   *http.Client
}

func (l *LocalProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
    // Call local LLM server via HTTP
}
```

---

## Provider Manager

```go
type ProviderManager struct {
    providers     map[string]LLMProvider
    defaultModel  string
    fallbacks     []string
    costOptimizer *CostOptimizer
}

// Initialize with multiple providers
func NewProviderManager(config ProviderConfig) *ProviderManager {
    pm := &ProviderManager{
        providers: make(map[string]LLMProvider),
    }
    
    if config.Gemini.Enabled {
        pm.providers["gemini"] = NewGeminiProvider(config.Gemini)
    }
    if config.Claude.Enabled {
        pm.providers["claude"] = NewClaudeProvider(config.Claude)
    }
    if config.OpenAI.Enabled {
        pm.providers["openai"] = NewOpenAIProvider(config.OpenAI)
    }
    if config.Local.Enabled {
        pm.providers["local"] = NewLocalProvider(config.Local)
    }
    
    pm.defaultModel = config.DefaultModel
    pm.fallbacks = config.Fallbacks
    
    return pm
}

// Smart provider selection
func (pm *ProviderManager) SelectProvider(task Task) LLMProvider {
    switch task.Type {
    case "critical_code_generation":
        // Use most capable model
        return pm.providers["claude"]
    case "code_search":
        // Use cheapest fast model
        return pm.providers["openai"]
    case "validation":
        // Use fast model
        return pm.providers["local"]
    default:
        return pm.providers[pm.defaultModel]
    }
}

// With fallback logic
func (pm *ProviderManager) CompleteWithFallback(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
    providers := []string{pm.defaultModel}
    providers = append(providers, pm.fallbacks...)
    
    var lastErr error
    for _, providerName := range providers {
        provider := pm.providers[providerName]
        if !provider.IsAvailable() {
            continue
        }
        
        resp, err := provider.Complete(ctx, req)
        if err == nil {
            return resp, nil
        }
        
        lastErr = err
        // Continue to next provider
    }
    
    return nil, fmt.Errorf("all providers failed: %v", lastErr)
}
```

---

## Cost Optimization

### Intelligent Routing

```go
type CostOptimizer struct {
    costThreshold float64 // Max acceptable cost per request
    budget        Budget  // Total budget tracking
}

type Budget struct {
    TotalSpent      float64
    BudgetLimit     float64
    RequestsCount   int
    LastResetTime   time.Time
}

// Route task to cheapest capable provider
func (co *CostOptimizer) RouteToCheapest(task Task, providers []LLMProvider) LLMProvider {
    capable := filterCapable(providers, task)
    
    var cheapest LLMProvider
    minCost := math.MaxFloat64
    
    for _, provider := range capable {
        cost := provider.GetCost()
        if cost.CostPerRequest < minCost {
            minCost = cost.CostPerRequest
            cheapest = provider
        }
    }
    
    // Check budget
    if co.budget.TotalSpent+minCost > co.budget.BudgetLimit {
        // Use free tier or error
        return nil
    }
    
    return cheapest
}

// Example usage:
// 
// Task: Simple code analysis → Use GPT-4o mini ($0.00015 per request)
// Task: Complex refactoring → Use Claude 3.5 ($0.003 per request)
// Task: Validation → Use Ollama local ($0 cost)
```

### Cost Tracking

```go
func (pm *ProviderManager) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
    provider := pm.SelectProvider(req)
    resp, err := provider.Complete(ctx, req)
    
    if err == nil {
        // Track cost
        cost := provider.GetCost()
        estimatedCost := float64(resp.TokensUsed.InputTokens) * cost.PricePerMillionInputTokens / 1e6
        estimatedCost += float64(resp.TokensUsed.OutputTokens) * cost.PricePerMillionOutputTokens / 1e6
        
        pm.costOptimizer.budget.TotalSpent += estimatedCost
        log.Printf("Request cost: $%.4f, Total: $%.2f", estimatedCost, pm.costOptimizer.budget.TotalSpent)
    }
    
    return resp, err
}
```

---

## Configuration

### YAML Configuration Example

```yaml
llm_providers:
  default_model: "claude"  # Primary provider
  fallbacks: ["openai", "gemini", "local"]
  
  providers:
    gemini:
      enabled: true
      api_key: ${GOOGLE_API_KEY}
      model: "gemini-2.5-flash"
      capabilities:
        max_context: 1000000
        streaming: true
    
    claude:
      enabled: true
      api_key: ${ANTHROPIC_API_KEY}
      model: "claude-3-5-sonnet-20241022"
      capabilities:
        max_context: 200000
        streaming: true
        vision: true
    
    openai:
      enabled: true
      api_key: ${OPENAI_API_KEY}
      model: "gpt-4o-mini"
      capabilities:
        max_context: 128000
        streaming: true
      cost_optimization: true
    
    local:
      enabled: false
      endpoint: "http://localhost:8000/v1"
      model: "llama-2-70b"
      capabilities:
        max_context: 4096
        streaming: true
        free: true
  
  routing_rules:
    - task_type: "critical_code_generation"
      provider: "claude"
      reason: "best_quality"
    
    - task_type: "code_search"
      provider: "openai"
      reason: "cost_optimized"
    
    - task_type: "validation"
      provider: "local"
      reason: "free_available"
    
    - task_type: "streaming_output"
      provider: "claude"
      reason: "streaming_capable"
  
  cost_management:
    budget_limit: 100.0  # $100 per day
    prefer_cheaper: true  # Choose cheaper capable model
    track_per_task: true
```

---

## Migration from Gemini-only

### Phase 1: Add Abstraction (Non-Breaking)

```go
// Keep existing Gemini as default
func (agent *CodingAgent) Complete(req CompletionRequest) (*CompletionResponse, error) {
    // Old code: Call Gemini directly
    // New code: Use provider manager
    return agent.providerManager.Complete(ctx, req)
}
```

### Phase 2: Add Alternative Providers

```go
// Initialize with multiple providers
providerManager := NewProviderManager(config)
// Gemini + Claude + OpenAI all available
```

### Phase 3: Smart Routing

```go
// Route based on task
provider := providerManager.SelectProvider(task)
resp, err := provider.Complete(ctx, req)
```

### Phase 4: Full Flexibility

```go
// Agent can switch providers mid-session
// Prompt engineering becomes provider-agnostic
// Fallback chains handle failures
```

---

## Handling Provider Differences

### API Differences

**Problem**: Different providers have different APIs

**Solution**: Translate at request/response boundary

```go
func (c *ClaudeProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
    // Translate abstract request to Claude format
    claudeReq := &anthropic.MessageParam{
        Model:       c.model,
        MaxTokens:   req.MaxTokens,
        Temperature: req.Temperature,
        SystemPrompt: req.SystemPrompt,
        Messages:    translateMessages(req.Messages),
    }
    
    // Call Claude
    resp, err := c.client.Messages.New(ctx, claudeReq)
    
    // Translate response back to abstract format
    return &CompletionResponse{
        Content: resp.Content[0].Text,
        TokensUsed: TokenUsage{
            InputTokens:  resp.Usage.InputTokens,
            OutputTokens: resp.Usage.OutputTokens,
        },
        Provider: "claude",
    }, nil
}
```

### Capability Differences

**Problem**: Not all models support all features

**Solution**: Capability checking

```go
func (pm *ProviderManager) CanHandleTask(task Task) bool {
    provider := pm.providers[pm.defaultModel]
    caps := provider.GetCapabilities()
    
    switch task.Type {
    case "vision_analysis":
        return caps.SupportsImages
    case "function_calling":
        return caps.SupportsFunctionCalling
    case "streaming":
        return caps.SupportsStreaming
    default:
        return true
    }
}
```

---

## Benefits

| Aspect | Gemini-Only | Multi-Provider |
|--------|------------|-----------------|
| **Cost** | Fixed, expensive | Optimize per task |
| **Reliability** | Single point of failure | Fallback chains |
| **Flexibility** | None | Choose best for task |
| **Quality** | One model | Use best for type |
| **Experimentation** | Slow (API changes) | Fast (switch providers) |
| **Compliance** | Limited | Choose compliant provider |

---

## Next Steps

1. **[03-multi-agent-orchestration.md](03-multi-agent-orchestration.md)** - Agent design
2. **[06-prompt-engineering-advanced.md](06-prompt-engineering-advanced.md)** - Provider-agnostic prompts
3. **[07-implementation-roadmap.md](07-implementation-roadmap.md)** - Implementation

---

## References

- **DeepCode provider selection**: `/research/DeepCode/utils/llm_utils.py`
- **Augmented LLM classes**: `/research/DeepCode/mcp_agent/workflows/llm/`
- **ADK provider abstraction**: `/research/adk-go/model/` (genai client foundation)
