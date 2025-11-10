# CLI Display Improvement - Quick Reference

**Quick Start Guide for Implementing Cline-quality Display**

---

## TL;DR - What to Do

1. **Install dependencies**
   ```bash
   go get github.com/charmbracelet/lipgloss
   go get github.com/charmbracelet/glamour
   go get golang.org/x/term
   ```

2. **Create display package**
   ```bash
   mkdir code_agent/display
   ```

3. **Copy architecture from Cline**
   - Use modular design
   - Separate rendering concerns
   - Support multiple output formats

4. **Replace printEvent() with Renderer**
   - Use markdown rendering
   - Add rich formatting
   - Support TTY detection

---

## Key Architectural Decisions

### ✅ DO

1. **Use Charmbracelet ecosystem**
   - lipgloss for styling
   - glamour for markdown
   - Well-maintained, widely used

2. **Modular architecture**
   - One file per responsibility
   - Clear interfaces
   - Easy to test

3. **Support multiple formats**
   - Rich (default)
   - Plain (piped)
   - JSON (programmatic)

4. **Detect terminal capabilities**
   - Check if TTY
   - Adapt to environment
   - Fallback gracefully

### ❌ DON'T

1. **Don't hardcode ANSI**
   - Use lipgloss instead
   - Terminal-agnostic

2. **Don't put everything in main.go**
   - Separate display logic
   - Keep main.go simple

3. **Don't force colors**
   - Respect NO_COLOR
   - Detect TTY
   - Provide plain option

---

## File Structure

```
code_agent/
├── main.go                    # CLI entry point
├── display/
│   ├── renderer.go           # Main facade (START HERE)
│   ├── markdown_renderer.go  # Markdown support
│   ├── tool_renderer.go      # Tool-specific display
│   ├── banner.go             # Session info
│   ├── typewriter.go         # Animation (optional)
│   ├── streaming.go          # Real-time updates (optional)
│   └── ansi.go               # Terminal utils
└── agent/
    └── ... (existing)
```

---

## Quick Implementation Guide

### Step 1: Create Renderer (renderer.go)

```go
package display

import (
    "github.com/charmbracelet/glamour"
    "github.com/charmbracelet/lipgloss"
)

type Renderer struct {
    mdRenderer   *glamour.TermRenderer
    outputFormat string
    
    // Styles
    dimStyle    lipgloss.Style
    greenStyle  lipgloss.Style
    redStyle    lipgloss.Style
}

func NewRenderer(outputFormat string) (*Renderer, error) {
    mdRenderer, err := glamour.NewTermRenderer(
        glamour.WithAutoStyle(),
    )
    if err != nil {
        return nil, err
    }
    
    return &Renderer{
        mdRenderer:   mdRenderer,
        outputFormat: outputFormat,
        dimStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
        greenStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("2")),
        redStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("1")),
    }, nil
}

func (r *Renderer) RenderMarkdown(text string) (string, error) {
    if r.outputFormat == "plain" || !isTTY() {
        return text, nil
    }
    return r.mdRenderer.Render(text)
}
```

### Step 2: Update main.go

```go
package main

import (
    "code_agent/display"
    // ... other imports
)

func main() {
    // ... existing setup ...
    
    // Create renderer
    renderer, err := display.NewRenderer("rich")
    if err != nil {
        log.Fatalf("Failed to create renderer: %v", err)
    }
    
    // Use renderer instead of printEvent
    for event, err := range agentRunner.Run(...) {
        if err != nil {
            fmt.Println(renderer.Error(err.Error()))
            break
        }
        
        if event != nil {
            if err := renderer.RenderEvent(event); err != nil {
                log.Printf("Render error: %v", err)
            }
        }
    }
}
```

### Step 3: Add TTY Detection (ansi.go)

```go
package display

import (
    "os"
    "golang.org/x/term"
)

func isTTY() bool {
    return term.IsTerminal(int(os.Stdout.Fd()))
}

func GetTerminalWidth() int {
    width, _, err := term.GetSize(int(os.Stdout.Fd()))
    if err != nil {
        return 80
    }
    return width
}
```

---

## Markdown Rendering Examples

### Before (Plain Text)

```
🤖 Agent: Thinking...

I'll fix the calculator by implementing proper expression parsing.

🔧 Tool: write_file
```

### After (Rich Markdown)

```
### Cline is thinking

I'll fix the calculator by implementing proper expression parsing:

1. Parse the expression into tokens
2. Apply operator precedence
3. Evaluate the result

### Cline is writing `calculator.c`
```

**Key:** Use markdown syntax in your agent's responses, and glamour will render it beautifully.

---

## Tool Display Patterns

### Pattern 1: File Operations

```go
func (r *Renderer) RenderFileRead(path string) string {
    header := fmt.Sprintf("### Cline is reading `%s`\n", path)
    rendered, _ := r.mdRenderer.Render(header)
    return "\n" + rendered
}

func (r *Renderer) RenderFileWrite(path, content string) string {
    header := fmt.Sprintf("### Cline is writing `%s`\n", path)
    preview := fmt.Sprintf("```\n%s\n```", truncate(content, 500))
    full := header + "\n" + preview
    rendered, _ := r.mdRenderer.Render(full)
    return "\n" + rendered
}
```

### Pattern 2: Command Execution

```go
func (r *Renderer) RenderCommand(cmd string) string {
    header := fmt.Sprintf("### Running command\n\n```shell\n%s\n```", cmd)
    rendered, _ := r.mdRenderer.Render(header)
    return "\n" + rendered
}
```

### Pattern 3: Diffs

```go
func (r *Renderer) RenderDiff(diff string) string {
    formatted := fmt.Sprintf("```diff\n%s\n```", diff)
    rendered, _ := r.mdRenderer.Render(formatted)
    return rendered
}
```

---

## Lipgloss Styling Cheat Sheet

```go
// Text colors
dimStyle    := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
greenStyle  := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
redStyle    := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
yellowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
blueStyle   := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))

// Text attributes
boldStyle   := lipgloss.NewStyle().Bold(true)
italicStyle := lipgloss.NewStyle().Italic(true)

// Borders
boxStyle := lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color("39")).
    Padding(1, 4)

// Usage
fmt.Println(greenStyle.Render("✓ Success"))
fmt.Println(redStyle.Render("✗ Error"))
fmt.Println(boxStyle.Render("Banner content"))
```

---

## Output Format Support

```go
type Renderer struct {
    outputFormat string // "rich", "plain", or "json"
}

func (r *Renderer) RenderEvent(event *Event) error {
    switch r.outputFormat {
    case "json":
        return r.renderJSON(event)
    case "plain":
        return r.renderPlain(event)
    default: // "rich"
        return r.renderRich(event)
    }
}

func (r *Renderer) renderRich(event *Event) error {
    // Use markdown, colors, formatting
    markdown := fmt.Sprintf("### %s\n\n%s", event.Title, event.Content)
    rendered, _ := r.mdRenderer.Render(markdown)
    fmt.Print(rendered)
    return nil
}

func (r *Renderer) renderPlain(event *Event) error {
    // Plain text only
    fmt.Printf("%s: %s\n", event.Title, event.Content)
    return nil
}

func (r *Renderer) renderJSON(event *Event) error {
    // JSON output
    data, _ := json.Marshal(event)
    fmt.Println(string(data))
    return nil
}
```

---

## Testing Checklist

- [ ] Works in iTerm2
- [ ] Works in Terminal.app
- [ ] Works in VS Code terminal
- [ ] Works with piped output (`./code-agent | tee log.txt`)
- [ ] Works over SSH
- [ ] Respects NO_COLOR environment variable
- [ ] Plain format has no ANSI codes
- [ ] JSON format is valid
- [ ] Markdown renders correctly
- [ ] Code blocks have syntax highlighting
- [ ] Diffs are colored
- [ ] No errors with `go vet`
- [ ] No errors with `staticcheck`

---

## Common Pitfalls

### 1. Not Checking TTY

**Wrong:**
```go
// Always uses colors
fmt.Println("\033[32mSuccess\033[0m")
```

**Right:**
```go
if isTTY() {
    fmt.Println(greenStyle.Render("Success"))
} else {
    fmt.Println("Success")
}
```

### 2. Ignoring Output Format

**Wrong:**
```go
// Always uses markdown
rendered, _ := mdRenderer.Render(text)
fmt.Print(rendered)
```

**Right:**
```go
if r.outputFormat == "rich" && isTTY() {
    rendered, _ := r.mdRenderer.Render(text)
    fmt.Print(rendered)
} else {
    fmt.Print(text)
}
```

### 3. Not Handling Errors

**Wrong:**
```go
rendered, _ := r.mdRenderer.Render(text)
```

**Right:**
```go
rendered, err := r.mdRenderer.Render(text)
if err != nil {
    // Fallback to plain text
    rendered = text
}
```

---

## Performance Tips

1. **Reuse renderer instances**
   ```go
   // Create once
   renderer := NewRenderer("rich")
   
   // Use many times
   for event := range events {
       renderer.RenderEvent(event)
   }
   ```

2. **Cache styles**
   ```go
   type Renderer struct {
       greenStyle lipgloss.Style // Cache, don't recreate
   }
   ```

3. **Avoid unnecessary rendering**
   ```go
   if r.outputFormat == "json" {
       // Skip markdown rendering entirely
       return r.renderJSON(event)
   }
   ```

4. **Truncate large outputs**
   ```go
   if len(content) > 10000 {
       content = content[:10000] + "\n... (truncated)"
   }
   ```

---

## Migration Path

### Phase 1: Minimal (1 day)
- [ ] Add dependencies
- [ ] Create display/renderer.go
- [ ] Replace printEvent with basic renderer
- [ ] Test basic functionality

### Phase 2: Rich Text (2-3 days)
- [ ] Add markdown rendering
- [ ] Implement TTY detection
- [ ] Add tool-specific rendering
- [ ] Test in different terminals

### Phase 3: Polish (2-3 days)
- [ ] Add session banners
- [ ] Implement multiple output formats
- [ ] Add API usage display
- [ ] Final testing and refinement

---

## Resources

- **Lipgloss Examples:** https://github.com/charmbracelet/lipgloss/tree/master/examples
- **Glamour Styles:** https://github.com/charmbracelet/glamour/tree/master/styles
- **Cline Source:** https://github.com/cline/cline/tree/main/cli/pkg/cli/display
- **Terminal Codes:** https://gist.github.com/fnky/458719343aabd01cfb17a3a4f7296797

---

## Quick Commands

```bash
# Install dependencies
go get github.com/charmbracelet/lipgloss
go get github.com/charmbracelet/glamour
go get golang.org/x/term

# Run with different formats
./code-agent --output-format rich "Create README"
./code-agent --output-format plain "Create README"
./code-agent --output-format json "Create README"

# Test in different environments
./code-agent "Test task"                    # Normal
./code-agent "Test task" | cat              # Piped
NO_COLOR=1 ./code-agent "Test task"         # No colors
```

---

## Summary

**Priority order:**
1. Create modular display package ⭐⭐⭐
2. Add markdown rendering ⭐⭐⭐
3. Implement TTY detection ⭐⭐⭐
4. Add tool-specific rendering ⭐⭐
5. Support multiple formats ⭐⭐
6. Add session banners ⭐
7. Optional: Typewriter effect ⭐
8. Optional: Streaming display ⭐

**Time estimate:**
- Minimal: 1 day
- Good: 3-5 days
- Excellent: 1-2 weeks

**Key success factor:** Use the Charmbracelet ecosystem and follow Cline's modular architecture pattern.
