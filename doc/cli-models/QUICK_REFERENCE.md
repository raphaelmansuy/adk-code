# Quick Reference: Provider/Model Selection Redesign

## One-Sentence Summary
Replace `--backend X --model Y` with intuitive `--model provider/model` syntax, treating providers as first-class concepts.

---

## Problem Statement
```
Current:  code-agent --backend vertexai --model gemini-1.5-pro-vertex
Issue:    Two flags, suffixed model IDs, duplicated definitions, not extensible

Better:   code-agent --model vertexai/1.5-pro
```

---

## Core Concepts

### Providers
- **gemini** - Google Gemini REST API (requires GOOGLE_API_KEY)
- **vertexai** - GCP-native endpoint (requires GOOGLE_CLOUD_PROJECT, GOOGLE_CLOUD_LOCATION)
- **Future** - openai, claude, anthropic, etc. (just add to registry)

### Model Syntax Levels

| Input | Resolves To | Use Case |
|-------|------------|----------|
| `gemini/gemini-2.5-flash` | Full ID (internal) | System code |
| `gemini/2.5-flash` | Explicit with version | Users who know versions |
| `gemini/flash` | Latest flash model | Users who want latest |
| `flash` | Latest flash (default provider) | Experienced users |

### Display Hierarchy

```
Provider 1 (gemini)
├─ Model A
│  ├─ Full ID: gemini-2.5-flash
│  ├─ Short: gemini/2.5-flash
│  └─ Alias: gemini/flash
├─ Model B
└─ Model C

Provider 2 (vertexai)
├─ Model A (same base as gemini's)
│  ├─ Full ID: gemini-2.5-flash (Vertex AI backend)
│  ├─ Short: vertexai/2.5-flash
│  └─ Alias: vertexai/flash
└─ Model B
```

---

## Proposed Changes

### CLI Syntax Examples

**Old (Current)**
```bash
# Gemini
code-agent --backend gemini --model gemini-2.5-flash

# Vertex AI
code-agent --backend vertexai --project proj --location loc --model gemini-1.5-pro-vertex
```

**New (Proposed)**
```bash
# Gemini
code-agent --model gemini/2.5-flash
code-agent --model gemini/flash    # Shorthand
code-agent --model flash           # Ultra-short with default provider

# Vertex AI
code-agent --model vertexai/2.5-flash --project proj --location loc
code-agent --model vertexai/flash --project proj --location loc
```

### New Commands

**`/providers`** - Show providers and their models
```
🔷 Gemini API (gemini/)
   ✓ gemini/2.5-flash  - Default, latest
   ○ gemini/1.5-pro
   ○ gemini/1.5-flash

🔶 Vertex AI (vertexai/)
   ○ vertexai/2.5-flash
   ○ vertexai/1.5-pro
```

### New Modules/Changes

- **Create: `provider.go`** - Provider definitions, metadata
- **Modify: `models.go`** - Registry refactoring, aliases
- **Modify: `cli.go`** - Add `ParseProviderModelSyntax()`, update help
- **Modify: `main.go`** - Use new parser, better errors
- **Add tests** - Unit tests for parsing and resolution

---

## Resolution Algorithm

```
Input: --model <string>

1. Parse string:
   Contains "/"?
     YES → Split into (provider, model)
     NO  → provider = "", model = string

2. Determine provider:
   If provider specified:
     Use it
   Else:
     Use default provider (usually "gemini")

3. Find model in provider:
   Try exact match
     Found → Return model
   Try shorthand/alias
     Found → Return model
   Not found → Error with suggestions

4. Check provider requirements:
   Gemini? Check GOOGLE_API_KEY
   Vertex AI? Check GOOGLE_CLOUD_PROJECT, GOOGLE_CLOUD_LOCATION
   Missing → Error with setup instructions

5. Create LLM → Done ✓
```

---

## Backward Compatibility

**None required** - this is a clean redesign, not a migration:

```bash
# Old syntax (will NOT be supported)
code-agent --backend gemini --model gemini-2.5-flash  # ❌ Removed
code-agent --backend vertexai --project proj --model gemini-1.5-pro-vertex  # ❌ Removed

# New syntax only (clean from day one)
code-agent --model gemini/2.5-flash  # ✓ Only supported
code-agent --model vertexai/1.5-pro --project my-proj --location us-central1  # ✓ Only supported
```

**Benefits:**
- ✅ No technical debt from supporting two syntaxes
- ✅ Cleaner code with no legacy branches
- ✅ Simpler user mental model (one way to do things)
- ✅ Faster implementation (2-3 weeks vs 4 weeks)

---

## Benefits

### For Users
✅ Intuitive syntax (provider/model is clear)  
✅ Shorthand aliases (less typing)  
✅ Better discovery (`/providers` command)  
✅ Helpful error messages with suggestions  
✅ Extensible to future providers  

### For Codebase
✅ No model duplication (single definition per model)  
✅ No new flags needed for new providers  
✅ Provider metadata centralized  
✅ Clean separation of concerns  
✅ Easier testing (alias resolution is pure function)  

---

## Implementation Phases

### Phase 1 (1 week) - Core Implementation
- Add `ParseProviderModelSyntax()` function
- Remove `--backend` flag from CLI
- Simplify flag parsing (no legacy handling)
- Update help text with new syntax examples

### Phase 2 (1 week) - Refactoring
- Create `provider.go` module
- Refactor `ModelRegistry` with clean alias system
- Remove duplicate model definitions
- Add `/providers` command

### Phase 3 (0.5-1 week) - Polish
- Enhanced error messages
- Full test coverage
- Documentation updates
- Ready for production

---

## Open Design Questions

1. **Shorthand Ambiguity**
   - `--model pro` when both gemini and vertexai have pro models?
   - Option A: Use default provider only ← Recommended
   - Option B: Show interactive menu
   - Option C: Error with suggestions

2. **Model Alias Count**
   - How many aliases per model?
   - `gemini/2.5-flash` + `gemini/flash` enough?
   - Or also `gemini/25`, `gemini/latest-flash`?

3. **Default Provider**
   - Always "gemini"?
   - Configurable in config file?
   - Auto-detect from environment?

4. **Model Version Matching**
   - `--model 2.5-flash` → find in default provider?
   - Or only full `--model gemini/2.5-flash`?

---

## Success Criteria

- [ ] `--model provider/model` syntax works
- [ ] `--model provider/shorthand` works
- [ ] `--model shorthand` works (default provider)
- [ ] `/providers` shows clean hierarchy
- [ ] All old syntax still works
- [ ] Error messages are helpful
- [ ] No model duplication
- [ ] Test coverage > 90%

---

## Future Extensions

### New Provider Support
```bash
# When added to registry, just works:
code-agent --model openai/gpt-4
code-agent --model claude/3.5-sonnet

# No new flags needed!
```

### New Model Versions
```bash
# When Gemini 3.0 released:
code-agent --model gemini/3.0-ultra

# Alias can point to it:
code-agent --model gemini/ultra
```

### Config File Support
```yaml
default_provider: gemini
default_model: 2.5-flash

providers:
  gemini:
    api_key: ${GOOGLE_API_KEY}
  vertexai:
    project: ${GCP_PROJECT}
    location: us-central1
```

---

## Key Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Separator | `/` | Familiar, intuitive, filesystem-like |
| Provider Term | "provider" not "backend" | More user-friendly, domain-specific |
| Registry Structure | Single model + aliases | Eliminate duplication |
| Resolution Precedence | Explicit > default | Principle of least surprise |
| Backward Compat | 100% preserved | Smooth migration path |

---

## Related Documents

1. **MODEL_SELECTION_REDESIGN.md** - Full design rationale (15+ pages)
2. **PROVIDER_MODEL_VISUAL_EXAMPLES.md** - UX mockups and examples (12+ pages)
3. **IMPLEMENTATION_ROADMAP.md** - Code changes and timeline (16+ pages)
4. **BRAINSTORM_SUMMARY.md** - Executive summary (9+ pages)

---

## Getting Started

### To Review This Design
1. Read this quick reference first
2. Check BRAINSTORM_SUMMARY.md for high-level overview
3. Review PROVIDER_MODEL_VISUAL_EXAMPLES.md for UX
4. Study MODEL_SELECTION_REDESIGN.md for architecture
5. Reference IMPLEMENTATION_ROADMAP.md for code details

### To Implement
1. Create GitHub issues for each phase
2. Start with Phase 1 (add parsing function)
3. Maintain backward compat throughout
4. Test extensively at each phase
5. Gather user feedback and iterate

### To Discuss
- Which design decisions need clarification?
- Are open questions resolved?
- Any concerns about backward compatibility?
- Timeline realistic?

---

## Contacts & Ownership

**Brainstorm Authors:** Code Agent Design Session  
**Date:** November 11, 2025  
**Status:** Ready for review and implementation planning

**Next Step:** Team discussion on design decisions and implementation timeline.

---

## Appendix: Syntax Examples

```bash
# All of these should work after implementation:

# Explicit provider/model
code-agent --model gemini/2.5-flash
code-agent --model vertexai/1.5-pro

# Shorthand (latest for type)
code-agent --model gemini/flash
code-agent --model vertexai/pro

# Ultra-short (with default provider)
code-agent --model flash
code-agent --model pro

# Backward compatible (old syntax)
code-agent --backend gemini --model gemini-2.5-flash
code-agent --backend vertexai --project proj --model gemini-1.5-pro-vertex

# Discovery
code-agent /providers        # Show all providers & models
code-agent /help            # Show help (updated with new syntax)
```

---

**Version:** 1.0  
**Last Updated:** November 11, 2025  
**Status:** Design Brainstorm Complete ✓
