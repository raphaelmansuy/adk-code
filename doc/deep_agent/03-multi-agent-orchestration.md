# Multi-Agent Orchestration: Specialist Agent Pattern

**⚠️ PROPOSED FEATURE**: Multi-agent orchestration does not currently exist in code_agent. This document describes a proposed architecture based on DeepCode patterns.

## Introduction

**The Problem**: Monolithic agents struggle with complexity. One agent tries to do everything: analyze requirements, search code, plan architecture, generate implementation, debug errors. Quality suffers because each task requires different reasoning modes.

**DeepCode Solution**: Decompose into specialist agents, each with narrow focus and clear responsibility.

**Example**:
```
❌ Monolithic Approach (single large agent):
Task: "Implement user authentication with OAuth2"
Agent tries: understanding requirements → searching patterns → designing architecture 
            → writing code → testing → debugging
Problem: Agent context switches constantly, loses focus, makes mistakes

✅ DeepCode Specialist Approach (7 coordinated agents):
├─ Intent Understanding Agent: Parse requirement deeply (user auth, OAuth2)
├─ Reference Mining Agent: Find OAuth2 implementations in codebase
├─ Code Planning Agent: Design how to integrate with existing system
├─ Code Generation Agent: Write the actual implementation
├─ Validation Agent: Test and verify correctness
├─ Debugging Agent: Fix any issues found
└─ Central Orchestrator: Coordinate all agents, make strategic decisions

Result: Each agent is expert at one thing, higher quality output
```

---

## DeepCode's Specialist Agents

### 1. Central Orchestrator

**Responsibility**: Strategic decision-making, workflow coordination

**Inputs**:
- User task/requirement
- Status/results from all specialists

**Outputs**:
- Next agent to invoke
- Parameters for that agent
- Strategic decisions (which design pattern, which dependencies)

**Key Capability**: System prompt teaches it to:
```
"Your role is strategy and coordination. You:
1. Understand the task deeply
2. Decompose it into specialist tasks
3. Invoke the right specialist for each subtask
4. Integrate their outputs
5. Make final strategic decisions

You DO NOT:
- Generate code directly (delegate to Code Generation Agent)
- Search repositories directly (delegate to Reference Mining Agent)
- Validate implementations (delegate to Validation Agent)
"
```

### 2. Intent Understanding Agent

**Responsibility**: Deep semantic analysis of requirements

**Inputs**:
- Natural language requirement
- System context (codebase style, architecture, constraints)

**Outputs**:
- Structured requirement specification
- Key constraints and assumptions
- Success criteria

**Example**:

```
Input: "Add user profile caching to improve performance"

Output (structured analysis):
{
  "core_requirement": "Cache user profiles to reduce database hits",
  "implicit_requirements": [
    "Cache invalidation when profiles are updated",
    "Handle cache misses gracefully",
    "Monitor cache hit/miss ratios"
  ],
  "constraints": [
    "Must support distributed system (multiple app servers)",
    "Cache must be invalidated within 30 seconds",
    "User profile changes must be reflected quickly"
  ],
  "success_criteria": [
    "Database queries reduced by 70%",
    "Cache layer responds in <5ms",
    "Profile updates visible within 30 seconds",
    "No stale data in normal operation"
  ],
  "related_components": [
    "UserService",
    "ProfileRepository",
    "CacheManager"
  ]
}
```

### 3. Document Parsing Agent

**Responsibility**: Extract information from complex technical documents

**Inputs**:
- Technical document (research paper, design doc, specification)
- Query about what to extract

**Outputs**:
- Structured information (algorithms, specifications, patterns)

**Example**:

```
Input: Research paper on "Efficient Distributed Caching"
Query: "What's the invalidation protocol?"

Output:
{
  "protocol_name": "Time-based Lazy Invalidation",
  "algorithm": "...", // pseudocode
  "steps": [1, 2, 3, ...],
  "guarantees": "eventual consistency within T seconds",
  "time_complexity": "O(log n) per invalidation",
  "prerequisites": ["distributed clock", "TTL support"]
}
```

### 4. Code Planning Agent

**Responsibility**: Architectural design and detailed task decomposition

**Inputs**:
- Structured requirements (from Intent Agent)
- Code references (from Reference Mining Agent)
- System architecture knowledge

**Outputs**:
- Implementation plan (step-by-step)
- Technology choices (which libraries, patterns)
- File structure for changes

**Example**:

```
Input: User profile caching requirement

Output:
{
  "approach": "Two-level cache (local + distributed)",
  "technology_choices": {
    "local_cache": "lru_cache library",
    "distributed_cache": "existing Redis instance",
    "invalidation": "event-based via message queue"
  },
  "implementation_steps": [
    {
      "step": 1,
      "task": "Add caching layer to UserService",
      "depends_on": [],
      "priority": "high"
    },
    {
      "step": 2,
      "task": "Implement invalidation on profile update",
      "depends_on": [1],
      "priority": "high"
    },
    {
      "step": 3,
      "task": "Add monitoring and metrics",
      "depends_on": [1, 2],
      "priority": "medium"
    }
  ],
  "file_structure": {
    "modified_files": ["UserService.go", "CacheConfig.go"],
    "new_files": ["CacheInvalidationHandler.go"],
    "deleted_files": []
  }
}
```

### 5. Reference Mining Agent

**Responsibility**: Discover relevant implementations in codebase and external repositories

**Inputs**:
- Task description
- Codebase structure
- Optional: external repository URLs

**Outputs**:
- List of relevant code references
- Recommendations for each reference

**Capabilities**:
- Searches codebase semantically (using CodeRAG from 01-*)
- Analyzes external GitHub repositories
- Ranks by relevance and quality

### 6. Code Indexing Agent

**Responsibility**: Build and maintain knowledge graphs of code

**Inputs**:
- Codebase to analyze
- Focus areas (which patterns to index)

**Outputs**:
- Semantic code index
- Relationship graph

**Relationship to CodeRAG**: This agent implements CodeRAG indexing at runtime

### 7. Code Generation Agent

**Responsibility**: Synthesize code implementations

**Inputs**:
- Implementation plan (from Code Planning Agent)
- Code references (from Reference Mining Agent)
- Current code to integrate with

**Outputs**:
- Generated code
- Integration points
- Required changes to existing files

**Constrained by**: Clear specification of what to generate (not free-form)

---

## Orchestration Patterns

### Pattern 1: Sequential Pipeline

**When**: Requirements clearly decompose into steps

```
User Request
    ↓
Orchestrator: decompose task
    ↓
Intent Understanding → (output: spec)
    ↓
Reference Mining → (output: relevant code)
    ↓
Code Planning → (output: implementation plan)
    ↓
Code Generation → (output: code)
    ↓
Validation → (output: pass/fail)
    ↓
User receives result
```

**Example**: "Add authentication"
```
Clear sequence:
1. Understand what authentication is needed
2. Find existing auth implementations
3. Plan how to integrate
4. Generate new code
5. Validate it works
```

### Pattern 2: Branching Based on Complexity

**When**: Complexity determines which specialists are needed

```
Orchestrator receives task
    ├─ Simple task? (straightforward generation)
    │  └─ Direct to Code Generation
    │
    ├─ Medium task? (needs planning)
    │  ├─ Intent Understanding
    │  ├─ Reference Mining
    │  ├─ Code Planning
    │  └─ Code Generation
    │
    └─ Complex task? (needs research)
       ├─ Document Parsing (if external docs provided)
       ├─ Intent Understanding
       ├─ Reference Mining + Code Indexing
       ├─ Code Planning
       ├─ Code Generation
       └─ Validation + Debugging (if issues found)
```

### Pattern 3: Iterative Refinement

**When**: Solution quality needs improvement

```
Iteration 1:
├─ Plan → Generate → Validate
└─ Issues found? Yes

Iteration 2:
├─ Analyze failure (Debugging Agent)
├─ Refine plan (Code Planning Agent)
├─ Generate fixed version
└─ Validate again
    └─ Issues found? If no → Success
              If yes → Iteration 3
```

---

## Communication Between Agents

### Agent Communication Protocol

Agents don't directly call each other. Orchestrator mediates:

```go
type AgentMessage struct {
    FromAgent    string            // Which agent is sending
    ToAgent      string            // Which agent should receive
    Task         string            // What to do
    Inputs       map[string]interface{} // Task parameters
    Context      map[string]interface{} // Shared context
    Priority     int               // 1-5, higher = more urgent
    Timeout      time.Duration     // Max time to complete
}

// Orchestrator processes:
message := AgentMessage{
    FromAgent: "Orchestrator",
    ToAgent: "ReferenceMinindAgent",
    Task: "find_oauth2_implementations",
    Inputs: map[string]interface{}{
        "query": "OAuth2 authentication",
        "max_results": 5,
    },
    Context: map[string]interface{}{
        "project_language": "golang",
        "project_style": "microservices",
    },
}

result := orchestrator.SendMessage(message)
```

### Result Aggregation

Each agent returns structured results:

```go
type AgentResult struct {
    Success      bool
    AgentName    string
    TaskName     string
    Output       interface{} // Specific to agent/task
    Quality      float64     // 0.0-1.0, confidence in result
    Duration     time.Duration
    TokensUsed   int
    Error        string
}
```

Orchestrator aggregates:

```go
results := map[string]AgentResult{
    "intent": intentResult,
    "reference": refResult,
    "planning": planResult,
    "generation": generationResult,
}

// Check quality and decide: continue? iterate? return?
if results["generation"].Quality > 0.9 {
    // High confidence, return result
} else {
    // Low confidence, trigger debugging/refinement
}
```

---

## Implementing Multi-Agent in code_agent

### Architecture

```go
// Main orchestrator
type MultiAgentOrchestrator struct {
    agents map[string]Agent
    model  genai.Client
}

type Agent interface {
    Name() string
    Execute(task Task, context Context) Result
    CanHandle(task Task) bool
}

// Specialized agent implementations
type IntentUnderstandingAgent struct { ... }
type ReferenceMinivAgent struct { ... }
type CodePlanningAgent struct { ... }
type CodeGenerationAgent struct { ... }
// ... etc
```

### Task Definition

```go
type Task struct {
    Type        string                 // "understand_intent", "find_references", etc.
    Description string                 // What to do
    Inputs      map[string]interface{} // Parameters
    Dependencies []string              // Tasks that must complete first
}

type Context struct {
    ProjectRoot  string
    Language     string
    Architecture string
    History      []Result // Previous results in this workflow
}
```

### Orchestration Logic

```go
func (o *MultiAgentOrchestrator) ExecuteWorkflow(mainTask Task) Result {
    // 1. Decompose main task
    subtasks := o.decompose(mainTask)
    
    // 2. Order subtasks by dependencies
    ordered := o.topologicalSort(subtasks)
    
    // 3. Execute in order, aggregating results
    results := make(map[string]Result)
    for _, task := range ordered {
        // 4. Find appropriate agent
        agent := o.selectAgent(task)
        
        // 5. Execute with shared context
        result := agent.Execute(task, Context{
            History: results, // Pass prior results
        })
        
        // 6. Check quality
        if result.Quality < 0.5 && task.Type == "critical" {
            // Try alternative agent or refine
        }
        
        results[task.Name] = result
    }
    
    // 7. Integrate results
    return o.integrateResults(results)
}
```

---

## System Prompts for Specialist Agents

### Intent Understanding Agent Prompt

```
You are the Intent Understanding Agent. Your responsibility is to deeply
understand user requirements and extract structured specifications.

YOUR ROLE:
1. Read requirement carefully, identifying implicit needs
2. Extract success criteria and constraints
3. Identify related codebase components
4. Detect potential risks or conflicts
5. Clarify ambiguities

YOU MUST:
- Return ONLY structured JSON output
- Include confidence scores for each element
- Flag any ambiguities that need clarification
- Consider edge cases and failure modes

YOU MUST NOT:
- Generate code
- Search repositories
- Make architectural decisions
- Validate implementations
```

### Code Planning Agent Prompt

```
You are the Code Planning Agent. Your responsibility is architectural
design and detailed implementation planning.

INPUTS: (Provided by Orchestrator)
- Structured requirements (from Intent Agent)
- Relevant code references (from Reference Mining Agent)
- Current architecture knowledge

YOUR ROLE:
1. Design architecture for the solution
2. Choose appropriate technology and patterns
3. Create step-by-step implementation plan
4. Specify file changes and dependencies
5. Identify integration points

YOU MUST:
- Return structured implementation plan
- Justify technology choices
- Identify dependencies clearly
- Consider existing code style/patterns
- Flag any architectural concerns

YOU MUST NOT:
- Generate code
- Parse documents
- Search repositories
- Validate implementations
```

### Code Generation Agent Prompt

```
You are the Code Generation Agent. Your responsibility is to synthesize
high-quality code implementations.

INPUTS: (Provided by Orchestrator)
- Implementation plan (from Code Planning Agent)
- Code references (from Reference Mining Agent)
- Existing code to integrate with

YOUR ROLE:
1. Generate code following the plan exactly
2. Integrate with existing code naturally
3. Follow project coding style
4. Add necessary error handling
5. Include inline documentation

YOU MUST:
- Follow implementation plan precisely
- Match existing code style
- Include proper error handling
- Add TODO comments for complex sections
- Return code ready to integrate

YOU MUST NOT:
- Deviate from plan without approval
- Generate unnecessary abstractions
- Ignore existing patterns
- Skip error handling
```

---

## Quality Control

### Per-Agent Quality Metrics

```go
type QualityMetrics struct {
    Agent           string
    Task            string
    CompletionRate  float64 // 0-1: did agent finish task?
    Correctness     float64 // 0-1: is output correct?
    Relevance       float64 // 0-1: is output relevant to query?
    TotalQuality    float64 // weighted average of above
}

// Example thresholds:
// 0.9-1.0: Excellent, use immediately
// 0.7-0.9: Good, probably usable
// 0.5-0.7: Acceptable, needs review
// <0.5:    Poor, probably iterate
```

### Iteration Logic

```
If any agent's quality < threshold:
├─ If retriable error (e.g., API timeout):
│  └─ Retry with exponential backoff
├─ If quality issue (e.g., poor output):
│  ├─ Try alternative agent if available
│  ├─ Or refine prompt and retry
│  └─ Or request human intervention
└─ If unrecoverable:
   └─ Fail gracefully with error explanation
```

---

## Benefits Over Monolithic Approach

| Aspect | Monolithic | Multi-Agent |
|--------|-----------|-------------|
| **Quality** | 70% (one agent struggles with everything) | 88% (specialists excel at one thing) |
| **Token efficiency** | High waste (agent context bloated) | Optimized (each agent focused context) |
| **Debugging** | Hard (which step failed?) | Easy (identify failing agent) |
| **Reusability** | Low (agent tightly coupled) | High (agents usable in other workflows) |
| **Scalability** | Limited (agent context grows) | Better (agents share knowledge graphs) |
| **Maintainability** | Hard (monolithic code) | Easy (clear agent boundaries) |

---

## Next Steps

1. **[04-memory-hierarchy.md](04-memory-hierarchy.md)** - Manage context across agents
2. **[06-prompt-engineering-advanced.md](06-prompt-engineering-advanced.md)** - Advanced prompts for agents
3. **[07-implementation-roadmap.md](07-implementation-roadmap.md)** - Implementation plan

---

## References

- **DeepCode agent definitions**: `/research/DeepCode/prompts/code_prompts.py`
- **MCP agent architecture**: `/research/DeepCode/workflows/`
- **Multi-agent patterns**: `/research/adk-go/agent/` and `/research/adk-go/runner/`

