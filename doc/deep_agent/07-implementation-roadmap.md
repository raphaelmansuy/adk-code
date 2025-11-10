# Implementation Roadmap: Bringing DeepCode Techniques to code_agent

**STATUS AS OF NOVEMBER 2025**: This is a **PROPOSED roadmap**. None of the features described in Phases 0-6 have been implemented yet. The codebase currently uses Google ADK Go with Gemini 2.5 Flash, basic file tools, and no advanced context engineering.

## Current State vs. Proposed State

### What Exists NOW
- ✅ Google ADK Go framework integration (llmagent pattern)
- ✅ Gemini 2.5 Flash model (hardcoded)
- ✅ Basic tools: read_file, write_file, grep_search, list_directory, search_files, execute_command
- ✅ Workspace management (single and multi-workspace support)
- ✅ Display/rendering system with ANSI colors and typewriter effect
- ✅ Tool registry system for dynamic tool registration

### What Does NOT Exist Yet (Proposed in this roadmap)
- ❌ LLMProvider abstraction interface
- ❌ Claude, OpenAI, or local model support
- ❌ CodeRAG (semantic code indexing)
- ❌ Semantic code search
- ❌ Document segmentation
- ❌ Multi-agent orchestration
- ❌ Memory hierarchy (4-level system)
- ❌ Agent-specific prompts
- ❌ Configuration system (YAML-based)

---

## Phased Approach

This roadmap breaks implementation into manageable phases with clear checkpoints.

---

## Phase 0: Foundation (Week 1)

**Goal**: Prepare codebase for new techniques

### 0.1: Add Configuration System

**Task**: Move hardcoded values to config

```
Files modified:
├─ code_agent/config/providers.yaml (new)
├─ code_agent/config/indexing.yaml (new)
├─ code_agent/config/memory.yaml (new)
├─ code_agent/agent/coding_agent.go

Changes:
├─ Load YAML configs at startup
├─ Add config validation
├─ Add sensible defaults
```

**Validation**: 
- Config loads without errors
- All defaults work in basic scenario
- Agent still functions with existing prompts

### 0.2: Add Abstraction Layer for Models

**Task**: Create provider interface (not switching models yet, just abstraction)

```go
// In code_agent/agent/provider_interface.go (new)
type LLMProvider interface {
    Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
    GetCapabilities() ProviderCapabilities
    IsAvailable() bool
}

// Keep Gemini as sole implementation for now
type GeminiProvider struct { ... }
```

**Validation**:
- Agent works exactly as before
- Prompts still route through GeminiProvider
- No behavioral changes

### 0.3: Add Memory Package

**Task**: Create memory hierarchy structure (not using yet)

```
Files created:
├─ code_agent/memory/hierarchy.go (new)
├─ code_agent/memory/level1.go (new, immediate context)
├─ code_agent/memory/level2.go (new, working set)
├─ code_agent/memory/manager.go (new, orchestrator)

Structure: 4-level system, initially unused
```

**Validation**:
- Memory package compiles
- No integration with agent yet
- Tests for memory isolation pass

---

## Phase 1: CodeRAG Foundation (Weeks 2-3)

**Goal**: Implement semantic code indexing and search

### 1.1: Code Indexing Tool

**Task**: Implement `index_codebase` tool

```go
// In code_agent/tools/code_indexing.go (new)
type IndexCodebaseInput struct {
    RepositoryPath string
    OutputPath     string
}

type IndexCodebaseOutput struct {
    Success        bool
    IndexPath      string
    FilesIndexed   int
    RelationshipsFound int
}
```

**Implementation**:
1. File traversal with filtering (exclude vendor, node_modules)
2. For each file: extract semantic profile (using LLM)
3. Store profiles in JSON
4. Compute relationship graph
5. Persist index to disk

**Validation**:
- Index creation completes for sample repo
- Index file is valid JSON
- Relationship graph is sensible

### 1.2: Semantic Search Tool

**Task**: Implement `semantic_code_search` tool

```go
type SemanticCodeSearchInput struct {
    Query         string
    IndexPath     string
    MaxResults    int
    MinConfidence float64
}
```

**Implementation**:
1. Load index from disk
2. Parse query into keywords/concepts
3. Score files based on relationships + keyword match
4. Return sorted results

**Validation**:
- Search completes in <1s for 1000-file repo
- Results are semantically relevant
- Confidence scores make sense

### 1.3: Integration with Agent

**Task**: Agent can use CodeRAG for code discovery

**Workflow**:
```
Agent working on feature X
├─ Calls: semantic_code_search("caching implementation")
├─ Gets: [cache_manager.go (0.95), redis_backend.go (0.88)]
├─ Loads these files for context
└─ Generates code informed by existing patterns
```

**Validation**:
- Agent successfully finds relevant code
- Generated code follows existing patterns
- Quality improves over baseline

---

## Phase 2: Document Segmentation (Week 4)

**Goal**: Handle large files intelligently

### 2.1: Document Analysis

**Task**: Implement document type detection and semantic boundary identification

```go
// In code_agent/tools/document_tools.go (new)
type SegmentDocumentInput struct {
    FilePath    string
    OutputPath  string
}

type DocumentSegment struct {
    ID          string
    Title       string
    Content     string
    ContentType string  // "algorithm", "explanation", etc.
    Keywords    []string
}
```

**Implementation**:
1. Read document
2. Detect type (paper, code, doc, etc.)
3. Identify semantic boundaries
4. Create segments
5. Compute relevance scores

**Validation**:
- Segmentation preserves algorithmic integrity
- Algorithm blocks not split
- Segments are coherent

### 2.2: Segment Retrieval Tool

**Task**: Query-aware segment retrieval

```go
type ReadDocumentSegmentsInput struct {
    SegmentIndexPath string
    QueryType        string  // "concept", "algorithm", "code_planning"
    MaxSegments      int
}
```

**Implementation**:
1. Load segment index
2. Score segments by query type
3. Return top-N, stopping when reaching char limit

**Validation**:
- Retrieved segments directly answer queries
- Token usage reduced significantly
- Quality of answers improves

---

## Phase 3: Multi-Agent Orchestration (Weeks 5-6)

**Goal**: Decompose agent into specialists

### 3.1: Define Specialist Agents

**Task**: Create agent interfaces and base implementations

```go
// In code_agent/agent/specialists/ (new)
type IntentUnderstandingAgent struct { ... }
type ReferenceMinedAgent struct { ... }
type CodePlanningAgent struct { ... }
type CodeGenerationAgent struct { ... }

type Agent interface {
    Execute(task Task, context Context) Result
}
```

**Implementation**:
1. Define agent interfaces
2. Create system prompts for each
3. Implement basic version (delegates to LLM)
4. Add quality metrics

**Validation**:
- Each agent executes independently
- System prompts are clear and effective
- Quality metrics are meaningful

### 3.2: Orchestrator

**Task**: Implement central coordinator

```go
type MultiAgentOrchestrator struct {
    agents  map[string]Agent
    history []Result
}

func (o *Orchestrator) ExecuteWorkflow(mainTask Task) Result {
    // Decompose task
    // Order by dependencies
    // Execute agents
    // Aggregate results
}
```

**Implementation**:
1. Task decomposition logic
2. Dependency resolution
3. Sequential execution
4. Error handling and retries

**Validation**:
- Complex tasks decompose properly
- Agents called in correct order
- Results integrate smoothly

---

## Phase 4: Memory Hierarchy Integration (Week 7)

**Goal**: Connect memory system to agents

### 4.1: Activate Memory Levels

**Task**: Integrate memory hierarchy into agent workflow

```go
// Agent now uses memory manager
agent.memory.SetImmediateContext(currentFile)
agent.memory.PromoteToWorkingSet(relatedFile)
context := agent.memory.GetWorkingSet()  // In prompt
```

**Implementation**:
1. Set Level 1 to current editing file
2. Promote searched files to Level 2
3. Query Level 4 CodeRAG for discovery
4. Track access patterns for promotion/demotion

**Validation**:
- Context stays under token limit
- Memory actually reduces token usage
- Agent performance improves

---

## Phase 5: Provider Abstraction (Week 8)

**Goal**: Support multiple LLM providers

### 5.1: Add Claude Provider

**Task**: Implement Claude as alternative provider

```go
type ClaudeProvider struct { ... }

func (c *ClaudeProvider) Complete(...) { ... }
```

**Implementation**:
1. Add Anthropic SDK dependency
2. Implement provider interface
3. Translate between abstract and Claude APIs
4. Handle differences (streaming, vision, etc.)

**Validation**:
- Claude provider works for same tasks
- Results quality is equivalent or better
- Cost comparison is accurate

### 5.2: Add Provider Selection

**Task**: Implement intelligent routing

```go
providerManager := NewProviderManager(config)
provider := providerManager.SelectProvider(task)
```

**Implementation**:
1. Route by task type
2. Fall back on provider failure
3. Track costs
4. Choose cheapest capable provider

**Validation**:
- Routing is intelligent (right model for task)
- Fallback works when primary unavailable
- Cost tracking is accurate

---

## Phase 6: Advanced Prompting (Week 9)

**Goal**: Upgrade system prompts using DeepCode lessons

### 6.1: Refactor Existing Prompts

**Task**: Apply prompt engineering principles to existing prompts

**Changes**:
- Add clarity on agent responsibilities
- Specify output format explicitly
- Add confidence scoring
- Include examples
- Add constraints

**Validation**:
- Prompts are clearer
- Output quality improves
- Token efficiency increases

### 6.2: Create Agent-Specific Prompts

**Task**: Separate system prompts per agent

**Files**:
```
code_agent/agent/prompts/
├─ intent_understanding.yaml
├─ reference_mining.yaml
├─ code_planning.yaml
├─ code_generation.yaml
└─ default.yaml
```

**Implementation**:
1. Load prompts from YAML
2. Per-agent system prompts
3. Version control for prompts
4. A/B testing framework

**Validation**:
- Agent-specific prompts improve quality
- Versioning allows rollback
- Testing shows improvement metrics

---

## Integration Checkpoints

After each phase, verify:

### Checkpoint 1: Core Functionality
```
Agent still works:
├─ Can read files
├─ Can search repositories
├─ Can generate code
├─ Can execute tools
└─ All tests pass
```

### Checkpoint 2: Quality Metrics
```
Measure across sample tasks:
├─ Code quality (test passing rate)
├─ Context efficiency (tokens per task)
├─ Correctness (does code work?)
├─ Relevance (is code appropriate?)
└─ Compare to baseline
```

### Checkpoint 3: Performance
```
Track for regressions:
├─ Task completion time
├─ API calls per task
├─ Cost per task
├─ Error rate
└─ User satisfaction
```

---

## Timeline Summary

| Phase | Duration | Key Deliverable |
|-------|----------|-----------------|
| 0: Foundation | 1 week | Config + abstraction layers |
| 1: CodeRAG | 2 weeks | Semantic code indexing |
| 2: Segmentation | 1 week | Smart document handling |
| 3: Multi-Agent | 2 weeks | Specialist orchestration |
| 4: Memory | 1 week | Hierarchy integration |
| 5: Providers | 1 week | Multi-model support |
| 6: Prompting | 1 week | Advanced system prompts |
| **Total** | **9 weeks** | **Full transformation** |

---

## Success Criteria

### Quantitative Goals

⚠️ **BASELINE DATA NOT COLLECTED**: No metrics exist yet for the current code_agent. These are aspirational targets based on DeepCode's improvements.

```
Proposed Baseline (To be established):
├─ Success rate on code tasks: TBD
├─ Avg tokens per task: TBD
├─ Cost per task: TBD
└─ User satisfaction: TBD

Target (After Implementation - Based on DeepCode results):
├─ Success rate: 88%+ (estimated 25% improvement over baseline)
├─ Avg tokens: 50% reduction via memory hierarchy
├─ Cost: 66% reduction via multi-provider optimization
└─ User satisfaction: Improved via better context engineering
```

**Action Item**: Establish baseline metrics BEFORE implementing any changes.

### Qualitative Goals

```
✓ Agent understands codebase architecture
✓ Agent reuses existing patterns naturally
✓ Agent handles large files without breaking
✓ Agent works with multiple LLM providers
✓ Codebase is maintainable and extensible
✓ Clear separation of concerns (agents)
```

---

## Risk Mitigation

### Risk: CodeRAG indexing is expensive

**Mitigation**:
- Cache indexes aggressively
- Amortize over multiple uses
- Profile real repositories
- Implement incremental indexing

### Risk: Multi-agent adds complexity

**Mitigation**:
- Start with simple orchestration
- Add agents incrementally
- Comprehensive testing at each step
- Monitor quality metrics

### Risk: Provider abstraction breaks compatibility

**Mitigation**:
- Keep Gemini as default
- Non-breaking changes throughout
- Extensive integration testing
- Easy rollback procedures

---

## Next Steps

1. **Phase 0**: Start configuration system today
2. **Weekly reviews**: Verify checkpoint criteria
3. **Adjust timeline**: Based on actual progress
4. **Gather feedback**: From users and tests

---

## Code Organization After Implementation

```
code_agent/
├─ agent/
│  ├─ coding_agent.go (updated with orchestrator)
│  ├─ enhanced_prompt.go (updated prompts)
│  ├─ orchestrator.go (new, central coordinator)
│  ├─ specialists/ (new)
│  │  ├─ intent_understanding.go
│  │  ├─ reference_mining.go
│  │  ├─ code_planning.go
│  │  └─ code_generation.go
│  └─ prompts/
│     └─ *.yaml (agent-specific prompts)
├─ tools/
│  ├─ code_indexing_tools.go (new CodeRAG tools)
│  ├─ document_tools.go (new segmentation tools)
│  ├─ provider_tools.go (new multi-provider)
│  └─ (existing tools)
├─ providers/ (new)
│  ├─ provider_interface.go
│  ├─ gemini_provider.go
│  ├─ claude_provider.go
│  ├─ openai_provider.go
│  └─ manager.go
├─ memory/ (new)
│  ├─ hierarchy.go
│  ├─ level1.go
│  ├─ level2.go
│  └─ manager.go
├─ config/
│  ├─ providers.yaml (new)
│  ├─ indexing.yaml (new)
│  └─ memory.yaml (new)
└─ (other files)
```

---

## Deployment Strategy

### Development
```
✓ Implement on feature branches
✓ Full test coverage
✓ Code review before merge
✓ Merge to dev branch
```

### Staging
```
✓ Deploy to staging
✓ Run integration tests
✓ Performance benchmarking
✓ Gather telemetry
```

### Production
```
✓ Feature flags (gradual rollout)
✓ Monitor quality metrics
✓ Easy rollback if needed
✓ Gradual traffic increase
```

---

## References

- **Previous documents**: See 01-06 for detailed techniques
- **DeepCode source**: `/research/DeepCode/`
- **ADK patterns**: `/research/adk-go/`

