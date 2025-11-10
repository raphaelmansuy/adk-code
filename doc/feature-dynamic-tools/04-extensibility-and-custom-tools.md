# Feature Comparison: Extensibility and Custom Tools

## Overview
This document explores how both systems allow developers to extend capabilities with custom tools.

---

## Code Agent: Tool Extension via ADK

### Tool Creation Pattern

Code Agent uses Google's ADK framework for tool creation, requiring developers to write Go code.

**Complete Tool Definition Example**:

```go
package tools

import (
    "fmt"
    "google.golang.org/adk/tool"
    "google.golang.org/adk/tool/functiontool"
)

// 1. Define Input struct with JSON schema tags
type MyToolInput struct {
    Param1 string `json:"param1" jsonschema:"description"`
    Param2 *int   `json:"param2,omitempty" jsonschema:"optional parameter"`
}

// 2. Define Output struct
type MyToolOutput struct {
    Result string `json:"result"`
    Error  string `json:"error,omitempty"`
}

// 3. Create the tool
func NewMyTool() (tool.Tool, error) {
    // Handler function
    handler := func(ctx tool.Context, input MyToolInput) MyToolOutput {
        // Implementation
        result := processData(input.Param1)
        return MyToolOutput{Result: result}
    }
    
    // Register with ADK
    t, err := functiontool.New(functiontool.Config{
        Name:        "my_tool",
        Description: "Does something useful",
    }, handler)
    
    if err == nil {
        // 4. Self-register with metadata
        Register(ToolMetadata{
            Tool:      t,
            Category:  CategoryCustom,
            Priority:  0,
            UsageHint: "Useful for X, Y, Z",
        })
    }
    
    return t, err
}
```

### Tool Registration

Tools register themselves via the global registry:

```go
// In tools/registry.go
var registry *ToolRegistry

func Register(metadata ToolMetadata) {
    registry.Add(metadata)
}

func GetRegistry() *ToolRegistry {
    return registry
}
```

### Integration into Agent

Tools are added to the agent in `coding_agent.go`:

```go
func NewCodingAgent(ctx context.Context, cfg Config) (agentiface.Agent, error) {
    // Create custom tool
    if _, err := tools.NewMyTool(); err != nil {
        return nil, err
    }
    
    // Get all tools from registry
    registry := tools.GetRegistry()
    registeredTools := registry.GetAllTools()
    
    // Create agent with tools
    codingAgent, err := llmagent.New(llmagent.Config{
        Tools: registeredTools,
    })
    
    return codingAgent, err
}
```

### Tool Categorization

Tools are organized by category for dynamic prompt generation:

```go
const (
    CategoryFileOperations ToolCategory = "File Operations"
    CategorySearchDiscovery ToolCategory = "Search & Discovery"
    CategoryCodeEditing    ToolCategory = "Code Editing"
    CategoryExecution      ToolCategory = "Execution"
    CategoryCustom         ToolCategory = "Custom Tools"
)
```

### Advanced Patterns

#### 1. Dry-Run Mode
```go
type MyToolInput struct {
    Param1 string
    DryRun bool `json:"dry_run,omitempty"`
}

handler := func(ctx tool.Context, input MyToolInput) MyToolOutput {
    if input.DryRun {
        // Preview without side effects
        return MyToolOutput{
            Result: "Would do X, Y, Z",
        }
    }
    // Execute for real
}
```

#### 2. Context-Aware Execution
```go
// Tool Context provides environment info
handler := func(ctx tool.Context, input MyToolInput) MyToolOutput {
    // Access execution context
    requestID := ctx.RequestID
    userID := ctx.UserID
    metadata := ctx.Metadata
}
```

#### 3. Error Handling
```go
type MyToolOutput struct {
    Success bool   `json:"success"`
    Result  string `json:"result,omitempty"`
    Error   string `json:"error,omitempty"`
}

handler := func(ctx tool.Context, input MyToolInput) MyToolOutput {
    if err != nil {
        return MyToolOutput{
            Success: false,
            Error:   fmt.Sprintf("Operation failed: %v", err),
        }
    }
    return MyToolOutput{
        Success: true,
        Result:  result,
    }
}
```

### Dynamic Prompt Generation

Tools automatically appear in system prompt:

```go
// From dynamic_prompt.go
func BuildToolsSection(registry *ToolRegistry) string {
    var builder strings.Builder
    builder.WriteString("## Available Tools\n\n")
    
    for _, category := range registry.GetCategories() {
        builder.WriteString(fmt.Sprintf("### %s\n\n", category))
        
        for _, metadata := range registry.GetByCategory(category) {
            tool := metadata.Tool
            builder.WriteString(fmt.Sprintf("**%s** - %s\n", 
                tool.Name(), tool.Description()))
            
            if metadata.UsageHint != "" {
                builder.WriteString(fmt.Sprintf("  → Usage: %s\n", 
                    metadata.UsageHint))
            }
            builder.WriteString("\n")
        }
    }
    
    return builder.String()
}
```

### Workflow: Adding a New Tool

1. **Create tool file**: `tools/my_tool.go`
2. **Define Input/Output structs**
3. **Implement handler function**
4. **Create registration function**: `NewMyTool()`
5. **Test tool directly**: Unit tests
6. **Add to `coding_agent.go`**: Tool initialization
7. **Rebuild agent**: `make build`
8. **Test end-to-end**: Manual testing

---

## Cline: Tool Extension via Model Context Protocol (MCP)

### MCP Protocol Overview

Model Context Protocol is an open protocol for connecting LLMs to tools and resources.

**Architecture**:
```
┌──────────────────────┐
│  Cline Extension     │
│  (Host/Client)       │
└──────────────────────┘
          ↓
    MCP Protocol
          ↓
┌──────────────────────┐
│  MCP Server          │
│  (Tool Provider)     │
└──────────────────────┘
```

### MCP Server Creation

MCP servers can be written in any language:

```typescript
// example-mcp-server/index.ts
import { 
    Server,
    Tool,
    TextContent,
    CallToolRequest,
} from "@modelcontextprotocol/sdk/server/index.js"
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js"

const server = new Server({
    name: "example-server",
    version: "1.0.0",
})

// Define tools
const tools: Tool[] = [
    {
        name: "my_tool",
        description: "Does something useful",
        inputSchema: {
            type: "object",
            properties: {
                param1: {
                    type: "string",
                    description: "First parameter",
                },
                param2: {
                    type: "number",
                    description: "Optional second parameter",
                },
            },
            required: ["param1"],
        },
    },
]

server.setRequestHandler(Tool.ListRequest, async () => ({
    tools: tools,
}))

// Handle tool calls
server.setRequestHandler(Tool.CallRequest, async (request: CallToolRequest) => {
    const { name, arguments: args } = request
    
    if (name === "my_tool") {
        const result = await processData(args.param1, args.param2)
        return {
            content: [{ type: "text", text: JSON.stringify(result) }],
        }
    }
    
    throw new Error(`Unknown tool: ${name}`)
})

const transport = new StdioServerTransport()
await server.connect(transport)
```

### MCP Server Configuration

Servers are configured in Cline settings:

```json
{
  "mcpServers": {
    "my-server": {
      "command": "node",
      "args": ["./dist/index.js"],
      "env": {
        "API_KEY": "${API_KEY}"
      },
      "disabled": false
    },
    "jira-tools": {
      "command": "python",
      "args": ["-m", "jira_mcp_server"],
      "env": {
        "JIRA_URL": "${JIRA_URL}",
        "JIRA_TOKEN": "${JIRA_TOKEN}"
      }
    }
  }
}
```

### MCP Hub Integration

Cline's `McpHub` manages tool discovery:

```typescript
class McpHub {
    // Load MCP servers from configuration
    async initializeMcpServers() {
        for (const [name, config] of Object.entries(mcpServers)) {
            const client = await this.createClient(config)
            const tools = await client.listTools()
            this.connections.push({
                serverName: name,
                client: client,
                tools: tools,
            })
        }
    }
    
    // Call tool from any server
    async callTool(
        serverName: string,
        toolName: string,
        args: Record<string, unknown>,
    ) {
        const connection = this.connections.find(c => c.serverName === serverName)
        return await connection.client.callTool({
            name: toolName,
            arguments: args,
        })
    }
    
    // Get all available tools
    getTools(): Array<{serverName: string; tool: Tool}> {
        const allTools = []
        for (const connection of this.connections) {
            for (const tool of connection.tools) {
                allTools.push({
                    serverName: connection.serverName,
                    tool: tool,
                })
            }
        }
        return allTools
    }
}
```

### Workflow: Adding an MCP Tool

1. **Create MCP server** (any language)
2. **Define tools** with JSON schema inputs
3. **Implement handlers** for tool calls
4. **Add to MCP config** (settings.json)
5. **Restart Cline** or reload MCP servers
6. **Use tool** - automatically available in chat

### MCP Transport Options

| Transport | Use Case | Configuration |
|-----------|----------|---|
| **stdio** | Local binaries, same machine | Command + args |
| **SSE** | HTTP servers | URL endpoint |
| **HTTP** | REST APIs | URL + auth |
| **WebSocket** | Real-time connections | URL + protocol |

### Example: Community MCP Servers

Pre-built MCP servers available:

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_PERSONAL_ACCESS_TOKEN": "${GITHUB_TOKEN}"
      }
    },
    "aws": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-aws-resources"],
      "env": {
        "AWS_PROFILE": "default"
      }
    }
  }
}
```

---

## Comparative Analysis

### Tool Development

| Aspect | Code Agent | Cline |
|--------|-----------|-------|
| **Language** | Go only | Any language (via MCP) |
| **Framework** | ADK required | MCP protocol (open) |
| **Setup** | Code + rebuild | Config file |
| **Discovery** | Static (build-time) | Dynamic (runtime) |
| **Communication** | In-process (fast) | Over protocol (flexible) |
| **Deployment** | Single binary | Distributed |
| **Isolation** | Minimal | Process isolation |

### Code Complexity

**Code Agent Tool** (simple):
```go
// ~50 lines including Input/Output structs
// Tightly integrated
// Type-safe
// Requires Go knowledge
```

**MCP Server** (simple):
```typescript
// ~80 lines for basic server
// Can be any language
// Uses JSON schema
// More flexible
```

### Execution Model

| Feature | Code Agent | Cline |
|---------|-----------|-------|
| Tool registration | Build-time | Runtime |
| Tool discovery | Agent constructor | MCP protocol |
| Error handling | Structured output | JSON response |
| Timeout | Per-tool config | MCP timeout |
| Logging | stdout/stderr | Server logs |
| Debugging | Direct code inspection | Server logs + protocol |

### Scalability

**Code Agent**:
- Add tools → Rebuild binary
- Scales vertically within single process
- All tools in same memory space
- Fast inter-tool communication

**Cline**:
- Add servers → Reload config
- Scales horizontally (separate processes)
- Process isolation
- Network overhead
- Can disable individual servers

---

## Best Practices

### Code Agent Custom Tools

1. **Follow the pattern**: Input struct, Output struct, handler, registration
2. **Add usage hints**: Help agent understand when to use tool
3. **Implement dry_run**: Preview without side effects
4. **Handle errors gracefully**: Return errors in Output struct
5. **Test independently**: Unit test handler function directly
6. **Document parameters**: JSON schema tags are visible to agent
7. **Return structured data**: Avoid ambiguous output formats

### MCP Server Development

1. **Use open protocol**: Ensures compatibility
2. **Implement error handling**: Proper MCP error responses
3. **Add timeout handling**: Connection timeouts, call timeouts
4. **Version API**: Backward compatible updates
5. **Document schema**: Clear input/output JSON schema
6. **Support hot reload**: Allow server restart in config
7. **Provide examples**: Usage examples in documentation

---

## Advanced Scenarios

### Code Agent: Multi-Tool Integration

```go
// Create specialized tool that uses other tools
func NewAggregatorTool() (tool.Tool, error) {
    handler := func(ctx tool.Context, input AggregatorInput) AggregatorOutput {
        // Use other tools internally
        readResult := readFile(input.FilePath)
        searchResult := grepSearch(input.FilePath, input.Pattern)
        return combine(readResult, searchResult)
    }
    // ...
}
```

### Cline: Tool Composition

```typescript
// MCP server that composes multiple APIs
// Example: Jira + GitHub tool that creates issue and links PR
const tools: Tool[] = [
    {
        name: "create_linked_issue",
        description: "Create Jira issue linked to GitHub PR",
        // Uses both Jira and GitHub APIs internally
    },
]
```

### Code Agent: Workspace-Aware Tools

```go
// Tool aware of workspace context
func NewProjectAwareTool() (tool.Tool, error) {
    handler := func(ctx tool.Context, input ProjectInput) ProjectOutput {
        // Access workspace manager
        wsManager := getWorkspaceManager()
        primaryRoot := wsManager.GetPrimaryRoot()
        // Tool operates on workspace
    }
}
```

### Cline: Context-Aware Tools

```typescript
// MCP tool that uses file context
{
    name: "analyze_code",
    description: "Analyze code with context",
    // Tool receives file content via @file context
    // Can operate on already-loaded context
}
```

---

## Migration Scenarios

### Migrate Code Agent Tool → MCP

```
Original (ADK):  Go tool struct → Rebuild
Equivalent (MCP): MCP server → Config reload

Both capable, different philosophies
```

### Migrate MCP → Code Agent

```
Original (MCP): Separate server process
Equivalent:     Go function in ADK tool

Better for: Monolithic deployment
```

---

## Conclusion

**Code Agent** offers compile-time extensibility with Go, suitable for closed, self-contained deployments with security requirements.

**Cline** offers runtime extensibility with MCP, ideal for open ecosystems and polyglot tool development.

**Choose Code Agent** if:
- Building closed systems
- Need maximum performance
- Want type safety
- Have Go expertise
- Monolithic deployment preferred

**Choose Cline (MCP)** if:
- Ecosystem compatibility important
- Multiple teams, different languages
- Dynamic tool loading needed
- Process isolation desired
- Distribute tool development

---

## See Also

- [01-architecture-and-framework.md](./01-architecture-and-framework.md) - Framework comparison
- [06-context-management.md](./06-context-management.md) - Tool context and state
- [07-deployment.md](./07-deployment.md) - Deployment scenarios
