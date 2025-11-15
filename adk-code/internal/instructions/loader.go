// Package instructions manages hierarchical user instructions
package instructions

import (
	"os"
	"path/filepath"
	"sort"
)

// InstructionLoader manages hierarchical user instructions
// Similar to Codex's AGENTS.md system
type InstructionLoader struct {
	globalPath  string // ~/.adk-code/AGENTS.md
	projectRoot string // Repository root
	workingDir  string // Current working directory
}

// LoadedInstructions represents merged instructions at runtime
type LoadedInstructions struct {
	Global      string            // Global instructions
	ProjectRoot string            // Root-level project instructions
	Nested      map[string]string // Nested directory instructions
	Merged      string            // All instructions combined
	MaxBytes    int               // Total size limit
	Truncated   bool              // True if merged was truncated
}

// NewInstructionLoader creates a new instruction loader
func NewInstructionLoader(workdir string) *InstructionLoader {
	home, _ := os.UserHomeDir()
	globalPath := filepath.Join(home, ".adk-code", "AGENTS.md")

	return &InstructionLoader{
		globalPath:  globalPath,
		projectRoot: findProjectRoot(workdir),
		workingDir:  workdir,
	}
}

// Load gathers instructions from all levels
func (il *InstructionLoader) Load() LoadedInstructions {
	result := LoadedInstructions{
		Nested:   make(map[string]string),
		MaxBytes: 32 * 1024, // 32 KiB default limit
	}

	// 1. Load global instructions (if present)
	result.Global = il.loadFileIfExists(il.globalPath)

	// 2. Load project root instructions
	if il.projectRoot != "" {
		rootAgents := filepath.Join(il.projectRoot, "AGENTS.md")
		result.ProjectRoot = il.loadFileIfExists(rootAgents)
	}

	// 3. Load nested directory instructions
	il.loadNestedInstructions(&result)

	// 4. Merge with size limit
	merged := il.mergeInstructions(result)

	// Check if truncation occurred
	if len(merged) > result.MaxBytes {
		result.Truncated = true
	}

	result.Merged = merged

	return result
}

func (il *InstructionLoader) loadFileIfExists(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(content)
}

func (il *InstructionLoader) loadNestedInstructions(result *LoadedInstructions) {
	if il.projectRoot == "" {
		return
	}

	// Walk from project root to working directory
	current := il.workingDir
	paths := []string{}

	// Collect all paths from working dir up to project root
	for current != il.projectRoot && current != "" && current != "/" {
		paths = append(paths, current)
		current = filepath.Dir(current)
	}

	// Reverse to go from root to leaf
	for i := len(paths) - 1; i >= 0; i-- {
		agentsFile := filepath.Join(paths[i], "AGENTS.md")
		content := il.loadFileIfExists(agentsFile)
		if content != "" {
			result.Nested[paths[i]] = content
		}
	}

	// Also check working directory itself
	agentsFile := filepath.Join(il.workingDir, "AGENTS.md")
	content := il.loadFileIfExists(agentsFile)
	if content != "" {
		result.Nested[il.workingDir] = content
	}
}

func (il *InstructionLoader) mergeInstructions(lr LoadedInstructions) string {
	var merged string

	// Order: global → project root → nested (root to leaf)
	if lr.Global != "" {
		merged += lr.Global + "\n\n"
	}

	if lr.ProjectRoot != "" {
		merged += lr.ProjectRoot + "\n\n"
	}

	// Add nested in sorted order (by path depth)
	paths := make([]string, 0, len(lr.Nested))
	for path := range lr.Nested {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	for _, path := range paths {
		content := lr.Nested[path]
		if content != "" {
			merged += content + "\n\n"
		}
	}

	// Truncate if needed
	if len(merged) > lr.MaxBytes {
		merged = merged[:lr.MaxBytes]
		merged += "\n\n[instructions truncated to fit limit]"
	}

	return merged
}

// findProjectRoot finds the root directory of the project
func findProjectRoot(workdir string) string {
	// Walk up looking for .git, .hg, go.mod, etc.
	current := workdir
	for current != "/" && current != "" {
		markers := []string{".git", ".hg", "go.mod", "package.json", "Cargo.toml"}
		for _, marker := range markers {
			if _, err := os.Stat(filepath.Join(current, marker)); err == nil {
				return current
			}
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return ""
}
