# Memory Hierarchy: Efficient Context Management

**⚠️ PROPOSED FEATURE**: Memory hierarchy does not currently exist in code_agent. This document describes a proposed 4-level system based on DeepCode patterns.

## Introduction

**The Problem**: LLM context windows are fixed (1M tokens). As agents work with larger codebases, context needs grow exponentially. Simple solutions (just throw more files at the prompt) lead to noise and token waste.

**DeepCode Solution**: Hierarchical memory system with intelligent promotion/demotion policies.

---

## Four-Level Hierarchy

### Level 1: Immediate Context (Always In Context)

**Purpose**: Files agent is actively editing right now

**Content**:
- Current file being modified
- Immediate dependencies (imports, class definitions used)
- Error stack traces from current operation

**Size**: 5-50KB (1-15K tokens)

**Example**:
```
Editing: UserService.go
Context Level 1 includes:
├─ UserService.go (full file)
├─ user_model.go (model definition)
└─ Current error message
```

### Level 2: Working Set (Cached, Accessible)

**Purpose**: Files needed for current task

**Content**:
- Files returned by semantic search (CodeRAG)
- Files identified as dependencies in architecture
- Recently accessed files
- Test files for current code

**Size**: 100-500KB (25-150K tokens)

**Cache policy**: 
- Keep for duration of task (e.g., 30 minutes)
- LRU eviction if cache grows too large
- Manual promotion from Level 3

**Example**:
```
Working on: Add caching feature
Working Set includes:
├─ cache_manager.go (directly needed)
├─ redis_backend.go (dependency)
├─ cache_config.go (configuration)
├─ cache_test.go (for writing tests)
└─ utils/logging.go (used by cache_manager)
```

### Level 3: Archive (Retrieved On Demand)

**Purpose**: Possible references, less frequently accessed

**Content**:
- All other project files (indexed via CodeRAG)
- Related but not immediately needed code
- Historical implementations (for reference)
- Examples and documentation

**Size**: 1MB-100MB (all indexed files)

**Access**: Retrieved on-demand when agent queries for specific patterns

**Example**:
```
Available in Archive but not loaded:
├─ auth_middleware.go (might be needed)
├─ payment_handler.go (different feature)
├─ database_schema.sql (reference)
├─ docs/architecture.md (documentation)
└─ examples/cache_usage.md (example)
```

### Level 4: Global Knowledge (Summarized)

**Purpose**: High-level understanding of entire codebase

**Content**:
- CodeRAG index (semantic relationships, not actual code)
- Architecture overview
- Design patterns used
- Technology stack

**Size**: 10-50KB (highly compressed)

**Type**: Not actual code, but semantic maps

**Example**:
```
Global knowledge (no actual code):
├─ Architecture: Service → Repository → Database
├─ Design patterns: Singleton (Logger), Factory (CacheManager)
├─ Key relationships: UserService uses CacheManager uses Redis
└─ Technology stack: Go 1.21, Redis 7, PostgreSQL 14
```

---

## Memory Management Policies

### Promotion (Moving from lower to higher level)

**Trigger 1: Explicit search**
```
Agent: "Find implementations of cache invalidation"
System:
├─ Query Level 4 (CodeRAG): Find relevant files
├─ Identify: cache_invalidation.go, cache_manager.go
├─ Promote: Move to Level 2 (working set)
└─ Return to agent
```

**Trigger 2: Dependency discovery**
```
Agent reads: cache_manager.go
Analyzer discovers imports:
├─ redis_backend.go (used)
├─ logging_utils.go (used)
Action: Auto-promote dependencies to Level 2
```

**Trigger 3: Frequent access**
```
Agent accesses same file 3+ times in 5 minutes
Action: Move to Level 1 (immediate context)
```

### Demotion (Moving from higher to lower level)

**Trigger 1: Task complete**
```
Agent finishes caching feature
Working set becomes stale
Demotion policy: After task completion, demote files to Level 3
```

**Trigger 2: Time-based expiration**
```
File in Level 2 not accessed for 30 minutes
Demotion: Move to Level 3, keep in Level 4 summary
```

**Trigger 3: Space pressure**
```
Level 2 exceeds 500KB limit
LRU eviction: Remove least recently used file
```

### Memory Pressure Handling

**When Level 1 is full** (> 50KB):
```
Option 1: Spill to Level 2 (file stays accessible)
Option 2: Split large file (break into smaller parts)
Option 3: Compress (summarize file, keep summary in L1)
```

**When Level 2 is full** (> 500KB):
```
Option 1: Move oldest files to Level 3
Option 2: Increase LLM context window if possible
Option 3: Reduce content retention time
```

---

## Integration with CodeRAG

**CodeRAG is Level 4**:

```
Level 1 ──── Direct edits
Level 2 ──── Working on feature  
Level 3 ──── Related files
Level 4 ──── CodeRAG index
            (semantic relationships
             and summaries)
```

**Agent Workflow**:

```
Agent needs context:
├─ Check Level 1: Is it the current file? Yes → use it
├─ Check Level 2: Is it in working set? Yes → use it
├─ Query Level 4: Is it related to task? 
│  ├─ Yes → Retrieve from Level 3, promote to Level 2
│  └─ No → Not needed
```

---

## Implementation

### Memory Manager Interface

```go
type MemoryManager interface {
    // Query operations
    GetImmediateContext() string          // Level 1
    GetWorkingSet() map[string]string     // Level 2
    GetArchiveItem(path string) (string, error) // Level 3
    QueryGlobalKnowledge(query string) string   // Level 4
    
    // Write operations
    SetImmediateContext(content string) error
    PromoteToWorkingSet(filePath string) error
    DemoteToArchive(filePath string) error
    
    // Management
    GetMemoryStats() MemoryStats
    Cleanup() error // Evict expired items
}

type MemoryStats struct {
    Level1Size int64 // bytes
    Level2Size int64
    Level3Size int64 // total available
    Level4Size int64
    HitRate    float64 // % of queries served from cache
}
```

### Promotion Logic

```go
type PromotionPolicy struct {
    AccessThreshold int           // Accesses before promotion
    TimeThreshold   time.Duration // How long before demotion
}

func (mm *MemoryManager) CheckPromotion(filePath string) {
    if mm.getAccessCount(filePath) >= policy.AccessThreshold {
        mm.PromoteToWorkingSet(filePath)
    }
}
```

---

## Cost Analysis

### Without Memory Hierarchy

```
Scenario: Work with 1000 files, 100MB codebase

Agent needs to read 10 files per task:
├─ Option A: Read all 1000 files into context
│  └─ Cost: ~50M tokens per task = $1500+ per task ❌
├─ Option B: Use grep + manual selection
│  └─ Problem: Agent misses relevant files ❌
└─ Option C: Pay for semantic search each time
   └─ Cost: $5-10 per search, multiple per task ❌
```

### With Memory Hierarchy

```
Same scenario with 4-level memory:

Session setup (one time):
├─ Index codebase with CodeRAG: $0.50
└─ Load initial working set: $0.10

Per task (10 tasks):
├─ Query Level 4 (CodeRAG): Free (already computed)
├─ Retrieve Level 3 files: Free (cached locally)
├─ Promote to Level 2: Free (just pointer)
├─ Include in prompt: ~50K tokens = $0.02
└─ Cost per task: $0.02

Total: $0.50 + $0.10 + (10 × $0.02) = $0.80 vs $15,000+
Savings: 18,000x cost reduction ✓
```

---

## Practical Tuning

### For Small Projects (< 1MB)

```
├─ Level 1: Reduce to 10KB (focused work only)
├─ Level 2: Reduce to 100KB (whole project fits)
├─ Level 3: Full project always available
└─ Level 4: Minimal (not needed for small projects)
```

### For Large Projects (> 100MB)

```
├─ Level 1: Keep strict at 30KB (only current file)
├─ Level 2: Increase to 2MB (working feature)
├─ Level 3: Reference Layer (rest of project)
└─ Level 4: Comprehensive CodeRAG index
```

### For Interactive Development

```
Increase time thresholds:
├─ Level 1 demotion: 1 hour (user might come back)
├─ Level 2 demotion: 2 hours (task might continue)
└─ Level 3 retention: Permanent (search from archive)
```

---

## Next Steps

1. **[01-advanced-context-engineering.md](01-advanced-context-engineering.md)** - CodeRAG foundation
2. **[05-llm-provider-abstraction.md](05-llm-provider-abstraction.md)** - Provider flexibility
3. **[07-implementation-roadmap.md](07-implementation-roadmap.md)** - Implementation

---

## References

- **DeepCode memory patterns**: `/research/DeepCode/tools/code_indexer.py`
- **ADK session management**: `/research/adk-go/session/`
