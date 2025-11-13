# Phase 3: Advanced Features (Future Enhancements)

## Overview

Phase 3 adds advanced capabilities building on Phases 1-2: configuration hot-reload, alternative file formats, additional MCP protocols, observability, and UI enhancements.

- **Duration**: 3-4 weeks (after Phase 2)
- **Complexity**: High (advanced features, optional scope)
- **Outcome**: Fully-featured, observable, and user-friendly MCP support
- **Prerequisites**: Phase 1 and Phase 2 complete

---

## Architecture Enhancement

### Phase 1 → Phase 2 → Phase 3 Integration

```
Phase 1 (MVP)          Phase 2 (Production)      Phase 3 (Advanced)
├── Config parsing     + OAuth/health checks    + Hot-reload config
├── Basic manager      + Caching/parallel load  + YAML support
├── Tool wrapper       + Metrics collection     + Config migration tool
├── CLI commands       + Auth commands          + Resource/Prompt protocols
└── Basic tests        + Performance tests      + Prometheus metrics
                                                + UI enhancements
```

### New Components

1. **Config Watcher** (`pkg/mcp/config_watcher.go`) - Hot-reload on file changes
2. **YAML Parser** (`pkg/config/yaml.go`) - YAML config support
3. **Protocol Adapter** (`pkg/mcp/protocol.go`) - Resource/Prompt protocols
4. **Metrics Exporter** (`pkg/mcp/prometheus.go`) - Prometheus integration
5. **Tool Browser UI** (`internal/cli/browser.go`) - Interactive tool discovery

---

## Implementation Tasks

### Sprint 3.1: Configuration Hot-Reload (Days 1-6)

**Deliverables**:
- File watcher for config changes
- Runtime config updates
- Graceful server reload
- Rollback on invalid config

#### Task 3.1.1: Configuration Watcher

**File**: `pkg/mcp/config_watcher.go` (NEW)

```go
package mcp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"code_agent/internal/config"
)

// ConfigWatcher monitors config file for changes
type ConfigWatcher struct {
	filePath string
	watcher  *fsnotify.Watcher
	mu       sync.RWMutex

	// Current config
	config     *config.MCPConfig
	lastLoaded time.Time

	// Callbacks
	onConfigChange func(*config.MCPConfig) error
	onError        func(error)

	done chan struct{}
}

// NewConfigWatcher creates a watcher for config file changes
func NewConfigWatcher(filePath string) (*ConfigWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &ConfigWatcher{
		filePath: filePath,
		watcher:  watcher,
		done:     make(chan struct{}),
	}, nil
}

// Start begins watching for config changes
func (cw *ConfigWatcher) Start(ctx context.Context) error {
	// Add file to watcher
	if err := cw.watcher.Add(cw.filePath); err != nil {
		return fmt.Errorf("failed to watch config file: %w", err)
	}

	// Also watch directory for rename operations
	dir := filepath.Dir(cw.filePath)
	if err := cw.watcher.Add(dir); err != nil {
		return fmt.Errorf("failed to watch config directory: %w", err)
	}

	go cw.watchLoop(ctx)
	return nil
}

func (cw *ConfigWatcher) watchLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-cw.done:
			return

		case event, ok := <-cw.watcher.Events:
			if !ok {
				return
			}

			// Debounce rapid changes
			if cw.isRelevantEvent(event) {
				cw.handleConfigChange()
			}

		case err, ok := <-cw.watcher.Errors:
			if !ok {
				return
			}

			if cw.onError != nil {
				cw.onError(fmt.Errorf("watcher error: %w", err))
			}
		}
	}
}

func (cw *ConfigWatcher) isRelevantEvent(event fsnotify.Event) bool {
	// Only care about our config file
	if event.Name != cw.filePath {
		// Could be temp file during editor save
		return false
	}

	// Ignore chmod, focus on write/rename
	return event.Op&(fsnotify.Write|fsnotify.Rename|fsnotify.Create) != 0
}

func (cw *ConfigWatcher) handleConfigChange() {
	// Debounce: wait for writes to stabilize
	time.Sleep(500 * time.Millisecond)

	newConfig, err := config.LoadMCP(cw.filePath)
	if err != nil {
		if cw.onError != nil {
			cw.onError(fmt.Errorf("failed to load updated config: %w", err))
		}
		return
	}

	// Validate it's actually different
	cw.mu.Lock()
	if cw.lastLoaded.Add(2 * time.Second).After(time.Now()) {
		cw.mu.Unlock()
		return // Skip if loaded very recently
	}
	cw.mu.Unlock()

	// Notify callback
	if cw.onConfigChange != nil {
		if err := cw.onConfigChange(newConfig); err != nil {
			if cw.onError != nil {
				cw.onError(fmt.Errorf("failed to apply config: %w", err))
			}
			return
		}
	}

	cw.mu.Lock()
	cw.config = newConfig
	cw.lastLoaded = time.Now()
	cw.mu.Unlock()
}

// Stop stops the watcher
func (cw *ConfigWatcher) Stop() error {
	close(cw.done)
	return cw.watcher.Close()
}

// GetConfig returns current config
func (cw *ConfigWatcher) GetConfig() *config.MCPConfig {
	cw.mu.RLock()
	defer cw.mu.RUnlock()
	return cw.config
}

// OnConfigChange registers a callback for config changes
func (cw *ConfigWatcher) OnConfigChange(cb func(*config.MCPConfig) error) {
	cw.onConfigChange = cb
}

// OnError registers a callback for errors
func (cw *ConfigWatcher) OnError(cb func(error)) {
	cw.onError = cb
}
```

#### Task 3.1.2: Manager with Hot-Reload

**File**: `pkg/mcp/manager.go` (UPDATE)

Add hot-reload methods:

```go
// In Manager struct, add:
type Manager struct {
	// ... existing fields ...
	watcher *ConfigWatcher
	mu      sync.RWMutex
}

// EnableConfigHotReload enables automatic config reloading
func (m *Manager) EnableConfigHotReload(ctx context.Context, configPath string) error {
	watcher, err := NewConfigWatcher(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config watcher: %w", err)
	}

	watcher.OnConfigChange(m.handleConfigChange)
	watcher.OnError(m.handleWatcherError)

	if err := watcher.Start(ctx); err != nil {
		return err
	}

	m.mu.Lock()
	m.watcher = watcher
	m.mu.Unlock()

	return nil
}

func (m *Manager) handleConfigChange(newConfig *config.MCPConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create backup of current servers
	oldServers := make(map[string]*serverInstance)
	for k, v := range m.servers {
		oldServers[k] = v
	}

	// Clear and reload
	m.servers = make(map[string]*serverInstance)
	m.toolsets = make([]tool.Toolset, 0)

	if err := m.loadServersFromConfig(context.Background(), newConfig); err != nil {
		// Rollback to old servers
		m.servers = oldServers
		return fmt.Errorf("failed to load new config, rolled back: %w", err)
	}

	return nil
}

func (m *Manager) handleWatcherError(err error) {
	fmt.Fprintf(os.Stderr, "Config watcher error: %v\n", err)
	// Continue operation even if watcher fails
}

// DisableConfigHotReload disables automatic reloading
func (m *Manager) DisableConfigHotReload() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.watcher == nil {
		return nil
	}

	err := m.watcher.Stop()
	m.watcher = nil
	return err
}
```

#### Task 3.1.3: CLI Hot-Reload Commands

**File**: `internal/cli/commands/mcp.go` (UPDATE)

```go
case "/mcp reload":
	fmt.Println("Reloading config...")
	// Trigger manual reload
	if err := mgr.handleConfigChange(mgr.watcher.GetConfig()); err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("Config reloaded successfully")
	}

case "/mcp watch":
	// Toggle hot-reload
	if mgr.watcher == nil {
		fmt.Println("Starting config watcher...")
		if err := mgr.EnableConfigHotReload(ctx, configPath); err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Println("Config watcher started. Changes will be loaded automatically.")
		}
	} else {
		fmt.Println("Stopping config watcher...")
		if err := mgr.DisableConfigHotReload(); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}
```

**Success Criteria**:
- [ ] File watcher detects config changes
- [ ] Invalid config doesn't crash agent
- [ ] Rollback works on failed config
- [ ] Enable/disable toggle works

---

### Sprint 3.2: YAML Support & Migration (Days 7-11)

**Deliverables**:
- YAML config file support
- JSON to YAML migration tool
- Config validation for both formats
- Documentation with examples

#### Task 3.2.1: YAML Parser

**File**: `internal/config/yaml.go` (NEW)

```go
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadMCPYAML loads MCP config from YAML file
func LoadMCPYAML(filePath string) (*MCPConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &MCPConfig{Enabled: false}, nil
		}
		return nil, err
	}

	var cfg MCPConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
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

// LoadMCPAuto loads MCP config from either JSON or YAML based on file extension
func LoadMCPAuto(filePath string) (*MCPConfig, error) {
	if filePath == "" {
		return &MCPConfig{Enabled: false}, nil
	}

	// Check file extension
	switch {
	case len(filePath) > 5 && filePath[len(filePath)-5:] == ".yaml":
		return LoadMCPYAML(filePath)
	case len(filePath) > 4 && filePath[len(filePath)-4:] == ".yml":
		return LoadMCPYAML(filePath)
	default:
		return LoadMCP(filePath)
	}
}

// SaveMCPYAML saves MCP config to YAML file
func SaveMCPYAML(filePath string, cfg *MCPConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write YAML file: %w", err)
	}

	return nil
}
```

#### Task 3.2.2: Config Migration Tool

**File**: `cmd/mcp-config-migrate/main.go` (NEW)

```go
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"code_agent/internal/config"
)

func main() {
	inputFile := flag.String("i", "", "Input config file (JSON or YAML)")
	outputFile := flag.String("o", "", "Output config file (JSON or YAML)")
	format := flag.String("f", "yaml", "Output format (json or yaml)")
	validateOnly := flag.Bool("validate", false, "Only validate, don't convert")

	flag.Parse()

	if *inputFile == "" {
		fmt.Fprintf(os.Stderr, "Error: -i flag required\n")
		os.Exit(1)
	}

	// Load config
	var cfg *config.MCPConfig
	var err error

	switch filepath.Ext(*inputFile) {
	case ".yaml", ".yml":
		cfg, err = config.LoadMCPYAML(*inputFile)
	default:
		cfg, err = config.LoadMCP(*inputFile)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	if *validateOnly {
		fmt.Println("✓ Config is valid")
		if cfg.Enabled {
			fmt.Printf("  Servers: %d\n", len(cfg.Servers))
		} else {
			fmt.Println("  MCP disabled")
		}
		return
	}

	if *outputFile == "" {
		// Auto-generate output filename
		base := filepath.Base(*inputFile)
		ext := filepath.Ext(base)
		name := base[:len(base)-len(ext)]

		if *format == "yaml" {
			*outputFile = name + ".yaml"
		} else {
			*outputFile = name + ".json"
		}
	}

	// Save in new format
	switch *format {
	case "yaml":
		if err := config.SaveMCPYAML(*outputFile, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "json":
		// Implement JSON save similarly
		fmt.Fprintf(os.Stderr, "JSON save not yet implemented\n")
		os.Exit(1)
	default:
		fmt.Fprintf(os.Stderr, "Unknown format: %s\n", *format)
		os.Exit(1)
	}

	fmt.Printf("✓ Migrated %s → %s\n", *inputFile, *outputFile)
}
```

#### Task 3.2.3: YAML Examples & Documentation

**File**: `examples/config.mcp.yaml` (NEW)

```yaml
mcp:
  enabled: true
  servers:
    filesystem:
      type: stdio
      command: mcp-server-filesystem
      args:
        - /home/user/documents
      timeout: 30000
      excludeTools:
        - delete_recursive
        - format_disk

    github:
      type: sse
      url: https://mcp.github.example.com/search
      headers:
        Authorization: "Bearer ${GITHUB_TOKEN}"
      timeout: 30000
      oauth:
        autoDiscover: true
        scopes:
          - repo
          - user

    web:
      type: http
      httpUrl: http://localhost:3000/mcp
      timeout: 60000
      debug: false

  globalSettings:
    timeout: 30000
    debug: false
```

**Success Criteria**:
- [ ] YAML parsing works for valid files
- [ ] YAML parsing rejects invalid files with clear errors
- [ ] Migration tool converts between formats
- [ ] All config examples work in both JSON and YAML

---

### Sprint 3.3: Additional MCP Protocols (Days 12-18)

**Deliverables**:
- Resource protocol support
- Prompt protocol support
- Multi-protocol tool discovery
- Unified tool execution

#### Task 3.3.1: Protocol Adapter Architecture

**File**: `pkg/mcp/protocol.go` (NEW)

```go
package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPProtocol represents different MCP protocol capabilities
type MCPProtocol string

const (
	ProtocolTools    MCPProtocol = "tools"
	ProtocolResources MCPProtocol = "resources"
	ProtocolPrompts   MCPProtocol = "prompts"
)

// ServerCapabilities describes what protocols a server supports
type ServerCapabilities struct {
	Tools     *ToolsCapability     `json:"tools,omitempty"`
	Resources *ResourcesCapability `json:"resources,omitempty"`
	Prompts   *PromptsCapability   `json:"prompts,omitempty"`
}

type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type ResourcesCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
	Subscribe   bool `json:"subscribe,omitempty"`
}

type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ResourceItem represents an MCP resource
type ResourceItem struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// PromptItem represents an MCP prompt
type PromptItem struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Arguments   []struct {
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
		Required    bool   `json:"required,omitempty"`
	} `json:"arguments,omitempty"`
}

// ProtocolClient wraps MCP session and provides protocol-agnostic interface
type ProtocolClient struct {
	session *mcp.ClientSession
	server  *serverInstance

	capabilities *ServerCapabilities
}

// NewProtocolClient creates a client for a server session
func NewProtocolClient(session *mcp.ClientSession, server *serverInstance) *ProtocolClient {
	return &ProtocolClient{
		session: session,
		server:  server,
	}
}

// DiscoverCapabilities checks what protocols server supports
func (pc *ProtocolClient) DiscoverCapabilities(ctx context.Context) (*ServerCapabilities, error) {
	if pc.capabilities != nil {
		return pc.capabilities, nil
	}

	// Initialize with server info
	cap := &ServerCapabilities{
		Tools: &ToolsCapability{},
	}

	// In real implementation, query server for capabilities
	// This would require protocol negotiation with mcp.ClientSession

	pc.capabilities = cap
	return cap, nil
}

// ListResources returns available resources from server
func (pc *ProtocolClient) ListResources(ctx context.Context) ([]ResourceItem, error) {
	// Will be implemented when MCP SDK adds Resource protocol support
	return nil, fmt.Errorf("resource protocol not yet supported in MCP SDK")
}

// ReadResource reads content of a specific resource
func (pc *ProtocolClient) ReadResource(ctx context.Context, uri string) (string, error) {
	return "", fmt.Errorf("resource protocol not yet supported in MCP SDK")
}

// ListPrompts returns available prompts from server
func (pc *ProtocolClient) ListPrompts(ctx context.Context) ([]PromptItem, error) {
	// Will be implemented when MCP SDK adds Prompt protocol support
	return nil, fmt.Errorf("prompt protocol not yet supported in MCP SDK")
}

// GetPrompt retrieves a prompt with arguments resolved
func (pc *ProtocolClient) GetPrompt(ctx context.Context, name string, args map[string]string) (string, error) {
	return "", fmt.Errorf("prompt protocol not yet supported in MCP SDK")
}
```

#### Task 3.3.2: Protocol-Based Tool Discovery

**File**: `pkg/mcp/discovery.go` (NEW)

```go
package mcp

import (
	"context"
	"fmt"

	"google.golang.org/adk/tool"
)

// DiscoveryManager discovers all available items across all protocols
type DiscoveryManager struct {
	servers map[string]*ProtocolClient
}

// NewDiscoveryManager creates a discovery manager
func NewDiscoveryManager() *DiscoveryManager {
	return &DiscoveryManager{
		servers: make(map[string]*ProtocolClient),
	}
}

// DiscoverAllItems discovers tools, resources, and prompts from all servers
func (dm *DiscoveryManager) DiscoverAllItems(ctx context.Context) (map[string]interface{}, error) {
	result := map[string]interface{}{
		"tools":     []tool.Tool{},
		"resources": []ResourceItem{},
		"prompts":   []PromptItem{},
	}

	for serverName, client := range dm.servers {
		cap, err := client.DiscoverCapabilities(ctx)
		if err != nil {
			fmt.Printf("Warning: could not discover capabilities for %s: %v\n", serverName, err)
			continue
		}

		if cap.Tools != nil {
			// Tools are already handled by mcptoolset
		}

		if cap.Resources != nil {
			resources, err := client.ListResources(ctx)
			if err == nil {
				result["resources"] = append(result["resources"].([]ResourceItem), resources...)
			}
		}

		if cap.Prompts != nil {
			prompts, err := client.ListPrompts(ctx)
			if err == nil {
				result["prompts"] = append(result["prompts"].([]PromptItem), prompts...)
			}
		}
	}

	return result, nil
}
```

#### Task 3.3.3: Tests for Protocol Support

**File**: `pkg/mcp/protocol_test.go` (NEW)

```go
package mcp

import (
	"context"
	"testing"
)

func TestDiscoverCapabilities(t *testing.T) {
	// Mock implementation
	server := &serverInstance{name: "test"}
	client := NewProtocolClient(nil, server)

	cap, err := client.DiscoverCapabilities(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cap.Tools == nil {
		t.Error("expected Tools capability")
	}
}

func TestResourcesNotYetSupported(t *testing.T) {
	server := &serverInstance{name: "test"}
	client := NewProtocolClient(nil, server)

	_, err := client.ListResources(context.Background())
	if err == nil {
		t.Error("expected error for unsupported resource protocol")
	}
}

func TestPromptsNotYetSupported(t *testing.T) {
	server := &serverInstance{name: "test"}
	client := NewProtocolClient(nil, server)

	_, err := client.ListPrompts(context.Background())
	if err == nil {
		t.Error("expected error for unsupported prompt protocol")
	}
}
```

**Success Criteria**:
- [ ] Capability discovery works
- [ ] Protocol adapter integrates with manager
- [ ] Clear error messages for unsupported protocols
- [ ] Tests pass

---

### Sprint 3.4: Observability - Prometheus Export (Days 19-22)

**Deliverables**:
- Prometheus metrics export
- `/metrics` endpoint
- Integration with monitoring systems
- Dashboard ready metrics

#### Task 3.4.1: Prometheus Exporter

**File**: `pkg/mcp/prometheus.go` (NEW)

```go
package mcp

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// PrometheusMetrics exports metrics in Prometheus format
type PrometheusMetrics struct {
	toolCallsTotal      prometheus.Counter
	toolErrorsTotal     prometheus.Counter
	toolLatencySeconds  prometheus.Histogram
	cacheHitsTotal      prometheus.Counter
	cacheMissesTotal    prometheus.Counter
	connectionFailures  prometheus.CounterVec
	serverHealthStatus  prometheus.GaugeVec
	toolCacheSize       prometheus.GaugeVec
}

// NewPrometheusMetrics creates a Prometheus metrics collector
func NewPrometheusMetrics(registry prometheus.Registerer) *PrometheusMetrics {
	return &PrometheusMetrics{
		toolCallsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "mcp_tool_calls_total",
			Help: "Total number of MCP tool calls",
		}),

		toolErrorsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "mcp_tool_errors_total",
			Help: "Total number of MCP tool errors",
		}),

		toolLatencySeconds: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "mcp_tool_latency_seconds",
			Help:    "MCP tool execution latency",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 10),
		}),

		cacheHitsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "mcp_cache_hits_total",
			Help: "Total MCP cache hits",
		}),

		cacheMissesTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "mcp_cache_misses_total",
			Help: "Total MCP cache misses",
		}),

		connectionFailures: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "mcp_connection_failures_total",
				Help: "Total connection failures per server",
			},
			[]string{"server"},
		),

		serverHealthStatus: *promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "mcp_server_health",
				Help: "Server health status (1=healthy, 0=unhealthy)",
			},
			[]string{"server"},
		),

		toolCacheSize: *promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "mcp_tool_cache_entries",
				Help: "Number of tools cached per server",
			},
			[]string{"server"},
		),
	}
}

// RecordToolCall records a tool call
func (pm *PrometheusMetrics) RecordToolCall(serverName string, latencySeconds float64, success bool) {
	pm.toolCallsTotal.Inc()
	pm.toolLatencySeconds.Observe(latencySeconds)

	if !success {
		pm.toolErrorsTotal.Inc()
		pm.connectionFailures.WithLabelValues(serverName).Inc()
	}
}

// RecordCacheHit records a cache hit
func (pm *PrometheusMetrics) RecordCacheHit() {
	pm.cacheHitsTotal.Inc()
}

// RecordCacheMiss records a cache miss
func (pm *PrometheusMetrics) RecordCacheMiss() {
	pm.cacheMissesTotal.Inc()
}

// SetServerHealth sets server health status
func (pm *PrometheusMetrics) SetServerHealth(serverName string, healthy bool) {
	status := 0.0
	if healthy {
		status = 1.0
	}
	pm.serverHealthStatus.WithLabelValues(serverName).Set(status)
}

// SetToolCacheSize sets number of cached tools
func (pm *PrometheusMetrics) SetToolCacheSize(serverName string, count int) {
	pm.toolCacheSize.WithLabelValues(serverName).Set(float64(count))
}
```

#### Task 3.4.2: Metrics Endpoint

**File**: `pkg/mcp/metrics_handler.go` (NEW)

```go
package mcp

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// StartMetricsServer starts HTTP server for Prometheus metrics
func StartMetricsServer(addr string, registry *prometheus.Registry) error {
	http.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	go http.ListenAndServe(addr, nil)
	return nil
}
```

**Success Criteria**:
- [ ] Metrics exported in Prometheus format
- [ ] `/metrics` endpoint responds correctly
- [ ] All relevant metrics collected
- [ ] Performance impact minimal

---

### Sprint 3.5: UI Enhancements (Days 23-25)

**Deliverables**:
- Interactive tool browser
- Configuration wizard
- Visual status dashboard
- Help improvements

#### Task 3.5.1: Tool Browser CLI UI

**File**: `internal/cli/browser.go` (NEW)

```go
package cli

import (
	"fmt"
	"strings"

	"code_agent/pkg/mcp"
	"google.golang.org/adk/tool"
)

// ToolBrowser provides interactive tool discovery UI
type ToolBrowser struct {
	manager *mcp.Manager
}

// NewToolBrowser creates a tool browser
func NewToolBrowser(manager *mcp.Manager) *ToolBrowser {
	return &ToolBrowser{manager: manager}
}

// Browse displays interactive tool browser
func (tb *ToolBrowser) Browse(ctx context.Context) error {
	fmt.Println("\n╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║          MCP Tool Browser                                   ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝\n")

	// Show servers
	states := tb.manager.GetServerStates()
	if len(states) == 0 {
		fmt.Println("No MCP servers configured.")
		return nil
	}

	fmt.Println("📡 Available Servers:\n")

	selectedServer := tb.selectServer(states)
	if selectedServer == "" {
		return nil
	}

	// Show tools for selected server
	return tb.showServerTools(ctx, selectedServer)
}

func (tb *ToolBrowser) selectServer(states map[string]mcp.ServerState) string {
	fmt.Println("Select a server to view its tools:")
	fmt.Println()

	serverNames := make([]string, 0, len(states))
	for name := range states {
		serverNames = append(serverNames, name)
	}

	for i, name := range serverNames {
		state := states[name]
		status := "🟢"
		if state.Status != mcp.StatusConnected {
			status = "🔴"
		}
		fmt.Printf("  [%d] %s %s (%d tools)\n", i+1, status, name, state.ToolCount)
	}

	fmt.Print("\nEnter server number (0 to cancel): ")

	var choice int
	fmt.Scanln(&choice)

	if choice < 1 || choice > len(serverNames) {
		return ""
	}

	return serverNames[choice-1]
}

func (tb *ToolBrowser) showServerTools(ctx context.Context, serverName string) error {
	tools, err := tb.manager.GetToolsWithCache(ctx, serverName)
	if err != nil {
		fmt.Printf("Error loading tools: %v\n", err)
		return err
	}

	if len(tools) == 0 {
		fmt.Printf("\nNo tools found in %s\n", serverName)
		return nil
	}

	fmt.Printf("\n📚 Tools in %s (%d total):\n\n", serverName, len(tools))

	for _, tool := range tools {
		fmt.Printf("  • %s\n", tool.Name())
		if desc := tool.Description(); desc != "" {
			fmt.Printf("    %s\n", desc)
		}
		fmt.Println()
	}

	return nil
}
```

#### Task 3.5.2: Configuration Wizard

**File**: `internal/cli/wizard.go` (NEW)

```go
package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"code_agent/internal/config"
)

// ConfigWizard helps users create MCP configuration
type ConfigWizard struct {
	reader *bufio.Reader
}

// NewConfigWizard creates a configuration wizard
func NewConfigWizard() *ConfigWizard {
	return &ConfigWizard{
		reader: bufio.NewReader(os.Stdin),
	}
}

// Run starts the configuration wizard
func (cw *ConfigWizard) Run() (*config.MCPConfig, error) {
	fmt.Println("\n╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║          MCP Configuration Wizard                           ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝\n")

	cfg := &config.MCPConfig{
		Enabled: true,
		Servers: make(map[string]config.ServerConfig),
	}

	for {
		fmt.Println("\nAdd a new server? (y/n): ")
		resp, _ := cw.reader.ReadString('\n')
		resp = strings.TrimSpace(resp)

		if resp != "y" && resp != "yes" {
			break
		}

		serverCfg := cw.promptServer()
		cfg.Servers[serverCfg.Name] = serverCfg
	}

	return cfg, nil
}

func (cw *ConfigWizard) promptServer() config.ServerConfig {
	fmt.Print("\nServer name: ")
	name, _ := cw.reader.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Print("Server type (stdio/sse/http): ")
	typ, _ := cw.reader.ReadString('\n')
	typ = strings.TrimSpace(typ)

	var command, url string

	switch typ {
	case "stdio":
		fmt.Print("Command: ")
		command, _ = cw.reader.ReadString('\n')
		command = strings.TrimSpace(command)
	case "sse", "http":
		fmt.Print("URL: ")
		url, _ = cw.reader.ReadString('\n')
		url = strings.TrimSpace(url)
	}

	return config.ServerConfig{
		Name:    name,
		Type:    typ,
		Command: command,
		URL:     url,
	}
}
```

**Success Criteria**:
- [ ] Tool browser displays available servers
- [ ] Configuration wizard guides user through setup
- [ ] Visual indicators show server status
- [ ] Help text is clear and helpful

---

## Integration Across Phases

### Data Flow

```
Phase 1 Config → Phase 2 OAuth/Cache → Phase 3 Hot-Reload/Protocols
                                          ↓
                                    Metrics Export
                                          ↓
                                    UI Display
```

### CLI Command Summary

**Phase 1 (Basic)**:
- `/mcp status` - Show servers
- `/mcp list-tools` - List all tools
- `/mcp debug <server>` - Debug info

**Phase 2 (Production)**:
- `/mcp enable/disable` - Control servers
- `/mcp reconnect` - Force reconnection
- `/mcp auth` - OAuth setup
- `/mcp metrics` - Performance stats

**Phase 3 (Advanced)**:
- `/mcp reload` - Manual config reload
- `/mcp watch` - Enable hot-reload
- `/mcp browse` - Interactive tool browser
- `/mcp resources` - List resources
- `/mcp prompts` - List prompts

---

## Testing Strategy

### Unit Tests
- Config watcher debouncing
- YAML parsing roundtrip
- Protocol capability discovery
- Prometheus metric recording

### Integration Tests
- Hot-reload with file changes
- Migration tool conversion
- Protocol adapter with mock servers
- Metrics export format

### Manual Testing
- Configure watch mode, change config file
- Migrate from JSON to YAML
- Prometheus scrape `/metrics` endpoint
- Use interactive tool browser

---

## Success Criteria

### Functionality
- [ ] Configuration hot-reload works
- [ ] YAML config fully supported
- [ ] JSON/YAML migration tool works
- [ ] Prometheus metrics exported
- [ ] Tool browser interactive and responsive
- [ ] Configuration wizard guides users

### Quality
- [ ] 80%+ test coverage for new code
- [ ] No performance impact from features
- [ ] Graceful error handling throughout
- [ ] Clear documentation and examples

### User Experience
- [ ] Hot-reload transparent to users
- [ ] YAML easier to read/write than JSON
- [ ] Metrics integrated with monitoring
- [ ] Tool browser helps discovery

---

## Timeline Summary

| Sprint | Duration | Tasks |
|--------|----------|-------|
| 3.1 | 6 days | Config hot-reload, watcher, CLI commands |
| 3.2 | 5 days | YAML support, migration tool, examples |
| 3.3 | 7 days | Protocol adapters (resources, prompts) |
| 3.4 | 4 days | Prometheus metrics, `/metrics` endpoint |
| 3.5 | 3 days | Tool browser, wizard, UI enhancements |
| **Phase 3 Total** | **~25 days** | **Advanced features & observability** |

---

## File Structure

```
code_agent/
├── pkg/mcp/
│   ├── config_watcher.go       # NEW: File watcher
│   ├── config_watcher_test.go
│   ├── protocol.go              # NEW: Multi-protocol support
│   ├── protocol_test.go
│   ├── discovery.go             # NEW: Multi-protocol discovery
│   ├── prometheus.go            # NEW: Metrics export
│   ├── prometheus_test.go
│   ├── metrics_handler.go       # NEW: Metrics endpoint
│   └── (updates to existing files)
├── internal/config/
│   ├── yaml.go                  # NEW: YAML support
│   ├── yaml_test.go
│   └── (updates to mcp.go)
├── internal/cli/
│   ├── browser.go               # NEW: Tool browser
│   ├── wizard.go                # NEW: Config wizard
│   └── commands/mcp.go          # UPDATE: New commands
├── cmd/mcp-config-migrate/
│   └── main.go                  # NEW: Migration tool
├── examples/
│   ├── config.mcp.yaml          # NEW: YAML example
│   └── (JSON examples)
└── docs/
    ├── PROMETHEUS.md            # NEW: Metrics documentation
    ├── HOT_RELOAD.md            # NEW: Hot-reload guide
    ├── YAML.md                  # NEW: YAML format guide
    └── (existing docs)
```

---

## Known Limitations (Phase 3)

1. **MCP Resource/Prompt Protocols** - Not yet in Go SDK. Phase 3 stub ready for SDK updates.

2. **Connection Pooling** - Not implemented. Can be added in future if performance testing shows benefit.

3. **Advanced OAuth Flows** - Only bearer token support. Full PKCE/code flow in future.

4. **Tool Categorization** - Tools shown flat. Hierarchical organization in future.

5. **Metrics Retention** - In-memory only. Persistent storage in future.

---

## Roadmap Beyond Phase 3

**Future Enhancements**:
- Tool usage analytics and logging
- Per-tool execution policies and restrictions
- Advanced authentication (mTLS, custom protocols)
- Tool result caching with invalidation policies
- Distributed agent with multi-user support
- Web UI dashboard for management
- Integration with CI/CD pipelines
- Tool marketplace and discovery service

---

## Risk Mitigation

### Risk 1: File Watcher Performance
- Use debouncing to prevent rapid reloads
- Test with large config files
- Monitor CPU/memory impact

### Risk 2: Protocol Incompatibility
- Stub code ready for MCP SDK updates
- Clear deprecation path if protocol changes
- Fallback to Phase 2 without protocols

### Risk 3: Metrics Overhead
- Prometheus instrumentation tested for performance
- Optional metrics collection
- Metrics endpoint isolated from main agent

### Risk 4: UI Complexity
- Keep UIs simple and discoverable
- Comprehensive help text
- Progressive disclosure of advanced features

---

## Phase 3 Success Criteria Checklist

### Must Have
- [ ] Config hot-reload works reliably
- [ ] YAML support fully functional
- [ ] Migration tool converts formats correctly
- [ ] Prometheus metrics exported properly

### Should Have
- [ ] Tool browser UI works
- [ ] Configuration wizard helpful
- [ ] Protocol adapters extensible
- [ ] Documentation comprehensive

### Nice to Have
- [ ] Performance optimizations
- [ ] Advanced filtering
- [ ] Tool categorization
- [ ] Usage analytics

---

## Conclusion

Phase 3 completes the MCP support journey:
- **Phase 1**: MVP that works
- **Phase 2**: Production-grade features
- **Phase 3**: Advanced capabilities and observability

By end of Phase 3, code_agent will have enterprise-ready MCP support with:
- Flexible configuration (JSON/YAML with hot-reload)
- Robust lifecycle management (health checks, reconnection, OAuth)
- Performance optimization (caching, parallel loading, metrics)
- Multiple protocol support (tools, resources, prompts)
- Complete observability (Prometheus, metrics)
- Excellent user experience (tool browser, wizard, help)

The phased approach allows:
- Quick delivery of MVP (Phase 1)
- Production readiness (Phase 2)
- Advanced features as needed (Phase 3)
- Clear path for future enhancements

All while maintaining code quality, testing rigor, and user documentation standards that ensure long-term maintainability.
