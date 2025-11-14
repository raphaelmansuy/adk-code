package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

// MCPClient represents a Model Context Protocol client for external tool integration
type MCPClient struct {
	// name is the name of the MCP server
	name string

	// cmd is the command to run the MCP server
	cmd string

	// args are the arguments for the MCP server
	args []string

	// stdin is the input stream to the server
	stdin io.WriteCloser

	// stdout is the output stream from the server
	stdout io.ReadCloser

	// process is the running process
	process *exec.Cmd

	// connected indicates if the client is connected
	connected bool

	// timeout for operations
	timeout time.Duration
}

// MCPTool represents a tool available through MCP
type MCPTool struct {
	// Name is the tool name
	Name string `json:"name"`

	// Description is the tool description
	Description string `json:"description"`

	// InputSchema is the tool input schema
	InputSchema map[string]interface{} `json:"input_schema"`

	// Server is the MCP server providing this tool
	Server string `json:"server"`
}

// MCPRequest represents a request to an MCP tool
type MCPRequest struct {
	// ID is the request ID
	ID string `json:"id"`

	// Method is the MCP method
	Method string `json:"method"`

	// Params are the request parameters
	Params map[string]interface{} `json:"params"`

	// JSONRPC is the JSON-RPC version
	JSONRPC string `json:"jsonrpc"`
}

// MCPResponse represents a response from an MCP tool
type MCPResponse struct {
	// ID is the response ID
	ID string `json:"id"`

	// Result is the tool result
	Result interface{} `json:"result,omitempty"`

	// Error is the error message if failed
	Error *MCPError `json:"error,omitempty"`

	// JSONRPC is the JSON-RPC version
	JSONRPC string `json:"jsonrpc"`
}

// MCPError represents an MCP error
type MCPError struct {
	// Code is the error code
	Code int `json:"code"`

	// Message is the error message
	Message string `json:"message"`

	// Data is additional error data
	Data interface{} `json:"data,omitempty"`
}

// NewMCPClient creates a new MCP client
func NewMCPClient(name, cmd string, args []string) *MCPClient {
	return &MCPClient{
		name:      name,
		cmd:       cmd,
		args:      args,
		connected: false,
		timeout:   30 * time.Second,
	}
}

// Connect connects to the MCP server
func (c *MCPClient) Connect(ctx context.Context) error {
	if c.connected {
		return fmt.Errorf("already connected")
	}

	// Create the process
	c.process = exec.CommandContext(ctx, c.cmd, c.args...)

	// Setup pipes
	stdin, err := c.process.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := c.process.StdoutPipe()
	if err != nil {
		stdin.Close()
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	c.process.Stderr = os.Stderr

	// Start the process
	if err := c.process.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		return fmt.Errorf("failed to start MCP server: %w", err)
	}

	c.stdin = stdin
	c.stdout = stdout
	c.connected = true

	return nil
}

// Disconnect disconnects from the MCP server
func (c *MCPClient) Disconnect(ctx context.Context) error {
	if !c.connected {
		return fmt.Errorf("not connected")
	}

	// Close stdin to signal end of input
	if c.stdin != nil {
		c.stdin.Close()
	}

	// Kill the process with timeout
	if c.process != nil && c.process.Process != nil {
		done := make(chan error, 1)
		go func() {
			done <- c.process.Wait()
		}()

		select {
		case <-ctx.Done():
			c.process.Process.Kill()
			return ctx.Err()
		case <-done:
		}
	}

	if c.stdout != nil {
		c.stdout.Close()
	}

	c.connected = false
	return nil
}

// ListTools lists available tools from the MCP server
func (c *MCPClient) ListTools(ctx context.Context) ([]*MCPTool, error) {
	if !c.connected {
		return nil, fmt.Errorf("not connected")
	}

	req := &MCPRequest{
		ID:      "list-tools",
		Method:  "tools/list",
		JSONRPC: "2.0",
		Params:  map[string]interface{}{},
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("server error: %s", resp.Error.Message)
	}

	// Parse tools from response
	var tools []*MCPTool
	if toolsData, ok := resp.Result.([]interface{}); ok {
		for _, toolData := range toolsData {
			if toolMap, ok := toolData.(map[string]interface{}); ok {
				tool := &MCPTool{
					Server: c.name,
				}

				if name, ok := toolMap["name"].(string); ok {
					tool.Name = name
				}

				if desc, ok := toolMap["description"].(string); ok {
					tool.Description = desc
				}

				if schema, ok := toolMap["input_schema"].(map[string]interface{}); ok {
					tool.InputSchema = schema
				}

				tools = append(tools, tool)
			}
		}
	}

	return tools, nil
}

// CallTool calls a tool on the MCP server
func (c *MCPClient) CallTool(ctx context.Context, toolName string, params map[string]interface{}) (interface{}, error) {
	if !c.connected {
		return nil, fmt.Errorf("not connected")
	}

	req := &MCPRequest{
		ID:      fmt.Sprintf("call-%s", toolName),
		Method:  "tools/call",
		JSONRPC: "2.0",
		Params: map[string]interface{}{
			"name": toolName,
			"args": params,
		},
	}

	resp, err := c.sendRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("tool error: %s", resp.Error.Message)
	}

	return resp.Result, nil
}

// sendRequest sends a request to the MCP server and waits for response
func (c *MCPClient) sendRequest(ctx context.Context, req *MCPRequest) (*MCPResponse, error) {
	// Encode and send request
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send request with context timeout
	done := make(chan error, 1)
	go func() {
		_, err := c.stdin.Write(data)
		if err == nil {
			_, err = c.stdin.Write([]byte("\n"))
		}
		done <- err
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-done:
		if err != nil {
			return nil, fmt.Errorf("failed to send request: %w", err)
		}
	}

	// Read response
	buf := make([]byte, 4096)
	n, err := c.stdout.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var resp MCPResponse
	if err := json.Unmarshal(buf[:n], &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &resp, nil
}

// IsConnected returns whether the client is connected
func (c *MCPClient) IsConnected() bool {
	return c.connected
}

// Name returns the MCP server name
func (c *MCPClient) Name() string {
	return c.name
}

// MCPRegistry manages multiple MCP clients
type MCPRegistry struct {
	clients map[string]*MCPClient
}

// NewMCPRegistry creates a new MCP registry
func NewMCPRegistry() *MCPRegistry {
	return &MCPRegistry{
		clients: make(map[string]*MCPClient),
	}
}

// Register registers an MCP client
func (r *MCPRegistry) Register(client *MCPClient) error {
	if client == nil {
		return fmt.Errorf("client is nil")
	}

	if _, exists := r.clients[client.name]; exists {
		return fmt.Errorf("client %q already registered", client.name)
	}

	r.clients[client.name] = client
	return nil
}

// Unregister unregisters an MCP client
func (r *MCPRegistry) Unregister(name string) error {
	if _, exists := r.clients[name]; !exists {
		return fmt.Errorf("client %q not found", name)
	}

	delete(r.clients, name)
	return nil
}

// GetClient gets an MCP client by name
func (r *MCPRegistry) GetClient(name string) (*MCPClient, error) {
	client, exists := r.clients[name]
	if !exists {
		return nil, fmt.Errorf("client %q not found", name)
	}

	return client, nil
}

// ListClients lists all registered clients
func (r *MCPRegistry) ListClients() []string {
	var names []string
	for name := range r.clients {
		names = append(names, name)
	}
	return names
}

// ConnectAll connects all registered clients
func (r *MCPRegistry) ConnectAll(ctx context.Context) error {
	for _, client := range r.clients {
		if err := client.Connect(ctx); err != nil {
			return fmt.Errorf("failed to connect %q: %w", client.name, err)
		}
	}
	return nil
}

// DisconnectAll disconnects all registered clients
func (r *MCPRegistry) DisconnectAll(ctx context.Context) error {
	for _, client := range r.clients {
		if client.connected {
			if err := client.Disconnect(ctx); err != nil {
				return fmt.Errorf("failed to disconnect %q: %w", client.name, err)
			}
		}
	}
	return nil
}

// ListAllTools lists all available tools from all connected servers
func (r *MCPRegistry) ListAllTools(ctx context.Context) (map[string][]*MCPTool, error) {
	result := make(map[string][]*MCPTool)

	for name, client := range r.clients {
		if !client.connected {
			continue
		}

		tools, err := client.ListTools(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list tools from %q: %w", name, err)
		}

		result[name] = tools
	}

	return result, nil
}

// CallToolAny calls a tool on any connected server that has it
func (r *MCPRegistry) CallToolAny(ctx context.Context, toolName string, params map[string]interface{}) (interface{}, error) {
	for _, client := range r.clients {
		if !client.connected {
			continue
		}

		// Try to call the tool
		result, err := client.CallTool(ctx, toolName, params)
		if err == nil {
			return result, nil
		}
	}

	return nil, fmt.Errorf("tool %q not found in any connected server", toolName)
}
