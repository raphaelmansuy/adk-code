# Phase 1 Implementation Plan - Multi-Path Agent Discovery

**Status**: 📋 Planning  
**Planned Start**: December 9, 2025 (Week 49)  
**Planned Duration**: 4 weeks  
**Target Completion**: January 6, 2026  
**Feature Branch**: `feat/agent-definition-support-phase1`  
**Base**: `feat/agent-definition-support-phase0`

## Overview

Phase 1 extends Phase 0's proof-of-concept into a production-ready agent discovery system supporting multiple agent sources and locations. This phase adds user-level agents, plugin agent support, and configuration-driven discovery.

## Scope Definition

### In Scope ✅

1. **Multi-Path Discovery**
   - User home directory agents (`~/.adk/agents/`)
   - Plugin directory agents (`$ADK_PLUGIN_PATH/agents/`)
   - Workspace directory agents (current `.adk/agents/`)
   - Environment variable override support

2. **Agent Source Expansion**
   - `SourceProject` (Phase 0) - already working
   - `SourceUser` - NEW - user-level agents
   - `SourcePlugin` - NEW - installed plugin agents
   - Source metadata and versioning

3. **Configuration System**
   - `.adk/config.yaml` for agent path configuration
   - Environment variable overrides (`ADK_AGENT_PATHS`, `ADK_PLUGINS`)
   - Discovery strategy configuration
   - Agent filtering rules

4. **Enhanced Filtering**
   - Filter by source (project, user, plugin)
   - Filter by type (subagent, skill, command, plugin)
   - Filter by discovery scope
   - Combining multiple filters

5. **Agent Metadata Expansion**
   - Version field (semantic versioning)
   - Author field
   - Tags for categorization
   - Dependencies on other agents
   - Execution requirements

6. **Testing & Documentation**
   - Integration tests with multi-path discovery
   - Configuration file testing
   - Documentation updates
   - User guide for agent definitions

### Out of Scope ❌

- Agent execution/invocation (Phase 2)
- Claude Code integration (Phase 3)
- Agent versioning/upgrading (Phase 2+)
- Remote agent discovery (Phase 3+)
- Agent marketplace (Phase 4+)

## Detailed Tasks

### Week 1: Configuration & Path Resolution (Dec 9-13)

**Goal**: 350-400 lines  
**Priority**: CRITICAL

#### Task 1.1: Configuration System
- [ ] Create `pkg/agents/config.go` (~150 lines)
  - `Config` struct with path settings
  - `LoadConfig()` from `.adk/config.yaml`
  - Environment variable merging
  - Path validation and expansion
  - Unit tests (10+ tests)

**Input/Output**:
```go
type Config struct {
    ProjectPath  string   // .adk/agents/
    UserPath     string   // ~/.adk/agents/
    PluginPaths  []string // $ADK_PLUGIN_PATH
    SearchOrder  []string // discovery order
}

func LoadConfig(projectRoot string) (*Config, error)
func (c *Config) GetAllPaths() []string
```

#### Task 1.2: Enhanced Discoverer
- [ ] Extend `pkg/agents/agents.go` (~100 lines)
  - Multi-path discovery support
  - `DiscoverAll()` updated to handle multiple paths
  - Path deduplication and conflict resolution
  - Source attribution for each agent
  - Performance optimization

**Changes**:
```go
func (d *Discoverer) DiscoverWithConfig(cfg *Config) (*DiscoveryResult, error)
func (d *Discoverer) DiscoverFromPath(path string, source AgentSource) (*DiscoveryResult, error)
```

#### Task 1.3: Tests & Documentation
- [ ] `pkg/agents/config_test.go` (100+ lines)
  - Config loading tests
  - Environment variable tests
  - Path expansion tests
  - Error handling tests

**Metrics Target**:
- Code: 350-400 lines
- Tests: 15+ test cases
- Coverage: Maintain >85%

---

### Week 2: Metadata Enhancement (Dec 16-20)

**Goal**: 250-300 lines  
**Priority**: HIGH

#### Task 2.1: Extended Agent Model
- [ ] Update `pkg/agents/agents.go` (~100 lines)
  - Add `Version` field
  - Add `Author` field
  - Add `Tags` field ([]string)
  - Add `Dependencies` field ([]string)
  - Update YAML unmarshaling

**Schema Example**:
```yaml
---
name: my-agent
description: Does something useful
version: 1.0.0
author: example@domain.com
tags: [refactoring, python, code-analysis]
dependencies: [base-agent]
---
```

#### Task 2.2: YAML Parser Updates
- [ ] Extend `pkg/agents/agents.go` (~100 lines)
  - Parse new metadata fields
  - Version validation
  - Tag parsing and validation
  - Dependency resolution (basic)
  - Error handling for missing optional fields

#### Task 2.3: Tests
- [ ] Update `pkg/agents/agents_test.go` (50+ lines)
  - New metadata field tests
  - Version validation tests
  - Tag parsing tests
  - Backward compatibility tests (Phase 0 files)

**Metrics Target**:
- Code: 200 lines
- New tests: 8+ test cases
- Coverage: Maintain >85%
- Backward compatibility: 100%

---

### Week 3: CLI Tool Enhancement (Dec 23-27)

**Goal**: 300-350 lines  
**Priority**: HIGH

#### Task 3.1: Enhanced List Tool
- [ ] Update `tools/agents/agents_tool.go` (~150 lines)
  - Add source filtering
  - Add version display
  - Add tag filtering
  - Add dependency information
  - Enhanced output formatting

**New Parameters**:
```go
type ListAgentsInput struct {
    AgentType  string   // filter by type
    Source     string   // filter by source (project, user, plugin)
    Tag        string   // filter by tag
    Author     string   // filter by author
    Detailed   bool     // include full metadata
    IncludeDeps bool    // show dependencies
}
```

#### Task 3.2: Discovery Tool
- [ ] Create `tools/agents/discover_paths.go` (~100 lines)
  - New `discover_paths` tool
  - List all agent search paths
  - Show configuration status
  - Verify path accessibility
  - Display search order

#### Task 3.3: Tests
- [ ] Update `tools/agents/agents_tool_test.go` (100+ lines)
  - Filter combination tests
  - Source filtering tests
  - Tag filtering tests
  - Path discovery tests
  - Integration scenarios

**Metrics Target**:
- Code: 250 lines
- New tests: 10+ test cases
- All tools registered and working

---

### Week 4: Integration & Documentation (Dec 30-Jan 3)

**Goal**: 200-250 lines + Documentation  
**Priority**: CRITICAL

#### Task 4.1: Integration Tests
- [ ] Create `pkg/agents/integration_test.go` (~150 lines)
  - Multi-path discovery scenarios
  - Configuration interaction tests
  - Source conflict resolution
  - Performance baseline tests
  - End-to-end workflows

#### Task 4.2: Documentation
- [ ] Create `docs/AGENT_DEFINITIONS.md` (~300 lines)
  - Agent file format specification
  - YAML frontmatter reference
  - Configuration guide
  - Examples for each source type
  - Best practices

#### Task 4.3: Examples & Verification
- [ ] Create example agent files
  - Project-level example
  - User-level example
  - Plugin-level example
  - Complete Phase 0+1 verification

#### Task 4.4: Finalization
- [ ] Code review preparation
- [ ] Coverage verification (>85%)
- [ ] Performance baseline
- [ ] Phase 1 completion report

**Deliverables**:
- 200+ lines integration code
- 300+ lines documentation
- 5+ example agents
- Full test coverage

---

## Architecture Changes

### File Structure
```
pkg/agents/
├── agents.go           # UPDATED: Multi-path discovery
├── agents_test.go      # UPDATED: New metadata tests
├── config.go           # NEW: Configuration system
├── config_test.go      # NEW: Config tests
└── integration_test.go # NEW: Integration tests

tools/agents/
├── agents_tool.go           # UPDATED: Enhanced filtering
├── agents_tool_test.go      # UPDATED: New test cases
├── discover_paths.go        # NEW: Path discovery tool
└── discover_paths_test.go   # NEW: Path tests

docs/
├── AGENT_DEFINITIONS.md     # NEW: Complete spec
└── PHASE1_COMPLETION.md     # NEW: Phase 1 report
```

### Data Flow

**Phase 0** (Simple):
```
.adk/agents/ → Discovery → Agent List
```

**Phase 1** (Enhanced):
```
Config File + Env Vars
        ↓
Path Resolution
        ↓
Multiple Paths: ~/.adk/agents/, .adk/agents/, plugins/
        ↓
Parallel Discovery (each path)
        ↓
Source Attribution
        ↓
Metadata Enhancement
        ↓
Filtering & Sorting
        ↓
Agent List (with metadata)
```

## Testing Strategy

### Unit Tests (~60-80 tests)
- Config loading and validation
- Multi-path discovery
- Metadata parsing
- Filter combinations
- Error handling

### Integration Tests (~20-30 tests)
- End-to-end discovery flows
- Multi-source scenarios
- Configuration overrides
- Path conflict resolution
- Performance benchmarks

### Coverage Goals
- Minimum: 85% coverage
- Target: 90%+ coverage
- Critical paths: 100%

## Risk Analysis

### Technical Risks

**Risk 1: Path Conflicts**
- **Impact**: High (incorrect agent discovery)
- **Probability**: Medium
- **Mitigation**: 
  - Deduplication by agent name
  - Source priority ordering
  - Conflict detection with warnings
  - Unit tests for all scenarios

**Risk 2: Configuration Complexity**
- **Impact**: Medium (user confusion)
- **Probability**: Medium
- **Mitigation**:
  - Sensible defaults
  - Clear documentation
  - Example configurations
  - Error messages for invalid configs

**Risk 3: Performance Degradation**
- **Impact**: Low (discovery is one-time)
- **Probability**: Low
- **Mitigation**:
  - Caching strategy (Phase 2)
  - Parallel path scanning
  - Early termination on errors
  - Benchmark baseline (Week 4)

### Schedule Risks

**Risk 1: Timeline Compression**
- 4 weeks for 5 features
- **Mitigation**: Focus on highest-value features first

**Risk 2: Scope Creep**
- **Mitigation**: Strict scope fence - defer Phase 2 features

## Success Criteria

### Code Quality ✅
- [ ] 85%+ code coverage
- [ ] Zero compilation errors
- [ ] Zero lint issues
- [ ] Clean git history (one commit per task)

### Functionality ✅
- [ ] Multi-path discovery working
- [ ] Configuration system functional
- [ ] All CLI tools operational
- [ ] 90+ tests passing

### Documentation ✅
- [ ] User guide complete
- [ ] API documentation updated
- [ ] Examples provided
- [ ] Phase 1 report written

## Comparison with Phase 0

| Aspect | Phase 0 | Phase 1 |
|--------|---------|---------|
| **Scope** | Single path discovery | Multi-path discovery |
| **Code** | ~500 lines | ~800 lines |
| **Tests** | 17 tests | 90+ tests |
| **Coverage** | 89% | 85%+ |
| **Duration** | 2 weeks | 4 weeks |
| **Complexity** | Low | Medium |
| **Risk** | Low | Low-Medium |

## Next Steps

1. **Code Review**: Get Phase 0 feedback (Week of Dec 2)
2. **Planning Refinement**: Adjust Phase 1 scope if needed
3. **Branch Creation**: Create feature branch from Phase 0
4. **Week 1 Start**: December 9, 2025
5. **Progress Reviews**: Weekly syncs Fridays 3pm PT

## Timeline Summary

```
Dec 9-13   Week 1: Config & Paths        (350-400 lines)
Dec 16-20  Week 2: Metadata             (250-300 lines)
Dec 23-27  Week 3: CLI Enhancement     (300-350 lines)
Dec 30-01  Week 4: Integration & Docs  (200+ lines + docs)
───────────────────────────────────────
Jan 6, 2026    Phase 1 Completion      (1,100-1,300 lines)
```

## Dependencies & Prerequisites

- **Before Phase 1 Start**:
  - Phase 0 merge to main ✅
  - Code review completion
  - No blocking issues

- **External Dependencies**:
  - None (self-contained)

- **Internal Dependencies**:
  - Phase 0 codebase (complete)
  - Existing tool infrastructure (available)

---

*Generated: 2025-11-14*  
*Phase: 1 (Multi-Path Discovery)*  
*Status: 📋 PLANNED*  
*Next Action: Code review of Phase 0, then Phase 1 branch creation*
