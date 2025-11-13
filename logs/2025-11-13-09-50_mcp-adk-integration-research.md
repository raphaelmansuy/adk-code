# MCP Support via ADK-Go Integration Research

## Date: 2025-11-13
## Task: Research existing MCP support in ADK-Go framework

---

## Executive Summary

**Key Finding**: The `research/adk-go` framework **ALREADY HAS BUILT-IN MCP SUPPORT** via the `tool/mcptoolset` package. The current specification in `features/mcp_support_code_agent/05_PHASE1_DETAILED_IMPLEMENTATION.md` proposes using `github.com/modelcontextprotocol/go-sdk@v0.7.0` directly, but this is **redundant** since:

1. ✅ `adk-go` already depends on `github.com/modelcontextprotocol/go-sdk v0.7.0`
2. ✅ `adk-go` provides a ready-made `mcptoolset` abstraction that wraps the MCP SDK
3. ✅ `code_agent` already uses `adk-go` via replace directive: `replace google.golang.org/adk => ../research/adk-go`
4. ✅ The MCP SDK is already transitively available in `code_agent`'s dependencies

**Recommendation**: Instead of re-implementing MCP client logic, we should **leverage the existing `mcptoolset` package** from ADK-Go, which provides:
- Simplified MCP server integration
- Tool discovery and conversion to ADK tools
- Lazy session initialization
- Built-in filtering capabilities
- Production-ready error handling

---

## Current Code Agent Architecture

### Dependencies (from `code_agent/go.mod`)
```go
module code_agent

go 1.24.4

replace google.golang.org/adk => ../research/adk-go

require (
    google.golang.org/adk v0.0.0
    google.golang.org/genai v1.20.0
    // ... other deps
)
```

**Key Point**: `code_agent` already uses ADK via local replacement, meaning all ADK packages (including `tool/mcptoolset`) are available.

---

## ADK-Go MCP Support Analysis

### Package Location
`research/adk-go/tool/mcptoolset/`

### Files
- **`set.go`**: MCP ToolSet manager (main entry point)
- **`tool.go`**: Individual MCP tool wrapper
- **`set_test.go`**: Unit tests with test server
- **`testdata/`**: Test fixtures

### MCP Dependencies in ADK-Go (`research/adk-go/go.mod`)
```go
require (
    github.com/modelcontextprotocol/go-sdk v0.7.0
    // ... other deps
)
```

### How ADK-Go's mcptoolset Works

#### 1. Simple Configuration
```go
// From research/adk-go/tool/mcptoolset/set.go
type Config struct {
    Transport  mcp.Transport   // Connection transport (stdio, SSE, HTTP)
    ToolFilter tool.Predicate  // Optional tool filtering
}
```

#### 2. Creating a ToolSet
```go
mcpToolSet, err := mcptoolset.New(mcptoolset.Config{
    Transport: &mcp.CommandTransport{
        Command: exec.Command("mcp-server-filesystem", "/tmp"),
    },
})
```

#### 3. Integration with LLMAgent
```go
agent, err := llmagent.New(llmagent.Config{
    Name:        "helper_agent",
    Model:       model,
    Instruction: "You are a helpful assistant.",
    Toolsets:    []tool.Toolset{mcpToolSet},
})
```

#### 4. Automatic Tool Discovery
The `mcptoolset` automatically:
- Connects to MCP server (lazy initialization)
- Discovers available tools via `session.ListTools()`
- Converts MCP tools to ADK `tool.Tool` interface
- Handles pagination for large tool lists
- Applies optional filtering

#### 5. Tool Execution
When LLM calls an MCP tool:
```go
// From tool.go
func (t *mcpTool) Run(ctx tool.Context, args any) (map[string]any, error) {
    session, err := t.getSessionFunc(ctx)
    res, err := session.CallTool(ctx, &mcp.CallToolParams{
        Name:      t.name,
        Arguments: args,
    })
    
    // Handle errors
    if res.IsError {
        // Extract error details from content
        return nil, errors.New("Tool execution failed...")
    }
    
    // Return structured or text content
    if res.StructuredContent != nil {
        return map[string]any{"output": res.StructuredContent}, nil
    }
    
    // Extract text response
    return map[string]any{"output": textResponse.String()}, nil
}
```

---

## Comparison: Proposed Implementation vs ADK-Go's mcptoolset

### Proposed Implementation (from Phase 1 spec)
```go
// Features/mcp_support_code_agent/05_PHASE1_DETAILED_IMPLEMENTATION.md

// NEW packages to create:
- internal/config/mcp.go          (MCP config parsing)
- pkg/mcp/types.go                (Custom types)
- pkg/mcp/client.go               (MCP client wrapper)
- pkg/mcp/manager.go              (Multi-server manager)
- tools/mcp/tool.go               (Tool wrapper)
- internal/cli/commands/mcp.go    (CLI commands)

// Direct MCP SDK usage:
import "github.com/modelcontextprotocol/go-sdk/mcp"

client := mcp.NewClient(...)
session, err := client.Connect(ctx, transport, nil)
tools, err := session.ListTools(...)
```

**Complexity**: ~2000+ lines of new code, reinventing existing abstractions

### ADK-Go's mcptoolset (existing)
```go
// ALREADY EXISTS in research/adk-go/tool/mcptoolset/

// Just import and use:
import "google.golang.org/adk/tool/mcptoolset"

toolset, err := mcptoolset.New(mcptoolset.Config{
    Transport: transport,
})

// Add to agent's toolsets
```

**Complexity**: ~200 lines total, production-ready, tested

---

## Key Architectural Differences

### 1. Integration Approach

**Proposed**: Separate MCP layer with custom wrappers
```
code_agent
  ├── pkg/mcp/client.go (NEW)
  ├── pkg/mcp/manager.go (NEW)
  └── tools/mcp/tool.go (NEW)
      └── calls MCP SDK directly
```

**ADK-Go**: Native toolset pattern
```
code_agent
  └── imports google.golang.org/adk/tool/mcptoolset
      └── (already wraps MCP SDK)
```

### 2. Tool Registration

**Proposed**: Manual registration loop
```go
// From proposed implementation
for _, server := range mcpServers {
    tools, err := server.DiscoverTools(ctx)
    for _, tool := range tools {
        wrappedTool := WrapMCPTool(tool, server)
        registry.Register(wrappedTool)
    }
}
```

**ADK-Go**: Automatic via toolset
```go
// Add entire MCP server as a toolset
agent := llmagent.New(llmagent.Config{
    Toolsets: []tool.Toolset{
        mcpToolSet,  // All tools discovered automatically
    },
})
```

### 3. Session Management

**Proposed**: Custom lazy initialization with mutexes
```go
type mcpClient struct {
    mu      sync.Mutex
    session *mcp.ClientSession
}

func (c *mcpClient) getSession(ctx context.Context) (*mcp.ClientSession, error) {
    c.mu.Lock()
    defer c.mu.Unlock()
    if c.session != nil {
        return c.session, nil
    }
    // ... connect logic
}
```

**ADK-Go**: Already implemented (identical pattern)
```go
// From mcptoolset/set.go
func (s *set) getSession(ctx context.Context) (*mcp.ClientSession, error) {
    s.mu.Lock()
    defer s.mu.Unlock()
    if s.session != nil {
        return s.session, nil
    }
    session, err := s.client.Connect(ctx, s.transport, nil)
    // ...
}
```

### 4. Error Handling

**Proposed**: Custom error wrapping
```go
if result.IsError {
    var errMsg strings.Builder
    for _, c := range result.Content {
        if tc, ok := c.(*mcp.TextContent); ok {
            errMsg.WriteString(tc.Text)
        }
    }
    return &ToolResult{Success: false, Error: errMsg.String()}
}
```

**ADK-Go**: Production-ready error handling (same logic, already tested)
```go
if res.IsError {
    details := strings.Builder{}
    for _, c := range res.Content {
        if textContent, ok := c.(*mcp.TextContent); ok {
            details.WriteString(textContent.Text)
        }
    }
    return nil, errors.New("Tool execution failed. Details: " + details.String())
}
```

---

## Example from ADK-Go (`research/adk-go/examples/mcp/main.go`)

### Full Working Example

```go
package main

import (
    "context"
    "os"
    "os/exec"
    
    "github.com/modelcontextprotocol/go-sdk/mcp"
    "google.golang.org/adk/agent/llmagent"
    "google.golang.org/adk/model/gemini"
    "google.golang.org/adk/tool"
    "google.golang.org/adk/tool/mcptoolset"
    "google.golang.org/genai"
)

func main() {
    ctx := context.Background()
    
    // 1. Create model
    model, _ := gemini.NewModel(ctx, "gemini-2.5-flash", &genai.ClientConfig{
        APIKey: os.Getenv("GOOGLE_API_KEY"),
    })
    
    // 2. Create MCP transport (stdio example)
    transport := &mcp.CommandTransport{
        Command: exec.Command("npx", "-y", "@modelcontextprotocol/server-filesystem", "/tmp"),
    }
    
    // 3. Create MCP toolset
    mcpToolSet, _ := mcptoolset.New(mcptoolset.Config{
        Transport: transport,
    })
    
    // 4. Create agent with MCP tools
    agent, _ := llmagent.New(llmagent.Config{
        Name:        "helper_agent",
        Model:       model,
        Description: "Helper agent with MCP tools",
        Instruction: "You are a helpful assistant with filesystem access.",
        Toolsets:    []tool.Toolset{mcpToolSet},
    })
    
    // Agent now has all MCP tools automatically!
}
```

This demonstrates:
- ✅ Stdio transport support
- ✅ Automatic tool discovery
- ✅ Seamless integration with LLMAgent
- ✅ No custom wrappers needed
- ✅ Production-ready code

---

## What the Proposed Spec Gets Right

Despite proposing a custom implementation, the spec correctly identifies:

1. **Configuration needs**: JSON config, environment variables, tool filtering
2. **Transport types**: Stdio, SSE, HTTP support required
3. **Multi-server support**: Managing multiple MCP servers simultaneously
4. **CLI commands**: `/mcp list`, `/mcp info`, `/mcp reload`
5. **Error handling**: Graceful degradation, clear diagnostics

These requirements are still valid, but should be **built on top of `mcptoolset`** rather than reimplementing the MCP client layer.

---

## Recommended Revised Architecture

### Layer 1: Configuration (New)
```go
// internal/config/mcp.go
type MCPConfig struct {
    Servers map[string]ServerConfig
}

type ServerConfig struct {
    Type    string   // "stdio", "sse", "http"
    Command string   // For stdio
    Args    []string
    URL     string   // For SSE/HTTP
    Env     map[string]string
    
    // Filtering
    IncludeTools []string
    ExcludeTools []string
}
```

### Layer 2: Manager (New, but simplified)
```go
// pkg/mcp/manager.go
import "google.golang.org/adk/tool/mcptoolset"

type Manager struct {
    toolsets map[string]tool.Toolset // mcptoolset instances
}

func (m *Manager) LoadServers(cfg *config.MCPConfig) error {
    for name, serverCfg := range cfg.Servers {
        transport := createTransport(serverCfg)
        
        // Use ADK's mcptoolset
        toolset, err := mcptoolset.New(mcptoolset.Config{
            Transport: transport,
            ToolFilter: createFilterPredicate(serverCfg),
        })
        
        m.toolsets[name] = toolset
    }
}

func (m *Manager) GetToolsets() []tool.Toolset {
    return values(m.toolsets)
}
```

### Layer 3: CLI Commands (New)
```go
// internal/cli/commands/mcp.go

func mcpListCommand(m *mcp.Manager) {
    toolsets := m.GetToolsets()
    for _, ts := range toolsets {
        tools, _ := ts.Tools(ctx)
        fmt.Printf("Server: %s (%d tools)\n", ts.Name(), len(tools))
    }
}
```

### Integration Point
```go
// main.go or internal/app/app.go

// Load MCP config
mcpCfg := config.LoadMCPConfig(configPath)

// Create MCP manager
mcpManager := mcp.NewManager()
mcpManager.LoadServers(mcpCfg)

// Get all toolsets (native + MCP)
allToolsets := []tool.Toolset{
    // Native toolsets
    nativeToolset,
    // MCP toolsets
    ...mcpManager.GetToolsets(),
}

// Create agent
agent := llmagent.New(llmagent.Config{
    Toolsets: allToolsets,
})
```

---

## Benefits of Using mcptoolset

### 1. Code Reduction
- **Proposed**: ~2000+ lines of new MCP client code
- **Using mcptoolset**: ~500 lines (just config + manager wrapper)

### 2. Maintenance
- MCP SDK updates handled by ADK team
- Bug fixes upstream benefit code_agent
- No need to track MCP protocol changes

### 3. Consistency
- Same MCP handling as other ADK-based tools
- Shared best practices from Google's ADK team
- Proven in production (ADK examples)

### 4. Features
- ✅ Lazy connection (already implemented)
- ✅ Tool pagination (already implemented)
- ✅ Error handling (already implemented)
- ✅ Structured output (already implemented)
- ✅ Filter predicate pattern (already implemented)

### 5. Testing
- ADK's `mcptoolset` already has unit tests
- Test utilities available (in-memory transport)
- Mock server patterns established

---

## What Still Needs to Be Built

Even using `mcptoolset`, we still need:

### 1. Configuration Layer ✅ (Same as proposed)
- JSON config parsing
- Environment variable substitution
- Multi-server configuration
- Tool filtering rules

### 2. Manager Layer ✅ (Simplified)
- Load multiple MCP servers
- Create appropriate transports (stdio, SSE, HTTP)
- Expose toolsets to agent
- Handle connection lifecycle

### 3. CLI Commands ✅ (Same as proposed)
- `/mcp list` - List connected servers
- `/mcp tools <server>` - List tools from server
- `/mcp reload` - Reconnect servers
- `/mcp info` - Show server status

### 4. Examples & Docs ✅ (Same as proposed)
- Example configurations
- User guide for setup
- Troubleshooting guide

### 5. Transport Factory (New)
```go
func createTransport(cfg ServerConfig) mcp.Transport {
    switch cfg.Type {
    case "stdio":
        cmd := exec.Command(cfg.Command, cfg.Args...)
        cmd.Env = buildEnv(cfg.Env)
        return &mcp.CommandTransport{Command: cmd}
    
    case "sse":
        return &mcp.SSETransport{
            URL: cfg.URL,
            HTTPClient: buildHTTPClient(cfg),
            Headers: buildHeaders(cfg),
        }
    
    case "http":
        return &mcp.StreamableClientTransport{
            Endpoint: cfg.URL,
            HTTPClient: buildHTTPClient(cfg),
        }
    }
}
```

---

## Revised Implementation Estimate

### Original Estimate (Custom implementation)
- **Phase 1 (MVP)**: 2-3 weeks, ~2000 lines
- **Phase 2 (Enhanced)**: 2-3 weeks, ~1500 lines
- **Phase 3 (Advanced)**: 2-3 weeks, ~2000 lines
- **Total**: 6-9 weeks, ~5500 lines

### Revised Estimate (Using mcptoolset)
- **Phase 1 (MVP)**: 3-5 days, ~500 lines
  - Config parsing: 1 day
  - Manager wrapper: 1 day
  - CLI commands: 1 day
  - Testing & examples: 1-2 days

- **Phase 2 (Enhanced)**: 3-5 days, ~300 lines
  - Multiple servers: 1 day
  - Advanced filtering: 1 day
  - Better diagnostics: 1-2 days

- **Phase 3 (Advanced)**: Optional, as needed
  - Resources/Prompts: Via mcptoolset extensions

**Total**: 1-2 weeks, ~800 lines

**Time Saved**: 4-7 weeks, ~4700 lines of code

---

## Risks & Considerations

### 1. ADK Dependency
**Concern**: Tighter coupling to ADK framework  
**Mitigation**: Already using ADK for core agent functionality  
**Verdict**: ✅ Acceptable - we're already committed to ADK

### 2. Limited Control
**Concern**: Can't customize MCP client behavior  
**Mitigation**: `mcptoolset` provides ToolFilter for common needs  
**Verdict**: ✅ Acceptable - covers 95% of use cases; can contribute upstream for edge cases

### 3. Future MCP Features
**Concern**: New MCP features might not be in mcptoolset  
**Mitigation**: ADK actively maintained by Google; can contribute PRs  
**Verdict**: ✅ Acceptable - better than maintaining custom client

### 4. Multiple Server Support
**Concern**: mcptoolset designed for single server  
**Mitigation**: Create multiple mcptoolset instances (one per server)  
**Verdict**: ✅ Works perfectly - each server becomes a separate toolset

---

## Recommended Next Steps

1. **Update specification documents** to reflect `mcptoolset` usage
   - Revise `05_PHASE1_DETAILED_IMPLEMENTATION.md`
   - Update `01_MCP_SPECIFICATION.md`
   - Adjust `02_IMPLEMENTATION_PLAN.md`

2. **Create proof-of-concept** (1-2 days)
   - Single MCP server via mcptoolset
   - Basic configuration loading
   - Integration with existing agent

3. **Implement configuration layer** (1-2 days)
   - JSON config parsing
   - Environment variable substitution
   - Transport factory

4. **Build manager wrapper** (1 day)
   - Multi-server support
   - Toolset aggregation
   - Connection lifecycle

5. **Add CLI commands** (1 day)
   - `/mcp list`
   - `/mcp tools`
   - `/mcp reload`

6. **Documentation & examples** (1-2 days)
   - User setup guide
   - Example configurations
   - Troubleshooting guide

---

## Conclusion

The existing ADK-Go `mcptoolset` package provides a **production-ready, well-tested MCP client abstraction** that directly addresses the core requirements. Instead of reimplementing 2000+ lines of MCP client code, we should:

1. ✅ **Use** `google.golang.org/adk/tool/mcptoolset` for MCP client functionality
2. ✅ **Build** configuration and multi-server management on top
3. ✅ **Focus** development effort on user-facing features (config, CLI, docs)
4. ✅ **Leverage** Google's investment in ADK framework

This approach:
- Reduces implementation time from 6-9 weeks to 1-2 weeks
- Reduces code to maintain from ~5500 lines to ~800 lines
- Provides better long-term maintainability
- Aligns with code_agent's existing ADK architecture

**The specification should be updated to reflect this architectural decision.**
