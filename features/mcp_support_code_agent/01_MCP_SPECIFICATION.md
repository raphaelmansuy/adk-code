# MCP Support Specification for code_agent

## Overview

This document specifies the implementation of Model Context Protocol (MCP) support in the code_agent CLI tool. The design leverages ADK-Go's production-ready `mcptoolset` abstraction for MCP protocol handling.

**Reference**: See `ARCHITECTURE_DECISION.md` for detailed rationale on using mcptoolset instead of custom MCP client implementation. Also see `logs/2025-11-13-09-50_mcp-adk-integration-research.md` for research findings.

## 1. Architecture Overview

### Design Pattern: Tool Aggregation via Toolsets

MCP integration uses ADK's **toolset aggregation** pattern:
- MCP servers are configured via JSON config
- Manager creates ADK `tool.Toolset` instances for each MCP server using `mcptoolset.New()`
- Each `toolset` automatically discovers and exposes MCP server's tools
- Agent combines native toolsets with MCP toolsets in single `llmagent.Config.Toolsets` array
- Tool execution routed through ADK's standard tool calling mechanism

### Architecture Flow

```
code_agent/main.go
  ↓
Load configuration
  ├─ Load app config (native tools)
  └─ Load MCP config (servers list)
  ↓
Create MCP Manager
  ├─ For each server config:
  │  ├─ Create transport (stdio/SSE/HTTP)
  │  └─ Create mcptoolset (ADK abstraction)
  └─ GetToolsets() returns all loaded toolsets
  ↓
Create Agent
  └─ llmagent.Config.Toolsets = [nativeTools, ...mcpToolsets]
  ↓
Agent.Run()
  └─ REPL accepts user requests
     └─ LLM calls tools from all toolsets (native + MCP)
        └─ ADK routes to appropriate tool handler
```

### Key Design Decisions

1. **Use ADK's mcptoolset**: Proven abstraction that handles MCP protocol details
2. **Lazy initialization**: Connections established only when needed (via mcptoolset internals)
3. **Graceful degradation**: MCP optional; agent works without configured servers
4. **Tool filtering**: Support include/exclude lists for fine-grained control
5. **Multi-transport**: Support stdio, SSE, and HTTP connections

## 2. Configuration Specification

### Configuration File Format (JSON)

**File Location**: `~/.code_agent/config.json` (default) or via `--mcp-config` flag or `CODE_AGENT_MCP_CONFIG_PATH` environment variable

**Structure**:
```json
{
  "mcp": {
    "enabled": true,
    "servers": {
      "<server_name>": {
        "type": "stdio|sse|http",
        "command": "...",           // for stdio only
        "args": [...],              // for stdio only
        "url": "...",               // for sse only
        "httpUrl": "...",           // for http only
        "headers": {...},           // optional, for sse/http
        "env": {...},               // optional, for stdio
        "cwd": "...",               // optional, for stdio
        "timeout": 30000,           // optional, milliseconds
        "trust": false,             // optional, skip confirmations
        "includeTools": [...],      // optional, whitelist
        "excludeTools": [...],      // optional, blacklist
        "debug": false              // optional
      }
    },
    "globalSettings": {
      "timeout": 30000,
      "debug": false
    }
  }
}
```

### Configuration via Environment Variable

For simple deployment, use `CODE_AGENT_MCP_SERVERS` environment variable:

```bash
export CODE_AGENT_MCP_SERVERS='[
  {"name": "filesystem", "type": "stdio", "command": "mcp-server-filesystem", "args": ["/"]},
  {"name": "web", "type": "sse", "url": "https://api.example.com/mcp"}
]'
```

### Configuration Priority

1. Environment variable `CODE_AGENT_MCP_SERVERS` (highest)
2. CLI flag `--config <file>` (if specified)
3. Default config file `~/.code_agent/config.json` (if exists)
4. No MCP servers (lowest)

### Server Type Details

**Type: `stdio`** - Local command/process
- `command`: Path to executable
- `args`: Command line arguments
- `env`: Environment variables to set
- `cwd`: Working directory
- Example: `{"type": "stdio", "command": "mcp-server-filesystem", "args": ["/home/user"]}`

**Type: `sse`** - Server-Sent Events over HTTP
- `url`: Server endpoint URL
- `headers`: HTTP headers (supports `${ENV_VAR}` substitution)
- Example: `{"type": "sse", "url": "https://api.example.com/mcp", "headers": {"Authorization": "Bearer ${API_TOKEN}"}}`

**Type: `http`** - Streamable HTTP
- `httpUrl`: Server endpoint URL
- `headers`: HTTP headers (supports `${ENV_VAR}` substitution)
- Example: `{"type": "http", "httpUrl": "http://localhost:8000/mcp"}`

### Tool Filtering

**Whitelist Mode** (if `includeTools` is specified):
```json
{
  "name": "filesystem",
  "includeTools": ["read_file", "write_file", "list_files"]
}
```
Only these tools are available.

**Blacklist Mode** (if `excludeTools` is specified):
```json
{
  "name": "filesystem",
  "excludeTools": ["delete_file", "format_disk"]
}
```
All tools except these are available.

**Default**: If neither is specified, all tools are available.

## 3. Tool Wrapper Specification

### Tool Naming Convention

MCP tools are registered with qualified names to avoid collisions:

```
<mcp_tool_name>@<server_name>
```

Examples:
- `read_file@filesystem`
- `list_repos@github`
- `search@web`

Alternative naming (configurable):
- Prefix: `mcp_<server>_<tool>`
- Namespace: `<server>::<tool>`

### Tool Attributes

Each MCP tool is wrapped with:
- **Name**: From MCP server (`read_file@filesystem`)
- **Description**: From MCP server + server name
- **Parameters**: From MCP InputSchema
- **Return Value**: Standardized to `map[string]any` or string

### Response Format Conversion

**Input**: MCP tool response (various formats: text, JSON, resources, images)

**Output**: Standardized format
```go
type ToolResult struct {
    Success bool        // true if no error
    Output  string      // human-readable output
    Data    map[string]any // structured data
    Error   string      // error message if failed
}
```

**Conversion Rules**:
- Text content → `Output` field
- Structured content → `Data` field
- Errors → `Success: false`, `Error: "<message>"`
- Images/resources → Reference in output with metadata

## 4. Lifecycle Management

### Initialization Phase

```
config.LoadFromEnv()
  ├─ Parse MCP servers config
  └─ Create MCPManager with config
  ↓
app.New()
  ├─ Validate MCP server configs
  ├─ Create MCP clients (don't connect yet)
  └─ Return application
  ↓
app.Run()
  ├─ Connect to MCP servers (or lazy on first use)
  ├─ Discover tools
  ├─ Register with ToolRegistry
  └─ Start REPL
```

### Server Connection Strategies

**Option A: Eager Connection** (current recommendation for Phase 1)
- Connect to all servers at startup
- Fail fast if config is wrong
- Know immediately if server is unavailable
- Simple error handling

**Option B: Lazy Connection** (future option)
- Connect on first tool discovery request
- Faster startup
- Better for optional MCP servers
- More complex error handling

### Server Lifecycle Events

1. **Discovery Phase**:
   - Establish connection to server
   - Fetch list of available tools
   - Filter tools based on config
   - Register with ToolRegistry

2. **Runtime Phase**:
   - Execute tool calls via MCP
   - Handle errors
   - Maintain connection

3. **Shutdown Phase**:
   - Gracefully close connections
   - Clean up resources

## 5. Error Handling

### Connection Errors

**Scenario**: Server unavailable or unreachable

**Behavior**:
- Log error with server name
- Continue loading other servers (don't fail entire startup)
- Show warning in CLI: "MCP server 'X' unavailable: ..."
- Provide `/mcp reconnect <server>` command

**Example Log**:
```
WARN: MCP server 'filesystem' connection failed: address already in use
INFO: Continuing with other MCP servers
```

### Tool Discovery Errors

**Scenario**: Tool list retrieval fails

**Behavior**:
- Log error with server and tool name
- Skip that specific tool
- Other tools from server still available
- Log count of successful discoveries

### Execution Errors

**Scenario**: Tool call fails

**Behavior**:
- Return error in tool result
- Log detailed error for debugging
- Don't crash the agent
- Show error message to user

### Configuration Errors

**Scenario**: Invalid config, missing required fields

**Behavior**:
- Validate config at startup
- Fail application with clear error message
- Suggest fixes: "Missing 'command' in stdio server 'X'"
- Support `--validate-config` flag to check without running

## 6. CLI Commands

### Internal REPL Commands

**New `/mcp` command group** for server management:

```
/mcp status                 # Show all MCP servers and status
/mcp list-tools             # List all MCP tools with server names
/mcp connect <server>       # Manually connect to a server
/mcp disconnect <server>    # Disconnect from a server
/mcp debug <server>         # Show detailed debug info
```

### Example Output

```
$ /mcp status

MCP Servers Status:
┌─────────────┬───────────┬─────────┬──────────┐
│ Server      │ Status    │ Tools   │ Updated  │
├─────────────┼───────────┼─────────┼──────────┤
│ filesystem  │ Connected │ 8       │ 2s ago   │
│ github      │ Connected │ 5       │ 10s ago  │
│ web-search  │ Failed    │ -       │ Error    │
└─────────────┴───────────┴─────────┴──────────┘

Use '/mcp list-tools' to see all tools
Use '/mcp debug <server>' for more details
```

## 7. Logging & Debugging

### Log Levels

- **INFO**: Server connections, tool discovery counts, status changes
- **WARN**: Configuration issues, non-critical failures
- **ERROR**: Critical failures, detailed error messages
- **DEBUG**: Detailed logging (enabled via `config.mcp.debug` or `-debug` flag)

### Debug Mode Output

When `debug: true` in config:
- MCP SDK detailed logs
- HTTP headers/response bodies
- Tool call parameters and results
- Connection timing and performance

### Environment Variable for Debug

```bash
CODE_AGENT_MCP_DEBUG=1 code-agent  # Enable MCP debug logging
```

## 8. Testing Strategy

### Unit Tests

**Config Parsing** (`internal/config/mcp_test.go`):
- Valid config parsing
- Invalid config detection
- Environment variable substitution
- Tool filtering logic

**Tool Wrapper** (`tools/mcp/tool_test.go`):
- Response format conversion
- Tool name generation
- Error handling

### Integration Tests

**With Mock MCP Server** (`tools/mcp/integration_test.go`):
- Connect to mock server
- Discover tools
- Execute tools
- Handle server errors

**Tool Execution** (`tools/mcp/tool_exec_test.go`):
- Run MCP tools
- Verify output format
- Test error scenarios

### Mock MCP Server

Provide a simple mock MCP server for testing:
```bash
make test-mcp-server  # Start mock server on :8888
```

Simple server that provides test tools:
- `echo` - Returns input text
- `fail` - Always returns error
- `slow` - Takes 5 seconds
- `json_response` - Returns structured JSON

## 9. Future Enhancements (Phase 2+)

### OAuth Support

- Auto-discover OAuth endpoints from `www-authenticate` headers
- Store tokens in `~/.code_agent/mcp-tokens/`
- Support `/mcp auth <server>` command
- Auto-refresh tokens

### Performance Optimization

- Connection pooling for multiple tools
- Tool result caching (with TTL)
- Parallel tool discovery for multiple servers
- Server health checks

### Advanced Configuration

- YAML config file support
- Configuration hot-reload (without restart)
- Per-tool timeout overrides
- Tool execution policies

### UI Enhancements

- Visual indicator for MCP tool sources
- Tool category grouping
- Server health dashboard
- Tool usage statistics

## 10. Success Criteria

### Phase 1 Completion Checklist

- [ ] Configuration file format implemented
- [ ] Environment variable parsing works
- [ ] MCP client connects to stdio servers
- [ ] Tool discovery from MCP servers works
- [ ] Tools registered in ToolRegistry
- [ ] Tool execution works for basic tools
- [ ] Error handling prevents crashes
- [ ] `/mcp list-tools` command works
- [ ] Tests pass (unit + integration)
- [ ] Documentation complete

### Testing Criteria

- [ ] Config validation catches invalid configs
- [ ] Connection failures don't crash app
- [ ] Tool name collisions are handled
- [ ] Response formatting works for text/JSON
- [ ] Tool filtering works (include/exclude)
- [ ] All existing tools still work
- [ ] MCP tools callable from agent

## 11. Implementation Checklist

### Code Files to Create

- [ ] `internal/config/mcp.go` - MCP configuration types and parsing
- [ ] `internal/config/mcp_test.go` - Config tests
- [ ] `pkg/mcp/client.go` - Single server client wrapper
- [ ] `pkg/mcp/manager.go` - Multi-server manager
- [ ] `pkg/mcp/types.go` - Shared types
- [ ] `tools/mcp/tool.go` - MCP tool wrapper
- [ ] `tools/mcp/tool_test.go` - Tool tests
- [ ] `internal/cli/commands/mcp.go` - `/mcp` CLI commands
- [ ] `test/mcp_mock_server.go` - Mock server for testing

### Code Files to Modify

- [ ] `internal/config/config.go` - Add MCP config field
- [ ] `internal/app/app.go` - Initialize MCPManager
- [ ] `tools/registry.go` - Document MCP registration
- [ ] `main.go` - Load MCP servers at startup

### Documentation Files

- [ ] `docs/MCP_SETUP.md` - User guide for setting up MCP servers
- [ ] `docs/MCP_DEV.md` - Developer guide for MCP support
- [ ] Example config: `examples/config.mcp.json`
- [ ] Example servers: `examples/mcp-servers.md`
