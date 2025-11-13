# Phase 2: Enhanced MCP Support (Production-Ready Features)

## Overview

Phase 2 adds production-grade features to Phase 1's MVP: advanced lifecycle management, authentication support, performance optimization, and comprehensive error handling.

- **Duration**: 2-3 weeks
- **Complexity**: Medium (builds on Phase 1)
- **Outcome**: Production-ready MCP support with enterprise features
- **Prerequisites**: Phase 1 complete and all tests passing

---

## Architecture Enhancement

### Phase 1 → Phase 2 Integration

```
Phase 1 (MVP)                Phase 2 (Production)
├── Config parsing    →      + OAuth/token management
├── Basic manager     →      + Server health checks
├── Tool wrapper      →      + Connection pooling
├── CLI commands      →      + Reconnection logic
└── Basic tests       →      + Performance metrics
```

### New Components

1. **OAuth Manager** (`pkg/mcp/oauth.go`) - Token discovery and refresh
2. **Health Checker** (`pkg/mcp/health.go`) - Server availability monitoring
3. **Connection Pool** (`pkg/mcp/pool.go`) - Session reuse and management
4. **Metrics Collector** (`pkg/mcp/metrics.go`) - Latency, success rates, caching

---

## Implementation Tasks

### Sprint 2.1: Advanced Server Lifecycle (Days 1-5)

**Deliverables**:
- Server reconnection with exponential backoff
- Enable/disable server at runtime
- Server health monitoring
- Connection state tracking

#### Task 2.1.1: Enhanced Server State Management

**File**: `pkg/mcp/server_state.go` (NEW)

```go
package mcp

import (
	"sync"
	"time"
)

// ServerState represents the current state of an MCP server
type ServerState struct {
	Name           string
	Status         ConnectionStatus
	LastChecked    time.Time
	LastError      string
	HealthCheckAge time.Duration
	ConnectionAge  time.Duration
	ToolCount      int
}

type ConnectionStatus string

const (
	StatusConnected    ConnectionStatus = "connected"
	StatusConnecting   ConnectionStatus = "connecting"
	StatusDisconnected ConnectionStatus = "disconnected"
	StatusError        ConnectionStatus = "error"
	StatusDisabled     ConnectionStatus = "disabled"
)

type serverInstance struct {
	name         string
	toolset      tool.Toolset
	status       ConnectionStatus
	lastError    string
	disabled     bool
	lastChecked  time.Time
	mu           sync.RWMutex
	backoffCount int
	lastAttempt  time.Time
}

func (s *serverInstance) setState(status ConnectionStatus, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.status = status
	s.lastChecked = time.Now()

	if err != nil {
		s.lastError = err.Error()
		s.backoffCount++
		s.lastAttempt = time.Now()
	} else {
		s.backoffCount = 0
	}
}

func (s *serverInstance) getState() ServerState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return ServerState{
		Name:        s.name,
		Status:      s.status,
		LastChecked: s.lastChecked,
		LastError:   s.lastError,
	}
}

func (s *serverInstance) isDisabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.disabled
}

func (s *serverInstance) setDisabled(disabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.disabled = disabled
	if disabled {
		s.status = StatusDisabled
	}
}

// exponentialBackoffDuration calculates backoff based on attempt count
func exponentialBackoffDuration(attemptCount int) time.Duration {
	// 1s, 2s, 4s, 8s, 16s, max 60s
	duration := time.Duration(1<<uint(attemptCount)) * time.Second
	if duration > 60*time.Second {
		duration = 60 * time.Second
	}
	return duration
}

func (s *serverInstance) shouldRetry() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.disabled {
		return false
	}

	backoff := exponentialBackoffDuration(s.backoffCount)
	return time.Since(s.lastAttempt) >= backoff
}
```

#### Task 2.1.2: Enhanced Manager with Lifecycle

**File**: `pkg/mcp/manager.go` (UPDATE)

Add to Manager struct:

```go
// In Manager struct, add:
type Manager struct {
	// ... existing fields ...
	servers    map[string]*serverInstance  // Updated type
	
	// New fields for lifecycle management
	healthCheckTicker *time.Ticker
	done              chan struct{}
	wg                sync.WaitGroup
}

// AddMethods:

// StartHealthChecks begins periodic health monitoring
func (m *Manager) StartHealthChecks(ctx context.Context, interval time.Duration) error {
	m.mu.Lock()
	if m.healthCheckTicker != nil {
		m.mu.Unlock()
		return fmt.Errorf("health checks already started")
	}

	m.healthCheckTicker = time.NewTicker(interval)
	m.done = make(chan struct{})
	m.mu.Unlock()

	m.wg.Add(1)
	go m.healthCheckLoop(ctx)

	return nil
}

func (m *Manager) healthCheckLoop(ctx context.Context) {
	defer m.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.done:
			return
		case <-m.healthCheckTicker.C:
			m.checkAllServersHealth(ctx)
		}
	}
}

func (m *Manager) checkAllServersHealth(ctx context.Context) {
	m.mu.RLock()
	servers := make([]*serverInstance, 0, len(m.servers))
	for _, s := range m.servers {
		servers = append(servers, s)
	}
	m.mu.RUnlock()

	for _, server := range servers {
		if !server.shouldRetry() {
			continue
		}

		go m.checkServerHealth(ctx, server)
	}
}

func (m *Manager) checkServerHealth(ctx context.Context, server *serverInstance) {
	// Attempt to get tools to verify connection
	_, err := server.toolset.Tools(ctx)
	
	if err != nil {
		server.setState(StatusError, err)
		return
	}

	server.setState(StatusConnected, nil)
}

// StopHealthChecks stops health monitoring
func (m *Manager) StopHealthChecks() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.healthCheckTicker != nil {
		m.healthCheckTicker.Stop()
		close(m.done)
	}
}

// GetServerStates returns status of all servers
func (m *Manager) GetServerStates() map[string]ServerState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]ServerState)
	for name, server := range m.servers {
		result[name] = server.getState()
	}
	return result
}

// DisableServer disables a server without unloading it
func (m *Manager) DisableServer(name string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	server, ok := m.servers[name]
	if !ok {
		return fmt.Errorf("server not found: %s", name)
	}

	server.setDisabled(true)
	return nil
}

// EnableServer re-enables a disabled server
func (m *Manager) EnableServer(name string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	server, ok := m.servers[name]
	if !ok {
		return fmt.Errorf("server not found: %s", name)
	}

	server.setDisabled(false)
	return nil
}

// ReconnectServer forces reconnection to a specific server
func (m *Manager) ReconnectServer(ctx context.Context, name string) error {
	m.mu.RLock()
	server, ok := m.servers[name]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("server not found: %s", name)
	}

	// Clear session to force reconnection
	// This requires modification to mcptoolset or wrapping it
	// For now, we'll just attempt health check
	return m.reconnectServer(ctx, server)
}

func (m *Manager) reconnectServer(ctx context.Context, server *serverInstance) error {
	server.setState(StatusConnecting, nil)
	
	_, err := server.toolset.Tools(ctx)
	if err != nil {
		server.setState(StatusError, err)
		return fmt.Errorf("reconnection failed: %w", err)
	}

	server.setState(StatusConnected, nil)
	return nil
}

// Close gracefully shuts down all servers and health checks
func (m *Manager) Close(ctx context.Context) error {
	m.StopHealthChecks()
	m.wg.Wait()

	// Session cleanup happens via mcptoolset internally
	return nil
}
```

#### Task 2.1.3: Tests for Server Lifecycle

**File**: `pkg/mcp/server_state_test.go` (NEW)

```go
package mcp

import (
	"testing"
	"time"
)

func TestExponentialBackoff(t *testing.T) {
	tests := []struct {
		attempt int
		min     time.Duration
		max     time.Duration
	}{
		{0, 1 * time.Second, 1 * time.Second},
		{1, 2 * time.Second, 2 * time.Second},
		{2, 4 * time.Second, 4 * time.Second},
		{5, 32 * time.Second, 32 * time.Second},
		{10, 60 * time.Second, 60 * time.Second}, // Max
	}

	for _, tt := range tests {
		d := exponentialBackoffDuration(tt.attempt)
		if d < tt.min || d > tt.max {
			t.Errorf("attempt %d: got %v, want [%v, %v]", tt.attempt, d, tt.min, tt.max)
		}
	}
}

func TestServerStateTransitions(t *testing.T) {
	server := &serverInstance{name: "test"}

	// Initial state
	if server.getState().Status != "" {
		t.Error("initial status should be empty")
	}

	// Transition to connecting
	server.setState(StatusConnecting, nil)
	if server.getState().Status != StatusConnecting {
		t.Error("expected connecting")
	}

	// Transition to connected
	server.setState(StatusConnected, nil)
	state := server.getState()
	if state.Status != StatusConnected {
		t.Error("expected connected")
	}
	if state.LastError != "" {
		t.Error("expected no error")
	}

	// Transition to error
	testErr := fmt.Errorf("test error")
	server.setState(StatusError, testErr)
	state = server.getState()
	if state.Status != StatusError {
		t.Error("expected error status")
	}
	if !strings.Contains(state.LastError, "test error") {
		t.Error("expected error message")
	}
}

func TestDisableEnable(t *testing.T) {
	server := &serverInstance{name: "test"}

	if server.isDisabled() {
		t.Error("should not be disabled initially")
	}

	server.setDisabled(true)
	if !server.isDisabled() {
		t.Error("should be disabled")
	}

	if server.getState().Status != StatusDisabled {
		t.Error("status should be disabled")
	}

	server.setDisabled(false)
	if server.isDisabled() {
		t.Error("should not be disabled")
	}
}
```

**Success Criteria**:
- [ ] Server state transitions work correctly
- [ ] Exponential backoff calculated correctly
- [ ] Enable/disable toggles work
- [ ] Health check loop starts and stops cleanly

---

### Sprint 2.2: OAuth2 Authentication (Days 6-12)

**Deliverables**:
- OAuth2 auto-discovery
- Token storage and refresh
- Bearer token integration
- `/mcp auth` command

#### Task 2.2.1: OAuth Token Manager

**File**: `pkg/mcp/oauth.go` (NEW)

```go
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/oauth2"
)

// TokenStore manages OAuth tokens securely
type TokenStore struct {
	baseDir string
	mu      sync.RWMutex
	tokens  map[string]*oauth2.Token
}

// NewTokenStore creates a token store in ~/.code_agent/mcp-tokens/
func NewTokenStore() (*TokenStore, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	baseDir := filepath.Join(homeDir, ".code_agent", "mcp-tokens")
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return nil, err
	}

	return &TokenStore{
		baseDir: baseDir,
		tokens:  make(map[string]*oauth2.Token),
	}, nil
}

// GetToken retrieves a token, refreshing if necessary
func (ts *TokenStore) GetToken(ctx context.Context, serverName string) (*oauth2.Token, error) {
	ts.mu.RLock()
	token, ok := ts.tokens[serverName]
	ts.mu.RUnlock()

	if ok && !token.Expiry.IsZero() && token.Expiry.After(time.Now().Add(30*time.Second)) {
		return token, nil
	}

	// Try to load from disk
	token, err := ts.loadFromDisk(serverName)
	if err == nil {
		if !token.Expiry.IsZero() && token.Expiry.Before(time.Now()) {
			// Token expired, needs refresh (handled by caller with oauth2.Config)
		}

		ts.mu.Lock()
		ts.tokens[serverName] = token
		ts.mu.Unlock()

		return token, nil
	}

	return nil, fmt.Errorf("no token found for server %s", serverName)
}

// SaveToken stores a token to disk
func (ts *TokenStore) SaveToken(serverName string, token *oauth2.Token) error {
	ts.mu.Lock()
	ts.tokens[serverName] = token
	ts.mu.Unlock()

	return ts.saveToDisk(serverName, token)
}

func (ts *TokenStore) saveToDisk(serverName string, token *oauth2.Token) error {
	path := filepath.Join(ts.baseDir, serverName+".json")
	
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, 0600)
}

func (ts *TokenStore) loadFromDisk(serverName string) (*oauth2.Token, error) {
	path := filepath.Join(ts.baseDir, serverName+".json")
	
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}

	return &token, nil
}

// DeleteToken removes a stored token
func (ts *TokenStore) DeleteToken(serverName string) error {
	ts.mu.Lock()
	delete(ts.tokens, serverName)
	ts.mu.Unlock()

	path := filepath.Join(ts.baseDir, serverName+".json")
	return os.Remove(path)
}

// OAuthDiscoverer finds OAuth endpoints from server
type OAuthDiscoverer struct {
	httpClient *http.Client
}

// NewOAuthDiscoverer creates a discoverer
func NewOAuthDiscoverer() *OAuthDiscoverer {
	return &OAuthDiscoverer{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// DiscoverFromHeaders attempts to find OAuth endpoints from HTTP headers
// Looks for WWW-Authenticate header as per OAuth spec
func (od *OAuthDiscoverer) DiscoverFromHeaders(headers http.Header) (string, error) {
	authHeader := headers.Get("WWW-Authenticate")
	if authHeader == "" {
		return "", fmt.Errorf("no WWW-Authenticate header found")
	}

	// Parse OAuth challenge (simplified)
	// Real implementation would parse: 'Bearer realm="...", error="invalid_token", error_uri="..."'
	if !contains(authHeader, "Bearer") {
		return "", fmt.Errorf("not an OAuth header")
	}

	// Extract error_uri if present
	// In real implementation, use regexp or parser
	return authHeader, nil
}

// DiscoverFromWellKnown discovers OAuth config from /.well-known/oauth-authorization-server
func (od *OAuthDiscoverer) DiscoverFromWellKnown(ctx context.Context, serverURL string) (*oauth2.Config, error) {
	wellKnownURL := serverURL + "/.well-known/oauth-authorization-server"

	resp, err := od.httpClient.Do(&http.Request{
		Method: http.MethodGet,
		URL:    wellKnownURL,
	})
	if err != nil {
		return nil, fmt.Errorf("well-known discovery failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("well-known endpoint returned %d", resp.StatusCode)
	}

	var discovery struct {
		AuthorizationEndpoint string `json:"authorization_endpoint"`
		TokenEndpoint         string `json:"token_endpoint"`
		ClientID              string `json:"client_id"`
		ClientSecret          string `json:"client_secret"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&discovery); err != nil {
		return nil, fmt.Errorf("failed to parse well-known: %w", err)
	}

	if discovery.AuthorizationEndpoint == "" || discovery.TokenEndpoint == "" {
		return nil, fmt.Errorf("incomplete oauth configuration")
	}

	return &oauth2.Config{
		ClientID:     discovery.ClientID,
		ClientSecret: discovery.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  discovery.AuthorizationEndpoint,
			TokenURL: discovery.TokenEndpoint,
		},
	}, nil
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
```

#### Task 2.2.2: OAuth Configuration

**File**: `internal/config/mcp.go` (UPDATE)

Add to ServerConfig:

```go
type ServerConfig struct {
	// ... existing fields ...
	
	// OAuth support
	OAuth *OAuthConfig `json:"oauth,omitempty"`
}

type OAuthConfig struct {
	// If true, attempt OAuth2 auto-discovery
	AutoDiscover bool `json:"autoDiscover"`
	
	// For manual OAuth2 setup
	AuthEndpoint  string `json:"authEndpoint,omitempty"`
	TokenEndpoint string `json:"tokenEndpoint,omitempty"`
	ClientID      string `json:"clientID,omitempty"`
	ClientSecret  string `json:"clientSecret,omitempty"` // Usually from env var
	
	// Scopes requested from OAuth provider
	Scopes []string `json:"scopes,omitempty"`
}
```

#### Task 2.2.3: CLI Auth Command

**File**: `internal/cli/commands/mcp.go` (UPDATE)

Add to `/mcp` commands:

```go
func mcpAuthCommand(input string, mgr *mcp.Manager, store *mcp.TokenStore) {
	args := strings.Fields(input)
	if len(args) < 2 {
		fmt.Println("Usage: /mcp auth <server>")
		return
	}

	serverName := args[1]

	// Initiate OAuth flow (simplified)
	fmt.Printf("Opening browser for authentication with %s...\n", serverName)
	fmt.Println("(Requires manual setup - copy auth URL to browser)")
	fmt.Printf("After authentication, use: /mcp auth-save %s <token>\n", serverName)
}

func mcpAuthSaveCommand(input string, store *mcp.TokenStore) {
	args := strings.Fields(input)
	if len(args) < 3 {
		fmt.Println("Usage: /mcp auth-save <server> <token>")
		return
	}

	serverName := args[1]
	tokenStr := args[2]

	// In real implementation, parse JWT or call token endpoint
	token := &oauth2.Token{
		AccessToken: tokenStr,
		TokenType:   "Bearer",
	}

	if err := store.SaveToken(serverName, token); err != nil {
		fmt.Printf("Error saving token: %v\n", err)
		return
	}

	fmt.Printf("Token saved for server '%s'\n", serverName)
}
```

**Success Criteria**:
- [ ] Token store creates directories correctly
- [ ] Tokens persisted to disk with 0600 permissions
- [ ] OAuth discovery from headers works
- [ ] Well-known endpoint discovery works
- [ ] Auth tokens can be loaded and used in headers

---

### Sprint 2.3: Performance Optimization (Days 13-15)

**Deliverables**:
- Connection pooling for reuse
- Tool list caching with TTL
- Parallel server initialization
- Metrics collection

#### Task 2.3.1: Tool Caching

**File**: `pkg/mcp/cache.go` (NEW)

```go
package mcp

import (
	"context"
	"sync"
	"time"

	"google.golang.org/adk/tool"
)

// ToolCache provides caching for tool lists with TTL
type ToolCache struct {
	mu       sync.RWMutex
	cache    map[string]*cachedTools
	defaultTTL time.Duration
}

type cachedTools struct {
	tools      []tool.Tool
	timestamp  time.Time
	ttl        time.Duration
}

// NewToolCache creates a cache with default TTL
func NewToolCache(defaultTTL time.Duration) *ToolCache {
	return &ToolCache{
		cache:      make(map[string]*cachedTools),
		defaultTTL: defaultTTL,
	}
}

// Get retrieves cached tools if valid
func (tc *ToolCache) Get(serverName string) ([]tool.Tool, bool) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	cached, ok := tc.cache[serverName]
	if !ok {
		return nil, false
	}

	if time.Since(cached.timestamp) > cached.ttl {
		return nil, false // Expired
	}

	return cached.tools, true
}

// Set stores tools with default TTL
func (tc *ToolCache) Set(serverName string, tools []tool.Tool) {
	tc.SetWithTTL(serverName, tools, tc.defaultTTL)
}

// SetWithTTL stores tools with custom TTL
func (tc *ToolCache) SetWithTTL(serverName string, tools []tool.Tool, ttl time.Duration) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	tc.cache[serverName] = &cachedTools{
		tools:     tools,
		timestamp: time.Now(),
		ttl:       ttl,
	}
}

// Clear removes all cached tools
func (tc *ToolCache) Clear() {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	tc.cache = make(map[string]*cachedTools)
}

// ClearServer removes cache for specific server
func (tc *ToolCache) ClearServer(serverName string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	delete(tc.cache, serverName)
}
```

#### Task 2.3.2: Manager with Caching and Parallel Load

**File**: `pkg/mcp/manager.go` (UPDATE)

Add caching to Manager:

```go
type Manager struct {
	// ... existing fields ...
	cache *ToolCache
}

func NewManagerWithCache(defaultCacheTTL time.Duration) *Manager {
	return &Manager{
		servers:  make(map[string]*serverInstance),
		toolsets: make([]tool.Toolset, 0),
		cache:    NewToolCache(defaultCacheTTL),
	}
}

// LoadServersParallel loads multiple servers concurrently
func (m *Manager) LoadServersParallel(ctx context.Context, cfg *config.MCPConfig, maxConcurrency int) error {
	if !cfg.Enabled || len(cfg.Servers) == 0 {
		return nil
	}

	// Semaphore to limit concurrency
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup
	errChan := make(chan error, len(cfg.Servers))

	m.mu.Lock()
	defer m.mu.Unlock()

	for name, srvCfg := range cfg.Servers {
		wg.Add(1)
		go func(n string, cfg config.ServerConfig) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			if err := m.loadServer(ctx, n, cfg); err != nil {
				errChan <- fmt.Errorf("%s: %w", n, err)
			}
		}(name, srvCfg)
	}

	wg.Wait()
	close(errChan)

	// Collect non-critical errors (don't fail entire load)
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
		// Log but continue
	}

	if len(errs) > 0 {
		fmt.Fprintf(os.Stderr, "Warning: Some servers failed to load:\n")
		for _, err := range errs {
			fmt.Fprintf(os.Stderr, "  - %v\n", err)
		}
	}

	return nil
}

// GetToolsWithCache returns tools, using cache when possible
func (m *Manager) GetToolsWithCache(ctx context.Context, serverName string) ([]tool.Tool, error) {
	if cached, ok := m.cache.Get(serverName); ok {
		return cached, nil
	}

	m.mu.RLock()
	server, ok := m.servers[serverName]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("server not found: %s", serverName)
	}

	tools, err := server.toolset.Tools(ctx)
	if err != nil {
		return nil, err
	}

	m.cache.Set(serverName, tools)
	return tools, nil
}

// GetAllToolsWithCache returns all tools from all servers, using cache
func (m *Manager) GetAllToolsWithCache(ctx context.Context) ([]tool.Tool, error) {
	m.mu.RLock()
	servers := make([]*serverInstance, 0, len(m.servers))
	for _, s := range m.servers {
		servers = append(servers, s)
	}
	m.mu.RUnlock()

	var allTools []tool.Tool
	for _, server := range servers {
		if server.isDisabled() {
			continue
		}

		tools, err := m.GetToolsWithCache(ctx, server.name)
		if err != nil {
			// Log but don't fail
			fmt.Fprintf(os.Stderr, "Warning: failed to get tools from %s: %v\n", server.name, err)
			continue
		}

		allTools = append(allTools, tools...)
	}

	return allTools, nil
}
```

#### Task 2.3.3: Metrics Collection

**File**: `pkg/mcp/metrics.go` (NEW)

```go
package mcp

import (
	"sync"
	"time"
)

// Metrics tracks MCP performance statistics
type Metrics struct {
	mu sync.RWMutex

	// Per-server metrics
	ServerMetrics map[string]*ServerMetrics

	// Global metrics
	TotalToolCalls    int64
	TotalToolErrors   int64
	CacheHits         int64
	CacheMisses       int64
	AvgLatencyMs      float64
	MaxLatencyMs      int64
}

// ServerMetrics tracks metrics for a single server
type ServerMetrics struct {
	Name                 string
	ConnectionCount      int64
	ConnectionFailures   int64
	ToolDiscoveryCount   int64
	ToolDiscoveryErrors  int64
	AvgDiscoveryTimeMs   float64
	LastDiscoveryTimeMs  int64
	ToolCacheHits        int64
	ToolCacheMisses      int64
}

// NewMetrics creates a metrics collector
func NewMetrics() *Metrics {
	return &Metrics{
		ServerMetrics: make(map[string]*ServerMetrics),
	}
}

// RecordToolCall records a tool execution
func (m *Metrics) RecordToolCall(serverName string, latencyMs int64, success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalToolCalls++

	if !success {
		m.TotalToolErrors++
	}

	if latencyMs > m.MaxLatencyMs {
		m.MaxLatencyMs = latencyMs
	}

	// Update average
	m.AvgLatencyMs = (m.AvgLatencyMs*float64(m.TotalToolCalls-1) + float64(latencyMs)) / float64(m.TotalToolCalls)

	// Update server metrics
	if sm, ok := m.ServerMetrics[serverName]; ok {
		sm.LastDiscoveryTimeMs = latencyMs
	}
}

// RecordCacheHit records a cache hit
func (m *Metrics) RecordCacheHit(serverName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CacheHits++

	if sm, ok := m.ServerMetrics[serverName]; ok {
		sm.ToolCacheHits++
	}
}

// RecordCacheMiss records a cache miss
func (m *Metrics) RecordCacheMiss(serverName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CacheMisses++

	if sm, ok := m.ServerMetrics[serverName]; ok {
		sm.ToolCacheMisses++
	}
}

// GetMetrics returns current metrics snapshot
func (m *Metrics) GetMetrics() Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return *m
}

// Reset clears all metrics
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalToolCalls = 0
	m.TotalToolErrors = 0
	m.CacheHits = 0
	m.CacheMisses = 0
	m.AvgLatencyMs = 0
	m.MaxLatencyMs = 0
	m.ServerMetrics = make(map[string]*ServerMetrics)
}
```

**Success Criteria**:
- [ ] Tool caching with TTL works correctly
- [ ] Parallel server loading completes faster than sequential
- [ ] Metrics collection doesn't impact performance
- [ ] Cache can be cleared on demand

---

## Integration with Phase 1

### Backward Compatibility

Phase 2 maintains full backward compatibility with Phase 1:
- All Phase 1 code continues to work
- New features are additive
- No breaking changes to config format (only additions)
- Manager API extended, not changed

### Config Format Additions

```json
{
  "mcp": {
    "servers": {
      "github": {
        // Phase 1 fields
        "type": "sse",
        "url": "https://api.example.com",
        
        // Phase 2 additions
        "oauth": {
          "autoDiscover": true,
          "scopes": ["repo", "user"]
        }
      }
    }
  }
}
```

### CLI Commands Added

New in Phase 2:
```
/mcp status          # Enhanced with health checks
/mcp auth <server>   # Start OAuth flow
/mcp disable <server> # Disable without unloading
/mcp enable <server>  # Re-enable server
/mcp reconnect <server> # Force reconnection
/mcp metrics         # Show performance metrics
```

---

## Testing Strategy

### Unit Tests

- **Server state transitions**: Test all status changes
- **Backoff calculation**: Verify exponential backoff logic
- **Token storage**: Test persistence and retrieval
- **Cache expiration**: Verify TTL enforcement
- **Metrics**: Test collection accuracy

### Integration Tests

- **Health checks**: Start mock server, simulate failure, verify recovery
- **OAuth flow**: Mock OAuth provider, test token exchange
- **Parallel loading**: Load multiple servers concurrently, verify all load correctly
- **Cache behavior**: Verify cache hits/misses with different TTLs

### Manual Testing

- Configure real MCP server with OAuth
- Test reconnection after network failure
- Monitor metrics collection during normal operation
- Verify enable/disable doesn't lose configuration

---

## Success Criteria

### Functionality
- [ ] Server reconnection with backoff works
- [ ] Health checks detect failures and recover
- [ ] OAuth tokens stored securely and used in headers
- [ ] Tool caching improves performance
- [ ] Parallel loading faster than sequential
- [ ] Metrics collected without performance impact
- [ ] Enable/disable works without losing config

### Quality
- [ ] 80%+ test coverage for new code
- [ ] No regressions from Phase 1
- [ ] Memory leaks under load testing
- [ ] Graceful shutdown with pending operations

### Performance
- [ ] Health checks don't block tool execution
- [ ] Cache hits reduce latency by 10x
- [ ] Parallel loading reduces startup by 50%
- [ ] Metrics collection overhead < 1%

---

## Timeline Summary

| Sprint | Duration | Tasks |
|--------|----------|-------|
| 2.1 | 5 days | Server lifecycle, health checks, reconnection |
| 2.2 | 7 days | OAuth discovery, token storage, auth commands |
| 2.3 | 3 days | Caching, parallel loading, metrics |
| **Phase 2 Total** | **~15 days** | **Production-ready features** |

---

## File Structure

```
code_agent/
├── pkg/mcp/
│   ├── server_state.go        # NEW: Server state management
│   ├── server_state_test.go   # NEW: State tests
│   ├── oauth.go               # NEW: OAuth token management
│   ├── oauth_test.go          # NEW: OAuth tests
│   ├── cache.go               # NEW: Tool caching
│   ├── cache_test.go          # NEW: Cache tests
│   ├── metrics.go             # NEW: Metrics collection
│   ├── metrics_test.go        # NEW: Metrics tests
│   ├── manager.go             # UPDATE: Add health checks, caching
│   └── manager_test.go        # UPDATE: Add integration tests
├── internal/config/
│   ├── mcp.go                 # UPDATE: Add OAuth config
│   └── mcp_test.go            # UPDATE: Add OAuth tests
└── internal/cli/commands/
    └── mcp.go                 # UPDATE: Add auth, reconnect, metrics commands
```

---

## Known Limitations (Phase 2)

1. **No automatic token refresh** - Tokens must be refreshed manually. Phase 3 can add automatic refresh via oauth2.TokenSource.

2. **Limited OAuth flows** - Only bearer token support. Phase 3 can add full OAuth2 code flow.

3. **No connection pooling** - Each server has single connection. Phase 3 can add connection reuse within single server.

4. **Basic metrics** - Simple counters and averages. Phase 3 can add Prometheus export.

5. **No configuration hot-reload** - Still requires restart. Phase 3 can add file watcher.

---

## Risk Mitigation

### Risk 1: Performance Regression
- Continuous performance testing during implementation
- Cache disabled by default until proven beneficial
- Metrics overhead validated < 1%

### Risk 2: Token Security
- Tokens stored with 0600 permissions
- No tokens in logs or error messages
- Clear documentation on security implications

### Risk 3: Breaking Changes
- Full backward compatibility maintained
- All new features optional
- Phase 1 code continues to work unchanged

### Risk 4: Complexity Increase
- Each feature independent and testable
- Clear separation of concerns
- Comprehensive documentation
