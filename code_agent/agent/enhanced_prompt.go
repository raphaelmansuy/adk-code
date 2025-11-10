// Enhanced system prompt for ADK Code Agent - Better than Cline
// This file combines modular prompt components for maintainability and easier customization.
package agent

// EnhancedSystemPrompt combines all prompt components into a single system prompt.
// Components are defined in separate files for modularity:
// - prompt_tools.go: Tool descriptions and APIs
// - prompt_guidance.go: Decision trees and best practices
// - prompt_pitfalls.go: Common mistakes and solutions
// - prompt_workflow.go: Workflow patterns and response styles
const EnhancedSystemPrompt = ToolsSection + "\n" + GuidanceSection + "\n" + PitfallsSection + "\n" + WorkflowSection
