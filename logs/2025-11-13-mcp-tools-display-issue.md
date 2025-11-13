# MCP Tools Display Issue

## Current Behavior

When you run:
```bash
./code-agent --mcp-config code_agent/examples/mcp/sse-server.json
```

And then type `/tools`, you only see the built-in tools, not MCP tools.

## Why This Happens

### Phase 1 Implementation

In Phase 1, MCP toolsets are added to the agent correctly:

```go
// In agent.go
mcpToolsets := mcpManager.Toolsets()

// Passed to agent
codingAgent, err := llmagent.New(llmagent.Config{
    Tools:    registeredTools,  // Built-in tools
    Toolsets: cfg.MCPToolsets,   // MCP tools ✅
})
```

**The MCP tools ARE loaded and available to the agent** - they just aren't displayed in the UI.

### The Display Issue

The `/tools` command in `repl_builders.go` shows a **hardcoded list** of built-in tools:

```go
func buildToolsListLines(renderer *display.Renderer) []string {
    lines = append(lines, "   ✓ read_file - Read file contents")
    lines = append(lines, "   ✓ write_file - Create files")
    // ... hardcoded list
}
```

It doesn't query:
- The tool registry
- The MCP manager
- The agent's actual toolsets

## Verification That MCP Tools Work

Even though MCP tools don't show in `/tools`, they ARE available. You can verify by:

1. **Asking the agent to use them**: The agent will automatically use MCP tools when appropriate
2. **Checking startup**: No errors means MCP config loaded successfully
3. **Looking at agent behavior**: The agent has access to MCP tools alongside built-in tools

## Solution for Phase 2

Phase 2 should add proper MCP tool visibility:

### 1. Add `/mcp` Commands
```
/mcp list       - List all MCP servers
/mcp status     - Show server connection status  
/mcp tools      - List tools from each MCP server
```

### 2. Update `/tools` Command

Option A: Add MCP section to existing `/tools` display:
```
Available Tools
═══════════════

📝 Core Editing Tools:
   ✓ read_file - Read file contents
   ...

🔌 MCP Tools (deepwiki):
   ✓ search_wiki - Search Wikipedia
   ✓ get_article - Get article content
   ...
```

Option B: Keep `/tools` for built-in only, use `/mcp tools` for MCP tools

### 3. Implementation Approach

Update `buildToolsListLines()` to:
- Accept MCP manager as parameter
- Query MCP toolsets
- Display MCP tools by server

## Current Status

✅ **Phase 1 Complete**: MCP tools are loaded and functional  
⏳ **Phase 2 Needed**: UI visibility for MCP tools

## Workaround for Now

To verify MCP tools are loaded, you can:
1. Ask the agent a question that would use MCP tools
2. Check if the agent mentions or uses them
3. Or wait for Phase 2 `/mcp` commands

## Example

If deepwiki MCP server provides Wikipedia search:

```
❯ Search Wikipedia for "Quantum Computing"

[Agent will use deepwiki's search_wiki tool if it's available]
```

The agent will automatically use the MCP tool even though it's not listed in `/tools`.
