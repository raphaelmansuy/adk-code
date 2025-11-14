package agents

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ExecutionContext contains parameters and configuration for agent execution.
type ExecutionContext struct {
	// Agent is the agent to execute
	Agent *Agent

	// Params contains input parameters for the agent
	Params map[string]interface{}

	// Timeout is the maximum execution time (0 = no timeout)
	Timeout time.Duration

	// WorkDir is the working directory for execution
	WorkDir string

	// Env contains additional environment variables
	Env map[string]string

	// CaptureOutput indicates whether to capture stdout/stderr
	CaptureOutput bool

	// ReturnRawOutput returns raw output without formatting
	ReturnRawOutput bool

	// Context for cancellation
	Context context.Context
}

// ExecutionResult contains the result of agent execution.
type ExecutionResult struct {
	// Output is the captured output from the agent
	Output string

	// Error is the error message if execution failed
	Error string

	// ExitCode is the exit code from the agent process
	ExitCode int

	// Duration is the execution time
	Duration time.Duration

	// Success indicates whether execution completed successfully
	Success bool

	// Stderr contains stderr output if captured separately
	Stderr string

	// StartTime is when execution started
	StartTime time.Time

	// EndTime is when execution completed
	EndTime time.Time
}

// ExecutionRequirements defines system requirements for agent execution.
type ExecutionRequirements struct {
	// SupportedOS is a list of supported operating systems (linux, darwin, windows)
	SupportedOS []string

	// MinGoVersion is the minimum required Go version
	MinGoVersion string

	// MinMemoryMB is the minimum memory required in MB
	MinMemoryMB int

	// TimeoutSeconds is the default execution timeout
	TimeoutSeconds int

	// RequiredEnv is a list of required environment variables
	RequiredEnv []string

	// Features is a list of required features (e.g., file-io, network)
	Features []string
}

// Executor defines the interface for agent execution.
type Executor interface {
	// Execute runs an agent with the given context and returns the result
	Execute(ctx ExecutionContext) (*ExecutionResult, error)

	// ValidateRequirements checks if the system meets the agent's requirements
	ValidateRequirements(req *ExecutionRequirements) error
}

// AgentRunner implements the Executor interface for executing agents.
type AgentRunner struct {
	// Discoverer is used to find agents
	discoverer *Discoverer

	// BaseWorkDir is the base directory for execution
	baseWorkDir string
}

// NewAgentRunner creates a new AgentRunner with a discoverer.
func NewAgentRunner(discoverer *Discoverer) *AgentRunner {
	return &AgentRunner{
		discoverer:  discoverer,
		baseWorkDir: os.TempDir(),
	}
}

// NewAgentRunnerWithWorkDir creates an AgentRunner with a specific base work directory.
func NewAgentRunnerWithWorkDir(discoverer *Discoverer, baseWorkDir string) *AgentRunner {
	return &AgentRunner{
		discoverer:  discoverer,
		baseWorkDir: baseWorkDir,
	}
}

// Execute runs the agent and returns the result.
func (r *AgentRunner) Execute(ctx ExecutionContext) (*ExecutionResult, error) {
	startTime := time.Now()
	result := &ExecutionResult{
		StartTime: startTime,
		ExitCode:  -1,
	}

	// Use provided context or create a timeout context
	execCtx := ctx.Context
	if execCtx == nil {
		execCtx = context.Background()
	}

	// Apply timeout if specified
	if ctx.Timeout > 0 {
		var cancel context.CancelFunc
		execCtx, cancel = context.WithTimeout(execCtx, ctx.Timeout)
		defer cancel()
	}

	// Validate agent exists
	if ctx.Agent == nil {
		result.Error = "agent is nil"
		result.Success = false
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(startTime)
		return result, fmt.Errorf("agent is nil")
	}

	// Validate agent has executable path
	if ctx.Agent.Path == "" {
		result.Error = "agent path is empty"
		result.Success = false
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(startTime)
		return result, fmt.Errorf("agent path is empty")
	}

	// Set working directory
	workDir := ctx.WorkDir
	if workDir == "" {
		workDir = r.baseWorkDir
	}

	// Create command
	cmd := exec.CommandContext(execCtx, ctx.Agent.Path)
	cmd.Dir = workDir

	// Set environment
	cmd.Env = os.Environ()
	if ctx.Env != nil {
		for k, v := range ctx.Env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// Add parameters as arguments (simple string conversion)
	if ctx.Params != nil {
		for k, v := range ctx.Params {
			cmd.Args = append(cmd.Args, fmt.Sprintf("--%s=%v", k, v))
		}
	}

	// Capture output if requested
	var stdout, stderr bytes.Buffer
	if ctx.CaptureOutput {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}

	// Execute
	err := cmd.Run()

	// Set result fields
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(startTime)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		if ctx.CaptureOutput {
			result.Error = fmt.Sprintf("execution failed: %v\nstderr: %s", err, stderr.String())
		}
		// Extract exit code if possible
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
	} else {
		result.Success = true
		result.ExitCode = 0
	}

	// Set output
	if ctx.CaptureOutput {
		result.Output = stdout.String()
		result.Stderr = stderr.String()
	}

	return result, nil
}

// ValidateRequirements checks if the system meets the agent's requirements.
func (r *AgentRunner) ValidateRequirements(req *ExecutionRequirements) error {
	if req == nil {
		return nil
	}

	// Check supported OS (simplified - would expand for real implementation)
	if len(req.SupportedOS) > 0 {
		currentOS := os.Getenv("GOOS")
		if currentOS == "" {
			// Fallback to hostname-based detection
			currentOS = "linux" // simplified
		}

		found := false
		for _, os := range req.SupportedOS {
			if os == currentOS {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("current OS %q not in supported list: %v", currentOS, req.SupportedOS)
		}
	}

	// Check required environment variables
	if len(req.RequiredEnv) > 0 {
		for _, envVar := range req.RequiredEnv {
			if _, exists := os.LookupEnv(envVar); !exists {
				return fmt.Errorf("required environment variable %q not set", envVar)
			}
		}
	}

	return nil
}

// GetExecutionRequirements extracts execution requirements from an agent.
func (a *Agent) GetExecutionRequirements() *ExecutionRequirements {
	if a == nil {
		return nil
	}

	// Parse requirements from agent metadata/fields
	// This will be populated from YAML parsing in the metadata system
	return &ExecutionRequirements{
		TimeoutSeconds: 300, // default 5 minutes
	}
}

// Validate checks if an agent is valid for execution.
func (a *Agent) Validate() error {
	if a == nil {
		return fmt.Errorf("agent is nil")
	}

	if a.Name == "" {
		return fmt.Errorf("agent name is empty")
	}

	if a.Path == "" {
		return fmt.Errorf("agent path is empty")
	}

	// Check if path exists
	info, err := os.Stat(a.Path)
	if err != nil {
		return fmt.Errorf("agent path does not exist: %w", err)
	}

	// Check if path is executable (simplified check)
	if info.IsDir() {
		return fmt.Errorf("agent path is a directory, not a file")
	}

	return nil
}

// ValidateParameters validates input parameters against agent expectations.
func (a *Agent) ValidateParameters(params map[string]interface{}) error {
	if a == nil {
		return fmt.Errorf("agent is nil")
	}

	// This would expand to validate against agent schema
	// For now, we accept any parameters
	return nil
}

// FormatOutput formats the execution output based on context.
func FormatOutput(result *ExecutionResult, returnRaw bool) string {
	if returnRaw || result.Output == "" {
		return result.Output
	}

	var output strings.Builder
	output.WriteString("=== Agent Execution Result ===\n")
	output.WriteString(fmt.Sprintf("Success: %v\n", result.Success))
	output.WriteString(fmt.Sprintf("Exit Code: %d\n", result.ExitCode))
	output.WriteString(fmt.Sprintf("Duration: %v\n", result.Duration))
	output.WriteString("--- Output ---\n")
	output.WriteString(result.Output)
	if result.Stderr != "" {
		output.WriteString("\n--- Errors ---\n")
		output.WriteString(result.Stderr)
	}

	return output.String()
}

// ExecuteAndStream executes an agent and returns a channel for streaming results.
func (r *AgentRunner) ExecuteAndStream(ctx ExecutionContext) <-chan *ExecutionResult {
	resultChan := make(chan *ExecutionResult, 1)

	go func() {
		defer close(resultChan)
		result, _ := r.Execute(ctx)
		resultChan <- result
	}()

	return resultChan
}

// GetAgentByName retrieves an agent by name using the discoverer.
func (r *AgentRunner) GetAgentByName(name string) (*Agent, error) {
	if r.discoverer == nil {
		return nil, fmt.Errorf("discoverer is not configured")
	}

	result, err := r.discoverer.DiscoverAll()
	if err != nil {
		return nil, fmt.Errorf("failed to discover agents: %w", err)
	}

	for _, agent := range result.Agents {
		if agent.Name == name {
			return agent, nil
		}
	}

	return nil, fmt.Errorf("agent %q not found", name)
}

// ExpandPath expands a path with home directory support.
func ExpandPath(p string) string {
	if strings.HasPrefix(p, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return p
		}
		return filepath.Join(homeDir, p[1:])
	}
	return p
}
