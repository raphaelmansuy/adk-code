# Refactoring Action Checklist

**Start Date**: ________  
**Team Members**: ________  
**Target Completion**: ________

---

## Pre-Flight Checklist

### Before Starting ANY Refactoring

- [ ] Read `docs/REFACTOR_SUMMARY.md` (executive summary)
- [ ] Read `docs/refactor_plan.md` (detailed implementation)
- [ ] Review `docs/draft.md` (analysis notes)
- [ ] Create dedicated feature branch: `git checkout -b refactor/phase-0`
- [ ] Ensure all current tests pass: `make test`
- [ ] Ensure code quality checks pass: `make check`
- [ ] Backup current working state: `git tag pre-refactor-backup`

---

## Phase 0: Safety Net (CRITICAL - DO FIRST)

**Objective**: Establish test coverage BEFORE making any changes  
**Time Estimate**: 2-3 days  
**Risk**: None (only adds tests)

### Week 1, Day 1-2: Add Core Tests

- [ ] Create `internal/app/testing/` directory
  - [ ] `fixtures.go` - Test data and sample configs
  - [ ] `mocks.go` - Mock implementations
  - [ ] `helpers.go` - Test utilities

- [ ] Create `internal/app/app_test.go`
  - [ ] `TestApplication_New_Success` - Happy path
  - [ ] `TestApplication_New_InvalidConfig` - Error cases
  - [ ] `TestApplication_InitializeDisplay` - Display setup
  - [ ] `TestApplication_InitializeModel` - Model creation
  - [ ] `TestApplication_InitializeAgent` - Agent creation
  - [ ] `TestApplication_InitializeSession` - Session management
  - [ ] `TestApplication_Close` - Cleanup

- [ ] Create `internal/app/repl_test.go`
  - [ ] `TestREPL_New_Success` - REPL creation
  - [ ] `TestREPL_Run_UserInput` - Input handling
  - [ ] `TestREPL_Run_Commands` - Built-in commands
  - [ ] `TestREPL_Run_ContextCancellation` - Ctrl+C handling

- [ ] Create `internal/app/session_test.go`
  - [ ] `TestSessionInitializer_NewSession` - Session creation
  - [ ] `TestSessionInitializer_ResumeSession` - Session resumption

### Verification Checkpoint

- [ ] Run `go test ./internal/app/...` - All tests pass
- [ ] Check coverage: `go test -cover ./internal/app/...` - Target: 80%+
- [ ] Commit: `git commit -m "test: Add comprehensive internal/app tests"`
- [ ] **STOP HERE if coverage < 80%** - Add more tests first!

---

## Phase 1: Structural Improvements

**Objective**: Reduce Application complexity  
**Time Estimate**: 1 day  
**Risk**: LOW (internal changes only, tests verify behavior)

### Week 1, Day 3: Create Component Groupings

- [ ] Create `internal/app/components.go`

```go
// DisplayComponents groups all display-related fields
type DisplayComponents struct {
    Renderer       *display.Renderer
    BannerRenderer *display.BannerRenderer
    Typewriter     *display.TypewriterPrinter
    StreamDisplay  *display.StreamingDisplay
}

// ModelComponents groups all model-related fields
type ModelComponents struct {
    Registry *models.Registry
    Selected models.Config
    LLM      model.LLM
}

// SessionComponents groups all session-related fields
type SessionComponents struct {
    Manager *persistence.SessionManager
    Runner  *runner.Runner
    Tokens  *tracking.SessionTokens
}
```

- [ ] Update `Application` struct in `app.go`
- [ ] Update `initializeDisplay()` to return `DisplayComponents`
- [ ] Update `initializeModel()` to return `ModelComponents`
- [ ] Update `initializeSession()` to return `SessionComponents`
- [ ] Update all references throughout codebase
  - [ ] `internal/app/repl.go`
  - [ ] `internal/app/session.go`
  - [ ] Any other files that access Application fields

### Verification Checkpoint

- [ ] Run `make test` - All tests pass
- [ ] Run `make check` - No new warnings
- [ ] Manual smoke test - App still works
- [ ] Commit: `git commit -m "refactor: Group Application components"`

### Week 1, Day 4: Simplify REPL Configuration

- [ ] Update `REPLConfig` in `repl.go`

```go
type REPLConfig struct {
    UserID      string
    SessionName string
    Display     *DisplayComponents
    Session     *SessionComponents
    Model       ModelComponents
}
```

- [ ] Update `NewREPL()` calls in `app.go`
- [ ] Update REPL methods that access config

### Verification Checkpoint

- [ ] Run `go test ./internal/app/...` - All tests pass
- [ ] Run `make check` - Clean
- [ ] Commit: `git commit -m "refactor: Simplify REPLConfig"`

### Week 1, Day 5: Group CLI Configuration

- [ ] Update `pkg/cli/config.go`

```go
type CLIConfig struct {
    Display DisplayConfig
    Session SessionConfig
    Model   ModelConfig
    AI      AIConfig
}

type DisplayConfig struct { ... }
type SessionConfig struct { ... }
type ModelConfig struct { ... }
type AIConfig struct { ... }
```

- [ ] Update `flags.go` to populate nested structs
- [ ] Update all references in `app.go`
- [ ] Update any other files using CLIConfig

### Verification Checkpoint

- [ ] Run `make test` - All tests pass
- [ ] Test CLI flags still work: `./code-agent --help`
- [ ] Commit: `git commit -m "refactor: Group CLI configuration"`

---

## Phase 2: Code Organization

**Objective**: Move code to appropriate packages  
**Time Estimate**: 2 hours  
**Risk**: VERY LOW (simple moves)

### Week 2, Day 1 Morning: Move GetProjectRoot

- [ ] Create `workspace/project_root.go`
- [ ] Move `GetProjectRoot()` function from `agent/coding_agent.go`
- [ ] Update import in `agent/coding_agent.go`
- [ ] Add tests in `workspace/project_root_test.go`

### Verification Checkpoint

- [ ] Run `go test ./workspace/...` - All tests pass
- [ ] Run `go test ./agent/...` - All tests pass
- [ ] Commit: `git commit -m "refactor: Move GetProjectRoot to workspace"`

### Week 2, Day 1 Afternoon: Create Display Factory

- [ ] Create `display/factory.go`

```go
type Config struct {
    OutputFormat      string
    TypewriterEnabled bool
}

type Components struct {
    Renderer       *Renderer
    BannerRenderer *BannerRenderer
    Typewriter     *TypewriterPrinter
    StreamDisplay  *StreamingDisplay
}

func NewComponents(cfg Config) (*Components, error) { ... }
```

- [ ] Update `app.go` to use `display.NewComponents()`
- [ ] Remove inline initialization code

### Verification Checkpoint

- [ ] Run `go test ./display/...` - All tests pass
- [ ] Run `go test ./internal/app/...` - All tests pass
- [ ] Commit: `git commit -m "refactor: Add display component factory"`

---

## Phase 3: Test Coverage Expansion

**Objective**: Comprehensive test coverage  
**Time Estimate**: 3-4 days  
**Risk**: None (only adds tests)

### Week 2, Day 2-3: Display Package Tests

- [ ] Create `display/renderer_test.go`
  - [ ] Test markdown rendering
  - [ ] Test ANSI color handling
  - [ ] Test TTY vs non-TTY behavior

- [ ] Create `display/formatters/tool_test.go`
  - [ ] Test tool call formatting
  - [ ] Test tool result formatting
  - [ ] Test error formatting

- [ ] Create `display/components/timeline_test.go`
  - [ ] Test event timeline creation
  - [ ] Test event rendering

### Verification Checkpoint

- [ ] Run `go test -cover ./display/...` - Target: 60%+
- [ ] Commit: `git commit -m "test: Add display package tests"`

### Week 2, Day 4-5: Agent Package Tests

- [ ] Create `agent/coding_agent_test.go`
  - [ ] Test agent creation
  - [ ] Test tool registration
  - [ ] Test workspace initialization
  - [ ] Test prompt building

- [ ] Expand `agent/xml_prompt_builder_test.go`
  - [ ] More comprehensive XML validation
  - [ ] Edge cases

### Verification Checkpoint

- [ ] Run `go test -cover ./agent/...` - Target: 70%+
- [ ] Commit: `git commit -m "test: Add agent package tests"`

---

## Phase 4: Polish (Optional)

**Objective**: Code quality improvements  
**Time Estimate**: 1 week  
**Risk**: LOW

### Standardize Error Handling

- [ ] Review all error returns
- [ ] Ensure consistent use of `%w` for wrapping
- [ ] Add context to all errors
- [ ] Define sentinel errors where appropriate

### Add Documentation

- [ ] Add package-level docs to all packages
- [ ] Document all public functions
- [ ] Add usage examples
- [ ] Update README if needed

### Extract Long Functions

- [ ] Identify functions > 50 lines
- [ ] Extract logical sub-operations
- [ ] Add tests for extracted functions

---

## Final Verification

### Before Merging to Main

- [ ] Run full test suite: `make test`
- [ ] Run code quality checks: `make check`
- [ ] Check test coverage: `go test -cover ./...`
- [ ] Manual testing of all features:
  - [ ] Agent starts successfully
  - [ ] Can create new session
  - [ ] Can resume session
  - [ ] Can list sessions
  - [ ] Tool execution works
  - [ ] Model selection works
  - [ ] Ctrl+C handling works
  - [ ] All CLI flags work
- [ ] Performance check (no regressions)
- [ ] Memory usage check (no leaks)

### Documentation

- [ ] Update CHANGELOG.md with refactoring summary
- [ ] Update any affected documentation
- [ ] Add migration notes if needed (should be none!)

### Code Review

- [ ] Self-review all changes
- [ ] Team code review
- [ ] Address review feedback
- [ ] Final approval from tech lead

### Merge & Release

- [ ] Merge to main: `git merge refactor/phase-0`
- [ ] Tag release: `git tag v1.1.0-refactored`
- [ ] Push: `git push origin main --tags`
- [ ] Monitor for issues in first 24 hours

---

## Rollback Procedure (If Needed)

### If Tests Fail

1. **DON'T PANIC** - This is why we have tests
2. Identify which test is failing
3. Check git diff to see what changed
4. Either fix the issue or revert the last commit
5. Re-run tests until all pass

### If Behavior Changed Unexpectedly

1. Stop immediately
2. Rollback to pre-refactor state: `git checkout pre-refactor-backup`
3. Review what went wrong
4. Fix the issue in a new commit
5. Re-test thoroughly

### Emergency Rollback

```bash
# If in production and something breaks
git revert HEAD~N  # Revert last N commits
# OR
git reset --hard pre-refactor-backup
git push --force  # Only if absolutely necessary!
```

---

## Success Metrics

### Quantitative

- [ ] Test coverage: internal/app from 0% → 80%+
- [ ] Test coverage: overall from ~40% → 70%+
- [ ] Application struct: 15 fields → 7 fields (53% reduction)
- [ ] REPLConfig: 10 fields → 5 fields (50% reduction)
- [ ] All tests pass (100%)
- [ ] Zero new lint warnings

### Qualitative

- [ ] Code is easier to understand
- [ ] Components are more testable
- [ ] Structure is more logical
- [ ] Team feedback is positive
- [ ] Onboarding new developers is easier

---

## Notes & Observations

Use this space to document:
- Issues encountered
- Deviations from the plan
- Time taken for each phase
- Team feedback
- Lessons learned

---

**Remember**: 
- Test first, refactor second
- Small changes, frequent commits
- If in doubt, stop and ask
- Our reputation is at stake - quality over speed
