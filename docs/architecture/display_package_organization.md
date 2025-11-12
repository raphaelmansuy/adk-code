# Display Package Organization

**Updated**: November 12, 2025  
**Phase**: 2 - Refactoring (Foundation Complete)

## Current Structure

The `display` package provides rich terminal display functionality for the code agent. It is partially organized into subpackages with a facade pattern for backward compatibility.

### Package Organization Diagram

```
display/                          # Main package (public facade)
├── facade.go                     # Public API re-exports
├── factory.go                    # Component factory
├── factory_test.go               # Tests
├── renderer.go                   # Renderer facade & re-exports
├── renderer_test.go              # Tests
│
├── CORE COMPONENTS (Root level)
│   ├── ansi.go                   # Terminal utilities (delegates to terminal/)
│   ├── spinner.go                # Spinner component
│   ├── spinner_test.go           # Tests
│   ├── typewriter.go             # Typewriter effect
│   ├── typewriter_test.go        # Tests
│   └── paginator.go              # Pagination logic
│
├── STREAMING DISPLAY (Root level)
│   ├── streaming_display.go      # Streaming message manager
│   ├── streaming_segment.go      # Message segments
│   ├── deduplicator.go           # Duplicate message prevention
│
├── EVENT HANDLING (Root level)
│   ├── event.go                  # Event types & timeline
│   ├── tool_adapter.go           # Tool execution listener
│   ├── tool_adapter_test.go      # Tests
│   ├── tool_renderer.go          # Tool call rendering
│   ├── tool_renderer_internals.go
│   ├── tool_result_parser.go     # Parse tool results
│   ├── tool_result_parser_test.go# Tests
│
├── banner/                       # Sub-package: Banner rendering
│   ├── banner.go                 # Banner component
│   └── banner_test.go            # Tests (root level)
│
├── components/                   # Sub-package: UI Components
│   ├── banner.go                 # Component banner (different from banner/)
│   ├── timeline.go               # Event timeline
│
├── styles/                       # Sub-package: Terminal styling
│   ├── colors.go                 # Color codes
│   ├── formatting.go             # Text formatting
│
├── terminal/                     # Sub-package: Terminal utilities
│   └── terminal.go               # TTY detection, cursor control
│
├── renderer/                     # Sub-package: Content renderers
│   ├── renderer.go               # Renderer interface & implementation
│   ├── markdown_renderer.go      # Markdown rendering
│
├── formatters/                   # Sub-package: Custom formatters
│   ├── registry.go               # Formatter registry
│   ├── registry_test.go          # Tests
│   ├── agent_formatter.go        # Agent message formatting
│   ├── error_formatter.go        # Error formatting
│   ├── tool_formatter.go         # Tool output formatting
│   └── metrics_formatter.go      # Metrics formatting
│
└── tooling/                      # Sub-package: (currently empty)
```

## Logical Grouping

### By Concern

**Terminal Primitives** (foundations)
- `terminal/` - TTY detection, cursor control
- `styles/` - Colors and text styling
- `ansi.go` - ANSI utilities (facade over terminal/)
- `formatters/` - Output formatting

**UI Components** (reusable pieces)
- `spinner.go` - Loading indicator
- `typewriter.go` - Text animation
- `banner/` + `components/banner.go` - Banners
- `paginator.go` - Pagination control

**Streaming Display** (real-time message display)
- `streaming_display.go` - Manages streaming output
- `streaming_segment.go` - Individual message segments
- `deduplicator.go` - Prevents duplicate rendering

**Content Rendering** (format-aware output)
- `renderer/` - Main renderer interface
- `renderer/markdown_renderer.go` - Markdown support

**Event Display** (agent interaction rendering)
- `event.go` - Event types and timeline
- `tool_adapter.go` - Tool execution listener
- `tool_renderer.go` - Render tool calls
- `tool_result_parser.go` - Parse tool results

**Factory** (component assembly)
- `factory.go` - Creates all components together
- `facade.go` - Public API re-exports

## Dependencies

### External Dependencies
- `google.golang.org/genai` - Used by streaming_display
- Terminal rendering libraries (charmbracelet, etc.)

### Internal Dependencies
```
facade.go
  ├─→ renderer/ (Renderer, MarkdownRenderer)
  └─→ banner/ (BannerRenderer)

factory.go
  ├─→ Renderer
  ├─→ BannerRenderer
  ├─→ TypewriterPrinter
  └─→ StreamingDisplay

StreamingDisplay
  ├─→ Renderer
  ├─→ TypewriterPrinter
  ├─→ MessageDeduplicator
  └─→ StreamingSegment

StreamingSegment
  ├─→ MarkdownRenderer
  ├─→ TypewriterPrinter
  ├─→ Event types
  └─→ Output format constants

ToolRenderer
  ├─→ Renderer
  └─→ Formatters

renderer/
  ├─→ components/ (EventTimeline, EventType)
  ├─→ formatters/
  └─→ styles/
```

## Stability Classification

### STABLE (Safe to depend on)
- `facade.go` - Public API
- `factory.go` - Component factory
- `renderer/` - Core rendering
- `styles/` - Color/style constants
- `terminal/` - Terminal utilities

### IN-TRANSITION (May change in refactoring)
- `spinner.go`, `typewriter.go` - May move to components/
- `streaming_*` - May consolidate to subpackage
- `tool_adapter.go`, `tool_renderer.go` - May consolidate

### IMPLEMENTATION DETAIL
- `formatters/` - Internal use primarily
- `components/` - Partial organization (incomplete)

## Design Patterns

### Facade Pattern
- `facade.go` re-exports all public types and constructors
- Allows internal reorganization without breaking imports
- All external imports should go through `display/` package root

### Factory Pattern
- `factory.go` bundles component creation
- `NewComponents()` creates coordinated set of components
- Simplifies application initialization

### Repository Pattern
- `formatters/registry.go` - Registry for custom formatters
- Allows extensibility for custom output formats

## Test Coverage

**Current**: 11.8% statement coverage

**By File**:
- `spinner_test.go` ✓ - Some tests exist
- `typewriter_test.go` ✓ - Some tests exist
- `tool_adapter_test.go` ✓ - Some tests exist
- `tool_result_parser_test.go` ✓ - Some tests exist
- Most others ✗ - No dedicated tests

## Recommendations for Phase 2+

### Short Term (Phase 2)
1. **Maintain current structure** - Too many dependencies to safely reorganize
2. **Improve facade** - Ensure all public APIs are re-exported (DONE)
3. **Add tests** - Focus on improving coverage to 50%+
4. **Document organization** - Make groupings clear (THIS DOCUMENT)

### Medium Term (Phase 3)
1. **Consolidate streaming** - Move streaming_* to streaming/ subpackage (after tests)
2. **Consolidate events** - Move event-related files to events/ subpackage
3. **Consolidate components** - Move spinner, typewriter to components/
4. **Update imports** - Use facade exclusively from outside display/

### Long Term (Phase 4+)
1. **Evaluate formatters** - May move to formatters/ subpackage
2. **Consider tool rendering** - May consolidate with tools/display
3. **Performance optimization** - Profile and optimize hot paths

## Critical Notes

### Circular Dependencies Risk ⚠️
- Many files in `display/` root are interdependent
- Moving them to subpackages risks circular imports
- Recommend moving only when dependencies can be clearly separated
- Use `facade.go` to manage visibility

### Test Coverage Gap 🔴
- Current 11.8% is very low
- Priority: Add tests before major refactoring
- Tests will reveal dependency issues naturally

### Import Strategy
```go
// ✓ GOOD: Import from main package
import "code_agent/display"

// ✓ OK: For formatter registry (special case)
import "code_agent/display/formatters"

// ✗ AVOID: Importing from subpackages directly
import "code_agent/display/styles"  // Use display.OutputFormatRich instead

// ✗ AVOID: Will cause issues
import "code_agent/display/renderer"  // Use display.NewRenderer instead
```

## How to Reorganize Safely

When moving files to subpackages:

1. **Identify dependencies** - List all imports and exports
2. **Create subpackage** - mkdir and create package
3. **Copy file** - Don't move, copy first
4. **Update package declaration** - Change package name
5. **Fix imports** - Update import paths
6. **Add re-exports** - Update facade.go
7. **Run tests** - Verify no breakage
8. **Delete original** - Only after tests pass
9. **Update external imports** - If any imports subpackage directly

## References

- `docs/architecture/dependency_graph.md` - Overall project dependencies
- `docs/architecture/api_surface.md` - Public API definitions
- Tests in `*_test.go` files throughout display/
