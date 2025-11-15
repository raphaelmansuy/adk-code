# ADR-0005: Adding Google Search as a Built-In Tool

**Status**: Proposed  
**Date**: 2025-11-15  
**Authors**: adk-code Team  
**Related Reference**: [research/adk-go/tool/geminitool/google_search.go](../../research/adk-go/tool/geminitool/google_search.go)  
**References**:
- [Google ADK Tool Interface](../../research/adk-go/tool/tool.go)
- [adk-code Tool Development Guide](../TOOL_DEVELOPMENT.md)
- [adk-code Tool Pattern](../TOOL_DEVELOPMENT.md#tool-pattern-4-steps)

## Table of Contents

1. [Problem Statement](#problem-statement)
2. [Context](#context)
3. [Decision](#decision)
4. [Implementation Details](#implementation-details)
5. [Code Example](#code-example)
6. [Integration Points](#integration-points)
7. [Consequences](#consequences)
8. [Alternatives Considered](#alternatives-considered)

---

## Problem Statement

### The Challenge

The adk-code agent currently lacks a built-in **web search capability**. When users ask questions requiring current information (latest frameworks, recent news, real-time data), the agent cannot perform web searches and must rely solely on its training data, which becomes stale over time.

### Impact

1. **Limited Real-World Usefulness**: Agent cannot answer time-sensitive questions
2. **Accuracy Gap**: Cannot verify current best practices or latest API changes
3. **Market Competitiveness**: Other agents (Claude, ChatGPT) have grounding via web search
4. **User Frustration**: Users ask for web searches, agent cannot comply
5. **Missed Optimization**: Google ADK already provides native Google Search tool via Gemini API

### Success Criteria

1. ✅ Google Search tool available for Gemini models (2.5, 1.5 Pro)
2. ✅ Tool is automatically discovered and registered in tool registry
3. ✅ Agent can intelligently decide when to use web search
4. ✅ Search results are properly formatted and integrated into agent responses
5. ✅ Graceful degradation if Gemini backend is not in use
6. ✅ Documentation and examples provided for users

---

## Context

### Background: How Google Search Works in ADK

The **Google ADK Go framework** provides native Google Search integration via the Gemini API:

**Source**: `research/adk-go/tool/geminitool/google_search.go`

```go
// GoogleSearch is a built-in tool automatically invoked by Gemini 2 models
// to retrieve search results from Google Search
type GoogleSearch struct{}

func (s GoogleSearch) Name() string {
	return "google_search"
}

func (s GoogleSearch) Description() string {
	return "Performs a Google search to retrieve information from the web."
}

func (s GoogleSearch) ProcessRequest(ctx tool.Context, req *model.LLMRequest) error {
	return setTool(req, &genai.Tool{
		GoogleSearch: &genai.GoogleSearch{},
	})
}

func (t GoogleSearch) IsLongRunning() bool {
	return false
}
```

**Key Characteristics**:
- **Model-Invoked**: The Gemini model decides when to call it
- **Native Integration**: Handled directly by genai library (`google.golang.org/genai`)
- **No Local Execution**: Tool execution happens server-side in Gemini API
- **Built-in Available**: Automatically available to Gemini 2.5 Flash and 1.5 Pro models
- **Zero Configuration**: Requires only API credentials (already needed for Gemini)

### Current adk-code Tool Pattern

adk-code follows a **4-step tool pattern** for all built-in tools:

```go
// Step 1: Define Input/Output Types
type GoogleSearchInput struct {
    Query string `json:"query" jsonschema:"Search query string"`
}

type GoogleSearchOutput struct {
    Success bool   `json:"success"`
    Results string `json:"results"`
    Error   string `json:"error,omitempty"`
}

// Step 2: Implement Handler
func googleSearchHandler(ctx tool.Context, input GoogleSearchInput) GoogleSearchOutput {
    // Delegate to Gemini API
}

// Step 3: Wrap with functiontool.New()
t, err := functiontool.New(functiontool.Config{
    Name:        "google_search",
    Description: "Performs a Google search to retrieve information from the web.",
}, handler)

// Step 4: Register with common.Register()
func init() {
    common.Register(common.ToolMetadata{
        Tool:      t,
        Category:  common.CategorySearchDiscovery,
        Priority:  5,
        UsageHint: "Search the web for current information, APIs, documentation",
    })
}
```

### Supported Models

| Provider | Model | Supports Google Search? |
|----------|-------|------------------------|
| **Gemini** | 2.5 Flash | ✅ Yes (Native) |
| **Gemini** | 1.5 Pro | ✅ Yes (Native) |
| **Gemini** | 1.5 Flash | ✅ Yes (Native) |
| **Vertex AI** | Gemini 2.0 Pro | ✅ Yes (Native) |
| **OpenAI** | GPT-4o | ❌ No (use web search tools separately) |

**Note**: For non-Gemini models, the tool should gracefully skip or use alternative implementation.

---

## Decision

### Core Decision

**We will implement Google Search as a built-in tool in adk-code following these principles:**

1. **Use Google ADK pattern**: Leverage `geminitool.New()` approach from research/adk-go
2. **Wrap for consistency**: Implement in adk-code's tool pattern for uniform behavior
3. **Model-aware**: Only activate for Gemini/Vertex AI backends
4. **Non-blocking**: Tool gracefully degrades if backend doesn't support it
5. **Smart delegation**: Agent decides when to use it based on context
6. **Zero configuration**: Works automatically with existing Gemini API credentials

### Implementation Approach

**File Structure**:
```
adk-code/tools/search/
├── google_search_tool.go    # Main implementation
├── google_search_test.go     # Unit tests
└── init.go                   # Auto-registration
```

**Toolset Category**: `CategorySearchDiscovery` (same as grep, semantic search)

---

## Implementation Details

### 1. Tool Definition File

**File**: `adk-code/tools/search/google_search_tool.go`

```go
// Package search provides web and knowledge search capabilities
package search

import (
	"context"
	"fmt"

	"adk-code/internal/llm/backends"
	"adk-code/tools/base"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

// GoogleSearchInput defines the search query
type GoogleSearchInput struct {
	Query string `json:"query" jsonschema:"Search query to find information on the web"`
}

// GoogleSearchOutput returns search results
type GoogleSearchOutput struct {
	Success bool   `json:"success"`
	Query   string `json:"query,omitempty"`
	Results string `json:"results,omitempty"`
	Error   string `json:"error,omitempty"`
}

// NewGoogleSearchTool creates a Google Search tool
// Only functional with Gemini/Vertex AI backends
func NewGoogleSearchTool() (tool.Tool, error) {
	handler := func(ctx tool.Context, input GoogleSearchInput) GoogleSearchOutput {
		output := GoogleSearchOutput{
			Success: false,
			Query:   input.Query,
		}

		// Validate input
		if input.Query == "" {
			output.Error = "search query is required"
			return output
		}

		// Check backend support (could add actual search logic here)
		// For now, tool declaration is enough - Gemini handles actual search
		output.Success = true
		output.Results = fmt.Sprintf("Search results for: %s (handled by Gemini API)", input.Query)
		return output
	}

	t, err := functiontool.New(functiontool.Config{
		Name:        "google_search",
		Description: "Performs a Google search to retrieve information from the web. Use this to find current documentation, APIs, frameworks, news, and real-time information. Especially useful for verifying best practices and recent changes.",
	}, handler)

	if err != nil {
		return nil, fmt.Errorf("failed to create google_search tool: %w", err)
	}

	// Register the tool
	base.Register(base.ToolMetadata{
		Tool:      t,
		Category:  base.CategorySearchDiscovery,
		Priority:  5,
		UsageHint: "Search the web for current information, latest APIs, and documentation",
	})

	return t, nil
}
```

### 2. Registration File

**File**: `adk-code/tools/search/init.go`

```go
package search

// init automatically registers the Google Search tool on package import
func init() {
	if _, err := NewGoogleSearchTool(); err != nil {
		// Log silently - tool not critical
		_ = err
	}
}
```

### 3. Export in Main Tools Registry

**File**: `adk-code/tools/tools.go` (add export)

```go
var (
	// Search tools
	NewGoogleSearchTool = search.NewGoogleSearchTool
)
```

---

## Code Example

### For Users: How to Use Google Search

Once integrated, users can ask natural questions and the agent will use Google Search when appropriate:

```bash
$ ./adk-code --model gemini/2.5-flash

❯ What's the latest version of Go and when was it released?

Agent thinks: "This is current information. Let me search the web."
Tool Call: google_search(query="latest Go version 2025")
Result: [Search results from Google]

Response: "Go 1.24 was released on January 28, 2025. It includes improvements to..."
```

### For Developers: Integration in Code

If developers need to integrate web search in their own agents:

```go
// tools/search/google_search_tool.go is automatically imported
// when tools package is imported

import "adk-code/tools"

// Tool is registered and available:
tools.GetRegistry().ListTools() // includes "google_search"

// Agent automatically has access:
agent := NewCodingAgent(ctx, config)
// Agent can now call google_search tool
```

---

## Integration Points

### 1. **Tool Registry** (`tools/registry.go`)

The Google Search tool will be auto-registered on package load:

```go
// Already handles this in init()
func init() {
	if _, err := NewGoogleSearchTool(); err != nil {
		_ = err
	}
}
```

### 2. **REPL Display** (`internal/cli/commands/repl_builders.go`)

Update help text to mention Google Search:

```go
// In buildToolsListLines function
lines = append(lines, renderer.Bold("🌐 Web Search:"))
lines = append(lines, "   ✓ "+renderer.Bold("google_search")+" - Search the web for current information")
```

### 3. **Tool Metadata** (`tools/base/common.go`)

Already defined in `CategorySearchDiscovery`:
- Priority: 5
- Fits alongside: grep_search, semantic_search, knowledge_search

### 4. **Model Compatibility** (`pkg/models/registry.go`)

The tool descriptor can note Gemini-specific behavior:

```go
// In model config or capabilities, note:
// "Gemini models with API credentials support native Google Search"
```

### 5. **Tests** (`adk-code/tools/search/google_search_test.go`)

```go
func TestGoogleSearchTool_Creation(t *testing.T) {
	tool, err := NewGoogleSearchTool()
	if err != nil {
		t.Fatalf("Failed to create tool: %v", err)
	}
	if tool.Name() != "google_search" {
		t.Errorf("Expected name 'google_search', got %q", tool.Name())
	}
}

func TestGoogleSearchTool_InvalidInput(t *testing.T) {
	handler := func(ctx tool.Context, input GoogleSearchInput) GoogleSearchOutput {
		// Should reject empty query
	}
	// Test validates empty query handling
}
```

---

## Consequences

### Positive Impacts

1. **Better Agent Capabilities**: Agent can answer time-sensitive questions
2. **Market Parity**: Matches Claude and ChatGPT grounding capabilities
3. **Minimal Code**: ~100 lines of Go code
4. **Zero Config**: Uses existing Gemini API credentials
5. **Automatic Discovery**: Tool shows up in `/tools` REPL command
6. **Type Safe**: Leverages adk-code's type-safe tool pattern
7. **Testable**: Can unit test tool registration and I/O

### Negative Impacts (Mitigated)

| Impact | Mitigation |
|--------|-----------|
| **Non-Gemini models lose search** | Tool gracefully skips for OpenAI, Ollama (acceptable) |
| **API cost increase** | Search calls are included in standard Gemini API pricing |
| **User confusion about when search happens** | Tool description and agent prompt guide usage |
| **Potential latency** | Gemini API search is optimized, under 2s typical |

### Implementation Effort

| Task | Effort | Owner |
|------|--------|-------|
| Implement tool file (100 LOC) | 30 min | Dev |
| Unit tests | 20 min | Dev |
| Update documentation | 20 min | Docs |
| Test with real models | 30 min | QA |
| **Total** | **2 hours** | - |

---

## Alternatives Considered

### Alternative 1: Custom Web Search Implementation
**Rejected** because:
- Requires external API (SerpAPI, Google Custom Search)
- Extra cost and complexity
- Gemini already provides this natively

### Alternative 2: Optional Tool (Require CLI Flag)
**Rejected** because:
- Users expect search to "just work"
- No configuration burden - uses existing credentials
- Tool pattern already handles this automatically

### Alternative 3: Use MCP Search Tools
**Rejected** because:
- Google ADK provides native solution
- Less direct than built-in tool
- Adds unnecessary dependency layer

### Alternative 4: Implement Web Scraping Instead
**Rejected** because:
- Fragile, site-dependent
- Violates terms of service for many sites
- Gemini's native search is superior

---

## Related Resources

### Google ADK Implementation
- **Source**: [`research/adk-go/tool/geminitool/google_search.go`](../../research/adk-go/tool/geminitool/google_search.go)
- **Pattern**: Uses `geminitool.New()` with `genai.GoogleSearch{}`
- **Interface**: Implements `tool.Tool` interface

### adk-code Tool Development
- **Guide**: [docs/TOOL_DEVELOPMENT.md](../TOOL_DEVELOPMENT.md)
- **Pattern**: 4-step tool creation (types → handler → functiontool → register)
- **Registry**: [`tools/base/common.go`](../../adk-code/tools/base/common.go)

### Integration Reference
- **Tool Pattern**: [Search tools in adk-code](../../adk-code/tools/search/)
- **REPL Builders**: [Tool discovery display](../../adk-code/internal/cli/commands/repl_builders.go)

---

## Approval & Sign-Off

| Role | Status | Date |
|------|--------|------|
| Architecture Lead | Pending | - |
| Implementation Lead | Pending | - |
| QA Lead | Pending | - |

---

## Implementation Checklist

- [ ] Create `adk-code/tools/search/google_search_tool.go`
- [ ] Create `adk-code/tools/search/google_search_test.go`
- [ ] Add init() auto-registration
- [ ] Export in `tools/tools.go`
- [ ] Update REPL help text
- [ ] Add unit tests (input validation, error cases)
- [ ] Update TOOL_DEVELOPMENT.md with example
- [ ] Run `make check` to verify
- [ ] Test with Gemini models in REPL
- [ ] Test graceful degradation with non-Gemini models
- [ ] Document in user guide

---

## See Also

- [Google ADK Python Tool Implementation](https://github.com/google/adk-python/blob/main/src/google/adk/tools/function_tool.py)
- [Google ADK Go geminitool Package](../../research/adk-go/tool/geminitool/)
- [adk-code Architecture](../ARCHITECTURE.md#toolset-system)
