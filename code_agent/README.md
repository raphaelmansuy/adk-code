# Code Agent 🤖

An AI-powered coding assistant CLI built with Google ADK Go, similar to Claude Code or Gemini Code CLI.

## Features

- **File Operations**: Read, write, and modify files with intelligent code editing
- **Terminal Execution**: Run commands, tests, and build tools
- **Code Search**: Search for patterns and files across your codebase
- **Iterative Problem Solving**: The agent works through problems step-by-step until completion
- **Interactive CLI**: Beautiful command-line interface with color-coded output

## Tools Available

The agent has access to the following tools:

- **read_file**: Read file contents to understand code
- **write_file**: Create new files or overwrite existing ones
- **replace_in_file**: Make precise edits by replacing text
- **list_directory**: Explore project structure
- **search_files**: Find files by pattern (e.g., *.go, test_*.py)
- **execute_command**: Run shell commands (tests, builds, installations)
- **grep_search**: Search for text patterns in files

## Prerequisites

- Go 1.24 or later
- Google API Key with Gemini access

## Installation

1. Clone this repository:
```bash
cd code_agent
```

2. Set up your Google API key:
```bash
export GOOGLE_API_KEY="your-api-key-here"
```

3. Install dependencies:
```bash
go mod tidy
```

4. Build the application:
```bash
go build -o code-agent
```

## Usage

Run the coding agent:

```bash
./code-agent
```

Or run directly with go:

```bash
go run main.go
```

### Example Interactions

**Example 1: Create a new Go function**
```
You: Create a function to calculate fibonacci numbers in a new file called fibonacci.go

Agent: [Reads project structure, creates file with implementation, runs tests]
```

**Example 2: Fix a bug**
```
You: Find and fix the off-by-one error in sort.go

Agent: [Reads the file, identifies the issue, makes the fix, runs tests to verify]
```

**Example 3: Add tests**
```
You: Add unit tests for the user authentication module

Agent: [Examines existing code, creates comprehensive tests, runs them]
```

## Architecture

```
code_agent/
├── main.go                 # CLI entry point with interactive loop
├── agent/
│   └── coding_agent.go     # Agent configuration and system prompt
├── tools/
│   ├── file_tools.go       # File operation tools
│   └── terminal_tools.go   # Terminal execution tools
└── go.mod                  # Go module definition
```

## How It Works

1. **User Input**: You describe what you want to accomplish
2. **Planning**: The agent analyzes the request and plans its approach
3. **Tool Execution**: The agent uses tools to read files, write code, run commands
4. **Iteration**: If something doesn't work, the agent analyzes errors and tries again
5. **Verification**: The agent tests its changes to ensure they work correctly

## System Prompt

The agent is guided by a comprehensive system prompt that instructs it to:

- Understand the codebase before making changes
- Make targeted, precise edits
- Test changes frequently
- Handle errors gracefully
- Work autonomously until tasks are complete

## Built With

- [Google ADK Go](https://github.com/google/adk-go) - Agent Development Kit
- [Gemini](https://ai.google.dev/) - Google's generative AI model
- Go standard library for file and terminal operations

## Comparison to Other Tools

| Feature | Code Agent | Claude Code | Gemini Code CLI |
|---------|-----------|-------------|-----------------|
| File Operations | ✅ | ✅ | ✅ |
| Command Execution | ✅ | ✅ | ✅ |
| Iterative Solving | ✅ | ✅ | ✅ |
| Open Source | ✅ | ❌ | ❌ |
| Customizable | ✅ | ❌ | ❌ |

## Contributing

Contributions are welcome! This is a demonstration project showing how to build sophisticated coding agents with Google ADK Go.

## License

Apache 2.0 License

## Acknowledgments

- Built with Google ADK Go
- Inspired by Claude Code and Gemini Code CLI
- Powered by Gemini 2.0 Flash
