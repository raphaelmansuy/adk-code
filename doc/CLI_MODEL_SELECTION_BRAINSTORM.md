# CLI Model & Provider Selection Brainstorm

**Complete brainstorm for provider/model selection redesign**

All documents have been organized in the `cli-models/` subdirectory. See `cli-models/README.md` for the complete guide.

## Quick Links

- **Start Here:** [`cli-models/README.md`](./cli-models/README.md)
- **Quick Reference:** [`cli-models/QUICK_REFERENCE.md`](./cli-models/QUICK_REFERENCE.md) - 3 pages, all key info
- **Clean Design Rationale:** [`cli-models/CLEAN_REDESIGN_RATIONALE.md`](./cli-models/CLEAN_REDESIGN_RATIONALE.md) - Why no backward compat
- **Architecture:** [`cli-models/MODEL_SELECTION_REDESIGN.md`](./cli-models/MODEL_SELECTION_REDESIGN.md) - Full design (15 pages)
- **Implementation:** [`cli-models/IMPLEMENTATION_ROADMAP.md`](./cli-models/IMPLEMENTATION_ROADMAP.md) - Code changes (16 pages)
- **UX Examples:** [`cli-models/PROVIDER_MODEL_VISUAL_EXAMPLES.md`](./cli-models/PROVIDER_MODEL_VISUAL_EXAMPLES.md) - Mockups (12 pages)
- **Summary:** [`cli-models/BRAINSTORM_SUMMARY.md`](./cli-models/BRAINSTORM_SUMMARY.md) - Executive summary (9 pages)
- **Index:** [`cli-models/BRAINSTORM_INDEX.md`](./cli-models/BRAINSTORM_INDEX.md) - Master guide

## Summary

### The Proposal
Replace `--backend X --model Y` with intuitive `--model provider/model` syntax:

```bash
# Current (verbose)
code-agent --backend vertexai --project my-proj --location us-central1 --model gemini-1.5-pro

# Proposed (clean)
code-agent --model vertexai/1.5-pro --project my-proj --location us-central1
```

### Key Benefits
- ✅ Single, clear syntax (provider/model separated by `/`)
- ✅ No model duplication in registry
- ✅ Extensible to future providers (OpenAI, Claude, etc.)
- ✅ Simpler codebase (no legacy branches)
- ✅ Faster implementation (2-3 weeks)

### Timeline
- **Week 1:** Core implementation (parsing + flag cleanup)
- **Week 2:** Registry refactoring (aliases + clean definitions)
- **Week 3:** Polish & release (tests + documentation)

---

**Status:** ✅ Complete brainstorm - ready for team review and implementation planning
