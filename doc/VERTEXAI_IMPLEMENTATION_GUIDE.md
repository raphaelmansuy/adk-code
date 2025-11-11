# Quick Implementation Guide: Vertex AI + Gemini Support

## Overview

This guide provides step-by-step instructions to add Vertex AI support to code_agent alongside existing Gemini API support.

## Architecture Summary

```
User Request
    ↓
[Backend Detection: env vars/CLI flags]
    ├→ GEMINI path: --backend gemini or GOOGLE_API_KEY set
    │   └→ model/gemini.NewModel() → genai.Client(BackendGeminiAPI)
    │
    └→ VERTEX AI path: --backend vertexai or GOOGLE_GENAI_USE_VERTEXAI=true
        └→ model/vertexai.NewModel() → genai.Client(BackendVertexAI)
```

Both paths return `model.LLM` interface → coding_agent.NewCodingAgent()

---

## Quick Start

### For Gemini API (Existing - No Changes Required)

```bash
export GOOGLE_API_KEY="your-api-key"
./code-agent
```

### For Vertex AI (New - With Proposed Changes)

```bash
# Setup GCP auth once
gcloud auth application-default login

# Run with auto-detection
export GOOGLE_CLOUD_PROJECT="my-project"
export GOOGLE_CLOUD_LOCATION="us-central1"
./code-agent
```

---

## Implementation Checklist

### 1. Create Vertex AI Model Package

**File**: `research/adk-go/model/vertexai/vertexai.go`

Key differences from Gemini:
- Uses `genai.BackendVertexAI` instead of `BackendGeminiAPI`
- Requires `Project` and `Location` in `ClientConfig`
- Uses Application Default Credentials (no API key needed)
- Same `model.LLM` interface implementation

**Copy-from**: `research/adk-go/model/gemini/gemini.go`
- Keep 95% of the code identical
- Change "Gemini" to "VertexAI" in function/type names and comments
- Set `Backend: genai.BackendVertexAI` instead of `BackendGeminiAPI`

### 2. Update code_agent CLI

**File**: `code_agent/cli.go`

Add these fields to `CLIConfig`:

```go
type CLIConfig struct {
    // ... existing fields ...
    
    // Backend selection (NEW)
    Backend          string // "gemini" | "vertexai"
    VertexAIProject  string // GCP project ID
    VertexAILocation string // GCP location
}
```

Add flags in `ParseCLIFlags()`:

```go
flag.StringVar(&config.Backend, "backend", "", 
    "Backend: 'gemini' or 'vertexai' (auto-detect if not set)")
flag.StringVar(&config.VertexAIProject, "project", 
    os.Getenv("GOOGLE_CLOUD_PROJECT"), 
    "GCP Project ID")
flag.StringVar(&config.VertexAILocation, "location", 
    os.Getenv("GOOGLE_CLOUD_LOCATION"), 
    "GCP Location")
```

Add auto-detection logic:

```go
// Auto-detect backend from environment
if config.Backend == "" {
    if os.Getenv("GOOGLE_GENAI_USE_VERTEXAI") == "true" {
        config.Backend = "vertexai"
    } else if os.Getenv("GOOGLE_API_KEY") != "" {
        config.Backend = "gemini"
    } else {
        config.Backend = "vertexai" // default
    }
}
```

### 3. Update code_agent Main Function

**File**: `code_agent/main.go`

Replace the hardcoded Gemini creation with factory pattern:

```go
// Old code (remove):
// model, err := gemini.NewModel(ctx, "gemini-2.5-flash", &genai.ClientConfig{
//     APIKey: apiKey,
// })

// New code (add):
var model model.LLM
var modelErr error

switch cliConfig.Backend {
case "vertexai":
    if cliConfig.VertexAIProject == "" {
        log.Fatal("Vertex AI requires GOOGLE_CLOUD_PROJECT")
    }
    if cliConfig.VertexAILocation == "" {
        log.Fatal("Vertex AI requires GOOGLE_CLOUD_LOCATION")
    }
    model, modelErr = vertexai.NewModel(ctx, "gemini-2.5-flash", &genai.ClientConfig{
        Project:  cliConfig.VertexAIProject,
        Location: cliConfig.VertexAILocation,
        Backend:  genai.BackendVertexAI,
    })
case "gemini":
    apiKey := os.Getenv("GOOGLE_API_KEY")
    if apiKey == "" {
        log.Fatal("Gemini API requires GOOGLE_API_KEY")
    }
    model, modelErr = gemini.NewModel(ctx, "gemini-2.5-flash", &genai.ClientConfig{
        APIKey:  apiKey,
        Backend: genai.BackendGeminiAPI,
    })
default:
    log.Fatalf("Unknown backend: %s", cliConfig.Backend)
}

if modelErr != nil {
    log.Fatalf("Failed to create model: %v", modelErr)
}
```

Also update banner to show active backend:

```go
backendName := cliConfig.Backend
if backendName == "" {
    backendName = "auto"
}
modelName := fmt.Sprintf("gemini-2.5-flash (%s)", backendName)
banner := bannerRenderer.RenderStartBanner(version, modelName, workingDir)
```

### 4. Add Import Statements

**File**: `code_agent/main.go`

Add imports at top:

```go
import (
    // ... existing imports ...
    
    "google.golang.org/adk/model"
    "google.golang.org/adk/model/gemini"
    "google.golang.org/adk/model/vertexai"  // NEW
    "google.golang.org/genai"
)
```

### 5. Testing

**File**: Add `research/adk-go/model/vertexai/vertexai_test.go`

Copy test patterns from `model/gemini/gemini_test.go` and adapt for Vertex AI:

```go
package vertexai

import (
    "context"
    "testing"
    "google.golang.org/genai"
)

func TestNewModel(t *testing.T) {
    ctx := context.Background()
    
    // This test requires valid GCP credentials
    cfg := &genai.ClientConfig{
        Project:  "test-project",
        Location: "us-central1",
        Backend:  genai.BackendVertexAI,
    }
    
    model, err := NewModel(ctx, "gemini-2.5-flash", cfg)
    if err != nil {
        t.Skipf("Skipping Vertex AI test (credentials not available): %v", err)
    }
    
    if model.Name() != "gemini-2.5-flash" {
        t.Errorf("Expected model name 'gemini-2.5-flash', got '%s'", model.Name())
    }
}
```

---

## Configuration Reference

### Environment Variables

| Variable | Purpose | Example |
|----------|---------|---------|
| `GOOGLE_API_KEY` | Gemini API key | `"AIza..."`  |
| `GOOGLE_GENAI_USE_VERTEXAI` | Force Vertex AI backend | `"true"` |
| `GOOGLE_CLOUD_PROJECT` | GCP Project ID | `"my-project"` |
| `GOOGLE_CLOUD_LOCATION` | GCP Region | `"us-central1"` |
| `GOOGLE_APPLICATION_CREDENTIALS` | Service account JSON path | `"/path/to/key.json"` |

### CLI Flags

```bash
# Gemini API
./code-agent --backend gemini --api-key <KEY>

# Vertex AI
./code-agent --backend vertexai --project <PROJECT> --location <LOCATION>

# Auto-detect (default)
./code-agent
```

---

## Minimal diff for existing code

The changes are minimal and focused:

1. **adk-go**: Add 1 new package (`model/vertexai/`) - mirrors existing `model/gemini/`
2. **code_agent/cli.go**: Add 3 config fields + flag parsing
3. **code_agent/main.go**: Add ~20 lines for backend selection + import
4. **Zero breaking changes** to existing Gemini API workflow

---

## Validation Checklist

After implementing:

- [ ] Existing Gemini API still works: `GOOGLE_API_KEY=xxx ./code-agent`
- [ ] Vertex AI works with credentials: `gcloud auth application-default login && GOOGLE_CLOUD_PROJECT=xxx GOOGLE_CLOUD_LOCATION=yyy ./code-agent`
- [ ] Help shows new flags: `./code-agent --help`
- [ ] Auto-detection works with env vars
- [ ] Docker images can use either backend via env vars
- [ ] Tests pass: `make test`
- [ ] Code passes lint: `make check`

---

## Files to Modify/Create

### Create (New)
- `research/adk-go/model/vertexai/vertexai.go`
- `research/adk-go/model/vertexai/vertexai_test.go`

### Modify
- `code_agent/cli.go` - Add backend selection flags
- `code_agent/main.go` - Add backend factory logic

### Documentation (Recommended)
- `README.md` - Update with backend selection instructions
- `doc/VERTEXAI_GEMINI_INTEGRATION.md` - This comprehensive guide

---

## Troubleshooting

### Vertex AI: "credentials not available"

```bash
# Ensure ADC is configured:
gcloud auth application-default login

# Or specify service account:
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/key.json
```

### Gemini API: "API key not valid"

```bash
# Check key format and set correctly:
export GOOGLE_API_KEY="AIza..." # Must start with AIza
```

### Backend not switching

```bash
# Check which backend is active:
./code-agent --help | grep backend

# Explicit override:
./code-agent --backend vertexai --project my-project --location us-central1
```

---

## Migration Path for Existing Users

**No action required** - Gemini API users continue working as before.

**To migrate to Vertex AI**:
1. Enable Vertex AI in your GCP project
2. Set environment variables: `GOOGLE_CLOUD_PROJECT`, `GOOGLE_CLOUD_LOCATION`
3. Run: `gcloud auth application-default login`
4. Restart code-agent (no code changes needed!)

---

## Next Steps

1. **Implement** the 5 changes above
2. **Test** with both backends in various environments
3. **Document** in main README.md
4. **Release** as minor version update (backward compatible)
5. **Monitor** community feedback on Vertex AI usage

This approach leverages the genai SDK's native unified backend abstraction, keeping code_agent clean and maintainable.
