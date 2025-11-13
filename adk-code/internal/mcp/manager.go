package mcp

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"adk-code/internal/config"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/mcptoolset"
)

// Manager manages multiple MCP server connections
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

// NewManager creates a new MCP manager
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
			// Log but don't fail entire load - store the error for later inspection
			m.servers[name] = &server{name: name, err: err}
			continue
		}
	}

	return nil
}

func (m *Manager) loadServer(ctx context.Context, name string, cfg config.ServerConfig) error {
	transport, err := createTransport(cfg)
	if err != nil {
		return fmt.Errorf("failed to create transport: %w", err)
	}

	toolset, err := mcptoolset.New(mcptoolset.Config{
		Transport: transport,
	})
	if err != nil {
		return fmt.Errorf("failed to create toolset: %w", err)
	}

	m.servers[name] = &server{name: name, toolset: toolset}
	m.toolsets = append(m.toolsets, toolset)

	// Log successful server load
	fmt.Printf("✓ MCP server '%s' loaded successfully\n", name)
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

// Status returns the status of all servers (name and error, if any)
func (m *Manager) Status() map[string]error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make(map[string]error)
	for name, srv := range m.servers {
		status[name] = srv.err
	}
	return status
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
		Command:           cmd,
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

	// Create custom HTTP client with two layers of transport:
	// 1. filteringTransport - filters out ping events (workaround for SDK bug #636)
	// 2. acceptHeaderTransport - adds Accept headers
	httpClient := &http.Client{
		Timeout: timeout,
		Transport: &filteringTransport{
			base: &acceptHeaderTransport{
				base: http.DefaultTransport.(*http.Transport),
			},
		},
	}

	return &mcp.SSEClientTransport{
		Endpoint:   cfg.URL,
		HTTPClient: httpClient,
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

	// Create custom HTTP client with two layers of transport:
	// 1. filteringTransport - filters out ping events (workaround for SDK bug #636)
	// 2. acceptHeaderTransport - adds Accept headers
	httpClient := &http.Client{
		Timeout: timeout,
		Transport: &filteringTransport{
			base: &acceptHeaderTransport{
				base: http.DefaultTransport.(*http.Transport),
			},
		},
	}

	return &mcp.StreamableClientTransport{
		Endpoint:   cfg.URL,
		HTTPClient: httpClient,
	}, nil
}

// acceptHeaderTransport wraps http.Transport to inject Accept headers required by some MCP servers
type acceptHeaderTransport struct {
	base *http.Transport
}

func (t *acceptHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Add Accept header for servers that require both application/json and text/event-stream
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "application/json, text/event-stream")
	}
	return t.base.RoundTrip(req)
}
