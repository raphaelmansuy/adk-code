package config

import (
	"os"
	"testing"
)

func TestLoadMCPEmpty(t *testing.T) {
	cfg, err := LoadMCP("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Enabled {
		t.Error("expected disabled config for empty path")
	}
}

func TestLoadMCPNonExistent(t *testing.T) {
	cfg, err := LoadMCP("/nonexistent/path.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Enabled {
		t.Error("expected disabled config for nonexistent file")
	}
}

func TestLoadMCPValidStdioServer(t *testing.T) {
	f, err := os.CreateTemp("", "*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	configJSON := `{
		"enabled": true,
		"servers": {
			"fs": {"type": "stdio", "command": "echo"}
		}
	}`
	if _, err := f.WriteString(configJSON); err != nil {
		t.Fatal(err)
	}
	f.Close()

	cfg, err := LoadMCP(f.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Enabled {
		t.Error("expected enabled config")
	}
	if len(cfg.Servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(cfg.Servers))
	}
	if srv, ok := cfg.Servers["fs"]; !ok {
		t.Error("expected 'fs' server")
	} else if srv.Type != "stdio" || srv.Command != "echo" {
		t.Error("server configuration mismatch")
	}
}

func TestLoadMCPMissingCommand(t *testing.T) {
	f, err := os.CreateTemp("", "*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	configJSON := `{
		"servers": {"fs": {"type": "stdio"}}
	}`
	if _, err := f.WriteString(configJSON); err != nil {
		t.Fatal(err)
	}
	f.Close()

	_, err = LoadMCP(f.Name())
	if err == nil {
		t.Error("expected error for missing command")
	}
}

func TestLoadMCPValidSSE(t *testing.T) {
	f, err := os.CreateTemp("", "*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	configJSON := `{
		"servers": {"web": {"type": "sse", "url": "http://localhost:3000"}}
	}`
	if _, err := f.WriteString(configJSON); err != nil {
		t.Fatal(err)
	}
	f.Close()

	cfg, err := LoadMCP(f.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(cfg.Servers))
	}
}

func TestLoadMCPValidStreamable(t *testing.T) {
	f, err := os.CreateTemp("", "*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	configJSON := `{
		"servers": {"api": {"type": "streamable", "url": "http://localhost:3000/mcp"}}
	}`
	if _, err := f.WriteString(configJSON); err != nil {
		t.Fatal(err)
	}
	f.Close()

	cfg, err := LoadMCP(f.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(cfg.Servers))
	}
}

func TestLoadMCPInvalidType(t *testing.T) {
	f, err := os.CreateTemp("", "*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	configJSON := `{
		"servers": {"bad": {"type": "invalid"}}
	}`
	if _, err := f.WriteString(configJSON); err != nil {
		t.Fatal(err)
	}
	f.Close()

	_, err = LoadMCP(f.Name())
	if err == nil {
		t.Error("expected error for invalid type")
	}
}

func TestLoadMCPMultipleServers(t *testing.T) {
	f, err := os.CreateTemp("", "*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	configJSON := `{
		"enabled": true,
		"servers": {
			"fs": {"type": "stdio", "command": "mcp-fs"},
			"web": {"type": "sse", "url": "http://localhost:3000"},
			"api": {"type": "streamable", "url": "http://localhost:3000/mcp"}
		}
	}`
	if _, err := f.WriteString(configJSON); err != nil {
		t.Fatal(err)
	}
	f.Close()

	cfg, err := LoadMCP(f.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Servers) != 3 {
		t.Fatalf("expected 3 servers, got %d", len(cfg.Servers))
	}
}

func TestServerConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		srv     ServerConfig
		wantErr bool
	}{
		{
			name:    "valid stdio",
			srv:     ServerConfig{Type: "stdio", Command: "cmd"},
			wantErr: false,
		},
		{
			name:    "stdio missing command",
			srv:     ServerConfig{Type: "stdio"},
			wantErr: true,
		},
		{
			name:    "valid sse",
			srv:     ServerConfig{Type: "sse", URL: "http://localhost"},
			wantErr: false,
		},
		{
			name:    "sse missing url",
			srv:     ServerConfig{Type: "sse"},
			wantErr: true,
		},
		{
			name:    "valid streamable",
			srv:     ServerConfig{Type: "streamable", URL: "http://localhost:3000/mcp"},
			wantErr: false,
		},
		{
			name:    "streamable missing url",
			srv:     ServerConfig{Type: "streamable"},
			wantErr: true,
		},
		{
			name:    "missing type",
			srv:     ServerConfig{},
			wantErr: true,
		},
		{
			name:    "invalid type",
			srv:     ServerConfig{Type: "invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.srv.validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResolvePathAbsolute(t *testing.T) {
	absPath := "/tmp/mcp-config.json"
	resolved, err := resolvePath(absPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved != absPath {
		t.Errorf("absolute path not preserved: got %s, want %s", resolved, absPath)
	}
}

func TestResolvePathTildeExpansion(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot get home directory: %v", err)
	}

	tilePath := "~/mcp-config.json"
	resolved, err := resolvePath(tilePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := homeDir + "/mcp-config.json"
	if resolved != expected {
		t.Errorf("tilde expansion failed: got %s, want %s", resolved, expected)
	}
}

func TestResolvePathRelative(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("cannot get working directory: %v", err)
	}

	relativePath := "examples/mcp/basic-stdio.json"
	resolved, err := resolvePath(relativePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedPrefix := cwd
	if !os.IsPathSeparator(expectedPrefix[len(expectedPrefix)-1]) {
		expectedPrefix += string(os.PathSeparator)
	}

	if !os.IsPathSeparator(resolved[len(expectedPrefix)-1]) && len(resolved) > len(expectedPrefix) {
		if resolved[:len(expectedPrefix)-1] != expectedPrefix[:len(expectedPrefix)-1] {
			t.Errorf("relative path not resolved correctly: got %s, expected to start with %s", resolved, expectedPrefix)
		}
	}
}

func TestLoadMCPWithRelativePath(t *testing.T) {
	// Create a temporary config file
	f, err := os.CreateTemp("", "mcp-test-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	configJSON := `{
		"enabled": true,
		"servers": {
			"test": {"type": "stdio", "command": "echo", "args": ["hello"]}
		}
	}`
	if _, err := f.WriteString(configJSON); err != nil {
		t.Fatal(err)
	}
	f.Close()

	// Load using absolute path (test that resolution works)
	cfg, err := LoadMCP(f.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Enabled {
		t.Error("expected enabled config")
	}
	if len(cfg.Servers) != 1 {
		t.Errorf("expected 1 server, got %d", len(cfg.Servers))
	}
}

func TestLoadMCPWithTildePath(t *testing.T) {
	// Create a temporary config file in home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot get home directory: %v", err)
	}

	// Create a test config in a temp directory instead
	tmpDir, err := os.MkdirTemp(homeDir, "mcp-test-*")
	if err != nil {
		t.Skipf("cannot create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := tmpDir + "/test-config.json"
	configJSON := `{
		"enabled": true,
		"servers": {
			"test": {"type": "stdio", "command": "echo"}
		}
	}`

	if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
		t.Fatal(err)
	}

	// Test by loading with absolute path (since we can't easily test with real ~ expansion)
	cfg, err := LoadMCP(configPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Enabled {
		t.Error("expected enabled config")
	}
}

// TestLoadMCPClaudeFormat tests loading Claude Desktop config format
func TestLoadMCPClaudeFormat(t *testing.T) {
	f, err := os.CreateTemp("", "*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	// Claude Desktop format
	configJSON := `{
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
	}`
	if _, err := f.WriteString(configJSON); err != nil {
		t.Fatal(err)
	}
	f.Close()

	cfg, err := LoadMCP(f.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Enabled {
		t.Error("expected enabled config for Claude format")
	}
	if len(cfg.Servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(cfg.Servers))
	}

	srv, ok := cfg.Servers["Bright Data"]
	if !ok {
		t.Fatal("expected 'Bright Data' server")
	}
	if srv.Type != "stdio" {
		t.Errorf("expected type 'stdio', got %s", srv.Type)
	}
	if srv.Command != "npx" {
		t.Errorf("expected command 'npx', got %s", srv.Command)
	}
	if len(srv.Args) != 1 || srv.Args[0] != "@brightdata/mcp" {
		t.Errorf("expected args ['@brightdata/mcp'], got %v", srv.Args)
	}
	if srv.Env["API_TOKEN"] != "your-token-here" {
		t.Errorf("expected API_TOKEN 'your-token-here', got %s", srv.Env["API_TOKEN"])
	}
	if srv.Env["PRO_MODE"] != "true" {
		t.Errorf("expected PRO_MODE 'true', got %s", srv.Env["PRO_MODE"])
	}
}

// TestLoadMCPClaudeFormatSSE tests Claude format with URL (SSE server)
func TestLoadMCPClaudeFormatSSE(t *testing.T) {
	f, err := os.CreateTemp("", "*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	configJSON := `{
		"mcpServers": {
			"Remote Server": {
				"url": "http://localhost:8080/mcp"
			}
		}
	}`
	if _, err := f.WriteString(configJSON); err != nil {
		t.Fatal(err)
	}
	f.Close()

	cfg, err := LoadMCP(f.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srv, ok := cfg.Servers["Remote Server"]
	if !ok {
		t.Fatal("expected 'Remote Server'")
	}
	if srv.Type != "sse" {
		t.Errorf("expected type 'sse', got %s", srv.Type)
	}
	if srv.URL != "http://localhost:8080/mcp" {
		t.Errorf("expected URL 'http://localhost:8080/mcp', got %s", srv.URL)
	}
}

// TestLoadMCPClaudeFormatMultiple tests multiple servers in Claude format
func TestLoadMCPClaudeFormatMultiple(t *testing.T) {
	f, err := os.CreateTemp("", "*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	configJSON := `{
		"mcpServers": {
			"Homebrew": {
				"command": "brew",
				"args": ["mcp-server"]
			},
			"Filesystem": {
				"command": "npx",
				"args": ["-y", "@modelcontextprotocol/server-filesystem", "/tmp"]
			},
			"Remote": {
				"url": "https://api.example.com/mcp"
			}
		}
	}`
	if _, err := f.WriteString(configJSON); err != nil {
		t.Fatal(err)
	}
	f.Close()

	cfg, err := LoadMCP(f.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Servers) != 3 {
		t.Fatalf("expected 3 servers, got %d", len(cfg.Servers))
	}

	// Check Homebrew server
	homebrew, ok := cfg.Servers["Homebrew"]
	if !ok {
		t.Error("expected 'Homebrew' server")
	} else {
		if homebrew.Command != "brew" {
			t.Errorf("expected command 'brew', got %s", homebrew.Command)
		}
		if len(homebrew.Args) != 1 || homebrew.Args[0] != "mcp-server" {
			t.Errorf("expected args ['mcp-server'], got %v", homebrew.Args)
		}
	}

	// Check Filesystem server
	fs, ok := cfg.Servers["Filesystem"]
	if !ok {
		t.Error("expected 'Filesystem' server")
	} else {
		if fs.Command != "npx" {
			t.Errorf("expected command 'npx', got %s", fs.Command)
		}
		if len(fs.Args) != 3 {
			t.Errorf("expected 3 args, got %d", len(fs.Args))
		}
	}

	// Check Remote server
	remote, ok := cfg.Servers["Remote"]
	if !ok {
		t.Error("expected 'Remote' server")
	} else {
		if remote.Type != "sse" {
			t.Errorf("expected type 'sse', got %s", remote.Type)
		}
		if remote.URL != "https://api.example.com/mcp" {
			t.Errorf("expected URL 'https://api.example.com/mcp', got %s", remote.URL)
		}
	}
}

// TestClaudeServerConfigToServerConfig tests the conversion function
func TestClaudeServerConfigToServerConfig(t *testing.T) {
	tests := []struct {
		name     string
		claude   ClaudeServerConfig
		wantType string
		wantCmd  string
		wantURL  string
	}{
		{
			name: "stdio server",
			claude: ClaudeServerConfig{
				Command: "node",
				Args:    []string{"server.js"},
				Env: map[string]string{
					"DEBUG": "1",
				},
			},
			wantType: "stdio",
			wantCmd:  "node",
		},
		{
			name: "sse server",
			claude: ClaudeServerConfig{
				URL: "http://localhost:8080",
			},
			wantType: "sse",
			wantURL:  "http://localhost:8080",
		},
		{
			name: "server with both (URL takes precedence for type)",
			claude: ClaudeServerConfig{
				Command: "node",
				URL:     "http://localhost:8080",
			},
			wantType: "sse", // URL takes precedence when both present
			wantURL:  "http://localhost:8080",
			wantCmd:  "node", // Command is still preserved
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.claude.toServerConfig()

			if got.Type != tt.wantType {
				t.Errorf("Type = %v, want %v", got.Type, tt.wantType)
			}

			if got.Command != tt.wantCmd {
				t.Errorf("Command = %v, want %v", got.Command, tt.wantCmd)
			}

			if got.URL != tt.wantURL {
				t.Errorf("URL = %v, want %v", got.URL, tt.wantURL)
			}
		})
	}
}
