# CLI Display: code_agent vs Cline - Side-by-Side Comparison

**Detailed examples showing the differences and improvements**

---

## Overview

This document provides concrete examples of how the CLI display differs between code_agent (current) and Cline, with specific code snippets showing how to achieve Cline-level quality.

---

## 1. Basic Text Output

### code_agent (Current)

**Code:**
```go
fmt.Printf("%s%s🤖 Agent:%s Thinking...\n\n", colorBold, colorBlue, colorReset)
```

**Output:**
```
🤖 Agent: Thinking...
```

**Issues:**
- Hardcoded emoji
- Basic color
- No context
- Not semantic

### Cline

**Code:**
```go
header := "### Cline is thinking\n"
rendered, _ := mdRenderer.Render(header)
fmt.Print(rendered)
```

**Output:**
```
### Cline is thinking
```

(Rendered as styled heading in terminal)

**Advantages:**
- Semantic markdown
- Professional styling
- Contextual
- Consistent format

---

## 2. Tool Execution Display

### code_agent (Current)

**Code:**
```go
if part.FunctionCall != nil {
    fmt.Printf("%s%s🔧 Calling tool:%s %s\n", colorBold, colorYellow, colorReset, part.FunctionCall.Name)
    if len(fmt.Sprintf("%v", part.FunctionCall.Args)) < 200 {
        fmt.Printf("%s   Args:%s %v\n", colorBold, colorReset, part.FunctionCall.Args)
    }
}
```

**Output:**
```
🔧 Calling tool: read_file
   Args: map[path:demo/file.c]
```

**Issues:**
- Generic display
- No file path highlighting
- Args shown as raw map
- Not user-friendly

### Cline

**Code:**
```go
func (tr *ToolRenderer) RenderToolExecution(tool *ToolMessage) string {
    header := tr.generateToolHeader(tool, "is")
    rendered := tr.renderMarkdown(header)
    return "\n" + rendered + "\n"
}

func (tr *ToolRenderer) generateToolHeader(tool *ToolMessage, verbTense string) string {
    switch tool.Tool {
    case "read_file":
        action := "is reading"
        return fmt.Sprintf("### Cline %s `%s`", action, tool.Path)
    // ... other cases
    }
}
```

**Output:**
```
### Cline is reading `demo/file.c`
```

**Advantages:**
- Clear action
- File path highlighted in backticks
- Natural language
- Professional appearance

---

## 3. File Write Operations

### code_agent (Current)

**Code:**
```go
fmt.Printf("%s%s🔧 Tool:%s write_file\n", colorBold, colorYellow, colorReset)
fmt.Printf("%s   Args:%s map[content:[...] path:demo/file.c]\n", colorBold, colorReset)
```

**Output:**
```
🔧 Tool: write_file
   Args: map[content:[...] path:demo/file.c]
```

**Issues:**
- Doesn't show content preview
- Generic tool name
- No diff view
- Hard to understand changes

### Cline

**Code:**
```go
func (tr *ToolRenderer) RenderToolExecution(tool *ToolMessage) string {
    var output strings.Builder
    
    // Header
    header := fmt.Sprintf("### Cline is writing `%s`\n", tool.Path)
    rendered := tr.renderMarkdown(header)
    output.WriteString("\n" + rendered + "\n")
    
    // Content preview
    preview := strings.TrimSpace(tool.Content)
    if len(preview) > 1000 {
        preview = preview[:1000] + "..."
    }
    contentMd := fmt.Sprintf("```\n%s\n```", preview)
    contentRendered := tr.renderMarkdown(contentMd)
    output.WriteString("\n" + contentRendered + "\n")
    
    return output.String()
}
```

**Output:**
```
### Cline is writing `demo/file.c`

```c
#include <stdio.h>

int main() {
    printf("Hello, World!\n");
    return 0;
}
```

**Advantages:**
- Clear action
- Shows actual content
- Syntax highlighting
- Easy to review

---

## 4. File Edit Operations (Diff Display)

### code_agent (Current)

**Code:**
```go
fmt.Printf("%s%s🔧 Tool:%s replace_in_file\n", colorBold, colorYellow, colorReset)
fmt.Printf("%s   Args:%s map[new_text:... old_text:... path:file.c]\n", colorBold, colorReset)
```

**Output:**
```
🔧 Tool: replace_in_file
   Args: map[new_text:... old_text:... path:file.c]
```

**Issues:**
- No visual diff
- Can't see what changed
- Not reviewable
- Truncated content

### Cline

**Code:**
```go
func (tr *ToolRenderer) RenderToolExecution(tool *ToolMessage) string {
    var output strings.Builder
    
    header := fmt.Sprintf("### Cline is editing `%s`\n", tool.Path)
    rendered := tr.renderMarkdown(header)
    output.WriteString("\n" + rendered + "\n")
    
    // Show diff
    diffMarkdown := fmt.Sprintf("```diff\n%s\n```", tool.Content)
    diffRendered := tr.renderMarkdown(diffMarkdown)
    output.WriteString("\n" + diffRendered + "\n")
    
    return output.String()
}
```

**Output:**
```
### Cline is editing `demo/calculator.c`

```diff
@@ -45,7 +45,7 @@
 int calculate(char* expr) {
-    result = a + b;
+    result = eval_expression(expr);
     return result;
 }
```

**Advantages:**
- Shows unified diff
- Colored (green/red)
- Easy to review changes
- Professional format

---

## 5. Command Execution

### code_agent (Current)

**Code:**
```go
fmt.Printf("%s%s🔧 Tool:%s execute_command\n", colorBold, colorYellow, colorReset)
fmt.Printf("%s   Args:%s map[command:gcc file.c working_dir:.]\n", colorBold, colorReset)
```

**Output:**
```
🔧 Tool: execute_command
   Args: map[command:gcc file.c working_dir:.]
```

**Issues:**
- Doesn't show actual command clearly
- No output display
- Generic format

### Cline

**Code:**
```go
func (tr *ToolRenderer) RenderCommandExecution(command string) string {
    header := fmt.Sprintf("### Running command\n\n```shell\n%s\n```", command)
    rendered := tr.renderMarkdown(header)
    return "\n" + rendered + "\n"
}

func (tr *ToolRenderer) RenderCommandOutput(output string) string {
    var result strings.Builder
    
    header := "### Terminal output"
    rendered := tr.renderMarkdown(header)
    result.WriteString("\n" + rendered + "\n\n")
    
    outputBlock := fmt.Sprintf("```\n%s\n```", strings.TrimSpace(output))
    outputRendered := tr.renderMarkdown(outputBlock)
    result.WriteString(outputRendered + "\n")
    
    return result.String()
}
```

**Output:**
```
### Running command

```shell
gcc demo/calculator.c -o demo/calculate
```

### Terminal output

```
compilation successful
```

**Advantages:**
- Command clearly visible
- Separated output section
- Easy to understand flow
- Professional presentation

---

## 6. Agent Response / Thinking

### code_agent (Current)

**Code:**
```go
if part.Text != "" {
    text := part.Text
    if strings.Contains(text, "read_file") || strings.Contains(text, "write_file") {
        fmt.Printf("%s%s🔧 Tool:%s %s\n", colorBold, colorYellow, colorReset, text)
    } else {
        fmt.Printf("%s", text)
    }
}
```

**Output:**
```
I'll fix the calculator by implementing proper expression parsing. First I need to read the file.
```

**Issues:**
- Plain text only
- No formatting
- Hard to read longer responses
- No markdown support

### Cline

**Code:**
```go
func (r *Renderer) RenderMessage(text string) error {
    // Agent responses use markdown
    markdown := text // Agent already provides markdown
    rendered, err := r.mdRenderer.Render(markdown)
    if err != nil {
        return err
    }
    fmt.Print(rendered)
    return nil
}
```

**Output:**
```
### Cline responds

I'll fix the calculator by implementing proper expression parsing:

1. Parse the expression into tokens
2. Apply operator precedence rules
3. Evaluate the expression

Here's the plan:

- Read the current implementation
- Identify the issue
- Apply the fix

Let me start by reading the file.
```

(With proper markdown rendering: bold headings, numbered lists, etc.)

**Advantages:**
- Full markdown support
- Lists, headings, code blocks
- Much easier to read
- Professional appearance

---

## 7. Session Banner

### code_agent (Current)

**Code:**
```go
func printBanner() {
    banner := `
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
`
    fmt.Printf("%s%s%s%s\n", colorBold, colorCyan, banner, colorReset)
}
```

**Output:**
```
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║   ██████╗ ██████╗ ██████╗ ███████╗     █████╗  ██████╗    ║
║  ... (ASCII art)
║                                                           ║
╚═══════════════════════════════════════════════════════════╝

Working directory: /Users/user/project
```

**Issues:**
- Large, static ASCII art
- No contextual information
- Doesn't show model info
- Takes up screen space

### Cline

**Code:**
```go
func RenderSessionBanner(info BannerInfo) string {
    titleStyle := lipgloss.NewStyle().
        Foreground(lipgloss.Color("15")).
        Bold(true)
    
    dimStyle := lipgloss.NewStyle().
        Foreground(lipgloss.AdaptiveColor{Light: "248", Dark: "238"})
    
    borderColor := lipgloss.Color("39") // Blue
    
    boxStyle := lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(borderColor).
        Padding(1, 4)
    
    var lines []string
    
    // Title line
    lines = append(lines, titleStyle.Render("code_agent") + " " + dimStyle.Render("v"+info.Version))
    
    // Model line
    if info.Provider != "" && info.ModelID != "" {
        lines = append(lines, dimStyle.Render(info.Provider+"/"+info.ModelID))
    }
    
    // Workspace line
    if info.Workdir != "" {
        lines = append(lines, dimStyle.Render(shortenPath(info.Workdir, 45)))
    }
    
    content := lipgloss.JoinVertical(lipgloss.Left, lines...)
    return boxStyle.Render(content)
}
```

**Output:**
```
╭────────────────────────────────────────────────────╮
│                                                    │
│  code_agent v1.0.0                                 │
│  google/gemini-2.5-flash                           │
│  ~/projects/myproject                              │
│                                                    │
╰────────────────────────────────────────────────────╯
```

**Advantages:**
- Compact
- Shows useful info (model, workspace)
- Professional styling
- Adapts to terminal width
- Easy to read

---

## 8. Error Display

### code_agent (Current)

**Code:**
```go
if err != nil {
    fmt.Printf("%s%sError:%s %v\n", colorBold, colorRed, colorReset, err)
    hasError = true
    break
}
```

**Output:**
```
Error: failed to read file: no such file or directory
```

**Issues:**
- Generic error format
- No context
- No suggestions
- Just the raw error

### Cline

**Code:**
```go
func (r *Renderer) RenderError(err error) error {
    markdown := fmt.Sprintf(`### ✗ Error

%s

**Suggestion:** Check the file path and ensure it exists.

Try: \`ls\` to see available files.
`, err.Error())
    
    rendered, _ := r.mdRenderer.Render(markdown)
    fmt.Println(rendered)
    return nil
}
```

**Output:**
```
### ✗ Error

failed to read file: demo/calculator.c: no such file or directory

**Suggestion:** Check the file path and ensure it exists.

Try: `ls demo/` to see available files.
```

**Advantages:**
- Clear error section
- Provides suggestions
- Actionable advice
- Professional format

---

## 9. API Usage Display

### code_agent (Current)

**Not implemented** - No API usage display

### Cline

**Code:**
```go
func (r *Renderer) RenderAPI(status string, apiInfo *APIRequestInfo) error {
    if apiInfo.Cost >= 0 {
        usageInfo := r.formatUsageInfo(
            apiInfo.TokensIn, 
            apiInfo.TokensOut, 
            apiInfo.CacheReads, 
            apiInfo.CacheWrites, 
            apiInfo.Cost
        )
        markdown := fmt.Sprintf("## API %s `%s`", status, usageInfo)
        rendered := r.RenderMarkdown(markdown)
        fmt.Print(rendered)
    }
    return nil
}

func (r *Renderer) formatUsageInfo(tokensIn, tokensOut, cacheReads, cacheWrites int, cost float64) string {
    parts := make([]string, 0, 4)
    
    if tokensIn != 0 {
        parts = append(parts, fmt.Sprintf("↑ %s", formatNumber(tokensIn)))
    }
    if tokensOut != 0 {
        parts = append(parts, fmt.Sprintf("↓ %s", formatNumber(tokensOut)))
    }
    if cacheReads != 0 {
        parts = append(parts, fmt.Sprintf("→ %s", formatNumber(cacheReads)))
    }
    if cacheWrites != 0 {
        parts = append(parts, fmt.Sprintf("← %s", formatNumber(cacheWrites)))
    }
    
    return fmt.Sprintf("%s $%.4f", strings.Join(parts, " "), cost)
}

func formatNumber(n int) string {
    if n >= 1000000 {
        return fmt.Sprintf("%.1fm", float64(n)/1000000.0)
    } else if n >= 1000 {
        return fmt.Sprintf("%.1fk", float64(n)/1000.0)
    }
    return fmt.Sprintf("%d", n)
}
```

**Output:**
```
## API complete `↑ 2.3k ↓ 856 → 1.5k $0.0023`
```

**Advantages:**
- Shows token usage
- Cache hits/writes
- Cost tracking
- Compact format
- Professional

---

## 10. Overall Session Comparison

### code_agent (Current) - Full Session

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

╭─ Enter your coding task (or 'exit' to quit)
╰─❯ Fix the calculator

🤖 Agent: Thinking...

🔧 Tool: read_file
   Args: map[path:demo/calculator.c]

✓ Tool result: read_file
   Result: [Large output - 2543 bytes]

I'll fix the expression parser.

🔧 Tool: write_file
   Args: map[content:[...] path:demo/calculator.c]

✓ Tool result: write_file
   Result: File written successfully

✓ Task completed

╭─ Next task?
╰─❯ 
```

### Cline - Full Session

```
╭────────────────────────────────────────────────────╮
│                                                    │
│  code_agent v1.0.0                                 │
│  google/gemini-2.5-flash                           │
│  ~/projects/myproject                              │
│                                                    │
╰────────────────────────────────────────────────────╯

╭─ What can I help you with?
╰─❯ Fix the calculator

### Cline is thinking

I'll analyze the calculator code and fix the expression parsing issue.

### Cline is reading `demo/calculator.c`

### Cline responds

I found the issue. The calculator doesn't handle operator precedence correctly. 
Here's the problem:

```c
result = a + b;  // Doesn't respect order of operations
```

I'll fix this by implementing a proper expression evaluator.

### Cline is editing `demo/calculator.c`

```diff
@@ -45,7 +45,7 @@
 int calculate(char* expr) {
-    result = a + b;
+    result = eval_expression(expr);
     return result;
 }
```

### Task completed

**Usage:** ↑ 2.3k ↓ 856 → 1.5k $0.0023

╭─ What's next?
╰─❯ 
```

**Comparison Summary:**

| Aspect | code_agent | Cline |
|--------|------------|-------|
| Banner | ASCII art, large | Compact, informative |
| Tool display | Generic | Contextual, clear |
| File operations | Basic | Rich, with previews |
| Diffs | Not shown | Colored diffs |
| Agent responses | Plain text | Full markdown |
| Command execution | Basic | Separated sections |
| API usage | Not shown | Token/cost tracking |
| Errors | Basic | With suggestions |
| Overall readability | ⭐⭐ | ⭐⭐⭐⭐⭐ |
| Professional appearance | ⭐⭐ | ⭐⭐⭐⭐⭐ |

---

## Key Takeaways

### What Makes Cline Better

1. **Markdown rendering** - Rich text with proper formatting
2. **Contextual display** - Each tool shows relevant info
3. **Visual hierarchy** - Clear sections and headings
4. **Code highlighting** - Syntax-colored code blocks
5. **Diff display** - Easy to review changes
6. **Compact banner** - Useful info without clutter
7. **API tracking** - Transparent usage and costs
8. **Professional styling** - Consistent, polished appearance

### Implementation Priorities

1. **Must Have:**
   - Markdown rendering (glamour)
   - Contextual tool display
   - Rich text formatting

2. **Should Have:**
   - Session banners with context
   - Diff display
   - API usage tracking

3. **Nice to Have:**
   - Typewriter effects
   - Streaming display
   - Advanced animations

### Bottom Line

The improvement from code_agent to Cline-quality display requires:
- Using proper rendering libraries (glamour, lipgloss)
- Modular architecture
- Contextual tool rendering
- Markdown support throughout

**Estimated effort:** 1-2 weeks to match Cline's quality.
**Result:** Professional, production-ready CLI that users will love.
