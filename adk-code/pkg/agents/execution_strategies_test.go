package agents

import (
	"testing"
	"time"
)

// Test Direct Execution Strategy
func TestNewDirectExecutionStrategy(t *testing.T) {
	strategy := NewDirectExecutionStrategy()

	if strategy == nil {
		t.Fatal("Strategy is nil")
	}

	if strategy.Name() != "direct" {
		t.Fatalf("Expected name 'direct', got '%s'", strategy.Name())
	}

	if strategy.Description() == "" {
		t.Fatal("Expected description, got empty string")
	}
}

func TestDirectExecutionStrategyValidate(t *testing.T) {
	strategy := NewDirectExecutionStrategy()
	agent := &Agent{
		Name: "test-agent",
	}

	err := strategy.Validate(agent)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestDirectExecutionStrategyValidateNil(t *testing.T) {
	strategy := NewDirectExecutionStrategy()

	err := strategy.Validate(nil)
	if err == nil {
		t.Fatal("Expected error for nil agent")
	}
}

// Test Docker Config
func TestNewDockerConfig(t *testing.T) {
	config := NewDockerConfig()

	if config.Image == "" {
		t.Fatal("Expected default image")
	}

	if config.Memory <= 0 {
		t.Fatal("Expected positive memory limit")
	}

	if config.CPUs == "" {
		t.Fatal("Expected CPU limit")
	}

	if config.Timeout == 0 {
		t.Fatal("Expected timeout")
	}
}

// Test Docker Execution Strategy
func TestNewDockerExecutionStrategy(t *testing.T) {
	strategy := NewDockerExecutionStrategy(nil)

	if strategy == nil {
		t.Fatal("Strategy is nil")
	}

	if strategy.Name() != "docker" {
		t.Fatalf("Expected name 'docker', got '%s'", strategy.Name())
	}

	if strategy.Description() == "" {
		t.Fatal("Expected description, got empty string")
	}
}

func TestDockerExecutionStrategyFullImageName(t *testing.T) {
	tests := []struct {
		name     string
		registry string
		image    string
		tag      string
		expected string
	}{
		{
			name:     "simple",
			registry: "docker.io",
			image:    "python",
			tag:      "3.11",
			expected: "docker.io/python:3.11",
		},
		{
			name:     "no registry",
			registry: "",
			image:    "python",
			tag:      "3.11",
			expected: "python:3.11",
		},
		{
			name:     "no tag",
			registry: "docker.io",
			image:    "python",
			tag:      "",
			expected: "docker.io/python",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewDockerConfig()
			config.Registry = tt.registry
			config.Image = tt.image
			config.Tag = tt.tag

			strategy := NewDockerExecutionStrategy(config)
			result := strategy.FullImageName()

			if result != tt.expected {
				t.Fatalf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestDockerExecutionStrategyValidate(t *testing.T) {
	strategy := NewDockerExecutionStrategy(nil)
	agent := &Agent{
		Name: "test-agent",
	}

	err := strategy.Validate(agent)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestDockerExecutionStrategySetMemory(t *testing.T) {
	strategy := NewDockerExecutionStrategy(nil)

	err := strategy.SetMemory(1024)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if strategy.config.Memory != 1024 {
		t.Fatalf("Expected memory 1024, got %d", strategy.config.Memory)
	}

	// Test invalid memory
	err = strategy.SetMemory(-1)
	if err == nil {
		t.Fatal("Expected error for negative memory")
	}
}

func TestDockerExecutionStrategySetCPU(t *testing.T) {
	strategy := NewDockerExecutionStrategy(nil)

	err := strategy.SetCPU("2.0")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if strategy.config.CPUs != "2.0" {
		t.Fatalf("Expected CPUs '2.0', got '%s'", strategy.config.CPUs)
	}

	// Test invalid CPU
	err = strategy.SetCPU("")
	if err == nil {
		t.Fatal("Expected error for empty CPU")
	}
}

func TestDockerExecutionStrategySetTimeout(t *testing.T) {
	strategy := NewDockerExecutionStrategy(nil)

	timeout := 5 * time.Minute
	err := strategy.SetTimeout(timeout)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if strategy.config.Timeout != timeout {
		t.Fatalf("Expected timeout %v, got %v", timeout, strategy.config.Timeout)
	}

	// Test invalid timeout
	err = strategy.SetTimeout(0)
	if err == nil {
		t.Fatal("Expected error for zero timeout")
	}
}

func TestDockerExecutionStrategyAddVolume(t *testing.T) {
	strategy := NewDockerExecutionStrategy(nil)

	err := strategy.AddVolume("/host/path", "/container/path")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if strategy.config.Volumes["/host/path"] != "/container/path" {
		t.Fatal("Expected volume to be added")
	}

	// Test invalid volume
	err = strategy.AddVolume("", "/container/path")
	if err == nil {
		t.Fatal("Expected error for empty host path")
	}
}

func TestDockerExecutionStrategyAddEnvironment(t *testing.T) {
	strategy := NewDockerExecutionStrategy(nil)

	err := strategy.AddEnvironment("TEST_VAR", "test_value")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if strategy.config.Environment["TEST_VAR"] != "test_value" {
		t.Fatal("Expected environment variable to be added")
	}

	// Test invalid environment
	err = strategy.AddEnvironment("", "test_value")
	if err == nil {
		t.Fatal("Expected error for empty key")
	}
}

// Test Execution Manager
func TestNewExecutionManager(t *testing.T) {
	manager := NewExecutionManager()

	if manager == nil {
		t.Fatal("Manager is nil")
	}

	strategies := manager.List()
	if len(strategies) != 2 {
		t.Fatalf("Expected 2 default strategies, got %d", len(strategies))
	}
}

func TestExecutionManagerRegister(t *testing.T) {
	manager := NewExecutionManager()

	// Register a new strategy
	custom := &CustomStrategy{
		name: "custom",
		desc: "Custom strategy",
	}

	err := manager.Register(custom)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	strategy, err := manager.Get("custom")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if strategy.Name() != "custom" {
		t.Fatalf("Expected 'custom', got '%s'", strategy.Name())
	}
}

func TestExecutionManagerGet(t *testing.T) {
	manager := NewExecutionManager()

	// Get direct strategy
	strategy, err := manager.Get("direct")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if strategy.Name() != "direct" {
		t.Fatalf("Expected 'direct', got '%s'", strategy.Name())
	}

	// Get docker strategy
	strategy, err = manager.Get("docker")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if strategy.Name() != "docker" {
		t.Fatalf("Expected 'docker', got '%s'", strategy.Name())
	}

	// Get non-existent strategy
	_, err = manager.Get("nonexistent")
	if err == nil {
		t.Fatal("Expected error for non-existent strategy")
	}
}

func TestExecutionManagerSetDefault(t *testing.T) {
	manager := NewExecutionManager()

	err := manager.SetDefault("docker")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if manager.default_.Name() != "docker" {
		t.Fatalf("Expected 'docker', got '%s'", manager.default_.Name())
	}

	// Test invalid default
	err = manager.SetDefault("nonexistent")
	if err == nil {
		t.Fatal("Expected error for non-existent strategy")
	}
}

func TestExecutionManagerList(t *testing.T) {
	manager := NewExecutionManager()

	strategies := manager.List()
	if len(strategies) == 0 {
		t.Fatal("Expected strategies, got empty list")
	}

	// Verify at least direct and docker are present
	hasNames := make(map[string]bool)
	for _, s := range strategies {
		hasNames[s.Name()] = true
	}

	if !hasNames["direct"] {
		t.Fatal("Expected 'direct' strategy")
	}

	if !hasNames["docker"] {
		t.Fatal("Expected 'docker' strategy")
	}
}

// Custom strategy for testing
type CustomStrategy struct {
	name string
	desc string
}

func (s *CustomStrategy) Execute(ctx ExecutionContext) (*ExecutionResult, error) {
	return &ExecutionResult{
		Success:   true,
		Output:    "Custom execution",
		ExitCode:  0,
		StartTime: time.Now(),
		EndTime:   time.Now(),
		Duration:  1 * time.Millisecond,
	}, nil
}

func (s *CustomStrategy) Name() string {
	return s.name
}

func (s *CustomStrategy) Description() string {
	return s.desc
}

func (s *CustomStrategy) Validate(agent *Agent) error {
	if agent == nil {
		return ErrInvalidAgent
	}
	return nil
}

// Error definitions for tests
var ErrInvalidAgent = &ValidationError{
	Message: "invalid agent",
	Field:   "agent",
}

// ValidationError is a custom error type for validation
type ValidationError struct {
	Message string
	Field   string
}

func (e *ValidationError) Error() string {
	return e.Message
}
