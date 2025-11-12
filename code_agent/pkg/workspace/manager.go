package workspace

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

// Manager manages one or more workspace roots with support for
// single-root (backward compatible) and multi-root workspaces
type Manager struct {
	roots        []WorkspaceRoot
	primaryIndex int
}

// NewManager creates a new workspace manager with the given roots
func NewManager(roots []WorkspaceRoot, primaryIndex int) *Manager {
	if len(roots) == 0 {
		return &Manager{
			roots:        []WorkspaceRoot{},
			primaryIndex: 0,
		}
	}

	// Validate and clamp primary index
	if primaryIndex < 0 || primaryIndex >= len(roots) {
		primaryIndex = 0
	}

	return &Manager{
		roots:        roots,
		primaryIndex: primaryIndex,
	}
}

// FromSingleDirectory creates a workspace manager from a single directory
// This is the backward-compatible mode for existing code
func FromSingleDirectory(cwd string) (*Manager, error) {
	// Detect VCS type
	vcs, err := detectVCS(cwd)
	if err != nil {
		// Not having VCS is okay, just log and continue
		vcs = VCSTypeNone
	}

	// Create workspace root
	root := WorkspaceRoot{
		Path: cwd,
		Name: filepath.Base(cwd),
		VCS:  vcs,
	}

	// Add Git metadata if this is a Git repository
	if vcs == VCSTypeGit {
		if hash, err := getGitCommitHash(cwd); err == nil {
			root.CommitHash = &hash
		}

		if urls, err := getGitRemoteURLs(cwd); err == nil && len(urls) > 0 {
			root.RemoteURLs = urls
		}
	}

	return NewManager([]WorkspaceRoot{root}, 0), nil
}

// GetRoots returns a copy of all workspace roots
func (m *Manager) GetRoots() []WorkspaceRoot {
	return append([]WorkspaceRoot{}, m.roots...)
}

// GetPrimaryRoot returns the primary workspace root
func (m *Manager) GetPrimaryRoot() *WorkspaceRoot {
	if len(m.roots) == 0 {
		return nil
	}
	root := m.roots[m.primaryIndex]
	return &root
}

// GetPrimaryIndex returns the index of the primary workspace root
func (m *Manager) GetPrimaryIndex() int {
	return m.primaryIndex
}

// SetPrimaryIndex sets the primary workspace root by index
func (m *Manager) SetPrimaryIndex(index int) error {
	if index < 0 || index >= len(m.roots) {
		return fmt.Errorf("invalid workspace index: %d (have %d roots)", index, len(m.roots))
	}
	m.primaryIndex = index
	return nil
}

// SetPrimaryByName sets the primary workspace root by name
func (m *Manager) SetPrimaryByName(name string) error {
	for i, root := range m.roots {
		if root.Name == name {
			m.primaryIndex = i
			return nil
		}
	}
	return fmt.Errorf("workspace not found: %s", name)
}

// SetPrimaryByPath sets the primary workspace root by path
func (m *Manager) SetPrimaryByPath(path string) error {
	for i, root := range m.roots {
		if root.Path == path {
			m.primaryIndex = i
			return nil
		}
	}
	return fmt.Errorf("workspace not found at path: %s", path)
}

// SwitchWorkspace switches to a different workspace and returns the new primary
func (m *Manager) SwitchWorkspace(identifier string) (*WorkspaceRoot, error) {
	// Try by name first
	if err := m.SetPrimaryByName(identifier); err == nil {
		return m.GetPrimaryRoot(), nil
	}

	// Try by path
	if err := m.SetPrimaryByPath(identifier); err == nil {
		return m.GetPrimaryRoot(), nil
	}

	return nil, fmt.Errorf("workspace not found: %s", identifier)
}

// ResolvePathToRoot finds the workspace root that contains the given absolute path
// Returns nil if no workspace contains the path
func (m *Manager) ResolvePathToRoot(absolutePath string) *WorkspaceRoot {
	// Sort by path length (longest first) to handle nested workspaces correctly
	// For example, if we have /home/user/project and /home/user/project/subdir,
	// we want to match /home/user/project/subdir first
	sortedRoots := append([]WorkspaceRoot{}, m.roots...)

	// Sort by descending path length
	for i := 0; i < len(sortedRoots)-1; i++ {
		for j := i + 1; j < len(sortedRoots); j++ {
			if len(sortedRoots[j].Path) > len(sortedRoots[i].Path) {
				sortedRoots[i], sortedRoots[j] = sortedRoots[j], sortedRoots[i]
			}
		}
	}

	// Find first matching root
	for _, root := range sortedRoots {
		if hasPrefix(absolutePath, root.Path) {
			rootCopy := root
			return &rootCopy
		}
	}

	return nil
}

// GetRootByName finds a workspace root by its name
func (m *Manager) GetRootByName(name string) *WorkspaceRoot {
	for _, root := range m.roots {
		if root.Name == name {
			rootCopy := root
			return &rootCopy
		}
	}
	return nil
}

// GetRootByIndex returns the workspace root at the given index
func (m *Manager) GetRootByIndex(index int) *WorkspaceRoot {
	if index < 0 || index >= len(m.roots) {
		return nil
	}
	root := m.roots[index]
	return &root
}

// IsPathInWorkspace checks if a path is within any workspace root
func (m *Manager) IsPathInWorkspace(absolutePath string) bool {
	return m.ResolvePathToRoot(absolutePath) != nil
}

// GetRelativePathFromRoot returns the relative path from a workspace root
// If root is nil, uses the workspace that contains the path
func (m *Manager) GetRelativePathFromRoot(absolutePath string, root *WorkspaceRoot) (string, error) {
	targetRoot := root
	if targetRoot == nil {
		targetRoot = m.ResolvePathToRoot(absolutePath)
	}

	if targetRoot == nil {
		return "", fmt.Errorf("path is not in any workspace: %s", absolutePath)
	}

	relPath, err := filepath.Rel(targetRoot.Path, absolutePath)
	if err != nil {
		return "", fmt.Errorf("failed to get relative path: %w", err)
	}

	return relPath, nil
}

// IsSingleRoot returns true if this manager has only one workspace root
func (m *Manager) IsSingleRoot() bool {
	return len(m.roots) == 1
}

// GetSingleRoot returns the single workspace root
// Returns an error if there are multiple roots
func (m *Manager) GetSingleRoot() (*WorkspaceRoot, error) {
	if len(m.roots) != 1 {
		return nil, fmt.Errorf("expected single root, but found %d roots", len(m.roots))
	}
	root := m.roots[0]
	return &root, nil
}

// CreateContext creates a workspace context for tool execution
func (m *Manager) CreateContext(currentRoot *WorkspaceRoot) WorkspaceContext {
	ctx := WorkspaceContext{
		Roots:       m.GetRoots(),
		PrimaryRoot: m.GetPrimaryRoot(),
	}

	if currentRoot != nil {
		ctx.CurrentRoot = currentRoot
	} else {
		ctx.CurrentRoot = m.GetPrimaryRoot()
	}

	return ctx
}

// UpdateCommitHashes refreshes the commit hashes for all Git workspaces
func (m *Manager) UpdateCommitHashes() error {
	for i := range m.roots {
		if m.roots[i].VCS == VCSTypeGit {
			if hash, err := getGitCommitHash(m.roots[i].Path); err == nil {
				m.roots[i].CommitHash = &hash
			}
		}
	}
	return nil
}

// BuildEnvironmentContext creates a structured environment context for LLM prompts
func (m *Manager) BuildEnvironmentContext() (string, error) {
	if len(m.roots) == 0 {
		return "", nil
	}

	envContext := EnvironmentContext{
		Workspaces: make(map[string]WorkspaceMetadata),
	}

	for _, root := range m.roots {
		metadata := WorkspaceMetadata{
			Hint: root.Name,
		}

		// Add Git information if available
		if root.VCS == VCSTypeGit {
			if len(root.RemoteURLs) > 0 {
				metadata.AssociatedRemoteURLs = root.RemoteURLs
			}

			if root.CommitHash != nil {
				metadata.LatestGitCommitHash = *root.CommitHash
			}
		}

		envContext.Workspaces[root.Path] = metadata
	}

	// Convert to JSON
	jsonBytes, err := json.MarshalIndent(envContext, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal environment context: %w", err)
	}

	return string(jsonBytes), nil
}

// GetSummary returns a human-readable summary of the workspace configuration
func (m *Manager) GetSummary() string {
	if len(m.roots) == 0 {
		return "No workspace roots configured"
	}

	if len(m.roots) == 1 {
		root := m.roots[0]
		summary := fmt.Sprintf("Single workspace: %s", root.Name)
		if root.VCS != VCSTypeNone {
			summary += fmt.Sprintf(" (%s)", root.VCS)
		}
		return summary
	}

	primary := m.GetPrimaryRoot()
	var otherNames []string
	for i, root := range m.roots {
		if i != m.primaryIndex {
			otherNames = append(otherNames, root.Name)
		}
	}

	return fmt.Sprintf("Multi-workspace (%d roots)\nPrimary: %s\nAdditional: %s",
		len(m.roots), primary.Name, strings.Join(otherNames, ", "))
}

// ToJSON serializes the manager state for storage
func (m *Manager) ToJSON() (string, error) {
	data := struct {
		Roots        []WorkspaceRoot `json:"roots"`
		PrimaryIndex int             `json:"primaryIndex"`
	}{
		Roots:        m.roots,
		PrimaryIndex: m.primaryIndex,
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal manager state: %w", err)
	}

	return string(jsonBytes), nil
}

// FromJSON deserializes a manager from JSON
func FromJSON(jsonData string) (*Manager, error) {
	var data struct {
		Roots        []WorkspaceRoot `json:"roots"`
		PrimaryIndex int             `json:"primaryIndex"`
	}

	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manager state: %w", err)
	}

	return NewManager(data.Roots, data.PrimaryIndex), nil
}

// hasPrefix checks if a path has the given prefix, handling path separators correctly
func hasPrefix(path, prefix string) bool {
	// Normalize paths
	path = filepath.Clean(path)
	prefix = filepath.Clean(prefix)

	// Check for exact match
	if path == prefix {
		return true
	}

	// Check for prefix with separator
	return strings.HasPrefix(path, prefix+string(filepath.Separator))
}
