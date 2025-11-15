# ADR 0005: Google Search Tool Integration

**Status:** Accepted  
**Date:** 2025-11-15  
**Decision Makers:** Development Team  
**Technical Story:** Adding Google Search capability to adk-code agent

## Context and Problem Statement

The adk-code agent currently lacks the ability to search the web for real-time information, external documentation, or current events. Users cannot ask questions that require web search capabilities, limiting the agent's usefulness for tasks requiring up-to-date information or external knowledge.

The Google ADK Go framework (`google.golang.org/adk`) provides a built-in `geminitool.GoogleSearch` tool that integrates seamlessly with Gemini models to perform web searches. We need to integrate this capability into adk-code while following the established tool architecture patterns.

## Decision Drivers

* **User Need**: Enable the agent to answer questions requiring web search
* **Framework Support**: Google ADK provides native Google Search integration
* **Consistency**: Must follow existing tool patterns in the codebase
* **Model Compatibility**: Google Search tool only works with Gemini 2+ models
* **Minimal Changes**: Leverage existing ADK functionality rather than building custom implementation

## Considered Options

### Option 1: Wrap ADK's geminitool.GoogleSearch (Chosen)
Wrap the built-in `geminitool.GoogleSearch` from Google ADK Go within our tool system architecture.

**Pros:**
- Leverages official, maintained implementation
- Minimal code to write and maintain
- Full feature support from Google
- Automatic updates with ADK upgrades
- Native integration with Gemini models

**Cons:**
- Only works with Gemini models (not Ollama, OpenAI, etc.)
- Requires Gemini 2.0+ models
- Limited customization options
- HTML rendering requirements

### Option 2: Implement Custom Search API Integration
Build a custom tool using Google Custom Search API or similar service.

**Pros:**
- Model-agnostic (works with any LLM)
- Full control over implementation
- Can customize search parameters

**Cons:**
- Requires separate API key management
- Additional maintenance burden
- Need to implement rate limiting, error handling
- Reinvents existing ADK functionality
- May incur additional API costs

### Option 3: Use Third-Party Search Library
Integrate a third-party Go library for web search (e.g., DuckDuckGo API, Brave Search).

**Pros:**
- No Google dependency
- Potentially works with all models
- Alternative search engines available

**Cons:**
- Quality may vary compared to Google Search
- Additional dependencies
- May require API keys
- Less integrated with ADK framework

## Decision Outcome

**Chosen Option:** Option 1 - Wrap ADK's geminitool.GoogleSearch

This option provides the best balance of functionality, maintainability, and integration quality. While it's limited to Gemini models, this aligns with the project's primary focus on Google ADK and Gemini models.

### Implementation Strategy

Follow the established tool architecture pattern in adk-code:

```
adk-code/tools/
├── websearch/          # New package for web search tools
│   ├── init.go        # Auto-registration
│   ├── google_search.go  # Implementation
│   └── google_search_test.go  # Unit tests
```

## Technical Details

### Architecture Integration

The Google Search tool follows the standard tool pattern:

1. **Input/Output Structs**: Define structured parameters and results
2. **Tool Constructor**: `NewGoogleSearchTool()` creates and registers the tool
3. **Auto-registration**: `init()` function ensures automatic discovery
4. **Category**: `CategorySearchDiscovery` for logical grouping
5. **Metadata**: Priority and usage hints for LLM guidance

### Code Structure

```go
package websearch

import (
    "google.golang.org/adk/tool"
    "google.golang.org/adk/tool/geminitool"
    common "adk-code/tools/base"
)

// GoogleSearchInput defines the input for Google Search
type GoogleSearchInput struct {
    Query string `json:"query" jsonschema:"Search query"`
}

// GoogleSearchOutput defines the output from Google Search
type GoogleSearchOutput struct {
    Results string `json:"results"`
    Success bool   `json:"success"`
    Error   string `json:"error,omitempty"`
}

// NewGoogleSearchTool creates a Google Search tool using ADK's built-in implementation
func NewGoogleSearchTool() (tool.Tool, error) {
    // Use ADK's native Google Search tool
    searchTool := geminitool.GoogleSearch{}
    
    // Register with metadata
    common.Register(common.ToolMetadata{
        Tool:      searchTool,
        Category:  common.CategorySearchDiscovery,
        Priority:  0,
        UsageHint: "Search the web for current information, documentation, or answers",
    })
    
    return searchTool, nil
}
```

### Registration Flow

1. Package `websearch/init.go` calls `NewGoogleSearchTool()` in `init()`
2. Tool is registered in the global registry with metadata
3. Tool automatically available when `tools` package is imported
4. Agent can use the tool with compatible models (Gemini 2+)

### Model Compatibility

The Google Search tool has specific requirements:
- **Supported Models**: Gemini 2.0 Flash, Gemini 2.0 Flash Thinking, and newer Gemini models
- **Unsupported Models**: Ollama, OpenAI, Gemini 1.5, and earlier versions
- **Behavior**: Tool will be registered but may fail gracefully with non-Gemini models

### Usage Example

Once integrated, users can ask questions like:

```
User: "What are the latest features in Go 1.24?"
Agent: [uses google_search tool to find current information]
Agent: "According to recent sources, Go 1.24 includes..."
```

The agent will automatically invoke the Google Search tool when web information is needed.

## References

### Google ADK Documentation
- **ADK Go Documentation**: https://google.github.io/adk-docs/get-started/go/
- **Built-in Tools Guide**: https://github.com/google/adk-docs/blob/main/docs/tools/built-in-tools.md
- **ADK Go GitHub**: https://github.com/google/adk-go

### Implementation Examples
- **LangDB Guide**: https://docs.langdb.ai/guides/building-agents/building-web-search-agent-with-google-adk
- **DeepWiki Tutorial**: https://deepwiki.com/google/adk-go/2.1-quick-start-tutorial

### Code References
```go
// From Google ADK Go source
import "google.golang.org/adk/tool/geminitool"

tools := []tool.Tool{
    geminitool.GoogleSearch{},  // Built-in Google Search
}
```

### Existing Tool Patterns in adk-code
- `tools/exec/terminal_tools.go`: Command execution tool pattern
- `tools/file/read_tool.go`: File operation tool pattern
- `tools/search/diff_tools.go`: Search/discovery tool pattern
- `tools/base/registry.go`: Registration system

## Consequences

### Positive
- ✅ Users can ask questions requiring web search
- ✅ Agent gains real-time information access
- ✅ Minimal code to maintain (leverages ADK)
- ✅ Follows established architecture patterns
- ✅ Official Google implementation ensures quality
- ✅ Automatic updates through ADK upgrades

### Negative
- ⚠️ Only works with Gemini 2+ models
- ⚠️ Requires Gemini API access
- ⚠️ Limited to Google Search results
- ⚠️ May return HTML that needs rendering
- ⚠️ Search quality depends on Google's algorithms

### Neutral
- ℹ️ Tool will be available but may show errors with non-Gemini models
- ℹ️ Users should be aware of model compatibility
- ℹ️ Future enhancements could add alternative search providers

## Validation and Testing

### Unit Tests
```go
func TestGoogleSearchTool(t *testing.T) {
    tool, err := NewGoogleSearchTool()
    assert.NoError(t, err)
    assert.NotNil(t, tool)
    
    // Verify registration
    registry := common.GetRegistry()
    tools := registry.GetByCategory(common.CategorySearchDiscovery)
    // Verify google_search is registered
}
```

### Integration Tests
1. Start agent with Gemini 2.0 model
2. Ask question requiring web search
3. Verify tool is invoked
4. Verify results are returned
5. Verify graceful degradation with non-Gemini models

### Manual Verification
```bash
# Build and test
cd adk-code
make build
make test

# Run agent with Google Search
export GOOGLE_API_KEY="your-key"
./bin/adk-code --model gemini-2.0-flash

# Test query
> What are the latest developments in Golang?
```

## Implementation Checklist

- [ ] Create `tools/websearch/` directory
- [ ] Implement `google_search.go` with Input/Output types
- [ ] Implement `NewGoogleSearchTool()` constructor
- [ ] Add `init.go` for auto-registration
- [ ] Write unit tests in `google_search_test.go`
- [ ] Update `tools/tools.go` to export new types
- [ ] Update `tools/registry.go` init() to trigger registration
- [ ] Add integration test with Gemini model
- [ ] Update README.md with tool documentation
- [ ] Update CHANGELOG.md with new feature
- [ ] Run `make check` for quality gates
- [ ] Manual testing with agent

## Future Enhancements

1. **Multi-Provider Support**: Add alternative search providers (DuckDuckGo, Brave)
2. **Search Customization**: Add parameters for search filters, date ranges, etc.
3. **Result Caching**: Cache recent search results to reduce API calls
4. **Fallback Strategy**: Automatically fall back to alternative providers if Google Search fails
5. **HTML Rendering**: Implement proper HTML rendering for search results
6. **Search History**: Track and display search history in session

## Notes

- The Google Search tool is a foundational capability that enables many advanced use cases
- Tool should fail gracefully with helpful error messages for incompatible models
- Documentation should clearly state Gemini 2+ requirement
- Consider adding a configuration flag to enable/disable web search
- Monitor API usage and costs if using Gemini API

## Related Decisions

- ADR 0001: Claude Code Agent Support (established tool architecture pattern)
- ADR 0002: Ollama Dynamic Model Discovery (model compatibility considerations)
- Future ADR: Multi-Model Tool Compatibility Strategy
