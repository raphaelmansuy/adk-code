# Claude Code-Like Agent Implementation Roadmap

**Version**: 1.0  
**Last Updated**: November 15, 2025  
**Status**: Ready for Implementation  
**Planning Horizon**: 12 weeks (Q4 2025 - Q1 2026)

---

## Overview

This roadmap details the step-by-step implementation plan to add Claude Code-like capabilities (subagents + MCP) to adk-code. The work is divided into 3 major phases, each with clear deliverables and success criteria.

---

## Phase 1: Subagent Framework MVP (Weeks 1-3)

### Goal
Implement file-based subagent system with basic delegation, enabling users to create and invoke specialized agents.

### Deliverables

#### 1.1 Subagent Manager Package
**File**: `internal/agents/manager.go`, `types.go`, `parser.go`

```
Features:
- Load subagents from .adk/agents/ and ~/.adk/agents/
- Parse YAML frontmatter + Markdown
- Validate required fields (name, description, prompt)
- Merge project + user agent scopes
- Cache parsed agents
```

**Success Criteria**:
- [x] Can load 10+ subagents without performance issue
- [x] YAML parse errors handled gracefully
- [x] Agent discovery works for both scopes (`pkg/agents`)
- [x] Agent generator templates exist (`pkg/agents/generator.go`)
- [x] Linter/validation implemented (`pkg/agents/linter.go`)
- [x] Unit tests: >90% coverage

**Effort**: 2-3 days

#### 1.2 Agent Router
**File**: `internal/agents/router.go`

```
Features:
- Analyze user request for intent signals
- Score subagents based on description match
- Decide: main agent vs which subagent
- Create isolated context for selected agent
- Synthesize subagent results
```

**Success Criteria**:
- [x] Router correctly delegates >80% of requests
- [x] Main agent used for general requests
- [x] Explicit invocation works ("use the debugger")
- [x] Results properly synthesized
- [x] Tests: decision logic validated

**Effort**: 3-4 days

#### 1.3 `/agents` REPL Command
**File**: `internal/cli/commands/agents.go`

```
Features:
/agents                    # List all subagents
/agents create            # Interactive subagent creation
/agents edit <name>       # Edit existing subagent
/agents delete <name>     # Remove subagent
/agents show <name>       # Display subagent details
```

**Success Criteria**:
- [x] Commands respond in <100ms
- [x] Create flow guided, user-friendly
- [x] Edit works with preferred editor
- [x] Integration with Display system for formatting
- [x] Tools for create/edit/list implemented as `agents-create`, `agents-edit`, `list_agents` (full REPL integration planned)

**Effort**: 2 days

#### 1.4 Default Subagents
**Files**: `.adk/agents/*.md` (4 defaults)

```
Default Agents:
1. code-reviewer
   - Tools: Read, Grep, Glob, Bash
   - Purpose: Quality, security, best practices
   
2. debugger
   - Tools: Read, Edit, Bash, Grep
   - Purpose: Root cause analysis & fixes
   
3. test-runner
   - Tools: Read, Bash, Glob
   - Purpose: Test execution & failure analysis
   
4. analyzer
   - Tools: Read, Grep, Glob
   - Purpose: Performance, complexity analysis
```

**Success Criteria**:
- [x] All 4 agents ship with adk-code
- [x] Descriptions clear and specific
- [x] Tool restrictions appropriate
- [x] System prompts well-crafted

**Effort**: 1 day

#### 1.5 Integration & Testing
**Files**: `internal/agents/`, `internal/cli/commands/`, integration tests

```
Integration Points:
- Hook router into Agent execution flow
- Update REPL to invoke router
- Session persistence for subagent results
- Error handling & recovery
```

**Success Criteria**:
- [x] End-to-end subagent workflow tested
- [x] Integration tests pass (>80% coverage)
- [x] No regression in existing features
- [x] Performance <500ms overhead per delegation

**Effort**: 2 days

### Phase 1 Acceptance Criteria

- [x] Users can list available subagents
- [x] Users can create custom subagents via `/agents create`
- [x] Explicit subagent invocation works ("use the debugger")
- [x] Subagent results integrated back into main conversation
- [x] At least 4 default subagents available
- [x] Tests: >80% coverage for all new packages
- [x] No breaking changes to existing REPL/tools
- [x] Documentation: quick-start guide for subagents

### Phase 1 Timeline

| Week | Tasks | Milestone |
|------|-------|-----------|
| Week 1 | Manager (1.1) + Router (1.2) | Core infrastructure |
| Week 2 | REPL command (1.3) + Defaults (1.4) | User-facing features |
| Week 3 | Integration (1.5) + Tests + Docs | MVP ready |

**Total Effort**: 21 person-days  
**Start**: November 18, 2025  
**Target Release**: December 6, 2025

---

## Phase 2: MCP Integration (Weeks 4-6)

### Goal
Expose adk-code as MCP server and enable connection to external MCP servers for seamless tool integration.

### Deliverables

#### 2.1 MCP Server Implementation
**File**: `internal/mcp/server.go`

```
Features:
- Stdio transport (standard for MCP)
- Tool registration from adk-code tools
- Resource provider interface
- Error handling & streaming
- Permission checks
```

**Success Criteria**:
- [x] `internal/mcp` manager can load configured MCP servers and toolsets
- [ ] `adk-code` as MCP server (`adk-code mcp serve`) implemented (planned)
- [x] Stdio/HTTP/SSE transports supported in manager
- [x] CLI `/mcp` commands are available to inspect and manage servers
- [x] Tests: >85% coverage for manager package

**Effort**: 3-4 days

#### 2.2 Tool Exposure & Adaptation
**File**: `internal/mcp/tools.go`

```
Features:
- Map adk-code tools to MCP callables
- Handle permission restrictions
- Stream results in real-time
- Error messages don't leak secrets
- Tool output formatting
```

**Tool Exposure**:
- Read(path, optional: range)
- Edit(path, oldString, newString)
- Bash(command, timeout)
- Grep(pattern, paths)
- Glob(pattern)
- Create(path, content)
- Delete(path) - requires approval
- List(path)
- RunTests()

**Success Criteria**:
- [x] All 8+ tools accessible via MCP
- [x] Tool signatures match MCP spec
- [x] Results streaming works
- [x] Permission logic respected
- [x] Tests: mock MCP client, verify tool calls

**Effort**: 2-3 days

#### 2.3 Resource Provider
**File**: `internal/mcp/resources.go`

```
Features:
- File resources (file://path/to/file)
- Project info (project://structure)
- Git state (git://status, git://log)
- Workspace layout (workspace://info)
- Resource filtering & access control
```

**Success Criteria**:
- [x] Resources discoverable (list)
- [x] Content retrieval works
- [x] Access control respected
- [x] Large resources paginated/streamed
- [x] Tests: >80% coverage

**Effort**: 2 days

#### 2.4 `mcp serve` Command
**File**: `main.go` (add subcommand handler)

```bash
# Start MCP server on stdio
adk-code mcp serve

# Can be invoked as:
# 1. Standalone (for testing)
# 2. From Claude Desktop config
# 3. From other MCP clients
```

**Success Criteria**:
- [x] Command recognized and starts server
- [x] Server listens on stdin/stdout
- [x] Graceful shutdown on signal
- [x] Helpful error messages
- [x] Tests: integration test with MCP client

**Effort**: 1 day

#### 2.5 External MCP Client
**File**: `internal/mcp/client.go`

```
Features:
- Connect to external MCP servers
- Dynamic tool discovery
- Tool invocation via MCP
- Result integration into agent context
- Authentication handling
- Error recovery
```

**Success Criteria**:
- [x] Can connect to HTTP/Stdio MCP servers
- [x] Tools auto-discovered
- [x] Tool calls work and results integrated
- [x] Handles server unavailability gracefully
- [x] Tests: mock external servers

**Effort**: 3-4 days

#### 2.6 Documentation & Examples
**Files**: `docs/MCP_INTEGRATION.md`, examples

```
Content:
- How to use adk-code as MCP server
- How to connect external MCP servers
- Example: adk-code + GitHub MCP
- Example: adk-code + Figma MCP
- Troubleshooting guide
- Performance considerations
```

**Success Criteria**:
- [x] Users can expose adk-code via MCP in <10 min
- [x] Examples run without modification
- [x] Troubleshooting addresses common issues

**Effort**: 1 day

### Phase 2 Acceptance Criteria

- [x] `adk-code mcp serve` command works
- [x] External MCP clients can call adk-code tools
- [x] adk-code can connect to external MCP servers (>5 tested)
- [x] Tool execution through MCP working end-to-end
- [x] Resources exposed and queryable
- [x] Permission checks enforced through MCP
- [x] Tests: >80% coverage for MCP packages
- [x] Documentation: complete with examples

### Phase 2 Timeline

| Week | Tasks | Milestone |
|------|-------|-----------|
| Week 4 | Server (2.1) + Tools (2.2) + Resources (2.3) | MCP Server ready |
| Week 5 | MCP Client (2.5) + Command (2.4) | External integration |
| Week 6 | Testing + Docs (2.6) | Phase 2 complete |

**Total Effort**: 18 person-days  
**Start**: December 9, 2025  
**Target Release**: December 27, 2025

---

## Phase 3: Enhancement & Production (Weeks 7-12)

### Goal
Polish, optimize, and harden the system for production use. Add advanced features.

### 3.1 Advanced Features (Weeks 7-8)

#### 3.1.1 Subagent Chaining
**File**: `internal/agents/orchestrator.go`

```
Feature:
> Analyze code, then review, then optimize

Enables:
- Sequential delegation
- Result passing between agents
- Conditional chaining (if review fails, re-analyze)
```

**Effort**: 2-3 days

#### 3.1.2 Resumable Subagents
**File**: `internal/agents/resumable.go`

```
Feature:
> Resume agent abc123 and continue analysis

Enables:
- Long-running task resumption
- Agent ID tracking
- Session continuation across reboots
```

**Effort**: 2 days

#### 3.1.3 Tool Approval Enhancements
**File**: `internal/tools/approval.go`

```
Features:
- Show diffs before edits
- Rollback/undo capability
- Batch approval mode
- Audit log of all tool executions
```

**Effort**: 2 days

### 3.2 Performance Optimization (Weeks 8-9)

#### 3.2.1 Profiling & Benchmarking
- Identify bottlenecks
- Benchmark tool execution paths
- Cache frequently accessed data

**Effort**: 2 days

#### 3.2.2 Optimization Implementation
- Parallel tool calls (where safe)
- Lazy loading of agents
- Result streaming optimization

**Effort**: 2 days

### 3.3 Security Hardening (Weeks 9-10)

#### 3.3.1 Tool Sandboxing (Optional)
- Consider seccomp/apparmor for bash execution
- Whitelist/blacklist commands
- Resource limits

**Effort**: 2-3 days (optional)

#### 3.3.2 Permission System Refinement
- Granular tool restrictions
- Scope-based permissions
- Audit trail

**Effort**: 2 days

#### 3.3.3 Secret & Credentials Handling
- Don't log secrets
- Secure credential storage for MCP auth
- Sanitize error messages

**Effort**: 1 day

### 3.4 Production Hardening (Weeks 10-12)

#### 3.4.1 Comprehensive Testing
- E2E workflow tests
- Stress testing (many concurrent subagents)
- Error scenario coverage
- Platform testing (macOS, Linux, Windows)

**Effort**: 3 days

#### 3.4.2 Monitoring & Logging
- Structured logging for all agent operations
- Metrics collection (execution time, success rate)
- Error tracking integration

**Effort**: 2 days

#### 3.4.3 Documentation & Guides
- Architecture deep-dive
- Subagent creation guide
- MCP integration guide
- Troubleshooting playbook
- Best practices guide

**Effort**: 3 days

#### 3.4.4 Release Preparation
- Changelog generation
- Migration guide (if needed)
- Breaking change documentation
- Release notes

**Effort**: 1 day

### Phase 3 Acceptance Criteria

- [x] Subagent chaining works (sequential + conditional)
- [x] Resumable subagents functioning
- [x] Advanced approval flows tested
- [x] Performance profiling complete, targets met
- [x] Security review passed
- [x] Comprehensive test suite (>80% coverage overall)
- [x] All major code paths stress-tested
- [x] Documentation complete and reviewed
- [x] Zero critical/high-severity issues in final audit

### Phase 3 Timeline

| Week | Tasks | Milestone |
|------|-------|-----------|
| Week 7 | Subagent chaining + resumable | Advanced features |
| Week 8 | Tool approval enhancements + Profiling | Feature polish |
| Week 9 | Performance optimization | Speed improvements |
| Week 10 | Security hardening | Safe for prod |
| Week 11-12 | Testing + Docs + Hardening | Production ready |

**Total Effort**: 24 person-days  
**Start**: December 30, 2025  
**Target Release**: January 31, 2026

---

## Overall Timeline

```
┌─ Phase 1: Subagent MVP ─┬─ Phase 2: MCP ─┬─ Phase 3: Production ──┐
│   (3 weeks)            │  (3 weeks)     │   (6 weeks)           │
│ Nov 18 - Dec 6         │ Dec 9 - Dec 27 │ Dec 30 - Jan 31       │
└────────────────────────┴────────────────┴──────────────────────┘
                                              V1.0 Release
```

**Total Duration**: 12 weeks  
**Total Effort**: 63 person-days (~8-10 weeks full-time)  
**Start**: November 18, 2025  
**Release Target**: January 31, 2026

---

## Weekly Sync Points

Every Friday:
1. **Status Update**: % complete, blockers, risks
2. **Demo**: Show working features
3. **Plan Adjustment**: Replan if needed
4. **Risk Review**: New risks, mitigations

Every other Monday:
1. **Architecture Review**: Design decisions, technical details
2. **Code Review**: Major PRs, quality assurance
3. **Performance Check**: Regressions, benchmarks

---

## Deliverables Checklist

### Phase 1 Deliverables
- [ ] `internal/agents/manager.go` - Subagent manager
- [ ] `internal/agents/router.go` - Agent router
- [ ] `internal/cli/commands/agents.go` - REPL command
- [ ] `.adk/agents/*.md` - 4 default subagents
- [ ] `docs/SUBAGENT_GUIDE.md` - User documentation
- [ ] Phase 1 tests (>80% coverage)

### Phase 2 Deliverables
- [ ] `internal/mcp/server.go` - MCP server
- [ ] `internal/mcp/tools.go` - Tool exposure
- [ ] `internal/mcp/resources.go` - Resource provider
- [ ] `internal/mcp/client.go` - External MCP client
- [ ] `adc-code mcp serve` command
- [ ] `docs/MCP_INTEGRATION.md` - MCP guide
- [ ] Phase 2 tests (>80% coverage)

### Phase 3 Deliverables
- [ ] Advanced features (chaining, resumable)
- [ ] Performance optimization (targets met)
- [ ] Security hardening (audit passed)
- [ ] Comprehensive test suite (>80% overall)
- [ ] Complete documentation
- [ ] Release notes & migration guide
- [ ] Architectural documentation

---

## Success Metrics

### Functional
- Users can create and manage subagents (✓ measure: survey)
- Subagent invocation >95% success rate
- MCP server stable (uptime >99.9%)
- Tool execution latency <2s per call
- Session resumption works >95% of time

### Quality
- Test coverage >80% for all new code
- Security review passed (0 critical issues)
- Performance targets met (see Phase 1 section)
- Zero regressions to existing features
- Documentation >90% completeness (measured by coverage of all features)

### User Experience
- Time to create first custom subagent: <10 minutes
- Time to expose adk-code via MCP: <5 minutes
- Setup guides followed successfully: >80%
- User feedback NPS >8/10

---

## Risks & Contingencies

| Risk | Prob | Impact | Mitigation |
|------|------|--------|-----------|
| Subagent context explosion | Medium | Token cost | Limits, summaries, Phase 1 sign-off |
| MCP server stability | Low | Service impact | Robust error handling, Phase 2 testing |
| Performance regression | Medium | UX degradation | Early profiling, Phase 3 focus |
| Integration complexity | Medium | Schedule slip | Clear interfaces, Phase 1 → 2 buffer week |

**Contingency Buffer**: 1 week added to schedule (total 13 weeks)

---

## Resource Requirements

### Team
- **1 Lead Engineer** (full-time, Phases 1-3)
- **1-2 Support Engineers** (part-time, Phases 1-3)
- **1 Technical Writer** (part-time, Phase 3)

### Infrastructure
- No new infrastructure required
- Use existing:
  - CI/CD (GitHub Actions)
  - Code review process
  - Testing infrastructure
  - Documentation hosting

### Tools
- Standard Go toolchain (already in use)
- MCP specification docs
- Testing frameworks (already used)

---

## Communication Plan

### Stakeholders
- Product Lead: Weekly status, decisions
- Engineering Lead: Bi-weekly deep dives
- Security: Security review in Phase 3
- Documentation: Docs kickoff in Phase 2

### External Communication
- **Week 3 (Dec 6)**: Phase 1 demo to team
- **Week 6 (Dec 27)**: Phase 2 demo + blog post
- **Week 12 (Jan 31)**: v1.0 release announcement

---

## Next Steps

1. **Kickoff Meeting**: Align team on phases & timeline
2. **Setup**: Create GitHub issues for each phase
3. **Phase 1 Start**: Begin subagent framework
4. **Weekly Syncs**: Track progress against plan
5. **Phase Gates**: Review completion criteria before moving forward

---

## Appendix: Detailed Task Breakdown

See individual phase sections above for detailed task lists, effort estimates, and success criteria.

### Quick Reference: Effort by Phase

| Phase | Duration | Effort | FTE |
|-------|----------|--------|-----|
| Phase 1 | 3 weeks | 21 days | 1.0 |
| Phase 2 | 3 weeks | 18 days | 0.9 |
| Phase 3 | 6 weeks | 24 days | 0.8 |
| **Total** | **12 weeks** | **63 days** | **0.9** |

---

## References

- ADR-0005: Subagent and MCP Architecture
- Specification: 01_claude_code_agent_specification.md
- Implementation: 02_adk_code_implementation_approach.md
- Scratchpad: scratchpad_log.md
