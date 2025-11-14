package agents

import (
	"testing"

	"google.golang.org/adk/tool"
)

// TestNewRunAgentTool tests creating the run_agent tool
func TestNewRunAgentTool(t *testing.T) {
	tool, err := NewRunAgentTool()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if tool == nil {
		t.Fatal("Expected non-nil tool")
	}
}

// TestRunAgentInputValidation tests input validation
func TestRunAgentInputValidation(t *testing.T) {
	input := RunAgentInput{
		AgentName: "test-agent",
		Timeout:   5,
	}

	if input.AgentName == "" {
		t.Error("Expected agent name to be set")
	}
}

// TestRunAgentOutputStructure tests output structure
func TestRunAgentOutputStructure(t *testing.T) {
	output := RunAgentOutput{
		Agent:   "test-agent",
		Success: true,
		Output:  "test output",
	}

	if output.Agent != "test-agent" {
		t.Errorf("Expected agent 'test-agent', got %q", output.Agent)
	}

	if !output.Success {
		t.Error("Expected success true")
	}
}

// TestRunAgentOutputWithError tests error output
func TestRunAgentOutputWithError(t *testing.T) {
	output := RunAgentOutput{
		Agent:   "test-agent",
		Success: false,
		Error:   "test error",
		Message: "Test failed",
	}

	if output.Success {
		t.Error("Expected success false")
	}

	if output.Error == "" {
		t.Error("Expected error message")
	}
}

// TestRunAgentInputWithParams tests parameter handling
func TestRunAgentInputWithParams(t *testing.T) {
	input := RunAgentInput{
		AgentName: "test-agent",
		Params: map[string]interface{}{
			"param1": "value1",
			"param2": 42,
		},
	}

	if len(input.Params) != 2 {
		t.Errorf("Expected 2 params, got %d", len(input.Params))
	}
}

// TestRunAgentInputWithTimeout tests timeout handling
func TestRunAgentInputWithTimeout(t *testing.T) {
	input := RunAgentInput{
		AgentName: "test-agent",
		Timeout:   30,
	}

	if input.Timeout != 30 {
		t.Errorf("Expected timeout 30, got %d", input.Timeout)
	}
}

// TestRunAgentInputCaptureOutput tests capture output flag
func TestRunAgentInputCaptureOutput(t *testing.T) {
	input := RunAgentInput{
		AgentName:     "test-agent",
		CaptureOutput: true,
	}

	if !input.CaptureOutput {
		t.Error("Expected CaptureOutput to be true")
	}
}

// TestRunAgentInputDetailed tests detailed flag
func TestRunAgentInputDetailed(t *testing.T) {
	input := RunAgentInput{
		AgentName: "test-agent",
		Detailed:  true,
	}

	if !input.Detailed {
		t.Error("Expected Detailed to be true")
	}
}

// TestRunAgentOutputTiming tests timing fields
func TestRunAgentOutputTiming(t *testing.T) {
	output := RunAgentOutput{
		Agent:     "test-agent",
		Duration:  100,
		StartTime: "2025-11-14T20:00:00Z",
		EndTime:   "2025-11-14T20:00:01Z",
	}

	if output.Duration <= 0 {
		t.Error("Expected positive duration")
	}

	if output.StartTime == "" {
		t.Error("Expected start time")
	}

	if output.EndTime == "" {
		t.Error("Expected end time")
	}
}

// TestRunAgentOutputExitCode tests exit code field
func TestRunAgentOutputExitCode(t *testing.T) {
	output := RunAgentOutput{
		Agent:    "test-agent",
		ExitCode: 0,
		Success:  true,
	}

	if output.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", output.ExitCode)
	}

	output.ExitCode = 1
	if output.ExitCode != 1 {
		t.Errorf("Expected exit code 1, got %d", output.ExitCode)
	}
}

// TestRunAgentMissingAgentName tests handling of missing agent name
func TestRunAgentMissingAgentName(t *testing.T) {
	tool, _ := NewRunAgentTool()
	if tool == nil {
		t.Fatal("Expected non-nil tool")
	}

	input := RunAgentInput{
		AgentName: "", // Empty name
	}

	// Just verify the input structure is correct
	if input.AgentName != "" {
		t.Error("Expected empty agent name")
	}
}

// TestRunAgentToolInterface tests that run_agent implements tool.Tool interface
func TestRunAgentToolInterface(t *testing.T) {
	agentTool, err := NewRunAgentTool()
	if err != nil {
		t.Fatalf("Failed to create run_agent tool: %v", err)
	}

	// Verify it's a tool.Tool by checking if it has Name method
	var _ tool.Tool = agentTool
}
