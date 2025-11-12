// Package common provides a dynamic tool registry for categorized tool management.
package common

import (
	"fmt"
	"sort"
	"sync"

	"google.golang.org/adk/tool"
)

// ToolCategory represents the functional category of a tool.
type ToolCategory string

const (
	// CategoryFileOperations includes basic file read/write/list operations
	CategoryFileOperations ToolCategory = "File Operations"
	// CategorySearchDiscovery includes tools for finding files and content
	CategorySearchDiscovery ToolCategory = "Search & Discovery"
	// CategoryCodeEditing includes advanced editing tools (patches, search/replace)
	CategoryCodeEditing ToolCategory = "Code Editing"
	// CategoryExecution includes command and program execution tools
	CategoryExecution ToolCategory = "Execution"
	// CategoryWorkspace includes workspace management tools
	CategoryWorkspace ToolCategory = "Workspace Management"
	// CategoryDisplay includes tools for displaying messages and task lists to the user
	CategoryDisplay ToolCategory = "Display & Communication"
)

// ToolMetadata contains a tool and its categorization metadata.
type ToolMetadata struct {
	Tool      tool.Tool
	Category  ToolCategory
	Priority  int    // Lower numbers appear first within category (0 = highest priority)
	UsageHint string // Brief usage guidance for the LLM
}

// ToolRegistry manages categorized tools for the coding agent.
type ToolRegistry struct {
	mu    sync.RWMutex
	tools map[ToolCategory][]ToolMetadata
}

// NewToolRegistry creates a new empty tool registry.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[ToolCategory][]ToolMetadata),
	}
}

// Register adds a tool with its metadata to the registry.
func (r *ToolRegistry) Register(metadata ToolMetadata) error {
	if metadata.Tool == nil {
		return fmt.Errorf("cannot register nil tool")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.tools[metadata.Category] = append(r.tools[metadata.Category], metadata)
	return nil
}

// GetByCategory returns all tools in a specific category, sorted by priority.
func (r *ToolRegistry) GetByCategory(category ToolCategory) []ToolMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := r.tools[category]
	// Sort by priority (lower number = higher priority)
	sorted := make([]ToolMetadata, len(tools))
	copy(sorted, tools)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Priority < sorted[j].Priority
	})
	return sorted
}

// GetAllTools returns all tools as a flat list.
func (r *ToolRegistry) GetAllTools() []tool.Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var allTools []tool.Tool
	for _, toolList := range r.tools {
		for _, metadata := range toolList {
			allTools = append(allTools, metadata.Tool)
		}
	}
	return allTools
}

// GetCategories returns all registered categories in a consistent order.
func (r *ToolRegistry) GetCategories() []ToolCategory {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return categories in a logical order for prompt generation
	order := []ToolCategory{
		CategoryFileOperations,
		CategoryCodeEditing,
		CategorySearchDiscovery,
		CategoryExecution,
		CategoryWorkspace,
		CategoryDisplay,
	}

	// Filter to only categories that have tools
	var categories []ToolCategory
	for _, cat := range order {
		if len(r.tools[cat]) > 0 {
			categories = append(categories, cat)
		}
	}
	return categories
}

// Count returns the total number of registered tools.
func (r *ToolRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, toolList := range r.tools {
		count += len(toolList)
	}
	return count
}

// Global registry instance
var globalRegistry = NewToolRegistry()

// Register adds a tool to the global registry (convenience function).
func Register(metadata ToolMetadata) error {
	return globalRegistry.Register(metadata)
}

// GetRegistry returns the global tool registry.
func GetRegistry() *ToolRegistry {
	return globalRegistry
}
