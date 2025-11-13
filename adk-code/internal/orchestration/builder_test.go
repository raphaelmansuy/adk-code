package orchestration

import (
	"context"
	"testing"

	"adk-code/internal/config"
)

// TestNewOrchestrator verifies Orchestrator creation
func TestNewOrchestrator(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{
		APIKey: "test-key",
		Model:  "gemini/2.5-flash",
	}

	orchestrator := NewOrchestrator(ctx, cfg)
	if orchestrator == nil {
		t.Fatalf("NewOrchestrator returned nil")
	}
	if orchestrator.ctx != ctx {
		t.Errorf("context not set correctly")
	}
	if orchestrator.cfg != cfg {
		t.Errorf("config not set correctly")
	}
}

// TestOrchestratorFluent verifies fluent API chaining
func TestOrchestratorFluent(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{
		APIKey: "test-key",
		Model:  "gemini/2.5-flash",
	}

	// Test that methods return *Orchestrator for chaining
	result := NewOrchestrator(ctx, cfg).
		WithDisplay().
		WithModel().
		WithAgent().
		WithSession()

	if result == nil {
		t.Fatalf("fluent chain returned nil")
	}
}

// TestOrchestratorWithDisplay verifies display component initialization
func TestOrchestratorWithDisplay(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{
		OutputFormat: "text",
	}

	orchestrator := NewOrchestrator(ctx, cfg)
	orchestrator.WithDisplay()

	if orchestrator.err != nil {
		t.Fatalf("WithDisplay failed: %v", orchestrator.err)
	}
	if orchestrator.displayComponents == nil {
		t.Errorf("display components not initialized")
	}
}

// TestOrchestratorWithModel verifies model component initialization
func TestOrchestratorWithModel(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{
		APIKey: "test-key",
		Model:  "gemini/2.5-flash",
	}

	orchestrator := NewOrchestrator(ctx, cfg)
	orchestrator.WithModel()

	if orchestrator.err != nil {
		t.Fatalf("WithModel failed: %v", orchestrator.err)
	}
	if orchestrator.modelComponents == nil {
		t.Errorf("model components not initialized")
	}
}

// TestOrchestratorWithAgent verifies agent component initialization
func TestOrchestratorWithAgent(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{
		APIKey: "test-key",
		Model:  "gemini/2.5-flash",
	}

	orchestrator := NewOrchestrator(ctx, cfg).
		WithModel().
		WithAgent()

	if orchestrator.err != nil {
		t.Fatalf("WithAgent failed: %v", orchestrator.err)
	}
	if orchestrator.agentComponent == nil {
		t.Errorf("agent component not initialized")
	}
}

// TestOrchestratorWithSession verifies session component initialization
func TestOrchestratorWithSession(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{
		APIKey:      "test-key",
		Model:       "gemini/2.5-flash",
		SessionName: "test-session",
		DBPath:      ":memory:",
	}

	orchestrator := NewOrchestrator(ctx, cfg).
		WithDisplay().
		WithModel().
		WithAgent().
		WithSession()

	if orchestrator.err != nil {
		t.Fatalf("WithSession failed: %v", orchestrator.err)
	}
	if orchestrator.sessionComponents == nil {
		t.Errorf("session components not initialized")
	}
}

// TestOrchestratorBuildSuccess verifies successful component building
func TestOrchestratorBuildSuccess(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{
		APIKey:      "test-key",
		Model:       "gemini/2.5-flash",
		SessionName: "test-session",
		DBPath:      ":memory:",
	}

	components, err := NewOrchestrator(ctx, cfg).
		WithDisplay().
		WithModel().
		WithAgent().
		WithSession().
		Build()

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	if components == nil {
		t.Fatalf("Build returned nil components")
	}
	if components.Display == nil {
		t.Errorf("Display component is nil")
	}
	if components.Model == nil {
		t.Errorf("Model component is nil")
	}
	if components.Agent == nil {
		t.Errorf("Agent component is nil")
	}
	if components.Session == nil {
		t.Errorf("Session component is nil")
	}
}

// TestOrchestratorBuildMissingDisplay verifies error when display not initialized
func TestOrchestratorBuildMissingDisplay(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{
		APIKey:      "test-key",
		Model:       "gemini/2.5-flash",
		SessionName: "test-session",
		DBPath:      ":memory:",
	}

	// Skip display component - error will happen at WithSession step since it requires display
	_, err := NewOrchestrator(ctx, cfg).
		WithModel().
		WithAgent().
		WithSession().
		Build()

	if err == nil {
		t.Fatalf("Build should fail without display component")
	}
	if err.Error() != "session requires display component; call WithDisplay() first" {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestOrchestratorBuildMissingModel verifies error when model not initialized
func TestOrchestratorBuildMissingModel(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{
		DBPath: ":memory:",
	}

	_, err := NewOrchestrator(ctx, cfg).
		WithDisplay().
		WithAgent().
		Build()

	// Should fail when trying to WithAgent because model is missing
	if err == nil {
		t.Fatalf("Build should fail without model component")
	}
}

// TestOrchestratorBuildMissingAgent verifies error when agent not initialized
func TestOrchestratorBuildMissingAgent(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{
		APIKey:      "test-key",
		Model:       "gemini/2.5-flash",
		SessionName: "test-session",
		DBPath:      ":memory:",
	}

	_, err := NewOrchestrator(ctx, cfg).
		WithDisplay().
		WithModel().
		WithSession().
		Build()

	// Should fail when trying to WithSession because agent is missing
	if err == nil {
		t.Fatalf("Build should fail without agent component")
	}
}

// TestOrchestratorAgentRequiresModel verifies dependency checking
func TestOrchestratorAgentRequiresModel(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{
		APIKey: "test-key",
		Model:  "gemini/2.5-flash",
	}

	orchestrator := NewOrchestrator(ctx, cfg).
		WithAgent() // Skip WithModel

	if orchestrator.err == nil {
		t.Fatalf("WithAgent should fail without model component")
	}
	if orchestrator.err.Error() != "agent requires model component; call WithModel() first" {
		t.Errorf("unexpected error: %v", orchestrator.err)
	}
}

// TestOrchestratorSessionRequiresAgent verifies dependency checking
func TestOrchestratorSessionRequiresAgent(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{
		APIKey: "test-key",
		Model:  "gemini/2.5-flash",
		DBPath: ":memory:",
	}

	orchestrator := NewOrchestrator(ctx, cfg).
		WithDisplay().
		WithModel().
		WithSession() // Skip WithAgent

	if orchestrator.err == nil {
		t.Fatalf("WithSession should fail without agent component")
	}
	if orchestrator.err.Error() != "session requires agent component; call WithAgent() first" {
		t.Errorf("unexpected error: %v", orchestrator.err)
	}
}

// TestOrchestratorSessionRequiresDisplay verifies dependency checking
func TestOrchestratorSessionRequiresDisplay(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{
		APIKey:      "test-key",
		Model:       "gemini/2.5-flash",
		SessionName: "test-session",
		DBPath:      ":memory:",
	}

	orchestrator := NewOrchestrator(ctx, cfg).
		WithModel().
		WithAgent().
		WithSession() // Skip WithDisplay

	if orchestrator.err == nil {
		t.Fatalf("WithSession should fail without display component")
	}
	if orchestrator.err.Error() != "session requires display component; call WithDisplay() first" {
		t.Errorf("unexpected error: %v", orchestrator.err)
	}
}

// TestOrchestratorErrorPropagation verifies errors stop the chain
func TestOrchestratorErrorPropagation(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{} // Invalid config

	orchestrator := NewOrchestrator(ctx, cfg).
		WithAgent(). // This will fail
		WithSession()

	if orchestrator.err == nil {
		t.Fatalf("error should propagate through chain")
	}
}

// TestComponentsAccessors verifies component accessor methods
func TestComponentsAccessors(t *testing.T) {
	ctx := context.Background()
	cfg := &config.Config{
		APIKey:      "test-key",
		Model:       "gemini/2.5-flash",
		SessionName: "test-session",
		DBPath:      ":memory:",
	}

	components, err := NewOrchestrator(ctx, cfg).
		WithDisplay().
		WithModel().
		WithAgent().
		WithSession().
		Build()

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Test accessor methods
	if components.DisplayRenderer() == nil {
		t.Errorf("DisplayRenderer() returned nil")
	}
	if components.ModelRegistry() == nil {
		t.Errorf("ModelRegistry() returned nil")
	}
	if components.AgentComponent() == nil {
		t.Errorf("AgentComponent() returned nil")
	}
	if components.SessionManager() == nil {
		t.Errorf("SessionManager() returned nil")
	}
}

// TestOrchestratorContextPropagation verifies context is passed correctly
func TestOrchestratorContextPropagation(t *testing.T) {
	ctx := context.WithValue(context.Background(), "test-key", "test-value")
	cfg := &config.Config{
		APIKey:      "test-key",
		Model:       "gemini/2.5-flash",
		SessionName: "test-session",
		DBPath:      ":memory:",
	}

	orchestrator := NewOrchestrator(ctx, cfg)
	if orchestrator.ctx.Value("test-key") != "test-value" {
		t.Errorf("context value not propagated correctly")
	}
}
