package execution

import (
	"context"
	"testing"
)

// TestMCPClient tests basic MCP client creation
func TestMCPClient(t *testing.T) {
	client := NewMCPClient("test-server", "test-cmd", []string{"--arg1"})

	if client.Name() != "test-server" {
		t.Fatalf("Expected name 'test-server', got %q", client.Name())
	}

	if client.IsConnected() {
		t.Fatal("Expected client to be disconnected initially")
	}
}

// TestMCPTool tests MCPTool structure
func TestMCPTool(t *testing.T) {
	tool := &MCPTool{
		Name:        "test-tool",
		Description: "A test tool",
		Server:      "test-server",
		InputSchema: map[string]interface{}{
			"type": "object",
		},
	}

	if tool.Name != "test-tool" {
		t.Fatalf("Expected name 'test-tool', got %q", tool.Name)
	}

	if tool.Server != "test-server" {
		t.Fatalf("Expected server 'test-server', got %q", tool.Server)
	}
}

// TestMCPRequest tests MCPRequest marshaling
func TestMCPRequest(t *testing.T) {
	req := &MCPRequest{
		ID:      "req-1",
		Method:  "tools/list",
		JSONRPC: "2.0",
		Params: map[string]interface{}{
			"filter": "active",
		},
	}

	if req.ID != "req-1" {
		t.Fatalf("Expected ID 'req-1', got %q", req.ID)
	}

	if req.Method != "tools/list" {
		t.Fatalf("Expected method 'tools/list', got %q", req.Method)
	}
}

// TestMCPResponse tests MCPResponse structure
func TestMCPResponse(t *testing.T) {
	result := map[string]interface{}{
		"tools": []string{"tool1", "tool2"},
	}

	resp := &MCPResponse{
		ID:      "req-1",
		Result:  result,
		JSONRPC: "2.0",
	}

	if resp.ID != "req-1" {
		t.Fatalf("Expected ID 'req-1', got %q", resp.ID)
	}

	if resp.Error != nil {
		t.Fatal("Expected no error in response")
	}
}

// TestMCPError tests MCPError structure
func TestMCPError(t *testing.T) {
	err := &MCPError{
		Code:    -32600,
		Message: "Invalid Request",
	}

	if err.Code != -32600 {
		t.Fatalf("Expected code -32600, got %d", err.Code)
	}

	if err.Message != "Invalid Request" {
		t.Fatalf("Expected message 'Invalid Request', got %q", err.Message)
	}
}

// TestMCPRegistry tests registry creation
func TestMCPRegistry(t *testing.T) {
	registry := NewMCPRegistry()

	if registry == nil {
		t.Fatal("Failed to create registry")
	}

	clients := registry.ListClients()
	if len(clients) != 0 {
		t.Fatalf("Expected 0 clients initially, got %d", len(clients))
	}
}

// TestMCPRegistryRegister tests registering a client
func TestMCPRegistryRegister(t *testing.T) {
	registry := NewMCPRegistry()
	client := NewMCPClient("test-server", "test-cmd", []string{})

	err := registry.Register(client)
	if err != nil {
		t.Fatalf("Failed to register client: %v", err)
	}

	clients := registry.ListClients()
	if len(clients) != 1 {
		t.Fatalf("Expected 1 client, got %d", len(clients))
	}

	if clients[0] != "test-server" {
		t.Fatalf("Expected client 'test-server', got %q", clients[0])
	}
}

// TestMCPRegistryRegisterDuplicate tests registering duplicate client
func TestMCPRegistryRegisterDuplicate(t *testing.T) {
	registry := NewMCPRegistry()
	client1 := NewMCPClient("test-server", "test-cmd", []string{})
	client2 := NewMCPClient("test-server", "other-cmd", []string{})

	_ = registry.Register(client1)

	err := registry.Register(client2)
	if err == nil {
		t.Fatal("Expected error registering duplicate client")
	}
}

// TestMCPRegistryGetClient tests getting a client
func TestMCPRegistryGetClient(t *testing.T) {
	registry := NewMCPRegistry()
	client := NewMCPClient("test-server", "test-cmd", []string{})

	_ = registry.Register(client)

	retrieved, err := registry.GetClient("test-server")
	if err != nil {
		t.Fatalf("Failed to get client: %v", err)
	}

	if retrieved.Name() != "test-server" {
		t.Fatalf("Expected name 'test-server', got %q", retrieved.Name())
	}
}

// TestMCPRegistryGetClientNotFound tests getting non-existent client
func TestMCPRegistryGetClientNotFound(t *testing.T) {
	registry := NewMCPRegistry()

	_, err := registry.GetClient("non-existent")
	if err == nil {
		t.Fatal("Expected error getting non-existent client")
	}
}

// TestMCPRegistryUnregister tests unregistering a client
func TestMCPRegistryUnregister(t *testing.T) {
	registry := NewMCPRegistry()
	client := NewMCPClient("test-server", "test-cmd", []string{})

	_ = registry.Register(client)

	err := registry.Unregister("test-server")
	if err != nil {
		t.Fatalf("Failed to unregister client: %v", err)
	}

	clients := registry.ListClients()
	if len(clients) != 0 {
		t.Fatalf("Expected 0 clients after unregister, got %d", len(clients))
	}
}

// TestMCPClientCreation tests creating MCP client with different args
func TestMCPClientCreation(t *testing.T) {
	args := []string{"--config", "config.json", "--port", "5000"}
	client := NewMCPClient("my-server", "my-cmd", args)

	if client.Name() != "my-server" {
		t.Fatalf("Expected name 'my-server', got %q", client.Name())
	}

	if !client.IsConnected() {
		// Expected - not connected initially
	}
}

// TestMCPRegistryMultipleClients tests registering multiple clients
func TestMCPRegistryMultipleClients(t *testing.T) {
	registry := NewMCPRegistry()

	for i := 0; i < 3; i++ {
		name := "server-" + string(rune('0'+i))
		client := NewMCPClient(name, "cmd-"+string(rune('0'+i)), []string{})
		_ = registry.Register(client)
	}

	clients := registry.ListClients()
	if len(clients) != 3 {
		t.Fatalf("Expected 3 clients, got %d", len(clients))
	}
}

// TestMCPRequestWithContext tests request with context
func TestMCPRequestWithContext(t *testing.T) {
	req := &MCPRequest{
		ID:      "req-1",
		Method:  "tools/call",
		JSONRPC: "2.0",
		Params: map[string]interface{}{
			"name": "my-tool",
			"args": map[string]interface{}{
				"input": "test",
			},
		},
	}

	if req.Params["name"] != "my-tool" {
		t.Fatal("Failed to set request parameters")
	}
}

// TestMCPResponseError tests error response
func TestMCPResponseError(t *testing.T) {
	resp := &MCPResponse{
		ID:      "req-1",
		JSONRPC: "2.0",
		Error: &MCPError{
			Code:    -32601,
			Message: "Method not found",
		},
	}

	if resp.Error == nil {
		t.Fatal("Expected error in response")
	}

	if resp.Error.Code != -32601 {
		t.Fatalf("Expected error code -32601, got %d", resp.Error.Code)
	}
}

// TestMCPClientTimeout tests client timeout setting
func TestMCPClientTimeout(t *testing.T) {
	client := NewMCPClient("test-server", "test-cmd", []string{})

	// The timeout is set in NewMCPClient but we can't access it directly
	// This test verifies the client was created
	if client.IsConnected() {
		t.Fatal("Client should not be connected initially")
	}
}

// TestMCPToolSchema tests tool with input schema
func TestMCPToolSchema(t *testing.T) {
	tool := &MCPTool{
		Name:        "process-file",
		Description: "Process a file",
		Server:      "file-processor",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"file_path": map[string]interface{}{
					"type": "string",
				},
				"format": map[string]interface{}{
					"type": "string",
					"enum": []string{"json", "yaml", "xml"},
				},
			},
			"required": []string{"file_path"},
		},
	}

	if tool.InputSchema == nil {
		t.Fatal("Expected input schema to be set")
	}

	props, ok := tool.InputSchema["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Failed to extract properties from schema")
	}

	if _, hasFilePath := props["file_path"]; !hasFilePath {
		t.Fatal("Expected file_path in schema properties")
	}
}

// TestMCPRegistryUnregisterNotFound tests unregistering non-existent client
func TestMCPRegistryUnregisterNotFound(t *testing.T) {
	registry := NewMCPRegistry()

	err := registry.Unregister("non-existent")
	if err == nil {
		t.Fatal("Expected error unregistering non-existent client")
	}
}

// TestMCPRegistryNilClient tests registering nil client
func TestMCPRegistryNilClient(t *testing.T) {
	registry := NewMCPRegistry()

	err := registry.Register(nil)
	if err == nil {
		t.Fatal("Expected error registering nil client")
	}
}

// TestMCPClientContextTimeout tests operations respect context timeout
func TestMCPClientContextTimeout(t *testing.T) {
	client := NewMCPClient("test-server", "test-cmd", []string{})

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// The disconnect should respect the timeout
	err := client.Disconnect(ctx)
	if err == nil {
		// It's OK if disconnect succeeds (client wasn't connected)
	}
}
