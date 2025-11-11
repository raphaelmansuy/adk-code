# Vertex AI + Gemini API Support Architecture

## Executive Summary

The `code_agent` and `research/adk-go` codebase can support both **Vertex AI** (with GCP project credentials) and **Gemini API** (with API keys) with minimal architectural changes. The new Google Gen AI SDK (`google.golang.org/genai`) already provides unified backend abstraction, which we can leverage perfectly.

**Recommended approach**: Create a parallel `model/vertexai` package in adk-go following the same pattern as `model/gemini`, with backend selection in `code_agent` based on environment variables or CLI flags.

---

## Current State Analysis

### code_agent Architecture

**Entry Point**: `code_agent/main.go`
```go
// Currently hardcoded to Gemini API
model, err := gemini.NewModel(ctx, "gemini-2.5-flash", &genai.ClientConfig{
    APIKey: apiKey,
})
```

**Model Layer**: Uses `model.LLM` interface from adk-go
```go
// From research/adk-go/model/llm.go
type LLM interface {
    Name() string
    GenerateContent(ctx context.Context, req *LLMRequest, stream bool) iter.Seq2[*LLMResponse, error]
}
```

**Implementation**: `research/adk-go/model/gemini/gemini.go`
- Creates `genai.Client` with `APIKey` and `BackendGeminiAPI`
- Implements `model.LLM` interface
- Handles streaming and non-streaming responses

### Google Gen AI SDK Capabilities

The current dependency is `google.golang.org/genai v1.20.0`, but the new SDK (v1.34+) provides:

```go
type Backend int

const (
    BackendUnspecified  // Auto-detect based on env vars
    BackendGeminiAPI    // Gemini Developer API
    BackendVertexAI     // Vertex AI in GCP
)

type ClientConfig struct {
    // For Gemini API
    APIKey  string
    
    // For Vertex AI
    Project   string
    Location  string
    
    // Common
    Backend     Backend
    Credentials *auth.Credentials
    HTTPClient  *http.Client
    HTTPOptions HTTPOptions
}
```

**Key Insight**: Both backends are already unified in the genai SDK. We just need to:
1. Pass the correct `Backend` flag
2. Provide appropriate credentials (APIKey for Gemini, Project/Location for Vertex)
3. Let the SDK handle the rest

---

## Recommended Architecture

### 1. **Parallel Model Packages in adk-go**

**Gemini Package** (existing, minimal changes):
```
research/adk-go/model/gemini/
├── gemini.go          // Existing implementation
└── gemini_test.go     // Existing tests
```

**Vertex AI Package** (new, mirrors Gemini):
```
research/adk-go/model/vertexai/
├── vertexai.go        // New implementation (mirrors gemini.go structure)
└── vertexai_test.go   // New tests
```

### 2. **Backend Selection in code_agent**

**New module**: `code_agent/model_factory.go`
```go
package main

import (
    "context"
    "google.golang.org/genai"
    "google.golang.org/adk/model"
    "code_agent/research/adk-go/model/gemini"
    "code_agent/research/adk-go/model/vertexai"
)

type ModelConfig struct {
    Backend  string // "gemini" or "vertexai"
    APIKey   string // For Gemini
    Project  string // For Vertex AI
    Location string // For Vertex AI
    ModelName string
}

func CreateModel(ctx context.Context, cfg ModelConfig) (model.LLM, error) {
    switch cfg.Backend {
    case "vertexai":
        return vertexai.NewModel(ctx, cfg.ModelName, &genai.ClientConfig{
            Project:  cfg.Project,
            Location: cfg.Location,
            Backend:  genai.BackendVertexAI,
        })
    case "gemini":
        fallthrough
    default:
        return gemini.NewModel(ctx, cfg.ModelName, &genai.ClientConfig{
            APIKey:  cfg.APIKey,
            Backend: genai.BackendGeminiAPI,
        })
    }
}
```

### 3. **Environment Variable & CLI Flag Support**

**Priority order** (from highest to lowest):
1. CLI flags (explicit user choice)
2. Environment variables (configuration layer)
3. Auto-detection (SDK default behavior)

**Environment Variables**:
```bash
# Vertex AI Mode
export GOOGLE_GENAI_USE_VERTEXAI=true
export GOOGLE_CLOUD_PROJECT=your-project-id
export GOOGLE_CLOUD_LOCATION=us-central1

# Gemini API Mode
export GOOGLE_API_KEY=your-api-key
```

**CLI Flags** (new additions to code_agent):
```bash
# Explicitly select backend
code-agent --backend vertexai --project my-project --location us-central1

# Gemini API (existing)
code-agent --api-key <key>

# Auto-detect (default)
code-agent
```

---

## Implementation Details

### Step 1: Create model/vertexai Package in adk-go

**File**: `research/adk-go/model/vertexai/vertexai.go`

```go
package vertexai

import (
    "context"
    "fmt"
    "iter"
    "runtime"
    "strings"

    "google.golang.org/adk/internal/llminternal"
    "google.golang.org/adk/internal/llminternal/converters"
    "google.golang.org/adk/internal/version"
    "google.golang.org/adk/model"
    "google.golang.org/genai"
)

type vertexAIModel struct {
    client             *genai.Client
    name               string
    project            string
    location           string
    versionHeaderValue string
}

// NewModel creates a Vertex AI model for use in agent applications.
// It requires a GCP project ID and location to be configured.
// Credentials are automatically discovered using Application Default Credentials (ADC).
func NewModel(ctx context.Context, modelName string, cfg *genai.ClientConfig) (model.LLM, error) {
    if cfg == nil {
        cfg = &genai.ClientConfig{}
    }
    
    // Force Vertex AI backend
    cfg.Backend = genai.BackendVertexAI
    
    client, err := genai.NewClient(ctx, cfg)
    if err != nil {
        return nil, fmt.Errorf("failed to create Vertex AI client: %w", err)
    }

    headerValue := fmt.Sprintf("google-adk-vertex/%s gl-go/%s", version.Version,
        strings.TrimPrefix(runtime.Version(), "go"))

    return &vertexAIModel{
        name:               modelName,
        client:             client,
        project:            cfg.Project,
        location:           cfg.Location,
        versionHeaderValue: headerValue,
    }, nil
}

func (m *vertexAIModel) Name() string {
    return m.name
}

// GenerateContent calls the Vertex AI model with streaming support.
func (m *vertexAIModel) GenerateContent(ctx context.Context, req *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
    m.maybeAppendUserContent(req)
    if req.Config == nil {
        req.Config = &genai.GenerateContentConfig{}
    }
    if req.Config.HTTPOptions == nil {
        req.Config.HTTPOptions = &genai.HTTPOptions{}
    }
    if req.Config.HTTPOptions.Headers == nil {
        req.Config.HTTPOptions.Headers = make(http.Header)
    }
    m.addHeaders(req.Config.HTTPOptions.Headers)

    if stream {
        return m.generateStream(ctx, req)
    }

    return func(yield func(*model.LLMResponse, error) bool) {
        resp, err := m.generate(ctx, req)
        yield(resp, err)
    }
}

func (m *vertexAIModel) addHeaders(headers http.Header) {
    headers.Set("x-goog-api-client", m.versionHeaderValue)
    headers.Set("user-agent", m.versionHeaderValue)
}

func (m *vertexAIModel) generate(ctx context.Context, req *model.LLMRequest) (*model.LLMResponse, error) {
    resp, err := m.client.Models.GenerateContent(ctx, m.name, req.Contents, req.Config)
    if err != nil {
        return nil, fmt.Errorf("failed to call Vertex AI model: %w", err)
    }
    if len(resp.Candidates) == 0 {
        return nil, fmt.Errorf("empty response from Vertex AI model")
    }
    return converters.Genai2LLMResponse(resp), nil
}

func (m *vertexAIModel) generateStream(ctx context.Context, req *model.LLMRequest) iter.Seq2[*model.LLMResponse, error] {
    aggregator := llminternal.NewStreamingResponseAggregator()

    return func(yield func(*model.LLMResponse, error) bool) {
        for resp, err := range m.client.Models.GenerateContentStream(ctx, m.name, req.Contents, req.Config) {
            if err != nil {
                yield(nil, err)
                return
            }
            for llmResponse, err := range aggregator.ProcessResponse(ctx, resp) {
                if !yield(llmResponse, err) {
                    return
                }
            }
        }
        if closeResult := aggregator.Close(); closeResult != nil {
            yield(closeResult, nil)
        }
    }
}

func (m *vertexAIModel) maybeAppendUserContent(req *model.LLMRequest) {
    if len(req.Contents) == 0 {
        req.Contents = append(req.Contents, genai.NewContentFromText(
            "Handle the requests as specified in the System Instruction.", "user"))
    }

    if last := req.Contents[len(req.Contents)-1]; last != nil && last.Role != "user" {
        req.Contents = append(req.Contents, genai.NewContentFromText(
            "Continue processing previous requests as instructed. Exit or provide a summary if no more outputs are needed.", "user"))
    }
}
```

### Step 2: Update code_agent/cli.go

Add backend selection flags:
```go
type CLIConfig struct {
    OutputFormat         string
    SessionName          string
    DBPath               string
    TypewriterEnabled    bool
    WorkingDirectory     string
    
    // New backend selection fields
    Backend              string // "gemini" or "vertexai"
    APIKey               string // For Gemini
    VertexAIProject      string // For Vertex AI
    VertexAILocation     string // For Vertex AI
}

func ParseCLIFlags() (*CLIConfig, []string) {
    config := &CLIConfig{}
    
    // Existing flags...
    
    // New flags for backend selection
    flag.StringVar(&config.Backend, "backend", "", 
        "Backend to use: 'gemini' or 'vertexai' (default: auto-detect from env vars)")
    flag.StringVar(&config.VertexAIProject, "project", 
        os.Getenv("GOOGLE_CLOUD_PROJECT"), 
        "GCP Project ID for Vertex AI")
    flag.StringVar(&config.VertexAILocation, "location", 
        os.Getenv("GOOGLE_CLOUD_LOCATION"), 
        "GCP Location for Vertex AI (e.g., us-central1)")
    
    flag.Parse()
    
    // Auto-detect backend from environment if not specified
    if config.Backend == "" {
        if os.Getenv("GOOGLE_GENAI_USE_VERTEXAI") == "true" || 
           os.Getenv("GOOGLE_GENAI_USE_VERTEXAI") == "1" {
            config.Backend = "vertexai"
        } else if os.Getenv("GOOGLE_API_KEY") != "" {
            config.Backend = "gemini"
        } else {
            // Default: try Vertex AI (use ADC if available)
            config.Backend = "vertexai"
        }
    }
    
    return config, flag.Args()
}
```

### Step 3: Update code_agent/main.go

```go
func main() {
    // ... existing setup code ...
    
    cliConfig, args := ParseCLIFlags()
    
    // Create model based on backend selection
    var modelErr error
    var llmModel model.LLM
    
    switch cliConfig.Backend {
    case "vertexai":
        if cliConfig.VertexAIProject == "" {
            log.Fatal("Vertex AI backend requires GOOGLE_CLOUD_PROJECT to be set")
        }
        if cliConfig.VertexAILocation == "" {
            log.Fatal("Vertex AI backend requires GOOGLE_CLOUD_LOCATION to be set")
        }
        llmModel, modelErr = vertexai.NewModel(ctx, "gemini-2.5-flash", &genai.ClientConfig{
            Project:  cliConfig.VertexAIProject,
            Location: cliConfig.VertexAILocation,
            Backend:  genai.BackendVertexAI,
        })
        
    case "gemini":
        apiKey := os.Getenv("GOOGLE_API_KEY")
        if apiKey == "" {
            log.Fatal("Gemini API backend requires GOOGLE_API_KEY environment variable")
        }
        llmModel, modelErr = gemini.NewModel(ctx, "gemini-2.5-flash", &genai.ClientConfig{
            APIKey:  apiKey,
            Backend: genai.BackendGeminiAPI,
        })
        
    default:
        log.Fatalf("Unknown backend: %s", cliConfig.Backend)
    }
    
    if modelErr != nil {
        log.Fatalf("Failed to create LLM model: %v", modelErr)
    }
    
    // Print banner showing which backend is active
    modelName := fmt.Sprintf("gemini-2.5-flash (%s)", cliConfig.Backend)
    banner := bannerRenderer.RenderStartBanner(version, modelName, workingDir)
    fmt.Print(banner)
    
    // ... rest of existing code ...
}
```

---

## Comparison: Gemini API vs Vertex AI

| Aspect | Gemini API | Vertex AI |
|--------|-----------|----------|
| **Authentication** | API Key | ADC (GCP credentials) |
| **Setup Requirement** | Get API key from ai.google.dev | GCP project + Vertex AI enabled |
| **Environment Vars** | `GOOGLE_API_KEY` | `GOOGLE_CLOUD_PROJECT`, `GOOGLE_CLOUD_LOCATION` |
| **Cost** | Direct billing | GCP billing |
| **Rate Limits** | API key based | GCP project quotas |
| **Data Residency** | US only | Multi-region support |
| **Enterprise Features** | Limited | Full Vertex AI features (RAG, grounding, etc.) |
| **Use Case** | Quick start, development | Production, compliance, grounding |

---

## Configuration Examples

### Example 1: Gemini API (Quick Start)

```bash
export GOOGLE_API_KEY="your-api-key-here"
./code-agent
```

Or explicitly:
```bash
./code-agent --backend gemini --api-key "your-api-key-here"
```

### Example 2: Vertex AI (Production)

```bash
# Setup GCP credentials (once)
gcloud auth application-default login

# Run with Vertex AI
export GOOGLE_CLOUD_PROJECT="my-gcp-project"
export GOOGLE_CLOUD_LOCATION="us-central1"
./code-agent --backend vertexai
```

Or with environment-based auto-detection:
```bash
export GOOGLE_GENAI_USE_VERTEXAI=true
export GOOGLE_CLOUD_PROJECT="my-gcp-project"
export GOOGLE_CLOUD_LOCATION="us-central1"
./code-agent
```

### Example 3: Docker with Vertex AI

```dockerfile
FROM golang:1.24

WORKDIR /app
COPY . .
RUN go build -o code-agent ./code_agent

# Use ADC via mounted service account
ENV GOOGLE_APPLICATION_CREDENTIALS=/var/secrets/google/key.json

ENTRYPOINT ["./code-agent", "--backend", "vertexai", "--project", "$GCP_PROJECT", "--location", "$GCP_LOCATION"]
```

---

## Implementation Phases

### Phase 1: Foundation (Week 1)
- [x] Create `research/adk-go/model/vertexai` package
- [x] Add backend selection to `code_agent/cli.go`
- [x] Update `code_agent/main.go` to use model factory
- [ ] Test with actual Vertex AI credentials

### Phase 2: Documentation & Cleanup (Week 2)
- [ ] Update README.md with backend selection instructions
- [ ] Add authentication guides (Gemini API vs Vertex AI)
- [ ] Create troubleshooting guide
- [ ] Update example Docker/K8s configs

### Phase 3: Advanced Features (Future)
- [ ] Support for Vertex AI specific features (grounding, RAG)
- [ ] Multi-region support for Vertex AI
- [ ] Cost tracking/monitoring per backend
- [ ] Backend switching within runtime

---

## Benefits of This Approach

✅ **Minimal Code Changes**
- Leverages existing `genai` SDK architecture
- No breaking changes to existing code_agent users

✅ **Backward Compatible**
- Existing Gemini API workflows continue unchanged
- New Vertex AI users can opt-in

✅ **Flexible Configuration**
- Environment variables for automation
- CLI flags for explicit control
- Auto-detection fallback

✅ **Future-Proof**
- Easy to add new backends (just follow the pattern)
- Both packages use same `model.LLM` interface
- Aligns with Google's unified genai SDK direction

✅ **Enterprise Ready**
- Vertex AI support for regulated environments
- Multi-region data residency options
- GCP IAM integration (no API key management)

---

## Summary

The `code_agent` can support both Vertex AI and Gemini API by:

1. **Creating a mirror `model/vertexai` package** in adk-go (follows the exact pattern of `model/gemini`)
2. **Adding backend detection in code_agent** with sensible defaults (auto-detect from env vars)
3. **Leveraging the genai SDK's native backend abstraction** (which both packages use)
4. **Using environment variables for configuration** (standard for cloud-native apps)

This approach is clean, maintainable, and requires minimal changes to the existing codebase while providing full support for both APIs with different authentication mechanisms and deployment scenarios.
