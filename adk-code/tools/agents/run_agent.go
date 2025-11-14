// Package agents provides tools for agent definition discovery and management
package agents

import (
	"fmt"
	"time"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"

	"adk-code/pkg/agents"
	common "adk-code/tools/base"
)

// RunAgentInput defines input parameters for running an agent
type RunAgentInput struct {
	AgentName     string                 `json:"agent_name" jsonschema:"Name of the agent to run (required)"`
	Params        map[string]interface{} `json:"params,omitempty" jsonschema:"Parameters to pass to the agent"`
	Timeout       int                    `json:"timeout,omitempty" jsonschema:"Timeout in seconds (0 = no timeout)"`
	CaptureOutput bool                   `json:"capture_output,omitempty" jsonschema:"Capture and return output"`
	Detailed      bool                   `json:"detailed,omitempty" jsonschema:"Include detailed timing and metadata"`
}

// RunAgentOutput defines the output of running an agent
type RunAgentOutput struct {
	Agent     string `json:"agent"`
	Output    string `json:"output"`
	Error     string `json:"error,omitempty"`
	ExitCode  int    `json:"exit_code"`
	Duration  int64  `json:"duration_ms"`
	Success   bool   `json:"success"`
	StartTime string `json:"start_time,omitempty"`
	EndTime   string `json:"end_time,omitempty"`
	Message   string `json:"message"`
}

// NewRunAgentTool creates a tool for running discovered agents
// Uses current working directory as project root for discovery
// Automatically loads configuration from .adk/config.yaml if present
func NewRunAgentTool() (tool.Tool, error) {
	handler := func(ctx tool.Context, input RunAgentInput) RunAgentOutput {
		output := RunAgentOutput{
			Agent: input.AgentName,
		}

		// Validate input
		if input.AgentName == "" {
			output.Success = false
			output.Error = "agent_name is required"
			output.Message = "Failed to run agent: agent_name is required"
			return output
		}

		// Use current working directory as project root
		projectRoot := "."

		// Load configuration
		cfg, err := agents.LoadConfig(projectRoot)
		if err != nil {
			// Fall back to default config if loading fails
			cfg = agents.NewConfig()
			cfg.ProjectPath = ".adk/agents"
		}

		// Create discoverer with configuration
		discoverer := agents.NewDiscovererWithConfig(projectRoot, cfg)
		runner := agents.NewAgentRunner(discoverer)

		// Get agent by name
		agent, err := runner.GetAgentByName(input.AgentName)
		if err != nil {
			output.Success = false
			output.Error = fmt.Sprintf("Agent not found: %v", err)
			output.Message = fmt.Sprintf("Failed to find agent %q: %v", input.AgentName, err)
			return output
		}

		// Validate agent before execution
		if err := agent.Validate(); err != nil {
			output.Success = false
			output.Error = fmt.Sprintf("Agent validation failed: %v", err)
			output.Message = fmt.Sprintf("Agent %q validation failed: %v", input.AgentName, err)
			return output
		}

		// Validate agent requirements
		req := agent.GetExecutionRequirements()
		if err := runner.ValidateRequirements(req); err != nil {
			output.Success = false
			output.Error = fmt.Sprintf("Agent requirements not met: %v", err)
			output.Message = fmt.Sprintf("Agent %q requirements not met: %v", input.AgentName, err)
			return output
		}

		// Build execution context
		execCtx := agents.ExecutionContext{
			Agent:           agent,
			Params:          input.Params,
			CaptureOutput:   input.CaptureOutput,
			ReturnRawOutput: !input.Detailed,
		}

		// Set timeout if specified
		if input.Timeout > 0 {
			execCtx.Timeout = time.Duration(input.Timeout) * time.Second
		}

		// Execute agent
		result, err := runner.Execute(execCtx)
		if err != nil {
			output.Success = false
			output.Error = fmt.Sprintf("Execution error: %v", err)
			output.Message = fmt.Sprintf("Agent %q execution failed: %v", input.AgentName, err)
			return output
		}

		// Populate output from execution result
		output.Output = result.Output
		output.Error = result.Error
		output.ExitCode = result.ExitCode
		output.Duration = result.Duration.Milliseconds()
		output.Success = result.Success

		// Add timestamps if detailed
		if input.Detailed {
			output.StartTime = result.StartTime.Format(time.RFC3339)
			output.EndTime = result.EndTime.Format(time.RFC3339)
		}

		// Set message based on success
		if result.Success {
			output.Message = fmt.Sprintf("Agent %q executed successfully in %v", input.AgentName, result.Duration)
		} else {
			output.Message = fmt.Sprintf("Agent %q failed with exit code %d", input.AgentName, result.ExitCode)
		}

		return output
	}

	t, err := functiontool.New(functiontool.Config{
		Name:        "run_agent",
		Description: "Execute a discovered agent with parameters and capture output. Agents are found using the agent discovery system (.adk/agents/, ~/.adk/agents/, plugin paths).",
	}, handler)

	if err != nil {
		return nil, fmt.Errorf("failed to create run_agent tool: %w", err)
	}

	// Register the tool
	common.Register(common.ToolMetadata{
		Tool:      t,
		Category:  common.CategoryExecution,
		Priority:  8,
		UsageHint: "Execute agents that have been discovered in the project",
	})

	return t, nil
}

// init automatically triggers tool registration at package initialization.
// This ensures run_agent tool is registered when the agents tool package is imported.
func init() {
	_ = NewRunAgentTool
}
