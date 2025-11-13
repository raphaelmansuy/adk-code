# MCP Configuration Examples

This directory contains example MCP (Model Context Protocol) server configurations for the adk-code.

## Quick Start

To use MCP servers with code_agent, pass the `--mcp-config` flag:

```bash
./code-agent --mcp-config examples/mcp/basic-stdio.json
```

## Configuration Files

### basic-stdio.json

Simple stdio-based MCP server configuration using the filesystem server.

```bash
./code-agent --mcp-config examples/mcp/basic-stdio.json
```

### sse-server.json

Server-Sent Events (SSE) transport configuration for HTTP-based MCP servers.

```bash
./code-agent --mcp-config examples/mcp/sse-server.json
```

### streamable-server.json

Streamable transport configuration for bidirectional HTTP communication.

```bash
./code-agent --mcp-config examples/mcp/streamable-server.json
```

### multi-server.json

Advanced configuration with multiple MCP servers of different types.

```bash
./code-agent --mcp-config examples/mcp/multi-server.json
```

## Configuration Formats

The adk-code supports **two configuration formats**: the native format and the Claude Desktop format.

### Native Format

```json
{
  "enabled": true,
  "servers": {
    "server-name": {
      "type": "stdio|sse|streamable",
      "command": "command-for-stdio",
      "args": ["arg1", "arg2"],
      "url": "url-for-sse-or-streamable",
      "headers": {"key": "value"},
      "env": {"ENV_VAR": "value"},
      "cwd": "/working/directory",
      "timeout": 30000
    }
  }
}
```

### Claude Desktop Format

The adk-code also supports the same configuration format used by Claude Desktop (from Anthropic). This makes it easy to share configurations between Claude Desktop and code_agent.

```json
{
  "mcpServers": {
    "Bright Data": {
      "command": "npx",
      "args": ["@brightdata/mcp"],
      "env": {
        "API_TOKEN": "your-token-here",
        "PRO_MODE": "true"
      }
    }
  }
}
```

**Claude Format Features:**

- No need to specify `"type"` - it's automatically inferred from the presence of `command` (stdio) or `url` (SSE)
- No need for `"enabled"` field - presence of `mcpServers` implies enabled
- Drop-in compatible with Claude Desktop's `claude_desktop_config.json`

**Example configurations:**

- `examples/mcp/claude-format.json` - Multiple stdio servers in Claude format
- `examples/mcp/claude-format-sse.json` - SSE server in Claude format

**Usage:**

```bash
# Use Claude Desktop format
./code-agent --mcp-config examples/mcp/claude-format.json

# Or use native format
./code-agent --mcp-config examples/mcp/basic-stdio.json
```

Both formats work identically - choose whichever you prefer!

## Claude Desktop Format Examples

### claude-format.json

Complete example using Claude Desktop's configuration format with multiple servers:

```bash
./code-agent --mcp-config examples/mcp/claude-format.json
```

This configuration includes:

- **Bright Data MCP Server** with environment variables
- **Filesystem Server** for file access
- **Homebrew MCP Server** for package management

### claude-format-sse.json

HTTP/SSE server example in Claude Desktop format:

```bash
./code-agent --mcp-config examples/mcp/claude-format-sse.json
```

## Transport Types

### stdio

Process-based transport that spawns a command and communicates via stdin/stdout.

**Required fields:**

- `command`: The command to execute
- `args`: (optional) Command arguments
- `env`: (optional) Environment variables
- `cwd`: (optional) Working directory

**Example:**

```json
{
  "type": "stdio",
  "command": "mcp-server-filesystem",
  "args": ["/path/to/files"]
}
```

### sse

HTTP Server-Sent Events transport for one-way server-to-client streaming.

**Required fields:**

- `url`: The SSE endpoint URL
- `headers`: (optional) HTTP headers
- `timeout`: (optional) Request timeout in milliseconds

**Example:**

```json
{
  "type": "sse",
  "url": "http://localhost:3000/sse",
  "headers": {
    "Authorization": "Bearer token"
  }
}
```

### streamable

HTTP transport with bidirectional streaming support.

**Required fields:**

- `url`: The server endpoint URL
- `headers`: (optional) HTTP headers
- `timeout`: (optional) Request timeout in milliseconds

**Example:**

```json
{
  "type": "streamable",
  "url": "http://localhost:3000/mcp",
  "timeout": 60000
}
```

## Available MCP Servers

### Official MCP Servers

1. **@modelcontextprotocol/server-filesystem**
   - File system access
   - Command: `npx -y @modelcontextprotocol/server-filesystem <path>`

2. **@modelcontextprotocol/server-github**
   - GitHub API integration
   - Command: `npx -y @modelcontextprotocol/server-github`
   - Requires: `GITHUB_PERSONAL_ACCESS_TOKEN` env var

3. **@modelcontextprotocol/server-postgres**
   - PostgreSQL database access
   - Command: `npx -y @modelcontextprotocol/server-postgres <connection-string>`

4. **@modelcontextprotocol/server-puppeteer**
   - Browser automation
   - Command: `npx -y @modelcontextprotocol/server-puppeteer`

### Environment Variable Substitution

You can use environment variables in configuration files using `${VAR_NAME}` syntax:

```json
{
  "env": {
    "API_KEY": "${MY_API_KEY}"
  }
}
```

Note: Environment variable substitution is not yet implemented in Phase 1. Variables will be read directly from the current environment.

## Testing Your Configuration

### Option 1: Run without MCP (Disabled)

To verify adk-code works without MCP servers:

```bash
./code-agent --mcp-config examples/mcp/disabled.json
```

This will start the agent normally with MCP disabled.

### Option 2: Test with a Real MCP Server

To test with actual MCP functionality, you need a real MCP server. For example, the official filesystem server:

```bash
# Install the MCP filesystem server (requires Node.js)
npm install -g @modelcontextprotocol/server-filesystem

# Create a config pointing to it
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

# Run adk-code with the MCP server
./code-agent --mcp-config /tmp/mcp-fs.json
```

### Important Notes

- **Don't use `echo` or other non-MCP commands**: MCP servers must implement the MCP protocol
- **Use `/mcp` commands to debug**: Use `/mcp status` and `/mcp list` to check server health
- **Server validation happens at startup**: Check terminal output for any MCP connection errors
- **deepwiki server is broken**: The example `sse-server.json` points to `https://mcp.deepwiki.com/sse` which returns invalid JSON. It's disabled by default. Don't enable it unless the server is fixed.

## Troubleshooting

### Error: "invalid character 'p' looking for beginning of value" (or similar)

This error occurs when the MCP server returns invalid JSON instead of proper MCP protocol responses. Common causes:

- **SSE/HTTP servers returning HTML error pages** instead of JSON
- External MCP servers that are down or misconfigured
- Using `echo` or other shell commands instead of an MCP server
- The command outputs text instead of JSON MCP protocol responses

**Examples of this error**:

- `invalid character 'p' looking for beginning of value` - Server returned HTML like `<html>`
- `invalid character 'h' looking for beginning of value` - Command returned plain text like "hello"

**Solution**:

1. Check `/mcp status` to see which server is failing
2. Test the server URL manually (for SSE/HTTP servers)
3. Use a real MCP server command for stdio servers
4. Set `"enabled": false` in your config to disable broken servers

### Server not connecting

- Check that the command exists and is executable
- Verify environment variables are set correctly
- Check server logs for errors

### stdio server fails to start

- Ensure the command is in your PATH
- Verify the working directory exists
- Check that required dependencies are installed

### HTTP server not reachable

- Confirm the server is running and reachable
- Verify the URL is correct
- Check firewall and network settings
- Verify authentication headers if required

## Next Steps

- Phase 2 will add dynamic server management (start/stop/reload)
- Phase 3 will add environment variable substitution
- See `../features/mcp_support_adk-code/README.md` for the full roadmap
