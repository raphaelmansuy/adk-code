// Package main demonstrates the context management system
package main

import (
	"fmt"
	"strings"
	"time"

	"adk-code/internal/context"
	"adk-code/internal/instructions"
	"adk-code/pkg/models"
)

func main() {
	fmt.Println("=== Context Management Example ===")
	fmt.Println()

	// Example 1: Basic Context Manager
	demonstrateContextManager()

	// Example 2: Token Tracking
	demonstrateTokenTracking()

	// Example 3: Output Truncation
	demonstrateOutputTruncation()

	// Example 4: Instruction Loading
	demonstrateInstructions()
}

func demonstrateContextManager() {
	fmt.Println("1. Context Manager")
	fmt.Println("------------------")

	// Create a context manager with a small context window for demonstration
	modelConfig := models.Config{
		Name:          "test-model",
		ContextWindow: 10000, // Small for demo
	}

	cm := context.NewContextManager(modelConfig)

	// Add some messages
	items := []context.ResponseItem{
		{
			ID:        "msg-1",
			Type:      context.ItemMessage,
			Role:      "user",
			Content:   "What is the capital of France?",
			Timestamp: time.Now(),
		},
		{
			ID:        "msg-2",
			Type:      context.ItemMessage,
			Role:      "assistant",
			Content:   "The capital of France is Paris.",
			Timestamp: time.Now(),
		},
		{
			ID:        "msg-3",
			Type:      context.ItemMessage,
			Role:      "user",
			Content:   "Tell me more about Paris.",
			Timestamp: time.Now(),
		},
	}

	for _, item := range items {
		err := cm.AddItem(item)
		if err == context.ErrCompactionNeeded {
			fmt.Println("⚠️  Compaction needed!")
		}
	}

	// Get token info
	info := cm.TokenInfo()
	fmt.Printf("📊 Token Usage:\n")
	fmt.Printf("   Used: %d tokens\n", info.UsedTokens)
	fmt.Printf("   Available: %d tokens\n", info.AvailableTokens)
	fmt.Printf("   Percentage: %.1f%%\n", info.PercentageUsed*100)
	fmt.Printf("   Compaction threshold: %.0f%%\n\n", info.CompactThreshold*100)
}

func demonstrateTokenTracking() {
	fmt.Println("2. Token Tracking")
	fmt.Println("-----------------")

	tracker := context.NewTokenTracker("demo-session", "gemini-2.5-flash", 1_000_000)

	// Simulate some turns
	turns := []struct {
		input  int
		output int
	}{
		{1500, 800},
		{1200, 600},
		{1800, 900},
		{1600, 750},
	}

	for i, turn := range turns {
		tracker.RecordTurn(turn.input, turn.output)
		fmt.Printf("Turn %d: %d input + %d output = %d total\n",
			i+1, turn.input, turn.output, turn.input+turn.output)
	}

	fmt.Printf("\n📈 Statistics:\n")
	fmt.Printf("   Total turns: %d\n", tracker.GetTurnCount())
	fmt.Printf("   Total tokens: %d\n", tracker.GetTotalTokens())
	fmt.Printf("   Average per turn: %d\n", tracker.AverageTurnSize())
	fmt.Printf("   Estimated remaining: %d turns\n\n",
		tracker.EstimateRemainingTurns(1_000_000, 100_000))
}

func demonstrateOutputTruncation() {
	fmt.Println("3. Output Truncation")
	fmt.Println("--------------------")

	// Create a large output
	lines := make([]string, 300)
	for i := 0; i < 300; i++ {
		lines[i] = fmt.Sprintf("Line %d: Some output content here", i+1)
	}
	largeOutput := strings.Join(lines, "\n")

	fmt.Printf("Original output: %d lines, %d bytes\n", 300, len(largeOutput))

	// Create context manager with default truncation settings
	modelConfig := models.Config{
		Name:          "test-model",
		ContextWindow: 1_000_000,
	}
	cm := context.NewContextManager(modelConfig)

	// Add tool output (will be truncated)
	item := context.ResponseItem{
		ID:        "output-1",
		Type:      context.ItemToolOutput,
		Role:      "tool",
		Content:   largeOutput,
		Timestamp: time.Now(),
	}

	cm.AddItem(item)

	// Get the truncated version
	history, _ := cm.GetHistory()
	truncated := history[0].Content

	fmt.Printf("Truncated output: %d bytes\n", len(truncated))

	// Show that it contains beginning and end
	hasBeginning := strings.Contains(truncated, "Line 1:")
	hasEnd := strings.Contains(truncated, "Line 299:") || strings.Contains(truncated, "Line 300:")
	hasMarker := strings.Contains(truncated, "omitted")

	fmt.Printf("   ✓ Contains beginning: %v\n", hasBeginning)
	fmt.Printf("   ✓ Contains end: %v\n", hasEnd)
	fmt.Printf("   ✓ Has elision marker: %v\n\n", hasMarker)
}

func demonstrateInstructions() {
	fmt.Println("4. Instruction Hierarchy")
	fmt.Println("------------------------")

	// Note: This will work with actual AGENTS.md files in the project
	workdir := "."
	loader := instructions.NewInstructionLoader(workdir)
	result := loader.Load()

	fmt.Printf("Instruction sources:\n")

	if result.Global != "" {
		fmt.Printf("   ✓ Global instructions loaded (%d bytes)\n", len(result.Global))
	} else {
		fmt.Printf("   ⚬ No global instructions\n")
	}

	if result.ProjectRoot != "" {
		fmt.Printf("   ✓ Project instructions loaded (%d bytes)\n", len(result.ProjectRoot))
	} else {
		fmt.Printf("   ⚬ No project instructions\n")
	}

	if len(result.Nested) > 0 {
		fmt.Printf("   ✓ Nested instructions: %d directories\n", len(result.Nested))
	} else {
		fmt.Printf("   ⚬ No nested instructions\n")
	}

	fmt.Printf("\nMerged instructions: %d bytes\n", len(result.Merged))
	if result.Truncated {
		fmt.Printf("   ⚠️  Instructions were truncated to fit limit\n")
	}

	fmt.Println("\n💡 Tip: Create AGENTS.md files to provide instructions to the agent")
	fmt.Println("   - ~/.adk-code/AGENTS.md (global)")
	fmt.Println("   - <project-root>/AGENTS.md (project-level)")
	fmt.Println("   - <any-directory>/AGENTS.md (directory-specific)")
}
