# ADR-0006: Context Management - Quick Start

**Status**: ✅ Ready for Development  
**File**: `0006-context-management.md` (558 lines)  
**Timeline**: 8-10 days to production  

---

## What's New

**Clean, production-ready ADR for context management system**:

- Output truncation (head+tail strategy)
- Token tracking (real-time metrics)
- Conversation compaction (automatic at 70%)

**Previously**: 9 verbose analysis documents (~35,000 words)
**Now**: 1 focused ADR (558 lines) + this quick-start

---

## For Different Roles

### 👨‍💼 **Decision Makers** (5 minutes)

1. Read **Decision** section in main ADR
2. Review **Success Criteria** (9 measurable metrics)
3. Check **Timeline** (8-10 days)

→ **Decision**: Proceed with 3-layer approach? Yes ✅

### 👨‍💻 **Developers** (Implementation)

1. Read main ADR once (15 min)
2. Follow **Implementation** section for each component
3. Implement in **Phase order** (1-6 in checklist)
4. Use **Success Criteria** to verify each phase
5. Pre-commit: Run `make check` + all tests

### 🧪 **QA / Testing**

1. Review **Testing Strategy** (50+ specific test cases)
2. Implement **Success Criteria** checks (9 metrics)
3. Run **Integration Tests** (50-turn workflow)
4. Verify all tests from **Checklist** pass

---

## The Five Components to Build

| # | Component | File | Size | Days | Purpose |
|---|-----------|------|------|------|---------|
| 1 | ContextManager | `internal/context/manager.go` | 300 L | 1.0 | Core context tracking |
| 2 | Truncation | `internal/context/truncate.go` | 80 L | 0.5 | Head+tail output limiting |
| 3 | Token Tracking | `internal/context/token_tracker.go` | 120 L | 0.5 | Per-turn metrics |
| 4 | Compaction | `internal/context/compaction.go` | 150 L | 1.0 | Auto conversation summarization |
| 5 | Instructions | `internal/instructions/loader.go` | 150 L | 0.5 | Hierarchical AGENTS.md loading |

**Plus**: Tests (~600 L, 2 days) + Integration (~100 L, 1.5 days) + Docs (0.5 days)

---

## The 9 Success Metrics

After you're done, verify these work:

1. ✅ **Truncation** – Outputs >10 KiB → reduced, start+end preserved
2. ✅ **Token Accuracy** – Estimated tokens within ±10% of actual
3. ✅ **Compaction Trigger** – ErrCompactionNeeded at exactly 70%
4. ✅ **Compaction Ratio** – 50K tokens → <5K tokens (~10x)
5. ✅ **Instructions Load** – All 3 levels (global/project/local) merged, <32 KiB
6. ✅ **History Valid** – No orphaned tool outputs after compaction
7. ✅ **REPL Metrics** – Token info displayed accurately
8. ✅ **No Silent Loss** – All truncations logged in audit trail
9. ✅ **Long Workflows** – 50+ turn test passes

---

## The 4 Integration Points

Where this system plugs into existing code:

```
1. Session Creation
   └─ Initialize ContextManager with model config

2. Agent Loop  
   └─ Call ContextManager.AddItem() after each tool
   └─ Handle ErrCompactionNeeded trigger

3. REPL Display
   └─ Show token usage metrics after each turn

4. Model Registry
   └─ Add ContextWindow field to ModelInfo
```

---

## Implementation Phases (Week-by-Week)

### **Week 1: Core + Integration**

**Days 1-2**: Build foundation

```plaintext
ContextManager + Truncation + Tests
Create: internal/context/manager.go (300 L)
Create: internal/context/truncate.go (80 L)
Tests: 200+ assertions covering both
Result: Truncation working end-to-end
```

**Days 2.5-3**: Tracking + Compaction

```plaintext
Token Tracking + Compaction
Create: internal/context/token_tracker.go (120 L)
Create: internal/context/compaction.go (150 L)
Tests: Compaction ratio, trigger threshold
Result: Both layers working with agent loop
```

**Days 3.5-4**: Integration

```plaintext
Wire into Session + Agent + Display
Modify: internal/session/manager.go
Modify: pkg/agents/agent.go
Create: internal/display/metrics.go
Result: Full system integrated, demo-ready
```

### **Week 2: Polish + Documentation**

**Day 1**: Instructions + Final

```plaintext
Create: internal/instructions/loader.go (150 L)
Tests: 3-level hierarchy with size limits
Result: Hierarchical instructions working
```

**Days 2-3**: Testing + Docs

```plaintext
Full integration test (50 turns)
Documentation updates
Pre-commit validation
Result: Production-ready, all tests pass
```

---

## Start Here Checklist

Before you begin implementation:

- [ ] **Read full ADR** – `docs/adr/0006-context-management.md` (15 min)
- [ ] **Understand the 3 layers** – Truncation → Tracking → Compaction
- [ ] **Know the 5 components** – What each file does
- [ ] **Review test strategy** – Know what you need to test
- [ ] **Check integration points** – Where in codebase you'll make changes
- [ ] **Confirm timeline** – 8-10 days matches your plan
- [ ] **Get questions answered** – Reference ADR for everything

---

## How to Get Unstuck

**Question**: "How do I implement [component]?"  
**Answer**: See **Implementation** section in main ADR for that component's API

**Question**: "What should I test?"  
**Answer**: See **Testing Strategy** for specific test names and assertions

**Question**: "How do I integrate with [system]?"  
**Answer**: See **Integration Points** for the exact location and code pattern

**Question**: "How do I know I'm done?"  
**Answer**: See **Success Criteria** - verify all 9 metrics pass

**Question**: "What's the timeline?"  
**Answer**: See **Timeline & Milestones** - it's broken into daily tasks

---

## Key Files

| File | Purpose | Audience |
|------|---------|----------|
| `0006-context-management.md` | Main ADR (read this) | Everyone |
| `README_0006.md` | This file (quick orientation) | Getting started |

---

## One More Thing

The ADR is **self-contained**. Everything you need is there:

- ✅ What to build (architecture)
- ✅ How to build it (implementation details)
- ✅ How to test it (specific test cases)
- ✅ How to know you're done (success criteria)

No need to look elsewhere. Reference the ADR for everything.

---

## Ready? 

1. **Read ADR-0006** (15 minutes)
2. **Create feature branch** (`feat/context-management`)
3. **Follow Week 1-2 phases**
4. **Verify all 9 metrics pass**
5. **Commit with confidence**

---

**🚀 Good to ship!**
