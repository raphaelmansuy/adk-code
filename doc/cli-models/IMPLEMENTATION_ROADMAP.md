# Implementation Roadmap: Provider/Model Selection

**Status:** Brainstorm & Design Phase  
**Target:** Code-Agent v1.1.0  
**Priority:** Medium (UX improvement, extensibility)

---

## Overview

This document outlines the specific code changes needed to implement the provider/model selection redesign described in `MODEL_SELECTION_REDESIGN.md` and `PROVIDER_MODEL_VISUAL_EXAMPLES.md`.

---

## Phase 1: Clean Implementation (MVP - Fresh Architecture)

### Goal
Implement `--model provider/model` syntax as the primary approach with no legacy support needed.

### Files to Modify

#### 1. `code_agent/cli.go`

**Changes needed:**
- Add `ParseProviderModelSyntax()` function to parse "provider/model" strings
- Modify `ParseCLIFlags()` to remove `--backend` flag entirely
- Remove `--api-key`, `--project`, `--location` CLI flags (use config file or env vars)
- Update help text with new syntax examples
- Update `printModelsList()` to show provider information with `/providers` command

```go
// Add near top of file
func ParseProviderModelSyntax(input string) (string, string, error) {
    // "gemini/2.5-flash" → ("gemini", "2.5-flash")
    // "flash" → ("", "flash")  // Default provider
    parts := strings.Split(input, "/")
    switch len(parts) {
    case 1:
        return "", parts[0], nil  // Shorthand without provider
    case 2:
        if parts[0] == "" || parts[1] == "" {
            return "", "", fmt.Errorf("invalid model syntax: %q", input)
        }
        return parts[0], parts[1], nil
    default:
        return "", "", fmt.Errorf("invalid model syntax: %q (use provider/model)", input)
    }
}

// Modify ParseCLIFlags help text
model := flag.String("model",
    "",
    "Model to use. Examples:\n"+
    "  --model gemini/2.5-flash     (explicit provider)\n"+
    "  --model gemini/flash          (shorthand)\n"+
    "  --model flash                 (uses default provider)\n"+
    "  --model vertexai/1.5-pro      (Vertex AI model)\n"+
    "Use '/providers' command to list all available models")

// In handleBuiltinCommand(), add case for "/providers"
case "/providers":
    printProvidersList(renderer, modelRegistry)
    return true
```

#### 2. `code_agent/models.go`

**Changes needed:**
- Add provider-aware lookup methods to `ModelRegistry`
- Implement shorthand alias resolution
- Improve error messages

```go
// Add to ModelRegistry struct
type ModelRegistry struct {
    models map[string]ModelConfig
    // NEW: Provider-specific models mapping
    // Key: "provider/modelid" or "provider/shorthand"
    aliases map[string]string  // Alias → Full model ID
}

// Add new methods
func (mr *ModelRegistry) ResolveFromProviderSyntax(
    providerName string,
    modelIdentifier string,
    defaultProvider string,
) (ModelConfig, error) {
    // If provider not specified, use default
    if providerName == "" {
        providerName = defaultProvider
    }

    // Try to find model in registry with provider prefix
    fullID := fmt.Sprintf("%s/%s", providerName, modelIdentifier)
    
    // Try exact match first
    if alias, exists := mr.aliases[fullID]; exists {
        return mr.GetModel(alias)
    }
    
    // Try to find by shorthand (e.g., "flash" → "gemini-2.5-flash")
    // ...
    
    return ModelConfig{}, fmt.Errorf(
        "model %q not found in provider %q",
        modelIdentifier, providerName)
}

func (mr *ModelRegistry) GetProviderModels(provider string) []ModelConfig {
    // Return all models for a specific provider
    var result []ModelConfig
    for _, model := range mr.models {
        if model.Backend == provider {
            result = append(result, model)
        }
    }
    return result
}

func (mr *ModelRegistry) ListProviders() []string {
    // Return list of available providers
    providers := make(map[string]bool)
    for _, model := range mr.models {
        providers[model.Backend] = true
    }
    
    result := make([]string, 0, len(providers))
    for p := range providers {
        result = append(result, p)
    }
    sort.Strings(result)
    return result
}
```

#### 3. `code_agent/main.go`

**Changes needed:**
- Update model resolution logic to use new parser
- Provide better error messages when model not found
- Display provider info in banner

```go
// In main(), model resolution is now much simpler
parsedProvider, parsedModel, err := ParseProviderModelSyntax(cliConfig.Model)
if err != nil {
    log.Fatalf("Invalid model syntax: %v (use provider/model like gemini/2.5-flash)", err)
}

// Require explicit provider - no default fallback
if parsedProvider == "" {
    log.Fatal("Provider required (use --model provider/modelid)")
}

// Resolve actual model
selectedModel, err := modelRegistry.GetModelByProviderAndID(
    parsedProvider,
    parsedModel,
)
if err != nil {
    // Provide helpful suggestions
    fmt.Fprintf(os.Stderr, "❌ Error: %v\n\n", err)
    suggestModels(os.Stderr, modelRegistry, parsedProvider)
    os.Exit(1)
}
```

---

## Phase 2: Registry Refactoring (Breaking Internal Changes)

### Goal
Eliminate duplicate model definitions while maintaining public API.

### Files to Create

#### 1. `code_agent/provider.go` (NEW)

```go
// Package main - Provider definitions and utilities
package main

import "sort"

// Provider represents a backend provider for LLMs
type Provider string

const (
    ProviderGemini   Provider = "gemini"
    ProviderVertexAI Provider = "vertexai"
)

// ProviderMetadata describes a provider
type ProviderMetadata struct {
    Name         string
    DisplayName  string
    Icon         string
    Description  string
    Requirements []string  // e.g., ["GOOGLE_API_KEY", "GOOGLE_CLOUD_PROJECT"]
    IsConfigured bool
}

// AllProviders returns list of all supported providers
func AllProviders() []Provider {
    return []Provider{ProviderGemini, ProviderVertexAI}
}

// GetProviderMetadata returns information about a provider
func GetProviderMetadata(provider Provider) ProviderMetadata {
    switch provider {
    case ProviderGemini:
        return ProviderMetadata{
            Name:        "gemini",
            DisplayName: "Gemini API",
            Icon:        "🔷",
            Description: "REST API with Google's Gemini models",
            Requirements: []string{"GOOGLE_API_KEY"},
        }
    case ProviderVertexAI:
        return ProviderMetadata{
            Name:        "vertexai",
            DisplayName: "Vertex AI",
            Icon:        "🔶",
            Description: "GCP-native endpoint for Google's Gemini models",
            Requirements: []string{"GOOGLE_CLOUD_PROJECT", "GOOGLE_CLOUD_LOCATION"},
        }
    default:
        return ProviderMetadata{}
    }
}

// String returns the provider name
func (p Provider) String() string {
    return string(p)
}

// SortedProviders returns providers in a consistent order
func SortedProviders() []Provider {
    providers := AllProviders()
    sort.Slice(providers, func(i, j int) bool {
        return providers[i].String() < providers[j].String()
    })
    return providers
}
```

### Files to Modify

#### 1. `code_agent/models.go` (Major Refactoring)

**Changes needed:**
- Separate model definitions from provider bindings
- Implement provider-based aliases
- Eliminate `-vertex` suffixed models

```go
// Restructure NewModelRegistry()
func NewModelRegistry() *ModelRegistry {
    registry := &ModelRegistry{
        models:     make(map[string]ModelConfig),
        aliases:    make(map[string]string),  // NEW
        modelsByProvider: make(map[string][]string),  // NEW
    }

    // Define base models ONCE
    registry.registerBaseModel(ModelConfig{
        ID:            "gemini-2.5-flash",
        Name:          "Gemini 2.5 Flash",
        // ... rest of config
    })
    
    // REGISTER for each provider
    registry.registerModelForProvider(ProviderGemini, "gemini-2.5-flash",
        []string{"2.5-flash", "flash"})  // Shorthand aliases
    registry.registerModelForProvider(ProviderVertexAI, "gemini-2.5-flash",
        []string{"2.5-flash", "flash"})
    
    // No more "gemini-2.5-flash-vertex" duplicates!
    
    return registry
}

// New methods
func (mr *ModelRegistry) registerModelForProvider(
    provider Provider,
    baseModelID string,
    shorthands []string,
) error {
    // Verify base model exists
    if _, exists := mr.models[baseModelID]; !exists {
        return fmt.Errorf("base model %q not found", baseModelID)
    }
    
    // Register provider/fullid → baseModelID
    key := fmt.Sprintf("%s/%s", provider, baseModelID)
    mr.aliases[key] = baseModelID
    
    // Register provider/shorthand → baseModelID
    for _, shorthand := range shorthands {
        key := fmt.Sprintf("%s/%s", provider, shorthand)
        mr.aliases[key] = baseModelID
    }
    
    // Track models by provider
    mr.modelsByProvider[provider.String()] = append(
        mr.modelsByProvider[provider.String()],
        baseModelID,
    )
}
```

---

## Phase 3: Enhanced UX (Display & Help)

### Files to Modify

#### 1. `code_agent/cli.go`

**Add new function:**

```go
// printProvidersList displays available providers and their models
func printProvidersList(renderer *display.Renderer, registry *ModelRegistry) {
    fmt.Print("\n" + renderer.Cyan("════════════════════════════════════════════════════════════════\n"))
    fmt.Print(renderer.Cyan("                    Available Providers & Models\n"))
    fmt.Print(renderer.Cyan("════════════════════════════════════════════════════════════════\n") + "\n")

    // Display each provider
    for _, provider := range AllProviders() {
        meta := GetProviderMetadata(provider)
        status := "✓ Configured"
        if !meta.IsConfigured {
            status = "⚠️  Not configured"
        }
        
        fmt.Printf("%s %s\n", meta.Icon, renderer.Bold(meta.DisplayName))
        fmt.Printf("   %s\n\n", meta.Description)
        
        // List models for this provider
        models := registry.GetProviderModels(provider.String())
        for _, model := range models {
            icon := "○"
            if model.IsDefault {
                icon = "✓"
            }
            fmt.Printf("   %s %s - %s\n",
                icon,
                renderer.Bold(fmt.Sprintf("%s/%s", provider, model.ID)),
                model.Description,
            )
        }
        
        fmt.Print("\n")
    }
}
```

---

## Phase 4: Documentation & Deprecation Warnings

### Files to Modify

#### 1. `code_agent/cli.go`

**Add deprecation warning for old syntax:**

```go
// In ParseCLIFlags(), when old flags detected
if *backend != "" && *model == "" {
    fmt.Fprintf(os.Stderr, "%s Old syntax: --backend %s\n",
        renderer.Yellow("⚠️  Tip:"),
        *backend)
    fmt.Fprintf(os.Stderr, "%s New syntax: --model %s/...\n\n",
        renderer.Yellow("   Try:"),
        *backend)
}
```

---

## Test Plan

### Unit Tests

```go
// test_cli_parsing.go (NEW)
func TestParseProviderModelSyntax(t *testing.T) {
    tests := []struct {
        input    string
        provider string
        model    string
        wantErr  bool
    }{
        {"gemini/2.5-flash", "gemini", "2.5-flash", false},
        {"vertexai/flash", "vertexai", "flash", false},
        {"flash", "", "flash", false},
        {"/flash", "", "", true},
        {"a/b/c", "", "", true},
    }
    // ... test cases
}

func TestModelResolution(t *testing.T) {
    registry := NewModelRegistry()
    
    tests := []struct {
        provider    string
        model       string
        expected    string
        wantErr     bool
    }{
        {"gemini", "2.5-flash", "gemini-2.5-flash", false},
        {"gemini", "flash", "gemini-2.5-flash", false},  // Latest
        {"vertexai", "flash", "gemini-2.5-flash", false},
        {"unknown", "any", "", true},
    }
    // ... test cases
}
```

### Integration Tests

```bash
# Test new syntax
code-agent --model gemini/2.5-flash /help
code-agent --model vertexai/1.5-pro /help

# Test backward compat
code-agent --backend gemini --model gemini-2.5-flash /help
code-agent --backend vertexai --model gemini-2.5-flash /help

# Test error handling
code-agent --model unknown/model  # Should show suggestions
```

---

## Timeline

**Total Timeline:** 2-3 weeks (significantly faster without backward compat concerns)

### Week 1: Core Implementation
- ✅ Implement `ParseProviderModelSyntax()` parsing
- ✅ Update `cli.go` to remove `--backend` flag
- ✅ Simplify flag structure (no legacy handling)
- ✅ Test CLI parsing thoroughly

### Week 2: Registry Refactoring
- ✅ Create `provider.go` module
- ✅ Refactor `ModelRegistry` with clean alias system
- ✅ Remove duplicate model definitions
- ✅ Update `/providers` command display

### Week 3: Polish & Release
- ✅ Enhanced error messages
- ✅ Full test coverage
- ✅ Documentation updates
- ✅ Ready for production release

---

## Success Criteria

- [ ] `--model gemini/2.5-flash` works
- [ ] `--model gemini/flash` works (shorthand)
- [ ] `--model flash` works (default provider)
- [ ] `/providers` shows clean hierarchy
- [ ] All old syntax still works
- [ ] Error messages are helpful
- [ ] No model duplication in registry
- [ ] Test coverage > 90%

---

## Rollback Plan

Since we're not supporting legacy syntax, there are no complex compatibility concerns:

1. If issues arise, simply revert the commits (clean git history)
2. No database migrations needed (in-memory registry only)
3. No feature flags needed to support both syntaxes
4. Users will use the new syntax from release day one

---

## Open Questions for Implementation

1. **Shorthand Alias Strategy**
   - Should `gemini/flash` → latest flash version?
   - Or should each version have an alias (gemini/2.5, gemini/2.0)?

2. **Default Provider in Config**
   - Should users be able to set `default_provider: vertexai` in config?
   - Or is `--backend gemini` sufficient?

3. **Provider Ambiguity Handling**
   - If user types `--model pro` and both providers have a `pro`:
     - Option A: Use default provider
     - Option B: Show interactive menu
     - Option C: Error with suggestions

4. **Model Version Matching**
   - Should `--model 2.5-flash` work (version-only)?
   - Or only `--model gemini/2.5-flash`?

---

## Related Documentation

- `MODEL_SELECTION_REDESIGN.md` - Design rationale
- `PROVIDER_MODEL_VISUAL_EXAMPLES.md` - UX mockups and examples

