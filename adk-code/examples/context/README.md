# Context Management Example

This example demonstrates the key features of the context management system implemented according to ADR-0006.

## Running the Example

```bash
cd examples/context
go run main.go
```

## What It Demonstrates

### 1. Context Manager

Shows how to:
- Create a ContextManager for a specific model
- Add conversation items
- Track token usage
- Detect when compaction is needed

### 2. Token Tracking

Demonstrates:
- Recording turns with input/output token counts
- Calculating statistics (average turn size, total tokens)
- Estimating remaining turns before context limit

### 3. Output Truncation

Illustrates:
- Automatic truncation of large tool outputs
- Head+tail strategy (preserves beginning and end)
- Elision markers showing what was omitted
- Size limits (10 KiB default)

### 4. Instruction Hierarchy

Shows:
- Loading instructions from multiple levels
- Global, project, and directory-specific AGENTS.md files
- Automatic merging with size limits

## Creating AGENTS.md Files

To see instruction loading in action, create AGENTS.md files:

```bash
# Global instructions (applies to all projects)
mkdir -p ~/.adk-code
echo "# Global Instructions\n\nAlways be helpful and precise." > ~/.adk-code/AGENTS.md

# Project instructions (applies to this project)
echo "# Project Instructions\n\nFollow Go best practices." > ../../AGENTS.md

# Directory instructions (applies to this directory)
echo "# Context Example Instructions\n\nKeep examples simple and clear." > AGENTS.md
```

Then run the example again to see the instructions loaded.

## Expected Output

The example will show:
- Token usage percentages
- Turn-by-turn statistics
- Truncation results with markers
- Instruction loading status

## Learn More

- [Context Management Guide](../../../docs/CONTEXT_MANAGEMENT.md)
- [ADR-0006](../../../docs/adr/0006-agent-context-management.md)
