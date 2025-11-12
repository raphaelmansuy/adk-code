# Code Agent Directory Naming Analysis & Opportunities

**Date**: November 12, 2025

## Current Structure Issues

### 🔴 CRITICAL ISSUES (Duplicate/Empty Packages)

#### 1. **Duplicate Session Packages**
```
session/              (Facade - contains manager.go as re-export)
  └── manager.go      → imports from internal/session

internal/session/     (Real implementation)
  ├── manager.go      (actual implementation)
  ├── models.go
  └── persistence/
```

**Problem**: Two packages with same purpose, one is just a facade  
**Impact**: Confusing for developers - which one to import?  
**Solution**: Remove `session/` facade, import `internal/session` directly (breaking change)

---

#### 2. **Duplicate Testutil Packages** 
```
pkg/testutil/         (Real - contains helpers.go)
  └── helpers.go

internal/testutils/   (Empty - no files)
  └── (empty)
```

**Problem**: Empty `internal/testutils/` package wastes structure  
**Impact**: Inconsistent naming (testutil vs testutils)  
**Solution**: Delete `internal/testutils/`, keep `pkg/testutil/`

---

### 🟡 NAMING INCONSISTENCIES

#### 3. **cmd/ Ambiguity**
```
cmd/                  (Abbreviated, purpose unclear)
  └── commands/       (Nested - redundant?)
      └── handlers.go
```

**Problem**: Is `cmd/` short for "command"? Why nested `commands/`?  
**Issue**: Could be confused with command execution (cmd package in Go)  
**Options**:
- Rename `cmd/` → `commands/` (flatten to one level)
- OR rename `cmd/` → `cli_commands/` (clarify purpose)
- OR move to `internal/commands/` (app-specific)

Current: `cmd/commands/handlers.go`  
Better: `commands/handlers.go` or `internal/commands/handlers.go`

---

#### 4. **Agent Package Unclear**
```
agent/                (What is this?)
  └── prompts/        (Ah, system prompts and templates)
```

**Problem**: `agent/` name doesn't indicate it's prompts/templates  
**Better Names**:
- `agent_prompts/` - More specific
- `prompts/` - Shorter, clearer (if unique at root)
- `internal/agent/templates/` - Better organization

---

### 🟠 ORGANIZATIONAL CONFUSION

#### 5. **Display Package Sprawl**
```
display/              (Root level - main package)
  ├── banner/
  ├── components/
  ├── core/          (What's "core"?)
  ├── formatters/
  ├── renderer/
  ├── streaming/
  ├── styles/
  ├── terminal/
  └── tooling/       (Confusing - sounds like tools/)

tools/
  └── display/       (Display tools?)
```

**Problems**:
- `display/tooling/` is confusing (sounds like tools/)
- `display/core/` unclear purpose
- `tools/display/` purpose different from root `display/`

**Recommendations**:
- Rename `display/tooling/` → `display/integration/` or `display/adapters/`
- Document `tools/display/` as "display tool implementations" vs root `display/` as "UI framework"

---

#### 6. **Tools Package Inconsistency**
```
tools/                (Plural)
  ├── common/        (What's "common"?)
  ├── display/
  ├── edit/
  ├── exec/
  ├── file/
  ├── search/
  ├── v4a/
  └── workspace/
```

**Problem**: `tools/common/` is vague  
**Better Options**:
- `tools/base/` - Foundation for tool implementations
- `tools/shared/` - Shared utilities
- `tools/internal/` - Internal to tools package

---

### 🟢 ANALYSIS SUMMARY

| Issue | Severity | Type | Solution |
|-------|----------|------|----------|
| Duplicate `session/` vs `internal/session/` | 🔴 High | Refactor | Delete facade, import internal |
| Unused `internal/testutils/` | 🔴 High | Cleanup | Delete empty package |
| Inconsistent `testutil/` vs `testutils/` | 🔴 High | Naming | Standardize to `testutil/` |
| Ambiguous `cmd/` naming | 🟡 Medium | Refactor | Rename to `commands/` or `internal/commands/` |
| Unclear `agent/` purpose | 🟡 Medium | Naming | Rename to `agent_prompts/` or `prompts/` |
| Confusing `display/tooling/` | 🟡 Medium | Naming | Rename to `display/integration/` |
| Vague `tools/common/` | 🟡 Medium | Naming | Rename to `tools/base/` or `tools/shared/` |

---

## Recommended Refactoring Priority

### **Phase 1: Critical (Delete/Consolidate)**
1. Delete `session/` facade - import `internal/session` directly
2. Delete `internal/testutils/` - it's empty
3. Standardize to `pkg/testutil/` (already correct)

**Impact**: 3 files deleted, 1-2 import updates  
**Risk**: Medium (facade removal, check all callers)

---

### **Phase 2: High Value (Rename)**
4. Rename `cmd/` → `internal/commands/` (clarify app-specific)
5. Rename `agent/` → `agent_prompts/` (clarify purpose)
6. Rename `tools/common/` → `tools/base/` (clearer intent)

**Impact**: 9 files moved, 5-7 import updates  
**Risk**: High (multiple moves and imports)

---

### **Phase 3: Documentation (Low Risk)**
7. Rename `display/tooling/` → `display/integration/`
8. Document difference: root `display/` = UI framework, `tools/display/` = tool adapters

**Impact**: 1 directory rename, ~3 import updates  
**Risk**: Low

---

## Opportunity Summary

```markdown
QUICK WINS (Low Risk):
- Delete internal/testutils/ (empty)
- Delete session/ (facade - import internal/session)

MEDIUM EFFORT:
- Rename cmd/ → internal/commands/
- Rename agent/ → agent_prompts/
- Rename tools/common/ → tools/base/

NICE TO HAVE:
- Rename display/tooling/ → display/integration/
```

## Recommendation

**Start with Phase 1** (critical cleanup):
1. Remove `session/` facade
2. Delete `internal/testutils/`
3. Update 1-2 imports

This gives immediate clarity without major refactoring effort.

**Then Phase 2** (naming consistency):
- Tackle `cmd/`, `agent/`, `tools/common/` when convenient

**Document Phase 3** for when you have time.
