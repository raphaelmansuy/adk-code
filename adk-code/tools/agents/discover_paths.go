// Package agents provides tools for agent definition discovery and management
package agents

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"

	"adk-code/pkg/agents"
	common "adk-code/tools/base"
)

// DiscoverPathsInput defines input parameters for path discovery
type DiscoverPathsInput struct {
	Verbose bool `json:"verbose,omitempty" jsonschema:"Show detailed path information including accessibility"`
}

// PathInfo represents a single agent search path
type PathInfo struct {
	Path        string `json:"path"`
	Source      string `json:"source"`
	Order       int    `json:"order"`
	Exists      bool   `json:"exists"`
	Accessible  bool   `json:"accessible"`
	AgentCount  int    `json:"agent_count,omitempty"`
	Description string `json:"description,omitempty"`
}

// DiscoverPathsOutput defines the output of path discovery
type DiscoverPathsOutput struct {
	Paths        []PathInfo `json:"paths"`
	SearchOrder  []string   `json:"search_order"`
	SkipMissing  bool       `json:"skip_missing"`
	TotalAgents  int        `json:"total_agents,omitempty"`
	Success      bool       `json:"success"`
	Error        string     `json:"error,omitempty"`
	Summary      string     `json:"summary"`
	ConfigFile   string     `json:"config_file,omitempty"`
	ConfigLoaded bool       `json:"config_loaded"`
}

// NewDiscoverPathsTool creates a tool for discovering and displaying agent paths
// Shows all configured agent search paths and their status
func NewDiscoverPathsTool() (tool.Tool, error) {
	handler := func(ctx tool.Context, input DiscoverPathsInput) DiscoverPathsOutput {
		projectRoot := "."

		// Load configuration
		cfg, err := agents.LoadConfig(projectRoot)
		if err != nil {
			cfg = agents.NewConfig()
			cfg.ProjectPath = ".adk/agents"
		}

		output := DiscoverPathsOutput{
			Paths:       make([]PathInfo, 0),
			SearchOrder: cfg.SearchOrder,
			SkipMissing: cfg.SkipMissing,
			ConfigFile:  filepath.Join(projectRoot, ".adk", "config.yaml"),
			Success:     true,
		}

		// Get all paths from config
		allPaths := cfg.GetAllPaths()

		// If verbose, also try to discover agents and count them
		var totalAgents int
		if input.Verbose {
			discoverer := agents.NewDiscovererWithConfig(projectRoot, cfg)
			result, _ := discoverer.DiscoverAll()
			totalAgents = result.Total
			output.TotalAgents = totalAgents
		}

		// Map source names for lookup
		sourceMap := map[string]string{
			cfg.ProjectPath: "project",
			cfg.UserPath:    "user",
		}
		for _, pluginPath := range cfg.PluginPaths {
			sourceMap[pluginPath] = "plugin"
		}

		// Build path info for each configured path
		for i, path := range allPaths {
			pathInfo := PathInfo{
				Path:   path,
				Order:  i + 1,
				Source: sourceMap[path],
			}

			// Check if path exists
			expanded := expandPathForCheck(path)
			pathInfo.Exists = pathExists(expanded)
			pathInfo.Accessible = isAccessible(expanded)

			// Add description based on source
			switch pathInfo.Source {
			case "project":
				pathInfo.Description = "Project-level agents directory"
			case "user":
				pathInfo.Description = "User-level agents directory (home)"
			case "plugin":
				pathInfo.Description = "Plugin-level agents directory"
			}

			output.Paths = append(output.Paths, pathInfo)
		}

		// Generate summary
		activeCount := 0
		for _, p := range output.Paths {
			if p.Exists && p.Accessible {
				activeCount++
			}
		}

		output.Summary = fmt.Sprintf("%d path(s) configured, %d active", len(output.Paths), activeCount)
		if input.Verbose {
			output.Summary = fmt.Sprintf("%s, %d total agent(s)", output.Summary, totalAgents)
		}

		return output
	}

	t, err := functiontool.New(functiontool.Config{
		Name:        "discover_paths",
		Description: "Shows all configured agent search paths and their status. Lists project, user, and plugin agent directories.",
	}, handler)

	if err == nil {
		common.Register(common.ToolMetadata{
			Tool:      t,
			Category:  common.CategorySearchDiscovery,
			Priority:  7,
			UsageHint: "View agent search path configuration",
		})
	}

	return t, err
}

// expandPathForCheck expands paths for file system checking
func expandPathForCheck(path string) string {
	if strings.HasPrefix(path, "~") {
		home := os.Getenv("HOME")
		if home != "" {
			return filepath.Join(home, path[1:])
		}
	}
	return path
}

// pathExists checks if a path exists
func pathExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

// isAccessible checks if a path is readable
func isAccessible(path string) bool {
	// Try to list the directory
	entries, err := os.ReadDir(path)
	return err == nil && entries != nil
}
