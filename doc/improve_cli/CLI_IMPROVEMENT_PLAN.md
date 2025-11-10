# CLI Display Improvement Plan: code_agent vs Cline

**Author:** AI Analysis  
**Date:** November 10, 2025  
**Status:** Planning Phase  

## Executive Summary

This document provides a comprehensive analysis of the code_agent CLI display system compared to Cline's implementation, and outlines a detailed improvement plan to achieve superior display quality and user experience.

**Current State:** code_agent has a basic CLI with hardcoded ANSI colors and minimal formatting.  
**Target State:** A professional, modular CLI display system that matches or exceeds Cline's capabilities.  
**Estimated Effort:** 3-4 weeks (1 developer)

---

## Table of Contents

1. [Gap Analysis](#gap-analysis)
2. [Architecture Comparison](#architecture-comparison)
3. [Key Improvements Required](#key-improvements-required)
4. [Implementation Plan](#implementation-plan)
5. [Technical Specifications](#technical-specifications)
6. [Dependencies](#dependencies)
7. [Success Metrics](#success-metrics)
8. [Risk Assessment](#risk-assessment)

---

## Gap Analysis

### Current code_agent Implementation

**File Structure:**
```
code_agent/
  main.go           (382 lines - contains ALL display logic)
  agent/
    coding_agent.go
    enhanced_prompt.go
```

**Display Capabilities:**
- ✗ Hardcoded ANSI color codes
- ✗ No markdown rendering
- ✗ Basic text output only
- ✗ Simple event printing
- ✗ No streaming effects
- ✗ Minimal tool call formatting
- ✗ Static ASCII banner
- ✗ No output format options
- ✗ No TTY detection
- ✓ Basic color differentiation

**Code Quality Issues:**
1. All display logic in main.go (poor separation of concerns)
2. No abstraction layers
3. Hardcoded formatting strings
4. Limited extensibility
5. No configuration options

### Cline Implementation

**File Structure:**
```
cline/cli/pkg/cli/display/
  ansi.go                    - Terminal utilities
  banner.go                  - Session banners
  deduplicator.go            - Message deduplication
  markdown_renderer.go       - Rich markdown rendering
  renderer.go                - Core rendering abstraction
  segment_streamer.go        - Streaming segment management
  streaming.go               - Streaming display manager
  system_renderer.go         - System message rendering
  tool_renderer.go           - Unified tool display
  tool_result_parser.go      - Tool output parsing
  typewriter.go              - Animated text output
```

**Display Capabilities:**
- ✓ Modular architecture with clear separation
- ✓ Markdown rendering with syntax highlighting
- ✓ Multiple output formats (rich/json/plain)
- ✓ TTY detection and adaptive rendering
- ✓ Streaming display with segments
- ✓ Typewriter effect (optional)
- ✓ Rich tool rendering with previews
- ✓ Token usage and cost display
- ✓ Session banners with context
- ✓ Message deduplication
- ✓ Diff rendering with colors
- ✓ Code block syntax highlighting
- ✓ Contextual headers and icons

**Key Libraries:**
- `github.com/charmbracelet/lipgloss` - Styling and layout
- `github.com/charmbracelet/glamour` - Markdown rendering
- `github.com/charmbracelet/huh` - Forms and prompts
- `github.com/charmbracelet/bubbles` - TUI components
- `github.com/spf13/cobra` - CLI framework
- `golang.org/x/term` - Terminal capabilities

---

## Architecture Comparison

### Current code_agent Architecture

```
┌─────────────────────────────────────┐
│           main.go                   │
│  ┌───────────────────────────────┐  │
│  │  printBanner()                │  │
│  │  printEvent()                 │  │
│  │  ANSI color constants         │  │
│  │  Event loop                   │  │
│  │  Direct fmt.Printf calls      │  │
│  └───────────────────────────────┘  │
└─────────────────────────────────────┘
```

**Issues:**
- Monolithic design
- No abstraction
- Hard to test
- Limited extensibility
- No format flexibility

### Proposed code_agent Architecture (Cline-inspired)

```
┌─────────────────────────────────────────────────────┐
│                    main.go                          │
│  ┌─────────────────────────────────────────────┐   │
│  │  CLI Setup (cobra/flags)                    │   │
│  │  Event Loop                                 │   │
│  │  Call display.Renderer methods              │   │
│  └─────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────┐
│            display/ package                         │
│  ┌─────────────────────────────────────────────┐   │
│  │  Renderer (main facade)                     │   │
│  │    - RenderBanner()                         │   │
│  │    - RenderEvent()                          │   │
│  │    - RenderTool()                           │   │
│  │    - RenderAPI()                            │   │
│  └─────────────────────────────────────────────┘   │
│          │         │         │         │            │
│          ▼         ▼         ▼         ▼            │
│  ┌────────┐ ┌──────────┐ ┌────────┐ ┌──────────┐  │
│  │ Banner │ │ Markdown │ │  Tool  │ │Typewriter│  │
│  │Renderer│ │ Renderer │ │Renderer│ │ Printer  │  │
│  └────────┘ └──────────┘ └────────┘ └──────────┘  │
│                                                     │
│  ┌─────────────────────────────────────────────┐   │
│  │  StreamingDisplay (real-time updates)       │   │
│  │    - HandlePartialMessage()                 │   │
│  │    - HandleCompleteMessage()                │   │
│  └─────────────────────────────────────────────┘   │
│                                                     │
│  ┌─────────────────────────────────────────────┐   │
│  │  Utilities                                  │   │
│  │    - TTY detection                          │   │
│  │    - ANSI helpers                           │   │
│  │    - Message deduplication                  │   │
│  └─────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────┘
```

**Benefits:**
- Clean separation of concerns
- Easy to test each component
- Extensible architecture
- Support for multiple output formats
- Professional code organization

---

## Key Improvements Required

### 1. Display Package Architecture ⭐⭐⭐ (Critical)

**Goal:** Create a modular display system with clear abstractions.

**Components to Build:**

#### 1.1 Core Renderer (`display/renderer.go`)
```go
type Renderer struct {
    typewriter   *TypewriterPrinter
    mdRenderer   *MarkdownRenderer
    toolRenderer *ToolRenderer
    outputFormat string
    
    // Styles (lipgloss)
    dimStyle     lipgloss.Style
    greenStyle   lipgloss.Style
    redStyle     lipgloss.Style
    // ... more styles
}

func NewRenderer(outputFormat string) *Renderer
func (r *Renderer) RenderEvent(event *session.Event) error
func (r *Renderer) RenderAPI(status string, info *APIInfo) error
func (r *Renderer) RenderBanner(info BannerInfo) string
```

#### 1.2 Markdown Renderer (`display/markdown_renderer.go`)
```go
type MarkdownRenderer struct {
    renderer *glamour.TermRenderer
    width    int
}

func NewMarkdownRenderer() (*MarkdownRenderer, error)
func (mr *MarkdownRenderer) Render(markdown string) (string, error)
```

#### 1.3 Tool Renderer (`display/tool_renderer.go`)
```go
type ToolRenderer struct {
    mdRenderer   *MarkdownRenderer
    outputFormat string
}

func (tr *ToolRenderer) RenderToolExecution(tool *ToolInfo) string
func (tr *ToolRenderer) RenderToolPreview(tool *ToolInfo) string
func (tr *ToolRenderer) RenderCommandExecution(cmd string) string
```

#### 1.4 Typewriter Printer (`display/typewriter.go`)
```go
type TypewriterPrinter struct {
    config *TypewriterConfig
}

func NewTypewriterPrinter(config *TypewriterConfig) *TypewriterPrinter
func (tp *TypewriterPrinter) Print(text string)
func (tp *TypewriterPrinter) SetEnabled(enabled bool)
```

#### 1.5 Banner Renderer (`display/banner.go`)
```go
type BannerInfo struct {
    Version    string
    Provider   string
    ModelID    string
    Workdir    string
}

func RenderSessionBanner(info BannerInfo) string
```

#### 1.6 Streaming Display (`display/streaming.go`)
```go
type StreamingDisplay struct {
    state      *ConversationState
    renderer   *Renderer
    dedupe     *MessageDeduplicator
    activeSegment *StreamingSegment
}

func NewStreamingDisplay(state *ConversationState, renderer *Renderer) *StreamingDisplay
func (sd *StreamingDisplay) HandlePartialMessage(msg *Message) error
func (sd *StreamingDisplay) FreezeActiveSegment()
```

### 2. Rich Text Formatting ⭐⭐⭐ (Critical)

**Goal:** Render markdown content with proper formatting.

**Features:**
- Syntax-highlighted code blocks
- Proper heading rendering
- List formatting (bullet and numbered)
- Bold, italic, and other markdown styles
- Link rendering
- Block quotes
- Tables
- Diff rendering for file changes

**Implementation:**
- Use `glamour` library for markdown rendering
- Detect terminal capabilities
- Apply appropriate theme (dark/light)
- Configure word wrapping

### 3. Tool Display Enhancement ⭐⭐⭐ (Critical)

**Goal:** Provide clear, contextual tool execution display.

**Features:**
- **Approval Phase:** "Cline wants to read `file.go`"
- **Execution Phase:** "Cline is reading `file.go`"
- **Preview Content:** Show diffs, file previews
- **Result Display:** Format tool outputs nicely
- **Icons/Symbols:** Use Unicode symbols for visual clarity

**Tool Types to Handle:**
1. File operations (read, write, edit, delete)
2. Command execution
3. Directory listings
4. Search operations
5. Patch applications

### 4. Output Format Support ⭐⭐ (Important)

**Goal:** Support multiple output formats for different use cases.

**Formats:**
1. **Rich** (default): Full markdown, colors, styling
2. **Plain**: No colors, no markdown, plain text
3. **JSON**: Structured JSON output for programmatic use

**Implementation:**
```go
const (
    OutputFormatRich  = "rich"
    OutputFormatPlain = "plain"
    OutputFormatJSON  = "json"
)

type OutputConfig struct {
    Format        string
    EnableColors  bool
    EnableMarkdown bool
    Width         int
}
```

### 5. TTY Detection & Adaptation ⭐⭐ (Important)

**Goal:** Adapt output based on terminal capabilities.

**Checks:**
- Is stdout a terminal?
- Is output being piped?
- Is in CI environment?
- Terminal width and height
- Color support (8-bit, 256-color, true color)

**Implementation:**
```go
import "golang.org/x/term"

func isTTY() bool {
    return term.IsTerminal(int(os.Stdout.Fd()))
}

func getTerminalWidth() int {
    width, _, err := term.GetSize(int(os.Stdout.Fd()))
    if err != nil {
        return 80 // fallback
    }
    return width
}
```

### 6. Session Context Banner ⭐⭐ (Important)

**Goal:** Display session information at start.

**Information to Show:**
- Agent name and version
- Model provider and ID
- Working directory
- Session ID (optional)
- Available commands

**Design:**
```
╭────────────────────────────────────────╮
│                                        │
│  code_agent v1.0.0                     │
│  google/gemini-2.5-flash               │
│  ~/projects/myapp                      │
│                                        │
╰────────────────────────────────────────╯
```

### 7. Streaming & Animation ⭐ (Nice-to-have)

**Goal:** Provide engaging real-time feedback.

**Features:**
- **Typewriter effect:** Character-by-character text display
- **Segment streaming:** Show headers immediately, content when ready
- **Progress indicators:** Spinners for long operations
- **Status updates:** "Thinking...", "Reading file...", etc.

**Configuration:**
```go
type TypewriterConfig struct {
    BaseDelay    time.Duration
    FastDelay    time.Duration
    SlowDelay    time.Duration
    PauseDelay   time.Duration
    Enabled      bool
    RandomFactor float64
}
```

### 8. Token Usage & Cost Display ⭐ (Nice-to-have)

**Goal:** Show API usage information clearly.

**Display Format:**
```
API Request: ↑ 2.3k ↓ 856 → 1.5k ← 0 $0.0023
```

**Components:**
- `↑` Input tokens
- `↓` Output tokens
- `→` Cache reads
- `←` Cache writes
- `$` Cost

**Abbreviations:**
- 1,234 → 1.2k
- 1,234,567 → 1.2m

### 9. Error Display Enhancement ⭐ (Nice-to-have)

**Goal:** Make errors clear and actionable.

**Features:**
- Red color for errors
- Clear error messages
- Suggestions for fixes
- Stack traces (in verbose mode)
- Error categories

**Design:**
```
✗ Error: Failed to read file

  File not found: demo/calculate.c

  Suggestion: Check the file path and ensure it exists.
  Try: ls demo/ to see available files.
```

### 10. Message Deduplication ⭐ (Nice-to-have)

**Goal:** Avoid showing duplicate messages.

**Implementation:**
```go
type MessageDeduplicator struct {
    mu           sync.RWMutex
    seenMessages map[string]time.Time
    ttl          time.Duration
}

func (md *MessageDeduplicator) IsDuplicate(msg *Message) bool
```

---

## Implementation Plan

### Phase 1: Foundation (Week 1)

**Goal:** Set up the basic architecture and dependencies.

#### Tasks:

1. **Create display package structure** (4 hours)
   - Create `code_agent/display/` directory
   - Set up initial file structure
   - Define package interfaces

2. **Add dependencies** (2 hours)
   ```bash
   go get github.com/charmbracelet/lipgloss
   go get github.com/charmbracelet/glamour
   go get github.com/spf13/cobra
   go get golang.org/x/term
   ```

3. **Implement core Renderer** (8 hours)
   - Create `renderer.go` with base structure
   - Implement constructor and basic methods
   - Add lipgloss style definitions
   - Implement TTY detection

4. **Implement MarkdownRenderer** (6 hours)
   - Create `markdown_renderer.go`
   - Integrate glamour library
   - Add theme detection
   - Test with sample markdown

5. **Refactor main.go** (4 hours)
   - Extract display logic
   - Use new Renderer
   - Clean up printEvent function
   - Add output format flag

**Deliverables:**
- Working display package structure
- Basic renderer with markdown support
- Refactored main.go using new display system
- Tests for core components

### Phase 2: Rich Display Features (Week 2)

**Goal:** Implement rich text formatting and tool display.

#### Tasks:

1. **Implement ToolRenderer** (8 hours)
   - Create `tool_renderer.go`
   - Add contextual headers
   - Implement content preview
   - Add tool-specific formatting

2. **Implement Banner system** (4 hours)
   - Create `banner.go`
   - Design session banner
   - Add version and model info
   - Implement working directory display

3. **Enhance Event Rendering** (6 hours)
   - Improve event type detection
   - Add rich formatting per event type
   - Implement diff rendering
   - Add code block highlighting

4. **Add output format support** (6 hours)
   - Implement plain format
   - Implement JSON format
   - Add format flag to CLI
   - Test all formats

**Deliverables:**
- Rich tool display
- Session banners
- Multiple output formats
- Enhanced event rendering

### Phase 3: Advanced Features (Week 3)

**Goal:** Add streaming, animation, and polish.

#### Tasks:

1. **Implement TypewriterPrinter** (6 hours)
   - Create `typewriter.go`
   - Implement character-by-character output
   - Add speed configuration
   - Add enable/disable flag

2. **Implement StreamingDisplay** (8 hours)
   - Create `streaming.go`
   - Add segment-based streaming
   - Implement message deduplication
   - Handle partial messages

3. **Add API usage display** (4 hours)
   - Implement token counting display
   - Add cost calculation
   - Format with abbreviations
   - Add to API response rendering

4. **Implement error display** (6 hours)
   - Create error formatting
   - Add suggestions
   - Implement severity levels
   - Add verbose mode

**Deliverables:**
- Typewriter effect
- Streaming display
- API usage tracking
- Enhanced error display

### Phase 4: Polish & Testing (Week 4)

**Goal:** Test, document, and refine.

#### Tasks:

1. **Comprehensive Testing** (8 hours)
   - Unit tests for all display components
   - Integration tests with agent
   - Manual testing in different terminals
   - Test with piped output

2. **Documentation** (6 hours)
   - Code documentation
   - User guide for display features
   - Examples and screenshots
   - Configuration reference

3. **Performance Optimization** (4 hours)
   - Profile rendering performance
   - Optimize markdown rendering
   - Reduce allocations
   - Benchmark improvements

4. **Final Polish** (6 hours)
   - Fix any bugs
   - Improve error messages
   - Refine styling
   - User feedback incorporation

**Deliverables:**
- Complete test coverage
- Documentation
- Performance benchmarks
- Production-ready display system

---

## Technical Specifications

### File Structure

```
code_agent/
├── main.go                 (Simplified, uses display package)
├── go.mod                  (Updated with new dependencies)
├── go.sum
├── display/
│   ├── ansi.go            (Terminal utilities)
│   ├── banner.go          (Session banners)
│   ├── deduplicator.go    (Message deduplication)
│   ├── markdown_renderer.go (Markdown rendering)
│   ├── renderer.go        (Core rendering facade)
│   ├── streaming.go       (Streaming display)
│   ├── tool_renderer.go   (Tool display)
│   ├── typewriter.go      (Animated output)
│   └── utils.go           (Shared utilities)
├── agent/
│   ├── coding_agent.go
│   └── enhanced_prompt.go
└── tools/
    └── ... (existing tools)
```

### Key Interfaces

```go
// Renderer is the main facade for all display operations
type Renderer interface {
    RenderBanner(info BannerInfo) error
    RenderEvent(event *session.Event) error
    RenderAPI(status string, info *APIInfo) error
    RenderError(err error) error
    RenderMessage(prefix, text string, newline bool) error
}

// MarkdownRenderer handles markdown formatting
type MarkdownRenderer interface {
    Render(markdown string) (string, error)
}

// ToolRenderer handles tool-specific display
type ToolRenderer interface {
    RenderToolExecution(tool *ToolInfo) string
    RenderToolPreview(tool *ToolInfo) string
}
```

### Configuration Options

```go
type DisplayConfig struct {
    // Output format: rich, plain, or json
    OutputFormat string
    
    // Enable/disable typewriter effect
    TypewriterEnabled bool
    
    // Typewriter speed multiplier
    TypewriterSpeed float64
    
    // Enable/disable colors
    ColorsEnabled bool
    
    // Enable/disable markdown rendering
    MarkdownEnabled bool
    
    // Terminal width (0 = auto-detect)
    Width int
    
    // Verbose mode
    Verbose bool
}
```

### CLI Flags

```bash
# Output format
--output-format, -o   Output format (rich|plain|json) [default: rich]

# Display options
--no-color            Disable colored output
--no-markdown         Disable markdown rendering
--no-typewriter       Disable typewriter effect
--typewriter-speed    Typewriter speed multiplier [default: 1.0]

# Existing flags
--verbose, -v         Verbose output
```

---

## Dependencies

### Required Go Modules

```go
require (
    // Charmbracelet ecosystem
    github.com/charmbracelet/lipgloss v1.0.0
    github.com/charmbracelet/glamour v0.10.0
    github.com/charmbracelet/huh v0.7.0        // For future input prompts
    
    // CLI framework
    github.com/spf13/cobra v1.8.0
    
    // Terminal utilities
    golang.org/x/term v0.18.0
    
    // Existing dependencies
    google.golang.org/adk/agent v0.x.x
    google.golang.org/adk/model/gemini v0.x.x
    google.golang.org/adk/runner v0.x.x
    google.golang.org/adk/session v0.x.x
    google.golang.org/genai v0.x.x
)
```

### Dependency Compatibility

All dependencies are:
- ✓ Pure Go (no C dependencies)
- ✓ Well-maintained
- ✓ Widely used (thousands of stars)
- ✓ MIT/Apache licensed
- ✓ Cross-platform (macOS, Linux, Windows)

---

## Success Metrics

### Functional Metrics

1. **Display Quality**
   - ✓ Markdown rendering works correctly
   - ✓ Syntax highlighting in code blocks
   - ✓ Proper diff rendering
   - ✓ Tool execution clearly visible
   - ✓ Errors are clear and actionable

2. **Format Support**
   - ✓ Rich format renders correctly
   - ✓ Plain format has no ANSI codes
   - ✓ JSON format is valid and complete
   - ✓ Piped output works correctly

3. **Compatibility**
   - ✓ Works in iTerm2
   - ✓ Works in Terminal.app
   - ✓ Works in VS Code terminal
   - ✓ Works over SSH
   - ✓ Works with piped output

### Performance Metrics

1. **Rendering Speed**
   - Target: < 50ms for typical event
   - Target: < 200ms for large markdown

2. **Memory Usage**
   - Target: < 10MB overhead for display system
   - No memory leaks in long-running sessions

3. **CPU Usage**
   - Minimal CPU when idle
   - Typewriter effect: < 5% CPU

### User Experience Metrics

1. **Readability**
   - Tool operations clearly visible
   - Thinking process easy to follow
   - Errors immediately noticeable

2. **Information Density**
   - All important info visible
   - Not cluttered
   - Appropriate whitespace

3. **Professional Appearance**
   - Looks polished and professional
   - Consistent styling
   - Appropriate use of color

---

## Risk Assessment

### Technical Risks

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Glamour rendering issues | Medium | Low | Fallback to plain text |
| Terminal compatibility | Medium | Medium | Extensive testing |
| Performance overhead | Low | Low | Profiling and optimization |
| Dependency conflicts | Low | Low | Version pinning |

### Timeline Risks

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Scope creep | High | Medium | Strict phase boundaries |
| Testing delays | Medium | Medium | Automated testing |
| Integration issues | Medium | Low | Incremental integration |

### Mitigation Strategies

1. **Technical:**
   - Always have fallback to plain output
   - Test on multiple terminals
   - Profile performance early
   - Use well-tested libraries

2. **Process:**
   - Stick to phased approach
   - Regular testing after each phase
   - Early integration with main.go
   - Continuous user feedback

---

## Comparison: Before vs After

### Before (Current)

```
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║   ██████╗ ██████╗ ██████╗ ███████╗     █████╗  ██████╗    ║
║  ██╔════╝██╔═══██╗██╔══██╗██╔════╝    ██╔══██╗██╔════╝    ║
║  ██║     ██║   ██║██║  ██║█████╗      ███████║██║  ███╗   ║
║  ██║     ██║   ██║██║  ██║██╔══╝      ██╔══██║██║   ██║   ║
║  ╚██████╗╚██████╔╝██████╔╝███████╗    ██║  ██║╚██████╔╝   ║
║   ╚═════╝ ╚═════╝ ╚═════╝ ╚══════╝    ╚═╝  ╚═╝ ╚═════╝    ║
║                                                           ║
║            AI-Powered Coding Assistant                    ║
║            Built with Google ADK Go                       ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝

Working directory: /Users/user/project

🤖 Agent: Thinking...

🔧 Tool: read_file
   Args: map[path:demo/file.c]

✓ Tool result: read_file
   Result: [Large output - 2543 bytes]

✓ Task completed
```

### After (Target)

```
╭────────────────────────────────────────────────────╮
│                                                    │
│  code_agent v1.0.0                                 │
│  google/gemini-2.5-flash                           │
│  ~/projects/myproject                              │
│                                                    │
╰────────────────────────────────────────────────────╯

╭─ What can I help you with?
╰─❯ 

### Cline is thinking

Looking at the file structure to understand the project...

### Cline is reading `demo/file.c`

### Cline responds

I found the issue in your calculator. The expression parser needs to handle 
operator precedence correctly. Let me show you the fix:

```diff
- result = a + b;
+ result = eval_expression(expr);
```

This change ensures that multiplication and division are evaluated before 
addition and subtraction.

### Cline is editing `demo/file.c`

```diff
@@ -45,7 +45,7 @@
 int calculate(char* expr) {
-    result = a + b;
+    result = eval_expression(expr);
     return result;
 }
```

### Task completed

Usage: ↑ 2.3k ↓ 856 → 1.5k $0.0023

╭─ Next task?
╰─❯ 
```

**Key Improvements Visible:**
1. Clean session banner with context
2. Markdown-rendered responses
3. Syntax-highlighted code blocks
4. Colored diff display
5. Clear section headers
6. Token usage summary
7. Professional, polished appearance

---

## Conclusion

Implementing these improvements will transform code_agent from a basic CLI tool into a professional, production-ready coding assistant with display quality that matches or exceeds Cline.

The modular architecture will make the code easier to maintain, test, and extend. The rich display features will significantly improve user experience and make the agent's thought process more transparent.

**Estimated Total Effort:** 120-160 hours (3-4 weeks for 1 developer)

**Priority:** High - Display quality significantly impacts user experience and perceived quality of the entire product.

---

## Next Steps

1. **Review this plan** with the team
2. **Approve dependencies** (all are standard, well-maintained libraries)
3. **Begin Phase 1** implementation
4. **Set up regular check-ins** to review progress
5. **Gather user feedback** throughout implementation

---

## Appendices

### A. Example Usage

```bash
# Rich output (default)
./code-agent "Fix the calculator bug"

# Plain output (for piping)
./code-agent --output-format plain "Create README" > output.txt

# JSON output (for programmatic use)
./code-agent --output-format json "List files" | jq .

# Disable animations
./code-agent --no-typewriter "Quick task"

# Verbose mode
./code-agent -v "Debug issue"
```

### B. Terminal Compatibility Matrix

| Terminal | Rich Format | Plain Format | JSON Format | Notes |
|----------|-------------|--------------|-------------|-------|
| iTerm2 | ✓ | ✓ | ✓ | Full support |
| Terminal.app | ✓ | ✓ | ✓ | Full support |
| VS Code | ✓ | ✓ | ✓ | Full support |
| tmux | ✓ | ✓ | ✓ | May need TERM=screen-256color |
| SSH | ✓ | ✓ | ✓ | Depends on client |
| Piped | - | ✓ | ✓ | Auto-detects, uses plain |
| CI/CD | - | ✓ | ✓ | Auto-detects, uses plain |

### C. Configuration File Example

```yaml
# ~/.code-agent/config.yaml
display:
  output_format: rich
  colors_enabled: true
  markdown_enabled: true
  typewriter_enabled: false
  typewriter_speed: 1.0
  verbose: false
```

### D. References

- Cline CLI: https://github.com/cline/cline/tree/main/cli
- Lipgloss: https://github.com/charmbracelet/lipgloss
- Glamour: https://github.com/charmbracelet/glamour
- Cobra: https://github.com/spf13/cobra
- Terminal Utilities: https://pkg.go.dev/golang.org/x/term

---

**Document Version:** 1.0  
**Last Updated:** November 10, 2025  
**Author:** AI Analysis Team  
**Status:** Ready for Review
