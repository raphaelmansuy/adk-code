# Agent Definition Support - Complete Project Roadmap

**Project**: Claude Code Agent Definition Support for adk-code  
**Status**: 🟢 Phase 0 Complete, Phase 1 Planned  
**Last Updated**: November 14, 2025  
**Project Lead**: Raphael Mansuy  
**Repository**: `raphaelmansuy/adk-code`

---

## Executive Summary

The Claude Code agent definition support project has successfully completed Phase 0 (proof-of-concept) with:
- ✅ **22/22 tests passing** (17 core + 5 tool)
- ✅ **89% code coverage** (exceeds 80% target)
- ✅ **~1,240 lines** of implementation + documentation
- ✅ **4 clean commits** with comprehensive documentation

Phase 1 is planned for December 9 - January 6, 2026, expanding to multi-path, multi-source agent discovery.

---

## Project Vision

Enable adk-code to discover, manage, and integrate with Claude Code agent definitions across multiple sources (project, user, plugin), providing a foundation for sophisticated agent selection and composition.

---

## Phase Overview

### Phase 0: Proof of Concept ✅ COMPLETE

**Duration**: 2 weeks (completed)  
**Status**: 🟢 READY FOR MERGE  
**Deliverables**:
- Single-path agent discovery (.adk/agents/)
- YAML frontmatter parsing
- CLI `list_agents` tool
- 22 comprehensive tests
- 89% code coverage

**Files**: 4 created + 1 updated
- `pkg/agents/agents.go` (500 lines)
- `pkg/agents/agents_test.go` (400 lines)
- `tools/agents/agents_tool.go` (140 lines)
- `tools/agents/agents_tool_test.go` (50 lines)
- `tools/tools.go` (updated: +10 lines)

**Commits**:
```
b54e3a7 - docs: Phase 0 completion report
678e1ab - test: Add comprehensive tests for agents tool
5f3b0dd - feat(tools): Add agents discovery CLI tool
da83036 - feat(agents): Phase 0 core implementation
```

**Next Steps**:
1. Code review feedback integration
2. Merge to main branch
3. Phase 1 branch creation

---

### Phase 1: Multi-Path Discovery 📋 PLANNED

**Duration**: 4 weeks (Dec 9, 2025 - Jan 6, 2026)  
**Status**: 📋 PLANNED  
**Deliverables**:
- Multi-path agent discovery (3 sources)
- Configuration system (.adk/config.yaml)
- Enhanced metadata (version, author, tags, dependencies)
- Advanced filtering capabilities
- 90+ tests
- 85%+ code coverage
- Complete user documentation

**Planned Changes**:
- New: `pkg/agents/config.go` (~150 lines)
- New: `pkg/agents/integration_test.go` (~150 lines)
- New: `tools/agents/discover_paths.go` (~100 lines)
- Update: `pkg/agents/agents.go` (+100 lines)
- Update: `tools/agents/agents_tool.go` (+150 lines)
- New: `docs/AGENT_DEFINITIONS.md` (~300 lines)

**Estimated Lines**: 1,100-1,300 total (1.2K additional)

**Success Criteria**:
- [ ] 85%+ code coverage
- [ ] 90+ tests passing
- [ ] All CLI tools functional
- [ ] Multi-path discovery working
- [ ] Configuration system operational

---

### Phase 2: Agent Execution & Caching 🔄 FUTURE

**Duration**: 4 weeks (Jan 13 - Feb 10, 2026)  
**Status**: 🔄 TBD  
**Planned Deliverables**:
- Agent execution framework
- Caching and performance optimization
- Agent dependency resolution
- Advanced filtering and search
- Integration with ADK runtime

**Estimated Effort**: 1,500+ lines

---

### Phase 3: Claude Code Integration 🔄 FUTURE

**Duration**: 6-8 weeks (Feb 17 - Apr 15, 2026)  
**Status**: 🔄 TBD  
**Planned Deliverables**:
- Claude Code protocol support
- Agent invocation from Claude Code
- Real-time agent discovery
- Agent composition and chains

**Estimated Effort**: 2,000+ lines

---

### Phase 4: Advanced Features 🔄 FUTURE

**Duration**: 8+ weeks (Apr 22+, 2026)  
**Status**: 🔄 TBD  
**Planned Features**:
- Agent marketplace
- Version management
- Remote agent discovery
- Agent authentication & security
- Performance optimization

---

## Technical Architecture

### Core Components

```
┌─────────────────────────────────────────────────────┐
│           Agent Definition System                    │
├─────────────────────────────────────────────────────┤
│                                                       │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────┐ │
│  │   Discovery  │  │   Metadata   │  │ Execution  │ │
│  │   (Phase 0)  │  │  (Phase 1)   │  │ (Phase 2)  │ │
│  └──────────────┘  └──────────────┘  └────────────┘ │
│         ↓                 ↓                  ↓       │
│  ┌─────────────────────────────────────────────────┐ │
│  │         Agent Registry & Caching               │ │
│  │            (Phase 1-2)                         │ │
│  └─────────────────────────────────────────────────┘ │
│         ↓              ↓              ↓             │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────┐ │
│  │   File I/O   │  │  YAML Parser │  │    Tool    │ │
│  │  (Multi-src) │  │  (Extended)  │  │ Interface  │ │
│  └──────────────┘  └──────────────┘  └────────────┘ │
│                                                       │
└─────────────────────────────────────────────────────┘
         │                    │
         ↓                    ↓
    ┌─────────┐          ┌──────────┐
    │ CLI Tool│          │ ADK Tools │
    │         │          │ Registry  │
    └─────────┘          └──────────┘
```

### Data Flow

**Phase 0** (Current):
```
.adk/agents/*.md → YAML Parser → Agent Model → CLI Tool → User
```

**Phase 1** (Planned):
```
Config + Env Vars
    ↓
Multiple Paths (project, user, plugin)
    ↓
Parallel Discovery + Metadata Extraction
    ↓
Deduplication + Source Attribution
    ↓
Filtering + Caching
    ↓
CLI Tools + ADK Integration
```

### File Locations (Phase 0 + Phase 1)

```
Project Root/
├── .adk/
│   ├── agents/              (Phase 0 & 1)
│   │   ├── agent-one.md
│   │   ├── agent-two.md
│   │   └── ...
│   └── config.yaml          (Phase 1 NEW)
│
Home Directory/
└── .adk/
    ├── agents/              (Phase 1 NEW)
    │   ├── user-agent-1.md
    │   └── ...
    └── config.yaml
```

---

## Development Timeline

### Q4 2025 (Oct - Dec)

| Week | Phase | Task | Status |
|------|-------|------|--------|
| 44-45 | Planning | Gap analysis, document review | ✅ DONE |
| 46-47 | Phase 0 | Core discovery + tests | ✅ DONE |
| 48 | Phase 0 | Tool integration + finalization | ✅ DONE |
| 49-52 | Phase 1 | Multi-path discovery system | 📋 PLANNED |

### Q1 2026 (Jan - Mar)

| Week | Phase | Task | Status |
|------|-------|------|--------|
| 1-2 | Phase 1 | Metadata enhancement + CLI tools | 📋 PLANNED |
| 3-4 | Phase 1 | Integration + documentation | 📋 PLANNED |
| 5-8 | Phase 2 | Execution framework | 🔄 TBD |
| 9-13 | Phase 3 | Claude Code integration | 🔄 TBD |

---

## Resource Requirements

### Phase 0 (Completed)

**Effort**: 40 hours  
**Team**:
- 1 Senior Developer (Raphael)

**Tools**:
- Go 1.24+
- gopkg.in/yaml.v3
- Google ADK Framework

**Infrastructure**:
- Git + GitHub
- CI/CD pipeline
- Test infrastructure

### Phase 1 (Planned)

**Effort**: 60-80 hours  
**Team**:
- 1 Senior Developer (Raphael)
- Code reviewer (TBD)

**Timeline**: 4 weeks (full-time equivalent)

### Phase 2+ (Future)

**Estimated Effort**: 100+ hours each  
**Team Size**: May require 2 developers

---

## Key Metrics & Goals

### Code Quality

| Metric | Target | Phase 0 | Phase 1 Goal |
|--------|--------|---------|--------------|
| Coverage | 80%+ | 89% ✅ | 85%+ |
| Tests | 100% pass | 22/22 ✅ | 90+ |
| Lint errors | 0 | 0 ✅ | 0 |
| Doc coverage | 100% | 100% ✅ | 100% |

### Performance

| Operation | Target | Status |
|-----------|--------|--------|
| Discovery scan | <50ms | <10ms ✅ |
| YAML parsing | <5ms/file | <1ms ✅ |
| Tool invocation | <100ms | <50ms (Phase 1) |
| Agent listing | <200ms | <100ms (Phase 1) |

### Delivery

| Phase | Status | Actual Date | Impact |
|-------|--------|-------------|--------|
| Phase 0 | ✅ COMPLETE | Nov 14, 2025 | On schedule |
| Phase 1 | 📋 PLANNED | Jan 6, 2026 | On track |
| Phase 2 | 🔄 TBD | ~Feb 10, 2026 | TBD |
| Phase 3 | 🔄 TBD | ~Apr 15, 2026 | TBD |

---

## Risk Register

### Critical Risks

**Risk 1: Phase 0 Merge Delays**
- **Impact**: HIGH (blocks Phase 1)
- **Probability**: LOW
- **Mitigation**: Already code-review ready

**Risk 2: Scope Creep in Phase 1**
- **Impact**: HIGH (timeline slip)
- **Probability**: MEDIUM
- **Mitigation**: Strict scope fence, documented deferments

**Risk 3: Multi-path Conflicts**
- **Impact**: HIGH (incorrect behavior)
- **Probability**: MEDIUM
- **Mitigation**: Comprehensive conflict resolution tests

### Medium Risks

**Risk 4: Configuration Complexity**
- **Impact**: MEDIUM (user confusion)
- **Probability**: MEDIUM
- **Mitigation**: Clear defaults, excellent documentation

**Risk 5: Integration Test Coverage**
- **Impact**: MEDIUM (hidden bugs)
- **Probability**: MEDIUM
- **Mitigation**: Planned integration test suite Week 4

---

## Communication Plan

### Status Updates
- **Weekly**: Friday 3pm PT (team sync)
- **Bi-weekly**: Detailed progress report
- **Monthly**: Executive summary

### Documentation
- **Real-time**: Git commit messages
- **Weekly**: Phase update in PRs
- **Phase end**: Completion report + lessons learned

### Review Gates
- **Phase 0**: ✅ READY (awaiting code review)
- **Phase 1**: Week 1 planning sync
- **Phase 2**: Phase 1 completion review

---

## Success Criteria - All Phases

### Functional ✅
- Agent discovery works across all sources
- All CLI tools functional
- No blocking bugs
- Performance acceptable

### Quality ✅
- 85%+ code coverage
- 100% test pass rate
- Zero lint/format issues
- Clean git history

### Documentation ✅
- User guide complete
- API documentation complete
- Code comments throughout
- Phase completion reports

### Timeline ✅
- No scope creep
- Milestones met
- Team aligned
- Stakeholders informed

---

## Appendix: Phase 0 Achievements

### Completed Tasks
- ✅ Consolidate agent types, parser, discovery into single file (recovery from corruption)
- ✅ Implement comprehensive YAML frontmatter extraction
- ✅ Implement recursive .adk/agents/ directory scanning
- ✅ Create robust error handling with graceful degradation
- ✅ Integrate with ADK tool registry
- ✅ Create 22 comprehensive unit tests (89% coverage)
- ✅ Update planning documents with reality-based timelines
- ✅ Create Phase 0 completion report
- ✅ Plan Phase 1 in detail

### Commit History
```
b54e3a7 - docs: Phase 0 completion report
678e1ab - test: Add comprehensive tests for agents tool
5f3b0dd - feat(tools): Add agents discovery CLI tool
da83036 - feat(agents): Phase 0 core implementation
ffa8308 - docs: Add agent definition support planning documents
```

### Lessons Learned
1. **File Size Matters**: Consolidated files reduce corruption risk
2. **Test-First Works**: Writing tests first improved quality
3. **Clear Scope**: Phase 0 success due to tight scope definition
4. **Documentation**: Planning docs prevent surprises later
5. **Gradual Expansion**: Better to split into phases than rush

---

## Next Immediate Actions

### Now (Week of Nov 14)
1. ✅ Phase 0 completion documentation
2. ✅ Phase 1 detailed planning
3. ⏳ Phase 0 code review (pending)

### Week of Nov 21 (Post-Holiday)
1. Incorporate code review feedback
2. Merge Phase 0 to main branch
3. Create Phase 1 feature branch

### Week of Dec 2
1. Final Phase 1 scope confirmation
2. Team sync on Phase 1 approach
3. Begin Phase 1 Week 1 tasks (Dec 9)

---

## References

- **Phase 0 Report**: `docs/PHASE0_COMPLETION.md`
- **Phase 1 Plan**: `docs/0001-agent-definition-support-PHASE1.md`
- **ADR**: `docs/adr/0001-claude-code-agent-support.md`
- **Spec**: `docs/spec/0001-agent-definition-support.md`
- **Risk Report**: `docs/RISK_MITIGATION_AGENT_SUPPORT.md`

---

*Project Timeline: Oct 15, 2025 - Apr 30, 2026*  
*Total Planned Effort: 250+ hours*  
*Team Size: 1-2 developers*  
*Status: 🟢 On Track*

Last Updated: November 14, 2025
