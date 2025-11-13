// Package workspace provides multi-root workspace management with VCS detection
// (Git, Mercurial) and path resolution.
//
// The Manager handles one or more workspace roots and provides utilities for:
// - Path resolution with workspace hints
// - VCS metadata (commit hash, remote URLs)
// - Multi-root workspace support for monorepos
// - Project root detection
//
// Key components:
// - Manager: Multi-root workspace manager
// - WorkspaceRoot: Individual workspace metadata
// - Resolver: Path resolution with VCS awareness
// - Detector: VCS detection (Git/Mercurial)
//
// Features:
// - Single-root mode: Backward compatible with existing code
// - Multi-root mode: Support for monorepos
// - VCS detection: Automatically detects Git or Mercurial repositories
// - Metadata: Captures commit hash and remote URLs for context
// - Path resolution: Resolves relative paths within workspace context
//
// Example:
//
//	manager, err := workspace.FromSingleDirectory(cwd)
//	if err != nil {
//		return err
//	}
//	root := manager.GetPrimaryRoot()
//	fmt.Printf("Workspace: %s\n", root.Path)
//	fmt.Printf("VCS: %s\n", root.VCS)
//	if root.CommitHash != nil {
//		fmt.Printf("Commit: %s\n", *root.CommitHash)
//	}
//
// The package provides:
// - Automatic workspace detection from working directory
// - VCS type detection (Git or Mercurial)
// - Git remote URL extraction
// - Git commit hash retrieval
// - Path resolution with custom hints
package workspace
