// Package workspace provides workspace management for the coding agent,
// including multi-root workspace support, VCS detection, and path resolution.
package workspace

// VCSType represents the version control system type for a workspace
type VCSType string

const (
	// VCSTypeGit represents a Git repository
	VCSTypeGit VCSType = "git"
	// VCSTypeMercurial represents a Mercurial repository
	VCSTypeMercurial VCSType = "mercurial"
	// VCSTypeNone represents no version control
	VCSTypeNone VCSType = "none"
)

// WorkspaceRoot represents a single workspace directory
type WorkspaceRoot struct {
	// Path is the absolute path to the workspace root directory
	Path string `json:"path"`
	// Name is the display name for the workspace (e.g., "frontend", "backend")
	Name string `json:"name"`
	// VCS is the version control system type detected for this workspace
	VCS VCSType `json:"vcs"`
	// CommitHash is the latest commit hash (for Git workspaces)
	CommitHash *string `json:"commitHash,omitempty"`
	// RemoteURLs are the Git remote URLs associated with this workspace
	RemoteURLs []string `json:"remoteUrls,omitempty"`
}

// WorkspaceContext provides context about the workspace environment
// for use in prompts and tool execution
type WorkspaceContext struct {
	// Roots is the list of all workspace roots
	Roots []WorkspaceRoot `json:"roots"`
	// PrimaryRoot is the main workspace root
	PrimaryRoot *WorkspaceRoot `json:"primaryRoot"`
	// CurrentRoot is the workspace root currently being operated on (optional)
	CurrentRoot *WorkspaceRoot `json:"currentRoot,omitempty"`
}

// ResolvedPath represents a path that has been resolved to a specific workspace
type ResolvedPath struct {
	// AbsolutePath is the resolved absolute path
	AbsolutePath string
	// Root is the workspace root that contains this path
	Root *WorkspaceRoot
	// RelativePath is the path relative to the workspace root
	RelativePath string
}

// WorkspaceMetadata represents metadata for a workspace (for LLM context)
type WorkspaceMetadata struct {
	// Hint is the workspace name/identifier
	Hint string `json:"hint"`
	// AssociatedRemoteURLs are the Git remote URLs
	AssociatedRemoteURLs []string `json:"associatedRemoteUrls,omitempty"`
	// LatestGitCommitHash is the latest commit hash
	LatestGitCommitHash string `json:"latestGitCommitHash,omitempty"`
}

// EnvironmentContext represents the complete workspace environment context
// for inclusion in LLM prompts
type EnvironmentContext struct {
	// Workspaces maps workspace paths to their metadata
	Workspaces map[string]WorkspaceMetadata `json:"workspaces"`
}
