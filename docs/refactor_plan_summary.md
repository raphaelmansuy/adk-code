# Refactoring Plan - Executive Summary

**Project**: code_agent  
**Date**: November 12, 2025  
**Commitment**: 0% Regression Guaranteed

---

## TL;DR - What We're Doing

**Goal**: Make code more organized, modular, and maintainable WITHOUT breaking anything.

**Strategy**: 5 incremental phases, each fully tested before moving to the next.

**Timeline**: 15-20 hours over 3 weeks (1-2 hours/day)

---

## Current State Assessment

✅ **What's Good**:
- Clean 30-line main.go with proper separation
- Well-designed tool registry with auto-registration
- Solid session persistence layer
- All 30 tests passing
- No circular dependencies

⚠️ **What Needs Improvement**:
- Display package: 24 files, too many responsibilities
- Error handling: Inconsistent (custom pkg/errors exists but underused)
- Agent prompts: Scattered across 5 files
- Model providers: Mixed together, could be per-provider packages

📊 **By The Numbers**:
- 138 Go files total
- 30 test files (21.7% coverage)
- Largest file: 432 lines (session/sqlite.go - reasonable)
- Zero critical complexity issues

---

## The Plan (5 Phases)

### Phase 1: Error Handling (3-4 hours, LOW RISK)

**What**: Standardize on `pkg/errors` across all packages

**Why**: Currently only 2 packages use our custom error types. Most use `fmt.Errorf()`, which makes debugging and testing harder.

**How**:
1. Create error adoption guide
2. Update display package (8-10 files)
3. Update agent package (3-4 files)  
4. Update remaining packages (5-6 files)

**Tests**: All existing tests must pass, add error code validation tests

---

### Phase 2: Agent Prompt Organization (2-3 hours, LOW RISK)

**What**: Move 5 scattered prompt files into `agent/prompts/` subpackage

**Before**:
```
agent/
├── dynamic_prompt.go
├── xml_prompt_builder.go
├── prompt_guidance.go
├── prompt_pitfalls.go
└── prompt_workflow.go
```

**After**:
```
agent/
├── coding_agent.go
└── prompts/
    ├── dynamic.go
    ├── builder.go
    ├── guidance.go
    ├── pitfalls.go
    └── workflow.go
```

**Compatibility**: Add re-exports in old locations (deprecated) for 2 release cycles

**Tests**: All agent tests must pass, no behavioral changes

---

### Phase 3: Display Package Refactoring (4-5 hours, MEDIUM RISK)

**What**: Extract tool-related code (4 files, 1000+ LOC) into `display/tooling/` subpackage

**Why**: Display package has 24 files with mixed responsibilities. Tool integration logic doesn't belong at root.

**Moving** (with backward-compatible facades):
```
display/tool_adapter.go (188 lines)        → display/tooling/adapter.go
display/tool_renderer.go (276 lines)       → display/tooling/renderer.go
display/tool_result_parser.go (361 lines)  → display/tooling/parser.go
display/tool_renderer_internals.go         → display/tooling/internal.go
```

**Result**: Display package reduced from 24 to ~18 files, clearer separation

**Tests**: All display tests must pass, add tooling package tests

---

### Phase 4: Model Provider Organization (3-4 hours, LOW RISK)

**What**: Organize model code by provider (Gemini, OpenAI, VertexAI)

**Before**:
```
pkg/models/
├── gemini.go
├── openai.go
├── openai_adapter.go (288 lines)
├── openai_adapter_helpers.go (426 lines)
└── vertexai.go
```

**After**:
```
pkg/models/
├── types.go
├── registry.go
├── gemini/
│   └── gemini.go
├── openai/
│   ├── openai.go
│   ├── adapter.go
│   └── adapter_helpers.go
└── vertexai/
    └── vertexai.go
```

**Benefits**: Easier to add new providers, clearer boundaries, better testability

**Tests**: All model tests must pass, add per-provider tests

---

### Phase 5: Testing & Quality (6-8 hours, ZERO RISK)

**What**: Expand test coverage and documentation

**Tasks**:
1. Add integration tests (agent workflow, session roundtrip, multi-workspace)
2. Add table-driven tests for edge cases
3. Create architecture documentation
4. Add code metrics to CI (gocyclo, golangci-lint)

**Target**: 40+ test files (currently 30), critical paths >80% coverage

**Tests**: Only additive, no modifications to existing code

---

## Implementation Schedule

### Week 1: Foundation
- Days 1-2: Phase 1 (Error handling)
- Day 3: Phase 2 (Agent prompts)
- Days 4-5: Phase 5 start (Tests)

### Week 2: Display
- Days 1-3: Phase 3 (Display refactoring)
- Days 4-5: Phase 5 continue (Tests)

### Week 3: Completion
- Days 1-2: Phase 4 (Model providers)
- Day 3: Phase 5 (Documentation)
- Days 4-5: Final validation & metrics

---

## Risk Mitigation - The 0% Regression Guarantee

### Before Each Phase
✓ Create git branch: `refactor/phase-{N}-{name}`  
✓ Run `make check` - baseline  
✓ Capture test coverage metrics  
✓ Document current behavior

### During Each Phase
✓ Run `make check` after every change  
✓ Commit after each working increment  
✓ Keep changes small (<200 lines/commit)  
✓ Add tests before moving code

### After Each Phase
✓ All tests pass: `make test`  
✓ All lints pass: `make check`  
✓ Manual CLI smoke test  
✓ Performance unchanged (measure with `time`)  
✓ Merge only if 100% green

### If Anything Breaks
🛑 **STOP IMMEDIATELY**  
🔄 Revert to last good commit  
📝 Document issue  
🤔 Review approach before retry

---

## Success Criteria (Must Achieve All)

### Functional (Non-Negotiable)
- [ ] All 30 existing test files still pass
- [ ] `make check` passes with zero warnings
- [ ] CLI behavior identical (manual verification)
- [ ] Session persistence works (roundtrip test)
- [ ] Tool execution unchanged (integration test)

### Structural (Goals)
- [ ] Error handling consistent across packages
- [ ] Package structure clearer (fewer files at roots)
- [ ] Test coverage increased (40+ files)
- [ ] Documentation improved (4+ new docs)

### Measurable
- [ ] Zero increase in cyclomatic complexity
- [ ] LOC per file reduced in refactored packages  
- [ ] Build time unchanged or improved
- [ ] No new linter warnings

---

## What We're NOT Doing

❌ Changing CLI interface or flags  
❌ Modifying tool functionality  
❌ Rewriting workspace detection  
❌ Updating external dependencies  
❌ Performance optimization  
❌ Adding new features  
❌ Changing database schema

**Focus**: Organization and maintainability ONLY

---

## Quick Reference - File Movements

**Phase 2 - Agent Prompts**: 5 files → `agent/prompts/`  
**Phase 3 - Display Tooling**: 4 files → `display/tooling/`  
**Phase 4 - Model Providers**: 6 files → `pkg/models/{gemini,openai,vertexai}/`

**Total Files Moved**: 15 files  
**Backward Compatibility**: 100% via re-export facades

---

## Approval Checklist

**Before starting**:
- [ ] Team reviewed this plan
- [ ] Timeline acceptable
- [ ] Risk mitigation understood
- [ ] Rollback strategy clear
- [ ] Commit to running tests after every change

**Sign-off**:
- [ ] Tech Lead: _________________
- [ ] Start Date: _________________

---

## Emergency Contact

**If issues arise**:
1. Stop work immediately
2. Document issue in `logs/YYYY-MM-DD-rollback.md`
3. Notify team
4. Execute rollback using git revert/reset
5. Schedule review meeting

**Rollback is not failure** - it's smart risk management.

---

**Remember**: Reputation is at stake. When in doubt, don't proceed. Test everything twice.

---

**Document Version**: 1.0  
**Detailed Plan**: See `docs/refactor_plan.md` for full technical details
