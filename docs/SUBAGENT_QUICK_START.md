# Subagent Quick Start Guide

## Overview

adk-code now supports **subagents** - specialized AI agents that handle specific tasks. The main agent automatically delegates to these specialists when appropriate.

## How It Works

**Agent as Tool Pattern:**
- Subagents are discovered from `.adk/agents/*.md` files
- Each subagent is converted to a tool using ADK's `agenttool.New()`
- The main agent decides when to use subagents naturally
- No manual routing needed - the LLM handles delegation

## Available Subagents (Built-in)

### 1. code-reviewer
**Purpose:** Code quality, security, and best practices review  
**Use when:** After writing or modifying code

```bash
# Example usage:
> Please review the authentication code in src/auth.go
# Main agent will automatically delegate to code-reviewer
```

### 2. debugger
**Purpose:** Finding and fixing bugs  
**Use when:** Code isn't working as expected

```bash
> The login function is crashing, can you help debug it?
```

### 3. test-engineer
**Purpose:** Writing and running tests  
**Use when:** Need test coverage or test failures

```bash
> Write unit tests for the UserService class
```

### 4. architect
**Purpose:** System design and architecture analysis  
**Use when:** Planning or analyzing system structure

```bash
> Analyze the architecture of this microservices system
```

### 5. documentation-writer
**Purpose:** Technical documentation and explanations  
**Use when:** Need documentation or code explanations

```bash
> Create API documentation for the REST endpoints
```

## Creating Custom Subagents

### 1. Create Agent Definition File

Create a file in `.adk/agents/my-agent.md`:

```markdown
---
name: my-agent
description: Brief description of what this agent does
tags: [tag1, tag2]
version: 1.0.0
author: your-email@example.com
---

# My Custom Agent

## Role and Purpose
Describe the agent's role and when it should be used.

## Instructions
Provide detailed instructions for how the agent should operate.

## Example Tasks
- Task 1
- Task 2
```

### 2. Required Fields

- **name**: Unique agent identifier (lowercase, hyphens allowed)
- **description**: One-line description (used for delegation decisions)

### 3. Optional Fields

- **tags**: Categories for filtering
- **version**: Semantic version (e.g., "1.0.0")
- **author**: Your email or name
- **dependencies**: Other agents this depends on

## File Format

```markdown
---
name: agent-name
description: What this agent does
---

# Agent Title

Your detailed system prompt here.
The main agent will use this as the system instruction.
```

## How Delegation Works

1. **Automatic**: Main agent analyzes your request
2. **Decision**: LLM decides if a specialist is needed
3. **Delegation**: Subagent executes with isolated context
4. **Synthesis**: Results integrated back to conversation

## Examples

### Review Code
```bash
> Review security in payment.go
✓ Delegating to code-reviewer...
[code-reviewer analyzes and provides feedback]
```

### Debug Issues
```bash
> Why is the API returning 500 errors?
✓ Delegating to debugger...
[debugger investigates and suggests fixes]
```

### Composite Tasks
```bash
> Review the code, then write tests for it
✓ Delegating to code-reviewer...
[review complete]
✓ Delegating to test-engineer...
[tests written]
```

## REPL Commands

```bash
/agents              # List all available subagents
/run-agent <name>    # Preview agent details
```

## Best Practices

1. **Keep descriptions concise** - They're used for delegation decisions
2. **Be specific in system prompts** - Clear instructions = better results
3. **Use tags for organization** - Makes filtering easier
4. **Version your agents** - Track changes over time
5. **Test agent behavior** - Try different requests to verify delegation

## Troubleshooting

### Subagent Not Loading?

Check:
- File is in `.adk/agents/` directory
- File has `.md` extension
- YAML frontmatter is valid
- Required fields (name, description) are present

### Agent Not Being Used?

- Make your request more specific
- Explicitly mention the agent: "use the code-reviewer"
- Check agent description matches your use case

## Technical Details

### Discovery Process
1. Scans `.adk/agents/` (project-level)
2. Scans `~/.adk/agents/` (user-level)
3. Parses YAML frontmatter
4. Validates required fields
5. Creates llmagent instances
6. Wraps as tools via `agenttool.New()`

### Phase 1 Limitations
- Subagents have no tools (analysis/recommendations only)
- No explicit tool restrictions yet
- No subagent chaining yet

### Future Enhancements (Phase 2+)
- Tool restrictions per agent
- Subagent chaining support
- Interactive creation via `/agents create`
- Resumable subagent sessions

## More Information

- **Architecture**: See `docs/spec/claude_code_like_agent_feature/02_adk_code_implementation_approach.md`
- **Implementation**: See `tools/agents/subagent_tools.go`
- **Examples**: See `.adk/agents/` for built-in agents

## Feedback

Found an issue or have suggestions? Please open an issue on GitHub.
