// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package workspace

import (
	"context"
)

// PathResolver provides intelligent path resolution across multiple workspaces
// It handles both relative and absolute paths with optional workspace hints.
//
// Example usage:
//
//	resolver := NewResolver(manager)
//	resolved, err := resolver.ResolvePath("src/main.go")
//	// Resolved against primary workspace
//
//	resolved, err := resolver.ResolvePath("src/main.go", &hint)
//	// Resolved against specific workspace
type PathResolver interface {
	// ResolvePath converts relative/absolute paths to an absolute path
	// workspaceHint, if provided, specifies which workspace to resolve against
	// Returns nil if the workspace hint workspace is not found
	ResolvePath(path string, workspaceHint *string) (*ResolvedPath, error)

	// GetWorkspaceForPath returns the workspace name for a given path
	// Returns empty string if path is not in any workspace
	GetWorkspaceForPath(path string) string

	// ResolvePathString is a convenience method that handles workspace hint syntax
	// Workspace hints use the format: @workspaceName:relative/path
	ResolvePathString(pathWithHint string) (*ResolvedPath, error)
}

// ContextBuilder constructs environment context strings for LLM prompts
// It gathers information about the workspace structure, layout, and configuration
// that helps the LLM understand the context of code operations.
//
// Example output:
//
//	"""
//	Workspace: /home/user/project
//	Primary Language: Go
//	Structure:
//	  - cmd/      (executables)
//	  - pkg/      (libraries)
//	  - internal/ (private packages)
//	Build Tool: Go (go.mod)
//	VCS: Git (main branch)
//	"""
type ContextBuilder interface {
	// BuildEnvironmentContext generates a formatted context string describing
	// the workspace(s), their structure, and relevant metadata.
	// This string is included in the system prompt.
	BuildEnvironmentContext() (string, error)

	// BuildWorkspaceContext generates context for a specific workspace
	BuildWorkspaceContext(workspace *WorkspaceRoot) (string, error)

	// SetIncludeStructure controls whether file structure is included
	SetIncludeStructure(include bool)

	// SetMaxDepth controls how deeply to traverse directory structure
	SetMaxDepth(depth int)
}

// VCSDetector identifies version control systems and extracts metadata
// It supports multiple VCS systems (Git, Mercurial, etc.) and provides
// consistent access to common VCS information.
//
// Example:
//
//	detector := NewVCSDetector()
//	vcsType, err := detector.Detect("/path/to/repo")
//	// Returns VCSTypeGit
//
//	hash, err := detector.GetCommitHash("/path/to/repo")
//	// Returns current commit hash
type VCSDetector interface {
	// Detect identifies the VCS type for a directory
	// Returns VCSTypeNone if no VCS is detected
	Detect(path string) (VCSType, error)

	// GetCommitHash returns the current commit hash/revision
	GetCommitHash(path string) (string, error)

	// GetRemoteURLs returns URLs for configured remotes (git) or servers (hg)
	GetRemoteURLs(path string) ([]string, error)

	// GetBranch returns the current branch or bookmark name
	GetBranch(path string) (string, error)

	// IsClean checks if the working directory is clean (no uncommitted changes)
	IsClean(path string) (bool, error)

	// GetStatus returns a human-readable status string
	GetStatus(path string) (string, error)
}

// DefaultPathResolver returns the standard path resolver for the Manager
// This is the implementation that ships with the workspace package
func DefaultPathResolver(manager *Manager) PathResolver {
	return NewResolver(manager)
}

// DefaultContextBuilder returns the standard context builder for the Manager
func DefaultContextBuilder(roots []WorkspaceRoot) ContextBuilder {
	return NewDefaultContextBuilder(roots)
}

// DefaultVCSDetector returns the standard VCS detector
// It supports Git and Mercurial
func DefaultVCSDetector() VCSDetector {
	return NewDefaultVCSDetector()
}

// VCSDetectorWithContext is an optional extended interface for VCS detectors
// that support context-aware operations
type VCSDetectorWithContext interface {
	VCSDetector

	// DetectWithContext detects VCS type with a timeout context
	DetectWithContext(ctx context.Context, path string) (VCSType, error)

	// GetCommitHashWithContext gets the commit hash with a timeout
	GetCommitHashWithContext(ctx context.Context, path string) (string, error)
}

// ContextBuilderWithMetrics is an optional extended interface for context builders
// that track performance metrics
type ContextBuilderWithMetrics interface {
	ContextBuilder

	// LastBuildDuration returns the duration of the last context build
	LastBuildDuration() int64

	// LastFileCount returns the number of files included in the last build
	LastFileCount() int

	// LastDirectoryCount returns the number of directories scanned
	LastDirectoryCount() int
}

// PathResolverWithCache is an optional extended interface for path resolvers
// that support caching to improve performance
type PathResolverWithCache interface {
	PathResolver

	// ClearCache clears the resolution cache
	ClearCache()

	// GetCacheStats returns cache hit/miss statistics
	GetCacheStats() (hits int64, misses int64)
}

// NewDefaultContextBuilder creates the default implementation of ContextBuilder
// It analyzes the provided workspace roots and builds a formatted context string
func NewDefaultContextBuilder(roots []WorkspaceRoot) ContextBuilder {
	return &defaultContextBuilder{
		roots:            roots,
		includeStructure: true,
		maxDepth:         2,
	}
}

// defaultContextBuilder is the standard implementation of ContextBuilder
type defaultContextBuilder struct {
	roots            []WorkspaceRoot
	includeStructure bool
	maxDepth         int
}

// BuildEnvironmentContext generates context for all workspaces
func (b *defaultContextBuilder) BuildEnvironmentContext() (string, error) {
	// This will be implemented in the context builder implementation file
	// For now, return a basic structure
	return "Workspace context", nil
}

// BuildWorkspaceContext generates context for a specific workspace
func (b *defaultContextBuilder) BuildWorkspaceContext(workspace *WorkspaceRoot) (string, error) {
	// This will be implemented in the context builder implementation file
	return "Workspace: " + workspace.Name, nil
}

// SetIncludeStructure sets whether to include file structure
func (b *defaultContextBuilder) SetIncludeStructure(include bool) {
	b.includeStructure = include
}

// SetMaxDepth sets the maximum directory traversal depth
func (b *defaultContextBuilder) SetMaxDepth(depth int) {
	b.maxDepth = depth
}

// NewDefaultVCSDetector creates the default implementation of VCSDetector
func NewDefaultVCSDetector() VCSDetector {
	return &defaultVCSDetector{}
}

// defaultVCSDetector is the standard implementation of VCSDetector
type defaultVCSDetector struct{}

// Detect identifies the VCS type
func (d *defaultVCSDetector) Detect(path string) (VCSType, error) {
	return detectVCS(path)
}

// GetCommitHash gets the current commit hash
func (d *defaultVCSDetector) GetCommitHash(path string) (string, error) {
	return getGitCommitHash(path)
}

// GetRemoteURLs gets remote URLs
func (d *defaultVCSDetector) GetRemoteURLs(path string) ([]string, error) {
	return getGitRemoteURLs(path)
}

// GetBranch gets the current branch
func (d *defaultVCSDetector) GetBranch(path string) (string, error) {
	// This will delegate to existing branch detection functions
	// For now, return a placeholder
	return "", nil
}

// IsClean checks if working directory is clean
func (d *defaultVCSDetector) IsClean(path string) (bool, error) {
	// This will be implemented to check git/hg status
	return true, nil
}

// GetStatus gets a status string
func (d *defaultVCSDetector) GetStatus(path string) (string, error) {
	// This will be implemented to return git/hg status
	return "", nil
}
