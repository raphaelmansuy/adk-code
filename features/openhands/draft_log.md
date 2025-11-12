# OpenHands Features Analysis for Code Agent

**Date**: November 12, 2025  
**Purpose**: Document high-value features from OpenHands that could enhance code_agent  
**Status**: Active Investigation  

## Executive Summary

This document catalogs features discovered in OpenHands (a mature, open-source AI coding agent platform) that could provide significant value to code_agent. OpenHands is production-ready with 65K+ GitHub stars and addresses several critical challenges that code_agent currently lacks.

**Key Finding**: OpenHands has solved several hard problems that code_agent could benefit from:
- Docker-based sandboxing with custom runtime images
- Multi-modal execution (headless, CLI, GUI, GitHub Actions)
- Plugin-based extensibility (VSCode, Jupyter, Agent Skills)
- Microagents for repository-specific customization
- Native MCP (Model Context Protocol) integration
- Integration with major development platforms (GitHub, GitLab, Bitbucket, Slack, Jira, Linear)
- Sophisticated runtime architecture with volume management
- Multi-run session persistence and conversation history

---

## 1. Docker-Based Sandboxing & Custom Runtime Images

### Discovery
OpenHands uses a sophisticated Docker-based sandbox system that allows executing arbitrary code safely. The runtime is containerized and supports custom base images.

### Key Features

**Sandboxing Approach**:
- Every agent execution runs in an isolated Docker container
- Host filesystem is protected via read-only or selective bind mounts
- Supports named volumes and overlay (copy-on-write) mounts
- Process isolation prevents resource exhaustion
- Network can be isolated or selectively exposed

**Custom Runtime Images**:
- Users can provide custom base Docker images (must be Debian-based)
- Automatically embeds OpenHands runtime client into the image
- Supports pre-installing tools, languages, and dependencies
- Multi-architecture support (amd64, arm64)
- Example: `FROM ruby:latest` - OpenHands injects its runtime layer

**Smart Image Caching**:
- Three-tier tagging system for reproducibility and speed:
  1. **Source Tag** (most specific): Includes source code hash
  2. **Lock Tag**: Freezes dependencies (poetry.lock)
  3. **Versioned Tag** (most generic): Base version compatibility
- Smart build optimization:
  - No rebuild if source unchanged
  - Fast rebuild using lock tag if dependencies unchanged
  - Reuses compatible images to speed up iteration

**Runtime Communication**:
- Docker container runs action execution server
- RESTful API communication between backend and container
- Supports shell commands, file operations, Python execution
- Returns observations for agent to process

### Value for Code Agent

**Current Limitation**: code_agent has no execution isolation. Tools run with full host permissions.

**Solution**: Implement Docker-based sandboxing enabling:
- ✅ Safe arbitrary code execution
- ✅ Language/tool isolation (Go, Python, Node, Ruby, etc.)
- ✅ Custom project-specific runtimes
- ✅ No pollution of host system
- ✅ Reproducible execution environments

**Implementation Path**:
1. Add Docker runtime abstraction layer
2. Implement base runtime image builder
3. Add custom image support to config
4. Implement action executor in container
5. Add volume mount management
6. Test with different base images

**Effort**: High (180-220 hours)  
**ROI**: Very High (enables safe execution, critical for production)

---

## 2. Multi-Modal Execution (CLI, Headless, GUI, GitHub Actions)

### Discovery
OpenHands supports four distinct execution modes, each optimized for different use cases.

### Key Features

**Execution Modes**:

1. **GUI Mode** (`uvx openhands serve`)
   - Web interface running on localhost:3000
   - Real-time agent activity visualization
   - Interactive conversation interface
   - Settings UI for model/API configuration
   - Best for: Interactive development, debugging

2. **CLI Mode** (`uvx openhands`)
   - Terminal-based interactive interface
   - Slash commands for settings, help, MCP management
   - Conversation history saved locally
   - Better for: Developer workflows, scriptability
   - Commands: `/settings`, `/help`, `/new`, `/mcp`

3. **Headless Mode** (`poetry run python -m openhands.core.main -t "task"`)
   - Non-interactive script execution
   - Perfect for CI/CD, automation, scripting
   - Single command with task input
   - Supports task files (`-f task.txt`)
   - Environment variable configuration
   - Can work with Docker or native Python
   - Best for: Automation, pipelines, batch processing

4. **GitHub Action**
   - Automatic issue resolution triggered by labels or mentions
   - `fix-me` label or `@openhands-agent` mentions
   - Creates PRs with proposed solutions
   - Iterative: users can comment for refinement
   - Supports custom configurations via repo variables
   - Best for: GitHub-native workflows

**CLI Slash Commands**:
- `/settings` - Configure model, API key, advanced options
- `/help` - View available commands
- `/new` - Start fresh conversation
- `/mcp` - View/manage MCP servers status

**Configuration Flexibility**:
- Environment variables override config
- `.toml` configuration files
- Per-repository custom settings
- GitHub Action variables (LLM_MODEL, OPENHANDS_MAX_ITER, etc.)

### Value for Code Agent

**Current Limitation**: code_agent is REPL-only. No automation mode, no GitHub integration, limited scripting support.

**Solution**: Implement multi-modal execution:
- ✅ GUI (already have web-based)
- ✅ CLI (interactive terminal mode)
- ✅ Headless (non-interactive automation)
- ✅ GitHub Action integration (auto-resolve issues)
- ✅ Slack/chat platform integration (future)

**Implementation Path**:
1. Refactor REPL into independent module
2. Create headless execution engine (non-interactive)
3. Build GitHub Action wrapper
4. Add CLI mode with interactive commands
5. Support task files and automation
6. Test with real GitHub repositories

**Effort**: Very High (200-280 hours for full suite)  
**ROI**: Very High (enables enterprise usage, CI/CD integration)

**Priority Breakdown**:
- Headless mode: 60-80 hours (high priority)
- CLI improvements: 40-60 hours (medium)
- GitHub Action: 80-120 hours (high priority)
- Slack integration: 100-140 hours (future)

---

## 3. Microagents: Repository-Specific Customization

### Discovery
OpenHands has a powerful microagent system that allows per-repository customization without code changes.

### Key Features

**Microagent Types**:

1. **General Microagents** (repo knowledge base)
   - File: `.openhands/microagents/repo.md`
   - Provides general repository guidelines
   - Automatically loaded and injected into context
   - Example: coding standards, architecture principles, testing practices
   - Can be auto-generated by OpenHands analyzing the repo

2. **Keyword-Triggered Microagents** (specialized prompts)
   - Files: `.openhands/microagents/<trigger_name>.md`
   - Activated when specific keywords appear in user prompts
   - Example: `testing.md` triggered by "test", "unit test", "qa"
   - Allows domain-specific behavior (e.g., "fix for deployment")
   - Requires frontmatter metadata (trigger keywords, version)

**Microagent Format**:
```markdown
---
keywords: ["migrate", "migration", "refactor"]
version: "1"
---

# Database Migration Guidelines

When migrating databases:
1. Always create a backup first
2. Test migrations on staging
3. Include rollback procedures
4. Document all schema changes
```

**Benefits**:
- Zero-code customization
- No agent retraining needed
- Per-project guidance
- Scaling guidance across monorepos
- Community-shareable microagents

### Value for Code Agent

**Current Limitation**: code_agent has generic AGENTS.md instructions but no structured microagent system.

**Solution**: Implement microagent system:
- ✅ Auto-discover `.openhands/microagents/` directory
- ✅ Load and merge microagents into system prompt
- ✅ Support keyword-triggered microagents
- ✅ Add microagent creation UI/command
- ✅ Enable per-directory overrides
- ✅ Handle microagent frontmatter

**Implementation Path**:
1. Add microagent discovery logic
2. Implement microagent loader/parser
3. Add keyword matching engine
4. Integrate into prompt building
5. Add `/init-microagent` command
6. Document microagent format

**Effort**: Medium (80-120 hours)  
**ROI**: High (significantly improves customization, enables community contributions)

---

## 4. Runtime Plugin System

### Discovery
OpenHands has a sophisticated plugin system that extends agent capabilities at runtime.

### Key Features

**Built-in Plugins**:

1. **VSCode Plugin**
   - Provides integrated VSCode editor in sandbox
   - Tokenized connection URLs for security
   - Automatic port management
   - Folder synchronization with sandbox
   - Useful for: Visual editing, debugging

2. **Jupyter Plugin**
   - IPython kernel gateway support
   - Kernel specification and management
   - Interactive notebook execution
   - Useful for: Data analysis, prototyping

3. **Agent Skills Plugin**
   - Extends agent with specialized capabilities
   - Framework for custom actions
   - Loaded from plugins directory
   - Useful for: Custom workflows, domain-specific tools

**Plugin Architecture**:
- Plugins inherit from base `Plugin` class
- Registered in `openhands/runtime/plugins/__init__.py`
- Associated with agents via `Agent.sandbox_plugins: list[PluginRequirement]`
- Initialized asynchronously when runtime starts
- Can expose HTTP endpoints (auto-proxied to host)

**Plugin Specification in Agent**:
```python
class MyAgent(Agent):
    sandbox_plugins = [
        PluginRequirement(name="vscode"),
        PluginRequirement(name="jupyter"),
        PluginRequirement(name="custom_tool"),
    ]
```

### Value for Code Agent

**Current Limitation**: code_agent has no extensible plugin system. Tools are hardcoded.

**Solution**: Implement runtime plugin system:
- ✅ Pluggable tool system
- ✅ VSCode editor integration
- ✅ Jupyter support for analysis tasks
- ✅ Custom tool discovery
- ✅ Per-agent plugin configuration
- ✅ Dynamic plugin loading

**Implementation Path**:
1. Design plugin interface
2. Implement plugin registry
3. Create VSCode plugin
4. Create Jupyter plugin
5. Add plugin discovery
6. Wire into runtime
7. Document plugin development

**Effort**: High (140-180 hours)  
**ROI**: High (future-proofs architecture, enables rich editing/analysis)

---

## 5. Native MCP (Model Context Protocol) Integration

### Discovery
OpenHands has deep, native MCP integration supporting multiple transport protocols.

### Key Features

**Supported MCP Transports**:
1. **SSE (Server-Sent Events)** - HTTP with server-sent events
2. **SHTTP (Streamable HTTP)** - Modern streamable HTTP protocol
3. **Stdio** - Direct process communication (via MCP proxy)

**Configuration**:
```toml
[mcp]
# SSE Servers - Recommended for HTTP-based servers
sse_servers = [
    "http://example.com:8080/mcp",
    {url="https://api.example.com/mcp/sse", api_key="..."}
]

# SHTTP Servers - Modern HTTP streaming protocol
shttp_servers = [
    "https://api.example.com/mcp/shttp",
    {url="https://files.example.com/mcp/shttp", timeout=1800}
]

# Stdio Servers - Via MCP proxy tools (recommended)
# Uses supergateway or similar MCP proxy to convert stdio to HTTP
```

**How It Works**:
1. OpenHands reads MCP configuration at startup
2. Connects to configured SSE/SHTTP servers
3. Registers tools from MCP servers
4. Agent can call MCP tools like built-in tools
5. OpenHands routes calls to appropriate server
6. Server responds, OpenHands converts to observation

**MCP Server Recommendations**:
- Proxy-based approach recommended (SSE/SHTTP)
- Direct stdio for development/testing
- Uses supergateway as proxy tool: `supergateway --stdio "tool-command" --port 8080`
- Enables connecting to ecosystem of MCP servers

**Built-in MCP Servers**:
- Filesystem operations
- Web fetch/scraping
- Custom domain tools

### Value for Code Agent

**Current Limitation**: code_agent has no MCP support. Tool ecosystem is hardcoded, not extensible.

**Solution**: Implement native MCP integration:
- ✅ SSE server support
- ✅ SHTTP server support
- ✅ Proxy-based stdio support
- ✅ Tool auto-discovery from MCP
- ✅ JSON configuration
- ✅ Multiple simultaneous MCP servers
- ✅ Error handling and fallbacks

**Implementation Path**:
1. Add MCP client library
2. Implement SSE transport
3. Implement SHTTP transport
4. Add MCP tool registration
5. Create tool proxy/router
6. Add configuration support
7. Test with public MCP servers
8. Document MCP setup

**Effort**: Very High (200-260 hours)  
**ROI**: Very High (future-proofs, connects to ecosystem, enables community extensions)

---

## 6. Integration with Development Platforms

### Discovery
OpenHands integrates natively with major development and project management platforms.

### Key Features

**Version Control Integration**:

1. **GitHub**
   - GitHub Action for auto-resolving issues
   - Label-based triggering (`fix-me` label)
   - Mention-based triggering (`@openhands-agent` comment)
   - Iterative: users can follow up with comments
   - Creates PRs with proposed changes
   - Cloud and local (token-based) support

2. **GitLab**
   - Similar label/mention triggering
   - Repository access via GitLab tokens
   - Merge request creation
   - Cloud and local support

3. **Bitbucket**
   - Bitbucket token authentication
   - Branch and pull request management
   - Repository operations

**Project Management Integration** (Coming Soon):

1. **Jira Cloud & Data Center**
   - Issue auto-resolution
   - Comment mention triggering
   - Label-based delegation
   - Webhook support
   - Service account setup

2. **Linear**
   - Issue auto-resolution
   - Comment-based task assignment
   - GitHub sync integration (auto-detects linked issues)
   - Modern PM platform support

**Chat Platform Integration**:

1. **Slack** (Beta, Cloud-only)
   - Mention `@openhands` in channels
   - Thread-based conversations
   - Repository selection via UI
   - Follow-up messages in threads
   - Format: `@openhands in MyRepo ...`

**Platform Features**:
- Webhook-based triggering
- OAuth/token authentication
- Multi-workspace support
- Custom configuration per platform
- Automatic Git repo detection
- Service account management

### Value for Code Agent

**Current Limitation**: code_agent is isolated. No GitHub, Slack, or PM tool integration.

**Solution**: Implement platform integrations:
- ✅ GitHub Action support
- ✅ GitHub issue/PR handling
- ✅ GitLab/Bitbucket support (future)
- ✅ Slack integration (future)
- ✅ Jira integration (future)
- ✅ Label-based triggering
- ✅ Mention-based triggering

**Implementation Path**:
1. Implement GitHub Action executor
2. Add GitHub API client
3. Add webhook receiver
4. Create PR/issue handlers
5. Add authentication layer
6. Test with real repositories
7. Document setup

**Effort**: High (140-200 hours for GitHub + Slack)  
**ROI**: Very High (enterprise integration, CI/CD adoption)

---

## 7. Session Persistence & Conversation History

### Discovery
OpenHands maintains persistent session state, allowing users to resume conversations and review history.

### Key Features

**Session Management**:
- Sessions stored in `~/.openhands/` directory
- Each session has unique ID
- Full conversation history preserved
- Auto-saved after each turn
- Local file system storage (JSON/SQLite)

**Conversation History**:
- Complete message and observation log
- File changes tracked
- Command execution history
- Agent state snapshots
- Accessible for review and analysis

**Resume Functionality**:
- Resume most recent session
- Resume specific session by ID
- Continue with new task from previous state
- Interactive session picker in CLI
- Shows conversation preview

**Headless Mode Persistence**:
- Logs events to files for analysis
- Structured event format
- Can be piped to analysis tools
- Trajectory storage for evaluation

### Value for Code Agent

**Current Limitation**: Each run is independent. No session history. Long tasks that exceed context are lost.

**Solution**: Add session persistence:
- ✅ Auto-save conversation after each turn
- ✅ Session picker UI
- ✅ Resume with `/resume` command
- ✅ Conversation history export
- ✅ Event logging and analysis
- ✅ Multi-run task continuation

**Implementation Path**:
1. Add session storage layer
2. Implement JSON/SQLite serialization
3. Add session picker to REPL
4. Add `/resume` command
5. Extend `/status` to show session info
6. Test persistence across restarts

**Effort**: Medium-High (120-160 hours)  
**ROI**: Very High (critical for long-running tasks, recovery from crashes)

---

## 8. Memory Condensation & Context Management

### Discovery
OpenHands automatically condenses conversation memory when approaching token limits.

### Key Features

**Memory Condensation**:
- Automatic summarization at context threshold
- Configurable trigger point (default: 75%)
- Preserves critical information
- Reduces context window pressure
- User can manually trigger at any time

**Context Awareness**:
- Token counting per turn
- Shows context utilization
- Warnings approaching limits
- Smart model-aware limits
- Supports multiple LLM providers

**Configuration Options**:
- Adjustable condensation threshold
- Max token limits per model
- Custom summarization prompts
- Model-specific context windows

### Value for Code Agent

**Current Limitation**: No token tracking. Tasks fail silently when context exhausted.

**Solution**: Add context management:
- ✅ Token counting and reporting
- ✅ Context utilization warnings
- ✅ Auto-condensation at threshold
- ✅ Manual `/compact` command
- ✅ Support for multiple models
- ✅ Memory-aware task planning

**Implementation Path**:
1. Add token counter (use model provider APIs)
2. Implement auto-summary detection
3. Create condensation prompt
4. Add `/compact` command
5. Integrate into display
6. Test with various models
7. Tune thresholds

**Effort**: Medium (100-140 hours)  
**ROI**: High (critical for long-running tasks, improves stability)

---

## 9. GitHub Resolver (GitHub Action Integration)

### Discovery
OpenHands provides a specialized GitHub resolver for GitHub Action integration.

### Key Features

**GitHub Action Triggers**:
- `fix-me` label on issues/PRs
- `@openhands-agent` mentions in comments
- Label macro vs comment macro (different scope)
- Macro-based approach (e.g., `@resolveit` custom macro)

**Iterative Resolution**:
1. Agent attempts to resolve issue
2. Creates PR with proposed solution
3. User reviews and comments
4. User adds `fix-me` label to PR or comments `@openhands-agent`
5. Agent refines based on feedback
6. Process repeats until resolved

**Configuration**:
- Custom macro names
- Max iterations limit
- Custom sandbox images
- Target branch specification
- Target runner selection
- Custom LLM models

**Environment Variables**:
- `LLM_MODEL` - Model selection
- `OPENHANDS_MAX_ITER` - Iteration limit
- `OPENHANDS_MACRO` - Custom macro name
- `OPENHANDS_BASE_CONTAINER_IMAGE` - Custom sandbox
- `TARGET_BRANCH` - Merge target
- `TARGET_RUNNER` - Custom runner

**Repository Instructions**:
- Custom `.openhands/repo.md` for project-specific guidance
- Picked up automatically by resolver
- Instructs agent about project conventions

### Value for Code Agent

**Current Limitation**: code_agent has no GitHub Action integration.

**Solution**: Implement GitHub resolver:
- ✅ GitHub Action workflow
- ✅ Label-based triggering
- ✅ Comment-based triggering
- ✅ Iterative resolution loop
- ✅ PR creation and management
- ✅ Custom configuration

**Implementation Path**:
1. Create GitHub Action YAML
2. Implement webhook receiver
3. Add GitHub API client
4. Create issue/PR handlers
5. Implement iteration loop
6. Add configuration support
7. Test with real repositories
8. Document setup steps

**Effort**: High (140-180 hours)  
**ROI**: Very High (enables GitHub workflow, widely requested)

---

## 10. Flexible Configuration System

### Discovery
OpenHands has a flexible, layered configuration system supporting multiple sources.

### Key Features

**Configuration Sources** (in precedence order):
1. CLI flags (highest priority)
2. Environment variables
3. `config.toml` file
4. Default values

**Configuration Sections**:
- Core settings (LLM, API keys, directories)
- Sandbox settings (Docker image, volumes, environment)
- Runtime settings (extra dependencies, startup env vars)
- MCP settings (server configurations)
- Advanced features (memory condensation, max iterations)

**Config Format**:
```toml
[core]
llm_model = "anthropic/claude-sonnet-4-5-20250929"
llm_api_key = "sk-..."

[sandbox]
base_container_image = "nikolaik/python-nodejs:python3.12-nodejs22"
runtime_extra_deps = "pip install numpy pandas"
runtime_startup_env_vars = { DATABASE_URL = "..." }

[mcp]
sse_servers = ["http://localhost:8080/mcp"]

[features]
memory_condensation_enabled = true
memory_condensation_threshold = 0.75
max_iterations = 50
max_budget_per_task = 10.0
```

**Environment Variable Support**:
- `LLM_MODEL`, `LLM_API_KEY`
- `SANDBOX_VOLUMES`, `SANDBOX_USER_ID`
- `LOG_ALL_EVENTS`, `LOG_LEVEL`
- Prefix-based (e.g., `OPENHANDS_*`)

### Value for Code Agent

**Current Limitation**: config_agent has CLI-only configuration. Limited persistence.

**Solution**: Enhance configuration system:
- ✅ TOML configuration files
- ✅ Environment variable support
- ✅ Settings precedence (CLI > ENV > config > defaults)
- ✅ Per-project config directories
- ✅ Configuration validation
- ✅ Settings UI (already have)

**Implementation Path**:
1. Add TOML support
2. Implement config loader
3. Add precedence system
4. Create config validator
5. Add config export/import
6. Test with various setups

**Effort**: Low-Medium (40-60 hours)  
**ROI**: Medium (improves usability, enables automation)

---

## 11. Event Logging & Structured Observability

### Discovery
OpenHands emits structured events for all agent activities, enabling monitoring and analysis.

### Key Features

**Event Types**:
- `thread.started` / `thread.completed`
- `turn.started` / `turn.completed`
- `item.started` / `item.updated` / `item.completed`
- `command_execution` - Shell commands
- `file_change` - File modifications
- `mcp_tool_call` - MCP tool invocations
- `reasoning` - Agent reasoning
- `agent_message` - Agent responses
- `error` - Error events

**Event Format** (JSON Lines):
```json
{type: "turn.started", timestamp: "...", turn_id: "..."}
{type: "command_execution", cmd: "...", exit_code: 0, stdout: "..."}
{type: "file_change", action: "write", path: "...", size: 1024}
{type: "item.completed", item_id: "...", status: "success"}
```

**Logging Options**:
- `LOG_ALL_EVENTS=true` - Log all events to file
- Structured JSON output
- Timestamped events
- Event hierarchy (threads, turns, items)

**Use Cases**:
- Agent behavior analysis
- Performance monitoring
- Debugging and troubleshooting
- Audit trails
- Integration with observability platforms

### Value for Code Agent

**Current Limitation**: code_agent has basic logging. No structured events.

**Solution**: Add structured event logging:
- ✅ Emit events for all actions
- ✅ JSON Lines format
- ✅ Event type hierarchy
- ✅ Timestamp tracking
- ✅ Structured fields
- ✅ Optional export to files
- ✅ Integration with monitoring

**Implementation Path**:
1. Design event schema
2. Add event emitter
3. Emit events from agent loop
4. Create JSON serializer
5. Add log file writer
6. Test event completeness
7. Document event types

**Effort**: Medium (80-120 hours)  
**ROI**: Medium-High (improves debugging, enables monitoring)

---

## 12. Repository-Aware Execution

### Discovery
OpenHands understands repository context and can work with specific branches, repositories, and configurations.

### Key Features

**Repository Selection**:
- `--selected-repo "owner/repo"` flag
- `SANDBOX_SELECTED_REPO` environment variable
- Interactive repo picker in UI
- Works with GitHub, GitLab, Bitbucket

**Branch Management**:
- Specific branch checkout
- Create branches for changes
- Push to branches
- Create PRs between branches

**Repository Detection**:
- Auto-detects from current directory
- Can manually specify
- Resolves repository URLs
- Supports multiple formats: "owner/repo", full HTTPS URL, etc.

**Repository Authentication**:
- GitHub tokens (GITHUB_TOKEN)
- GitLab tokens (GITLAB_TOKEN)
- Bitbucket credentials
- Token exported to sandbox environment

### Value for Code Agent

**Current Limitation**: code_agent works with local files only. Limited repository context.

**Solution**: Enhance repository awareness:
- ✅ Repository auto-detection
- ✅ Repository selection UI
- ✅ Branch management
- ✅ Remote repository operations
- ✅ Token-based authentication
- ✅ PR/MR creation

**Implementation Path**:
1. Add repository detector
2. Add token management
3. Implement Git operations
4. Create branch manager
5. Add PR/MR handler
6. Wire into workspace

**Effort**: Medium (100-140 hours)  
**ROI**: High (enables cloud operations, GitHub workflows)

---

## 13. Multi-LLM Support & Model Switching

### Discovery
OpenHands supports multiple LLM providers with dynamic model switching.

### Key Features

**Supported Providers**:
- Anthropic Claude (multiple versions)
- OpenAI GPT models
- Google models (Gemini, etc.)
- Custom LLM endpoints
- Model registry for extensibility

**Model Switching**:
- `/settings` command to change model
- CLI flag `--model` or env var
- Models with different token limits
- Supports reasoning models (o1, o3)
- Per-model configuration

**Provider Abstraction**:
- Unified provider interface
- Easy to add new providers
- Registry-based discovery
- Capability tracking (token limits, reasoning, etc.)

**Cost Tracking**:
- Per-turn token counting
- Cost estimation
- Budget limit enforcement
- Model-specific pricing

### Value for Code Agent

**Current Limitation**: code_agent supports multiple providers but no easy runtime switching.

**Solution**: Enhance model switching:
- ✅ `/model` command for switching
- ✅ Easy provider configuration
- ✅ Token tracking and budgets
- ✅ Cost estimation
- ✅ Model capability detection

**Implementation Path**:
1. Create unified model interface
2. Build provider registry
3. Add model discovery
4. Implement capability tracking
5. Add budget enforcement
6. Create provider factory
7. Test with multiple providers

**Effort**: Medium (80-120 hours)  
**ROI**: Medium (improves flexibility, enables cost control)

---

## Feature Priority Matrix

| Feature | Value | Effort | Impact | ROI | Priority |
|---------|-------|--------|--------|-----|----------|
| Docker Sandboxing | Very High | High | Very High | 1.8x | 🔴 P0 |
| Multi-Modal Execution | Very High | Very High | Very High | 1.6x | 🔴 P0 |
| Microagents | High | Medium | High | 2.0x | 🟠 P1 |
| MCP Integration | Very High | Very High | Very High | 1.4x | 🟠 P1 |
| Runtime Plugins | High | High | Medium | 1.5x | 🟠 P1 |
| Platform Integrations | High | High | High | 1.7x | 🟠 P1 |
| Session Persistence | Very High | Medium | Very High | 2.2x | 🔴 P0 |
| Memory Condensation | High | Medium | High | 1.9x | 🟠 P1 |
| GitHub Resolver | High | High | High | 1.8x | 🟠 P1 |
| Config System | Medium | Low | Low | 1.4x | 🟢 P2 |
| Event Logging | Medium | Medium | Medium | 1.3x | 🟢 P2 |
| Repository Awareness | High | Medium | Medium | 1.6x | 🟠 P1 |
| Multi-LLM Support | Medium | Medium | Medium | 1.2x | 🟢 P2 |

---

## Implementation Roadmap

### Phase 1: Core Execution Safety (Weeks 1-3)
**Essential for production deployment** - 120-160 hours

- [ ] Docker-based sandboxing
- [ ] Custom runtime image support
- [ ] Session persistence (basic)
- [ ] Event logging

**Target**: Safe, reproducible, traceable execution

### Phase 2: Operational Modes (Weeks 4-6)
**Enable diverse deployment patterns** - 180-240 hours

- [ ] Headless mode (for automation)
- [ ] CLI mode improvements (interactive)
- [ ] GitHub Action integration
- [ ] Configuration system enhancement

**Target**: Enterprise-ready deployment flexibility

### Phase 3: Extensibility & Customization (Weeks 7-9)
**Future-proof architecture** - 200-260 hours

- [ ] Microagents implementation
- [ ] MCP integration
- [ ] Runtime plugin system
- [ ] Custom tool support

**Target**: Community-driven extensions, ecosystem

### Phase 4: Platform Integrations (Weeks 10-12)
**Enterprise & team workflows** - 140-200 hours

- [ ] GitHub/GitLab/Bitbucket integration
- [ ] Slack integration
- [ ] Jira/Linear integration (future)
- [ ] Project management workflows

**Target**: Enterprise adoption, team collaboration

### Phase 5: Intelligence & Context (Weeks 13-14)
**Improve reasoning & efficiency** - 100-140 hours

- [ ] Memory condensation
- [ ] Repository awareness
- [ ] Smart context management
- [ ] Cost optimization

**Target**: Longer tasks, bigger codebases, cost control

---

## Comparison: OpenHands vs Code Agent

| Aspect | OpenHands | Code Agent | Gap | Priority |
|--------|-----------|-----------|-----|----------|
| **Execution Safety** | Docker sandbox | Native host execution | ❌ Critical | P0 |
| **Session Persistence** | Yes (SQLite) | No | ❌ Critical | P0 |
| **Multi-Modal Execution** | GUI, CLI, Headless, GitHub | REPL only | ❌ Critical | P0 |
| **Docker Support** | First-class | None | ❌ Major | P0 |
| **Plugin System** | Extensive | None | ❌ Major | P1 |
| **Microagents** | Yes (keyword + repo) | Basic AGENTS.md | ❌ Significant | P1 |
| **MCP Support** | Native (SSE, SHTTP, stdio) | None | ❌ Major | P1 |
| **GitHub Integration** | Full (Issues, PRs, Action) | None | ❌ Major | P1 |
| **Slack Integration** | Beta (Cloud) | None | ❌ Important | P1 |
| **Memory Condensation** | Auto-summary | Manual | ❌ Important | P1 |
| **Event Logging** | Structured JSON | Basic logs | ❌ Useful | P2 |
| **Config System** | TOML + ENV | CLI + ENV | ✅ Minor | P2 |
| **Repository Awareness** | Smart detection | Manual | ❌ Useful | P1 |
| **Model Flexibility** | Multiple providers | Multiple providers | ✅ Similar | P2 |

---

## Key Learnings from OpenHands

### 1. Sandbox First
Docker sandboxing should be foundational, not an afterthought. It's essential for production use and user trust.

### 2. Multiple Execution Modes Matter
Different users need different modes:
- Developers want interactive CLI
- CI/CD needs headless automation
- Teams want GitHub/Slack integration

### 3. Customization Without Code Changes
Microagents enable per-project customization without agent retraining. Powerful pattern.

### 4. Extensibility Through Standards
MCP provides standard way to extend. Avoids tool sprawl.

### 5. Session Persistence is Critical
Crashing or long tasks are inevitable. Resume capability essential for production.

### 6. Event-Driven Architecture
Structured events enable monitoring, debugging, and integration.

### 7. Repository Context Matters
Understanding repository structure, configuration, and conventions improves agent effectiveness.

### 8. Cost and Context Control
Token tracking, budget limits, and auto-summary prevent runaway costs.

---

## Integration Patterns

### Pattern 1: Docker-Based Execution Flow
```
User Task → Backend → Docker Client → Start Container → 
Action Executor → Execute Command → Return Observation → Backend → User
```

### Pattern 2: Multi-Modal Dispatcher
```
User Input (CLI/GUI/GitHub/Slack) → Normalize to Task → Execute via Agent → 
Format Output (Terminal/Web/PR/Message) → Deliver to User
```

### Pattern 3: Microagent Injection
```
Repository Detected → Scan .openhands/microagents/ → Load Microagent Files → 
Merge into System Prompt → Execute with Enhanced Context
```

### Pattern 4: MCP Tool Bridge
```
Agent Requests Tool → Route to MCP Server → MCP Handler → Execute Tool → 
Return Result → Convert to Observation → Agent Continues
```

---

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Docker/Host compatibility issues | Start with standard Debian images, test extensively |
| Context explosion with multiple data sources | Implement careful prioritization, auto-summary |
| MCP server reliability | Implement timeouts, fallbacks, error handling |
| GitHub Action rate limits | Implement backoff, caching, cost control |
| Session storage bloat | Implement compression, archival, cleanup policies |
| Plugin conflicts | Namespace isolation, version pinning, conflict detection |

---

## Next Steps

### Immediate (This Week)
1. ✅ Complete OpenHands analysis
2. Create comparison with Codex findings
3. Get stakeholder feedback on priorities

### Short-term (Next 2 Weeks)
1. Design Docker sandbox integration
2. Create headless execution prototype
3. Plan microagent system
4. Evaluate MCP libraries

### Medium-term (Weeks 3-6)
1. Implement Phase 1 features
2. Beta test with community
3. Refine based on feedback

### Long-term (Weeks 7+)
1. Implement Phase 2-5 features
2. Build ecosystem (microagents, plugins)
3. Community growth and support

---

## References

- **OpenHands Docs**: https://docs.all-hands.dev/
- **OpenHands GitHub**: https://github.com/OpenHands/OpenHands
- **OpenHands Architecture**: https://docs.all-hands.dev/openhands/usage/architecture/backend
- **MCP Specification**: https://modelcontextprotocol.io/
- **Paper on ArXiv**: https://arxiv.org/abs/2407.16741

---

## Document History

| Date | Change | Author |
|------|--------|--------|
| 2025-11-12 | Initial analysis | Copilot |

---

**Status**: ✅ Complete for initial investigation  
**Next Review**: After Phase 1 planning  
**Recommended Action**: Prioritize Docker sandboxing + multi-modal execution for Phase 1
