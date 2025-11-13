package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WorkspaceMarker represents a file or directory that indicates a project root
type WorkspaceMarker struct {
	Name        string
	IsDirectory bool
	Priority    int // Higher priority markers are preferred
	Type        string
}

// Common workspace markers, ordered by priority
var commonMarkers = []WorkspaceMarker{
	// VCS markers (highest priority)
	{Name: ".git", IsDirectory: true, Priority: 100, Type: "vcs"},
	{Name: ".hg", IsDirectory: true, Priority: 100, Type: "vcs"},

	// Go
	{Name: "go.mod", IsDirectory: false, Priority: 90, Type: "language"},

	// Node.js / JavaScript / TypeScript
	{Name: "package.json", IsDirectory: false, Priority: 85, Type: "language"},

	// Rust
	{Name: "Cargo.toml", IsDirectory: false, Priority: 85, Type: "language"},

	// Python
	{Name: "setup.py", IsDirectory: false, Priority: 80, Type: "language"},
	{Name: "pyproject.toml", IsDirectory: false, Priority: 80, Type: "language"},
	{Name: "Pipfile", IsDirectory: false, Priority: 80, Type: "language"},

	// Java / Kotlin
	{Name: "pom.xml", IsDirectory: false, Priority: 80, Type: "language"},
	{Name: "build.gradle", IsDirectory: false, Priority: 80, Type: "language"},
	{Name: "build.gradle.kts", IsDirectory: false, Priority: 80, Type: "language"},

	// .NET / C#
	{Name: "*.csproj", IsDirectory: false, Priority: 80, Type: "language"},
	{Name: "*.sln", IsDirectory: false, Priority: 80, Type: "language"},

	// Ruby
	{Name: "Gemfile", IsDirectory: false, Priority: 75, Type: "language"},

	// PHP
	{Name: "composer.json", IsDirectory: false, Priority: 75, Type: "language"},

	// Generic build tools
	{Name: "Makefile", IsDirectory: false, Priority: 70, Type: "build"},
	{Name: "CMakeLists.txt", IsDirectory: false, Priority: 70, Type: "build"},
}

// DetectionOptions configures workspace detection behavior
type DetectionOptions struct {
	// MaxDepth limits how deep to search for workspaces
	MaxDepth int

	// MaxWorkspaces limits the number of workspaces to find
	MaxWorkspaces int

	// IncludeHidden includes hidden directories (starting with .)
	IncludeHidden bool

	// PreferVCSRoots prioritizes directories with VCS markers
	PreferVCSRoots bool

	// CustomMarkers adds additional workspace markers
	CustomMarkers []WorkspaceMarker

	// ExcludePaths excludes specific paths from detection
	ExcludePaths []string
}

// DefaultDetectionOptions returns sensible defaults
func DefaultDetectionOptions() DetectionOptions {
	return DetectionOptions{
		MaxDepth:       3,
		MaxWorkspaces:  10,
		IncludeHidden:  false,
		PreferVCSRoots: true,
		CustomMarkers:  []WorkspaceMarker{},
		ExcludePaths: []string{
			"node_modules",
			"vendor",
			"target",
			"build",
			"dist",
			".git",
			".hg",
			".svn",
		},
	}
}

// DetectWorkspaces discovers workspace roots in a directory tree
func DetectWorkspaces(rootPath string, options DetectionOptions) ([]WorkspaceRoot, error) {
	// Normalize path
	rootPath = filepath.Clean(rootPath)

	// Check if root exists
	info, err := os.Stat(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat root path: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("root path is not a directory: %s", rootPath)
	}

	// Combine default markers with custom markers
	markers := append([]WorkspaceMarker{}, commonMarkers...)
	markers = append(markers, options.CustomMarkers...)

	// Find all potential workspace roots
	candidates := make(map[string]workspaceCandidate)
	err = findWorkspaceCandidates(rootPath, rootPath, 0, markers, options, candidates)
	if err != nil {
		return nil, fmt.Errorf("failed to find workspace candidates: %w", err)
	}

	// Convert candidates to workspace roots
	var roots []WorkspaceRoot
	for path := range candidates {
		// Detect VCS
		vcs, _ := detectVCS(path)

		root := WorkspaceRoot{
			Path: path,
			Name: filepath.Base(path),
			VCS:  vcs,
		}

		// Add Git metadata if available
		if vcs == VCSTypeGit {
			if hash, err := getGitCommitHash(path); err == nil {
				root.CommitHash = &hash
			}

			if urls, err := getGitRemoteURLs(path); err == nil && len(urls) > 0 {
				root.RemoteURLs = urls
			}
		}

		roots = append(roots, root)

		// Stop if we've reached the limit
		if len(roots) >= options.MaxWorkspaces {
			break
		}
	}

	// Sort by priority if preferring VCS roots
	if options.PreferVCSRoots {
		sortWorkspacesByVCS(roots)
	}

	return roots, nil
}

// workspaceCandidate represents a potential workspace root
type workspaceCandidate struct {
	path     string
	priority int
	markers  []string
}

// findWorkspaceCandidates recursively searches for workspace markers
func findWorkspaceCandidates(
	basePath string,
	currentPath string,
	depth int,
	markers []WorkspaceMarker,
	options DetectionOptions,
	candidates map[string]workspaceCandidate,
) error {
	// Check depth limit
	if depth > options.MaxDepth {
		return nil
	}

	// Check workspace limit
	if len(candidates) >= options.MaxWorkspaces {
		return nil
	}

	// Check if path should be excluded
	relPath, _ := filepath.Rel(basePath, currentPath)
	for _, exclude := range options.ExcludePaths {
		if strings.Contains(relPath, exclude) {
			return nil
		}
	}

	// Check for workspace markers in current directory
	entries, err := os.ReadDir(currentPath)
	if err != nil {
		// Skip directories we can't read
		return nil
	}

	var foundMarkers []string
	maxPriority := 0

	for _, entry := range entries {
		name := entry.Name()

		// Skip hidden files/directories unless included
		if !options.IncludeHidden && strings.HasPrefix(name, ".") && name != ".git" && name != ".hg" {
			continue
		}

		// Check against markers
		for _, marker := range markers {
			matched := false

			if marker.IsDirectory && entry.IsDir() {
				matched = name == marker.Name
			} else if !marker.IsDirectory && !entry.IsDir() {
				// Support glob patterns for file markers
				if strings.Contains(marker.Name, "*") {
					matched, _ = filepath.Match(marker.Name, name)
				} else {
					matched = name == marker.Name
				}
			}

			if matched {
				foundMarkers = append(foundMarkers, marker.Name)
				if marker.Priority > maxPriority {
					maxPriority = marker.Priority
				}
			}
		}
	}

	// If we found markers, this is a candidate workspace
	if len(foundMarkers) > 0 {
		candidates[currentPath] = workspaceCandidate{
			path:     currentPath,
			priority: maxPriority,
			markers:  foundMarkers,
		}
	}

	// Recurse into subdirectories
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Skip hidden directories unless included
		if !options.IncludeHidden && strings.HasPrefix(name, ".") {
			continue
		}

		// Skip excluded paths
		skip := false
		for _, exclude := range options.ExcludePaths {
			if name == exclude {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		subPath := filepath.Join(currentPath, name)
		err := findWorkspaceCandidates(basePath, subPath, depth+1, markers, options, candidates)
		if err != nil {
			// Continue even if one subdirectory fails
			continue
		}
	}

	return nil
}

// sortWorkspacesByVCS sorts workspaces with VCS roots first
func sortWorkspacesByVCS(roots []WorkspaceRoot) {
	// Simple bubble sort - VCS roots first, then alphabetical
	for i := 0; i < len(roots)-1; i++ {
		for j := i + 1; j < len(roots); j++ {
			// VCS roots come first
			iHasVCS := roots[i].VCS != VCSTypeNone
			jHasVCS := roots[j].VCS != VCSTypeNone

			if !iHasVCS && jHasVCS {
				roots[i], roots[j] = roots[j], roots[i]
			} else if iHasVCS == jHasVCS {
				// Both have or don't have VCS, sort alphabetically
				if roots[i].Name > roots[j].Name {
					roots[i], roots[j] = roots[j], roots[i]
				}
			}
		}
	}
}

// DetectWorkspacesFromPreferences uses preferences to detect workspaces
func DetectWorkspacesFromPreferences(rootPath string, prefs Preferences) ([]WorkspaceRoot, error) {
	if !prefs.AutoDetectWorkspaces {
		return nil, nil
	}

	options := DefaultDetectionOptions()
	options.MaxWorkspaces = prefs.MaxWorkspaces
	options.PreferVCSRoots = prefs.PreferVCSRoots
	options.IncludeHidden = prefs.IncludeHidden

	return DetectWorkspaces(rootPath, options)
}

// SmartWorkspaceInitialization initializes workspaces intelligently
// 1. Try loading from config file
// 2. If no config, detect workspaces automatically
// 3. If detection finds nothing or fails, use single directory
func SmartWorkspaceInitialization(rootPath string) (*Manager, error) {
	// Try loading from config
	manager, _, err := LoadManagerFromDirectory(rootPath)
	if err == nil && manager != nil {
		// Config loaded successfully
		return manager, nil
	}

	// Try detecting workspaces
	options := DefaultDetectionOptions()
	roots, err := DetectWorkspaces(rootPath, options)
	if err == nil && len(roots) > 0 {
		// Found workspaces
		primaryIndex := 0

		// If rootPath itself is one of the roots, make it primary
		for i, root := range roots {
			if root.Path == rootPath {
				primaryIndex = i
				break
			}
		}

		manager = NewManager(roots, primaryIndex)

		// Save config for next time
		SaveManagerToDirectory(rootPath, manager, nil)

		return manager, nil
	}

	// Fall back to single directory
	return FromSingleDirectory(rootPath)
}
