package mcp

import (
	"context"
	"testing"

	"adk-code/internal/config"
)

func TestManagerNewCreatesEmptyStructure(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Error("expected non-nil manager")
	}
	if len(m.Toolsets()) != 0 {
		t.Error("expected empty toolsets")
	}
	if len(m.List()) != 0 {
		t.Error("expected empty server list")
	}
}

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

func TestListServersEmpty(t *testing.T) {
	m := NewManager()
	cfg := &config.MCPConfig{Enabled: true}
	m.LoadServers(context.Background(), cfg)
	list := m.List()
	if len(list) != 0 {
		t.Error("expected empty list")
	}
}

func TestStatusEmpty(t *testing.T) {
	m := NewManager()
	status := m.Status()
	if len(status) != 0 {
		t.Error("expected empty status map")
	}
}

func TestCreateTransportStdioValid(t *testing.T) {
	cfg := config.ServerConfig{
		Type:    "stdio",
		Command: "echo",
	}
	transport, err := createTransport(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if transport == nil {
		t.Error("expected transport to be created")
	}
}

func TestCreateTransportStdioMissingCommand(t *testing.T) {
	cfg := config.ServerConfig{
		Type: "stdio",
	}
	_, err := createTransport(cfg)
	if err == nil {
		t.Error("expected error for missing command")
	}
}

func TestCreateTransportSSEValid(t *testing.T) {
	cfg := config.ServerConfig{
		Type: "sse",
		URL:  "http://localhost:3000",
	}
	transport, err := createTransport(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if transport == nil {
		t.Error("expected transport to be created")
	}
}

func TestCreateTransportSSEMissingURL(t *testing.T) {
	cfg := config.ServerConfig{
		Type: "sse",
	}
	_, err := createTransport(cfg)
	if err == nil {
		t.Error("expected error for missing URL")
	}
}

func TestCreateTransportStreamableValid(t *testing.T) {
	cfg := config.ServerConfig{
		Type: "streamable",
		URL:  "http://localhost:3000/mcp",
	}
	transport, err := createTransport(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if transport == nil {
		t.Error("expected transport to be created")
	}
}

func TestCreateTransportStreamableMissingURL(t *testing.T) {
	cfg := config.ServerConfig{
		Type: "streamable",
	}
	_, err := createTransport(cfg)
	if err == nil {
		t.Error("expected error for missing URL")
	}
}

func TestCreateTransportInvalidType(t *testing.T) {
	cfg := config.ServerConfig{
		Type: "invalid",
	}
	_, err := createTransport(cfg)
	if err == nil {
		t.Error("expected error for invalid type")
	}
}
