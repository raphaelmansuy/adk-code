package agents

import (
	"context"
	"iter"
	"os"
	"path/filepath"
	"testing"

	"google.golang.org/adk/model"

	"adk-code/pkg/models"
)

// mockLLM is a minimal mock for testing
type mockLLM struct{}

func (m *mockLLM) Name() string {
	return "mock-llm"
}

func (m *mockLLM) GenerateContent(ctx context.Context, request *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		yield(&model.LLMResponse{}, nil)
	}
}

func TestNewSubAgentManager(t *testing.T) {
	modelConfig := models.Config{Name: "test-model", ContextWindow: 100000}
	manager := NewSubAgentManager("/tmp", &mockLLM{}, modelConfig)
	if manager == nil {
		t.Fatal("NewSubAgentManager returned nil")
	}
	if manager.projectRoot != "/tmp" {
		t.Errorf("Expected projectRoot=/tmp, got %s", manager.projectRoot)
	}
}

func TestLoadSubAgentTools_NoAgents(t *testing.T) {
	tmpDir := t.TempDir()
	modelConfig := models.Config{Name: "test-model", ContextWindow: 100000}
	manager := NewSubAgentManager(tmpDir, &mockLLM{}, modelConfig)

	ctx := context.Background()
	tools, err := manager.LoadSubAgentTools(ctx)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(tools) != 0 {
		t.Errorf("Expected 0 tools, got %d", len(tools))
	}
}

func TestLoadSubAgentTools_WithAgents(t *testing.T) {
	// Create temp directory with agent definitions
	tmpDir := t.TempDir()
	agentsDir := filepath.Join(tmpDir, ".adk", "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents directory: %v", err)
	}

	// Create a test agent
	agentContent := `---
name: test-agent
description: A test agent for unit testing
---

You are a test agent.`

	agentPath := filepath.Join(agentsDir, "test-agent.md")
	if err := os.WriteFile(agentPath, []byte(agentContent), 0644); err != nil {
		t.Fatalf("Failed to write agent file: %v", err)
	}

	modelConfig := models.Config{Name: "test-model", ContextWindow: 100000}
	manager := NewSubAgentManager(tmpDir, &mockLLM{}, modelConfig)

	ctx := context.Background()
	tools, err := manager.LoadSubAgentTools(ctx)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(tools))
	}
}

func TestInitSubAgentTools(t *testing.T) {
	// This is mainly a smoke test to ensure the convenience function works
	tmpDir := t.TempDir()

	ctx := context.Background()
	modelConfig := models.Config{Name: "test-model", ContextWindow: 100000}
	tools, err := InitSubAgentTools(ctx, tmpDir, &mockLLM{}, modelConfig)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	// Should return empty list for empty directory
	if tools == nil {
		t.Error("Expected non-nil tools slice")
	}
}
