# Workspace Package

The `workspace` package provides sophisticated workspace management for the adk-code, including multi-root workspace support, VCS detection, and intelligent path resolution.

## Features

### 1. Multi-Root Workspace Support

- Manage multiple project directories simultaneously
- Primary workspace concept for default operations
- Workspace detection and configuration
- Backward compatible with single-workspace mode

### 2. VCS Detection

- Automatic Git repository detection
- Mercurial repository detection  
- Commit hash tracking for Git repositories
- Remote URL extraction for Git repositories
- VCS-aware workspace metadata

### 3. Intelligent Path Resolution

- Resolve paths across multiple workspaces
- Handle both absolute and relative paths
- Workspace hint syntax (`@workspace:path`)
- Disambiguation for ambiguous paths
- Workspace-aware path operations

### 4. Rich Environment Context

- Generate structured workspace metadata for LLM
- Include Git remote URLs and commit hashes
- Workspace names and paths
- JSON format for easy parsing

## Quick Start

### Single Workspace (Default)

```go
import "adk-code/workspace"

// Create a workspace manager from a single directory
manager, err := workspace.FromSingleDirectory("/path/to/project")
if err != nil {
    log.Fatal(err)
}

// Get workspace information
primary := manager.GetPrimaryRoot()
fmt.Printf("Workspace: %s\n", primary.Name)
fmt.Printf("Path: %s\n", primary.Path)
fmt.Printf("VCS: %s\n", primary.VCS)

// Build environment context for LLM
envContext, err := manager.BuildEnvironmentContext()
if err != nil {
    log.Fatal(err)
}
fmt.Println(envContext)
```

### Path Resolution

```go
// Create a resolver
resolver := workspace.NewResolver(manager)

// Resolve a relative path
resolved, err := resolver.ResolvePath("src/main.go", nil)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Absolute path: %s\n", resolved.AbsolutePath)
fmt.Printf("Workspace: %s\n", resolved.Root.Name)
fmt.Printf("Relative path: %s\n", resolved.RelativePath)
```

### Workspace Hints

```go
// Parse workspace hint from path
hint, path := workspace.ParseWorkspaceHint("@frontend:src/index.ts")
// hint = "frontend", path = "src/index.ts"

// Resolve with workspace hint
resolved, err := resolver.ResolvePath(path, hint)
if err != nil {
    log.Fatal(err)
}

// Format path with hint
formatted := workspace.FormatPathWithHint("backend", "api/server.go")
// formatted = "@backend:api/server.go"
```

## API Reference

### Types

#### WorkspaceRoot

Represents a single workspace directory with metadata.

```go
type WorkspaceRoot struct {
    Path       string   // Absolute path to workspace root
    Name       string   // Display name (e.g., "frontend", "backend")
    VCS        VCSType  // Version control system type
    CommitHash *string  // Latest commit hash (for Git)
    RemoteURLs []string // Git remote URLs
}
```

#### VCSType

Version control system type constants.

```go
type VCSType string

const (
    VCSTypeGit       VCSType = "git"
    VCSTypeMercurial VCSType = "mercurial"
    VCSTypeNone      VCSType = "none"
)
```

#### ResolvedPath

Result of path resolution.

```go
type ResolvedPath struct {
    AbsolutePath string        // Resolved absolute path
    Root         *WorkspaceRoot // Containing workspace
    RelativePath string        // Path relative to workspace root
}
```

### Manager Methods

#### NewManager

```go
func NewManager(roots []WorkspaceRoot, primaryIndex int) *Manager
```

Creates a new workspace manager with the given roots.

#### FromSingleDirectory

```go
func FromSingleDirectory(cwd string) (*Manager, error)
```

Creates a workspace manager from a single directory (backward compatible mode).

#### GetRoots

```go
func (m *Manager) GetRoots() []WorkspaceRoot
```

Returns all workspace roots.

#### GetPrimaryRoot

```go
func (m *Manager) GetPrimaryRoot() *WorkspaceRoot
```

Returns the primary workspace root.

#### ResolvePathToRoot

```go
func (m *Manager) ResolvePathToRoot(absolutePath string) *WorkspaceRoot
```

Finds the workspace root that contains the given absolute path.

#### GetRootByName

```go
func (m *Manager) GetRootByName(name string) *WorkspaceRoot
```

Finds a workspace root by its name.

#### IsPathInWorkspace

```go
func (m *Manager) IsPathInWorkspace(absolutePath string) bool
```

Checks if a path is within any workspace root.

#### GetRelativePathFromRoot

```go
func (m *Manager) GetRelativePathFromRoot(absolutePath string, root *WorkspaceRoot) (string, error)
```

Returns the relative path from a workspace root.

#### BuildEnvironmentContext

```go
func (m *Manager) BuildEnvironmentContext() (string, error)
```

Creates structured workspace metadata JSON for LLM context.

#### GetSummary

```go
func (m *Manager) GetSummary() string
```

Returns a human-readable summary of the workspace configuration.

### Resolver Methods

#### NewResolver

```go
func NewResolver(manager *Manager) *Resolver
```

Creates a new path resolver.

#### ResolvePath

```go
func (r *Resolver) ResolvePath(path string, workspaceHint *string) (*ResolvedPath, error)
```

Resolves a path (with optional workspace hint) to an absolute path.

#### ResolvePathString

```go
func (r *Resolver) ResolvePathString(pathWithHint string) (*ResolvedPath, error)
```

Convenience method that resolves a path string that may contain workspace hints.

#### ParseWorkspaceHint

```go
func ParseWorkspaceHint(input string) (workspaceHint *string, path string)
```

Parses the workspace hint syntax: `@workspaceName:path`

#### FormatPathWithHint

```go
func FormatPathWithHint(workspaceName, path string) string
```

Formats a path with a workspace hint.

## Examples

### Example 1: Basic Workspace Setup

```go
package main

import (
    "fmt"
    "log"
    "adk-code/workspace"
)

func main() {
    // Create workspace manager
    manager, err := workspace.FromSingleDirectory("/home/user/myproject")
    if err != nil {
        log.Fatal(err)
    }
    
    // Get workspace info
    root := manager.GetPrimaryRoot()
    fmt.Printf("Working in: %s\n", root.Name)
    fmt.Printf("VCS: %s\n", root.VCS)
    
    if root.CommitHash != nil {
        fmt.Printf("Commit: %s\n", *root.CommitHash)
    }
    
    // Print summary
    fmt.Println(manager.GetSummary())
}
```

### Example 2: Path Resolution

```go
package main

import (
    "fmt"
    "log"
    "adk-code/workspace"
)

func main() {
    manager, _ := workspace.FromSingleDirectory("/home/user/myproject")
    resolver := workspace.NewResolver(manager)
    
    // Resolve relative path
    resolved, err := resolver.ResolvePath("src/main.go", nil)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Absolute: %s\n", resolved.AbsolutePath)
    fmt.Printf("Workspace: %s\n", resolved.Root.Name)
}
```

### Example 3: Environment Context for LLM

```go
package main

import (
    "fmt"
    "log"
    "adk-code/workspace"
)

func main() {
    manager, _ := workspace.FromSingleDirectory("/home/user/myproject")
    
    // Build environment context
    envContext, err := manager.BuildEnvironmentContext()
    if err != nil {
        log.Fatal(err)
    }
    
    // This JSON can be included in LLM prompts
    fmt.Println(envContext)
    // Output:
    // {
    //   "workspaces": {
    //     "/home/user/myproject": {
    //       "hint": "myproject",
    //       "associatedRemoteUrls": [
    //         "https://github.com/user/myproject.git"
    //       ],
    //       "latestGitCommitHash": "abc123..."
    //     }
    //   }
    // }
}
```

## Architecture

### Design Principles

1. **Backward Compatibility**: Single-workspace mode works exactly like before
2. **Incremental Adoption**: Multi-workspace features can be enabled gradually
3. **Fail Gracefully**: VCS detection failures don't prevent operation
4. **Performance**: Caching and lazy loading for efficiency
5. **Extensibility**: Easy to add new VCS types or features

### Components

```
workspace/
├── types.go      - Core data structures
├── manager.go    - Workspace management logic
├── resolver.go   - Path resolution logic
├── vcs.go        - VCS detection and metadata
└── README.md     - This file
```

### Data Flow

```
┌─────────────┐
│   Agent     │
└──────┬──────┘
       │ creates
       ▼
┌─────────────┐
│  Manager    │◄─────┐
└──────┬──────┘      │
       │ uses        │
       ▼             │
┌─────────────┐      │
│  Resolver   │──────┘
└──────┬──────┘
       │ resolves paths
       ▼
┌─────────────┐
│   Tools     │
└─────────────┘
```

## Phase 2 Features (NEW!)

### File Existence Checking

The resolver now checks if files actually exist before claiming a workspace contains them:

```go
resolver := workspace.NewResolver(manager)

// DisambiguatePath only returns workspaces that actually contain the file
matches := resolver.DisambiguatePath("src/main.go")
fmt.Printf("File exists in: %v\n", matches)

// ResolvePathWithDisambiguation intelligently chooses the best match
resolved, err := resolver.ResolvePathWithDisambiguation("config.yaml")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Resolved to: %s\n", resolved.Root.Name)

// FileExists checks if a file exists in any workspace
if resolver.FileExists("README.md") {
    fmt.Println("README found!")
}
```

### Workspace Configuration Persistence

Save and load workspace configurations to/from `.workspace.json`:

```go
// Save configuration
manager := workspace.NewManager(roots, primaryIndex)
prefs := workspace.DefaultPreferences()
prefs.AutoDetectWorkspaces = true
prefs.MaxWorkspaces = 10

err := workspace.SaveManagerToDirectory("/path/to/project", manager, &prefs)
if err != nil {
    log.Fatal(err)
}

// Load configuration
loadedManager, prefs, err := workspace.LoadManagerFromDirectory("/path/to/project")
if err != nil {
    log.Fatal(err)
}

// Check if config exists
if workspace.ConfigExists("/path/to/project") {
    fmt.Println("Workspace configuration found")
}
```

#### Configuration File Format

The `.workspace.json` file structure:

```json
{
  "version": 1,
  "roots": [
    {
      "path": "/home/user/frontend",
      "name": "frontend",
      "vcs": "git",
      "commitHash": "abc123...",
      "remoteUrls": ["https://github.com/user/frontend.git"]
    },
    {
      "path": "/home/user/backend",
      "name": "backend",
      "vcs": "git",
      "commitHash": "def456...",
      "remoteUrls": ["https://github.com/user/backend.git"]
    }
  ],
  "primaryIndex": 0,
  "preferences": {
    "autoDetectWorkspaces": true,
    "maxWorkspaces": 10,
    "preferVCSRoots": true,
    "includeHidden": false
  }
}
```

### Multi-Workspace Detection

Automatically discover multiple projects in a directory:

```go
// Use default detection options
options := workspace.DefaultDetectionOptions()
roots, err := workspace.DetectWorkspaces("/path/to/parent", options)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Found %d workspaces:\n", len(roots))
for _, root := range roots {
    fmt.Printf("  - %s (%s)\n", root.Name, root.Path)
}

// Customize detection
options.MaxDepth = 2              // Limit search depth
options.MaxWorkspaces = 5         // Limit number of workspaces
options.PreferVCSRoots = true     // Prioritize VCS repositories
options.IncludeHidden = false     // Skip hidden directories

// Detection works with various project markers:
// - VCS: .git, .hg
// - Go: go.mod
// - Node: package.json
// - Rust: Cargo.toml
// - Python: setup.py, pyproject.toml, Pipfile
// - Java: pom.xml, build.gradle
// - .NET: *.csproj, *.sln
// - And more...
```

### Smart Workspace Initialization

One-step initialization that tries config → detection → fallback:

```go
// Tries in order:
// 1. Load from .workspace.json if it exists
// 2. Auto-detect workspaces if no config
// 3. Fall back to single directory
manager, err := workspace.SmartWorkspaceInitialization("/path/to/project")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Initialized with %d workspaces\n", len(manager.GetRoots()))
```

### Workspace Switching

Dynamically change the primary workspace:

```go
manager := workspace.NewManager(roots, 0)

// Switch by name
err := manager.SetPrimaryByName("backend")
if err != nil {
    log.Fatal(err)
}

// Switch by path
err = manager.SetPrimaryByPath("/path/to/frontend")
if err != nil {
    log.Fatal(err)
}

// Switch by index
err = manager.SetPrimaryIndex(1)
if err != nil {
    log.Fatal(err)
}

// Universal switch method (tries name, then path)
newPrimary, err := manager.SwitchWorkspace("backend")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Switched to: %s\n", newPrimary.Name)
```

### Workspace-Aware Tools

Tools can now resolve paths with workspace support:

```go
import "adk-code/tools"

workspaceTools := tools.NewWorkspaceTools(resolver)

// Resolve a path (supports @workspace:path hints)
resolvedPath, err := workspaceTools.ResolvePath("@frontend:src/index.ts")
if err != nil {
    log.Fatal(err)
}

// Format a path with its workspace hint
hint := workspaceTools.FormatPathWithHint("/home/user/frontend/src/main.go")
// Returns: "@frontend:src/main.go"

// Parse workspace hints
name, path, hasHint := tools.ParseWorkspaceHint("@backend:api/server.go")
// name = "backend", path = "api/server.go", hasHint = true
```

## Enhanced Examples

### Example 4: Multi-Workspace Project

```go
package main

import (
    "fmt"
    "log"
    "adk-code/workspace"
)

func main() {
    // Auto-detect workspaces in a monorepo
    manager, err := workspace.SmartWorkspaceInitialization("/home/user/monorepo")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Detected workspaces:\n")
    for _, root := range manager.GetRoots() {
        fmt.Printf("  - %s at %s (%s)\n", root.Name, root.Path, root.VCS)
    }
    
    // Create resolver
    resolver := workspace.NewResolver(manager)
    
    // Check file existence across workspaces
    matches := resolver.DisambiguatePath("README.md")
    fmt.Printf("\nREADME.md found in: %v\n", matches)
    
    // Switch primary workspace
    manager.SwitchWorkspace("backend")
    fmt.Printf("Primary workspace: %s\n", manager.GetPrimaryRoot().Name)
}
```

### Example 5: Configuration Management

```go
package main

import (
    "fmt"
    "log"
    "adk-code/workspace"
)

func main() {
    // Create a multi-workspace setup
    roots := []workspace.WorkspaceRoot{
        {Path: "/home/user/frontend", Name: "frontend", VCS: workspace.VCSTypeGit},
        {Path: "/home/user/backend", Name: "backend", VCS: workspace.VCSTypeGit},
    }
    manager := workspace.NewManager(roots, 0)
    
    // Configure preferences
    prefs := workspace.DefaultPreferences()
    prefs.MaxWorkspaces = 5
    prefs.PreferVCSRoots = true
    
    // Save configuration
    err := workspace.SaveManagerToDirectory("/home/user", manager, &prefs)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Configuration saved to .workspace.json")
    
    // Later: load configuration
    loadedManager, loadedPrefs, err := workspace.LoadManagerFromDirectory("/home/user")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Loaded %d workspaces\n", len(loadedManager.GetRoots()))
    fmt.Printf("Max workspaces: %d\n", loadedPrefs.MaxWorkspaces)
}
```

## Implementation Status

### Phase 1 ✅
- ✅ Single workspace support
- ✅ VCS detection (Git, Mercurial)
- ✅ Path resolution
- ✅ Environment context
- ✅ Workspace hints support

### Phase 2 ✅
- ✅ Multi-workspace detection
- ✅ Workspace switching
- ✅ File existence checking
- ✅ Workspace configuration persistence
- ✅ Workspace-aware tools
- ✅ Smart initialization

### Phase 3 (Planned)
- [ ] Cross-workspace operations
- [ ] Workspace templates
- [ ] Advanced VCS integration (branches, tags, diffs)
- [ ] Workspace dependency tracking

## Contributing

When adding new features:

1. Maintain backward compatibility
2. Add tests for new functionality
3. Update this README
4. Follow Go conventions and best practices
5. Consider performance implications

## See Also

- [CLINE Comparison Document](../../doc/WORKSPACE_ENVIRONMENT_COMPARISON.md) - Detailed comparison with CLINE
- [Agent Documentation](../agent/README.md) - How the agent uses workspaces
- [Tool Documentation](../tools/README.md) - How tools interact with workspaces
