# CLI Model Selection & Provider Architecture - Complete Brainstorm

**Date:** November 11, 2025  
**Status:** Complete brainstorm - ready for implementation  
**Timeline:** 2-3 weeks to production

---

## Overview

This directory contains a comprehensive brainstorm exploring the redesign of CLI model/provider selection for the Code Agent. The proposal introduces a clean `provider/model` syntax that treats providers as first-class concepts.

### The Problem
Current CLI requires separate flags to specify provider and model:
```bash
code-agent --backend vertexai --project my-proj --location us-central1 --model gemini-1.5-pro
```

### The Solution
Single, intuitive syntax:
```bash
code-agent --model vertexai/1.5-pro --project my-proj --location us-central1
```

---

## Documents in This Directory

### 📖 Start Here

**1. QUICK_REFERENCE.md** (⚡ 3 pages)
- Quick overview and decision reference
- One-sentence problem/solution
- Core concepts and syntax
- All syntax examples for quick lookup
- **Best for:** Quick answers, stakeholder demos

**2. CLEAN_REDESIGN_RATIONALE.md** (📋 New!)
- Why we don't need backward compatibility
- Benefits of clean approach (simpler code, faster)
- Comparison: old vs new timeline
- Implementation approach (3 weeks)
- **Best for:** Understanding why this is the right approach

---

### 🏗️ Deep Dive

**3. MODEL_SELECTION_REDESIGN.md** (🏗️ 15 pages)
- Complete architectural design
- Current limitations analysis
- Provider-first architecture concepts
- Unified registry design
- Enhanced CLI syntax options
- Clear resolution logic with precedence
- Open design questions
- **Best for:** Architecture review, technical planning

**4. IMPLEMENTATION_ROADMAP.md** (🛣️ 16 pages)
- Step-by-step implementation guide
- Specific code changes for each phase
- Code examples showing new approach
- Test plan with concrete test cases
- Timeline: 3 phases over 2-3 weeks
- **Best for:** Developers implementing the feature

---

### 📊 Design & Examples

**5. PROVIDER_MODEL_VISUAL_EXAMPLES.md** (🎨 12 pages)
- Before/after CLI syntax comparison
- Mockups of `/providers` command output
- Model resolution flow diagrams
- Real-world usage examples
- Error messages and suggestions
- Shorthand alias reference table
- Design rationale for each decision
- **Best for:** UX/product review, understanding user experience

**6. BRAINSTORM_SUMMARY.md** (📋 9 pages)
- Executive summary with context
- Three-document overview
- Key design decisions with rationale
- Impact analysis (user benefits, code benefits)
- Future extensibility scenarios
- Success metrics to measure
- **Best for:** Stakeholder presentations, strategic alignment

---

### 🗺️ Navigation

**7. BRAINSTORM_INDEX.md** (📑 Complete guide)
- Master index of all documents
- Document statistics and word counts
- How to navigate by role
- Document map and relationships
- **Best for:** Finding what you need

---

## Reading Paths by Role

### 👨‍💼 For Product Managers/Stakeholders
1. Read: `QUICK_REFERENCE.md` (5 min)
2. Read: `BRAINSTORM_SUMMARY.md` (15 min)
3. Read: `CLEAN_REDESIGN_RATIONALE.md` (10 min)
4. Decision: Approve for implementation?

### 🏗️ For Architects
1. Read: `MODEL_SELECTION_REDESIGN.md` (30 min)
2. Review: `PROVIDER_MODEL_VISUAL_EXAMPLES.md` (15 min)
3. Check: Key design decisions section
4. Feedback: Sound architecture? Extensible?

### 👨‍💻 For Developers
1. Read: `IMPLEMENTATION_ROADMAP.md` (30 min)
2. Review: Phase 1-3 specific code changes
3. Study: Code examples and parsing logic
4. Plan: Implementation sprint breakdown

### 🧪 For QA/Testers
1. Skim: `PROVIDER_MODEL_VISUAL_EXAMPLES.md` (10 min)
2. Study: `IMPLEMENTATION_ROADMAP.md` test section (10 min)
3. Create: Test cases for each phase
4. Execute: Testing strategy

---

## Key Insights

### Clean Design (No Legacy Support)
- ✅ Only `--model provider/model` syntax (no `--backend` flag)
- ✅ Single model definitions (no `-vertex` suffix duplication)
- ✅ Simpler parsing (one code path, not two)
- ✅ Faster implementation (2-3 weeks vs 4+ weeks)
- ✅ Zero technical debt from launch

### Provider/Model Syntax
```bash
# Explicit provider specification
code-agent --model gemini/2.5-flash
code-agent --model vertexai/1.5-pro

# Shorthand (latest of type)
code-agent --model gemini/flash
code-agent --model vertexai/pro

# Future providers just work
code-agent --model openai/gpt-4
code-agent --model claude/3.5-sonnet
```

### Timeline
- **Week 1:** Core implementation (parsing + flag simplification)
- **Week 2:** Registry refactoring (aliases + clean definitions)
- **Week 3:** Polish & release (tests + documentation)
- **Total:** 2-3 weeks to production

---

## Implementation Checklist

- [ ] Review all documents
- [ ] Approve design approach
- [ ] Resolve open questions (if any)
- [ ] Create GitHub issues for each phase
- [ ] Assign implementation owner
- [ ] Begin Phase 1 (Week 1)
- [ ] Begin Phase 2 (Week 2)
- [ ] Begin Phase 3 (Week 3)
- [ ] Release v1.1.0

---

## Files Modified

All brainstorm documents have been updated to reflect the **clean redesign approach** (no backward compatibility needed):

✅ `MODEL_SELECTION_REDESIGN.md` - Clean architecture section  
✅ `IMPLEMENTATION_ROADMAP.md` - New 2-3 week timeline  
✅ `QUICK_REFERENCE.md` - Simplified phases  
✅ `BRAINSTORM_SUMMARY.md` - Updated timeline  
✅ `CLEAN_REDESIGN_RATIONALE.md` - New document explaining why  

---

## Quick Facts

| Metric | Value |
|--------|-------|
| **Total Pages** | ~55 |
| **Total Words** | ~22,000 |
| **Implementation Time** | 2-3 weeks |
| **Breaking Changes** | 0 (fresh version) |
| **Technical Debt** | 0 (clean from start) |
| **Code Branches** | 1 (single syntax) |

---

## Next Steps

1. **Review:** Have team members read appropriate documents
2. **Discuss:** Schedule architecture/design review meeting
3. **Approve:** Confirm implementation approach
4. **Plan:** Create GitHub issues for 3 phases
5. **Start:** Begin Week 1 implementation

---

## Questions?

Refer to the document that addresses your question:

- **How does the new syntax work?** → `QUICK_REFERENCE.md` or `PROVIDER_MODEL_VISUAL_EXAMPLES.md`
- **Why this design?** → `MODEL_SELECTION_REDESIGN.md` or `CLEAN_REDESIGN_RATIONALE.md`
- **How do I implement it?** → `IMPLEMENTATION_ROADMAP.md`
- **What's the overall strategy?** → `BRAINSTORM_SUMMARY.md`
- **Where do I start?** → This README or `BRAINSTORM_INDEX.md`

---

**Status:** ✅ Complete brainstorm  
**Ready for:** Implementation planning  
**Approved for:** Team review  
