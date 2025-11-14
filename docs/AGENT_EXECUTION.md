# Agent Execution System Documentation

## Overview

The Agent Execution System is a comprehensive framework for discovering, validating, managing, and executing AI agents within the ADK CLI. It provides a two-layer architecture separating pure utilities from ADK tool integrations, enabling robust agent definition support with dependency resolution, semantic versioning, and metadata validation.

**Key Features:**
- **Agent Discovery**: Automatically discover agents from filesystem and metadata
- **Execution Management**: Execute agents with context management and streaming output
- **Dependency Resolution**: Detect cycles, resolve dependencies, and execute in correct order
- **Semantic Versioning**: Parse and validate version constraints with 8 constraint types
- **Metadata Validation**: Comprehensive validation of agent definitions and requirements
- **Tool Integration**: 4 ADK tools for agent management (run_agent, resolve_deps, validate_agent, and optional dependency_graph)

## Architecture

### Two-Layer Design

```
┌─────────────────────────────────────────────┐
│        Layer 2: ADK Tools                   │
│  (functiontool wrappers, automatic registry)│
├─────────────────────────────────────────────┤
│  run_agent.go    │  resolve_deps.go         │
│  validate_agent.go  (optional tools)        │
├─────────────────────────────────────────────┤
│        Layer 1: Pure Utilities              │
│   (Zero external deps, fully testable)      │
├─────────────────────────────────────────────┤
│  execution.go    │ dependencies.go          │
│  version.go      │ metadata_integration.go  │
├─────────────────────────────────────────────┤
│     Standard Library + YAML parsing         │
└─────────────────────────────────────────────┘
```

**Design Philosophy:**
- **Layer 1** provides pure, testable utilities with no external dependencies beyond Go stdlib
- **Layer 2** wraps Layer 1 utilities with ADK framework integration patterns
- Clear separation enables independent testing and code reusability
- Automatic registration via `init()` functions ensures tools are available at runtime

### Core Components

#### Layer 1: Pure Utilities

1. **execution.go** - Agent execution engine
   - `ExecutionContext`: Manages execution state and streams
   - `ExecutionResult`: Captures execution outcome and metrics
   - `ExecutionRequirements`: Defines agent requirements and constraints
   - `AgentRunner`: Core execution engine with streaming support

2. **dependencies.go** - Dependency graph and resolution
   - `DependencyGraph`: Directed graph of agent relationships
   - Topological sorting: O(V+E) execution order calculation
   - Cycle detection: O(V+E) DFS-based algorithm
   - Transitive dependency resolution

3. **version.go** - Semantic versioning
   - `Version`: Parsed semantic version (Major.Minor.Patch-Prerelease)
   - `Constraint`: Version constraint matching
   - 8 constraint types: `==`, `>`, `>=`, `<`, `<=`, `^` (caret), `~` (tilde), `range`
   - Prerelease and build metadata support

4. **metadata_integration.go** - Unified validation
   - `AgentMetadataValidator`: Bridges all systems
   - 13 public validation methods covering all aspects
   - `ValidationReport`: Structured validation results
   - `CompatibilityReport`: Agent compatibility analysis

#### Layer 2: ADK Tool Wrappers

1. **run_agent.go** - Agent execution tool
   - Input: `agent_name`, `parameters`, `timeout_seconds`, `stream_output`
   - Output: `execution_id`, `status`, `result`, `metrics`
   - Priority: 6 (Medium-high)
   - Category: Search/Discovery

2. **resolve_deps.go** - Dependency analysis tool
   - Input: `agent_name`, `show_transitive`, `check_versions`, `format` (list/tree/json)
   - Output: Dependencies array, transitive dependencies, version issues, summary
   - Priority: 7 (High)
   - Supports multiple output formats
   - Category: Search/Discovery

3. **validate_agent.go** - Agent validation tool
   - Input: `agent_name`, `check_dependencies`, `check_versions`, `check_execution_requirements`, `detailed`
   - Output: Validation status, issues array, warnings array, summary
   - Priority: 8 (Highest)
   - Comprehensive validation across all systems
   - Category: Search/Discovery

4. **dependency_graph.go** (optional) - Graph visualization tool
   - Input: `format` (graphviz/json/text), `max_depth` (optional)
   - Output: Graph data in requested format
   - Supports visual and data export formats

## Usage Guide

### 1. Running an Agent

Execute an agent with the `run_agent` tool:

```bash
adk-code /agent_run_agent \
  --agent_name "my-agent" \
  --parameters '{"key": "value"}' \
  --timeout_seconds 30 \
  --stream_output true
```

**Response:**
```json
{
  "execution_id": "exec-1234-5678",
  "status": "success",
  "result": {
    "output": "Agent execution result",
    "metrics": {
      "duration_ms": 2500,
      "memory_used_mb": 128
    }
  }
}
```

**Parameters:**
- `agent_name` (string, required): Name of the agent to run
- `parameters` (JSON object, optional): Input parameters for the agent
- `timeout_seconds` (integer, optional): Execution timeout (default: 300)
- `stream_output` (boolean, optional): Stream output in real-time (default: true)

**Response Fields:**
- `execution_id`: Unique identifier for this execution
- `status`: "success", "failed", "timeout", "cancelled"
- `result`: Execution output and metrics
- `error`: Error message if execution failed

### 2. Resolving Dependencies

Analyze agent dependencies with the `resolve_deps` tool:

```bash
adk-code /agent_resolve_deps \
  --agent_name "my-agent" \
  --show_transitive true \
  --check_versions true \
  --format "tree"
```

**Response (tree format):**
```
my-agent
├── base-agent (>=1.0.0)
│   ├── util-agent (^1.5.0)
│   └── logger-agent (>=1.0.0)
└── processor-agent (>=2.0.0)
```

**Response (JSON format):**
```json
{
  "dependencies": [
    {
      "name": "base-agent",
      "version": "1.2.0",
      "constraint": ">=1.0.0",
      "satisfied": true
    }
  ],
  "transitive_dependencies": [
    {
      "name": "util-agent",
      "path": "base-agent -> util-agent"
    }
  ],
  "version_issues": [],
  "summary": "All dependencies resolved: 3 direct, 2 transitive"
}
```

**Parameters:**
- `agent_name` (string, required): Agent to analyze
- `show_transitive` (boolean, optional): Include transitive dependencies (default: true)
- `check_versions` (boolean, optional): Validate version constraints (default: true)
- `format` (string, optional): Output format - "list", "tree", "json" (default: "tree")

### 3. Validating an Agent

Validate agent definitions with the `validate_agent` tool:

```bash
adk-code /agent_validate_agent \
  --agent_name "my-agent" \
  --check_dependencies true \
  --check_versions true \
  --check_execution_requirements true \
  --detailed true
```

**Response:**
```json
{
  "agent_name": "my-agent",
  "valid": true,
  "dependency_validation": {
    "name": "Dependencies",
    "valid": true,
    "details": [
      "Dependency chain: 3 agents in execution order"
    ]
  },
  "version_validation": {
    "name": "Versions",
    "valid": true,
    "details": [
      "Version: 1.0.0"
    ]
  },
  "execution_validation": {
    "compatible": true,
    "issues": [],
    "warnings": [],
    "details": [
      "Type: utility",
      "Source: local",
      "Version: 1.0.0"
    ]
  },
  "issues": [],
  "warnings": [],
  "summary": "Validation Report for \"my-agent\"\n✓ VALID: Agent passes all validation checks"
}
```

**Parameters:**
- `agent_name` (string, required): Agent to validate
- `check_dependencies` (boolean, optional): Validate dependencies (default: true if none specified)
- `check_versions` (boolean, optional): Validate versions (default: true if none specified)
- `check_execution_requirements` (boolean, optional): Validate requirements (default: true if none specified)
- `detailed` (boolean, optional): Include detailed information (default: false)

**Validation Checks:**
1. **Dependency Validation**: Detects cycles, verifies all dependencies exist, checks resolution order
2. **Version Validation**: Parses versions, validates format, checks semantic version constraints
3. **Execution Validation**: Verifies required fields (name, description), checks metadata completeness

### 4. Viewing Dependency Graph (Optional)

Visualize or export agent dependency graphs:

```bash
adk-code /agent_dependency_graph \
  --format "graphviz" \
  --max_depth 3
```

**Output (graphviz format):**
```
digraph {
  "my-agent" -> "base-agent"
  "base-agent" -> "util-agent"
  "base-agent" -> "logger-agent"
}
```

## Version Constraints

The system supports semantic versioning with 8 constraint types:

### Exact Match
```
"version_constraint": "==1.2.3"
```
Matches exactly version 1.2.3.

### Comparison Operators
```
">1.0.0"      # Greater than 1.0.0
">=1.0.0"     # Greater than or equal to 1.0.0
"<2.0.0"      # Less than 2.0.0
"<=2.0.0"     # Less than or equal to 2.0.0
```

### Caret (^) - Compatible Versions
```
"^1.2.3"      # >=1.2.3, <2.0.0 (allows minor/patch changes)
```
Allows changes that don't modify the left-most non-zero digit.

### Tilde (~) - Patch Version Changes
```
"~1.2.3"      # >=1.2.3, <1.3.0 (allows patch changes only)
```
Allows patch-level changes only.

### Range
```
"1.2.0-2.0.0" # >=1.2.0, <=2.0.0
```
Inclusive range between two versions.

## Dependency Resolution

### Topological Sorting

Agents are executed in dependency order using topological sort:

```
Input: Agent "web-api" depends on ["database", "cache"]
       Agent "database" depends on ["config"]
       Agent "cache" depends on ["config"]
       Agent "config" depends on []

Output (execution order):
1. config    (no dependencies)
2. database  (depends on config only)
3. cache     (depends on config only)
4. web-api   (depends on database and cache)
```

**Algorithm:**
- Time Complexity: O(V + E) where V = agents, E = dependencies
- Uses DFS-based topological sort
- Post-order traversal ensures correct execution order

### Cycle Detection

Circular dependencies are detected and reported:

```
Graph:
  agent-a -> agent-b
  agent-b -> agent-c
  agent-c -> agent-a  ❌ Creates cycle

Error:
  "circular dependency detected: agent-a -> agent-b -> agent-c -> agent-a"
```

## Execution Flow

### Complete Execution Workflow

```
User Request (run_agent)
    ↓
Parse Input & Validate
    ↓
Discover All Agents (filesystem + metadata)
    ↓
Build Dependency Graph
    ↓
Detect Cycles
    ↓
Resolve Dependencies (topological sort)
    ↓
For Each Agent (in execution order):
    ├── Create ExecutionContext
    ├── Validate Execution Requirements
    ├── Execute Agent (with streaming)
    ├── Capture ExecutionResult
    └── Update Overall Result
    ↓
Return Combined Results
```

### Error Handling

The system provides clear error messages for common issues:

**Missing Agent:**
```json
{
  "error": "agent 'non-existent' not found"
}
```

**Circular Dependency:**
```json
{
  "error": "circular dependency detected: agent-a -> agent-b -> agent-a"
}
```

**Version Mismatch:**
```json
{
  "error": "version constraint '^1.0.0' not satisfied by version '0.5.0'"
}
```

**Missing Dependency:**
```json
{
  "error": "dependency 'required-agent' not found in discovered agents"
}
```

## Metadata Integration

### Agent Definition Example

```yaml
# agent.yaml
name: "data-processor"
description: "Processes and transforms data"
version: "1.2.0"
type: "processor"
source: "local"
author: "Engineering Team"
tags:
  - "data"
  - "etl"
  - "processing"
dependencies:
  - "database-connector"
  - "validation-engine"
version_constraints:
  - "database-connector": ">=1.0.0"
  - "validation-engine": "^2.0.0"
execution_requirements:
  timeout_seconds: 60
  memory_mb: 512
  requires_network: true
```

### Validation Rules

1. **Name & Description Required**: All agents must have name and description
2. **Version Format**: If specified, must follow semantic versioning (Major.Minor.Patch)
3. **Dependencies Exist**: All listed dependencies must be discoverable
4. **No Cycles**: Dependency graph must be acyclic
5. **Version Constraints Satisfied**: All version constraints must be satisfied by discovered versions
6. **Execution Requirements Valid**: Timeout and resource limits must be positive integers

## Troubleshooting

### Agent Not Found

**Problem:**
```
error: agent 'my-agent' not found
```

**Solutions:**
1. Verify agent name spelling and case sensitivity
2. Check agent file location is within discovery paths
3. Ensure agent metadata is valid YAML/JSON
4. Run discovery to see all available agents

### Circular Dependency

**Problem:**
```
error: circular dependency detected
```

**Solutions:**
1. Review agent dependency declarations
2. Remove unnecessary dependencies
3. Restructure dependencies to be acyclic
4. Use `resolve_deps` tool to visualize dependency chain

### Version Constraint Mismatch

**Problem:**
```
error: version constraint '^1.0.0' not satisfied by version '0.9.5'
```

**Solutions:**
1. Update dependency agent to required version
2. Relax version constraint (e.g., `>=0.9.0` instead of `^1.0.0`)
3. Check available versions with `resolve_deps --check_versions true`

### Execution Timeout

**Problem:**
```
error: execution timeout after 30 seconds
```

**Solutions:**
1. Increase timeout: `--timeout_seconds 60`
2. Optimize agent implementation
3. Check for long-running operations or network calls
4. Validate resource availability (CPU, memory, network)

## Advanced Usage

### Streaming Large Outputs

For agents producing large outputs, use streaming:

```bash
adk-code /agent_run_agent \
  --agent_name "data-processor" \
  --stream_output true
```

Output is streamed line-by-line, reducing memory usage.

### Validating Before Execution

Always validate agents before running in production:

```bash
# Comprehensive validation
adk-code /agent_validate_agent \
  --agent_name "my-agent" \
  --detailed true

# If validation passes, execute
if [ $? -eq 0 ]; then
  adk-code /agent_run_agent --agent_name "my-agent"
fi
```

### Analyzing Complex Dependencies

For multi-level dependencies, use detailed analysis:

```bash
# Show all transitive dependencies in tree format
adk-code /agent_resolve_deps \
  --agent_name "root-agent" \
  --show_transitive true \
  --format "tree"

# Export as JSON for programmatic use
adk-code /agent_resolve_deps \
  --agent_name "root-agent" \
  --format "json" > dependencies.json
```

### Batch Validation

Validate multiple agents programmatically:

```bash
for agent in agent1 agent2 agent3; do
  echo "Validating $agent..."
  adk-code /agent_validate_agent --agent_name "$agent" || exit 1
done
echo "All agents valid!"
```

## Performance Considerations

### Execution Optimization

1. **Parallel Execution**: Agents with no inter-dependencies can be executed in parallel
2. **Caching**: Validation results are cached to avoid re-computation
3. **Streaming**: Large outputs are streamed to reduce memory overhead
4. **Lazy Loading**: Agents are loaded only when needed

### Resource Requirements

**Per-Agent Execution:**
- Default timeout: 300 seconds
- Default memory: 256 MB
- Network: Optional (specify in metadata)

**Graph Operations:**
- Discovery: Linear in number of agent files
- Cycle detection: O(V + E)
- Dependency resolution: O(V + E)
- Version validation: O(n * m) where n = versions, m = constraints

## Best Practices

1. **Define Clear Dependencies**: Keep dependency graphs simple and well-defined
2. **Use Semantic Versioning**: Follow SemVer for all agent versions
3. **Validate Early**: Run `validate_agent` before deploying agents
4. **Document Requirements**: Specify execution requirements in metadata
5. **Test Dependencies**: Verify dependency resolution with `resolve_deps` tool
6. **Monitor Execution**: Use streaming output to track long-running agents
7. **Handle Errors Gracefully**: Check return status and error messages

## API Reference

### ExecutionContext

```go
type ExecutionContext struct {
    AgentName        string
    Parameters       map[string]interface{}
    TimeoutSeconds   int
    StreamOutput     bool
    OutputChan       chan string  // For streaming
    DoneChan         chan struct{} // Cancellation
}
```

### ExecutionResult

```go
type ExecutionResult struct {
    ExecutionID   string
    AgentName     string
    Status        string      // "success", "failed", "timeout"
    Output        string
    Error         string
    StartTime     time.Time
    EndTime       time.Time
    MetricsData   map[string]interface{}
}
```

### DependencyGraph

```go
type DependencyGraph struct {
    Agents map[string]*Agent
    Edges  map[string][]string
}

// Key Methods:
func (dg *DependencyGraph) AddAgent(agent *Agent) error
func (dg *DependencyGraph) AddEdge(from, to string) error
func (dg *DependencyGraph) ResolveDependencies(agentName string) ([]*Agent, error)
func (dg *DependencyGraph) GetTransitiveDependencies(agentName string) ([]string, error)
```

### Version

```go
type Version struct {
    Major      int
    Minor      int
    Patch      int
    Prerelease string
}

// Key Methods:
func ParseVersion(versionStr string) (*Version, error)
func (v *Version) Matches(constraint *Constraint) bool
func (v *Version) String() string
```

## Testing

All components include comprehensive test coverage:

- **execution_test.go**: 20 tests (100% coverage)
- **dependencies_test.go**: 20 tests (complete graph operations)
- **version_test.go**: 14 tests (all constraint types)
- **metadata_integration_test.go**: 26 tests (complete validation)
- **run_agent_test.go**: 12 tool-specific tests
- **validate_agent_test.go**: 22 validation tests
- **resolve_deps_test.go**: Tool integration tests
- **End-to-end tests**: Complete workflow validation

**Run Tests:**
```bash
make test                    # All tests
make test-verbose            # With details
make coverage                # Coverage report
```

## Integration with ADK Framework

The system integrates seamlessly with ADK's agent lifecycle:

1. **Tool Registration**: Tools automatically register via `init()` functions
2. **Input/Output Contracts**: Strict type definitions for tool inputs/outputs
3. **Error Propagation**: Clear error messages for debugging
4. **Streaming Support**: Output streaming for large results
5. **Context Management**: Proper context handling for cancellation

## Future Enhancements

1. **Parallel Execution**: Execute independent agents concurrently
2. **Agent Templating**: Support agent definition templates and inheritance
3. **Version Auto-Update**: Automatic dependency version management
4. **Execution History**: Track and replay previous executions
5. **Performance Metrics**: Detailed execution performance analysis
6. **Remote Agent Support**: Execute agents from remote sources
7. **Agent Signing**: Cryptographic signing and verification of agents
8. **Dependency Caching**: Cache dependency graphs for faster resolution

## Contributing

When extending the Agent Execution System:

1. **Follow Two-Layer Design**: Keep utilities pure, wrappers light
2. **Write Tests First**: Maintain >80% coverage
3. **Document Changes**: Update this guide and code comments
4. **Run Quality Checks**: `make check` before committing
5. **Test Integration**: Verify ADK framework integration

## Support

For issues or questions:

1. Check **Troubleshooting** section above
2. Run `adk-code /help agent` for tool-specific help
3. Review **API Reference** for detailed signatures
4. Check test files for usage examples
5. Review logs for detailed error information

---

**Last Updated:** Phase 2 Week 4  
**Version:** 1.0.0  
**Maintenance:** Active
