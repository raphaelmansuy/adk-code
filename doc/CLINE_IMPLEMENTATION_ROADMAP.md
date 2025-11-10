# Implementation Roadmap: Cline Feature Parity

## Quick Start: Phases at a Glance

| Phase | Duration | Focus | Impact |
|-------|----------|-------|--------|
| Phase 1 | 2 weeks | Streaming & Permissions | UX & Safety |
| Phase 2 | 2 weeks | Error Handling & Tracking | Autonomy |
| Phase 3 | 2 weeks | Context Management | Large Projects |
| Phase 4 | 3 weeks | MCP & Multi-API | Extensibility |
| Phase 5 | 3 weeks | Browser & Checkpoints | Completeness |
| **Total** | **~12 weeks** | **Full Parity** | **Production Ready** |

---

## Phase 1: Streaming & Permissions (Weeks 1-2)

### Goal
Enable real-time feedback and safe agent operations

### 1.1 Streaming Output Support (2-3 days)

**What to build:**
- Stream LLM responses token-by-token
- Stream terminal output in real-time
- Progress visualization

**Implementation:**
```go
// In agent/streaming.go
package agent

import "io"

type StreamHandler interface {
    SendToken(token string)
    SendTerminalOutput(output string)
    SendProgress(step, total int)
    SendError(err error)
}

type StreamingConfig struct {
    Enabled    bool
    BufferSize int
    UpdateFreq time.Duration
}

func (a *CodingAgent) GenerateContentStreaming(
    ctx context.Context,
    req *genai.GenerateContentRequest,
    handler StreamHandler,
) error {
    // Use Gemini streaming API
    iter := a.model.GenerateContentStream(ctx, req.Contents...)
    
    for {
        resp, err := iter.Next()
        if err == iterator.Done {
            break
        }
        
        for _, part := range resp.Content.Parts {
            if text, ok := part.(genai.Text); ok {
                handler.SendToken(string(text))
            }
        }
    }
    return nil
}
```

**Testing:**
- Stream large responses
- Handle network interruptions
- Verify token counting

---

### 1.2 Permission & Approval System (2-3 days)

**What to build:**
- Track pending actions (file edits, commands)
- Show diffs before execution
- Require user approval
- Record approval decisions

**Implementation:**
```go
// In tools/approval.go
package tools

type PendingAction struct {
    ID       string
    Type     string // "edit", "run", "create", "delete"
    Target   string // File path or command
    Content  string // New content or command text
    Diff     *DiffView
    CreatedAt time.Time
}

type DiffView struct {
    FileName string
    Before   string
    After    string
    Hunks    []DiffHunk
}

type ApprovalManager struct {
    pending map[string]*PendingAction
    handler ApprovalHandler
}

type ApprovalHandler interface {
    RequestApproval(action *PendingAction) (bool, string, error) // approved, comment, error
}

func (am *ApprovalManager) RequestFileEdit(
    path, oldContent, newContent string,
) error {
    action := &PendingAction{
        ID:   uuid.New().String(),
        Type: "edit",
        Target: path,
        Content: newContent,
        Diff: generateDiff(path, oldContent, newContent),
    }
    
    am.pending[action.ID] = action
    approved, comment, err := am.handler.RequestApproval(action)
    
    if !approved {
        // Log rejection with comment
        return fmt.Errorf("user rejected edit: %s", comment)
    }
    
    // Proceed with edit
    return os.WriteFile(path, []byte(newContent), 0644)
}
```

**Testing:**
- Approval flow works end-to-end
- Diffs are accurate
- Rejections are handled

---

### 1.3 Permission UI Mockup (1-2 days)

**For CLI/Interactive Mode:**
```
Agent wants to: Create file src/main.go
────────────────────────────────────────
--- /dev/null
+++ src/main.go
@@ -0,0 +1,10 @@
+package main
+
+import "fmt"
+
+func main() {
+    fmt.Println("Hello, World!")
+}

[A]pprove  [R]eject  [E]dit  [C]ancel
```

**For VS Code Extension (if applicable):**
- Diff view in side panel
- Inline approve/reject buttons
- Comment field for feedback

---

## Phase 2: Error Handling & Tracking (Weeks 3-4)

### Goal
Enable autonomous error detection and recovery

### 2.1 Error Monitoring from Terminal (2-3 days)

**What to build:**
- Parse terminal output for errors
- Detect common error patterns
- Classify errors (syntax, type, runtime, etc.)
- Pass errors back to agent for fixing

**Implementation:**
```go
// In tools/error_monitor.go
package tools

type ErrorPattern struct {
    Regex     *regexp.Regexp
    Type      string // "syntax", "type", "import", "runtime"
    Extractor func(match []string) *CompileError
}

type CompileError struct {
    Type     string // Error classification
    File     string
    Line     int
    Column   int
    Message  string
    Severity string // "error", "warning"
}

type ErrorMonitor struct {
    patterns []*ErrorPattern
}

func NewErrorMonitor() *ErrorMonitor {
    return &ErrorMonitor{
        patterns: []*ErrorPattern{
            // TypeScript/JavaScript
            {
                Regex: regexp.MustCompile(`error TS(\d+): (.+?) at (.+?):(\d+):(\d+)`),
                Type:  "type",
            },
            // Go
            {
                Regex: regexp.MustCompile(`^(.+?):(\d+):(\d+): (.+)$`),
                Type:  "syntax",
            },
            // Python
            {
                Regex: regexp.MustCompile(`File "(.+?)", line (\d+).*\n(.+Error: .+)`),
                Type:  "runtime",
            },
            // Missing imports (common pattern)
            {
                Regex: regexp.MustCompile(`(?i)cannot find module|no such file|not found|undefined`),
                Type:  "import",
            },
        },
    }
}

func (em *ErrorMonitor) ParseErrors(output string) []*CompileError {
    var errors []*CompileError
    
    for _, pattern := range em.patterns {
        matches := pattern.Regex.FindAllStringSubmatch(output, -1)
        for _, match := range matches {
            err := pattern.Extractor(match)
            errors = append(errors, err)
        }
    }
    
    return errors
}

func (em *ErrorMonitor) ShouldAutoFix(err *CompileError) bool {
    // Agent can auto-fix these
    autoFixable := map[string]bool{
        "import": true,
        "syntax": true,
        "type":   true,
    }
    return autoFixable[err.Type]
}
```

**Integration with execute_command:**
```go
func (ect *ExecuteCommandTool) ExecuteWithErrorMonitoring(
    ctx context.Context,
    command, workingDir string,
    onError func(*CompileError) error,
) error {
    // ... execute command ...
    
    // Parse output for errors
    errors := errorMonitor.ParseErrors(combinedOutput)
    
    for _, err := range errors {
        if errorMonitor.ShouldAutoFix(err) {
            if err := onError(err); err != nil {
                // Pass back to agent for fixing
                return err
            }
        }
    }
    
    return nil
}
```

---

### 2.2 Token Counting (1-2 days)

**What to build:**
- Estimate tokens for requests
- Track cumulative usage
- Calculate API costs
- Display to user

**Implementation:**
```go
// In agent/token_tracker.go
package agent

import "github.com/tiktoken-go/tokenizer"

type TokenTracker struct {
    totalTokens      int
    totalCost        float64
    requestsTokens   []int
    modelPricing     map[string]Pricing
}

type Pricing struct {
    InputPerMTok  float64 // Price per million tokens
    OutputPerMTok float64
}

func NewTokenTracker() *TokenTracker {
    return &TokenTracker{
        modelPricing: map[string]Pricing{
            "gemini-2.5-flash": {
                InputPerMTok:  0.075 / 1000, // $0.075 per 1M
                OutputPerMTok: 0.3 / 1000,   // $0.3 per 1M
            },
            "gpt-4": {
                InputPerMTok:  0.03,
                OutputPerMTok: 0.06,
            },
            "claude-3-sonnet": {
                InputPerMTok:  0.003,
                OutputPerMTok: 0.015,
            },
        },
    }
}

func (tt *TokenTracker) EstimateTokens(text string, model string) int {
    // Use tiktoken for estimation
    enc := tokenizer.NewEncoding("cl100k_base")
    tokens, _, err := enc.Encode(text)
    if err != nil {
        // Fallback: rough estimation
        return len(text) / 4
    }
    return len(tokens)
}

func (tt *TokenTracker) RecordRequest(
    model string,
    inputTokens, outputTokens int,
) {
    pricing, exists := tt.modelPricing[model]
    if !exists {
        return
    }
    
    cost := (float64(inputTokens) * pricing.InputPerMTok) +
            (float64(outputTokens) * pricing.OutputPerMTok)
    
    tt.totalTokens += inputTokens + outputTokens
    tt.totalCost += cost
    tt.requestsTokens = append(tt.requestsTokens, inputTokens+outputTokens)
}

func (tt *TokenTracker) Report() string {
    return fmt.Sprintf(
        "Tokens: %d | Cost: $%.2f | Requests: %d",
        tt.totalTokens,
        tt.totalCost,
        len(tt.requestsTokens),
    )
}
```

---

### 2.3 Auto-Fix Common Issues (1-2 days)

**What to build:**
- Fix missing imports automatically
- Fix common syntax errors
- Fix type errors when obvious

**Implementation:**
```go
// In agent/auto_fix.go
package agent

type AutoFixer struct {
    fileOps *FileOperations
}

func (af *AutoFixer) FixMissingImport(
    file string,
    moduleName string,
    language string,
) error {
    content, _ := os.ReadFile(file)
    
    switch language {
    case "go":
        return af.fixGoImport(file, string(content), moduleName)
    case "python":
        return af.fixPythonImport(file, string(content), moduleName)
    case "typescript":
        return af.fixTypeScriptImport(file, string(content), moduleName)
    }
    return fmt.Errorf("unsupported language: %s", language)
}

func (af *AutoFixer) fixGoImport(file, content, moduleName string) error {
    // Find import block
    // Add new import
    // Update file
    // Return no error if fixed
    return nil
}
```

---

## Phase 3: Context Management (Weeks 5-6)

### Goal
Handle large projects with smart context selection

### 3.1 @file, @folder, @url, @problems Support (2-3 days)

**What to build:**
- Parse context tokens from user input
- Fetch and process context
- Add to agent context window

**Implementation:**
```go
// In agent/context.go
package agent

type ContextToken struct {
    Type    string // "file", "folder", "url", "problems"
    Value   string // Path or URL
    Content string
    Tokens  int
}

type ContextManager struct {
    tokens      []*ContextToken
    budget      int // Max tokens for context
    remaining   int
}

func (cm *ContextManager) ParseContextFromInput(input string) []*ContextToken {
    var tokens []*ContextToken
    
    // Find @file mentions: @file path/to/file
    fileMatches := regexp.MustCompile(`@file\s+(\S+)`).FindAllStringSubmatch(input, -1)
    for _, match := range fileMatches {
        content, _ := os.ReadFile(match[1])
        token := &ContextToken{
            Type:    "file",
            Value:   match[1],
            Content: string(content),
            Tokens:  estimateTokens(string(content)),
        }
        tokens = append(tokens, token)
    }
    
    // Find @folder mentions: @folder path/to/folder
    folderMatches := regexp.MustCompile(`@folder\s+(\S+)`).FindAllStringSubmatch(input, -1)
    for _, match := range folderMatches {
        contents := readFolder(match[1])
        token := &ContextToken{
            Type:    "folder",
            Value:   match[1],
            Content: contents,
            Tokens:  estimateTokens(contents),
        }
        tokens = append(tokens, token)
    }
    
    // Find @url mentions: @url https://...
    urlMatches := regexp.MustCompile(`@url\s+(https?://\S+)`).FindAllStringSubmatch(input, -1)
    for _, match := range urlMatches {
        content, _ := fetchAndConvertURL(match[1])
        token := &ContextToken{
            Type:    "url",
            Value:   match[1],
            Content: content,
            Tokens:  estimateTokens(content),
        }
        tokens = append(tokens, token)
    }
    
    return tokens
}

func (cm *ContextManager) AddContextToPrompt(basePrompt string, tokens []*ContextToken) string {
    var contextParts []string
    
    for _, token := range tokens {
        if cm.remaining < token.Tokens {
            break // Out of budget
        }
        
        contextParts = append(contextParts, fmt.Sprintf(
            "```%s\n%s\n```",
            token.Type,
            token.Content,
        ))
        
        cm.remaining -= token.Tokens
    }
    
    if len(contextParts) == 0 {
        return basePrompt
    }
    
    return fmt.Sprintf(
        "%s\n\n## Additional Context\n\n%s",
        basePrompt,
        strings.Join(contextParts, "\n\n"),
    )
}
```

---

### 3.2 @problems Support (1-2 days)

**What to build:**
- Read VS Code problems panel
- Format and add to context

**Implementation:**
```go
// In agent/problems.go
package agent

type Problem struct {
    File     string
    Line     int
    Column   int
    Message  string
    Severity string // "error", "warning"
}

func GetWorkspaceProblems(workspaceRoot string) []*Problem {
    // Check linters/compilers
    problems := []*Problem{}
    
    // Run linters (example for Go)
    if fileExists(filepath.Join(workspaceRoot, "go.mod")) {
        goProblems := runGoLinter(workspaceRoot)
        problems = append(problems, goProblems...)
    }
    
    // Run other linters as needed
    
    return problems
}

func (cm *ContextManager) AddProblems() error {
    problems := GetWorkspaceProblems(".")
    
    var content strings.Builder
    for _, p := range problems {
        content.WriteString(fmt.Sprintf(
            "%s:%d:%d [%s] %s\n",
            p.File, p.Line, p.Column, p.Severity, p.Message,
        ))
    }
    
    token := &ContextToken{
        Type:    "problems",
        Value:   "workspace_problems",
        Content: content.String(),
        Tokens:  estimateTokens(content.String()),
    }
    
    cm.tokens = append(cm.tokens, token)
    return nil
}
```

---

## Phase 4: MCP & Multi-API Support (Weeks 7-9)

### Goal
Enable extensibility and flexibility in model/API choice

### 4.1 Multi-API Support (2-3 days)

**What to build:**
- Support multiple AI providers
- Configuration-based selection
- Cost tracking per provider

**Implementation:**
```go
// In model/provider.go
package model

type ModelProvider interface {
    GenerateContent(ctx context.Context, req *genai.GenerateContentRequest) (*genai.GenerateContentResponse, error)
    CountTokens(ctx context.Context, content ...interface{}) (int, error)
    ListModels(ctx context.Context) ([]string, error)
    GetPricing(model string) *PricingInfo
}

// Anthropic provider
type AnthropicProvider struct {
    client *anthropic.Client
    apiKey string
}

func NewAnthropicProvider(apiKey string) *AnthropicProvider {
    return &AnthropicProvider{
        apiKey: apiKey,
        client: anthropic.NewClient(apiKey),
    }
}

func (ap *AnthropicProvider) GenerateContent(
    ctx context.Context,
    req *genai.GenerateContentRequest,
) (*genai.GenerateContentResponse, error) {
    // Convert genai request to anthropic format
    // Make API call
    // Convert response back
    return nil, nil
}

// OpenAI provider
type OpenAIProvider struct {
    client *openai.Client
    apiKey string
}

// ... similar implementation ...

// Factory
type ProviderFactory struct {
    providers map[string]ModelProvider
}

func (pf *ProviderFactory) GetProvider(providerName string) ModelProvider {
    return pf.providers[providerName]
}

func (pf *ProviderFactory) Register(name string, provider ModelProvider) {
    pf.providers[name] = provider
}
```

**Configuration:**
```yaml
# config.yaml
model:
  provider: "anthropic" # or "openai", "gemini", "bedrock"
  apiKey: "${ANTHROPIC_API_KEY}"
  model: "claude-3-sonnet"
  
providers:
  anthropic:
    apiKey: "${ANTHROPIC_API_KEY}"
  openai:
    apiKey: "${OPENAI_API_KEY}"
  gemini:
    apiKey: "${GOOGLE_API_KEY}"
```

---

### 4.2 MCP Server Implementation (3-4 days)

**What to build:**
- MCP server that hosts tools
- Tool registry
- JSON-RPC communication

**Implementation:**
```go
// In mcp/server.go
package mcp

import "github.com/modelcontextprotocol/go-sdk"

type MCPServer struct {
    server    *mcp.Server
    tools     map[string]*ToolDefinition
    resources map[string]*Resource
}

type ToolDefinition struct {
    Name        string
    Description string
    InputSchema interface{} // JSON schema
    Handler     ToolHandler
}

type ToolHandler func(args map[string]interface{}) (interface{}, error)

func NewMCPServer() *MCPServer {
    server := mcp.NewServer()
    return &MCPServer{
        server:    server,
        tools:     make(map[string]*ToolDefinition),
        resources: make(map[string]*Resource),
    }
}

func (ms *MCPServer) RegisterTool(def *ToolDefinition) {
    ms.tools[def.Name] = def
    
    // Register with MCP server
    ms.server.AddTool(&mcp.Tool{
        Name:        def.Name,
        Description: def.Description,
        InputSchema: def.InputSchema,
    })
}

func (ms *MCPServer) CallTool(name string, args map[string]interface{}) (interface{}, error) {
    tool, exists := ms.tools[name]
    if !exists {
        return nil, fmt.Errorf("tool not found: %s", name)
    }
    
    return tool.Handler(args)
}

func (ms *MCPServer) Start(port int) error {
    return ms.server.Listen(port)
}
```

**Custom Tool Creation:**
```go
// Example: GitHub tool
func createGitHubTool() *ToolDefinition {
    return &ToolDefinition{
        Name:        "github",
        Description: "Fetch GitHub issues and create PRs",
        InputSchema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "action": map[string]interface{}{
                    "type": "string",
                    "enum": []string{"list_issues", "create_pr"},
                },
                "owner":  map[string]interface{}{"type": "string"},
                "repo":   map[string]interface{}{"type": "string"},
            },
        },
        Handler: func(args map[string]interface{}) (interface{}, error) {
            action := args["action"].(string)
            owner := args["owner"].(string)
            repo := args["repo"].(string)
            
            switch action {
            case "list_issues":
                return github.ListIssues(owner, repo)
            case "create_pr":
                // ... handle PR creation ...
            }
            
            return nil, nil
        },
    }
}
```

---

## Phase 5: Browser Automation & Checkpoints (Weeks 10-12)

### Goal
Complete feature parity with browser testing and safe exploration

### 5.1 Browser Automation (3-4 days)

**What to build:**
- Launch and control browser
- Take screenshots
- Click elements
- Type text
- Monitor console

**Implementation:**
```go
// In tools/browser.go
package tools

import "github.com/playwright-community/playwright-go"

type BrowserTool struct {
    browser *playwright.Browser
    page    *playwright.Page
}

func NewBrowserTool() (*BrowserTool, error) {
    pw := playwright.PlaywrightOptions{}
    if err := playwright.Install(&pw); err != nil {
        return nil, err
    }
    
    browser, err := playwright.Chromium.Launch()
    if err != nil {
        return nil, err
    }
    
    page, err := browser.NewPage()
    if err != nil {
        return nil, err
    }
    
    return &BrowserTool{
        browser: browser,
        page:    page,
    }, nil
}

func (bt *BrowserTool) Navigate(url string) error {
    _, err := bt.page.Goto(url)
    return err
}

func (bt *BrowserTool) TakeScreenshot() ([]byte, error) {
    return bt.page.Screenshot(playwright.PageScreenshotOptions{
        FullPage: playwright.Bool(true),
    })
}

func (bt *BrowserTool) Click(selector string) error {
    return bt.page.Click(selector)
}

func (bt *BrowserTool) Type(selector, text string) error {
    return bt.page.Fill(selector, text)
}

func (bt *BrowserTool) GetConsoleLogs() []string {
    // Collect console messages
    return []string{} // Simplified
}

func (bt *BrowserTool) Close() error {
    return bt.browser.Close()
}
```

---

### 5.2 Checkpoint System (2-3 days)

**What to build:**
- Snapshot workspace at each step
- Compare snapshots (diffs)
- Restore to previous state

**Implementation:**
```go
// In agent/checkpoint.go
package agent

type Checkpoint struct {
    ID        string
    Timestamp time.Time
    Files     map[string]string // path -> content
    Description string
}

type CheckpointManager struct {
    checkpoints []*Checkpoint
    current     int
}

func (cm *CheckpointManager) CreateCheckpoint(description string) (*Checkpoint, error) {
    cp := &Checkpoint{
        ID:          uuid.New().String(),
        Timestamp:   time.Now(),
        Files:       make(map[string]string),
        Description: description,
    }
    
    // Walk filesystem and capture file contents
    filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
        if !info.IsDir() && !shouldIgnore(path) {
            content, _ := os.ReadFile(path)
            cp.Files[path] = string(content)
        }
        return nil
    })
    
    cm.checkpoints = append(cm.checkpoints, cp)
    cm.current = len(cm.checkpoints) - 1
    
    return cp, nil
}

func (cm *CheckpointManager) CompareCheckpoints(cp1, cp2 *Checkpoint) *DiffReport {
    report := &DiffReport{
        Added:    []string{},
        Deleted:  []string{},
        Modified: map[string]*FileDiff{},
    }
    
    // Compare files in cp1 vs cp2
    allFiles := make(map[string]bool)
    for file := range cp1.Files {
        allFiles[file] = true
    }
    for file := range cp2.Files {
        allFiles[file] = true
    }
    
    for file := range allFiles {
        content1, exists1 := cp1.Files[file]
        content2, exists2 := cp2.Files[file]
        
        if exists1 && !exists2 {
            report.Deleted = append(report.Deleted, file)
        } else if !exists1 && exists2 {
            report.Added = append(report.Added, file)
        } else if content1 != content2 {
            report.Modified[file] = generateDiff(file, content1, content2)
        }
    }
    
    return report
}

func (cm *CheckpointManager) Restore(cp *Checkpoint) error {
    // Clear current workspace
    filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
        if !shouldIgnore(path) {
            os.RemoveAll(path)
        }
        return nil
    })
    
    // Restore from checkpoint
    for path, content := range cp.Files {
        os.MkdirAll(filepath.Dir(path), 0755)
        os.WriteFile(path, []byte(content), 0644)
    }
    
    return nil
}
```

---

## Quick Implementation Checklist

### Week 1-2 (Phase 1)
- [ ] Implement streaming output
- [ ] Build approval system with diffs
- [ ] Create approval UI for CLI

### Week 3-4 (Phase 2)
- [ ] Error monitoring and parsing
- [ ] Token counting integration
- [ ] Auto-fix common errors

### Week 5-6 (Phase 3)
- [ ] @file context support
- [ ] @folder context support
- [ ] @url fetching and markdown conversion
- [ ] @problems workspace integration

### Week 7-9 (Phase 4)
- [ ] Multi-API support (OpenAI, Anthropic, etc.)
- [ ] MCP server framework
- [ ] Tool registration
- [ ] Example MCP tools

### Week 10-12 (Phase 5)
- [ ] Browser automation setup
- [ ] Screenshot and interaction tools
- [ ] Checkpoint creation and restore
- [ ] Diff comparison UI

---

## Success Criteria

- [ ] Streaming output works smoothly
- [ ] All file edits require approval
- [ ] Errors are detected and fixed autonomously
- [ ] Context tokens work (@file, @folder, @url, @problems)
- [ ] Multiple APIs can be used
- [ ] Custom tools can be created via MCP
- [ ] Browser can be controlled for testing
- [ ] Workspace can be checkpointed and restored
- [ ] Token usage is tracked and displayed
- [ ] Large codebases are handled efficiently

---

## Testing Strategy

For each phase:
1. Unit tests for new functions
2. Integration tests with real LLM
3. Manual testing with different project types
4. Performance testing for large codebases

---

## Deployment Plan

1. Release Phase 1 features in v0.2
2. Release Phase 2 features in v0.3
3. Release Phase 3 features in v0.4
4. Release Phase 4 features in v0.5
5. Release Phase 5 features in v1.0 (full Cline parity)

Each release should work independently and add value.

---

## References & Resources

- [Cline GitHub](https://github.com/cline/cline)
- [Cline Architecture](https://docs.cline.bot)
- [Playwright Go](https://playwright.dev/go/)
- [MCP Specification](https://modelcontextprotocol.io/)
- [Anthropic SDK Go](https://github.com/anthropics/anthropic-sdk-go)
- [OpenAI Go SDK](https://github.com/openai/openai-go)
