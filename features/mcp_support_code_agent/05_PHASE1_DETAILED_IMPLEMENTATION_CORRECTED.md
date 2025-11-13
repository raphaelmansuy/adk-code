# Phase 1: MCP Support Implementation - CORRECTED

> **STATUS**: Updated from original `05_PHASE1_DETAILED_IMPLEMENTATION.md` with verified fixes
>
> **Last Verified**: November 13, 2025  
> **Reference**: VERIFICATION_REPORT.md

## Overview

Add MCP server support to `code_agent` leveraging ADK-Go's built-in `mcptoolset` package.

- **Duration**: 1 week
- **Complexity**: Low (relies on existing ADK abstractions)
- **Outcome**: Agent discovers and executes tools from configured MCP servers

## Architecture

**Decision**: Use `google.golang.org/adk/tool/mcptoolset` instead of custom MCP client.

**Why**: 
- ADK-Go already provides production-tested MCP support
- Handles all protocol details, transports, and tool discovery
- MCP SDK updates managed upstream by ADK maintainers
- 90% less custom code to maintain

---

## Implementation Tasks

### Task 1: MCP Configuration (1 day)

**Deliverable**: `internal/config/mcp.go`

```go
package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type MCPConfig struct {
	Enabled bool                    `json:"enabled"`
	Servers map[string]ServerConfig `json:"servers"`
}

type ServerConfig struct {
	Name    string            `json:"-"` // Set from map key
	Type    string            `json:"type"` // "stdio", "sse", "streamable"
	Command string            `json:"command,omitempty"` // For stdio
	Args    []string          `json:"args,omitempty"`
	URL     string            `json:"url,omitempty"` // For sse/streamable
	Headers map[string]string `json:"headers,omitempty"`
	Env     map[string]string `json:"env,omitempty"` // Environment variables for stdio
	Cwd     string            `json:"cwd,omitempty"` // Working directory for stdio
	Timeout int               `json:"timeout,omitempty"` // milliseconds, default 30000
}

// LoadMCP loads MCP config from file or environment
func LoadMCP(configPath string) (*MCPConfig, error) {
	if configPath == "" {
		return &MCPConfig{Enabled: false}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &MCPConfig{Enabled: false}, nil
		}
		return nil, err
	}

	var cfg MCPConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Set server names from map keys
	for name, srv := range cfg.Servers {
		s := srv
		s.Name = name
		cfg.Servers[name] = s
	}

	// Validate
	for name, srv := range cfg.Servers {
		if err := srv.validate(); err != nil {
			return nil, fmt.Errorf("server '%s': %w", name, err)
		}
	}

	return &cfg, nil
}

func (s ServerConfig) validate() error {
	if s.Type == "" {
		return fmt.Errorf("type required")
	}
	switch s.Type {
	case "stdio":
		if s.Command == "" {
			return fmt.Errorf("command required for stdio")
		}
	case "sse", "streamable":
		if s.URL == "" {
			return fmt.Errorf("url required for %s", s.Type)
		}
	default:
		return fmt.Errorf("unsupported type: %s", s.Type)
	}
	return nil
}
```

**Tests**: `internal/config/mcp_test.go`

```go
package config

import (
	"os"
	"testing"
)

func TestLoadMCP(t *testing.T) {
	tests := []struct {
		name      string
		configJSON string
		wantErr   bool
	}{
		{
			name: "valid stdio server",
			configJSON: `{
				"enabled": true,
				"servers": {
					"fs": {"type": "stdio", "command": "echo"}
				}
			}`,
			wantErr: false,
		},
		{
			name: "missing command",
			configJSON: `{
				"servers": {"fs": {"type": "stdio"}}
			}`,
			wantErr: true,
		},
		{
			name: "valid sse",
			configJSON: `{
				"servers": {"web": {"type": "sse", "url": "http://localhost"}}
			}`,
			wantErr: false,
		},
		{
			name: "valid streamable",
			configJSON: `{
				"servers": {"api": {"type": "streamable", "url": "http://localhost:3000/mcp"}}
			}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := os.CreateTemp("", "*.json")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(f.Name())

			f.WriteString(tt.configJSON)
			f.Close()

			cfg, err := LoadMCP(f.Name())
			if (err != nil) != tt.wantErr {
				t.Fatalf("got error %v, want %v", err, tt.wantErr)
			}
			if err == nil && !cfg.Enabled && tt.configJSON != "" {
				// Basic sanity check
				if len(cfg.Servers) == 0 {
					t.Error("expected servers to be loaded")
				}
			}
		})
	}
}
```

**Config Update**: Modify `internal/config/config.go` to add MCP fields:

```go
type Config struct {
	// ... existing fields ...
	MCPConfigPath string
	MCPConfig     *MCPConfig
}

// In LoadFromEnv():
mcpPath := flag.String("mcp-config", "", "Path to MCP config file")
// ... after flag.Parse() ...
mcpCfg, _ := LoadMCP(*mcpPath) // Fail silently for optional feature
```

---

### Task 2: Manager Integration (2 days)

**Deliverable**: `internal/mcp/manager.go`

This wraps ADK-Go's `mcptoolset` and manages multiple servers.

```go
package mcp

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"code_agent/internal/config"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/mcptoolset"
)

type Manager struct {
	mu       sync.RWMutex
	toolsets []tool.Toolset // All loaded MCP toolsets
	servers  map[string]*server
}

type server struct {
	name    string
	toolset tool.Toolset
	err     error
}

func NewManager() *Manager {
	return &Manager{
		toolsets: make([]tool.Toolset, 0),
		servers:  make(map[string]*server),
	}
}

// LoadServers initializes all configured MCP servers
func (m *Manager) LoadServers(ctx context.Context, cfg *config.MCPConfig) error {
	if !cfg.Enabled || len(cfg.Servers) == 0 {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for name, srvCfg := range cfg.Servers {
		if err := m.loadServer(ctx, name, srvCfg); err != nil {
			// Log but don't fail entire load
			m.servers[name] = &server{name: name, err: err}
			continue
		}
	}

	return nil
}

func (m *Manager) loadServer(ctx context.Context, name string, cfg config.ServerConfig) error {
	transport, err := createTransport(cfg)
	if err != nil {
		return err
	}

	toolset, err := mcptoolset.New(mcptoolset.Config{
		Transport: transport,
	})
	if err != nil {
		return fmt.Errorf("failed to create toolset: %w", err)
	}

	m.servers[name] = &server{name: name, toolset: toolset}
	m.toolsets = append(m.toolsets, toolset)
	return nil
}

// Toolsets returns all loaded MCP toolsets for agent integration
func (m *Manager) Toolsets() []tool.Toolset {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]tool.Toolset{}, m.toolsets...)
}

// List returns server names and status
func (m *Manager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.servers))
	for name := range m.servers {
		names = append(names, name)
	}
	return names
}

func createTransport(cfg config.ServerConfig) (mcp.Transport, error) {
	switch cfg.Type {
	case "stdio":
		return createStdioTransport(cfg)
	case "sse":
		return createSSETransport(cfg)
	case "streamable":
		return createStreamableTransport(cfg)
	default:
		return nil, fmt.Errorf("unsupported transport: %s", cfg.Type)
	}
}

func createStdioTransport(cfg config.ServerConfig) (mcp.Transport, error) {
	if cfg.Command == "" {
		return nil, fmt.Errorf("command required for stdio transport")
	}

	cmd := exec.Command(cfg.Command, cfg.Args...)
	
	// Set environment variables if provided
	if cfg.Env != nil {
		cmd.Env = os.Environ() // Start with existing environment
		for k, v := range cfg.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}
	
	// Set working directory if provided
	if cfg.Cwd != "" {
		cmd.Dir = cfg.Cwd
	}

	return &mcp.CommandTransport{
		Command: cmd,
		TerminateDuration: 5 * time.Second, // Graceful termination timeout
	}, nil
}

func createSSETransport(cfg config.ServerConfig) (mcp.Transport, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("url required for sse transport")
	}

	headers := http.Header{}
	for k, v := range cfg.Headers {
		headers.Add(k, v)
	}

	timeout := 30 * time.Second
	if cfg.Timeout > 0 {
		timeout = time.Duration(cfg.Timeout) * time.Millisecond
	}

	return &mcp.SSEClientTransport{
		Endpoint: cfg.URL,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

func createStreamableTransport(cfg config.ServerConfig) (mcp.Transport, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("url required for streamable transport")
	}

	timeout := 30 * time.Second
	if cfg.Timeout > 0 {
		timeout = time.Duration(cfg.Timeout) * time.Millisecond
	}

	return &mcp.StreamableClientTransport{
		Endpoint: cfg.URL,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}
```

**Tests**: `internal/mcp/manager_test.go`

```go
package mcp

import (
	"context"
	"testing"

	"code_agent/internal/config"
)

func TestManagerEmpty(t *testing.T) {
	m := NewManager()
	cfg := &config.MCPConfig{Enabled: false}

	if err := m.LoadServers(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}

	if len(m.Toolsets()) != 0 {
		t.Error("expected no toolsets")
	}
}

func TestListServers(t *testing.T) {
	m := NewManager()
	cfg := &config.MCPConfig{Enabled: true}

	m.LoadServers(context.Background(), cfg)
	list := m.List()
	if len(list) != 0 {
		t.Error("expected empty list")
	}
}
```

---

### Task 3: Agent Integration (1 day)

**Deliverable**: Update `main.go` to include MCP toolsets

```go
package main

import (
	"context"
	"fmt"
	"os"

	"code_agent/internal/config"
	"code_agent/internal/mcp"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/tool"
)

func main() {
	ctx := context.Background()

	appCfg := config.LoadConfig()

	// Load MCP if configured
	var mcpToolsets []tool.Toolset
	if appCfg.MCPConfig != nil && appCfg.MCPConfig.Enabled {
		mgr := mcp.NewManager()
		if err := mgr.LoadServers(ctx, appCfg.MCPConfig); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: MCP load failed: %v\n", err)
		}
		mcpToolsets = mgr.Toolsets()
	}

	// Note: This assumes llmagent.Config supports Toolsets parameter
	// Verify: Use either Tools OR Toolsets, not both, based on actual llmagent API
	
	// ... rest of main ...
}
```

**IMPORTANT VERIFICATION NEEDED**: 
Check `llmagent.New()` Config struct to determine the correct pattern:
- Does it support `Toolsets` field?
- Should native tools be wrapped in a Toolset for consistency?
- Can both `Tools` and `Toolsets` be used together?

---

### Task 4: CLI Commands (1 day)

**Deliverable**: Add `/mcp` command to REPL

In `internal/repl/repl.go`, add:

```go
case "/mcp":
	args := strings.Fields(input)
	if len(args) < 2 {
		fmt.Println("Usage: /mcp list|tools")
		return
	}

	switch args[1] {
	case "list":
		servers := r.mcpManager.List()
		fmt.Printf("MCP Servers: %v\n", servers)
	case "tools":
		// TODO: Implement tool listing per server
		fmt.Println("TODO: /mcp tools implementation")
	default:
		fmt.Println("Unknown subcommand. Use: list, tools")
	}
```

---

### Task 5: Documentation (0.5 day)

**Deliverable**: Example configs and usage guide

**File**: `code_agent/examples/mcp.config.json`

```json
{
  "enabled": true,
  "servers": {
    "filesystem": {
      "type": "stdio",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"],
      "timeout": 30000
    },
    "web": {
      "type": "sse",
      "url": "http://localhost:3000/sse",
      "headers": {"Authorization": "Bearer token"}
    },
    "modern-api": {
      "type": "streamable",
      "url": "http://localhost:3000/mcp",
      "timeout": 30000
    }
  }
}
```

**Usage**:

```bash
./code-agent --mcp-config ./examples/mcp.config.json
```

---

## Implementation Checklist

### Sprint 1: Configuration (Day 1)
- [ ] Create `internal/config/mcp.go` with types and LoadMCP()
- [ ] Create `internal/config/mcp_test.go` with unit tests
- [ ] Update `internal/config/config.go` to include MCPConfig
- [ ] All config tests pass: `go test ./internal/config`

### Sprint 2: Manager (Days 2-3)
- [ ] Create `internal/mcp/manager.go` with Manager type
- [ ] Implement NewManager() and LoadServers()
- [ ] Implement transport factories (stdio, sse, streamable)
- [ ] Create `internal/mcp/manager_test.go`
- [ ] All manager tests pass: `go test ./internal/mcp`

### Sprint 3: Integration (Day 4)
- [ ] **VERIFY**: Check llmagent.Config for correct Toolsets usage
- [ ] Update `main.go` to load MCP manager
- [ ] Wire MCPConfig through REPL
- [ ] Verify native tools still work
- [ ] Run `make test` - no regressions

### Sprint 4: CLI & Docs (Days 5-6)
- [ ] Add `/mcp list` command to REPL
- [ ] Add `/mcp tools` command to REPL
- [ ] Create example configs
- [ ] Update README with MCP usage
- [ ] Full system test: `make build && ./code-agent --help`

### Final: Quality Gate (Day 7)
- [ ] All tests pass: `go test ./...`
- [ ] Code formatted: `make fmt`
- [ ] No lint errors: `make lint`
- [ ] Manual integration test with real MCP server
- [ ] Documentation review and polish

---

## Key Differences from Original Document

| Section | Original | Corrected | Reason |
|---------|----------|-----------|--------|
| Transport names | `mcp.SSETransport`, `mcp.HTTPTransport` | `mcp.SSEClientTransport`, `mcp.StreamableClientTransport` | Verified against actual SDK |
| Config fields | Missing `Env`, `Cwd` | Added both fields | Required for subprocess control |
| CommandTransport | Missing `TerminateDuration` | Added field | Graceful shutdown support |
| Supported types | stdio, sse, http | stdio, sse, streamable | "streamable" is current HTTP standard |
| Integration notes | None | Added VERIFY task | Critical for llmagent integration |

---

## Known Limitations & TODOs

1. **llmagent.Config Verification**: Need to confirm Toolsets vs. Tools usage before Task 3
2. **Tool Filtering**: Phase 1 doesn't support include/exclude tool lists. Can be added in Phase 2.
3. **Error Recovery**: If an MCP server fails, currently skips it. Consider retry logic in Phase 2.
4. **Session Management**: mcptoolset handles this internally. Monitor for memory leaks with long-running agents.
5. **CLI Tool Discovery**: `/mcp tools` command implementation deferred to Phase 2.

---

## References

- **ADK-Go**: `research/adk-go/tool/mcptoolset/`
- **ADK-Go Example**: `research/adk-go/examples/mcp/main.go`
- **MCP SDK**: `github.com/modelcontextprotocol/go-sdk` (v0.7.0+)
- **Transport Types**: https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp
- **Code Agent Architecture**: `docs/ARCHITECTURE.md`
- **Tool Development**: `docs/TOOL_DEVELOPMENT.md`
- **Verification Report**: `VERIFICATION_REPORT.md` (NEW)
