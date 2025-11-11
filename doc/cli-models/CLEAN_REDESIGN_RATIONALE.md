# Provider/Model Selection - Clean Redesign (No Legacy Support)

**Date:** November 11, 2025  
**Update:** Documents updated to reflect clean redesign approach  
**Status:** Ready for implementation

---

## Key Change: No Backward Compatibility Required

The brainstorm documents have been updated to remove all backward compatibility constraints. This enables a **cleaner, simpler, faster implementation**.

---

## What Changed

### Old Approach (Not Used)
- Support `--backend gemini|vertexai` flag alongside new syntax
- Keep duplicate `-vertex` suffix models in registry
- Handle both old and new CLI patterns
- Timeline: 4 weeks with compatibility code

### New Approach (Cleaner)
- **Only** `--model provider/model` syntax
- Single model definitions with provider aliases
- No legacy CLI branches
- Timeline: 2-3 weeks with clean code

---

## Benefits of Clean Redesign

### Simpler Codebase
```go
// OLD: Complex flag handling
func ParseCLIFlags() {
    backend := flag.String("backend", "gemini", ...)
    model := flag.String("model", "gemini-2.5-flash", ...)
    // ... complex parsing logic ...
    // ... backward compat checks ...
}

// NEW: Single, clear syntax
func ParseCLIFlags() {
    model := flag.String("model", "", 
        "Model to use (e.g., gemini/2.5-flash, vertexai/1.5-pro)")
    // ... simple parsing: provider/model ...
}
```

### No Model Duplication
```go
// OLD: Duplicate definitions
models["gemini-2.5-flash"]        // Gemini API
models["gemini-2.5-flash-vertex"] // Vertex AI

// NEW: Single definition + provider aliases
models["gemini-2.5-flash"]  // Base definition
aliases["gemini/2.5-flash"]     → gemini-2.5-flash (Gemini API)
aliases["vertexai/2.5-flash"]   → gemini-2.5-flash (Vertex AI)
```

### Clearer User Model
```bash
# OLD: Two concepts needed
code-agent --backend vertexai --model gemini-1.5-pro-vertex
# Users must understand: "backend" is different from "model"
# Users must know "-vertex" suffix convention

# NEW: Single concept
code-agent --model vertexai/1.5-pro
# Users see: provider/model - intuitive pairing
# No suffixes to remember
```

### Faster Implementation
- **Phase 1 (1 week):** Core parsing + flag cleanup
- **Phase 2 (1 week):** Registry refactoring
- **Phase 3 (0.5-1 week):** Polish + testing
- **Total:** 2-3 weeks vs 4+ weeks

### Zero Technical Debt
- No legacy branches in code
- No deprecation warnings to maintain
- No migration timeline concerns
- Clean codebase from release day one

---

## Updated Documents

All brainstorm documents have been updated:

### ✅ MODEL_SELECTION_REDESIGN.md
- Replaced backward compat section with "Clean Architecture"
- Removed references to supporting old flags
- Simplified design rationale

### ✅ IMPLEMENTATION_ROADMAP.md
- Changed Phase 1 goal to "Clean Implementation"
- Removed "no breaking changes" language
- Updated code examples (simpler parsing)
- Updated timeline: 2-3 weeks instead of 4
- Simplified rollback plan

### ✅ QUICK_REFERENCE.md
- Updated backward compatibility section to show clean approach
- Simplified implementation phases (3 instead of 4)
- Clearer messaging about no legacy support

### ✅ BRAINSTORM_SUMMARY.md
- Updated next steps timeline
- Changed "migration path" to "clean architecture"
- Added benefits of no legacy support
- Simplified success metrics

---

## CLI Changes Summary

### What's Removed
```bash
# These will NOT be supported
code-agent --backend gemini --model gemini-2.5-flash  # ❌
code-agent --backend vertexai --project proj --model gemini-1.5-pro-vertex  # ❌
code-agent -api-key=sk-...  # ❌ Use config or env vars
code-agent --project my-proj  # ❌ Use config or env vars
code-agent --location us-central1  # ❌ Use config or env vars
```

### What's Added
```bash
# Only supported syntax
code-agent --model gemini/2.5-flash
code-agent --model gemini/flash  # Shorthand
code-agent --model vertexai/1.5-pro --project my-proj --location us-central1
code-agent --model openai/gpt-4  # Future providers just work
```

### Configuration
```bash
# Credentials via environment variables
export GOOGLE_API_KEY=sk-...  # Gemini API
export GOOGLE_CLOUD_PROJECT=my-proj  # Vertex AI
export GOOGLE_CLOUD_LOCATION=us-central1  # Vertex AI
export OPENAI_API_KEY=sk-...  # Future OpenAI support

# No `--api-key`, `--project`, `--location` flags needed
```

---

## Model Registry Changes

### Current (Has Duplication)
```go
models := map[string]ModelConfig{
    "gemini-2.5-flash":        {Backend: "gemini", ...},
    "gemini-2.5-flash-vertex": {Backend: "vertexai", ...},  // Duplicate!
    "gemini-1.5-pro":          {Backend: "gemini", ...},
    "gemini-1.5-pro-vertex":   {Backend: "vertexai", ...},  // Duplicate!
}
```

### New (Clean)
```go
models := map[string]ModelConfig{
    "gemini-2.5-flash": {Capabilities: {...}, ...},
    "gemini-1.5-pro":   {Capabilities: {...}, ...},
    // No duplication!
}

aliases := map[string]string{
    "gemini/2.5-flash":  "gemini-2.5-flash",
    "gemini/flash":      "gemini-2.5-flash",  // Latest flash
    "vertexai/2.5-flash": "gemini-2.5-flash",
    "vertexai/flash":     "gemini-2.5-flash",
    // Multiple paths to same model, single definition
}
```

---

## Implementation Approach

### Week 1: Core Implementation
1. Remove `--backend` flag from CLI
2. Implement `ParseProviderModelSyntax("vertexai/2.5-flash")`
3. Simplify `CLIConfig` struct (no backend field)
4. Test parsing thoroughly

### Week 2: Registry Refactoring
1. Create `provider.go` with provider definitions
2. Refactor `ModelRegistry` with alias system
3. Remove duplicate `-vertex` suffix models
4. Update `/providers` command

### Week 3: Polish & Release
1. Error messages for invalid syntax
2. Comprehensive test coverage
3. Update documentation and help text
4. Ready for production

---

## Example Implementation: Clean Parsing

```go
// Parse "vertexai/1.5-pro" → provider, model, error
func ParseProviderModelSyntax(input string) (string, string, error) {
    parts := strings.Split(input, "/")
    
    if len(parts) != 2 {
        return "", "", fmt.Errorf(
            "invalid model syntax: %q (use provider/model like gemini/2.5-flash)",
            input)
    }
    
    provider := parts[0]
    model := parts[1]
    
    // Validate provider exists
    if !isValidProvider(provider) {
        return "", "", fmt.Errorf(
            "unknown provider: %q\n\nAvailable providers:\n%s",
            provider, listProviders())
    }
    
    return provider, model, nil
}

// In main.go, much simpler:
provider, modelID, err := ParseProviderModelSyntax(cliConfig.Model)
if err != nil {
    log.Fatal(err)
}

selectedModel, err := modelRegistry.GetByProviderAndModel(provider, modelID)
if err != nil {
    log.Fatal(err)
}

// Create LLM based on provider
switch provider {
case "gemini":
    // Create Gemini API client
case "vertexai":
    // Create Vertex AI client
default:
    // Should not happen, validated in ParseProviderModelSyntax
}
```

---

## Comparison: Old vs New

| Aspect | Old (4 weeks) | New (2-3 weeks) |
|--------|---------------|-----------------|
| **Syntax** | `--backend X --model Y` | `--model provider/model` |
| **Flags** | backend, api-key, project, location | model only |
| **Code Branches** | Dual parsing paths | Single path |
| **Model Definitions** | Duplicated (2 per base) | Single (1 per base) |
| **Complexity** | Medium | Low |
| **Backward Compat** | Full support needed | Not needed |
| **Mental Model** | "backend" + "model" | "provider/model" |
| **Extensibility** | Requires new flags | Just add provider |
| **Test Complexity** | Higher (2 paths) | Lower (1 path) |
| **Implementation Time** | 4+ weeks | 2-3 weeks |

---

## Q&A: Why This Approach?

**Q: What about users upgrading from old version?**  
A: Users get cleaner, simpler syntax from day one. No migration burden.

**Q: How do we communicate the change?**  
A: Release notes show new syntax. Help text is very clear. One way to do it.

**Q: Is there risk?**  
A: No. All syntax is validated in `ParseProviderModelSyntax()`. Errors are clear.

**Q: Can we add old syntax later if needed?**  
A: Yes, but unlikely to be needed. New syntax is intuitive and simple.

**Q: What about CI/CD pipelines using old flags?**  
A: Release notes clearly document the change. Teams update their scripts.

---

## Files Updated

✅ `doc/MODEL_SELECTION_REDESIGN.md` - Clean architecture section  
✅ `doc/IMPLEMENTATION_ROADMAP.md` - New timeline and simpler approach  
✅ `doc/QUICK_REFERENCE.md` - Updated implementation phases  
✅ `doc/BRAINSTORM_SUMMARY.md` - Updated next steps and timeline  

**Not yet updated (reference):**
- `doc/PROVIDER_MODEL_VISUAL_EXAMPLES.md` - Still valid (shows clean syntax)
- `doc/BRAINSTORM_INDEX.md` - Still valid (document structure unchanged)

---

## Next Steps

1. ✅ Review this clean redesign approach
2. ✅ Approve faster timeline (2-3 weeks)
3. ✅ Create GitHub issues for implementation
4. ✅ Begin Week 1: Core Implementation
5. ✅ Begin Week 2: Registry Refactoring
6. ✅ Begin Week 3: Polish & Release

---

## Summary

By removing backward compatibility constraints, we get:

- 🚀 **Faster implementation** (2-3 weeks vs 4+)
- 🧹 **Cleaner codebase** (no legacy branches)
- 🎯 **Clearer design** (one way to do things)
- 📦 **Easier maintenance** (simpler mental model)
- 🔧 **Better architecture** (provider-first, extensible)

**This is the right approach for v1.1.0 of Code Agent.**

---

**Status:** ✅ All documents updated for clean redesign  
**Timeline:** 2-3 weeks to production  
**Approved for:** Implementation  
