# Model Selection & Provider Architecture - Redesign Brainstorm

**Date**: November 11, 2025
**Topic**: Improving CLI model/provider selection syntax and architecture
**Status**: Brainstorm / Design Proposal

---

## Current State Analysis

### What Works
- ✅ Model registry with clean abstraction (`models.go`)
- ✅ Backend separation (Gemini API vs Vertex AI)
- ✅ Clear model capabilities metadata
- ✅ Environment variable auto-detection for backends
- ✅ Flag-based backend selection (`--backend gemini|vertexai`)
- ✅ Model selection via `--model` flag

### Current Limitations
1. **Duplicate Model Definitions**: Gemini models registered twice (once per backend)
   - `gemini-2.5-flash` (Gemini API)
   - `gemini-2.5-flash-vertex` (Vertex AI)
   - Model metadata is duplicated; only backend differs

2. **Awkward Provider Selection**:
   - Must specify both `--backend` and `--model` separately
   - Or remember model IDs with `-vertex` suffix
   - Inconsistent: `/models` shows both backend variants as separate entries

3. **No Quick Syntax for Provider/Model Pairing**:
   - Can't do: `code-agent --model gemini/2.5-flash`
   - Can't do: `code-agent --model vertexai/gemini-2.5-flash`
   - Requires two flags to switch providers

4. **Display Confusions**:
   - `/models` lists models duplicated (one per backend)
   - No clear visual hierarchy of providers
   - User might forget which `-vertex` suffix belongs to which model

5. **Runtime Model Resolution Issues**:
   - `ResolveModel()` tries multiple heuristics
   - Unclear which takes precedence: explicit backend vs. model ID
   - Model name matching is case-insensitive string search

---

## Proposed Solution

### 1. **Provider-First Architecture**

Treat **providers** as first-class concepts alongside models:

```go
type Provider string
const (
    ProviderGeminiAPI Provider = "gemini"
    ProviderVertexAI  Provider = "vertexai"
)

type ModelReference struct {
    Provider    Provider  // "gemini" or "vertexai"
    ModelID     string    // "gemini-2.5-flash", "gemini-1.5-pro"
    ContextKey  string    // Human-friendly key for display (e.g., "2.5-flash")
}
```

**Benefits:**
- Clean separation of concerns
- No duplicate model registrations
- Single source of truth for model metadata

### 2. **Unified Model Registry**

Restructure registry to store models once, with multiple provider mappings:

```go
type ModelRegistry struct {
    // Base model definitions (once per unique model)
    models map[string]ModelConfig
    
    // Provider-specific model mappings
    // Key: "provider/modelID" or "provider/shorthand"
    // Value: pointer to ModelConfig
    providerModels map[string]*ModelConfig
}

// Examples in registry:
// "gemini/2.5-flash" → gemini-2.5-flash model (Gemini API)
// "vertexai/2.5-flash" → gemini-2.5-flash model (Vertex AI backend)
// "gemini/flash" → shorthand alias
```

**Benefits:**
- Single definition per model
- Multiple provider access paths
- Clean lookup semantics

### 3. **Enhanced CLI Syntax Options**

Support multiple syntaxes for flexibility:

```bash
# Current syntax (still works)
code-agent --backend gemini --model gemini-2.5-flash
code-agent --backend vertexai --project my-project --location us-central1 --model gemini-2.5-flash

# NEW: Provider/Model syntax (recommended)
code-agent --model gemini/2.5-flash
code-agent --model vertexai/gemini-1.5-pro
code-agent --model vertexai/1.5-pro  # shorthand
code-agent --model vertexai/pro       # ultra-shorthand

# NEW: Shorthand aliases
code-agent --model flash              # defaults to default provider's flash
code-agent --model pro                # defaults to default provider's pro

# NEW: Interactive provider/model selection
code-agent --select-model             # shows TUI menu
```

### 4. **New `/providers` Command**

Display providers and their available models:

```
════════════════════════════════════════════════════════════════
                    Available Providers & Models
════════════════════════════════════════════════════════════════

🔷 Gemini API (gemini)
   Default: gemini/2.5-flash (gemini-2.5-flash)
   
   Models:
   ✓ gemini/2.5-flash       - Gemini 2.5 Flash (economy, 1M ctx)
   ○ gemini/2.0-flash       - Gemini 2.0 Flash (economy, 1M ctx)
   ○ gemini/1.5-flash       - Gemini 1.5 Flash (economy, 1M ctx)
   ○ gemini/1.5-pro         - Gemini 1.5 Pro (premium, 2M ctx)
   
   Shortcuts:
   • --model gemini/flash   → gemini/2.5-flash
   • --model gemini/pro     → gemini/1.5-pro

🔶 Vertex AI (vertexai)
   Default: vertexai/2.5-flash (Vertex AI backend)
   
   Models:
   ○ vertexai/2.5-flash     - Gemini 2.5 Flash via Vertex AI (economy, 1M ctx)
   ○ vertexai/1.5-pro       - Gemini 1.5 Pro via Vertex AI (premium, 2M ctx)
   
   Shortcuts:
   • --model vertexai/flash → vertexai/2.5-flash
   • --model vertexai/pro   → vertexai/1.5-pro
   
   Requirements:
   • GOOGLE_CLOUD_PROJECT environment variable
   • GOOGLE_CLOUD_LOCATION environment variable

════════════════════════════════════════════════════════════════
```

### 5. **Updated Model Resolution Logic**

Clear precedence for model selection:

```
Priority Order:
1. Explicit --model with provider/model syntax (e.g., gemini/2.5-flash)
2. Explicit --model with shorthand (e.g., pro)
3. Explicit --backend flag (use default for that provider)
4. Environment variable detection (GOOGLE_GENAI_USE_VERTEXAI)
5. Global default (gemini/2.5-flash)
```

### 6. **Configuration File Option** (Future Enhancement)

Store provider preferences:

```yaml
# ~/.code_agent/config.yaml
default_provider: gemini
default_model: 2.5-flash

providers:
  gemini:
    api_key: ${GOOGLE_API_KEY}
    
  vertexai:
    project: my-gcp-project
    location: us-central1
    credentials: ${GOOGLE_APPLICATION_CREDENTIALS}
```

---

## Implementation Plan

### Phase 1: Minimal Changes
1. Add `--model provider/modelid` parsing to CLI
2. Update `ResolveModel()` to parse provider/model syntax
3. Add `/providers` display command
4. Keep backward compatibility with existing flags

### Phase 2: Registry Refactoring
1. Restructure `ModelRegistry` to use provider-based lookups
2. Eliminate duplicate model registrations
3. Add model alias/shorthand support
4. Update model list display

### Phase 3: Enhanced UX
1. Interactive provider selection (`--select-model` flag)
2. Better error messages for invalid provider/model combos
3. Configuration file support
4. Environment variable override improvements

### Phase 4: Documentation & Deprecation
1. Update help text and examples
2. Deprecate old syntax gradually (still support it)
3. Add migration guide
4. Update README with new syntax

---

## Code Structure Changes

### New Files/Modules
- `provider.go` - Provider definitions and utilities
- Update `models.go` - Registry refactoring
- Update `cli.go` - Enhanced flag parsing

### Key Functions to Add/Modify
```go
// provider.go (NEW)
func ParseProviderModel(input string) (Provider, string, error) {
    // Parse "gemini/2.5-flash" → (gemini, "2.5-flash")
}

func GetProviderMetadata(provider Provider) ProviderInfo {
    // Return name, description, requirements
}

// models.go (MODIFIED)
func (mr *ModelRegistry) RegisterModelForProvider(
    provider Provider,
    modelID string,
    shorthand string,
    config ModelConfig,
) error

func (mr *ModelRegistry) GetModelByProviderAndID(
    provider Provider,
    modelID string,
) (ModelConfig, error)

func (mr *ModelRegistry) ResolveFromSyntax(
    syntax string, // "provider/model" or "model"
    defaultProvider Provider,
) (Provider, ModelConfig, error)

// cli.go (MODIFIED)
func ParseProviderModelSyntax(input string) (Provider, string, error) {
    // Enhanced parsing for provider/model syntax
}
```

---

## Example Usage Scenarios

### Scenario 1: Quick Switch Between Providers
```bash
# Using current approach (verbose)
code-agent --backend gemini --model gemini-2.5-flash

# Using new syntax (clean)
code-agent --model gemini/2.5-flash

# Even shorter with shorthand
code-agent --model gemini/flash
```

### Scenario 2: Using Vertex AI
```bash
# Current (complex)
code-agent \
  --backend vertexai \
  --project my-project \
  --location us-central1 \
  --model gemini-1.5-pro-vertex

# New approach (clear provider association)
code-agent \
  --model vertexai/1.5-pro \
  --project my-project \
  --location us-central1
```

### Scenario 3: Discovering Available Models
```bash
# Current
code-agent /models
# Shows: "gemini-2.5-flash", "gemini-2.5-flash-vertex", "gemini-1.5-pro", "gemini-1.5-pro-vertex"

# New
code-agent /providers
# Shows:
# 🔷 Gemini API: gemini/2.5-flash, gemini/1.5-pro
# 🔶 Vertex AI: vertexai/2.5-flash, vertexai/1.5-pro
```

---

## Clean Architecture (No Legacy Support Needed)

Since backward compatibility is not required, we can implement the **cleanest design** without supporting old flag combinations:

**Removed:**
- ❌ `--backend` flag (replaced by provider prefix in `--model`)
- ❌ Old model ID format with `-vertex` suffix
- ❌ Separate `--api-key`, `--project`, `--location` flags (configuration-driven)

**Replaced with:**
- ✅ `--model provider/model` - Single, clear syntax
- ✅ `--config` or env vars - Provider-specific credentials
- ✅ `/providers` command - Shows all available options

**Example:**
```bash
# Old (no longer supported)
code-agent --backend vertexai --project my-proj --location us-central1 --model gemini-1.5-pro

# New (clean, unified)
code-agent --model vertexai/1.5-pro --project my-proj --location us-central1
```

This enables us to:
- ✅ Eliminate all model duplication completely
- ✅ Simplify CLI flag parsing significantly
- ✅ Create cleaner mental model for users
- ✅ Design provider system extensibly from day one
- ✅ No technical debt from supporting two syntaxes

---

## Open Design Questions

1. **Shorthand Ambiguity**: If user types `--model pro`, should we search across all providers or just the default?
   - Option A: Search default provider only (safest)
   - Option B: Show menu if ambiguous (better UX)
   - Option C: Default to Gemini, require explicit provider for others (simpler)

2. **Provider Auto-Detection**: When user specifies `--model gemini-1.5-pro-vertex`, should we auto-detect provider?
   - Current: Would fail with "model not found"
   - Better: Extract `-vertex` suffix and auto-select Vertex AI provider

3. **Model Aliases**: How many shorthand aliases should each model have?
   - Example: `gemini/2.5-flash` could be `gemini/flash` or `gemini/25`
   - Tradeoff: More flexibility vs. confusion

4. **Config File Precedence**: If user sets default provider in config, but uses `--backend` flag, which wins?
   - Recommendation: CLI flags > config file > environment > defaults

5. **Provider Requirements**: How should we surface that Vertex AI requires project/location?
   - Show in `/providers` output ✅
   - Validate early and provide clear error message ✅
   - Suggest `--project` and `--location` in help ✅

---

## Benefits Summary

| Aspect | Current | Proposed |
|--------|---------|----------|
| **Syntax Clarity** | `--backend gemini --model gemini-2.5-flash` | `--model gemini/2.5-flash` |
| **Model Duplication** | Yes (once per provider) | No (single definition) |
| **Provider Visibility** | Hidden in flags | First-class in API |
| **Discoverability** | `/models` shows flat list | `/providers` shows hierarchy |
| **Shorthand** | None | Supported (e.g., `gemini/flash`) |
| **Default Provider** | Implicit (Gemini API) | Explicit & configurable |
| **Error Messages** | Generic | Provider-aware |

---

## Summary

This redesign elevates **provider** from an implicit backend choice to a first-class concept alongside models. The `provider/model` syntax is:
- ✅ **Intuitive**: Clear provider/model pairing
- ✅ **Flexible**: Supports full IDs and shorthand
- ✅ **Discoverable**: `/providers` shows structure
- ✅ **Backward Compatible**: Old flags still work
- ✅ **Extensible**: Easy to add new providers (e.g., Claude, OpenAI in future)

The main win: **Users understand what provider they're using, and can switch between Gemini and Vertex AI effortlessly**.
