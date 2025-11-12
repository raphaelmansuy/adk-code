# Refactoring Validation Checklist

## Purpose
This checklist ensures zero-regression during refactoring. Run after each phase.

---

## Gate 1: Code Quality

```bash
cd /Users/raphaelmansuy/Github/03-working/adk_training_go/code_agent
make fmt      # Format code
make vet      # Run go vet
# make lint   # Run linters (if golangci-lint installed)
```

**Pass Criteria:**
- [ ] No formatting changes needed
- [ ] No go vet warnings
- [ ] No new lint errors (if linter available)

---

## Gate 2: Tests

```bash
make test     # Run all tests
```

**Pass Criteria:**
- [ ] All tests pass
- [ ] No test failures
- [ ] No test skips (unless intentional)

---

## Gate 3: Build

```bash
make clean
make build
```

**Pass Criteria:**
- [ ] Build succeeds without errors
- [ ] Binary created in ../bin/code-agent
- [ ] No compilation warnings

---

## Gate 4: Integration Testing

```bash
# Test help command
./bin/code-agent --help

# Test session commands
./bin/code-agent new-session test-refactor-session
./bin/code-agent list-sessions
./bin/code-agent delete-session test-refactor-session

# Test basic agent interaction (optional - requires API key)
# GOOGLE_API_KEY=<key> ./bin/code-agent
```

**Pass Criteria:**
- [ ] Help displays correctly
- [ ] Session creation works
- [ ] Session listing works
- [ ] Session deletion works
- [ ] Agent starts without errors (if tested)

---

## Gate 5: Code Review Checklist

**For each change:**
- [ ] No unintended file modifications
- [ ] Imports updated correctly
- [ ] No commented-out code added
- [ ] Proper git commit message
- [ ] Changes match plan exactly

---

## Phase-Specific Validation

### Phase 2.1: Orchestration Facades Removal
- [ ] `internal/app/orchestration.go` deleted
- [ ] `internal/app/app_init_test.go` imports updated
- [ ] All references to old functions removed
- [ ] No build errors from missing imports

### Phase 2.2: REPL Facade Removal
- [ ] `internal/app/repl.go` deleted
- [ ] `internal/app/app.go` imports `internal/repl` directly
- [ ] No type alias references remain
- [ ] REPL functionality works correctly

### Phase 2.3: Command Handler Consolidation
- [ ] `cmd/commands/` directory removed
- [ ] `internal/commands/` directory removed
- [ ] `main.go` imports `internal/cli/commands` directly
- [ ] All special commands work (new-session, list-sessions, delete-session)

---

## Rollback Procedure

If any validation gate fails:

```bash
# Review the error
git status
git diff

# Rollback if needed
git reset --hard HEAD~1

# Or stash changes
git stash

# Fix the issue
# ... make corrections ...

# Re-run validation
make check
```

---

## Success Confirmation

After all Phase 2 steps complete:

```bash
# Full validation
make clean
make check
make build

# Verify binary works
./bin/code-agent --help
./bin/code-agent new-session test-final
./bin/code-agent list-sessions | grep test-final
./bin/code-agent delete-session test-final

echo "✅ Phase 2 validation complete!"
```

---

## Notes

- Run validation after EVERY change, not just at phase end
- If a test fails, understand WHY before proceeding
- Document any unexpected issues in docs/draft.md
- Keep validation logs for reference

**Last Updated:** November 12, 2025
