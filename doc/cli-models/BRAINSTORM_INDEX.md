# Brainstorm Index: Provider/Model Selection Architecture

**Date:** November 11, 2025  
**Topic:** CLI Model Selection & Provider Architecture Redesign  
**Status:** Complete Brainstorm - Ready for Review

---

## Brainstorm Deliverables

Four comprehensive documents have been created exploring the provider/model selection redesign:

### 1. QUICK_REFERENCE.md ⚡
**Length:** ~3 pages | **Audience:** Everyone  
**Purpose:** Quick overview and decision reference

**Key Content:**
- One-sentence problem/solution
- Core concepts and syntax levels
- Proposed CLI changes with examples
- Resolution algorithm diagram
- Success criteria checklist
- All syntax examples for quick lookup

**Use This For:** Quick answers, showing stakeholders, making decisions

---

### 2. BRAINSTORM_SUMMARY.md 📋
**Length:** ~9 pages | **Audience:** Decision makers, stakeholders  
**Purpose:** Executive summary with strategic context

**Key Content:**
- Problem statement and proposed solution
- Overview of all three detailed documents
- Key design decisions with rationale
- Implementation strategy (3 layers)
- Impact analysis (user benefits, code benefits, migration path)
- Future extensibility scenarios
- Comparison table with current architecture
- Success metrics to measure

**Use This For:** Presenting to stakeholders, strategic planning, approval

---

### 3. PROVIDER_MODEL_VISUAL_EXAMPLES.md 🎨
**Length:** ~12 pages | **Audience:** UX/Product, Developers  
**Purpose:** Visual designs, mockups, and concrete examples

**Key Content:**
- Before/after CLI syntax comparison
- Display mockup of `/providers` command output
- Model resolution state diagram
- Real-world implementation examples (successful and error cases)
- Backward compatibility examples
- Shorthand alias tables
- Design decision rationale for each choice
- Future extensions (new providers, versions)

**Use This For:** Understanding UX flow, user-facing design, acceptance criteria

---

### 4. MODEL_SELECTION_REDESIGN.md 🏗️
**Length:** ~15 pages | **Audience:** Architects, Senior Developers  
**Purpose:** Detailed architecture and design rationale

**Key Content:**
- Current state analysis (what works, limitations)
- Proposed provider-first architecture with code concepts
- Unified registry design to eliminate duplication
- Enhanced CLI syntax options (multiple approaches)
- New `/providers` command specification
- Clear model resolution logic with precedence
- 4-phase implementation plan
- Backward compatibility guarantees
- Open design questions needing discussion
- Benefits summary

**Use This For:** Architecture review, technical planning, code decisions

---

### 5. IMPLEMENTATION_ROADMAP.md 🛣️
**Length:** ~16 pages | **Audience:** Developers, QA  
**Purpose:** Step-by-step implementation guide with code details

**Key Content:**
- Phase 1 (MVP) - Minimal changes, full backward compat
  - Specific changes to cli.go, models.go, main.go
  - Code examples for each change
  
- Phase 2 (Refactoring) - Registry restructuring
  - New provider.go module creation
  - Detailed models.go refactoring
  - Alias system implementation
  
- Phase 3 (Enhanced UX) - Display improvements
  - New `/providers` command implementation
  
- Phase 4 (Polish) - Documentation and tests
  - Test plan with specific test cases
  - Rollback strategy
  - 4-week timeline breakdown
  
**Use This For:** Implementation planning, code review preparation, testing strategy

---

## Document Navigation Guide

### "What is this about?"
→ Start with **QUICK_REFERENCE.md** (3 pages)

### "Why should we do this?"
→ Read **BRAINSTORM_SUMMARY.md** (9 pages)

### "What will users see?"
→ Check **PROVIDER_MODEL_VISUAL_EXAMPLES.md** (12 pages)

### "How should we build this?"
→ Study **MODEL_SELECTION_REDESIGN.md** (15 pages)

### "What's the implementation plan?"
→ Reference **IMPLEMENTATION_ROADMAP.md** (16 pages)

### "I need everything at once"
→ Read all documents in order above

---

## Problem Summary

**Current State:**
```bash
code-agent --backend vertexai --project my-proj --location us-central1 --model gemini-1.5-pro-vertex
```

**Proposed State:**
```bash
code-agent --model vertexai/1.5-pro --project my-proj --location us-central1
```

**Why Better:**
- ✅ Clear provider/model pairing with `/` separator
- ✅ No model duplication in registry (single definition per model)
- ✅ Scales to multiple providers naturally
- ✅ Shorthand aliases reduce typing
- ✅ Better error messages with suggestions
- ✅ Extensible architecture for future providers

---

## Key Design Decisions

### 1. Separator Choice: `/`
- Familiar to developers (DNS, file paths, package managers)
- Prevents ambiguity with dashes in model names
- Intuitive hierarchy (provider/model)

### 2. Providers as First-Class Concept
- Current `--backend` is implementation detail, `provider` is user concept
- New providers fit naturally without new flags
- Provider requirements can be surfaced (e.g., GOOGLE_CLOUD_PROJECT)

### 3. Alias Levels
- **Level 3:** Full ID (gemini-2.5-flash) - internal
- **Level 2:** Explicit (gemini/2.5-flash) - users who know versions
- **Level 1:** Shorthand (gemini/flash) - latest of type
- **Level 0:** Ultra-short (flash) - default provider

### 4. Single Model Definition with Provider Aliases
```
Before: Model duplication (once per provider)
After: Single model + multiple alias paths
```

---

## Implementation Phases Overview

| Phase | Week | Task | Impact |
|-------|------|------|--------|
| **Phase 1** | W1 | Add parsing + help text | MVP ready, zero breaking changes |
| **Phase 2** | W2-3 | Registry refactoring | Eliminate duplication |
| **Phase 3** | W3-4 | Enhanced UX | New commands and displays |
| **Phase 4** | W4 | Polish + tests | Production ready |

**Total Timeline:** 4 weeks with no breaking changes throughout

---

## Backward Compatibility Guarantee

✅ **All existing syntax continues to work:**
```bash
code-agent --backend gemini --model gemini-2.5-flash
code-agent --backend vertexai --model gemini-1.5-pro-vertex
export GOOGLE_GENAI_USE_VERTEXAI=true && code-agent
```

✅ **Gradual migration path** - users can opt-in to new syntax at their pace

✅ **No database migrations or data changes** needed (in-memory registry only)

✅ **Safe rollback** - if issues arise, revert safely

---

## Success Metrics After Implementation

**Adoption:**
- % of users using new syntax vs. old
- Command-line argument length reduction
- Support ticket reduction for model selection

**Satisfaction:**
- User feedback on UX improvements
- Time-to-first-model-selection reduction
- Discoverability of model options

**Maintainability:**
- Code duplication reduction (eliminate -vertex suffix models)
- Files touched for new provider support (should decrease)
- Test coverage increase

---

## Open Questions (For Team Discussion)

1. **Shorthand Ambiguity Handling**
   - When user types `--model pro` and both Gemini and Vertex AI have `pro`:
     - Use default provider only? (Recommended)
     - Show interactive menu?
     - Error with suggestions?

2. **Model Alias Strategy**
   - How many aliases should each model have?
   - `gemini/2.5-flash` + `gemini/flash` sufficient?
   - Or also `gemini/25`, `gemini/latest-flash`?

3. **Default Provider Configurability**
   - Always "gemini" for backward compat?
   - Configurable in ~/.code_agent/config.yaml?
   - Auto-detect from environment?

4. **Model Version-Only Matching**
   - Should `--model 2.5-flash` work (without provider prefix)?
   - Or only `--model gemini/2.5-flash`?

5. **Deprecation Timeline**
   - When to deprecate old `--backend` syntax?
   - How long to support both?
   - Gradual or cut-off date?

---

## Extensibility Examples

### Adding OpenAI Provider (Future)
```bash
# Syntax automatically supports it - no new flags needed!
code-agent --model openai/gpt-4
code-agent --model openai/4      # Shorthand
code-agent --model gpt-4         # With default provider = openai

# Just register in ModelRegistry, display works automatically:
/providers → shows openai section
```

### Adding New Gemini Version
```bash
# When Google releases Gemini 3.0:
code-agent --model gemini/3.0-ultra
code-agent --model gemini/ultra    # Shorthand

# No code changes needed - just registry update
```

---

## Related Project Context

### Current Architecture Limitations
- `--backend` and `--model` separated (cognitive load)
- Model IDs duplicated in registry (maintenance burden)
- No visual hierarchy in model listing (discoverability)
- New providers require multiple flag additions (not extensible)

### Why This Matters
- Code Agent may expand to support more providers (OpenAI, Claude, etc.)
- User base will grow; simpler syntax reduces support burden
- Team will maintain code longer; cleaner design is maintainable
- First-class provider concept enables future features (config files, interactive selection)

---

## How to Use These Documents

### For Product Managers/Stakeholders
1. Read: QUICK_REFERENCE.md (5 min)
2. Read: BRAINSTORM_SUMMARY.md (15 min)
3. Decision: Approve for implementation? (Yes/No/Discuss)

### For Architects
1. Read: MODEL_SELECTION_REDESIGN.md (30 min)
2. Review: Key design decisions section
3. Answer: Open design questions
4. Feedback: Architecture sound? Extensible? Maintainable?

### For Developers
1. Read: IMPLEMENTATION_ROADMAP.md (30 min)
2. Review: Phase 1 specific code changes
3. Plan: How to tackle each phase
4. Test: Unit test strategy for alias resolution

### For QA/Testers
1. Skim: PROVIDER_MODEL_VISUAL_EXAMPLES.md (10 min)
2. Study: IMPLEMENTATION_ROADMAP.md test section (10 min)
3. Create: Test cases for each phase
4. Execute: Regression testing for backward compat

---

## Document Statistics

| Document | Pages | Words | Focus |
|----------|-------|-------|-------|
| QUICK_REFERENCE.md | 3 | ~1,500 | Decision reference |
| BRAINSTORM_SUMMARY.md | 9 | ~3,500 | Executive summary |
| PROVIDER_MODEL_VISUAL_EXAMPLES.md | 12 | ~4,500 | UX design |
| MODEL_SELECTION_REDESIGN.md | 15 | ~6,000 | Architecture |
| IMPLEMENTATION_ROADMAP.md | 16 | ~6,500 | Implementation |
| **Total** | **~55** | **~22,000** | **Complete design** |

---

## Next Steps

### Immediate (This Week)
- [ ] Distribute brainstorm documents to team
- [ ] Schedule architecture review meeting
- [ ] Collect feedback on design decisions
- [ ] Clarify open questions with stakeholders

### Short Term (Next Week)
- [ ] Finalize design decisions
- [ ] Create GitHub issues for each phase
- [ ] Assign implementation owners
- [ ] Plan Phase 1 sprint

### Medium Term (Weeks 2-4)
- [ ] Execute Phase 1 (MVP)
- [ ] Phase 2-4 implementation
- [ ] Comprehensive testing
- [ ] User documentation updates

### Long Term (Post-Release)
- [ ] Gather user feedback
- [ ] Measure success metrics
- [ ] Plan Phase 2 enhancements (config files, interactive selection)
- [ ] Document lessons learned

---

## Contact & Ownership

**Brainstorm Documents:** Design Review Session  
**Created:** November 11, 2025  
**Status:** ✅ Complete - Ready for Team Review

**Next Action:** Schedule architecture review with team

---

## Appendix: Document Map

```
Brainstorm Index (This Document)
├── QUICK_REFERENCE.md (⚡ Quick lookup)
│   └── For: Everyone, quick answers
│
├── BRAINSTORM_SUMMARY.md (📋 Executive)
│   └── For: Decision makers, stakeholders
│
├── PROVIDER_MODEL_VISUAL_EXAMPLES.md (🎨 UX Design)
│   └── For: Product, Designers, Developers
│
├── MODEL_SELECTION_REDESIGN.md (🏗️ Architecture)
│   └── For: Architects, Senior Engineers
│
└── IMPLEMENTATION_ROADMAP.md (🛣️ Code)
    └── For: Developers, QA, Implementation
```

**Total Package:** Complete design brainstorm with visual mockups, architecture, implementation plan, and reference guide.

---

**Ready to discuss and implement!** 🚀
