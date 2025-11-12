# Code Agent Refactoring Plan - Executive Summary

**Date**: November 12, 2025  
**Status**: ✅ Analysis Complete  
**Full Audit**: See `docs/audit.md` (1,162 lines, 32KB)

---

## Quick Assessment

**Current Architecture Grade: 7.5/10** (Good → Target: 9/10 Excellent)

### Strengths ✅
- Clean layered architecture
- Zero circular dependencies
- 100% test pass rate
- Strong design patterns
- Well-documented refactoring history

### Issues Identified ⚠️
1. Package location inconsistencies (workspace/, tracking/, agent_prompts/ at root)
2. Implicit tool registration (init() functions)
3. Mixed responsibilities in internal/app
4. Display package may be over-engineered (23+ subpackages)

---

## Pragmatic Recommendations

### Priority 1: High-Impact, Low-Risk ⭐

| ID | Recommendation | Effort | Risk | Impact |
|----|----------------|--------|------|--------|
| R1.1 | Reorganize root packages to pkg/internal | 2-4h | LOW | HIGH |
| R1.2 | Replace init() with explicit tool registration | 3-6h | LOW | HIGH |
| R1.3 | Consolidate internal/app responsibilities | 4-6h | MEDIUM | HIGH |

**Total P1 Effort: 9-16 hours**

### Priority 2: Medium-Impact, Low-Risk

| ID | Recommendation | Effort | Risk | Impact |
|----|----------------|--------|------|--------|
| R2.1 | Add package-level documentation (doc.go) | 1-2h | ZERO | MEDIUM |
| R2.2 | Extract common factory interface | 2-3h | LOW | MEDIUM |
| R2.3 | Simplify display package structure | 2-4h | LOW | MEDIUM |

**Total P2 Effort: 5-9 hours**

### Priority 3: Future Enhancements (Optional)

- R3.1: Generic component lifecycle (4-8h, MEDIUM-HIGH risk)
- R3.2: Plugin architecture (8-12h, HIGH risk)

**Note:** P3 items are optional and should only be pursued if needed.

---

## Implementation Roadmap

### Sprint 1: Foundation (1 week)
- Day 1-2: Package documentation (R2.1)
- Day 3-4: Reorganize packages (R1.1)
- Day 5: Explicit tool registration (R1.2)

### Sprint 2: Consolidation (1 week)
- Day 1-3: Consolidate internal/app (R1.3)
- Day 4-5: Simplify display package (R2.3)

### Sprint 3: Polish (3-5 days)
- Day 1-2: Common factory interface (R2.2)
- Day 3-4: Testing, validation, documentation

**Total Effort: 13-15 days (P1 + P2 only)**

---

## Key Changes Summary

### R1.1: Package Reorganization (2-4h)

**Before:**
```
code_agent/
├── workspace/        # Root level (wrong)
├── tracking/         # Root level (wrong)
└── agent_prompts/    # Root level (wrong)
```

**After:**
```
code_agent/
├── pkg/
│   └── workspace/    # Reusable workspace logic
└── internal/
    ├── tracking/     # App-specific tracking
    └── prompts/      # App-specific prompts
```

### R1.2: Explicit Tool Registration (3-6h)

**Before:**
```go
// Magic init() in each tool subpackage
func init() {
    registry.Register(ReadFileTool())
}
```

**After:**
```go
// Clear, explicit registration in one place
func RegisterAllTools(reg *base.ToolRegistry) error {
    reg.Register(file.ReadFileTool())
    reg.Register(file.WriteFileTool())
    // ... all tools clearly listed
}
```

### R1.3: Consolidate internal/app (4-6h)

**Before:**
```
internal/app/
├── app.go           # Main
├── components.go    # Type aliases (band-aid)
├── factories.go     # Creates OTHER packages' components
├── session.go
├── signals.go
├── utils.go
```

**After:**
```
internal/app/
├── app.go           # Focused application lifecycle
└── lifecycle.go     # Init/cleanup

internal/orchestration/
└── factories/       # Component creation logic
    ├── display.go
    ├── model.go
    ├── agent.go
    └── session.go
```

---

## Success Criteria

All changes must meet these criteria:

- ✅ All tests pass (100% pass rate)
- ✅ No build warnings
- ✅ No new circular dependencies
- ✅ Backward compatibility maintained
- ✅ Documentation updated
- ✅ Code review approved

---

## Risk Mitigation

### Development Process
1. One recommendation per feature branch
2. Comprehensive testing after each change
3. Peer review before merge
4. Git tags before major changes

### Validation
```bash
# After each change
make check           # fmt, vet, test, build
go mod graph         # Check for circular deps
```

### Rollback Plan
- Keep git history clean
- Tag before changes: `git tag pre-refactor-r1.1`
- Document rollback procedures

---

## Expected Outcomes

### Architecture Improvements
- ✅ Clear package organization (pkg/ vs internal/)
- ✅ Explicit dependencies (no magic init())
- ✅ Better separation of concerns
- ✅ Comprehensive documentation
- ✅ Simpler navigation

### Quality Metrics
- Maintain: 100% test pass rate
- Maintain: 0 circular dependencies
- Improve: Documentation coverage (40% → 90%)
- Improve: Code organization score (7.5/10 → 9/10)

---

## Commitment

**This plan prioritizes:**
1. ✅ **Zero Regression** - All existing functionality preserved
2. ✅ **Pragmatic Approach** - High-impact, low-risk changes first
3. ✅ **Incremental Progress** - One change at a time
4. ✅ **Comprehensive Testing** - Validate after every step
5. ✅ **Quality Assurance** - Your reputation protected

**The codebase is already good. These changes will make it excellent.**

---

## Next Steps

1. **Review this plan** - Validate recommendations align with goals
2. **Prioritize items** - Confirm P1 items are the right focus
3. **Start Sprint 1** - Begin with low-risk documentation and reorganization
4. **Track progress** - Update this document as work completes
5. **Log outcomes** - Create session logs in logs/ directory

---

## Reference Documents

- **Full Audit**: `docs/audit.md` (comprehensive 1,162-line analysis)
- **Analysis Notes**: `docs/draft.md` (detailed working notes)
- **Architecture Docs**: `docs/architecture/` (existing documentation)
- **Refactoring History**: `logs/2025-11-*.md` (past refactoring sessions)

---

**For questions or clarifications, refer to the full audit document.**

**End of Executive Summary**
