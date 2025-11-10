# Code Agent - User Guide

## Quick Start

```bash
export GOOGLE_API_KEY="your-api-key"
./code-agent
```

## Available Commands

### Natural Language Requests

Simply type what you want in plain English:

```
❯ Add comments to main.py
❯ Create a README.md with project overview
❯ Refactor the calculate function
❯ Run tests and fix any failures
❯ Add error handling to the server code
```

### Built-in Commands

| Command | Aliases | Description |
|---------|---------|-------------|
| `help` | `.help` | Show help message with examples |
| `exit` | `quit` | Exit the agent |
| `.prompt` | `debug prompt`, `show prompt` | Display the full system prompt |
| `.tools` | - | List all available tools |

## Command Examples

### Get Help

```
❯ help
```

Shows:
- Available commands
- Tool categories
- Usage examples

### View System Prompt

```
❯ .prompt
```

Displays the complete system prompt that guides the agent's behavior. Useful for:
- Understanding agent capabilities
- Debugging agent behavior
- Verifying prompt customizations

### List Available Tools

```
❯ .tools
```

Shows all tools organized by category:
- **Core Editing:** read_file, write_file, search_replace, edit_lines, apply_patch, apply_v4a_patch
- **Discovery:** list_files, search_files, grep_search
- **Execution:** execute_command, execute_program

## Command-Line Options

```bash
./code-agent [options]
```

| Option | Default | Description |
|--------|---------|-------------|
| `--output-format` | `rich` | Output format: `rich`, `plain`, or `json` |
| `--typewriter` | `false` | Enable typewriter effect for text output |

### Examples

**Plain text output (no colors):**
```bash
./code-agent --output-format=plain
```

**Enable typewriter effect:**
```bash
./code-agent --typewriter
```

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `GOOGLE_API_KEY` | Yes | Google Gemini API key |
| `GEMINI_API_KEY` | No | Alternative name for API key |

**Note:** If both are set, `GOOGLE_API_KEY` takes precedence.

## Working with Files

### The Agent Can:

- Read files (with line ranges for large files)
- Create new files
- Modify existing files using multiple methods:
  - **search_replace** - Targeted changes (recommended)
  - **edit_lines** - Line-based edits
  - **apply_patch** - Unified diff patches
  - **apply_v4a_patch** - Semantic patches (new!)
  - **write_file** - Complete rewrites

### Path Resolution

All file paths are **relative to the working directory** (where you run code-agent).

Examples:
- `main.py` - File in current directory
- `src/handler.go` - File in src/ subdirectory
- `./demo/test.c` - Explicit relative path

## Agent Capabilities

### Core Editing Tools

**read_file** - Read and examine code
- Supports line ranges for large files
- Use: Understanding code, checking contents

**write_file** - Create or overwrite files
- Atomic writes for safety
- Size validation prevents data loss
- Use: New files, complete rewrites

**search_replace** - Make targeted changes (RECOMMENDED)
- Uses SEARCH/REPLACE blocks
- Whitespace-tolerant matching
- Multiple blocks in one call
- Use: Bug fixes, small edits, refactoring

**edit_lines** - Edit by line number
- Insert/replace/delete specific lines
- Use: Structural changes, fixing braces, adding imports

**apply_patch** - Apply unified diff patches
- Standard RFC 3881 format
- Dry run mode for previews
- Use: Large refactoring, complex changes

**apply_v4a_patch** - Apply V4A semantic patches (NEW!)
- Uses class/function names instead of line numbers
- More resilient to code changes
- Better readability
- Use: Refactoring within classes/functions

### Discovery Tools

**list_files** - Explore project structure  
**search_files** - Find files by pattern (*.go, test_*.py)  
**grep_search** - Search for text in files

### Execution Tools

**execute_command** - Run shell commands
- Supports pipes and redirects
- Use: `ls -la | grep test`, `make build`

**execute_program** - Run programs directly
- No shell quoting issues
- Arguments passed directly
- Use: `./calculate "5 + 3"`, `gcc -o output input.c`

## Tips & Best Practices

### For Best Results

1. **Be specific** - Provide clear, detailed instructions
2. **Use natural language** - No special syntax needed
3. **Let the agent explore** - It can list directories and read files
4. **Review changes** - Agent uses dry_run/preview modes

### Common Patterns

**Adding a feature:**
```
❯ Add logging to the server module with timestamps and log levels
```

**Fixing issues:**
```
❯ Fix the null pointer error in process_request function
```

**Refactoring:**
```
❯ Refactor calculate_total to use a helper function for validation
```

**Testing:**
```
❯ Run all tests and fix any failures you find
```

**Documentation:**
```
❯ Add docstrings to all functions in utils.py
```

## Troubleshooting

### Agent can't find files

Check your working directory:
```
❯ list current directory
```

Agent automatically explores the project structure.

### Unclear instructions

Use the help command for examples:
```
❯ help
```

### Want to see how agent works

View the system prompt:
```
❯ .prompt
```

### Need to know available tools

List all tools:
```
❯ .tools
```

## Advanced Features

### V4A Patch Format (NEW!)

V4A patches use semantic context (class/function names) instead of line numbers:

```
@@ class User
@@     def validate():
-          return True
+          if not self.email:
+              raise ValueError("Email required")
+          return True
```

**Benefits:**
- More resilient to code changes
- Better readability
- Easier to understand intent

**When to use:**
- Refactoring within classes/functions
- File is frequently modified
- Semantic clarity matters

### Batch Operations

The agent can handle multiple related changes:
```
❯ Add error handling to all functions in handler.go and update tests
```

### Interactive Workflows

The agent iterates automatically:
```
❯ Create a calculator program, test it, and fix any issues
```

It will:
1. Create the program
2. Run tests
3. Debug and fix errors
4. Verify fixes work

## Examples

### Example 1: Add Comments

```
❯ Add comprehensive comments to demo/prolog/main.py
```

Agent will:
- Read the file
- Understand the code structure
- Add appropriate comments
- Write back the file

### Example 2: Create Documentation

```
❯ Create a README.md with project overview, installation, and usage instructions
```

Agent will:
- Explore project structure
- Read relevant files
- Generate comprehensive README
- Include code examples

### Example 3: Refactoring

```
❯ Refactor the calculate function to separate validation and computation
```

Agent will:
- Read the function
- Understand logic
- Split into two functions
- Update all references
- Verify changes work

### Example 4: Fix Compilation Errors

```
❯ Compile the project and fix any errors
```

Agent will:
- Run build command
- Read error messages
- Identify issues
- Make fixes
- Rebuild to verify

## Getting More Help

1. **In-app help:** Type `help` in the agent
2. **System prompt:** Type `.prompt` to see full capabilities
3. **Tool list:** Type `.tools` to see all available tools
4. **Project docs:** Check `/doc` directory for detailed documentation

## Version Information

Current version: 1.0.0  
Model: gemini-2.5-flash

For latest updates and documentation, see the project repository.
