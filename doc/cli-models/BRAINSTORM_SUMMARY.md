# Model Selection & Provider Architecture - Complete Brainstorm Summary

**Date:** November 11, 2025  
**Topic:** Improving CLI model/provider selection  
**Status:** Design Phase Complete

---

## Executive Summary

This brainstorm proposes a **provider-first architecture** for model selection in the Code Agent CLI. The key innovation is adopting a clean `provider/model` syntax that makes provider selection intuitive and extensible.

### The Problem
- Current CLI requires two separate flags (`--backend` and `--model`) to specify provider + model
- Model IDs are duplicated in registry (e.g., `gemini-2.5-flash` and `gemini-2.5-flash-vertex`)
- No visual hierarchy showing which models belong to which provider
- Adding new providers (OpenAI, Claude, etc.) requires multiple new flags
- Users must remember provider-specific model ID suffixes

### The Solution
Replace verbose multi-flag approach with provider/model syntax:

**Current (Verbose):**
```bash
code-agent --backend vertexai --project my-proj --location us-central1 --model gemini-1.5-pro-vertex
```

**Proposed (Clear & Intuitive):**
```bash
code-agent --model vertexai/1.5-pro --project my-proj --location us-central1
```

---

## Three Brainstorm Documents Created

### 1. **MODEL_SELECTION_REDESIGN.md**
Comprehensive design document covering:
- **Current state analysis** - What works, what doesn't
- **Provider-first architecture** - Conceptual foundation
- **Unified model registry** - Eliminate duplicates
- **Enhanced CLI syntax** - Multiple ways to specify models
- **New commands** - `/providers` for discovery
- **Clear resolution logic** - Deterministic precedence
- **Backward compatibility** - Old syntax still works
- **Design questions** - Open decisions needing discussion

**Key sections:**
- Problem identification
- Proposed solutions with code examples
- Implementation phases (4 phases)
- Benefits table
- Summary of wins

### 2. **PROVIDER_MODEL_VISUAL_EXAMPLES.md**
Concrete examples showing:
- **Before/after CLI syntax** - Current vs. proposed
- **Display mockups** - How `/providers` command looks
- **State diagrams** - Model resolution flow
- **Implementation examples** - Real usage patterns
- **Error handling** - Helpful error messages with suggestions
- **Backward compatibility examples** - What still works
- **Shorthand alias tables** - Available shortcuts
- **Design decisions** - Rationale for choices
- **Future extensions** - How to add new providers

**Key sections:**
- CLI syntax comparison
- Interactive display changes
- Real-world usage scenarios
- Error message examples
- Extensibility for future providers

### 3. **IMPLEMENTATION_ROADMAP.md**
Step-by-step implementation guide:
- **Phase 1 (MVP)** - Minimal changes, full backward compat
  - Add `ParseProviderModelSyntax()` function
  - Update help text
  - Display provider info
  
- **Phase 2** - Registry refactoring
  - Create `provider.go`
  - New alias system
  - Eliminate duplicates
  
- **Phase 3** - Enhanced UX
  - New `/providers` command
  - Better error messages
  - Interactive selection
  
- **Phase 4** - Polish & docs
  - Deprecation warnings
  - Documentation updates
  - Full test suite

**Key sections:**
- Specific code changes needed
- File-by-file modifications
- Test plan and success criteria
- Rollback strategy
- Timeline (4 weeks)

---

## Key Design Decisions

### 1. Syntax Choice: `provider/model`

**Why this syntax?**
- ✅ Familiar to developers (DNS, file paths, package managers)
- ✅ Prevents ambiguity with dashes in model names
- ✅ Scales to multiple providers naturally
- ✅ No new flags needed - works with existing `--model` flag

**Alternatives considered:**
- `provider:model` - Too SQL-like, less familiar
- `provider-model` - Ambiguous with dashes
- `provider.model` - Less common in CLI tools

### 2. Provider as First-Class Concept

**Why separate providers?**
- ✅ Current `--backend` is hidden in implementation details
- ✅ New providers (OpenAI, Claude) need clear representation
- ✅ Provider-specific requirements can be surfaced (e.g., GOOGLE_CLOUD_PROJECT)
- ✅ Enables future features like provider selection in config files

### 3. Model Alias Levels

```
Level 3 (Internal): gemini-2.5-flash (full model ID)
Level 2 (Explicit): gemini/2.5-flash (provider/full)
Level 1 (Short):    gemini/flash     (provider/alias)
Level 0 (Ultra):    flash            (uses default provider)
```

**Why this hierarchy?**
- ✅ Supports different user skill levels
- ✅ Experienced users can type minimally
- ✅ New users see full explicit form
- ✅ Backward compatible with existing IDs

### 4. Registry Restructuring

**From:**
```
Model Registry
├─ gemini-2.5-flash (Gemini API)
├─ gemini-2.5-flash-vertex (Vertex AI)
├─ gemini-1.5-pro
├─ gemini-1.5-pro-vertex
└─ ...
```

**To:**
```
Base Models
├─ gemini-2.5-flash
├─ gemini-1.5-pro
└─ ...

Provider Aliases
├─ gemini/2.5-flash → gemini-2.5-flash
├─ gemini/flash → gemini-2.5-flash
├─ vertexai/2.5-flash → gemini-2.5-flash
├─ vertexai/flash → gemini-2.5-flash
└─ ...
```

**Why this structure?**
- ✅ Single source of truth per model
- ✅ Multiple access paths per provider
- ✅ Easy to add new providers
- ✅ Clean separation of concerns

---

## Implementation Strategy

### Three-Layer Rollout

**Layer 1: Parsing (Week 1)**
- Add `ParseProviderModelSyntax()` function
- Update CLI help text
- No breaking changes
- All existing flags still work

**Layer 2: Registry (Weeks 2-3)**
- Refactor `ModelRegistry` with aliases
- Create `provider.go` module
- Eliminate duplicate model definitions
- Add `/providers` command

**Layer 3: Polish (Week 4)**
- Deprecation warnings for old syntax
- Enhanced error messages
- Full documentation
- Test coverage

---

## Impact Analysis

### Users Benefit From:
- ✅ **Clarity** - Provider and model are visually paired
- ✅ **Discoverability** - `/providers` shows hierarchical structure
- ✅ **Consistency** - Same syntax works for Gemini and future providers
- ✅ **Efficiency** - Shorthand aliases reduce typing
- ✅ **Friendliness** - Better error messages with suggestions

### Codebase Benefits From:
- ✅ **No duplication** - Models defined once, accessed multiple ways
- ✅ **Extensibility** - New providers don't need new flags
- ✅ **Clarity** - Provider/model relationship explicit in code
- ✅ **Maintainability** - Provider config centralized
- ✅ **Testability** - Alias resolution can be unit tested

### Clean Architecture (No Legacy Support):
- ✅ **Simpler Codebase** - No backward compat branches
- ✅ **Clearer Mental Model** - One way to do things
- ✅ **Faster Implementation** - 2-3 weeks vs 4+ weeks
- ✅ **Zero Technical Debt** - No old syntax baggage
- ✅ **Easy Maintenance** - Cleaner code from day one

---

## Future Extensibility

### Adding New Providers (e.g., OpenAI)

**Current approach would require:**
```bash
code-agent --backend openai --api-key=sk-... --model gpt-4
```

**With new design, just works:**
```bash
code-agent --model openai/gpt-4
code-agent --model openai/4
code-agent --model gpt-4    # With default provider = openai
```

### Adding New Model Versions

**Gemini 3.0 Ultra release:**
```bash
code-agent --model gemini/3.0-ultra
code-agent --model gemini/ultra  # Shorthand
```

**No code changes needed** - just register in model registry.

### Configuration File Support

```yaml
# ~/.code_agent/config.yaml (future)
default_provider: gemini
default_model: 2.5-flash

providers:
  gemini:
    api_key: ${GOOGLE_API_KEY}
  
  vertexai:
    project: my-gcp-project
    location: us-central1
  
  openai:
    api_key: ${OPENAI_API_KEY}
```

---

## Comparison with Current Architecture

| Feature | Current | Proposed |
|---------|---------|----------|
| **Model Selection** | `--backend X --model Y` | `--model provider/model` |
| **Shorthand** | None | Supported |
| **Model Definitions** | Duplicated per provider | Single source of truth |
| **Provider Visibility** | Implicit (backend term) | Explicit (provider term) |
| **Discovery** | `/models` flat list | `/providers` hierarchical |
| **New Provider Setup** | Multiple new flags | Just add to registry |
| **Error Messages** | Generic | Provider/model aware |
| **Backward Compat** | N/A | 100% preserved |
| **User Mental Model** | "backend" technical term | "provider" domain-specific |

---

## Next Steps

### Immediate (For Discussion)
1. ✅ Review the brainstorm documents
2. ✅ Validate design decisions
3. ✅ Approve clean redesign (no legacy support)

### Short Term (Next 2-3 weeks)
1. Finalize design decisions
2. Create GitHub issues for each phase
3. Begin implementation
4. Complete all phases

### Medium Term (1 month)
1. Comprehensive testing
2. Update documentation
3. Prepare release notes
4. Production release

---

## Related Documents

| Document | Purpose | Audience |
|----------|---------|----------|
| `MODEL_SELECTION_REDESIGN.md` | Design rationale & architecture | Architects, tech leads |
| `PROVIDER_MODEL_VISUAL_EXAMPLES.md` | UX mockups & usage examples | UX designers, product team |
| `IMPLEMENTATION_ROADMAP.md` | Code changes & implementation plan | Developers, QA |
| **This document** | Executive summary & overview | Everyone |

---

## Success Metrics

After implementation, we should measure:

✅ **Adoption**
- % of users using new syntax vs. old
- Command length reduction in logs
- CLI flag parsing error reduction

✅ **Satisfaction**
- User feedback on model selection UX
- Support ticket reduction for model-related issues
- Time-to-first-model selection

✅ **Maintainability**
- Code duplication reduction
- Number of files touched for new provider
- Test coverage increase

---

## Open Questions (For Discussion)

1. **Shorthand Scope**
   - Should `--model pro` search across all providers or just default?
   - Should we show an interactive menu if ambiguous?

2. **Model Aliasing**
   - How many aliases should each model have?
   - Who decides what aliases are "official"?

3. **Default Provider**
   - Should default provider be configurable?
   - Or always "gemini" for backward compat?

4. **Error Messages**
   - How verbose should "did you mean?" suggestions be?
   - Should we suggest `--project` and `--location` flags in errors?

5. **Backward Compat Timeline**
   - When should we deprecate old `--backend` syntax?
   - How long should we support both before removing?

---

## Conclusion

This redesign elevates **provider** from an implicit implementation detail to a first-class concept in the user's mental model. The result is:

- 🎯 **More intuitive** - Users understand provider/model relationship
- 📦 **Better structured** - No duplication, single source of truth
- 🚀 **Extensible** - New providers fit naturally
- 🔄 **Backward compatible** - Zero breaking changes
- ✨ **Future-proof** - Scales with new models and providers

The provider/model syntax is a small change with big benefits for both users and the codebase.

---

**Created by:** Design Brainstorm Session  
**Date:** November 11, 2025  
**Status:** Ready for review and discussion
