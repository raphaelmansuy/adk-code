# Vertex AI vs Gemini API: Architecture Comparison

## Side-by-Side Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        code_agent/main.go                   │
│  User → CLI flags/env vars → Backend Detection Logic        │
└─────────────────────────┬───────────────────────────────────┘
                          │
                ┌─────────┴──────────┐
                ▼                    ▼
        ┌──────────────┐      ┌──────────────────┐
        │ Gemini API   │      │   Vertex AI      │
        │   Branch     │      │     Branch       │
        └──────┬───────┘      └────────┬─────────┘
               │                       │
               ▼                       ▼
    ┌────────────────────┐  ┌──────────────────────┐
    │ model/gemini/      │  │ model/vertexai/      │
    │ gemini.go          │  │ vertexai.go (NEW)    │
    │ ✓ Existing         │  │ ✓ To be created      │
    └────────┬───────────┘  └──────────┬───────────┘
             │                         │
             ▼                         ▼
    ┌──────────────────────────────────────────┐
    │   google.golang.org/genai SDK            │
    │                                          │
    │  genai.NewClient(ctx, &ClientConfig{   │
    │    Backend: genai.BackendGeminiAPI,  ✓ │
    │    APIKey: "AIza...",                   │
    │  })                                      │
    └──────────────────────────────────────────┘
             ▲
             │
    ┌──────────────────────────────────────────┐
    │  google.golang.org/genai SDK             │
    │                                          │
    │  genai.NewClient(ctx, &ClientConfig{   │
    │    Backend: genai.BackendVertexAI,   ✓ │
    │    Project: "my-project",               │
    │    Location: "us-central1",             │
    │  })                                      │
    └──────────────────────────────────────────┘
             │
             ▼
    ┌────────────────────┐  ┌──────────────────────┐
    │ Gemini Models:     │  │ Vertex AI Models:    │
    │ • gemini-2.5-flash │  │ • gemini-2.5-flash   │
    │ • gemini-2.0       │  │ • gemini-2.0         │
    │ • gemini-1.5-pro   │  │ • gemini-1.5-pro     │
    │                    │  │ • claude-3-sonnet    │
    └────────────────────┘  └──────────────────────┘
```

## Code Comparison Matrix

| Aspect | Gemini API | Vertex AI |
|--------|-----------|----------|
| **Package Location** | `research/adk-go/model/gemini/gemini.go` | `research/adk-go/model/vertexai/vertexai.go` |
| **Type Name** | `geminiModel struct` | `vertexAIModel struct` |
| **Constructor** | `func NewModel(ctx, modelName, *genai.ClientConfig)` | `func NewModel(ctx, modelName, *genai.ClientConfig)` |
| **Backend Flag** | `genai.BackendGeminiAPI` | `genai.BackendVertexAI` |
| **Authentication** | `cfg.APIKey = "AIza..."` | `cfg.Project, cfg.Location + ADC` |
| **Error Messages** | "Gemini" in strings | "Vertex AI" in strings |
| **Implementation** | 90% identical | 90% identical |

## Initialization Flow Comparison

### Gemini API Flow

```go
func main() {
    apiKey := os.Getenv("GOOGLE_API_KEY")  // 1. Get API key
    
    // 2. Create client
    model, err := gemini.NewModel(ctx, "gemini-2.5-flash", 
        &genai.ClientConfig{
            APIKey: apiKey,
            Backend: genai.BackendGeminiAPI,
        })
    
    // 3. Use in agent
    codingAgent, _ := codingagent.NewCodingAgent(ctx, 
        codingagent.Config{Model: model})
}
```

### Vertex AI Flow (Proposed)

```go
func main() {
    project := os.Getenv("GOOGLE_CLOUD_PROJECT")      // 1. Get project
    location := os.Getenv("GOOGLE_CLOUD_LOCATION")    // 2. Get location
    // ADC (gcloud credentials) used automatically
    
    // 3. Create client
    model, err := vertexai.NewModel(ctx, "gemini-2.5-flash",
        &genai.ClientConfig{
            Project:  project,
            Location: location,
            Backend:  genai.BackendVertexAI,
        })
    
    // 4. Use in agent (identical)
    codingAgent, _ := codingagent.NewCodingAgent(ctx,
        codingagent.Config{Model: model})
}
```

## Implementation Differences

### File: `model/gemini/gemini.go` vs `model/vertexai/vertexai.go`

```diff
+ Package name: vertexai instead of gemini
+ Type name: vertexAIModel instead of geminiModel
+ Constructor: handles Project/Location instead of APIKey
+ Backend: genai.BackendVertexAI instead of BackendGeminiAPI
+ Error messages: "Vertex AI" instead of "Gemini"

✓ Everything else is identical:
  - GenerateContent() implementation
  - generateStream() logic
  - maybeAppendUserContent() function
  - addHeaders() function
  - Interface implementation (model.LLM)
```

## Detailed Code: Gemini Package

```go
// research/adk-go/model/gemini/gemini.go
package gemini

import "google.golang.org/genai"

type geminiModel struct {
    client             *genai.Client
    name               string
    versionHeaderValue string
}

func NewModel(ctx context.Context, modelName string, 
    cfg *genai.ClientConfig) (model.LLM, error) {
    
    // Note: Client created with genai.NewClient()
    client, err := genai.NewClient(ctx, cfg)
    if err != nil {
        return nil, err
    }
    
    return &geminiModel{
        client: client,
        name:   modelName,
    }, nil
}

func (m *geminiModel) Name() string {
    return m.name
}

func (m *geminiModel) GenerateContent(ctx context.Context,
    req *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
    // ... implementation ...
}
```

## Detailed Code: Vertex AI Package (Proposed)

```go
// research/adk-go/model/vertexai/vertexai.go  [NEW FILE]
package vertexai

import "google.golang.org/genai"

type vertexAIModel struct {  // ← Different name
    client             *genai.Client
    name               string
    project            string  // ← Store for reference
    location           string  // ← Store for reference
    versionHeaderValue string
}

func NewModel(ctx context.Context, modelName string,
    cfg *genai.ClientConfig) (model.LLM, error) {
    
    // Force Vertex AI backend
    cfg.Backend = genai.BackendVertexAI
    
    // genai.NewClient() auto-handles ADC when Project/Location set
    client, err := genai.NewClient(ctx, cfg)
    if err != nil {
        return nil, fmt.Errorf("failed to create Vertex AI client: %w", err)
    }
    
    return &vertexAIModel{
        client:   client,
        name:     modelName,
        project:  cfg.Project,
        location: cfg.Location,
    }, nil
}

func (m *vertexAIModel) Name() string {
    return m.name
}

func (m *vertexAIModel) GenerateContent(ctx context.Context,
    req *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
    // ... IDENTICAL implementation ...
}
```

## Client Configuration Differences

### Gemini API ClientConfig

```go
cfg := &genai.ClientConfig{
    APIKey:  "AIza...",              // Required for Gemini
    Backend: genai.BackendGeminiAPI,
    // No Project/Location needed
}
```

### Vertex AI ClientConfig

```go
cfg := &genai.ClientConfig{
    Project:  "my-gcp-project",      // Required for Vertex
    Location: "us-central1",          // Required for Vertex
    Backend:  genai.BackendVertexAI,
    // No APIKey needed (uses ADC)
}
```

## Environment Variable Initialization

### Gemini API

```bash
# Set API key
export GOOGLE_API_KEY="AIza..."

# Run with auto-detection
./code-agent
# OR explicit
./code-agent --backend gemini
```

### Vertex AI

```bash
# Setup ADC (once per machine/container)
gcloud auth application-default login

# Set project and location
export GOOGLE_CLOUD_PROJECT="my-project"
export GOOGLE_CLOUD_LOCATION="us-central1"

# Run with auto-detection
./code-agent
# OR explicit
./code-agent --backend vertexai
```

## Backend Auto-Detection Logic

```go
// In code_agent/cli.go ParseCLIFlags()
if config.Backend == "" {
    // Priority: explicit flag > env var > default
    
    // Check for Vertex AI env vars
    if os.Getenv("GOOGLE_GENAI_USE_VERTEXAI") == "true" {
        config.Backend = "vertexai"
        return
    }
    
    // Check for Gemini API
    if os.Getenv("GOOGLE_API_KEY") != "" {
        config.Backend = "gemini"
        return
    }
    
    // Default: Vertex AI (requires ADC setup)
    config.Backend = "vertexai"
}
```

## Error Messages

### Gemini API Errors

```
Failed to create model: failed to call model: <error details>
```

### Vertex AI Errors

```
Failed to create model: failed to create Vertex AI client: <error details>
Failed to call Vertex AI model: <error details>
```

## Testing Strategy

Both packages should have identical test patterns:

```go
// test_initialization.go (both backends)
func TestNewModel(t *testing.T) {
    ctx := context.Background()
    model, err := NewModel(ctx, "gemini-2.5-flash", &genai.ClientConfig{...})
    // ... same assertions for both
}

// test_generate_content.go (both backends)
func TestGenerateContent(t *testing.T) {
    // ... same test cases for both
}

// test_streaming.go (both backends)  
func TestGenerateContentStream(t *testing.T) {
    // ... same stream tests for both
}
```

## Migration Matrix

| Scenario | Before | After | Effort |
|----------|--------|-------|--------|
| Gemini API user | Works | Works (no change) | None |
| Vertex AI user | Not supported | Works via new package | Just flag it |
| Developer switching backends | Manual code edit | CLI flag/env var | Minimal |
| Docker image switching | Must rebuild | Just env vars | Zero |
| Kubernetes deployment | Gemini only | Either backend | env vars |

## Package Structure After Implementation

```
research/adk-go/
├── model/
│   ├── llm.go                    (unchanged)
│   ├── gemini/
│   │   ├── gemini.go             (unchanged)
│   │   └── gemini_test.go        (unchanged)
│   └── vertexai/                 [NEW DIRECTORY]
│       ├── vertexai.go           [NEW FILE] 
│       └── vertexai_test.go      [NEW FILE]
└── ... other packages ...
```

## Summary of Changes

| File | Change Type | Impact |
|------|------------|--------|
| `adk-go/model/vertexai/vertexai.go` | Create new | +250 lines (mirrors gemini.go) |
| `adk-go/model/vertexai/vertexai_test.go` | Create new | +100 lines (mirrors tests) |
| `code_agent/cli.go` | Modify | +15 lines (backend flags) |
| `code_agent/main.go` | Modify | +25 lines (backend factory) |
| `code_agent/go.mod` | No change | No changes |
| `adk-go/go.mod` | Possibly update genai version | Depends on required features |

**Total new code**: ~390 lines (mostly copy-paste with minor renames)
**Breaking changes**: Zero ✓
**Backward compatibility**: 100% maintained ✓

---

## Conclusion

The architecture is remarkably clean because:

1. **genai SDK unifies both backends** - Both use the same `genai.Client` with different `Backend` flags
2. **model.LLM interface is universal** - Both Gemini and Vertex AI implementations return the same interface
3. **Code is ~95% identical** - Only authentication and header values differ
4. **No breaking changes** - Existing Gemini API workflows continue unchanged
5. **Environment-based configuration** - Standard cloud-native patterns (env vars)

This makes it one of the cleanest possible implementations for dual-backend support.
