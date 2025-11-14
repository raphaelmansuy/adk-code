package agents

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestExecutionContextCreation tests creating execution contexts
func TestExecutionContextCreation(t *testing.T) {
	agent := &Agent{
		Name:        "test-agent",
		Description: "Test agent",
		Path:        "/tmp/test",
	}

	ctx := ExecutionContext{
		Agent:         agent,
		Timeout:       5 * time.Second,
		CaptureOutput: true,
	}

	if ctx.Agent.Name != "test-agent" {
		t.Errorf("Expected agent name 'test-agent', got %q", ctx.Agent.Name)
	}

	if ctx.Timeout != 5*time.Second {
		t.Errorf("Expected timeout 5s, got %v", ctx.Timeout)
	}
}

// TestExecutionResultFields tests ExecutionResult structure
func TestExecutionResultFields(t *testing.T) {
	now := time.Now()
	later := now.Add(100 * time.Millisecond)

	result := &ExecutionResult{
		Output:    "test output",
		Error:     "",
		ExitCode:  0,
		Success:   true,
		StartTime: now,
		EndTime:   later,
		Duration:  later.Sub(now),
	}

	if result.Output != "test output" {
		t.Errorf("Expected output 'test output', got %q", result.Output)
	}

	if !result.Success {
		t.Error("Expected success true")
	}

	if result.Duration != 100*time.Millisecond {
		t.Errorf("Expected duration 100ms, got %v", result.Duration)
	}
}

// TestAgentValidationNilAgent tests validation with nil agent
func TestAgentValidationNilAgent(t *testing.T) {
	var agent *Agent
	err := agent.Validate()
	if err == nil {
		t.Error("Expected error for nil agent, got nil")
	}
}

// TestAgentValidationEmptyName tests validation with empty name
func TestAgentValidationEmptyName(t *testing.T) {
	agent := &Agent{
		Name: "",
		Path: "/tmp/test",
	}

	err := agent.Validate()
	if err == nil {
		t.Error("Expected error for empty name, got nil")
	}
}

// TestAgentValidationEmptyPath tests validation with empty path
func TestAgentValidationEmptyPath(t *testing.T) {
	agent := &Agent{
		Name: "test",
		Path: "",
	}

	err := agent.Validate()
	if err == nil {
		t.Error("Expected error for empty path, got nil")
	}
}

// TestAgentValidationNonexistentPath tests validation with nonexistent path
func TestAgentValidationNonexistentPath(t *testing.T) {
	agent := &Agent{
		Name: "test",
		Path: "/nonexistent/path/to/agent",
	}

	err := agent.Validate()
	if err == nil {
		t.Error("Expected error for nonexistent path, got nil")
	}
}

// TestAgentValidationDirectoryPath tests validation with directory instead of file
func TestAgentValidationDirectoryPath(t *testing.T) {
	tmpDir := t.TempDir()
	agent := &Agent{
		Name: "test",
		Path: tmpDir,
	}

	err := agent.Validate()
	if err == nil {
		t.Error("Expected error for directory path, got nil")
	}
}

// TestAgentValidationValidAgent tests validation with valid agent
func TestAgentValidationValidAgent(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test-agent")
	f, err := os.Create(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	f.Close()

	agent := &Agent{
		Name: "test",
		Path: tmpFile,
	}

	err = agent.Validate()
	if err != nil {
		t.Errorf("Expected no error for valid agent, got %v", err)
	}
}

// TestAgentParameterValidation tests parameter validation
func TestAgentParameterValidation(t *testing.T) {
	agent := &Agent{
		Name: "test",
	}

	params := map[string]interface{}{
		"param1": "value1",
		"param2": 42,
	}

	err := agent.ValidateParameters(params)
	if err != nil {
		t.Errorf("Expected no error for valid parameters, got %v", err)
	}
}

// TestAgentParameterValidationNilAgent tests parameter validation with nil agent
func TestAgentParameterValidationNilAgent(t *testing.T) {
	var agent *Agent
	err := agent.ValidateParameters(nil)
	if err == nil {
		t.Error("Expected error for nil agent, got nil")
	}
}

// TestExecutionRequirementsValidation tests environment variable checking
func TestExecutionRequirementsValidation(t *testing.T) {
	runner := NewAgentRunner(nil)

	// Test with required env var set
	os.Setenv("TEST_AGENT_VAR", "value")
	defer os.Unsetenv("TEST_AGENT_VAR")

	req := &ExecutionRequirements{
		RequiredEnv: []string{"TEST_AGENT_VAR"},
	}

	err := runner.ValidateRequirements(req)
	if err != nil {
		t.Errorf("Expected no error for set env var, got %v", err)
	}
}

// TestExecutionRequirementsValidationMissingEnv tests with missing env var
func TestExecutionRequirementsValidationMissingEnv(t *testing.T) {
	runner := NewAgentRunner(nil)

	req := &ExecutionRequirements{
		RequiredEnv: []string{"NONEXISTENT_VAR_12345"},
	}

	err := runner.ValidateRequirements(req)
	if err == nil {
		t.Error("Expected error for missing env var, got nil")
	}
}

// TestExecutionContextNilAgent tests execution with nil agent
func TestExecutionContextNilAgent(t *testing.T) {
	runner := NewAgentRunner(nil)

	ctx := ExecutionContext{
		Agent: nil,
	}

	result, err := runner.Execute(ctx)
	if err == nil {
		t.Error("Expected error for nil agent, got nil")
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.Success {
		t.Error("Expected success false")
	}
}

// TestExecutionContextEmptyPath tests execution with empty path
func TestExecutionContextEmptyPath(t *testing.T) {
	runner := NewAgentRunner(nil)

	agent := &Agent{
		Name: "test",
		Path: "",
	}

	ctx := ExecutionContext{
		Agent: agent,
	}

	result, err := runner.Execute(ctx)
	if err == nil {
		t.Error("Expected error for empty path, got nil")
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.Success {
		t.Error("Expected success false")
	}
}

// TestExecutionWithTimeout tests timeout context
func TestExecutionWithTimeout(t *testing.T) {
	runner := NewAgentRunner(nil)

	agent := &Agent{
		Name: "test",
		Path: "/bin/sleep",
	}

	ctx := ExecutionContext{
		Agent:   agent,
		Timeout: 100 * time.Millisecond,
		Params: map[string]interface{}{
			"duration": "10",
		},
		Context: context.Background(),
	}

	result, _ := runner.Execute(ctx)
	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Timeout should cause failure (unless /bin/sleep doesn't exist)
	if result.EndTime.Before(result.StartTime) {
		t.Error("End time should be after start time")
	}
}

// TestExecutionResultTiming tests execution timing
func TestExecutionResultTiming(t *testing.T) {
	runner := NewAgentRunner(nil)

	agent := &Agent{
		Name: "test",
		Path: "/bin/true",
	}

	ctx := ExecutionContext{
		Agent:         agent,
		CaptureOutput: true,
	}

	result, _ := runner.Execute(ctx)
	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.StartTime.IsZero() {
		t.Error("Expected non-zero start time")
	}

	if result.EndTime.IsZero() {
		t.Error("Expected non-zero end time")
	}

	if result.EndTime.Before(result.StartTime) {
		t.Error("End time should be after start time")
	}

	if result.Duration == 0 {
		t.Error("Expected non-zero duration")
	}
}

// TestFormatOutputRaw tests raw output formatting
func TestFormatOutputRaw(t *testing.T) {
	result := &ExecutionResult{
		Output: "test output\n",
	}

	formatted := FormatOutput(result, true)
	if formatted != "test output\n" {
		t.Errorf("Expected raw output, got formatted: %q", formatted)
	}
}

// TestFormatOutputFormatted tests formatted output
func TestFormatOutputFormatted(t *testing.T) {
	result := &ExecutionResult{
		Output:   "test output\n",
		ExitCode: 0,
		Success:  true,
		Duration: 100 * time.Millisecond,
	}

	formatted := FormatOutput(result, false)
	if !contains(formatted, "Agent Execution Result") {
		t.Error("Expected formatted output to contain result header")
	}

	if !contains(formatted, "test output") {
		t.Error("Expected formatted output to contain agent output")
	}
}

// TestFormatOutputWithStderr tests formatted output with stderr
func TestFormatOutputWithStderr(t *testing.T) {
	result := &ExecutionResult{
		Output: "output",
		Stderr: "error",
	}

	formatted := FormatOutput(result, false)
	if !contains(formatted, "Errors") {
		t.Error("Expected formatted output to contain Errors section")
	}
}

// TestExpandPathHome tests home directory expansion
func TestExpandPathHome(t *testing.T) {
	path := "~/test/agent"
	expanded := ExpandPath(path)

	if expanded == path {
		t.Errorf("Expected expanded path, got %q", expanded)
	}

	if !filepath.IsAbs(expanded) {
		t.Errorf("Expected absolute path, got %q", expanded)
	}
}

// TestExpandPathAbsolute tests absolute path
func TestExpandPathAbsolute(t *testing.T) {
	path := "/usr/bin/test"
	expanded := ExpandPath(path)

	if expanded != path {
		t.Errorf("Expected same path, got %q", expanded)
	}
}

// TestNewAgentRunner tests AgentRunner creation
func TestNewAgentRunner(t *testing.T) {
	discoverer := NewDiscoverer(".")
	runner := NewAgentRunner(discoverer)

	if runner.discoverer != discoverer {
		t.Error("Expected discoverer to be set")
	}

	if runner.baseWorkDir == "" {
		t.Error("Expected baseWorkDir to be set")
	}
}

// TestNewAgentRunnerWithWorkDir tests AgentRunner creation with custom work dir
func TestNewAgentRunnerWithWorkDir(t *testing.T) {
	tmpDir := t.TempDir()
	discoverer := NewDiscoverer(".")
	runner := NewAgentRunnerWithWorkDir(discoverer, tmpDir)

	if runner.baseWorkDir != tmpDir {
		t.Errorf("Expected baseWorkDir %q, got %q", tmpDir, runner.baseWorkDir)
	}
}

// TestGetExecutionRequirements tests getting execution requirements
func TestGetExecutionRequirements(t *testing.T) {
	agent := &Agent{
		Name: "test",
	}

	req := agent.GetExecutionRequirements()
	if req == nil {
		t.Fatal("Expected non-nil requirements")
	}

	if req.TimeoutSeconds != 300 {
		t.Errorf("Expected timeout 300s, got %d", req.TimeoutSeconds)
	}
}

// TestGetExecutionRequirementsNilAgent tests with nil agent
func TestGetExecutionRequirementsNilAgent(t *testing.T) {
	var agent *Agent
	req := agent.GetExecutionRequirements()
	if req != nil {
		t.Error("Expected nil requirements for nil agent")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
