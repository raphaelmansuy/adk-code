# API Surface Documentation

**Generated**: November 12, 2025  
**Purpose**: Document all exported types and functions to identify public vs internal APIs  
**Scope**: Catalog of all exported identifiers across packages

## Executive Summary

This document catalogs all exported Go identifiers (types, functions, constants, variables) across the codebase. Items are marked as:
- **STABLE** ✅ - Public API, maintain backward compatibility
- **INTERNAL** 🔒 - Implementation detail, can change freely
- **DEPRECATE** ⚠️ - Old API, plan replacement

---

## Package: agent

**Stability**: Core, STABLE  
**Purpose**: ADK LLMAgent wrapper and tool coordination

### Types
- `Agent` ✅ STABLE - Main agent orchestrator
- `Config` ✅ STABLE - Agent configuration
- `PromptContext` ✅ STABLE - Prompt generation context

### Functions
- `NewAgent(ctx, config)` ✅ STABLE - Create agent instance
- `BuildEnhancedPrompt(context, tools)` ✅ STABLE - Generate system prompt

### Interfaces
- (None exposed at package level)

### Constants
- Various prompt-related constants (INTERNAL for most)

---

## Package: internal/app

**Stability**: Application layer, INTERNAL  
**Purpose**: Application orchestration and lifecycle

### Types
- `Application` 🔒 INTERNAL - Main app struct
- `Config` 🔒 INTERNAL - App configuration
- Various component structs (Display, Session, Model, etc.)

### Functions
- `NewApplication(ctx)` 🔒 INTERNAL - Create application
- `InitializeXxx()` functions (Display, Agent, Session, etc.) - 🔒 INTERNAL

### Purpose
- These should NOT be used outside internal/app package
- May be refactored in Phase 3

---

## Package: internal/config

**Stability**: Configuration, INTERNAL  
**Purpose**: Configuration type definitions

### Types
- `Config` 🔒 INTERNAL - Application configuration struct
- Various field types for config

### Functions
- None (data-only package)

### Note
- This package exists but has minimal usage
- May consolidate with internal/app in Phase 3

---

## Package: internal/data

**Stability**: Data layer abstraction, STABLE  
**Purpose**: Repository pattern interfaces

### Types
- `Repository` ✅ STABLE - Session repository interface
- Various query/result types

### Interfaces
- `SessionRepository` ✅ STABLE - Define contract for session storage

### Functions
- None (interface definitions)

### Note
- Core abstraction for data persistence
- Implementations in internal/data/sqlite and internal/data/memory

---

## Package: internal/data/sqlite

**Stability**: SQLite implementation, INTERNAL  
**Purpose**: SQLite-based session storage

### Types
- `SQLiteSessionRepository` 🔒 INTERNAL - Implementation type

### Functions
- `NewSQLiteSessionRepository(dbPath)` 🔒 INTERNAL - Create repository

### Note
- Implement via Repository interface
- Direct usage discouraged, use session.Manager instead

---

## Package: internal/data/memory

**Stability**: In-memory implementation, INTERNAL  
**Purpose**: In-memory session storage (testing)

### Types
- `MemorySessionRepository` 🔒 INTERNAL - In-memory implementation

### Functions
- `NewMemorySessionRepository()` 🔒 INTERNAL - Create repository

---

## Package: internal/llm

**Stability**: LLM abstraction, STABLE  
**Purpose**: LLM provider abstraction layer

### Types
- `Provider` ✅ STABLE - LLM provider interface
- Various backend-specific types

### Interfaces
- `Provider` ✅ STABLE - Abstract LLM provider

### Functions
- `GetProvider(providerName)` ✅ STABLE - Get provider instance
- `NewProvider(config)` ✅ STABLE - Create provider

---

## Package: internal/llm/backends

**Stability**: Provider implementations, INTERNAL  
**Purpose**: Specific LLM provider implementations

### Types
- `GeminiBackend` 🔒 INTERNAL
- `OpenAIBackend` 🔒 INTERNAL
- `VertexAIBackend` 🔒 INTERNAL

### Functions
- `NewGeminiBackend(apiKey)` 🔒 INTERNAL
- `NewOpenAIBackend(apiKey)` 🔒 INTERNAL
- `NewVertexAIBackend(config)` 🔒 INTERNAL

### Note
- Use internal/llm provider interface instead
- Direct usage discouraged

---

## Package: pkg/cli

**Stability**: CLI utilities, STABLE  
**Purpose**: Command-line interface utilities and model resolution

### Types
- `Config` ✅ STABLE - CLI configuration
- `ModelResolver` ✅ STABLE - Model resolution logic
- Various command types ✅ STABLE

### Functions
- `NewModelResolver()` ✅ STABLE - Create resolver
- `ParseModelString(input)` ✅ STABLE - Parse model specification
- `ResolveModel(input)` ✅ STABLE - Resolve to full model ID
- Various command functions ✅ STABLE

### Constants
- (Model registries and command definitions)

---

## Package: pkg/cli/commands

**Stability**: Command handlers, STABLE  
**Purpose**: CLI command implementations

### Types
- Various command handler types ✅ STABLE

### Functions
- `HandleXxxCommand(ctx, args)` ✅ STABLE - Command handlers

---

## Package: pkg/errors

**Stability**: Error types, STABLE  
**Purpose**: Application error types and utilities

### Types
- `AgentError` ✅ STABLE - Base error type
- `FileNotFoundError` ✅ STABLE
- `PermissionDeniedError` ✅ STABLE
- `PathTraversalError` ✅ STABLE
- `SymlinkEscapeError` ✅ STABLE
- `ExecutionError` ✅ STABLE
- `TimeoutError` ✅ STABLE
- `APIKeyError` ✅ STABLE
- `ModelNotFoundError` ✅ STABLE
- `ProviderError` ✅ STABLE
- `PatchFailedError` ✅ STABLE
- `InternalError` ✅ STABLE
- `NotSupportedError` ✅ STABLE

### Functions
- `NewAgentError(...)` ✅ STABLE - Create error
- `WrapError(...)` ✅ STABLE - Wrap existing error
- `WithContext(...)` ✅ STABLE - Add context to error
- `IsFunction(func)` ✅ STABLE - Check if error type

### Note
- Excellent test coverage (92.3%)
- High stability for error handling

---

## Package: pkg/models

**Stability**: Model configuration, STABLE  
**Purpose**: Model registry and resolution

### Types
- `Model` ✅ STABLE - Model configuration
- `Backend` ✅ STABLE - Backend identifier
- `ModelRegistry` ✅ STABLE - Registry of available models

### Functions
- `GetModel(id)` ✅ STABLE - Get model by ID
- `ListModelsByBackend(backend)` ✅ STABLE - List backend models
- `GetBackends()` ✅ STABLE - List supported backends
- `ResolveModel(backend, model)` ✅ STABLE - Resolve model

### Constants
- Backend constants (Gemini, OpenAI, VertexAI)
- Model ID constants

---

## Package: pkg/models/factories

**Stability**: Model factories, INTERNAL  
**Purpose**: Factory implementations for model creation

### Types
- Various factory types 🔒 INTERNAL

### Functions
- Various factory functions 🔒 INTERNAL

---

## Package: session

**Stability**: Session management, STABLE  
**Purpose**: High-level session API

### Types
- `Manager` ✅ STABLE - Session manager
- `Session` ✅ STABLE - Session model
- Various event types ✅ STABLE

### Functions
- `NewManager(appName, dbPath)` ✅ STABLE - Create manager
- `CreateSession(name)` ✅ STABLE - Create new session
- `LoadSession(name)` ✅ STABLE - Load existing session
- `ListSessions()` ✅ STABLE - List all sessions
- `DeleteSession(name)` ✅ STABLE - Delete session
- `AppendEvent(sessionName, event)` ✅ STABLE - Add event to session

### Interfaces
- `Repository` ✅ STABLE - Repository pattern interface

---

## Package: tools (main)

**Stability**: Tool system, STABLE  
**Purpose**: Tool registry and common utilities

### Types
- `Tool` ✅ STABLE - Tool interface from ADK
- `ToolRegistry` ✅ STABLE - Registry of available tools
- Various tool-related types

### Functions
- `Register(tool)` ✅ STABLE - Register a tool
- `Get(toolName)` ✅ STABLE - Get tool by name
- `GetAll()` ✅ STABLE - Get all tools

---

## Package: tools/common

**Stability**: Tool utilities, INTERNAL  
**Purpose**: Shared tool utilities and registry

### Types
- `ToolRegistry` 🔒 INTERNAL - Registry implementation
- Various utility types 🔒 INTERNAL

### Functions
- Registry functions 🔒 INTERNAL

---

## Package: tools/file

**Stability**: File operations, STABLE  
**Purpose**: File I/O tool implementations

### Types
- `ReadFileTool` ✅ STABLE - Read file tool
- `WriteFileTool` ✅ STABLE - Write file tool
- Various input/output types ✅ STABLE

### Functions
- `NewReadFileTool()` ✅ STABLE
- `NewWriteFileTool()` ✅ STABLE
- `ValidatePath(path, basePath)` ✅ STABLE - Path validation
- `AtomicWrite(path, content)` ✅ STABLE - Safe file writing

---

## Package: tools/edit

**Stability**: File editing, STABLE  
**Purpose**: File editing operations

### Types
- `EditTool` ✅ STABLE - Edit tool
- `ReplaceInFileTool` ✅ STABLE - Replace text tool

### Functions
- `NewEditTool()` ✅ STABLE
- `NewReplaceInFileTool()` ✅ STABLE

---

## Package: tools/exec

**Stability**: Command execution, STABLE  
**Purpose**: Execute shell commands

### Types
- `ExecuteCommandTool` ✅ STABLE - Execute tool
- Various execution types ✅ STABLE

### Functions
- `NewExecuteCommandTool()` ✅ STABLE

---

## Package: tools/search

**Stability**: Search operations, STABLE  
**Purpose**: Workspace search tools

### Types
- `SearchTool` ✅ STABLE - Search implementation
- Various result types ✅ STABLE

### Functions
- `NewSearchTool()` ✅ STABLE

---

## Package: tools/workspace

**Stability**: Workspace analysis, STABLE  
**Purpose**: Workspace manipulation tools

### Types
- Various workspace tool types ✅ STABLE

### Functions
- Tool constructors ✅ STABLE

---

## Package: tools/display

**Stability**: Display messaging, STABLE  
**Purpose**: Tool for agents to display messages

### Types
- `DisplayMessageTool` ✅ STABLE - Message display tool
- `UpdateTaskListTool` ✅ STABLE - Task list update tool
- Input/output types ✅ STABLE

### Functions
- `NewDisplayMessageTool()` ✅ STABLE
- `NewUpdateTaskListTool()` ✅ STABLE

---

## Package: tools/v4a

**Stability**: V4A patch format, STABLE  
**Purpose**: Unified V4A patch format support

### Types
- `Patch` ✅ STABLE - Patch representation
- `Hunk` ✅ STABLE - Patch hunk
- Various types ✅ STABLE

### Functions
- `Parse(patchString)` ✅ STABLE - Parse V4A patch
- `Apply(content, patch)` ✅ STABLE - Apply patch to content
- `ApplyDryRun(content, patch)` ✅ STABLE - Preview patch

---

## Package: display

**Stability**: Display facade, STABLE  
**Purpose**: Main display API and component factory

### Types
- `Renderer` ✅ STABLE - Main display renderer
- `Components` ✅ STABLE - UI component collection
- Event types ✅ STABLE
- Various component types ✅ STABLE

### Functions
- `NewRenderer(format)` ✅ STABLE - Create renderer
- `NewComponents(config)` ✅ STABLE - Create component set
- Rendering methods ✅ STABLE

### Note
- This is the main display facade
- Backward compatibility maintained via re-exports

---

## Package: display/styles

**Stability**: Styling, STABLE  
**Purpose**: ANSI colors and text styling

### Types
- Color type constants ✅ STABLE
- Style type constants ✅ STABLE

### Functions
- `Colorize(text, color)` ✅ STABLE
- `Style(text, style)` ✅ STABLE

---

## Package: display/components

**Stability**: UI components, STABLE  
**Purpose**: Reusable UI components

### Types
- `Spinner` ✅ STABLE
- `Banner` ✅ STABLE
- `Typewriter` ✅ STABLE
- Various component types ✅ STABLE

### Functions
- `NewSpinner()` ✅ STABLE
- `NewBanner()` ✅ STABLE
- Component methods ✅ STABLE

---

## Package: display/renderers

**Stability**: Content rendering, STABLE  
**Purpose**: Render various content types

### Types
- `MarkdownRenderer` ✅ STABLE
- `TextRenderer` ✅ STABLE
- Various types ✅ STABLE

### Functions
- `NewMarkdownRenderer()` ✅ STABLE
- Rendering methods ✅ STABLE

---

## Package: workspace

**Stability**: Workspace detection, STABLE  
**Purpose**: Detect and manage workspace structure

### Types
- `WorkspaceRoot` ✅ STABLE - Workspace information
- `Config` ✅ STABLE - Workspace configuration

### Functions
- `GetProjectRoot(startPath)` ✅ STABLE - Find project root
- `DetectWorkspaces(rootPath)` ✅ STABLE - Detect workspaces
- Configuration functions ✅ STABLE

---

## Package: tracking

**Stability**: Token tracking, STABLE  
**Purpose**: Track token usage across sessions

### Types
- `SessionTokens` ✅ STABLE - Session token metrics
- `GlobalTracker` ✅ STABLE - Global token tracking

### Functions
- `NewSessionTokens()` ✅ STABLE
- `GetOrCreateSession(sessionName)` ✅ STABLE
- `GetGlobalSummary()` ✅ STABLE
- Formatting functions ✅ STABLE

---

## Dependency Stability Matrix

### STABLE Dependencies (Safe to use)
- pkg/errors - 92.3% test coverage
- pkg/models - 19.1% coverage but stable interface
- tools/v4a - 80.6% coverage
- display - Facade pattern ensures compatibility
- agent - Core orchestration

### IN-PROGRESS (Good but could improve)
- session - 49% coverage, stable interface
- workspace - 48% coverage, solid interface
- pkg/cli - 19.6% coverage

### NEEDS TESTING (No test files)
- internal/data - Interface only, OK
- internal/data/sqlite - Implementation needed
- internal/data/memory - Implementation, testing needed
- internal/llm/* - Provider integration untested
- tools/* (most) - Tool implementations need tests

---

## Backward Compatibility Guarantees

### GUARANTEED STABLE (Semantic Versioning)
- All `pkg/*` exports
- All public tool interfaces
- All display facades
- Error types in pkg/errors

### CAN CHANGE (Internal only)
- All `internal/*` types
- All command implementations in cmd/
- All factory implementations

### DEPRECATED (With migration path)
- To be determined in Phase 3

---

## Recommendations for Phase 2+

### Priority 1: Add Tests
- internal/data implementations
- internal/llm backends
- tools/* implementations

### Priority 2: Stabilize Interfaces
- Define clear contracts in internal/data
- Formalize provider abstraction in internal/llm
- Lock down tool interface contract

### Priority 3: Document Usage Patterns
- How to add new tools
- How to add new LLM providers
- How to use display system
- How to create custom formatters
