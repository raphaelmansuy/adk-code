package agents

import (
	"fmt"
	"time"
)

// ExecutionStrategy defines an interface for different execution methods
type ExecutionStrategy interface {
	// Execute runs an agent with the given context
	Execute(ctx ExecutionContext) (*ExecutionResult, error)

	// Name returns the name of the execution strategy
	Name() string

	// Description returns a description of the execution strategy
	Description() string

	// Validate checks if the strategy can execute the given agent
	Validate(agent *Agent) error
}

// DirectExecutionStrategy executes agents directly on the host system
type DirectExecutionStrategy struct {
	name        string
	description string
}

// NewDirectExecutionStrategy creates a new direct execution strategy
func NewDirectExecutionStrategy() *DirectExecutionStrategy {
	return &DirectExecutionStrategy{
		name:        "direct",
		description: "Executes agents directly on the host system",
	}
}

// Execute runs an agent directly
func (s *DirectExecutionStrategy) Execute(ctx ExecutionContext) (*ExecutionResult, error) {
	result := &ExecutionResult{
		StartTime: time.Now(),
	}

	// Validate agent
	if err := ctx.Agent.Validate(); err != nil {
		return nil, fmt.Errorf("agent validation failed: %w", err)
	}

	// For now, return a placeholder execution result
	// In a real implementation, this would execute the agent
	result.Output = fmt.Sprintf("Direct execution: %s\n", ctx.Agent.Name)
	result.ExitCode = 0
	result.Success = true
	result.Duration = 100 * time.Millisecond

	result.EndTime = result.StartTime.Add(result.Duration)
	return result, nil
}

// Name returns the strategy name
func (s *DirectExecutionStrategy) Name() string {
	return s.name
}

// Description returns the strategy description
func (s *DirectExecutionStrategy) Description() string {
	return s.description
}

// Validate checks if the strategy can execute the agent
func (s *DirectExecutionStrategy) Validate(agent *Agent) error {
	if agent == nil {
		return fmt.Errorf("agent is nil")
	}
	return nil
}

// DockerConfig contains configuration for Docker execution
type DockerConfig struct {
	// Image is the Docker image to use for execution
	Image string

	// Tag is the Docker image tag
	Tag string

	// Registry is the Docker registry (e.g., docker.io)
	Registry string

	// Memory is the memory limit in MB
	Memory int

	// CPUs is the CPU limit (e.g., 0.5, 1.0, 2.0)
	CPUs string

	// Timeout is the execution timeout
	Timeout time.Duration

	// Network is the Docker network mode (bridge, host, none, container:name)
	Network string

	// Volumes maps host paths to container paths
	Volumes map[string]string

	// Environment variables to pass to the container
	Environment map[string]string

	// WorkDir is the working directory inside the container
	WorkDir string

	// Privileged indicates whether to run in privileged mode
	Privileged bool

	// RemoveAfterExecution indicates whether to remove the container after execution
	RemoveAfterExecution bool
}

// NewDockerConfig creates a new Docker configuration with defaults
func NewDockerConfig() *DockerConfig {
	return &DockerConfig{
		Image:                "python:3.11",
		Tag:                  "latest",
		Registry:             "docker.io",
		Memory:               512,
		CPUs:                 "1.0",
		Timeout:              30 * time.Second,
		Network:              "bridge",
		Volumes:              make(map[string]string),
		Environment:          make(map[string]string),
		WorkDir:              "/workspace",
		Privileged:           false,
		RemoveAfterExecution: true,
	}
}

// DockerExecutionStrategy executes agents inside Docker containers
type DockerExecutionStrategy struct {
	config *DockerConfig
}

// NewDockerExecutionStrategy creates a new Docker execution strategy
func NewDockerExecutionStrategy(config *DockerConfig) *DockerExecutionStrategy {
	if config == nil {
		config = NewDockerConfig()
	}
	return &DockerExecutionStrategy{
		config: config,
	}
}

// Execute runs an agent inside a Docker container
func (s *DockerExecutionStrategy) Execute(ctx ExecutionContext) (*ExecutionResult, error) {
	result := &ExecutionResult{
		StartTime: time.Now(),
	}

	// Validate agent
	if err := ctx.Agent.Validate(); err != nil {
		return nil, fmt.Errorf("agent validation failed: %w", err)
	}

	// Validate Docker config
	if err := s.ValidateDockerConfig(); err != nil {
		return nil, fmt.Errorf("Docker config validation failed: %w", err)
	}

	// For now, return a placeholder execution result
	// In a real implementation, this would create and run a Docker container
	result.Output = fmt.Sprintf("Docker execution: %s (image: %s)\n", ctx.Agent.Name, s.FullImageName())
	result.ExitCode = 0
	result.Success = true
	result.Duration = 200 * time.Millisecond

	result.EndTime = result.StartTime.Add(result.Duration)
	return result, nil
}

// Name returns the strategy name
func (s *DockerExecutionStrategy) Name() string {
	return "docker"
}

// Description returns the strategy description
func (s *DockerExecutionStrategy) Description() string {
	return "Executes agents inside isolated Docker containers"
}

// Validate checks if the strategy can execute the agent
func (s *DockerExecutionStrategy) Validate(agent *Agent) error {
	if agent == nil {
		return fmt.Errorf("agent is nil")
	}

	if err := s.ValidateDockerConfig(); err != nil {
		return err
	}

	return nil
}

// ValidateDockerConfig validates the Docker configuration
func (s *DockerExecutionStrategy) ValidateDockerConfig() error {
	if s.config.Image == "" {
		return fmt.Errorf("Docker image is required")
	}

	if s.config.Memory <= 0 {
		return fmt.Errorf("Docker memory limit must be greater than 0")
	}

	if s.config.Timeout <= 0 {
		return fmt.Errorf("Docker timeout must be greater than 0")
	}

	if s.config.CPUs == "" {
		return fmt.Errorf("Docker CPU limit is required")
	}

	return nil
}

// FullImageName returns the full Docker image name with registry
func (s *DockerExecutionStrategy) FullImageName() string {
	image := s.config.Image
	if s.config.Tag != "" {
		image = fmt.Sprintf("%s:%s", image, s.config.Tag)
	}
	if s.config.Registry != "" {
		image = fmt.Sprintf("%s/%s", s.config.Registry, image)
	}
	return image
}

// SetMemory sets the memory limit
func (s *DockerExecutionStrategy) SetMemory(mb int) error {
	if mb <= 0 {
		return fmt.Errorf("memory limit must be greater than 0")
	}
	s.config.Memory = mb
	return nil
}

// SetCPU sets the CPU limit
func (s *DockerExecutionStrategy) SetCPU(cpus string) error {
	if cpus == "" {
		return fmt.Errorf("CPU limit cannot be empty")
	}
	s.config.CPUs = cpus
	return nil
}

// SetTimeout sets the execution timeout
func (s *DockerExecutionStrategy) SetTimeout(timeout time.Duration) error {
	if timeout <= 0 {
		return fmt.Errorf("timeout must be greater than 0")
	}
	s.config.Timeout = timeout
	return nil
}

// AddVolume adds a volume mount
func (s *DockerExecutionStrategy) AddVolume(hostPath, containerPath string) error {
	if hostPath == "" || containerPath == "" {
		return fmt.Errorf("both host path and container path are required")
	}
	s.config.Volumes[hostPath] = containerPath
	return nil
}

// AddEnvironment adds an environment variable
func (s *DockerExecutionStrategy) AddEnvironment(key, value string) error {
	if key == "" {
		return fmt.Errorf("environment variable key cannot be empty")
	}
	s.config.Environment[key] = value
	return nil
}

// ExecutionManager manages execution strategies
type ExecutionManager struct {
	strategies map[string]ExecutionStrategy
	default_   ExecutionStrategy
}

// NewExecutionManager creates a new execution manager
func NewExecutionManager() *ExecutionManager {
	manager := &ExecutionManager{
		strategies: make(map[string]ExecutionStrategy),
		default_:   NewDirectExecutionStrategy(),
	}

	// Register default strategies
	manager.Register(NewDirectExecutionStrategy())
	manager.Register(NewDockerExecutionStrategy(nil))

	return manager
}

// Register registers an execution strategy
func (em *ExecutionManager) Register(strategy ExecutionStrategy) error {
	if strategy == nil {
		return fmt.Errorf("strategy is nil")
	}

	if strategy.Name() == "" {
		return fmt.Errorf("strategy name cannot be empty")
	}

	em.strategies[strategy.Name()] = strategy
	return nil
}

// Get gets an execution strategy by name
func (em *ExecutionManager) Get(name string) (ExecutionStrategy, error) {
	if name == "" {
		return em.default_, nil
	}

	strategy, ok := em.strategies[name]
	if !ok {
		return nil, fmt.Errorf("execution strategy %q not found", name)
	}

	return strategy, nil
}

// SetDefault sets the default execution strategy
func (em *ExecutionManager) SetDefault(name string) error {
	strategy, err := em.Get(name)
	if err != nil {
		return err
	}

	em.default_ = strategy
	return nil
}

// List returns a list of registered strategies
func (em *ExecutionManager) List() []ExecutionStrategy {
	var strategies []ExecutionStrategy
	for _, strategy := range em.strategies {
		strategies = append(strategies, strategy)
	}
	return strategies
}

// Execute executes an agent using the specified strategy
func (em *ExecutionManager) Execute(ctx ExecutionContext, strategyName string) (*ExecutionResult, error) {
	strategy, err := em.Get(strategyName)
	if err != nil {
		return nil, err
	}

	return strategy.Execute(ctx)
}
