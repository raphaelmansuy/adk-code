# adk-code

> **An intelligent CLI coding assistant powered by Google's ADK framework**

[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square)](LICENSE)
[![GitHub Stars](https://img.shields.io/github/stars/raphaelmansuy/adk-code?style=flat-square)](https://github.com/raphaelmansuy/adk-code)


## What is adk-code?

`adk-code` is a **multi-model AI coding assistant** that runs directly in your terminal. Ask natural language questions about your code—it reads files, executes commands, makes edits, and runs searches autonomously.

### Key Features

- **🤖 Multi-Model Support**: Seamlessly switch between Gemini, OpenAI, and Vertex AI
- **🛠️ 21 Built-in Tools**: File operations, code editing, execution, web search, and more
- **🔌 MCP Integration**: Unlimited extensibility via Model Context Protocol
- **💾 Session Persistence**: Maintain context across conversations with automatic history
- **🧠 Smart Context Management**: Automatic token tracking, output truncation, and conversation compaction for 50+ turn workflows
- **⚡ Streaming Responses**: Real-time output as the model thinks and executes
- **🎨 Beautiful Terminal UI**: Rich formatting, colors, and interactive displays
- **📦 Zero External Dependencies**: No Langchain, Claude Code, or Cline baggage

## Quick Start

### Installation

#### Option A: Homebrew (macOS) — Recommended

The easiest way to install on macOS:

```bash
# Add the tap (one-time)
brew tap raphaelmansuy/adk-code

# Install adk-code
brew install adk-code

# Verify installation
adk-code --version
```

**Supported on:**

- macOS 10.13+ (High Sierra and later)
- Intel (x86_64) and Apple Silicon (M-series) Macs

**Update to latest:**

```bash
brew upgrade adk-code
```

**Uninstall:**

```bash
brew uninstall adk-code
```

See [homebrew-adk-code](https://github.com/raphaelmansuy/homebrew-adk-code) for more details.

#### Option B: Build from Source

Clone and build manually:

```bash
# Clone and build
git clone https://github.com/raphaelmansuy/adk-code.git
cd adk-code/adk-code
make build

# Binary is now at ../bin/adk-code
```

### 1-Minute Setup

```bash
# Set your API key
export GOOGLE_API_KEY=your-key-here

# Run adk-code
../bin/adk-code
```

That's it! You're ready to ask questions about your code.

### Examples

```bash
# Interactive mode (default)
❯ How do I add error handling to ReadFile?
[adk-code reads files, analyzes, and suggests changes]

❯ Create a CLI parser for flags
[adk-code implements, tests, and explains]

# Session mode
❯ adk-code --session my-project --model gpt-4o

# Batch mode
❯ echo "Write a test for userAuth()" | adk-code
```

## 🎯 Use Cases

| Use Case | Benefit |
|----------|---------|
| **Code Review** | Understand complex codebases quickly |
| **Bug Fixes** | Trace errors and implement solutions |
| **Refactoring** | Improve code quality with AI guidance |
| **Documentation** | Generate docs and comments |
| **Testing** | Write and run test suites |
| **Learning** | Study patterns and best practices |

## 🏗️ Architecture at a Glance

```
┌─────────────────────────────────┐
│     User Terminal (REPL)        │
└────────────┬────────────────────┘
             │
┌────────────▼────────────────────┐
│   Agent Loop (ADK Framework)    │
│  ┌──────────────────────────┐   │
│  │ 1. Call LLM with context │   │
│  │ 2. Parse tool calls      │   │
│  │ 3. Execute tools         │   │
│  │ 4. Append results        │   │
│  │ 5. Loop until complete   │   │
│  └──────────────────────────┘   │
└────────────┬────────────────────┘
             │
    ┌────────┴────────┬──────────┐
    ▼                 ▼          ▼
┌────────────┐  ┌─────────┐  ┌──────────┐
│ 21 Tools   │  │LLM APIs │  │ Display  │
├────────────┤  ├─────────┤  ├──────────┤
│ File Ops   │  │ Gemini  │  │ Rich UI  │
│ Execution  │  │ OpenAI  │  │ Colors   │
│ Search     │  │ Vertex  │  │ Markdown │
└────────────┘  └─────────┘  └──────────┘
```

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for details.

## 🔀 Dynamic Sub-Agent Architecture

adk-code supports a dynamic sub-agent system that enables intent-driven delegation and modular, reusable agent definitions. This architecture is designed to make complex tasks more manageable by letting dedicated sub-agents handle specialized responsibilities while the main agent orchestrates work.

- Discovery: `pkg/agents` discovers agent definitions from `.adk/agents/` (YAML frontmatter + Markdown). Agents are validated for name, description, and metadata (version, author, tags).
- Agent Router (planned): We define the router in the spec (`internal/agents/router.go`) as a small decision layer that will run intent scoring and select the right handler. Note: the router is a planned Phase 1 component.

  Current behavior: adk-code already supports actionable subagents using ADK's agent-as-tool pattern — see `tools/agents/subagent_tools.go` and `internal/prompts/coding_agent.go`. `SubAgentManager` discovers `.adk/agents/*.md`, creates `llmagent` instances, and registers them as tools; the LLM naturally selects an agent/tool at runtime. This provides pragmatic delegation today while the router design remains part of Phase 1.
- Sub-Agents: Sub-agents are self-contained agent definition files with metadata and behavior (skills/commands). They can be added, versioned, and discovered at runtime.
- Delegation Flow (current): User request → LLM (main agent) selects a tool/subagent → If a sub-agent tool is invoked, it runs in its own context and executes allowed tools or MCP services → Return result → Main agent synthesizes final answer.
-
- Delegation Flow (future router): User request → Agent Router (intent scoring, heuristic) → Select sub-agent → Sub-agent executes tools or MCP services → Return result → Main agent synthesizes final answer.
- Tools & MCP: Sub-agents call local tools or external MCP servers for actions (filesystem edits, Git, build, cloud APIs). This separation keeps tool execution deterministic and traceable.
- Audit & Replay: All agent actions (intent scores, chosen sub-agent, tool calls, and MCP interactions) are logged to the session history. This enables replays, debugging, and reproducibility.

Benefits: concise intent routing, modular agent definitions, scalable delegation to domain-specific sub-agents, and transparent tool/MCP integration.

## 📚 Documentation

- **[QUICK_REFERENCE.md](docs/QUICK_REFERENCE.md)** — Daily commands & flags (2 min)
- **[ARCHITECTURE.md](docs/ARCHITECTURE.md)** — System design & components (15 min)
- **[TOOL_DEVELOPMENT.md](docs/TOOL_DEVELOPMENT.md)** — Build your own tools (20 min)
- **[docs/](docs/)** — Complete documentation suite

## 💻 Requirements

- **Go 1.24+**
- One API key:
  - `GOOGLE_API_KEY` (Gemini - free tier available)
  - `OPENAI_API_KEY` (OpenAI)
  - GCP project (Vertex AI)

## 🚀 Getting Started

### Option 1: Gemini (Recommended)

Free tier, fastest setup:

```bash
export GOOGLE_API_KEY=your-key
cd adk-code && make run
```

### Option 2: OpenAI

```bash
export OPENAI_API_KEY=sk-...
cd adk-code && make run -- --model gpt-4o
```

### Option 3: Vertex AI (GCP)

```bash
export GOOGLE_CLOUD_PROJECT=your-project
export GOOGLE_CLOUD_LOCATION=us-central1
export GOOGLE_GENAI_USE_VERTEXAI=true
cd adk-code && make run
```

## 🛠️ Development

```bash
cd adk-code

# Build
make build

# Test
make test

# Quality checks (required before commit)
make check

# Development watch mode
make watch
```

## 🔧 CLI Flags

```bash
./adk-code --model gemini-2.5-flash           # Specify model
./adk-code --session my-project               # Named session
./adk-code --output-format plain              # Output format
./adk-code --enable-thinking                  # Extended reasoning
./adk-code --working-directory /path/to/src   # Set working dir
```

See [QUICK_REFERENCE.md](docs/QUICK_REFERENCE.md) for all flags.

## 🧠 How It Works

1. **You ask a question** in natural language
2. **Agent receives context** (system prompt, tools, history)
3. **LLM generates response** with tool calls (read file, run command, etc.)
4. **Tools execute** and return results
5. **Agent loops** until response is complete
6. **Result streams** to your terminal in real-time

Example: "How many lines in main.go?"

```
Agent thinks: "User wants line count. I'll use count_lines tool."
  ↓
Calls: count_lines(path="main.go")
  ↓
Gets: {success: true, total_lines: 140, ...}
  ↓
Returns: "main.go has 140 lines"
```

## 🌐 Extensibility

### Add Tools

Create your own tools without modifying core code. See [TOOL_DEVELOPMENT.md](docs/TOOL_DEVELOPMENT.md).

```go
// 4-step pattern
type MyToolInput struct { Path string }
type MyToolOutput struct { Result string }

func handler(ctx Context, input MyToolInput) MyToolOutput {
    // Your logic
    return MyToolOutput{Result: "..."}
}

func init() {
    // Register automatically
}
```

### Connect External Tools (MCP)

Use Model Context Protocol servers instead of building tools:

```json
{
  "mcp": {
    "servers": {
      "github": {
        "type": "stdio",
        "command": "mcp-server-github"
      }
    }
  }
}
```

## 🧠 Context Management

`adk-code` includes intelligent context window management to support long-running conversations (50+ turns):

- **Automatic Token Tracking**: Real-time visibility into token usage after each turn
- **Smart Output Truncation**: Head+tail strategy preserves critical information (first/last 128 lines)
- **Conversation Compaction**: LLM-powered summarization at 70% threshold (configurable)
- **Sub-Agent Support**: Context shared across main agent and all sub-agents
- **No Silent Data Loss**: Clear markers show when content is truncated

```bash
# Token usage displayed after each turn
User: Create a file test.py
📊 Context: 1,250/1,000,000 tokens (0.1%) • Compaction at 70%

# Warning when approaching limit
User: Run all tests
⚠️  Context approaching limit - compaction recommended
📊 Context: 710,000/1,000,000 tokens (71.0%) • Compaction at 70%
```

**Configure the threshold** programmatically:

```go
// Use custom compaction threshold (e.g., 80%)
cm := context.NewContextManagerWithOptions(modelConfig, llm, 0.80)

// Or update dynamically
cm.SetCompactThreshold(0.60) // Compact at 60%
```

See [CONTEXT_MANAGEMENT.md](docs/CONTEXT_MANAGEMENT.md) and [CONTEXT_INTEGRATION.md](docs/CONTEXT_INTEGRATION.md) for details.

## 📊 Performance

| Metric | Value |
|--------|-------|
| **Binary Size** | ~15MB (release) |
| **Startup Time** | <500ms |
| **Context Window** | Up to 2M tokens (Gemini 1.5 Pro) |
| **Max Conversation** | 50+ turns with auto-compaction |
| **Tool Execution** | <1s typical |
| **Memory Usage** | ~50MB baseline |

## 🤝 Contributing

Contributions welcome! Please:

1. Fork and create a branch
2. Make changes following Go conventions
3. Run `make check` before committing
4. Submit a pull request with description

See [TOOL_DEVELOPMENT.md](docs/TOOL_DEVELOPMENT.md) for architecture details.

## 📄 License

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.

```
Copyright 2025 adk-code contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```

## 🎓 Learning Path

**New to adk-code?**

```
5 min   → QUICK_REFERENCE.md → Start using
1 hour  → ARCHITECTURE.md    → Understand design
3 hours → Full docs          → Contribute
```

## ❓ FAQ

**Q: What's the difference between adk-code and ChatGPT?**  
A: adk-code runs in your terminal with direct filesystem access. No copy-pasting code—just ask.

**Q: Can I use this offline?**  
A: No, it requires an LLM API. But you can use any of 3 providers (Gemini/OpenAI/Vertex).

**Q: Is my code private?**  
A: Yes, only sent to your chosen API provider. Self-hosted options available on request.

**Q: How much does it cost?**  
A: Depends on provider. Gemini has a free tier. OpenAI is ~$0.03/1K tokens.

**Q: Can I build custom tools?**  
A: Yes! Follow the 4-step pattern in TOOL_DEVELOPMENT.md.

## 🚀 What's Next?

- [ ] Add more tool categories (database, API, etc.)
- [ ] Support for local LLMs
- [ ] Web UI option
- [ ] Plugin marketplace

## 💬 Community

- **Issues**: [GitHub Issues](https://github.com/raphaelmansuy/adk-code/issues)
- **Discussions**: [GitHub Discussions](https://github.com/raphaelmansuy/adk-code/discussions)
- **Contributing**: See [CONTRIBUTING.md](CONTRIBUTING.md)

## 🙏 Acknowledgments

Built on:
- [Google ADK](https://github.com/googleapis/google-cloud-go) — Agent framework
- [Charmbracelet](https://github.com/charmbracelet) — Terminal UI
- [Gemini/OpenAI/Vertex AI](https://ai.google.dev) — LLM APIs

## 📈 Stats

- **~1000 lines** of critical code
- **30+ tools** across 8 categories
- **3 LLM backends** supported
- **100% test coverage** target

---

<div align="center">

**Made with ❤️ by the adk-code community**

[⭐ Star us on GitHub](https://github.com/raphaelmansuy/adk-code) | [🐛 Report Bug](https://github.com/raphaelmansuy/adk-code/issues) | [💡 Request Feature](https://github.com/raphaelmansuy/adk-code/issues)

</div>
