# MCP Testing Notes

## Issue Discovered: Invalid MCP Server Command

When testing Phase 1 MCP implementation, we discovered that using non-MCP commands (like `echo`) causes errors:

```
Error: invalid character 'h' looking for beginning of value
```

## Root Cause

The test configuration `test-simple.json` used `echo` as the MCP server command:

```json
{
  "enabled": true,
  "servers": {
    "test-server": {
      "type": "stdio",
      "command": "echo",
      "args": ["hello"]
    }
  }
}
```

`echo` just outputs "hello" and exits - it doesn't implement the MCP protocol. The agent expects JSON responses from MCP servers.

## Solution

### For Testing Without MCP Servers

Use the disabled configuration:

```bash
./code-agent --session test-session --mcp-config code_agent/examples/mcp/disabled.json
```

**Important**: Use `--session <unique-name>` to avoid session cache issues.

### For Testing With Real MCP Servers

Install and use a real MCP server:

```bash
# Install MCP filesystem server
npm install -g @modelcontextprotocol/server-filesystem

# Create config
cat > /tmp/mcp-fs.json << 'EOF'
{
  "enabled": true,
  "servers": {
    "filesystem": {
      "type": "stdio",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"]
    }
  }
}
EOF

# Run with new session
./code-agent --session mcp-test --mcp-config /tmp/mcp-fs.json
```

## Files Updated

1. **test-simple.json**: Changed to `"enabled": false` (not a real MCP server)
2. **disabled.json**: Created proper disabled config
3. **README.md**: Added testing section and troubleshooting

## Recommendation for Phase 2

Consider adding validation that checks if the MCP server responds correctly before adding it to the agent's toolset. This would provide better error messages at startup rather than during runtime.

## Session Cache Note

MCP configuration is stored in sessions. When testing different MCP configs, always use unique session names with `--session <name>` to avoid cached configurations.
