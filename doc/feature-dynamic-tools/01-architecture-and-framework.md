# Feature Comparison: Architecture & Framework Foundation

## Overview
This document compares the foundational architecture and frameworks that power code_agent and Cline, two sophisticated AI coding assistants with different approaches to agent development.

---

## Code Agent Architecture

### Framework: Google ADK (Agent Development Kit) for Go
- **Language**: Go 1.24+
- **Foundation**: Google's ADK Go framework (llmagent pattern)
- **Model**: Gemini 2.5 Flash (via google.golang.org/genai)
- **Paradigm**: Code-first, modular agent development

### Architecture Layers

```
┌─────────────────────────────────────────────────────────────┐
│ CLI Interface (main.go)                                     │
│ - Interactive REPL                                           │
│ - Command history                                            │
│ - Rich terminal rendering                                    │
└─────────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────────┐
│ Agent Layer (coding_agent.go)                               │
│ - LLM Agent (llmagent.New)                                   │
│ - Dynamic tool registration                                  │
│ - System prompt construction                                 │
│ - Session management (ADK runner)                            │
└─────────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────────┐
│ Tool System (tools/ package)                                 │
│ - Registry pattern                                           │
│ - 14+ categorized tools                                      │
│ - Self-registering tool constructors                         │
│ - Input/Output struct definitions                            │
└─────────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────────┐
│ Supporting Systems                                           │
│ - Workspace management (multi-workspace support)             │
│ - Display system (rich terminal rendering)                   │
│ - File I/O operations                                        │
│ - Execution environment                                      │
└─────────────────────────────────────────────────────────────┘
```

### Key Architectural Components

| Component | Purpose | Implementation |
|-----------|---------|-----------------|
| **ADK Runner** | Orchestrates agent lifecycle | `google.golang.org/adk/runner` |
| **LLM Agent** | Hosts tool definitions & prompts | `llmagent.New()` with Config |
| **Tool Registry** | Central tool management | Custom registry in `tools/registry.go` |
| **Session Service** | Maintains conversation history | ADK in-memory service (productionizable) |
| **Workspace Manager** | Multi-workspace/monorepo support | `workspace/manager.go` with smart detection |
| **Display System** | Rich terminal output | `display/renderer.go` + streaming |

### Tool Registration Pattern

```go
func NewReadFileTool() (tool.Tool, error) {
    // 1. Define handler function
    handler := func(ctx tool.Context, input ReadFileInput) ReadFileOutput {
        // Implementation
    }
    
    // 2. Create function tool
    t, err := functiontool.New(functiontool.Config{
        Name:        "read_file",
        Description: "Reads file contents...",
    }, handler)
    
    // 3. Self-register with metadata
    if err == nil {
        Register(ToolMetadata{
            Tool:      t,
            Category:  CategoryFileOperations,
            Priority:  0,
            UsageHint: "Examine code, read configs...",
        })
    }
    
    return t, err
}
```

### Deployment Model
- **Execution**: CLI binary (`./code-agent`)
- **Environment**: Local/server-based
- **API Key**: GOOGLE_API_KEY (Gemini API)
- **Session Store**: In-memory (can be productionized to database)
- **Scalability**: Cloud Run/GKE ready (ADK deployment patterns)

---

## Cline Architecture

### Framework: VS Code Extension + Model Context Protocol (MCP)
- **Language**: TypeScript
- **Foundation**: VS Code Extension API + Model Context Protocol
- **Model**: Claude Sonnet + support for OpenRouter, Anthropic, OpenAI, Google Gemini, AWS Bedrock, Azure, Cerebras, Groq, Local LMs
- **Paradigm**: Human-in-the-loop, GUI-driven agent within IDE

### Architecture Layers

```
┌─────────────────────────────────────────────────────────────┐
│ VS Code Webview UI (VscodeDiffViewProvider)                  │
│ - Diff view for file changes                                 │
│ - Chat interface                                             │
│ - Human approval gates                                       │
│ - Checkpoint/restore functionality                           │
└─────────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────────┐
│ Extension Controller (core/controller/)                      │
│ - Command handlers                                           │
│ - WebviewProvider management                                 │
│ - VS Code API integration                                    │
│ - Model selection and auth                                   │
└─────────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────────┐
│ Agent Core (core/assistant-message/)                         │
│ - LLM-driven task execution                                  │
│ - Tool calling orchestration                                 │
│ - Context management                                         │
│ - Terminal integration                                       │
└─────────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────────┐
│ Tool Systems                                                 │
│ - MCP Hub (McpHub.ts) - Dynamic MCP tool loading             │
│ - Built-in tools (read/write/execute)                        │
│ - Browser automation (Computer Use)                          │
│ - Terminal execution (shell integration)                     │
└─────────────────────────────────────────────────────────────┘
                        ↓
┌─────────────────────────────────────────────────────────────┐
│ Supporting Systems                                           │
│ - Authentication (AuthService)                               │
│ - Telemetry (TelemetryService)                               │
│ - Tree-sitter for AST analysis                               │
│ - File system watcher (chokidar)                             │
│ - Ripgrep integration for search                             │
└─────────────────────────────────────────────────────────────┘
```

### Key Architectural Components

| Component | Purpose | Implementation |
|-----------|---------|-----------------|
| **VS Code Extension** | Host environment & UI | `extension.ts` + `VscodeWebviewProvider` |
| **Diff View Provider** | Visual change preview | `VscodeDiffViewProvider` |
| **MCP Hub** | Dynamic tool loading | `McpHub.ts` manages MCP connections |
| **Controller** | Event handling & task flow | `core/controller/` + `WebviewProvider` |
| **Auth Service** | Multi-provider model support | `AuthService` handles API keys/credentials |
| **Terminal Integration** | Shell execution | VS Code terminal shell integration API |
| **AST Analysis** | Tree-sitter code parsing | `tree-sitter` for syntax understanding |

### MCP Integration

```typescript
// MCP Server Discovery & Tool Loading
class McpHub {
    private connections: McpConnection[] = [];
    
    // Loads MCP servers from configuration
    async initializeMcpServers() {
        // Watch for MCP config changes
        // Connect to each MCP server
        // Register tools from each server
        // Handle notifications
    }
    
    // Calls tools from any MCP server
    async callTool(serverName: string, toolName: string, args: any) {
        // Route to appropriate MCP connection
        // Execute tool with timeout handling
    }
}
```

### Deployment Model
- **Execution**: VS Code Extension Marketplace
- **Environment**: Developer IDE (VS Code)
- **Model Selection**: Multi-provider architecture (30+ model options)
- **UI**: Sidebar webview with diff preview
- **Human Approval**: Required for all file changes & terminal commands
- **Scalability**: Built-in for enterprise (Auth, telemetry, security)

---

## Comparative Analysis

### Execution Model
| Aspect | Code Agent | Cline |
|--------|-----------|-------|
| **Deployment** | Standalone CLI | VS Code Extension |
| **Runtime** | Local/Server binary | IDE process |
| **Models** | Gemini only | 30+ providers |
| **User Input** | Text-based REPL | GUI sidebar + webview |
| **Approval Flow** | Agent autonomous | Human-in-the-loop gates |
| **Extensibility** | Tool registry | MCP protocol |

### Framework Philosophy
| Aspect | Code Agent | Cline |
|--------|-----------|-------|
| **Design** | Framework-first (ADK) | Extension-first (VS Code) |
| **Paradigm** | Autonomous agent | Human-supervised agent |
| **Code Style** | Go (compiled, typed) | TypeScript (dynamic, bundled) |
| **Development** | SDK/API focused | IDE integration focused |
| **Complexity** | Simpler internal structure | Complex IDE integration |

### Strengths

**Code Agent**:
- Clean, modular Go architecture
- Type-safe tool definitions
- Modern ADK framework patterns
- Production-ready deployment (Cloud Run/GKE)
- Lightweight binary
- Easy local testing

**Cline**:
- Integrated IDE experience (no context switching)
- Human approval gates (safety critical)
- Multi-model support (flexibility)
- Visual diff preview (code review experience)
- Browser automation (UI testing)
- Workspace checkpoint/restore

### Limitations

**Code Agent**:
- Single model (Gemini)
- No browser automation
- CLI-based UX (no visual diff)
- Limited to terminal users
- Requires manual context management

**Cline**:
- Complex VS Code API dependencies
- Heavier resource footprint
- Longer startup time
- Requires IDE
- MCP adds complexity

---

## Framework Comparison: ADK vs MCP

### ADK (Google Agent Development Kit)

**Design Philosophy**: Developer SDK for agent systems
- **Tool Definition**: Function-based (Input/Output Go structs)
- **Registration**: Static registration in agent constructor
- **Communication**: Process-internal (Go function calls)
- **Extensibility**: Code-based tool implementation
- **Multi-agent**: Built-in composition patterns (Sequential, Parallel)
- **Deployment**: Containerizable, cloud-native ready

**Tool System**:
```go
tools := []tool.Tool{
    readFileTool,
    writeFileTool,
    executeCommandTool,
    // ... more tools
}

agent := llmagent.New(llmagent.Config{
    Tools: tools,
    // ...
})
```

### MCP (Model Context Protocol)

**Design Philosophy**: Open protocol for tool discovery
- **Tool Definition**: JSON schema descriptions over protocol
- **Registration**: Dynamic discovery via MCP servers
- **Communication**: Remote (HTTP, SSE, stdio transports)
- **Extensibility**: Process-based (separate MCP server binaries)
- **Multi-agent**: Not built-in (focuses on tool layer)
- **Deployment**: Distributed, service-oriented

**Tool System**:
```typescript
// Tools exposed via MCP servers
// Discovery via client connections
// Remote calls via standardized protocol
mcpHub.getTools() // All available tools across all servers
mcpHub.callTool(serverName, toolName, args)
```

### Comparative Matrix

| Dimension | ADK | MCP |
|-----------|-----|-----|
| **Scope** | Full agent framework | Tool protocol only |
| **Coupling** | Tightly integrated | Loosely coupled |
| **Transport** | In-process (Go functions) | Over protocol (HTTP/stdio) |
| **Configuration** | Code-based | Config file based |
| **Tool Discovery** | Static (at startup) | Dynamic (live) |
| **Scalability** | Vertical (single process) | Horizontal (distributed servers) |
| **Latency** | Minimal | Moderate (network overhead) |
| **Complexity** | Lower (single language) | Higher (protocol/serialization) |
| **Standardization** | Google-specific | Industry-wide protocol |
| **Adoption** | Growing (official framework) | Rapid (industry standard) |

---

## Conclusion

**Code Agent** leverages the modern ADK framework to provide a clean, type-safe agent architecture optimized for cloud deployment. The tool system is straightforward and performant, suitable for backend/server-side agent applications.

**Cline** embraces the extensible MCP protocol within a VS Code extension, prioritizing IDE integration, human oversight, and flexibility through multi-model support. It's designed for interactive development workflows where the developer remains in control.

**For Next Steps**:
- See [Tool System Comparison](./02-tool-system.md) for deep dive into available tools
- See [Extensibility & Customization](./04-extensibility.md) for tool creation patterns
- See [Deployment & Scalability](./07-deployment.md) for production considerations

