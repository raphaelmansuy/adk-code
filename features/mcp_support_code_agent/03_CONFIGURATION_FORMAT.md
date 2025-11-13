# MCP Configuration Format Guide

**Architecture Note**: Configuration is used with the mcptoolset-based manager from Phase 1. See `05_PHASE1_DETAILED_IMPLEMENTATION.md` for implementation details.

## Quick Start

### Simplest Configuration (1 Stdio Server)

File: `~/.code_agent/config.json`
```json
{
  "mcp": {
    "servers": {
      "filesystem": {
        "type": "stdio",
        "command": "mcp-server-filesystem"
      }
    }
  }
}
```

Then run:
```bash
code-agent --config ~/.code_agent/config.json
```

---

## Full Configuration Reference

### Root Structure

```json
{
  "mcp": {
    "enabled": true,
    "servers": { ... },
    "globalSettings": { ... }
  }
}
```

### `mcp.enabled` (Boolean, Optional)

Default: `true` if `servers` are defined, otherwise `false`

Allows disabling all MCP without removing server definitions:
```json
{
  "mcp": {
    "enabled": false,
    "servers": { ... }
  }
}
```

---

## Server Definitions

### General Structure

```json
{
  "mcp": {
    "servers": {
      "<server_name>": {
        "type": "stdio|sse|http",
        "timeout": 30000,
        "trust": false,
        "includeTools": [...],
        "excludeTools": [...],
        "debug": false,
        
        // Type-specific fields
        "command": "...",    // for stdio
        "args": [...],       // for stdio
        "env": {...},        // for stdio
        "cwd": "...",        // for stdio
        
        "url": "...",        // for sse
        
        "httpUrl": "...",    // for http
        
        "headers": {...}     // for sse/http
      }
    }
  }
}
```

### Server Types

#### 1. **Stdio** - Local Command/Process

The server runs as a child process. Communication via stdin/stdout.

**Required Fields**:
- `type`: `"stdio"`
- `command`: Path to executable (absolute or in PATH)

**Optional Fields**:
- `args`: Array of command-line arguments
- `env`: Object of environment variables to set
- `cwd`: Working directory

**Examples**:

Simple command:
```json
{
  "filesystem": {
    "type": "stdio",
    "command": "mcp-server-filesystem"
  }
}
```

With arguments:
```json
{
  "filesystem": {
    "type": "stdio",
    "command": "mcp-server-filesystem",
    "args": ["/home/user/documents"]
  }
}
```

With custom environment:
```json
{
  "github": {
    "type": "stdio",
    "command": "./servers/github-mcp",
    "args": ["--token", "${GITHUB_TOKEN}"],
    "env": {
      "GITHUB_TOKEN": "${GITHUB_TOKEN}",
      "LOG_LEVEL": "debug"
    }
  }
}
```

With working directory:
```json
{
  "workspace": {
    "type": "stdio",
    "command": "node",
    "args": ["./mcp-server.js"],
    "cwd": "/path/to/mcp-servers"
  }
}
```

#### 2. **SSE** - Server-Sent Events

HTTP-based server using Server-Sent Events for streaming responses.

**Required Fields**:
- `type`: `"sse"`
- `url`: Server endpoint URL (HTTP or HTTPS)

**Optional Fields**:
- `headers`: Custom HTTP headers
- `timeout`: Request timeout in milliseconds

**Examples**:

Simple SSE server:
```json
{
  "web-search": {
    "type": "sse",
    "url": "https://api.example.com/mcp"
  }
}
```

With authentication header:
```json
{
  "github": {
    "type": "sse",
    "url": "https://mcp.github.example.com/search",
    "headers": {
      "Authorization": "Bearer ${GITHUB_TOKEN}"
    }
  }
}
```

With custom timeout:
```json
{
  "slow-api": {
    "type": "sse",
    "url": "https://api.example.com/mcp",
    "timeout": 60000
  }
}
```

#### 3. **HTTP** - Streamable HTTP

Bi-directional HTTP communication using streamable HTTP protocol.

**Required Fields**:
- `type`: `"http"`
- `httpUrl`: Server endpoint URL

**Optional Fields**:
- `headers`: Custom HTTP headers
- `timeout`: Request timeout in milliseconds

**Examples**:

Local HTTP server:
```json
{
  "local-mcp": {
    "type": "http",
    "httpUrl": "http://localhost:8000/mcp"
  }
}
```

Remote HTTP server with auth:
```json
{
  "enterprise-mcp": {
    "type": "http",
    "httpUrl": "https://mcp.company.internal/api",
    "headers": {
      "Authorization": "Bearer ${ENTERPRISE_TOKEN}",
      "X-API-Key": "${MCP_API_KEY}"
    }
  }
}
```

---

## Tool Filtering

### Include Only Whitelist (if specified)

When `includeTools` is defined, **only these tools are available**:

```json
{
  "filesystem": {
    "type": "stdio",
    "command": "mcp-server-filesystem",
    "includeTools": [
      "read_file",
      "write_file",
      "list_files"
    ]
  }
}
```

Result: Only these 3 tools from the filesystem server are registered. Others are ignored.

Pattern matching supported:
```json
{
  "includeTools": [
    "read_file",
    "write_file",
    "search*"  // Matches: search, search_code, search_web, etc.
  ]
}
```

### Exclude from Blacklist (if specified)

When `excludeTools` is defined, **these tools are hidden**:

```json
{
  "filesystem": {
    "type": "stdio",
    "command": "mcp-server-filesystem",
    "excludeTools": [
      "delete_recursive",
      "format_disk",
      "mount_device"
    ]
  }
}
```

Result: All filesystem tools are available except the 3 listed.

### Default Behavior

If neither `includeTools` nor `excludeTools` is specified, **all tools are available**:

```json
{
  "filesystem": {
    "type": "stdio",
    "command": "mcp-server-filesystem"
    // All tools from this server are available
  }
}
```

---

## Common Settings

### Timeout

Timeout for MCP server requests (milliseconds):

```json
{
  "slow-server": {
    "type": "sse",
    "url": "https://api.example.com/mcp",
    "timeout": 60000  // 60 seconds
  }
}
```

Default: 30000 (30 seconds)

### Trust

Skip tool execution confirmation for trusted servers:

```json
{
  "filesystem": {
    "type": "stdio",
    "command": "mcp-server-filesystem",
    "trust": true  // Don't ask for confirmation before running tools
  }
}
```

Default: `false` (ask for confirmation)

### Debug

Enable detailed logging for this server:

```json
{
  "filesystem": {
    "type": "stdio",
    "command": "mcp-server-filesystem",
    "debug": true  // Log all MCP messages
  }
}
```

Default: `false`

---

## Global Settings

### globalSettings Structure

Settings that apply to all MCP servers unless overridden:

```json
{
  "mcp": {
    "globalSettings": {
      "timeout": 30000,
      "debug": false,
      "enableOAuth": false
    }
  }
}
```

### Available Global Settings

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `timeout` | number | 30000 | Default timeout for all servers (ms) |
| `debug` | boolean | false | Enable debug logging |
| `enableOAuth` | boolean | false | Enable OAuth support (future) |

---

## Environment Variable Substitution

### Variable Syntax

Use `${VAR_NAME}` to reference environment variables:

```json
{
  "github": {
    "type": "sse",
    "url": "https://api.example.com/mcp",
    "headers": {
      "Authorization": "Bearer ${GITHUB_TOKEN}"
    }
  }
}
```

If `GITHUB_TOKEN=abc123`, this becomes:
```
Authorization: Bearer abc123
```

### Multiple Variables

```json
{
  "headers": {
    "Authorization": "Bearer ${API_TOKEN}",
    "X-User-ID": "${USER_ID}",
    "User-Agent": "code-agent/1.0 (${OS})"
  }
}
```

### Missing Variables

If a variable is not set:
- **During validation**: Error with clear message: `"Environment variable 'GITHUB_TOKEN' not found"`
- **Resolution**: Set the variable before running code-agent

```bash
export GITHUB_TOKEN=your-token
code-agent --config config.json
```

---

## Complete Examples

### Example 1: Multiple Local Servers

```json
{
  "mcp": {
    "servers": {
      "filesystem": {
        "type": "stdio",
        "command": "mcp-server-filesystem",
        "includeTools": ["read_file", "write_file", "list_files"]
      },
      "git": {
        "type": "stdio",
        "command": "./mcp-servers/git-server",
        "cwd": "/home/user/projects"
      }
    }
  }
}
```

### Example 2: Remote Servers with Auth

```json
{
  "mcp": {
    "servers": {
      "github": {
        "type": "sse",
        "url": "https://mcp.github.example.com/search",
        "headers": {
          "Authorization": "Bearer ${GITHUB_TOKEN}"
        },
        "timeout": 30000
      },
      "jira": {
        "type": "http",
        "httpUrl": "https://jira.company.com/mcp",
        "headers": {
          "Authorization": "Basic ${JIRA_CREDENTIALS}",
          "X-API-Version": "2"
        }
      }
    },
    "globalSettings": {
      "timeout": 45000,
      "debug": false
    }
  }
}
```

### Example 3: Mixed Setup with Filtering

```json
{
  "mcp": {
    "servers": {
      "filesystem": {
        "type": "stdio",
        "command": "mcp-server-filesystem",
        "excludeTools": ["delete_recursive"],
        "trust": true
      },
      "web": {
        "type": "sse",
        "url": "https://api.example.com/web",
        "headers": {
          "Authorization": "Bearer ${WEB_API_KEY}"
        },
        "includeTools": ["search", "fetch"]
      },
      "local": {
        "type": "http",
        "httpUrl": "http://localhost:3000/mcp",
        "debug": true
      }
    }
  }
}
```

---

## Configuration Validation

### Check Configuration Without Running

```bash
code-agent --validate-config ~/.code_agent/config.json
```

Output for valid config:
```
✓ Configuration is valid
  - Servers: 3
  - Total tools: 24 (estimated)
  - Server status: OK
```

Output for invalid config:
```
✗ Configuration validation failed

Error in server 'github':
  Missing required field: 'url' (type: sse)
  Suggestion: Add "url": "https://..." to server definition

Error in server 'filesystem':
  Environment variable '${GITHUB_TOKEN}' not found
  Suggestion: Set GITHUB_TOKEN environment variable before running
```

---

## Migration from Existing Setups

### From Environment Variable

If using:
```bash
export CODE_AGENT_MCP_SERVERS='[{"name":"fs","command":"server"}]'
```

Convert to:
```json
{
  "mcp": {
    "servers": {
      "fs": {
        "type": "stdio",
        "command": "server"
      }
    }
  }
}
```

### From Other Tools

If migrating from Cline or similar:
1. Extract MCP server definitions
2. Convert to `code_agent` format
3. Test with `--validate-config` flag
4. Start `code-agent` with config

---

## Troubleshooting

### "Server not found" Error

**Problem**: 
```
Error: failed to start server 'filesystem': command not found: mcp-server-filesystem
```

**Solution**:
1. Ensure the command is in PATH: `which mcp-server-filesystem`
2. Or use absolute path: `"/usr/local/bin/mcp-server-filesystem"`
3. Or use relative path with cwd: `{"command": "./servers/filesystem"}`

### Connection Timeout

**Problem**:
```
Error: connection to 'github' timed out after 30000ms
```

**Solution**:
1. Increase timeout: `"timeout": 60000`
2. Check URL is correct
3. Check network/firewall
4. Verify server is running

### Environment Variables Not Substituted

**Problem**:
```
Headers show literally: "Bearer ${GITHUB_TOKEN}"
```

**Solution**:
1. Variable must be set before running: `export GITHUB_TOKEN=...`
2. Variable must be in format `${NAME}` (no spaces)
3. Verify with: `echo $GITHUB_TOKEN`
