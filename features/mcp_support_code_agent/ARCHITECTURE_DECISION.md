# Architecture Decision: MCP Integration via ADK-Go mcptoolset

**Date**: November 13, 2025  
**Decision**: Use ADK-Go's existing `mcptoolset` package instead of implementing custom MCP client  
**Status**: APPROVED  
**Impact**: Reduces implementation from 6-9 weeks to 1-2 weeks, ~4700 lines of code saved

---

## Context

The code_agent project already depends on ADK-Go via a local replace directive:
```go
replace google.golang.org/adk => ../research/adk-go
```

Research revealed that ADK-Go includes a production-ready MCP support package (`tool/mcptoolset`) that:
- ✅ Already imports `github.com/modelcontextprotocol/go-sdk v0.7.0`
- ✅ Provides all MCP client abstractions needed
- ✅ Follows ADK's standard `tool.Toolset` pattern
- ✅ Has comprehensive error handling and session management
- ✅ Includes unit tests and examples

The original implementation plan proposed custom MCP client code (~2000+ lines) that duplicates logic already present in ADK-Go's `mcptoolset`.

---

## Decision

**Use ADK-Go's `google.golang.org/adk/tool/mcptoolset` for all MCP protocol handling.**

This means:
- ✅ Leverage existing, tested abstractions
- ✅ Reduce custom code by ~1500+ lines
- ✅ Eliminate MCP SDK version management concerns
- ✅ Benefit from upstream improvements
- ✅ Stay consistent with ADK-based architecture

---

## Architecture Comparison

### Option A: Custom Implementation (Proposed, ~2000+ lines)

```
code_agent/
├── pkg/mcp/
│   ├── client.go         # MCP SDK wrapper
│   ├── manager.go        # Multi-server manager
│   ├── types.go          # Custom types
│   └── ...
├── tools/mcp/
│   ├── tool.go           # Tool conversion
│   ├── response.go       # Response transformation
│   └── ...
├── internal/config/
│   └── mcp.go            # Config parsing
└── internal/cli/commands/
    └── mcp.go            # CLI commands
```

**Pros**:
- Full control over implementation
- Custom optimizations possible

**Cons**:
- Reimplements existing abstractions
- Maintains separate MCP client logic
- Duplicates error handling, session management
- Must track MCP SDK changes independently

---

### Option B: Use mcptoolset (Recommended, ~800 lines)

```
code_agent/
├── pkg/mcp/
│   ├── manager.go        # Multi-server aggregator (NEW)
│   └── manager_test.go
├── internal/config/
│   ├── mcp.go            # Config parsing (NEW)
│   └── mcp_test.go
└── internal/cli/commands/
    └── mcp.go            # CLI commands (NEW)

// Imports:
import "google.golang.org/adk/tool/mcptoolset"
```

**Pros**:
- Proven, tested MCP client
- Aligns with existing ADK architecture
- Smaller codebase to maintain
- Benefits from ADK improvements
- Faster implementation (1-2 weeks vs 6-9 weeks)

**Cons**:
- Slight dependency on ADK roadmap
- Can't customize MCP client internals (rarely needed)

---

## Implementation Layers

### Layer 1: Configuration (Required)
```go
// internal/config/mcp.go - NEW (~200 lines)
type MCPConfig struct {
    Servers map[string]ServerConfig
}

type ServerConfig struct {
    Type    string              // "stdio", "sse", "http"
    Command string              // For stdio
    Args    []string
    URL     string              // For SSE/HTTP
    Headers map[string]string
    IncludeTools []string       // Filtering
    ExcludeTools []string
}
```

### Layer 2: Manager (Wrapper)
```go
// pkg/mcp/manager.go - NEW (~200 lines)
type Manager struct {
    toolsets map[string]tool.Toolset  // mcptoolset instances
}

func (m *Manager) LoadServers(cfg *config.MCPConfig) error {
    for name, serverCfg := range cfg.Servers {
        transport := createTransport(serverCfg)
        
        // Use ADK's mcptoolset
        toolset, err := mcptoolset.New(mcptoolset.Config{
            Transport:  transport,
            ToolFilter: createFilterPredicate(serverCfg),
        })
        
        m.toolsets[name] = toolset
    }
    return nil
}

func (m *Manager) GetToolsets() []tool.Toolset {
    result := make([]tool.Toolset, 0)
    for _, ts := range m.toolsets {
        result = append(result, ts)
    }
    return result
}
```

### Layer 3: Transport Factory
```go
// pkg/mcp/transport.go - NEW (~150 lines)
func createTransport(cfg ServerConfig) mcp.Transport {
    switch cfg.Type {
    case "stdio":
        return &mcp.CommandTransport{
            Command: exec.Command(cfg.Command, cfg.Args...),
        }
    case "sse":
        return &mcp.SSETransport{
            URL: cfg.URL,
            Headers: cfg.Headers,
        }
    case "http":
        return &mcp.HTTPTransport{
            URL: cfg.URL,
            Headers: cfg.Headers,
        }
    }
    return nil
}
```

### Layer 4: CLI Commands
```go
// internal/cli/commands/mcp.go - NEW (~150 lines)
func mcpListCommand(m *mcp.Manager) {
    toolsets := m.GetToolsets()
    for _, ts := range toolsets {
        tools, _ := ts.Tools(ctx)
        fmt.Printf("Server: %s (%d tools)\n", ts.Name(), len(tools))
        for _, tool := range tools {
            fmt.Printf("  - %s\n", tool.Name())
        }
    }
}

func mcpStatusCommand(m *mcp.Manager) {
    // Show connection status for each server
}
```

### Layer 5: Agent Integration
```go
// main.go or app.go
mcpCfg := config.LoadMCPConfig(configPath)

mcpManager := mcp.NewManager()
mcpManager.LoadServers(mcpCfg)

// Add MCP toolsets to agent
allToolsets := []tool.Toolset{
    nativeToolsets...,
    ...mcpManager.GetToolsets(),  // All MCP tools
}

agent := llmagent.New(llmagent.Config{
    Toolsets: allToolsets,
})
```

---

## What mcptoolset Provides

### From `research/adk-go/tool/mcptoolset/set.go`

**Configuration**:
```go
type Config struct {
    Transport  mcp.Transport   // Connection
    ToolFilter tool.Predicate  // Optional filtering
}
```

**API**:
```go
mcpToolSet, err := mcptoolset.New(config)
tools, err := mcpToolSet.Tools(ctx)  // Discover tools
tool, err := mcpToolSet.Tool(ctx, name)  // Get specific tool
```

**Session Management** (Already handles):
- ✅ Lazy initialization
- ✅ Thread-safe access via sync.Mutex
- ✅ Automatic reconnection
- ✅ Graceful shutdown

**Tool Execution** (Already handles):
- ✅ MCP protocol marshalling
- ✅ Structured/text output handling
- ✅ Error extraction
- ✅ Context propagation

**Examples**: See `research/adk-go/examples/mcp/main.go`

---

## Migration Path

### If Future MCP Features Needed

If ADK-Go's mcptoolset doesn't support a new MCP feature:

1. **Report Issue**: Open issue in ADK-Go repo
2. **Contribute**: Submit PR to ADK-Go if feature is general
3. **Workaround**: Create thin wrapper around MCP SDK if urgent
4. **Upgrade**: Integrate when ADK-Go is updated

This is much lower risk than maintaining custom client code.

---

## Benefits Summary

| Aspect | Custom | mcptoolset |
|--------|--------|-----------|
| Development Time | 6-9 weeks | 1-2 weeks |
| Code to Maintain | ~5500 lines | ~800 lines |
| MCP Client Logic | Custom (risky) | Proven, tested |
| Version Management | Manual | Automatic (via ADK) |
| Error Handling | DIY | Production-ready |
| Session Management | DIY | Tested & documented |
| Future MCP Support | Must implement | ADK handles |
| Consistency | Custom | ADK patterns |
| Documentation | Must write | ADK examples |

---

## What Still Needs Implementation

✅ **Configuration** (200 lines) - JSON/env config parsing  
✅ **Manager** (200 lines) - Multi-server aggregation  
✅ **Transport Factory** (150 lines) - Create stdio/SSE/HTTP transports  
✅ **CLI Commands** (150 lines) - /mcp list, /mcp status, etc  
✅ **Documentation** (Various) - User guide, examples  
✅ **Tests** (Various) - Config, transport factory tests

---

## Risks & Mitigations

### Risk 1: ADK-Go Dependency Change
**Risk**: ADK-Go removes or changes mcptoolset  
**Mitigation**: ADK is Google-maintained, actively used; low probability  
**Fallback**: Switch to custom implementation if needed (still possible)

### Risk 2: Limited Customization
**Risk**: mcptoolset doesn't support niche use case  
**Mitigation**: ToolFilter predicate handles most filtering needs  
**Fallback**: Wrap mcptoolset with custom logic if needed

### Risk 3: Version Lock
**Risk**: We're locked to whatever ADK version provides  
**Mitigation**: Local replacement allows quick ADK updates  
**Fallback**: Can fork ADK-Go if critical features needed

---

## Decision Record

**Decided**: November 13, 2025  
**By**: Architecture Review (based on research log 2025-11-13-09-50_mcp-adk-integration-research.md)  
**Rationale**: Leverage existing, proven abstractions in ADK-Go rather than duplicate 2000+ lines of MCP client code

**Implementation**: Begin Phase 1 using mcptoolset as foundation

**Review**: Revisit if new requirements emerge that mcptoolset can't support

---

## References

- Research log: `logs/2025-11-13-09-50_mcp-adk-integration-research.md`
- ADK-Go mcptoolset: `research/adk-go/tool/mcptoolset/`
- ADK-Go examples: `research/adk-go/examples/mcp/main.go`
- MCP SDK docs: https://github.com/modelcontextprotocol/go-sdk
